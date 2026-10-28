package policy

import (
	"context"
	"path"
	"time"

	adomain "github.com/illuminarean/energy-share-core-api/domain"
	cdomain "github.com/illuminarean/energy-shares-backend-repo/clean-arch-domain"
	"github.com/illuminarean/energy-shares-backend-repo/domain"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/config"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/filehelper"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/jsonutils"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/s3utils"
	"github.com/pkg/errors"
	"github.com/samber/lo"
)

type policyUsecase struct {
	cfg                     *config.Config
	s3pkg                   s3utils.S3
	fileHelper              filehelper.FileHelper
	attachmentRepository    cdomain.AttachmentRepository
	policyRepository        cdomain.PolicyRepository
	policyHistoryRepository cdomain.PolicyHistoryRepository
	policyService           cdomain.PolicyService
}

func New(
	cfg *config.Config,
	s3pkg s3utils.S3,
	fileHelper filehelper.FileHelper,
	attachmentRepository cdomain.AttachmentRepository,
	policyRepository cdomain.PolicyRepository,
	policyHistoryRepository cdomain.PolicyHistoryRepository,
	policyService cdomain.PolicyService,
) adomain.PolicyUsecase {
	return &policyUsecase{
		cfg:                     cfg,
		s3pkg:                   s3pkg,
		fileHelper:              fileHelper,
		attachmentRepository:    attachmentRepository,
		policyRepository:        policyRepository,
		policyHistoryRepository: policyHistoryRepository,
		policyService:           policyService,
	}
}

