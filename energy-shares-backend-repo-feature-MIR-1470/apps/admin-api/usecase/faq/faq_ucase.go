package faq

import (
	"context"
	"path"
	"strings"

	"github.com/illuminarean/energy-shares-backend-repo/pkg/filehelper"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/idgenerator"
	"github.com/samber/lo"

	"github.com/pkg/errors"

	adomain "github.com/illuminarean/energy-share-core-api/domain"
	cdomain "github.com/illuminarean/energy-shares-backend-repo/clean-arch-domain"
	"github.com/illuminarean/energy-shares-backend-repo/domain"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/config"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/s3utils"
)

const (
	maxPinCount = 4
)

type faqUsecase struct {
	cfg                       *config.Config
	idGenerator               idgenerator.IdGenerator
	s3pkg                     s3utils.S3
	fileHelper                filehelper.FileHelper
	faqRepository             cdomain.FaqRepository
	attachmentRepository      cdomain.AttachmentRepository
	contentImageRepository    cdomain.ContentImageRepository
	operatorHistoryRepository cdomain.OperatorHistoryRepository
}

func New(
	cfg *config.Config,
	idGenerator idgenerator.IdGenerator,
	s3pkg s3utils.S3,
	fileHelper filehelper.FileHelper,
	faqRepository cdomain.FaqRepository,
	attachmentRepository cdomain.AttachmentRepository,
	contentImageRepository cdomain.ContentImageRepository,
	operatorHistoryRepository cdomain.OperatorHistoryRepository,
) adomain.FaqUsecase {
	return &faqUsecase{
		cfg:                       cfg,
		idGenerator:               idGenerator,
		s3pkg:                     s3pkg,
		fileHelper:                fileHelper,
		faqRepository:             faqRepository,
		attachmentRepository:      attachmentRepository,
		operatorHistoryRepository: operatorHistoryRepository,
		contentImageRepository:    contentImageRepository,
	}
}

func (f faqUsecase) Create(ctx context.Context, request adomain.FaqCreateRequest) error {
	var commonId = f.idGenerator.GenUuidV4().String()

	// 전체 FAQ 중에서 Pin이 true인 것을 조회
	if lo.FromPtr(request.Pin) == true {
		count, _, err := f.faqRepository.GetAll(ctx, cdomain.FaqQueryOptionFunc().WithPin(request.Pin))
		if err != nil {
			return err
		}

		// Pin이 true인 것이 4개를 초과하면 (생성건을 포함) 에러를 반환
		if count+1 > maxPinCount {
			return cdomain.ErrRequestLimitExceeded
		}
	}

	// 첨부 파일관련 파일 찾기
	for _, attachmentUrl := range request.AttachmentUrls {
		filePath, err := f.fileHelper.ExtractFilePath(attachmentUrl)
		if err != nil {
			return errors.Wrap(err, "failed to extract file path")
		}
		normalizedPath := filehelper.NormalizeS3FilePath(f.cfg.AWS.S3_PUBLIC_BUCKET, filePath)

		// S3 에서 파일 메타정보 가저오기
		header, err := f.s3pkg.GetHeadObject(f.cfg.AWS.S3_PUBLIC_BUCKET, normalizedPath)
		if err != nil {
			return errors.Wrap(err, "failed to get head object")
		}

		// Attachment Repository 생성
		if _, err = f.attachmentRepository.Create(ctx, domain.Attachment{
			FileName:      f.fileHelper.ExtractFileName(normalizedPath),
			OriginalName:  header.OriginalFileName,
			FileSize:      header.FileSize,
			FileExtension: path.Ext(normalizedPath),
			FilePath:      path.Dir(normalizedPath) + "/",
			FileUrl:       s3utils.MakeCDNURL(normalizedPath),
			ResourceId:    commonId,
		}); err != nil {
			return err
		}
	}

	// 컨텐츠 이미지 파일 찾기
	for _, contentImageUrl := range request.ContentImageUrls {
		filePath, err := f.fileHelper.ExtractFilePath(contentImageUrl)
		if err != nil {
			return errors.Wrap(err, "failed to extract file path")
		}
		normalizedPath := filehelper.NormalizeS3FilePath(f.cfg.AWS.S3_PUBLIC_BUCKET, filePath)

		// S3 에서 파일 메타정보 가저오기
		header, err := f.s3pkg.GetHeadObject(f.cfg.AWS.S3_PUBLIC_BUCKET, normalizedPath)
		if err != nil {
			return errors.Wrap(err, "failed to get head object")
		}

		// Content Image Repository 생성
		if _, err = f.contentImageRepository.Create(ctx, domain.ContentImage{
			FileName:      f.fileHelper.ExtractFileName(filePath),
			OriginalName:  header.OriginalFileName,
			FileSize:      header.FileSize,
			FileExtension: path.Ext(normalizedPath),
			FilePath:      path.Dir(normalizedPath) + "/",
			FileUrl:       s3utils.MakeCDNURL(normalizedPath),
			ResourceId:    commonId,
		}); err != nil {
			return err
		}
	}

	if _, err := f.faqRepository.Create(ctx, domain.Faq{
		CommonId:      commonId,
		Pin:           lo.FromPtr(request.Pin),
		Title:         request.Title,
		OwnerId:       request.JWTUser.Id,
		UpdateId:      request.JWTUser.Id,
		ContentHtml:   request.ContentHtml,
		ClosingDate:   lo.FromPtr(request.ClosingDate),
		Posted:        request.Posted,
		CategoryType:  domain.FAQ_CATEGORY_TYPE(request.CategoryType),
		AttachmentCnt: len(request.AttachmentUrls),
	}); err != nil {
		return err
	}

	// OperatorHistory 등록
	if err := f.operatorHistoryRepository.Create(ctx, domain.OperatorHistory{
		UserId: request.JWTUser.Id,
		Name:   request.JWTUser.Name,
		Email:  request.JWTUser.Email,
		Menu:   "menu.manage_website.manage_faq_register",
		Ip:     request.ClientIp,
		State:  domain.REGISTER,
	}); err != nil {
		return err
	}

	return nil
}

