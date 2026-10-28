package http

import (
	"time"

	"github.com/illuminarean/energy-shares-backend-repo/domain"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/jsonutils"
)

var (
	t1, _          = time.ParseInLocation("2006-01-02 15:04:05", "2021-07-09 17:41:45", time.Local)
	mockNoticeMsg1 = domain.NoticeMsg{
		Id:              1,
		Title:           "test",
		ContentHtml:     "<p>test</p>",
		NoticeDate:      jsonutils.JSONTime(t1),
		NoticeUserCount: 1,
		Created:         jsonutils.JSONTime(t1),
		NoticeUserType:  domain.NOTICE_NEWSLETTER,
		ProjectId:       0,
	}
)