func (p policyUsecase) Update(ctx context.Context, request adomain.PolicyUpdateRequest) (adomain.PolicyUpdateResponse, error) {

	var (
		beforePolicy  domain.Policy
		afterPolicy   domain.Policy
		newAttachment domain.Attachment
		err           error
	)

	// 수정 전 약관 정보 조회
	beforePolicy, err = p.policyRepository.GetById(ctx, request.Id)
	if err != nil {
		return adomain.PolicyUpdateResponse{}, errors.Wrap(err, "failed to get policy by ID")
	}

	opts := cdomain.AttachmentQueryOptionFunc()
	beforeAttachments, err := p.attachmentRepository.GetAll(ctx,
		opts.WithResourceId(beforePolicy.CommonId),
		opts.WithSortBy("id"),
		opts.WithOrder("desc"),
		opts.WithOffset(0),
		opts.WithLimit(1),
	)
	if err != nil {
		return adomain.PolicyUpdateResponse{}, errors.Wrap(err, "failed to get attachments")
	}

	beforePolicy.Attachment = lo.IfF(len(beforeAttachments) > 0, func() *domain.Attachment {
		return &beforeAttachments[0]
	}).Else(nil)

	// 약관 수정
	afterPolicy = domain.Policy{
		Id:             request.Id,
		PolicyViewType: request.PolicyViewType,
		ContentHtml:    request.ContentHtml,
		ModifierId:     request.JWTUser.Id,
		ModifierEmail:  request.JWTUser.Email,
	}
	if err := p.policyRepository.Update(ctx, afterPolicy); err != nil {
		return adomain.PolicyUpdateResponse{}, errors.Wrap(err, "failed to update policy")
	}

	// 첨부파일 수정
	if len(request.AttachmentUrls) == 0 { // 인풋 첨부파일이 없음
		afterPolicy.Attachment = nil

	} else { // 인풋 첨부파일이 있음
		newFileName := p.fileHelper.ExtractFileName(request.AttachmentUrls[0])
		newFilePath, err := p.fileHelper.ExtractFilePath(request.AttachmentUrls[0])
		if err != nil {
			return adomain.PolicyUpdateResponse{}, errors.Wrap(err, "failed to extract file path")
		}
		normalizedPath := filehelper.NormalizeS3FilePath(p.cfg.AWS.S3_PUBLIC_BUCKET, newFilePath)

		// 기존 첨부파일이 없거나, FileName 이 기존과 다르면 새로운 첨부파일을 저장
		if beforePolicy.Attachment == nil || beforePolicy.Attachment.FileName != newFileName {
			// 기존 첨부파일 삭제
			if beforePolicy.Attachment != nil {
				if err = p.attachmentRepository.Delete(ctx, beforePolicy.Attachment.Id); err != nil {
					return adomain.PolicyUpdateResponse{}, errors.Wrap(err, "failed to delete attachment")
				}
			}

			// 새로운 첨부파일 저장
			header, err := p.s3pkg.GetHeadObject(p.cfg.AWS.S3_PUBLIC_BUCKET, normalizedPath)
			if err != nil {
				return adomain.PolicyUpdateResponse{}, errors.Wrap(err, "failed to get head object")
			}

			newAttachment = domain.Attachment{
				FileName:      newFileName,
				OriginalName:  header.OriginalFileName,
				FileSize:      header.FileSize,
				FileExtension: path.Ext(normalizedPath),
				FilePath:      path.Dir(normalizedPath) + "/",
				FileUrl:       s3utils.MakeCDNURL(normalizedPath),
				ResourceId:    beforePolicy.CommonId,
			}
			if newAttachment.Id, err = p.attachmentRepository.Create(ctx, newAttachment); err != nil {
				return adomain.PolicyUpdateResponse{}, errors.Wrap(err, "failed to create attachment")
			}

			afterPolicy.Attachment = &newAttachment

		} else {
			// 기존 첨부파일 유지
			afterPolicy.Attachment = beforePolicy.Attachment
		}
	}

	// 약관 히스토리 기록
	if err := p.policyService.CreatePolicyHistory(ctx, cdomain.PolicyHistoryInput{
		Before:   beforePolicy,
		After:    afterPolicy,
		Modifier: request.JWTUser,
	}); err != nil {
		return adomain.PolicyUpdateResponse{}, errors.Wrap(err, "failed to create policy history")
	}

	return adomain.PolicyUpdateResponse{
		Id:             afterPolicy.Id,
		PolicyViewType: afterPolicy.PolicyViewType,
		ContentHtml:    afterPolicy.ContentHtml,
		Attachment:     afterPolicy.Attachment,
		ModifierId:     afterPolicy.ModifierId,
		ModifierEmail:  afterPolicy.ModifierEmail,
		Updated:        afterPolicy.Updated,

		CommonId:   beforePolicy.CommonId,
		OwnerId:    beforePolicy.OwnerId,
		PolicyType: beforePolicy.PolicyType,
		Created:    beforePolicy.Created,
	}, nil
}

func (p policyUsecase) GetAll(ctx context.Context, request adomain.PolicyGetAllRequest) (adomain.PolicyGetAllResponse, error) {

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
		return adomain.PolicyGetAllResponse{}, errors.Wrap(err, "failed to get all policies")
	}

	items := lo.Map(policies, func(policy domain.Policy, _ int) adomain.PolicyGetAllItem {
		return adomain.PolicyGetAllItem{
			Id:             policy.Id,
			CommonId:       policy.CommonId,
			OwnerId:        policy.OwnerId,
			ModifierId:     policy.ModifierId,
			PolicyViewType: policy.PolicyViewType,
			PolicyType:     policy.PolicyType,
			Created:        policy.Created,
			Updated:        policy.Updated,
		}
	})

	return adomain.PolicyGetAllResponse{
		TotalCount: totalCount,
		Items:      items,
	}, nil
}