func (f faqUsecase) Update(ctx context.Context, request adomain.FaqUpdateRequest) error {
	faq, err := f.faqRepository.GetById(ctx, request.Id)
	if err != nil {
		return errors.Wrap(err, "err in faqRepository.GetById")
	}

	// Pin을 true로 변경할 때 현재 업데이트하려는 Pin이 false 인경우
	if lo.FromPtr(request.Pin) == true && faq.Pin == false {
		count, _, err := f.faqRepository.GetAll(ctx, cdomain.FaqQueryOptionFunc().WithPin(request.Pin))
		if err != nil {
			return err
		}

		// Pin이 true인 것이 최대 핀 개수를 초과하면 (생성건을 포함) 에러를 반환.
		if count >= maxPinCount {
			return cdomain.ErrRequestLimitExceeded
		}
	}

	// 현재 FAQ에 연관된 첨부파일 조회
	attachments, err := f.attachmentRepository.GetAll(ctx, cdomain.AttachmentQueryOptionFunc().WithResourceId(faq.CommonId))
	if err != nil {
		return errors.Wrap(err, "failed to get attachments")
	}

	var (
		attachmentRequestFileNames = f.fileHelper.ExtractFileNames(request.AttachmentUrls)
		attachmentSkipFileNames    = make([]string, 0)
	)

	for _, attachment := range attachments {
		// 이미 존재하고 있는 파일이라면 삭제하지 않습니다.
		if lo.Contains(attachmentRequestFileNames, attachment.FileName) {
			attachmentSkipFileNames = append(attachmentSkipFileNames, attachment.FileName)
			continue
		}

		if err = f.attachmentRepository.Delete(ctx, attachment.Id); err != nil {
			return err
		}
	}

	contentImages, err := f.contentImageRepository.GetAll(ctx, cdomain.ContentImageQueryOptionFunc().WithResourceId(faq.CommonId))
	if err != nil {
		return errors.Wrap(err, "failed to get content images")
	}

	var (
		contentImageRequestFileNames = f.fileHelper.ExtractFileNames(request.ContentImageUrls)
		contentImageSkipFileNames    = make([]string, 0)
	)

	for _, contentImage := range contentImages {
		// 이미 존재하고 있는 파일이라면 삭제하지 않습니다.
		if lo.Contains(contentImageRequestFileNames, contentImage.FileName) {
			contentImageSkipFileNames = append(contentImageSkipFileNames, contentImage.FileName)
			continue
		}

		if err = f.contentImageRepository.Delete(ctx, contentImage.Id); err != nil {
			return err
		}
	}

	for _, attachmentUrl := range request.AttachmentUrls {
		filePath, err := f.fileHelper.ExtractFilePath(attachmentUrl)
		if err != nil {
			return errors.Wrap(err, "failed to extract file path")
		}
		normalizedPath := filehelper.NormalizeS3FilePath(f.cfg.AWS.S3_PUBLIC_BUCKET, filePath)

		fileName := f.fileHelper.ExtractFileName(normalizedPath)

		if lo.Contains(attachmentSkipFileNames, fileName) {
			continue
		}

		// S3 에서 파일 메타정보 가저오기
		header, err := f.s3pkg.GetHeadObject(f.cfg.AWS.S3_PUBLIC_BUCKET, normalizedPath)
		if err != nil {
			return errors.Wrap(err, "failed to get head object")
		}

		// Attachment Repository 생성
		if _, err = f.attachmentRepository.Create(ctx, domain.Attachment{
			FileName:      fileName,
			OriginalName:  header.OriginalFileName,
			FileSize:      header.FileSize,
			FileExtension: path.Ext(normalizedPath),
			FilePath:      path.Dir(normalizedPath) + "/",
			FileUrl:       s3utils.MakeCDNURL(normalizedPath),
			ResourceId:    faq.CommonId,
		}); err != nil {
			return err
		}
	}

	// 컨텐츠 이미지 파일관련 생성
	for _, contentImageUrl := range request.ContentImageUrls {
		filePath, err := f.fileHelper.ExtractFilePath(contentImageUrl)
		if err != nil {
			return errors.Wrap(err, "failed to extract file path")
		}
		normalizedPath := filehelper.NormalizeS3FilePath(f.cfg.AWS.S3_PUBLIC_BUCKET, filePath)

		fileName := f.fileHelper.ExtractFileName(normalizedPath)

		if lo.Contains(contentImageSkipFileNames, fileName) {
			continue
		}

		// S3 에서 파일 메타정보 가저오기
		header, err := f.s3pkg.GetHeadObject(f.cfg.AWS.S3_PUBLIC_BUCKET, normalizedPath)
		if err != nil {
			return errors.Wrap(err, "failed to get head object")
		}

		// Content Image Repository 생성
		if _, err = f.contentImageRepository.Create(ctx, domain.ContentImage{
			FileName:      fileName,
			OriginalName:  header.OriginalFileName,
			FileSize:      header.FileSize,
			FileExtension: path.Ext(normalizedPath),
			FilePath:      path.Dir(normalizedPath) + "/",
			FileUrl:       s3utils.MakeCDNURL(normalizedPath),
			ResourceId:    faq.CommonId,
		}); err != nil {
			return err
		}
	}

	if err = f.faqRepository.Update(ctx, domain.Faq{
		Id:            request.Id,
		CommonId:      faq.CommonId,
		Pin:           lo.FromPtr(request.Pin),
		Title:         request.Title,
		UpdateId:      request.JWTUser.Id,
		ContentHtml:   request.ContentHtml,
		Posted:        request.Posted,
		CategoryType:  domain.FAQ_CATEGORY_TYPE(request.CategoryType),
		ClosingDate:   lo.FromPtr(request.ClosingDate),
		AttachmentCnt: len(request.AttachmentUrls),
	}); err != nil {
		return err
	}

	// OperatorHistory 등록
	if err = f.operatorHistoryRepository.Create(ctx, domain.OperatorHistory{
		UserId: request.JWTUser.Id,
		Name:   request.JWTUser.Name,
		Email:  request.JWTUser.Email,
		Menu:   "menu.manage_website.manage_faq_detail",
		Title:  request.Title,
		Ip:     request.ClientIp,
		State:  domain.CHANGE,
	}); err != nil {
		return err
	}

	return nil
}

