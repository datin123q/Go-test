package news

import (
	"context"
	"path"
	"strings"

	"github.com/pkg/errors"
	"github.com/samber/lo"
	"github.com/sirupsen/logrus"

	"github.com/illuminarean/energy-shares-backend-repo/pkg/filehelper"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/idgenerator"

	adomain "github.com/illuminarean/energy-share-core-api/domain"
	cdomain "github.com/illuminarean/energy-shares-backend-repo/clean-arch-domain"
	"github.com/illuminarean/energy-shares-backend-repo/domain"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/config"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/factory"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/s3utils"
)

type newsUsecase struct {
	cfg                               *config.Config
	idGenerator                       idgenerator.IdGenerator
	s3pkg                             s3utils.S3
	fileHelper                        filehelper.FileHelper
	logger                            factory.LoggerFactory
	mailService                       cdomain.MailService
	newsRepository                    cdomain.NewsRepository
	attachmentRepository              cdomain.AttachmentRepository
	contentImageRepository            cdomain.ContentImageRepository
	operatorHistoryRepository         cdomain.OperatorHistoryRepository
	getAllEmailsOfUsersAndSubscribers func(ctx context.Context) ([]string, error)
}

func New(
	cfg *config.Config,
	idGenerator idgenerator.IdGenerator,
	s3pkg s3utils.S3,
	fileHelper filehelper.FileHelper,
	logger factory.LoggerFactory,
	mailService cdomain.MailService,
	newsRepository cdomain.NewsRepository,
	attachmentRepository cdomain.AttachmentRepository,
	contentImageRepository cdomain.ContentImageRepository,
	operatorHistoryRepository cdomain.OperatorHistoryRepository,
) adomain.NewsUsecase {
	return &newsUsecase{
		cfg:                               cfg,
		idGenerator:                       idGenerator,
		s3pkg:                             s3pkg,
		fileHelper:                        fileHelper,
		logger:                            logger,
		mailService:                       mailService,
		getAllEmailsOfUsersAndSubscribers: getAllEmailsOfUsersAndSubscribers,
		newsRepository:                    newsRepository,
		attachmentRepository:              attachmentRepository,
		contentImageRepository:            contentImageRepository,
		operatorHistoryRepository:         operatorHistoryRepository,
	}
}

func (u newsUsecase) GetAll(ctx context.Context, request adomain.NewsGetAllRequest) (adomain.NewsGetAllResponse, error) {

	total, items, err := u.newsRepository.GetAll(ctx,
		cdomain.NewsQueryOptionFunc().WithTitle(request.Title),
		cdomain.NewsQueryOptionFunc().WithPosted(request.Posted),
		cdomain.NewsQueryOptionFunc().WithClosingDate(request.OnlyNotExpired),
		cdomain.NewsQueryOptionFunc().WithStartDate(request.StartDate),
		cdomain.NewsQueryOptionFunc().WithEndDate(request.EndDate),
		cdomain.NewsQueryOptionFunc().WithSortBy(request.Sortby),
		cdomain.NewsQueryOptionFunc().WithOrder(request.Order),
		cdomain.NewsQueryOptionFunc().WithOffset(request.SkipCount),
		cdomain.NewsQueryOptionFunc().WithLimit(request.MaxResultCount),
	)

	if err != nil {
		return adomain.NewsGetAllResponse{}, errors.Wrap(err, "err in newsRepository.GetAll")
	}

	for i := range items {
		items[i].Attachments = make([]*domain.Attachment, 0)
		attachments, err := u.attachmentRepository.GetAll(ctx,
			cdomain.AttachmentQueryOptionFunc().WithResourceId(items[i].CommonId),
		)
		if err != nil {
			return adomain.NewsGetAllResponse{}, errors.Wrap(err, "err in attachmentRepository.GetAll")
		}

		for _, attachment := range attachments {
			items[i].Attachments = append(items[i].Attachments, &attachment)
		}
	}

	if err = u.operatorHistoryRepository.Create(ctx, domain.OperatorHistory{
		UserId: request.JWTUser.Id,
		Name:   request.JWTUser.Name,
		Email:  request.JWTUser.Email,
		Menu:   "menu.manage_website.manage_news_list",
		Ip:     request.ClientIp,
		State:  domain.LOOKUP,
	}); err != nil {
		return adomain.NewsGetAllResponse{}, errors.Wrap(err, "err in operatorHistoryRepository.Create")
	}

	return adomain.NewsGetAllResponse{
		TotalCount: total,
		Items:      items,
	}, nil
}

