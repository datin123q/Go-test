package faq

import (
	"github.com/illuminarean/energy-shares-backend-repo/domain"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/jsonutils"
)

var (
	mockFaq1 = domain.Faq{
		Id:            1000,
		CommonId:      "bddad349-d0cc-4a1d-8843-431d54fcd2c7",
		Pin:           false,
		Title:         "重要なお知らせです。",
		OwnerId:       1000,
		UpdateId:      1000,
		ContentHtml:   "",
		Posted:        true,
		CategoryType:  domain.FAQ_CATEGORY_ACCOUNT_REGISTRATION,
		ClosingDate:   jsonutils.JSONTime{},
		Created:       jsonutils.JSONTime{},
		Updated:       jsonutils.JSONTime{},
		Deleted:       jsonutils.JSONTime{},
		Attachments:   nil,
		AttachmentCnt: 1,
	}

	mockAttachment1 = domain.Attachment{
		Id:            1000,
		FileName:      "bpbmq9297s524a3jhgn1",
		OriginalName:  "bpbmhsbipt35d0tqvg3f.pdf",
		FileSize:      1398,
		FileExtension: "pdf",
		FilePath:      "/attachment_file/2702602/",
		FileUrl:       "https://dev-energyshares-file-storage.s3.ap-northeast-2.amazonaws.com/attachment_file/2702602/bpbmq9297s524a3jhgn1.pdf",
		ResourceId:    "bddad349-d0cc-4a1d-8843-431d54fcd2c7",
		Created:       jsonutils.JSONTime{},
		Updated:       jsonutils.JSONTime{},
	}
)
