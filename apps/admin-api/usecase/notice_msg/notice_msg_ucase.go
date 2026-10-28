package notice_msg

import (
	"context"
	"time"

	"github.com/pkg/errors"

	adomain "github.com/illuminarean/energy-share-core-api/domain"
	cdomain "github.com/illuminarean/energy-shares-backend-repo/clean-arch-domain"
	"github.com/illuminarean/energy-shares-backend-repo/domain"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/jsonutils"
)

type noticeMsgUsecase struct {
	noticeMsgRepository         cdomain.NoticeMsgRepository
	userNoticeMappingRepository cdomain.UserNoticeMappingRepository
	projectRepository           cdomain.ProjectRepository
	shareRepository             cdomain.ShareRepository
	userRepository              cdomain.UserRepository
}

func New(
	noticeMsgRepository cdomain.NoticeMsgRepository,
	userNoticeMappingRepository cdomain.UserNoticeMappingRepository,
	projectRepository cdomain.ProjectRepository,
	shareRepository cdomain.ShareRepository,
	userRepository cdomain.UserRepository,
) adomain.NoticeMsgUsecase {
	return &noticeMsgUsecase{
		noticeMsgRepository:         noticeMsgRepository,
		userNoticeMappingRepository: userNoticeMappingRepository,
		projectRepository:           projectRepository,
		shareRepository:             shareRepository,
		userRepository:              userRepository,
	}
}

func (uc noticeMsgUsecase) GetAll(ctx context.Context, request adomain.NoticeMsgGetAllRequest) (adomain.NoticeMsgGetAllResponse, error) {
	count, items, err := uc.noticeMsgRepository.GetAll(ctx,
		cdomain.NoticeMsgQueryOptionFunc().WithTitle(request.Title),
		cdomain.NoticeMsgQueryOptionFunc().WithStartDate(request.StartDate),
		cdomain.NoticeMsgQueryOptionFunc().WithEndDate(request.EndDate),
		cdomain.NoticeMsgQueryOptionFunc().WithSortBy(request.Sortby),
		cdomain.NoticeMsgQueryOptionFunc().WithOrder(request.Order),
		cdomain.NoticeMsgQueryOptionFunc().WithOffset(request.SkipCount),
		cdomain.NoticeMsgQueryOptionFunc().WithLimit(request.MaxResultCount),
	)
	if err != nil {
		return adomain.NoticeMsgGetAllResponse{}, err
	}

	// get user count of each notice msg
	msgIds := make([]int64, 0)
	for _, item := range items {
		msgIds = append(msgIds, item.Id)
	}
	_, msgUserCounts, err := uc.userNoticeMappingRepository.CountByMsgIds(ctx, msgIds)
	if err != nil {
		return adomain.NoticeMsgGetAllResponse{}, err
	}

	// set user count
	for i, item := range items {
		for _, msgUserCount := range msgUserCounts {
			if item.Id == msgUserCount.MsgId {
				items[i].NoticeUserCount = msgUserCount.UserCount
				break
			}
		}
	}

	return adomain.NoticeMsgGetAllResponse{
		TotalCount: count,
		Items:      items,
	}, nil
}

func (uc noticeMsgUsecase) GetOne(ctx context.Context, request adomain.NoticeMsgGetOneRequest) (adomain.NoticeMsgGetOneResponse, error) {
	item, err := uc.noticeMsgRepository.GetById(ctx, request.Id)
	if err != nil {
		return adomain.NoticeMsgGetOneResponse{}, err
	}

	// find users of the notice msg
	var mappings domain.UserNoticeMappings
	_, mappings, err = uc.userNoticeMappingRepository.GetAll(ctx, cdomain.UserNoticeMappingQueryOptionFunc().WithMsgId(item.Id))
	if err != nil {
		return adomain.NoticeMsgGetOneResponse{}, errors.Wrap(err, "failed to find user ids")
	}
	_, users, err := uc.userRepository.GetAll(ctx, cdomain.UserQueryOptionFunc().WithIds(mappings.UserIds()))
	if err != nil {
		return adomain.NoticeMsgGetOneResponse{}, errors.Wrap(err, "failed to find users")
	}
	item.Users = users

	// find investor types or project name of the notice msg
	switch item.NoticeUserType {
	case domain.NOTICE_USER:
		investorTypes := make([]domain.INVESTOR_TYPE, 0)
		investorTypeMap := make(map[domain.INVESTOR_TYPE]bool)
		for _, user := range users {
			if !investorTypeMap[user.InvestorType] { // avoid duplicate investor types
				investorTypeMap[user.InvestorType] = true
				investorTypes = append(investorTypes, user.InvestorType)
			}
		}
		item.UserInvestorTypeList = investorTypes
	case domain.NOTICE_PROJECT:
		project, err := uc.projectRepository.GetById(ctx, item.ProjectId)
		if err != nil {
			return adomain.NoticeMsgGetOneResponse{}, errors.Wrap(err, "failed to find project")
		}
		item.ProjectName = project.Name
	}

	// set user count
	_, msgUserCounts, err := uc.userNoticeMappingRepository.CountByMsgIds(ctx, []int64{item.Id})
	if err != nil {
		return adomain.NoticeMsgGetOneResponse{}, err
	}
	for _, msgUserCount := range msgUserCounts {
		if item.Id == msgUserCount.MsgId {
			item.NoticeUserCount = msgUserCount.UserCount
			break
		}
	}

	return adomain.NoticeMsgGetOneResponse{NoticeMsg: item}, nil
}

