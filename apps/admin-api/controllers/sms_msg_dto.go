package controllers

import "github.com/illuminarean/energy-shares-backend-repo/domain"

type SmsSearchInput struct {
	Content   *string `query:"sms_content"`
	Name      *string `query:"name"`
	Sender    *string `query:"sender"`
	Receiver  *string `query:"receiver"`
	Status    *string `query:"status"`
	StartDate string  `query:"start_date"`
	EndDate   string  `query:"end_date"`
	SearchInput
}

type SmsInput struct {
	Title           string   `json:"title" swagger:"desc(題目)"`
	ReceiverNumbers []string `json:"receiver_numbers" swagger:"desc(受信番号)"`
	Content         string   `json:"content" swagger:"desc(内容)"`
}

func (m *SmsInput) ToModel() (*domain.SmsMsg, error) {

	return &domain.SmsMsg{
		Title:      m.Title,
		Receiver:   m.ReceiverNumbers,
		Status:     domain.REQUEST_SMS,
		SmsType:    domain.NORMAL_SMS,
		SmsContent: m.Content,
		Count:      int64(len(m.ReceiverNumbers)),
	}, nil
}
