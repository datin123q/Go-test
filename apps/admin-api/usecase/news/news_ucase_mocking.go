package news

import (
	"github.com/illuminarean/energy-shares-backend-repo/domain"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/jsonutils"
)

var (
	mockNews1 = domain.News{
		Id:            1000,
		CommonId:      "bddad349-d0cc-4a1d-8843-431d54fcd2c7",
		Pin:           false,
		Title:         "重要なお知らせです。",
		OwnerId:       12378123,
		UpdateId:      12378123,
		ContentHtml:   "",
		Posted:        true,
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
		FileExtension: ".pdf",
		FilePath:      "/attachment_file/2702602/",
		FileUrl:       "https://dev-energyshares-file-storage.s3.ap-northeast-2.amazonaws.com/attachment_file/2702602/bpbmq9297s524a3jhgn1.pdf",
		ResourceId:    "bddad349-d0cc-4a1d-8843-431d54fcd2c7",
		Created:       jsonutils.JSONTime{},
		Updated:       jsonutils.JSONTime{},
	}

	mockContentImage1 = domain.ContentImage{
		Id:            1000,
		FileName:      "bpbmq9297s524a3jhgn1",
		OriginalName:  "bpbmhsbipt35d0tqvg3f.jpg",
		FileSize:      1398,
		FileExtension: ".jpg",
		FilePath:      "/content_image/2702602/",
		FileUrl:       "https://dev-energyshares-file-storage.s3.ap-northeast-2.amazonaws.com/attachment_file/2702602/bpbmq9297s524a3jhgn1.pdf",
		ResourceId:    "bddad349-d0cc-4a1d-8843-431d54fcd2c7",
		Created:       jsonutils.JSONTime{},
		Updated:       jsonutils.JSONTime{},
	}
)