func (u newsUsecase) GetOne(ctx context.Context, request adomain.NewsGetOneRequest) (adomain.NewsGetOneResponse, error) {

	news, err := u.newsRepository.GetById(ctx, request.NewsId)
	if err != nil {
		return adomain.NewsGetOneResponse{}, errors.Wrap(err, "err in newsRepository.GetById")
	}

	attachments, err := u.attachmentRepository.GetAll(ctx,
		cdomain.AttachmentQueryOptionFunc().WithResourceId(news.CommonId),
	)
	if err != nil {
		return adomain.NewsGetOneResponse{}, errors.Wrap(err, "failed to get attachments")
	}

	news.Attachments = make([]*domain.Attachment, 0)
	for _, attachment := range attachments {
		news.Attachments = append(news.Attachments, &attachment)
	}

	if err = u.operatorHistoryRepository.Create(ctx, domain.OperatorHistory{
		UserId: request.JWTUser.Id,
		Name:   request.JWTUser.Name,
		Email:  request.JWTUser.Email,
		Menu:   "menu.manage_website.manage_news_detail",
		Ip:     request.ClientIp,
		State:  domain.LOOKUP,
	}); err != nil {
		return adomain.NewsGetOneResponse{}, errors.Wrap(err, "failed to create operator history")
	}

	return adomain.NewsGetOneResponse{News: news}, nil
}

const _maxPinCount = 4

