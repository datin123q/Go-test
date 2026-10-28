package domain

import (
	"context"

	"github.com/pkg/errors"

	"github.com/illuminarean/energy-shares-backend-repo/pkg/commoncode"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/factory"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/i18nutils"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/jsonutils"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/modelhelper"
)

type NoticeMsg struct { // ユーザーニュース
	Id                   int64              `json:"id" swagger:"desc(ID)"`
	Title                string             `json:"title" swagger:"desc(제목)"`
	ContentHtml          string             `json:"content_html" xorm:"LONGTEXT" swagger:"desc(내용)"`
	NoticeDate           jsonutils.JSONTime `json:"notice_date" swagger:"desc(소식 전송일자)"`
	NoticeUserType       NOTICE_USER_TYPE   `json:"notice_user_type" swagger:"desc(소식 수신 유저 정보)"`
	IsRead               bool               `json:"is_read" xorm:"-" swagger:"desc(소식 읽음 여부)"`
	ProjectId            int64              `json:"project_id" swagger:"desc(투자상품별 회원시 상품 정보 저장)"`
	ProjectName          string             `json:"project_name" swagger:"desc(투자상품별 회원시 투자상품 이름)" xorm:"-"`
	NoticeUserCount      int64              `json:"notice_user_count" swagger:"desc(소식 전송 유저 카운터)" xorm:"-"`
	Users                []User             `json:"users" swagger:"desc(수신 사용자 목록)" xorm:"-"`
	UserInvestorTypeList []INVESTOR_TYPE    `json:"user_investor_type_list" swagger:"desc(수신 사용자 회원 투자자 타입 리스트)" xorm:"-"`
	Created              jsonutils.JSONTime `json:"-"`
}

const (
	NOTICE_LIST = "notice_msg.id,notice_msg.title,notice_msg.content_html,notice_msg.notice_date,count(user_notice_mapping.msg_id) AS notice_user_count"
)

type NOTICE_USER_TYPE string

const (
	NOTICE_USER_TYPE_CODE                  = "NOTICE_USER_TYPE"
	NOTICE_PROJECT        NOTICE_USER_TYPE = "NOTICE_PROJECT"
	NOTICE_DIRECT         NOTICE_USER_TYPE = "NOTICE_DIRECT"
	NOTICE_USER           NOTICE_USER_TYPE = "NOTICE_USER"
	NOTICE_NEWSLETTER     NOTICE_USER_TYPE = "NOTICE_NEWSLETTER"
)

func (n NoticeMsg) TableName() string {
	return "notice_msg"
}

func (n NoticeMsg) SelectAll() string {
	return n.TableName() + ".*"
}

func (n NoticeMsg) IsEmpty() bool {
	return n.Id == 0
}

func (NoticeMsg) GetNoticeUserTypeGroup(loc string) commoncode.CodeGroup {

	codes := make([]commoncode.Code, 0, 0)
	codes = append(codes, commoncode.Code{Code: string(NOTICE_PROJECT), Name: i18nutils.Translate("commoncode.NOTICE_USER_TYPE.NOTICE_PROJECT", loc), Order: 1})
	codes = append(codes, commoncode.Code{Code: string(NOTICE_DIRECT), Name: i18nutils.Translate("commoncode.NOTICE_USER_TYPE.NOTICE_DIRECT", loc), Order: 2})
	codes = append(codes, commoncode.Code{Code: string(NOTICE_USER), Name: i18nutils.Translate("commoncode.NOTICE_USER_TYPE.NOTICE_USER", loc), Order: 3})
	codes = append(codes, commoncode.Code{Code: string(NOTICE_NEWSLETTER), Name: i18nutils.Translate("commoncode.NOTICE_USER_TYPE.NOTICE_NEWSLETTER", loc), Order: 3})

	return commoncode.CodeGroup{
		GroupCode: NOTICE_USER_TYPE_CODE,
		Codes:     codes,
	}
}

func (NoticeMsg) GetAll(ctx context.Context, title *string, startDate, endDate, sortby, order string, offset, limit int) (int64, []NoticeMsg, error) {
	query := factory.DB(ctx).Table("notice_msg").Select(NOTICE_LIST)
	query.Join("INNER", "user_notice_mapping", "notice_msg.id = user_notice_mapping.msg_id")
	query.GroupBy("user_notice_mapping.msg_id")

	if err := modelhelper.SetSortOrder(query, sortby, order); err != nil {
		factory.Logger(ctx).Error(err)
	}

	if err := modelhelper.SearchLikeQuery(query, "notice_msg", "title", title); err != nil {
		factory.Logger(ctx).Error(err)
	}

	if err := modelhelper.SetBetweenDate(query, "notice_msg.notice_date", startDate, endDate); err != nil {
		factory.Logger(ctx).Error(err)
	}

	var noticeList []NoticeMsg
	totalCount, err := query.Limit(limit, offset).FindAndCount(&noticeList)
	if err != nil {
		return 0, nil, err
	}

	return totalCount, noticeList, nil
}

