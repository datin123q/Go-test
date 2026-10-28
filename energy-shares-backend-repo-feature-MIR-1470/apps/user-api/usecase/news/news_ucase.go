package news

import (
	"context"

	"github.com/pkg/errors"

	cdomain "github.com/illuminarean/energy-shares-backend-repo/clean-arch-domain"
	"github.com/illuminarean/energy-shares-backend-repo/domain"
	udomain "github.com/illuminarean/energyshares-user-api/domain"
)

type newsUsecase struct {
	newsRepository            cdomain.NewsRepository
	attachmentRepository      cdomain.AttachmentRepository
	operatorHistoryRepository cdomain.OperatorHistoryRepository
}

func New(
	newsRepository cdomain.NewsRepository,
	attachmentRepository cdomain.AttachmentRepository,
	operatorHistoryRepository cdomain.OperatorHistoryRepository,
) udomain.NewsUsecase {
	return &newsUsecase{
		newsRepository:            newsRepository,
		attachmentRepository:      attachmentRepository,
		operatorHistoryRepository: operatorHistoryRepository,
	}
}

func (u newsUsecase) GetAll(ctx context.Context, request udomain.NewsGetAllRequest) (udomain.NewsGetAllResponse, error) {

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
		return udomain.NewsGetAllResponse{}, err
	}

	for i := range items {
		if items[i].AttachmentCnt == 0 {
			items[i].Attachments = make([]*domain.Attachment, 0)
			continue
		}
		attachments, err := u.attachmentRepository.GetAll(ctx,
			cdomain.AttachmentQueryOptionFunc().WithResourceId(items[i].CommonId),
		)
		if err != nil {
			return udomain.NewsGetAllResponse{}, err
		}

		items[i].AddAttachments(attachments)
	}

	if err = u.operatorHistoryRepository.Create(ctx, domain.OperatorHistory{
		UserId: request.JWTUser.Id,
		Name:   request.JWTUser.Name,
		Email:  request.JWTUser.Email,
		Menu:   "menu.manage_website.manage_news_list",
		Ip:     request.ClientIp,
		State:  domain.LOOKUP,
	}); err != nil {
		return udomain.NewsGetAllResponse{}, errors.Wrap(err, "failed to create operator history")
	}

	return udomain.NewsGetAllResponse{
		TotalCount: total,
		Items:      items,
	}, nil
}
