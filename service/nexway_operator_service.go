package service

import (
	"context"
	"fmt"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/illuminarean/energy-shares-backend-repo/domain"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/jsonutils"
)

// SaveOperators Nexway API から取得したオペレーター情報をDBに保存
func SaveOperators(ctx context.Context, ekycRequestLogId int64, operators []domain.NexwayOperator) error {
	for _, op := range operators {
		// OperationEndAtをパース (RFC3339形式: 2024-01-27T12:34:56+09:00)
		operationEndAt, err := time.Parse(time.RFC3339, op.OperationEndAt)
		if err != nil {
			logrus.WithFields(logrus.Fields{
				"error":            err,
				"operation_end_at": op.OperationEndAt,
			}).Error("Failed to parse operation_end_at")
			return fmt.Errorf("failed to parse operation_end_at: %w", err)
		}

		nexwayOp := &domain.NexwayBpoOperator{
			EkycRequestLogId: ekycRequestLogId,
			CompanyId:        op.CompanyId,
			CompanyName:      op.CompanyName,
			UserId:           op.UserId,
			UserName:         op.UserName,
			LoginId:          op.LoginId,
			OperatorType:     op.OperatorType,
			OperationEndAt:   jsonutils.JSONTime(operationEndAt),
		}

		if _, err := nexwayOp.Create(ctx); err != nil {
			logrus.WithFields(logrus.Fields{
				"error":               err,
				"ekyc_request_log_id": ekycRequestLogId,
				"operator_type":       op.OperatorType,
			}).Error("Failed to save operator")
			return fmt.Errorf("failed to save operator to DB: %w", err)
		}

		logrus.WithFields(logrus.Fields{
			"ekyc_request_log_id": ekycRequestLogId,
			"operator_type":       op.OperatorType,
			"user_name":           op.UserName,
		}).Info("Saved operator")
	}

	return nil
}