func (f faqUsecase) GetAll(ctx context.Context, request adomain.FaqGetAllRequest) (adomain.FaqGetAllResponse, error) {
	count, items, err := f.faqRepository.GetAll(ctx,
		cdomain.FaqQueryOptionFunc().WithTitle(request.Title),
		cdomain.FaqQueryOptionFunc().WithCategoryType(request.CategoryType),
		cdomain.FaqQueryOptionFunc().WithPosted(request.Posted),
		cdomain.FaqQueryOptionFunc().WithStartDate(request.StartDate),
		cdomain.FaqQueryOptionFunc().WithEndDate(request.EndDate),
		cdomain.FaqQueryOptionFunc().WithSortBy(strings.Join([]string{"pin", request.Sortby}, ",")),
		cdomain.FaqQueryOptionFunc().WithOrder(strings.Join([]string{"desc", request.Order}, ",")),
		cdomain.FaqQueryOptionFunc().WithOffset(request.SkipCount),
		cdomain.FaqQueryOptionFunc().WithLimit(request.MaxResultCount),
	)
	if err != nil {
		return adomain.FaqGetAllResponse{}, err
	}

	for i := range items {
		items[i].Attachments = make([]*domain.Attachment, 0)

		attachments, err := f.attachmentRepository.GetAll(ctx,
			cdomain.AttachmentQueryOptionFunc().WithResourceId(items[i].CommonId),
		)
		if err != nil {
			return adomain.FaqGetAllResponse{}, err
		}

		for _, attachment := range attachments {
			items[i].Attachments = append(items[i].Attachments, &attachment)
		}
	}

	if err = f.operatorHistoryRepository.Create(ctx, domain.OperatorHistory{
		UserId: request.JWTUser.Id,
		Name:   request.JWTUser.Name,
		Email:  request.JWTUser.Email,
		Menu:   "menu.manage_website.manage_faq_list",
		Ip:     request.ClientIp,
		State:  domain.LOOKUP,
	}); err != nil {
		return adomain.FaqGetAllResponse{}, err
	}

	return adomain.FaqGetAllResponse{
		TotalCount: count,
		Items:      items,
	}, nil
}

