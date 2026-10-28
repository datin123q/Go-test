package notice_msg

import (
	"context"

	cdomain "github.com/illuminarean/energy-shares-backend-repo/clean-arch-domain"
	"github.com/illuminarean/energy-shares-backend-repo/domain"
	udomain "github.com/illuminarean/energyshares-user-api/domain"
)

type noticeMsgUsecase struct {
	noticeMsgRepository         cdomain.NoticeMsgRepository
	noticeMsgReadRepository     cdomain.NoticeMsgReadRepository
	userNoticeMappingRepository cdomain.UserNoticeMappingRepository
}

func New(
	noticeMsgRepository cdomain.NoticeMsgRepository,
	noticeMsgReadRepository cdomain.NoticeMsgReadRepository,
	userNoticeMappingRepository cdomain.UserNoticeMappingRepository,
) udomain.NoticeMsgUsecase {
	return &noticeMsgUsecase{
		noticeMsgRepository:         noticeMsgRepository,
		noticeMsgReadRepository:     noticeMsgReadRepository,
		userNoticeMappingRepository: userNoticeMappingRepository,
	}
}

func (uc noticeMsgUsecase) GetAll(ctx context.Context, request udomain.NoticeMsgGetAllRequest) (udomain.NoticeMsgGetAllResponse, error) {
	// get mappings
	var mappings domain.UserNoticeMappings
	_, mappings, err := uc.userNoticeMappingRepository.GetAll(ctx,
		cdomain.UserNoticeMappingQueryOptionFunc().WithUserId(request.JWTUser.Id),
	)
	if err != nil {
		return udomain.NoticeMsgGetAllResponse{}, err
	}

	// get notice msgs
	count, items, err := uc.noticeMsgRepository.GetAll(ctx,
		cdomain.NoticeMsgQueryOptionFunc().WithMsgIds(mappings.MsgIds()),
		cdomain.NoticeMsgQueryOptionFunc().WithTitle(request.Title),
		cdomain.NoticeMsgQueryOptionFunc().WithStartDate(request.StartDate),
		cdomain.NoticeMsgQueryOptionFunc().WithEndDate(request.EndDate),
		cdomain.NoticeMsgQueryOptionFunc().WithSortBy(request.Sortby),
		cdomain.NoticeMsgQueryOptionFunc().WithOrder(request.Order),
		cdomain.NoticeMsgQueryOptionFunc().WithOffset(request.SkipCount),
		cdomain.NoticeMsgQueryOptionFunc().WithLimit(request.MaxResultCount),
	)
	if err != nil {
		return udomain.NoticeMsgGetAllResponse{}, err
	}

	// get user count of each notice msg
	msgIds := make([]int64, 0)
	for _, item := range items {
		msgIds = append(msgIds, item.Id)
	}
	_, msgUserCounts, err := uc.userNoticeMappingRepository.CountByMsgIds(ctx, msgIds)
	if err != nil {
		return udomain.NoticeMsgGetAllResponse{}, err
	}

	// get read status
	_, noticeMsgReads, err := uc.noticeMsgReadRepository.GetAll(ctx,
		cdomain.NoticeMsgReadQueryOptionFunc().WithUserId(request.JWTUser.Id),
	)
	if err != nil {
		return udomain.NoticeMsgGetAllResponse{}, err
	}

	for i, noticeMsg := range items {
		// set user count
		for _, msgUserCount := range msgUserCounts {
			if noticeMsg.Id == msgUserCount.MsgId {
				items[i].NoticeUserCount = msgUserCount.UserCount
				break
			}
		}
		// set isRead flag
		for _, noticeMsgRead := range noticeMsgReads {
			if noticeMsg.Id == noticeMsgRead.MsgId {
				items[i].IsRead = noticeMsgRead.IsRead
				break
			}
		}
	}

	return udomain.NoticeMsgGetAllResponse{
		TotalCount: count,
		Items:      items,
	}, nil
}

func (uc noticeMsgUsecase) GetOne(ctx context.Context, request udomain.NoticeMsgGetOneRequest) (udomain.NoticeMsgGetOneResponse, error) {
	// get mapping
	count, _, err := uc.userNoticeMappingRepository.GetAll(ctx,
		cdomain.UserNoticeMappingQueryOptionFunc().WithMsgId(request.Id),
		cdomain.UserNoticeMappingQueryOptionFunc().WithUserId(request.JWTUser.Id),
	)
	if err != nil {
		return udomain.NoticeMsgGetOneResponse{}, err
	}
	if count == 0 {
		return udomain.NoticeMsgGetOneResponse{}, cdomain.ErrItemNotFound
	}

	// get notice msg
	item, err := uc.noticeMsgRepository.GetById(ctx, request.Id)
	if err != nil {
		return udomain.NoticeMsgGetOneResponse{}, err
	}

	// set user count
	_, msgUserCounts, err := uc.userNoticeMappingRepository.CountByMsgIds(ctx, []int64{item.Id})
	if err != nil {
		return udomain.NoticeMsgGetOneResponse{}, err
	}
	for _, msgUserCount := range msgUserCounts {
		if item.Id == msgUserCount.MsgId {
			item.NoticeUserCount = msgUserCount.UserCount
			break
		}
	}

	// set isRead flag
	err = uc.setNoticeMsgRead(ctx, request)
	if err != nil {
		return udomain.NoticeMsgGetOneResponse{}, err
	}
	item.IsRead = true

	return udomain.NoticeMsgGetOneResponse{NoticeMsg: item}, nil
}

func (uc noticeMsgUsecase) setNoticeMsgRead(ctx context.Context, request udomain.NoticeMsgGetOneRequest) error {
	has, _, err := uc.noticeMsgReadRepository.GetAll(ctx,
		cdomain.NoticeMsgReadQueryOptionFunc().WithMsgId(request.Id),
		cdomain.NoticeMsgReadQueryOptionFunc().WithUserId(request.JWTUser.Id),
	)
	if err != nil {
		return err
	}

	if has == 0 {
		_, err = uc.noticeMsgReadRepository.Create(ctx, domain.NoticeMsgRead{
			UserId: request.JWTUser.Id,
			MsgId:  request.Id,
			IsRead: true,
		})
		if err != nil {
			return err
		}
	} else {
		err = uc.noticeMsgReadRepository.Update(ctx, domain.NoticeMsgRead{
			UserId: request.JWTUser.Id,
			MsgId:  request.Id,
			IsRead: true,
		})
		if err != nil {
			return err
		}
	}

	return nil
}