func (u newsUsecase) Create(ctx context.Context, request adomain.NewsCreateRequest) error {
	var commonId = u.idGenerator.GenUuidV4().String()

	// 이미 pin이 true인 것이 최대 갯수 또는 그 이상이라면 에러 리턴(물론 이상이 되면 안된다)
	if lo.FromPtr(request.Pin) == true {
		count, _, err := u.newsRepository.GetAll(ctx, cdomain.NewsQueryOptionFunc().WithPin(request.Pin))
		if err != nil {
			return errors.Wrap(err, "err in newsRepository.GetAll")
		}

		if count >= _maxPinCount {
			return cdomain.ErrRequestLimitExceeded
		}
	}

	// 첨부 파일관련 파일 찾기
	for _, attachmentUrl := range request.AttachmentUrls {
		filePath, err := u.fileHelper.ExtractFilePath(attachmentUrl)
		if err != nil {
			return errors.Wrap(err, "failed to extract file path")
		}

		normalizedPath := filehelper.NormalizeS3FilePath(u.cfg.AWS.S3_PUBLIC_BUCKET, filePath)
		s3Key := strings.TrimPrefix(normalizedPath, "/")

		// S3 에서 파일 메타정보 가저오기
		header, err := u.s3pkg.GetHeadObject(u.cfg.AWS.S3_PUBLIC_BUCKET, s3Key)
		if err != nil {
			return errors.Wrap(err, "err in s3pkg.GetHeadObject")
		}

		// Attachment Repository 생성
		if _, err = u.attachmentRepository.Create(ctx, domain.Attachment{
			FileName:      u.fileHelper.ExtractFileName(normalizedPath),
			OriginalName:  header.OriginalFileName,
			FileSize:      header.FileSize,
			FileExtension: path.Ext(normalizedPath),
			FilePath:      path.Dir(normalizedPath) + "/",
			FileUrl:       s3utils.MakeCDNURL(normalizedPath),
			ResourceId:    commonId,
		}); err != nil {
			return errors.Wrap(err, "err in attachmentRepository.Create")
		}
	}

	// 컨텐츠 이미지 파일 찾기
	for _, contentImageUrl := range request.ContentImageUrls {
		filePath, err := u.fileHelper.ExtractFilePath(contentImageUrl)
		if err != nil {
			return errors.Wrap(err, "failed to extract file path")
		}

		normalizedPath := filehelper.NormalizeS3FilePath(u.cfg.AWS.S3_PUBLIC_BUCKET, filePath)
		s3Key := strings.TrimPrefix(normalizedPath, "/")

		// S3 에서 파일 메타정보 가저오기
		header, err := u.s3pkg.GetHeadObject(u.cfg.AWS.S3_PUBLIC_BUCKET, s3Key)
		if err != nil {
			return errors.Wrap(err, "err in s3pkg.GetHeadObject")
		}

		// Content Image Repository 생성
		if _, err = u.contentImageRepository.Create(ctx, domain.ContentImage{
			FileName:      u.fileHelper.ExtractFileName(normalizedPath),
			OriginalName:  header.OriginalFileName,
			FileSize:      header.FileSize,
			FileExtension: path.Ext(normalizedPath),
			FilePath:      path.Dir(normalizedPath) + "/",
			FileUrl:       s3utils.MakeCDNURL(normalizedPath),
			ResourceId:    commonId,
		}); err != nil {
			return errors.Wrap(err, "err in contentImageRepository.Create")
		}
	}

	newsId, err := u.newsRepository.Create(ctx, domain.News{
		CommonId:      commonId,
		Pin:           lo.FromPtr(request.Pin),
		Title:         request.Title,
		OwnerId:       request.JWTUser.Id,
		UpdateId:      request.JWTUser.Id,
		ContentHtml:   request.ContentHtml,
		Posted:        request.Posted,
		ClosingDate:   lo.FromPtr(request.ClosingDate),
		AttachmentCnt: len(request.AttachmentUrls),
	})
	if err != nil {
		return errors.Wrap(err, "err in newsRepository.Create")
	}

	// OperatorHistory 등록
	if err := u.operatorHistoryRepository.Create(ctx, domain.OperatorHistory{
		UserId: request.JWTUser.Id,
		Name:   request.JWTUser.Name,
		Email:  request.JWTUser.Email,
		Menu:   "menu.manage_website.manage_news_register",
		Ip:     request.ClientIp,
		State:  domain.REGISTER,
	}); err != nil {
		return errors.Wrap(err, "err in operatorHistoryRepository.Create")
	}

	// 메일 발송 프로세스
	if request.SendMailStatus == true {
		// 메일 생성시간 확인용
		createdNews, err := u.newsRepository.GetById(ctx, newsId)
		if err != nil {
			return errors.Wrap(err, "err in newsRepository.GetById")
		}

		// 수신자 추출
		receivers, err := u.getAllEmailsOfUsersAndSubscribers(ctx)
		if err != nil {
			return errors.Wrap(err, "err in getAllEmailsOfUsersAndSubscribers")
		}

		// 메일 발송
		err = u.mailService.SendMail(ctx, cdomain.SendMailRequest{
			TemplateType: domain.NEWS_UPLOAD_TEMPLATE,
			Sender:       u.cfg.ServiceConfig.Mail.Info,
			Receivers:    receivers,
			MailType:     domain.MAIL_SEND_PRIORITY_NORMAL,
			Details: map[string]interface{}{
				"Title":      request.Title,
				"Content":    request.ContentHtml,
				"Time":       createdNews.Created.ToString(),
				"NewsView":   path.Join(u.cfg.ServiceConfig.Url.UserServiceDomain, "user-support") + "?tab=0",
				"ServiceUrl": u.cfg.ServiceConfig.Url.UserServiceDomain,
			},
		})
		// 실패시에도 에러는 리턴하지 않고 로그만 남긴다.
		if err != nil {
			u.logger.Entry(ctx).WithFields(logrus.Fields{
				"templateType": domain.NEWS_UPLOAD_TEMPLATE,
				"time":         createdNews.Created.ToString(),
			}).Error("Sending Mail Failed")
		}
	}

	return nil
}