func (f faqUsecase) GetOne(ctx context.Context, request adomain.FaqGetOneRequest) (adomain.FaqGetOneResponse, error) {
	faq, err := f.faqRepository.GetById(ctx, request.Id)
	if err != nil {
		return adomain.FaqGetOneResponse{}, errors.Wrap(err, "err in faqRepository.GetById")
	}

	attachments, err := f.attachmentRepository.GetAll(ctx,
		cdomain.AttachmentQueryOptionFunc().WithResourceId(faq.CommonId),
	)

	for _, attachment := range attachments {
		faq.Attachments = append(faq.Attachments, &attachment)
	}

	if err = f.operatorHistoryRepository.Create(ctx, domain.OperatorHistory{
		UserId: request.JWTUser.Id,
		Name:   request.JWTUser.Name,
		Email:  request.JWTUser.Email,
		Menu:   "menu.manage_website.manage_faq_detail",
		Ip:     request.ClientIp,
		State:  domain.LOOKUP,
	}); err != nil {
		return adomain.FaqGetOneResponse{}, errors.Wrap(err, "err in operatorHistoryRepository.Create")
	}

	return adomain.FaqGetOneResponse{Faq: faq}, nil
}

func (f faqUsecase) Delete(ctx context.Context, request adomain.FaqDeleteRequest) error {
	faq, err := f.faqRepository.GetById(ctx, request.Id)
	if err != nil {
		return errors.Wrap(err, "err in faqRepository.GetById")
	}

	if err = f.faqRepository.Delete(ctx, request.Id); err != nil {
		return errors.Wrap(err, "err in faqRepository.Delete")
	}

	// 첨부파일 삭제
	if err = f.attachmentRepository.DeleteByResourceId(ctx, faq.CommonId); err != nil {
		return errors.Wrap(err, "err in attachmentRepository.DeleteByResourceId")
	}

	// S3의 첨부 파일 삭제는 별도 배치작업으로 진행한다.

	// TODO 컨텐츠 이미지 부분은 기존 히스토리 확인 후 다시 작성

	if err = f.operatorHistoryRepository.Create(ctx, domain.OperatorHistory{
		UserId: request.JWTUser.Id,
		Name:   request.JWTUser.Name,
		Email:  request.JWTUser.Email,
		Menu:   "menu.manage_website.manage_faq_detail",
		Ip:     request.ClientIp,
		Title:  faq.Title,
		State:  domain.DELETE,
	}); err != nil {
		return errors.Wrap(err, "err in operatorHistoryRepository.Create")
	}

	return nil
}
