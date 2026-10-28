package domain

import (
	"github.com/samber/lo"

	"github.com/illuminarean/energy-shares-backend-repo/pkg/jsonutils"
)

type News struct {
	Id            int64              `json:"id" description:"お知らせID" example:"61"`
	CommonId      string             `json:"common_id" xorm:"notnull UUID" description:"共通ID" xorm:"notnull UUID" swagger:"desc(共通ID)" example:"aeee0ebc-cb8a-43bd-9eb6-94c34b8686ae"`
	Pin           bool               `json:"pin" xorm:"not null" description:"ピン固定" example:"false"`
	Title         string             `json:"title" xorm:"varchar(255) not null" description:"タイトル" example:"お知らせテストです"`
	OwnerId       int64              `json:"owner_id" xorm:"index" description:"作成者" example:"110511"`
	UpdateId      int64              `json:"update_id" xorm:"index" description:"最終更新者" example:"110511"`
	ContentHtml   string             `json:"content_html" xorm:"LONGTEXT not null" description:"コンテンツHTML" example:"<p>こんにちは</p>\r\n<p><img title=\"Untitled.png\" src=\"https://cdn.dev.energyshares.jp/temp_file/2024-07-26/cqhkgsanulri1lp6i7eg.png\" alt=\"\" width=\"878\" height=\"228\" /></p>\r\n<p>FAQテストです。</p>\r\n<p>ありがとうございます。</p>"`
	Posted        bool               `json:"posted" description:"掲示有無" example:"true"`
	ClosingDate   jsonutils.JSONTime `json:"closing_date" xorm:"not null" description:"締切日" example:"2024-07-30 09:18:59"`
	Created       jsonutils.JSONTime `json:"created" xorm:"created" description:"作成日" example:"2024-07-26 15:56:32"`
	Updated       jsonutils.JSONTime `json:"updated" xorm:"updated" description:"更新日" example:"2024-07-26 15:56:32"`
	Deleted       jsonutils.JSONTime `json:"deleted" xorm:"deleted" description:"削除日" example:""`
	Attachments   []*Attachment      `json:"attachments" xorm:"-" description:"添付ファイルリスト"`
	AttachmentCnt int                `json:"attachment_cnt" description:"添付ファイル個数" example:"5"`
}

func (News) TableName() string {
	return "news"
}

func (n *News) AddAttachments(attachments []Attachment) {
	n.Attachments = lo.Map(attachments, func(a Attachment, _ int) *Attachment {
		return &a
	})
	return
}

func (n News) WithAttachments(attachments []Attachment) News {
	for _, attachment := range attachments {
		n.Attachments = append(n.Attachments, &attachment)
	}

	return n
}