// Update 시에는 메일 발송을 하지 않는다.
func (u newsUsecase) Update(ctx context.Context, request adomain.NewsUpdateRequest) error {

	news, err := u.newsRepository.GetById(ctx, request.Id)
	if err != nil {
		return errors.Wrap(err, "err in newsRepository.GetById")
	}

	// pin 개수가 증가하는 경우 최대 개수 체크
	if lo.FromPtr(request.Pin) == true && news.Pin == false {
		pinnedNewsCount, _, err := u.newsRepository.GetAll(ctx, cdomain.NewsQueryOptionFunc().WithPin(request.Pin))
		if err != nil {
			return errors.Wrap(err, "err in newsRepository.GetAll")
		}

		if pinnedNewsCount >= _maxPinCount {
			return cdomain.ErrRequestLimitExceeded
		}
	}

	// 현재 공지사항에 연관된 첨부파일 조회
	attachments, err := u.attachmentRepository.GetAll(ctx, cdomain.AttachmentQueryOptionFunc().WithResourceId(news.CommonId))
	if err != nil {
		return errors.Wrap(err, "failed to get attachments")
	}

	var (
		attachmentRequestFileNames = u.fileHelper.ExtractFileNames(request.AttachmentUrls)
		attachmentSkipFileNames    = make([]string, 0)
	)

	for _, attachment := range attachments {
		// 이미 존재하고 있는 파일이라면 삭제하지 않습니다.
		if lo.Contains(attachmentRequestFileNames, attachment.FileName) {
			attachmentSkipFileNames = append(attachmentSkipFileNames, attachment.FileName)
			continue
		}

		if err = u.attachmentRepository.Delete(ctx, attachment.Id); err != nil {
			return err
		}
	}
	contentImages, err := u.contentImageRepository.GetAll(ctx, cdomain.ContentImageQueryOptionFunc().WithResourceId(news.CommonId))
	if err != nil {
		return errors.Wrap(err, "failed to get content images")
	}

	var (
		contentImageRequestFileNames = u.fileHelper.ExtractFileNames(request.ContentImageUrls)
		contentImageSkipFileNames    = make([]string, 0)
	)

	for _, contentImage := range contentImages {
		// 이미 존재하고 있는 파일이라면 삭제하지 않습니다.
		if lo.Contains(contentImageRequestFileNames, contentImage.FileName) {
			contentImageSkipFileNames = append(contentImageSkipFileNames, contentImage.FileName)
			continue
		}

		if err = u.contentImageRepository.Delete(ctx, contentImage.Id); err != nil {
			return err
		}
	}

	for _, attachmentUrl := range request.AttachmentUrls {
		filePath, err := u.fileHelper.ExtractFilePath(attachmentUrl)
		if err != nil {
			return errors.Wrap(err, "failed to extract file path")
		}

		normalizedPath := filehelper.NormalizeS3FilePath(u.cfg.AWS.S3_PUBLIC_BUCKET, filePath)
		s3Key := strings.TrimPrefix(normalizedPath, "/")

		fileName := u.fileHelper.ExtractFileName(normalizedPath)

		if lo.Contains(attachmentSkipFileNames, fileName) {
			continue
		}

		// S3 에서 파일 메타정보 가저오기
		header, err := u.s3pkg.GetHeadObject(u.cfg.AWS.S3_PUBLIC_BUCKET, s3Key)
		if err != nil {
			return errors.Wrap(err, "failed to get head object")
		}

		// Attachment Repository 생성
		if _, err = u.attachmentRepository.Create(ctx, domain.Attachment{
			FileName:      fileName,
			OriginalName:  header.OriginalFileName,
			FileSize:      header.FileSize,
			FileExtension: path.Ext(normalizedPath),
			FilePath:      path.Dir(normalizedPath) + "/",
			FileUrl:       s3utils.MakeCDNURL(normalizedPath),
			ResourceId:    news.CommonId,
		}); err != nil {
			return err
		}
	}

	// 컨텐츠 이미지 파일관련 생성
	for _, contentImageUrl := range request.ContentImageUrls {
		filePath, err := u.fileHelper.ExtractFilePath(contentImageUrl)
		if err != nil {
			return errors.Wrap(err, "failed to extract file path")
		}

		normalizedPath := filehelper.NormalizeS3FilePath(u.cfg.AWS.S3_PUBLIC_BUCKET, filePath)
		s3Key := strings.TrimPrefix(normalizedPath, "/")

		fileName := u.fileHelper.ExtractFileName(normalizedPath)

		if lo.Contains(contentImageSkipFileNames, fileName) {
			continue
		}

		// S3 에서 파일 메타정보 가저오기
		header, err := u.s3pkg.GetHeadObject(u.cfg.AWS.S3_PUBLIC_BUCKET, s3Key)
		if err != nil {
			return errors.Wrap(err, "failed to get head object")
		}

		// 하위 호환성: 예전 파일은 OriginalFileName이 없을 수 있으므로 fileName으로 대체
		if header.OriginalFileName == "" {
			header.OriginalFileName = fileName
		}

		// Content Image Repository 생성
		if _, err = u.contentImageRepository.Create(ctx, domain.ContentImage{
			FileName:      fileName,
			OriginalName:  header.OriginalFileName,
			FileSize:      header.FileSize,
			FileExtension: path.Ext(normalizedPath),
			FilePath:      path.Dir(normalizedPath) + "/",
			FileUrl:       s3utils.MakeCDNURL(normalizedPath),
			ResourceId:    news.CommonId,
		}); err != nil {
			return err
		}
	}

	if err = u.newsRepository.Update(ctx, domain.News{
		Id:            request.Id,
		CommonId:      news.CommonId,
		Pin:           lo.FromPtr(request.Pin),
		Title:         request.Title,
		OwnerId:       request.JWTUser.Id,
		UpdateId:      request.JWTUser.Id,
		ContentHtml:   request.ContentHtml,
		Posted:        request.Posted,
		ClosingDate:   lo.FromPtr(request.ClosingDate),
		AttachmentCnt: len(request.AttachmentUrls),
	}); err != nil {
		return err
	}

	// OperatorHistory 등록
	if err = u.operatorHistoryRepository.Create(ctx, domain.OperatorHistory{
		UserId: request.JWTUser.Id,
		Name:   request.JWTUser.Name,
		Email:  request.JWTUser.Email,
		Menu:   "menu.manage_website.manage_news_detail",
		Title:  request.Title,
		Ip:     request.ClientIp,
		State:  domain.CHANGE,
	}); err != nil {
		return err
	}

	return nil
}

