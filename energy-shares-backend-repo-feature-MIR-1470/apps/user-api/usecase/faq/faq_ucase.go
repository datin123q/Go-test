package faq

import (
	"context"
	cdomain "github.com/illuminarean/energy-shares-backend-repo/clean-arch-domain"
	"github.com/illuminarean/energy-shares-backend-repo/domain"
	udomain "github.com/illuminarean/energyshares-user-api/domain"
	"github.com/samber/lo"
	"strings"
)

type faqUsecase struct {
	faqRepository        cdomain.FaqRepository
	attachmentRepository cdomain.AttachmentRepository
}

func New(faqRepository cdomain.FaqRepository, attachmentRepository cdomain.AttachmentRepository) udomain.FaqUsecase {
	return &faqUsecase{
		faqRepository:        faqRepository,
		attachmentRepository: attachmentRepository,
	}
}

func (f faqUsecase) GetAll(ctx context.Context, request udomain.FaqGetAllRequest) (udomain.FaqGetAllResponse, error) {
	count, items, err := f.faqRepository.GetAll(ctx,
		cdomain.FaqQueryOptionFunc().WithTitle(request.Title),
		cdomain.FaqQueryOptionFunc().WithCategoryType(request.CategoryType),
		cdomain.FaqQueryOptionFunc().WithPosted(lo.ToPtr(true)),
		cdomain.FaqQueryOptionFunc().WithSortBy(strings.Join([]string{"pin", request.Sortby}, ",")),
		cdomain.FaqQueryOptionFunc().WithOrder(strings.Join([]string{"desc", request.Order}, ",")),
		cdomain.FaqQueryOptionFunc().WithOffset(request.SkipCount),
		cdomain.FaqQueryOptionFunc().WithLimit(request.MaxResultCount),
	)
	if err != nil {
		return udomain.FaqGetAllResponse{}, err
	}

	for i := range items {
		items[i].Attachments = make([]*domain.Attachment, 0)

		attachments, err := f.attachmentRepository.GetAll(ctx,
			cdomain.AttachmentQueryOptionFunc().WithResourceId(items[i].CommonId),
		)
		if err != nil {
			return udomain.FaqGetAllResponse{}, err
		}

		for _, attachment := range attachments {
			items[i].Attachments = append(items[i].Attachments, &attachment)
		}
	}

	return udomain.FaqGetAllResponse{
		TotalCount: count,
		Items:      items,
	}, nil
}