func (p policyUsecase) GetOne(ctx context.Context, request adomain.PolicyGetOneRequest) (adomain.PolicyGetOneResponse, error) {

	policy, err := p.policyRepository.GetById(ctx, request.Id)
	if err != nil {
		return adomain.PolicyGetOneResponse{}, errors.Wrap(err, "failed to get policy by ID")
	}

	opts := cdomain.AttachmentQueryOptionFunc()
	attachments, err := p.attachmentRepository.GetAll(ctx,
		opts.WithResourceId(policy.CommonId),
		opts.WithSortBy("id"),
		opts.WithOrder("desc"),
		opts.WithOffset(0),
		opts.WithLimit(1),
	)
	if err != nil {
		return adomain.PolicyGetOneResponse{}, errors.Wrap(err, "failed to get attachments")
	}

	return adomain.PolicyGetOneResponse{
		Id:             policy.Id,
		CommonId:       policy.CommonId,
		OwnerId:        policy.OwnerId,
		ModifierId:     policy.ModifierId,
		PolicyViewType: policy.PolicyViewType,
		PolicyType:     policy.PolicyType,
		ContentHtml:    policy.ContentHtml,
		ModifierEmail:  policy.ModifierEmail,
		Attachment: lo.IfF(len(attachments) > 0, func() *domain.Attachment {
			return &attachments[0]
		}).Else(nil),
		Created: policy.Created,
		Updated: policy.Updated,
	}, nil
}

func (p policyUsecase) GetAllHistory(ctx context.Context, request adomain.PolicyGetAllHistoryRequest) (adomain.PolicyGetAllHistoryResponse, error) {
	opts := cdomain.PolicyHistoryQueryOptionFunc()
	totalCount, histories, err := p.policyHistoryRepository.GetAll(ctx,
		opts.WithPolicyId(request.Id),
		opts.WithIsPosted(request.IsPosted),
		opts.WithSortBy(request.Sortby),
		opts.WithOrder(request.Order),
		opts.WithOffset(request.SkipCount),
		opts.WithLimit(request.MaxResultCount),
	)
	if err != nil {
		return adomain.PolicyGetAllHistoryResponse{}, errors.Wrap(err, "failed to get all policies")
	}

	items := lo.Map(histories, func(history domain.PolicyInfoHistory, _ int) adomain.PolicyHistoryItem {
		return adomain.PolicyHistoryItem{
			Id:              history.Id,
			PolicyId:        history.PolicyId,
			SubstanceType:   history.SubstanceType,
			ContentType:     history.ContentType,
			Name:            history.Name,
			Email:           history.Email,
			IsPosted:        history.IsPosted,
			Label:           lo.If(history.Label.IsEmpty(), (*string)(nil)).Else(lo.ToPtr(history.Label.ToFormat(time.DateOnly))),
			PrevContent:     history.PrevContent,
			PrevContentHtml: history.PrevContentHtml,
			PrevUrl:         history.PrevUrl,
			CurrContent:     history.CurrContent,
			CurrContentHtml: history.CurrContentHtml,
			CurrUrl:         history.CurrUrl,
			Created:         history.Created,
		}
	})

	return adomain.PolicyGetAllHistoryResponse{
		TotalCount: totalCount,
		Items:      items,
	}, nil
}