func (u newsUsecase) Delete(ctx context.Context, request adomain.NewsDeleteRequest) error {

	news, err := u.newsRepository.GetById(ctx, request.NewsId)
	if err != nil {
		return errors.Wrap(err, "err in newsRepository.GetById")
	}

	if err := u.newsRepository.Delete(ctx, request.NewsId); err != nil {
		return errors.Wrap(err, "err in newsRepository.Delete")
	}

	if err := u.attachmentRepository.DeleteByResourceId(ctx, news.CommonId); err != nil {
		return errors.Wrap(err, "failed to delete attachments")
	}

	// S3의 첨부 파일 삭제는 별도 배치작업으로 진행한다.

	// TODO 컨텐츠 이미지 부분은 기존 히스토리 확인 후 다시 작성

	if err = u.operatorHistoryRepository.Create(ctx, domain.OperatorHistory{
		UserId: request.JWTUser.Id,
		Name:   request.JWTUser.Name,
		Email:  request.JWTUser.Email,
		Menu:   "menu.manage_website.manage_news_detail",
		Ip:     request.ClientIp,
		State:  domain.DELETE,
	}); err != nil {
		return errors.Wrap(err, "err in operatorHistoryRepository.Create")
	}

	return nil
}

// TODO: 추후 repository 패턴으로 구현해야 한다.
func getAllEmailsOfUsersAndSubscribers(ctx context.Context) ([]string, error) {
	userMails, err := domain.User{}.GetNewsLetterEmails(ctx)
	if err != nil {
		return nil, err
	}

	subscriberMails, err := domain.SubscriptionMail{}.GetSubscriptionEmail(ctx)
	if err != nil {
		return nil, err
	}

	// deduplication using samber/lo
	allMails := append(userMails, subscriberMails...)
	uniqueMails := lo.Uniq(allMails)
	return uniqueMails, nil
}
