package domain

import (
	"context"
	"encoding/json"
	"time"

	"github.com/samber/lo"

	"github.com/illuminarean/energy-shares-backend-repo/pkg/emailutils"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/factory"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/jsonutils"
)

type MailRetry struct {
	Id           int64              `xorm:"pk autoincr 'id'"`                     // Primary key
	EmailRequest json.RawMessage    `xorm:"json 'email_request'"`                 // JSON object
	Receivers    json.RawMessage    `xorm:"json 'receivers'"`                     // JSON array
	TemplateName string             `xorm:"varchar(255) notnull 'template_name'"` // Template name
	RetryStatus  string             `xorm:"varchar(50) notnull 'retry_status'"`   // PENDING, IN_PROGRESS, FAILED
	Created      jsonutils.JSONTime `xorm:"created 'created'"`                    // Timestamp (auto-filled)
}

type MailFailed struct {
	Id           int64              `xorm:"pk autoincr 'id'"`                     // Primary key
	EmailRequest json.RawMessage    `xorm:"json 'email_request'"`                 // JSON object
	Receivers    json.RawMessage    `xorm:"json 'receivers'"`                     // JSON array
	TemplateName string             `xorm:"varchar(255) notnull 'template_name'"` // Template name
	RetryStatus  string             `xorm:"varchar(50) notnull 'retry_status'"`   // PENDING, IN_PROGRESS, FAILED
	Created      jsonutils.JSONTime `xorm:"created 'created'"`                    // Timestamp (auto-filled)
}

func (m MailRetry) TableName() string {
	return "mail_retry"
}

func (m MailFailed) TableName() string {
	return "mail_failed"
}

func (m *MailRetry) ToMailFailed() MailFailed {
	return MailFailed{
		EmailRequest: m.EmailRequest,
		Receivers:    m.Receivers,
		TemplateName: m.TemplateName,
		RetryStatus:  m.RetryStatus,
	}
}

func (m *MailRetry) ToEmailRequest() (emailutils.EmailRequest, error) {
	var emailRequest emailutils.EmailRequest
	if err := json.Unmarshal(m.EmailRequest, &emailRequest); err != nil {
		return emailutils.EmailRequest{}, err
	}
	return emailRequest, nil
}

func (m *MailRetry) EmailRequestToMailRetry(emailRequest emailutils.EmailRequest) error {
	// EmailRequestをJSONでシリアライズ
	emailRequestBytes, err := json.Marshal(emailRequest)
	if err != nil {
		return err
	}

	// すべてのMailDataのReceiversを1つのスライスにマージ
	var allReceivers []string
	for _, mailData := range emailRequest.MailData {
		allReceivers = append(allReceivers, mailData.Receivers...)
	}

	// ReceiversスライスをインしJSON with serialize
	receiversBytes, err := json.Marshal(allReceivers)
	if err != nil {
		return err
	}

	m.EmailRequest = emailRequestBytes
	m.Receivers = receiversBytes
	m.TemplateName = emailRequest.TemplateName
	m.RetryStatus = "PENDING" // デフォルト値として設定
	m.Created = jsonutils.JSONTime(time.Now())

	return nil
}

func (m *MailRetry) Create(ctx context.Context) (int64, error) {
	return factory.DB(ctx).Insert(m)
}

// GetAllByStatusは特定の状態のすべてのレコードを取得するメソッド。最大1000件のレコードを取得できます
const _MaxMailRetry = 1000

func (MailRetry) GetAllByStatus(ctx context.Context, retryStatus string) ([]MailRetry, error) {
	var records []MailRetry
	err := factory.DB(ctx).
		Where("retry_status = ?", retryStatus).
		Limit(_MaxMailRetry).
		Find(&records)
	return records, err
}

// UpdateStatusManyは複数のレコードの状態を一度に更新するメソッド
func (MailRetry) UpdateStatusMany(ctx context.Context, retries []MailRetry, newStatus string) (err error) {

	// Extract the IDs from the retries slice
	ids := lo.Map(retries, func(retry MailRetry, _ int) int64 {
		return retry.Id
	})

	// Update all records with a single query
	_, err = factory.DB(ctx).In("id", ids).Update(&MailRetry{RetryStatus: newStatus})
	if err != nil {
		return err
	}
	return nil
}

// 特定の状態のすべてのレコードを削除するメソッド
func (MailRetry) DeleteMany(ctx context.Context, retries []MailRetry) error {
	// Extract the IDs from the retries slice
	ids := lo.Map(retries, func(retry MailRetry, _ int) int64 {
		return retry.Id
	})

	_, err := factory.DB(ctx).In("id", ids).Delete(&MailRetry{})
	if err != nil {
		return err
	}
	return nil
}

func (m *MailFailed) Create(ctx context.Context) (int64, error) {
	return factory.DB(ctx).Insert(m)
}