func (p policyUsecase) UpdateHistory(ctx context.Context, request adomain.PolicyUpdateHistoryRequest) (adomain.PolicyUpdateHistoryResponse, error) {

	// check label if is_posted is true
	if request.IsPosted {
		// get duplicated policy history
		opts := cdomain.PolicyHistoryQueryOptionFunc()
		totalCount, duplicated, err := p.policyHistoryRepository.GetAll(ctx,
			opts.WithPolicyId(request.Id),
			opts.WithIsPosted(lo.ToPtr(true)),
			opts.WithLabel(request.Label),
		)
		if err != nil {
			return adomain.PolicyUpdateHistoryResponse{}, errors.Wrap(err, "failed to get all policies")
		}

		// if duplicated policy history exists, set is_posted to false
		if totalCount > 0 {
			duplicated[0].IsPosted = false
			if err := p.policyHistoryRepository.Update(ctx, duplicated[0]); err != nil {
				return adomain.PolicyUpdateHistoryResponse{}, errors.Wrap(err, "failed to update policy history")
			}
		}
	}

	// get policy history by historyId
	history, err := p.policyHistoryRepository.GetById(ctx, request.HistoryId)
	if err != nil {
		return adomain.PolicyUpdateHistoryResponse{}, errors.Wrap(err, "failed to get policy history by ID")
	}

	// set policy history data
	// MIR-1384: Labelはxorm:"date"のカラムへ保存するため、専用のパーサーを使う
	label, err := jsonutils.ParseJSONDateJST(time.DateOnly, lo.FromPtr(request.Label))
	if err != nil {
		return adomain.PolicyUpdateHistoryResponse{}, errors.Wrap(err, "failed to parse Label")
	}
	history.Label = label
	history.IsPosted = request.IsPosted

	// update policy history
	if err := p.policyHistoryRepository.Update(ctx, history); err != nil {
		return adomain.PolicyUpdateHistoryResponse{}, errors.Wrap(err, "failed to update policy history")
	}

	return adomain.PolicyUpdateHistoryResponse{
		PolicyHistoryItem: adomain.PolicyHistoryItem{
			Id:              history.Id,
			PolicyId:        history.PolicyId,
			SubstanceType:   history.SubstanceType,
			ContentType:     history.ContentType,
			Name:            history.Name,
			Email:           history.Email,
			IsPosted:        history.IsPosted,
			Label:           lo.If(history.Label.IsEmpty(), (*string)(nil)).Else(lo.ToPtr(history.Label.ToFormat(time.DateOnly))),
			PrevContent:     history.PrevContent,
			PrevContentHtml: history.PrevContentHtml,
			PrevUrl:         history.PrevUrl,
			CurrContent:     history.CurrContent,
			CurrContentHtml: history.CurrContentHtml,
			CurrUrl:         history.CurrUrl,
			Created:         history.Created,
		},
	}, nil
}

func (p policyUsecase) Delete(ctx context.Context, request adomain.PolicyDeleteRequest) error {

	policy, err := p.policyRepository.GetById(ctx, request.Id)
	if err != nil {
		return errors.Wrap(err, "failed to get policy by ID")
	}

	policy.ModifierId = request.JWTUser.Id
	policy.ModifierEmail = request.JWTUser.Email
	if err := p.policyRepository.Update(ctx, policy); err != nil {
		return errors.Wrap(err, "failed to update policy")
	}

	if err := p.policyRepository.Delete(ctx, request.Id); err != nil {
		return errors.Wrap(err, "failed to delete policy")
	}

	return nil
}

func (p policyUsecase) CheckLabelDuplicate(ctx context.Context, request adomain.PolicyCheckLabelDuplicateRequest) (adomain.PolicyCheckLabelDuplicateResponse, error) {

	opts := cdomain.PolicyHistoryQueryOptionFunc()
	totalCount, histories, err := p.policyHistoryRepository.GetAll(ctx,
		opts.WithPolicyId(request.Id),
		opts.WithIsPosted(lo.ToPtr(true)),
		opts.WithLabel(request.Label),
	)
	if err != nil {
		return adomain.PolicyCheckLabelDuplicateResponse{}, errors.Wrap(err, "failed to get all policies")
	}

	if totalCount > 0 {
		return adomain.PolicyCheckLabelDuplicateResponse{
			IsDuplicate: true,
			PolicyHistory: &adomain.PolicyHistoryItem{
				Id:              histories[0].Id,
				PolicyId:        histories[0].PolicyId,
				SubstanceType:   histories[0].SubstanceType,
				ContentType:     histories[0].ContentType,
				Name:            histories[0].Name,
				Email:           histories[0].Email,
				IsPosted:        histories[0].IsPosted,
				Label:           lo.ToPtr(histories[0].Label.ToFormat(time.DateOnly)),
				PrevContent:     histories[0].PrevContent,
				PrevContentHtml: histories[0].PrevContentHtml,
				PrevUrl:         histories[0].PrevUrl,
				CurrContent:     histories[0].CurrContent,
				CurrContentHtml: histories[0].CurrContentHtml,
				CurrUrl:         histories[0].CurrUrl,
				Created:         histories[0].Created,
			},
		}, nil
	}

	return adomain.PolicyCheckLabelDuplicateResponse{
		IsDuplicate:   false,
		PolicyHistory: nil,
	}, nil
}
