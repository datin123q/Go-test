package http

import (
	"github.com/illuminarean/energy-shares-backend-repo/domain"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/jsonutils"
)

var (
	mockNews1 = domain.News{
		Id:            1000,
		CommonId:      "2000",
		Pin:           true,
		Title:         "News Title1",
		OwnerId:       1000,
		UpdateId:      1000,
		ContentHtml:   "",
		Posted:        true,
		ClosingDate:   jsonutils.JSONTime{},
		Created:       jsonutils.JSONTime{},
		Updated:       jsonutils.JSONTime{},
		Deleted:       jsonutils.JSONTime{},
		Attachments:   nil,
		AttachmentCnt: 0,
	}
)
