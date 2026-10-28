package policy

import (
	"context"
	"time"

	cdomain "github.com/illuminarean/energy-shares-backend-repo/clean-arch-domain"
	"github.com/illuminarean/energy-shares-backend-repo/domain"
	udomain "github.com/illuminarean/energyshares-user-api/domain"
	"github.com/pkg/errors"
	"github.com/samber/lo"
)

type policyUsecase struct {
	attachmentRepository    cdomain.AttachmentRepository
	policyRepository        cdomain.PolicyRepository
	policyHistoryRepository cdomain.PolicyHistoryRepository
}

func New(
	attachmentRepository cdomain.AttachmentRepository,
	policyRepository cdomain.PolicyRepository,
	policyHistoryRepository cdomain.PolicyHistoryRepository,
) udomain.PolicyUsecase {
	return &policyUsecase{
		attachmentRepository:    attachmentRepository,
		policyRepository:        policyRepository,
		policyHistoryRepository: policyHistoryRepository,
	}
}

func (p policyUsecase) GetAll(ctx context.Context, request udomain.PolicyGetAllRequest) (udomain.PolicyGetAllResponse, error) {

	// get all policies
	opts := cdomain.PolicyQueryOptionFunc()
	totalCount, policies, err := p.policyRepository.GetAll(ctx,
		opts.WithSortBy(request.Sortby),
		opts.WithOrder(request.Order),
		opts.WithOffset(request.SkipCount),
		opts.WithLimit(request.MaxResultCount),
		opts.WithPolicyType(request.PolicyType),
		opts.WithPolicyViewType(request.PolicyViewType),
	)
	if err != nil {
		return udomain.PolicyGetAllResponse{}, errors.Wrap(err, "failed to get all policies")
	}

	// get attachment
	attachmentOpts := cdomain.AttachmentQueryOptionFunc()
	for i := range policies {
		policy := &policies[i] // スライス要素をポインタとして取得
		attachments, err := p.attachmentRepository.GetAll(ctx,
			attachmentOpts.WithResourceId(policy.CommonId),
			attachmentOpts.WithSortBy("id"),
			attachmentOpts.WithOrder("desc"),
			attachmentOpts.WithOffset(0),
			attachmentOpts.WithLimit(1),
		)
		if err != nil {
			return udomain.PolicyGetAllResponse{}, errors.Wrap(err, "failed to get attachments")
		}

		policy.Attachment = lo.IfF(len(attachments) > 0, func() *domain.Attachment {
			return &attachments[0]
		}).Else(nil)
	}

	// map to response
	items := lo.Map(policies, func(policy domain.Policy, _ int) udomain.PolicyGetAllItem {
		return udomain.PolicyGetAllItem{
			Id:             policy.Id,
			PolicyViewType: policy.PolicyViewType,
			PolicyType:     policy.PolicyType,
			ContentHtml:    policy.ContentHtml,
			Attachment:     policy.Attachment,
			Created:        policy.Created,
			Updated:        policy.Updated,
		}
	})

	return udomain.PolicyGetAllResponse{
		TotalCount: totalCount,
		Items:      items,
	}, nil
}

func (p policyUsecase) GetAllHistory(ctx context.Context, request udomain.PolicyGetAllHistoryRequest) (udomain.PolicyGetAllHistoryResponse, error) {

	// get latest policy by PolicyType and PolicyViewType
	opts := cdomain.PolicyQueryOptionFunc()
	_, policies, err := p.policyRepository.GetAll(ctx,
		opts.WithSortBy("id"),
		opts.WithOrder("desc"),
		opts.WithOffset(0),
		opts.WithLimit(1),
		opts.WithPolicyType(request.PolicyType),
		opts.WithPolicyViewType(request.PolicyViewType),
	)
	if err != nil || len(policies) == 0 {
		return udomain.PolicyGetAllHistoryResponse{}, errors.Wrap(err, "failed to get all policies")
	}

	// get histories by PolicyId
	historyOpts := cdomain.PolicyHistoryQueryOptionFunc()
	totalCount, histories, err := p.policyHistoryRepository.GetAll(ctx,
		historyOpts.WithPolicyId(policies[0].Id),
		historyOpts.WithIsPosted(lo.ToPtr(true)),
		historyOpts.WithSortBy(request.Sortby),
		historyOpts.WithOrder(request.Order),
		historyOpts.WithOffset(request.SkipCount),
		historyOpts.WithLimit(request.MaxResultCount),
	)
	if err != nil {
		return udomain.PolicyGetAllHistoryResponse{}, errors.Wrap(err, "failed to get all policies")
	}

	items := lo.Map(histories, func(history domain.PolicyInfoHistory, _ int) udomain.PolicyGetAllHistoryItem {
		return udomain.PolicyGetAllHistoryItem{
			Id:             history.Id,
			PolicyType:     domain.POLICY_TYPE(lo.FromPtr(request.PolicyType)),
			PolicyViewType: domain.POLICY_VIEW_TYPE(lo.FromPtr(request.PolicyViewType)),
			ContentHtml:    history.CurrContentHtml,
			IsPosted:       history.IsPosted,
			Label:          lo.If(history.Label.IsEmpty(), (*string)(nil)).Else(lo.ToPtr(history.Label.ToFormat(time.DateOnly))),
			Created:        history.Created,
		}
	})

	return udomain.PolicyGetAllHistoryResponse{
		TotalCount: totalCount,
		Items:      items,
	}, nil
}
