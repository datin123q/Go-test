package policy

import (
	"context"
	"crypto/sha256"

	cdomain "github.com/illuminarean/energy-shares-backend-repo/clean-arch-domain"
	"github.com/illuminarean/energy-shares-backend-repo/domain"
)

type policyService struct {
	adminUserRepository     cdomain.AdminUserRepository
	policyHistoryRepository cdomain.PolicyHistoryRepository
}

func New(policyHistoryRepository cdomain.PolicyHistoryRepository) cdomain.PolicyService {
	return &policyService{
		policyHistoryRepository: policyHistoryRepository,
	}
}

func (svc policyService) CreatePolicyHistory(ctx context.Context, input cdomain.PolicyHistoryInput) error {

	// 内容の変更履歴
	beforeContent := sha256.Sum256([]byte(input.Before.ContentHtml))
	afterContent := sha256.Sum256([]byte(input.After.ContentHtml))
	if beforeContent != afterContent {
		// 内容が変更された場合は履歴を記録
		policyInfoHistory := domain.PolicyInfoHistory{
			PolicyId:        input.After.Id,
			SubstanceType:   string(domain.POLICY_HISTORY_CONTENTS),
			ContentType:     string(domain.POLICY_CHANGED_CONTENTS),
			Name:            input.Modifier.Name,
			Email:           input.Modifier.Email,
			PrevContentHtml: input.Before.ContentHtml,
			CurrContentHtml: input.After.ContentHtml,
		}

		if _, err := svc.policyHistoryRepository.Create(ctx, policyInfoHistory); err != nil {
			return err
		}
	}

	// 添付ファイルの変更履歴
	if isAttachmentChanged(input.Before, input.After) {
		// 内容が変更された場合は履歴を記録
		policyInfoHistory := domain.PolicyInfoHistory{
			PolicyId:      input.After.Id,
			SubstanceType: string(domain.POLICY_HISTORY_CONTENTS),
			ContentType:   string(domain.POLICY_CHANGED_CONTENTS),
			Name:          input.Modifier.Name,
			Email:         input.Modifier.Email,
		}

		// 新しい添付ファイルが存在する場合のみ、現在の内容とURLを設定
		if input.After.Attachment != nil {
			policyInfoHistory.CurrContent = input.After.Attachment.OriginalName
			policyInfoHistory.CurrUrl = input.After.Attachment.FileUrl
		}

		// 以前の添付ファイルが存在する場合のみ、以前の内容とURLを設定
		if input.Before.Attachment != nil {
			policyInfoHistory.PrevContent = input.Before.Attachment.OriginalName
			policyInfoHistory.PrevUrl = input.Before.Attachment.FileUrl
		}

		if _, err := svc.policyHistoryRepository.Create(ctx, policyInfoHistory); err != nil {
			return err
		}
	}

	return nil
}

func isAttachmentChanged(before, after domain.Policy) bool {
	switch {
	case before.Attachment == nil && after.Attachment == nil:
		// 変更なし: 既存の約款と新しい約款に添付ファイルがない場合
		return false
	case before.Attachment == nil || after.Attachment == nil:
		// 変更あり: 既存の約款にのみ添付ファイルがある場合、または新しい約款にのみ添付ファイルがある場合
		return true
	case before.Attachment.FileName != after.Attachment.FileName:
		// 変更あり: 既存の約款と新しい約款のFileNameが異なる場合
		return true
	default:
		return false
	}
}