func (NoticeMsg) GetById(ctx context.Context, id int64) (*NoticeMsg, error) {
	var noticeMsg NoticeMsg
	has, err := factory.DB(ctx).ID(id).Get(&noticeMsg)

	if err != nil {
		return nil, err
	} else if !has {
		return nil, nil
	}
	var userList []User
	err = factory.DB(ctx).Table("user").Select("user.*").Where("id IN (SELECT user_id FROM user_notice_mapping WHERE msg_id = ?)", id).Asc("user.email").Find(&userList)
	if err != nil {
		return nil, err
	}
	noticeMsg.Users = userList

	switch noticeMsg.NoticeUserType {
	case NOTICE_USER:
		{
			var investorTypeList []INVESTOR_TYPE
			err = factory.DB(ctx).Table("user").Select("user.investor_type").Where("id IN (select user_id FROM user_notice_mapping WHERE msg_id = ?)", id).GroupBy("user.investor_type").Find(&investorTypeList)
			if err != nil {
				return nil, err
			}
			noticeMsg.UserInvestorTypeList = investorTypeList
		}
	case NOTICE_PROJECT:
		factory.DB(ctx).Table("project").Select("project.name").Where("ID = ?", noticeMsg.ProjectId).Get(&noticeMsg.ProjectName)
	}

	return &noticeMsg, nil
}

func (NoticeMsg) GetByIdWithUserId(ctx context.Context, id, userId int64) (*NoticeMsg, error) {
	var noticeMsg NoticeMsg
	cnt, err := factory.DB(ctx).Table("user_notice_mapping").Where("msg_id = ? AND user_id = ?", id, userId).Count(&noticeMsg)
	if err != nil {
		return nil, err
	}
	if cnt <= 0 {
		return nil, factory.ErrorNotHaveRole.New(errors.New("not have role"))
	}
	has, err := factory.DB(ctx).ID(id).Get(&noticeMsg)
	if err != nil {
		return nil, err
	} else if !has {
		return nil, nil
	}
	return &noticeMsg, nil
}

func (noticeMsg *NoticeMsg) Create(ctx context.Context) (int64, error) {
	return factory.DB(ctx).Insert(noticeMsg)
}

func (NoticeMsg) GetUserNoticeMsg(ctx context.Context, title *string, startDate, endDate, sortby, order string, offset, limit int, userId int64) (int64, []NoticeMsg, error) {
	query := factory.DB(ctx).Table("notice_msg").Select("notice_msg.*")
	query.Join("INNER", "user_notice_mapping", "notice_msg.id = user_notice_mapping.msg_id")
	query.Where("user_notice_mapping.user_id=?", userId)
	if err := modelhelper.SetSortOrder(query, sortby, order); err != nil {
		factory.Logger(ctx).Error(err)
	}

	if err := modelhelper.SearchLikeQuery(query, "notice_msg", "title", title); err != nil {
		factory.Logger(ctx).Error(err)
	}
	if err := modelhelper.SetBetweenDate(query, "notice_msg.notice_date", startDate, endDate); err != nil {
		factory.Logger(ctx).Error(err)
	}

	var resList []NoticeMsg
	var totalCount int64

	totalCount, err := query.Limit(limit, offset).FindAndCount(&resList)
	if err != nil {
		return 0, nil, err
	}

	return totalCount, resList, nil
}

type NoticeMsgRead struct {
	Id     int64              `json:"id" xorm:"pk autoincr"`
	MsgId  int64              `json:"msg_id" xorm:"not null"`
	UserId int64              `json:"user_id" xorm:"not null"`
	IsRead bool               `json:"is_read" xorm:"not null"`
	ReadAt jsonutils.JSONTime `json:"read_at" xorm:"created"`
}

func (n NoticeMsgRead) TableName() string {
	return "notice_msg_read"
}

func (n NoticeMsgRead) SelectAll() string {
	return n.TableName() + ".*"
}

type UserNoticeMapping struct {
	MsgId  int64 `json:"msg_id" xorm:"not null"`
	UserId int64 `json:"user_id" xorm:"not null"`
}

type UserNoticeMappings []UserNoticeMapping

func (m UserNoticeMapping) TableName() string {
	return "user_notice_mapping"
}

func (m UserNoticeMapping) SelectAll() string {
	return m.TableName() + ".*"
}

func (m UserNoticeMapping) IsEmpty() bool {
	return m.MsgId == 0
}

func (m *UserNoticeMapping) Create(ctx context.Context) (int64, error) {
	return factory.DB(ctx).Insert(m)
}

func (ms UserNoticeMappings) UserIds() []int64 {
	userIds := make([]int64, 0)
	for _, mapping := range ms {
		userIds = append(userIds, mapping.UserId)
	}
	return userIds
}

func (ms UserNoticeMappings) MsgIds() []int64 {
	msgIds := make([]int64, 0)
	for _, mapping := range ms {
		msgIds = append(msgIds, mapping.MsgId)
	}
	return msgIds
}

type NoticeMsgUserCount struct {
	MsgId     int64 `json:"msg_id"`
	UserCount int64 `json:"user_count"`
}