func (uc noticeMsgUsecase) Create(ctx context.Context, request adomain.NoticeMsgCreateRequest) error {
	// insert notice msg
	id, err := uc.noticeMsgRepository.Create(ctx, domain.NoticeMsg{
		Title:          request.Title,
		ContentHtml:    request.ContentHtml,
		NoticeDate:     jsonutils.JSONTime(time.Now()),
		NoticeUserType: domain.NOTICE_DIRECT,
	})
	if err != nil {
		return err
	}

	// insert notice msg mapping
	for _, userId := range request.UserIds {
		mapping := domain.UserNoticeMapping{
			MsgId:  id,
			UserId: userId,
		}
		_, err := uc.userNoticeMappingRepository.Create(ctx, mapping)
		if err != nil {
			return err
		}
	}

	return nil
}

func (uc noticeMsgUsecase) CreateByProject(ctx context.Context, request adomain.NoticeMsgCreateByProjectRequest) error {
	// find project
	project, err := uc.projectRepository.GetById(ctx, request.ProjectId)
	if err != nil {
		return err
	}

	// find investors
	investors, err := uc.shareRepository.GetInvestor(ctx,
		cdomain.ShareQueryOptionFunc().WithProjectId(request.ProjectId),
		cdomain.ShareQueryOptionFunc().WithGroupBy("owner_id"),
	)
	if err != nil {
		return err
	}
	if len(investors) == 0 {
		return errors.Wrap(nil, "no investors found")
	}

	// insert notice msg
	id, err := uc.noticeMsgRepository.Create(ctx, domain.NoticeMsg{
		Title:          request.Title,
		ContentHtml:    request.ContentHtml,
		NoticeDate:     jsonutils.JSONTime(time.Now()),
		NoticeUserType: domain.NOTICE_PROJECT,
		ProjectId:      project.Id,
	})
	if err != nil {
		return err
	}

	// insert notice msg mapping
	for _, investor := range investors {
		mapping := domain.UserNoticeMapping{
			MsgId:  id,
			UserId: investor.Id,
		}
		_, err := uc.userNoticeMappingRepository.Create(ctx, mapping)
		if err != nil {
			return err
		}
	}

	return nil
}

func (uc noticeMsgUsecase) CreateByNewsletter(ctx context.Context, request adomain.NoticeMsgCreateByNewsletterRequest) error {
	// find users who agreed to receive newsletter
	count, users, err := uc.userRepository.GetAll(ctx, cdomain.UserQueryOptionFunc().WithNewsletterAgreement(true))
	if err != nil {
		return err
	}
	if count == 0 {
		return errors.Wrap(nil, "no users found")
	}

	// insert notice msg
	id, err := uc.noticeMsgRepository.Create(ctx, domain.NoticeMsg{
		Title:          request.Title,
		ContentHtml:    request.ContentHtml,
		NoticeDate:     jsonutils.JSONTime(time.Now()),
		NoticeUserType: domain.NOTICE_NEWSLETTER,
	})
	if err != nil {
		return err
	}

	// insert notice msg mapping
	for _, user := range users {
		mapping := domain.UserNoticeMapping{
			MsgId:  id,
			UserId: user.Id,
		}
		_, err := uc.userNoticeMappingRepository.Create(ctx, mapping)
		if err != nil {
			return err
		}
	}

	return nil
}

func (uc noticeMsgUsecase) CreateByUserType(ctx context.Context, request adomain.NoticeMsgCreateByUserTypeRequest) error {
	// find users by investor types
	count, users, err := uc.userRepository.GetAll(ctx, cdomain.UserQueryOptionFunc().WithInvestorTypes(request.UserTypes))
	if err != nil {
		return err
	}
	if count == 0 {
		return errors.Wrap(nil, "no users found")
	}

	// insert notice msg
	id, err := uc.noticeMsgRepository.Create(ctx, domain.NoticeMsg{
		Title:          request.Title,
		ContentHtml:    request.ContentHtml,
		NoticeDate:     jsonutils.JSONTime(time.Now()),
		NoticeUserType: domain.NOTICE_USER,
	})
	if err != nil {
		return err
	}

	// insert notice msg mapping
	for _, user := range users {
		mapping := domain.UserNoticeMapping{
			MsgId:  id,
			UserId: user.Id,
		}
		_, err := uc.userNoticeMappingRepository.Create(ctx, mapping)
		if err != nil {
			return err
		}
	}

	return nil
}
