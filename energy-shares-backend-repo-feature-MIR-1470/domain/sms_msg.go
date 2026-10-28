package domain

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"time"

	"xorm.io/xorm"

	"github.com/sirupsen/logrus"

	"github.com/illuminarean/energy-shares-backend-repo/pkg/commoncode"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/factory"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/i18nutils"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/jsonutils"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/modelhelper"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/serviceutils"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/snsutils"
)

type SMS_TYPE string

const (
	SMS_TYPE_CODE                        = "SMS_TYPE"
	NORMAL_SMS                  SMS_TYPE = "NORMAL_SMS"
	AUTO_SMS                    SMS_TYPE = "AUTO_SMS"
	SIGN_UP_COMPLETE_SMS        SMS_TYPE = "SIGN_UP_COMPLETE_SMS"        // 회원가입 완료
	PASSWORD_RESET_COMPLETE_SMS SMS_TYPE = "PASSWORD_RESET_COMPLETE_SMS" // 비밀번호 변경 완료
)

func (SmsMsg) GetSmsTypeCodeGroup(loc string) commoncode.CodeGroup {

	codes := make([]commoncode.Code, 0, 0)
	codes = append(codes, commoncode.Code{Code: string(NORMAL_SMS), Name: i18nutils.Translate("commoncode.SMS_TYPE.NORMAL_SMS", loc), Order: len(codes) + 1})
	codes = append(codes, commoncode.Code{Code: string(AUTO_SMS), Name: i18nutils.Translate("commoncode.SMS_TYPE.AUTO_SMS", loc), Order: len(codes) + 1})
	codes = append(codes, commoncode.Code{Code: string(SIGN_UP_COMPLETE_SMS), Name: i18nutils.Translate("commoncode.SMS_TYPE.SIGN_UP_COMPLETE_SMS", loc), Order: len(codes) + 1})
	codes = append(codes, commoncode.Code{Code: string(PASSWORD_RESET_COMPLETE_SMS), Name: i18nutils.Translate("commoncode.SMS_TYPE.PASSWORD_RESET_COMPLETE_SMS", loc), Order: len(codes) + 1})

	return commoncode.CodeGroup{
		GroupCode: SMS_TYPE_CODE,
		Codes:     codes,
	}
}

type SMS_STATUS string

const (
	SMS_STATUS_CODE            = "SMS_STATUS"
	REQUEST_SMS     SMS_STATUS = "REQUEST_SMS"
	SUCCESS_SMS     SMS_STATUS = "SUCCESS_SMS"
	FAIL_SMS        SMS_STATUS = "FAIL_SMS"
)

func (SmsMsg) GetSmsStatusCodeGroup(loc string) commoncode.CodeGroup {
	codes := make([]commoncode.Code, 0, 0)
	codes = append(codes, commoncode.Code{Code: string(REQUEST_SMS), Name: i18nutils.Translate("commoncode.SMS_STATUS.REQUEST_SMS", loc), Order: 1})
	codes = append(codes, commoncode.Code{Code: string(SUCCESS_SMS), Name: i18nutils.Translate("commoncode.SMS_STATUS.SUCCESS_SMS", loc), Order: 2})
	codes = append(codes, commoncode.Code{Code: string(FAIL_SMS), Name: i18nutils.Translate("commoncode.SMS_STATUS.FAIL_SMS", loc), Order: 3})

	return commoncode.CodeGroup{
		GroupCode: SMS_STATUS_CODE,
		Codes:     codes,
	}
}

type SmsMsg struct {
	Id         int64              `json:"id"`
	MsgId      string             `json:"msg_id"`
	Title      string             `json:"title"`
	Name       string             `json:"name"`
	Sender     string             `json:"sender" xorm:"not null"`
	Receiver   []string           `json:"receiver" xorm:"-"`
	Status     SMS_STATUS         `json:"status"`
	SmsType    SMS_TYPE           `json:"sms_type"`
	SmsContent string             `json:"sms_content" xorm:"TEXT"`
	Count      int64              `json:"count"`
	SendDate   jsonutils.JSONTime `json:"send_date"`
	Created    jsonutils.JSONTime `json:"created" xorm:"created"` // 생성 날짜
}

type MapSmsReceiver struct {
	Id       int64              `json:"id"`
	SmsId    int64              `json:"sms_id" xorm:"not null"`   // sms_msg.id
	Receiver string             `json:"receiver" xorm:"not null"` // 전화 번호
	Created  jsonutils.JSONTime `json:"created" xorm:"created"`   // 생성 날짜
}

type SmsCertInfo struct {
	Code string
}

type SmsPasswordChangedInfo struct {
	ServiceCall string
}

func (m *SmsMsg) Message() string {
	return m.SmsContent
}

func (m *SmsMsg) Subject() string {
	return m.Title
}

// Send 문자 메시지 전송
func (m *SmsMsg) Send(ctx context.Context) error {
	if m.Status != REQUEST_SMS {
		return errors.New("invalid sms status")
	}
	if m.Title == "" {
		m.Title = "Empty"
	}

	var err error
	for _, receiverNumber := range m.Receiver {
		m.MsgId, err = snsutils.SendSMS(ctx, receiverNumber, m)
		if err != nil {
			factory.Logger(ctx).WithFields(logrus.Fields{
				"msgId":    m.MsgId,
				"smsType":  m.SmsType,
				"receiver": receiverNumber,
				"content":  m.SmsContent,
				"error":    err,
			}).Error("Fail SMS Send")
		}
	}

	m.Status = SUCCESS_SMS
	m.SendDate = jsonutils.JSONTime(time.Now())

	if err := m.Update(ctx); err != nil {
		return err
	}

	return nil
}

func (m *SmsMsg) Update(ctx context.Context) (err error) {
	if _, err := factory.DB(ctx).ID(m.Id).Update(m); err != nil {
		return err
	}
	return nil
}

func (m *SmsMsg) Create(ctx context.Context) (int64, error) {
	m.Status = REQUEST_SMS
	return factory.DB(ctx).Insert(m)
}

func (SmsMsg) GetById(ctx context.Context, id int64) (*SmsMsg, error) {
	var v SmsMsg
	if has, err := factory.DB(ctx).ID(id).Get(&v); err != nil {
		return nil, err
	} else if !has {
		return nil, nil
	}
	return &v, nil
}

func (SmsMsg) GetAll(ctx context.Context, content, name, sender, receiver, status *string, startDate, endDate string, sortby, order string, offset, limit int) (int64, []SmsMsg, error) {
	var q *xorm.Session

	if receiver != nil {
		q = factory.DB(ctx).Table("sms_msg").
			Join("LEFT", "map_sms_receiver", "sms_msg.id = map_sms_receiver.sms_id").
			Select("sms_msg.*").
			GroupBy("sms_msg.id")

		if err := modelhelper.SearchLikeQuery(q, "map_sms_receiver", "receiver", receiver); err != nil {
			factory.Logger(ctx).Error(err)
		}
	} else {
		q = factory.DB(ctx).Table("sms_msg").
			Select("sms_msg.*").
			GroupBy("sms_msg.id")
	}

	// MIR-1113: 以前は空要素にも "sms_msg." を前置してしまい(例: sortby="created,"
	// -> "sms_msg.created,sms_msg.")、末尾の不正な列名で並び替え全体が弾かれて
	// いた。QualifySortColumns は空要素を前置せずそのまま残すため、
	// SetSortOrderForTable に置き換える。
	if err := modelhelper.SetSortOrderForTable(q, "sms_msg", sortby, order); err != nil {
		factory.Logger(ctx).Error(err)
	}

	if err := modelhelper.SearchLikeQuery(q, "sms_msg", "sms_content", content); err != nil {
		factory.Logger(ctx).Error(err)
	}

	if err := modelhelper.SearchLikeQuery(q, "sms_msg", "name", name); err != nil {
		factory.Logger(ctx).Error(err)
	}

	if err := modelhelper.SearchLikeQuery(q, "sms_msg", "sender", sender); err != nil {
		factory.Logger(ctx).Error(err)
	}

	if err := modelhelper.SearchEqQuery(q, "sms_msg", "status", status); err != nil {
		factory.Logger(ctx).Error(err)
	}

	if err := modelhelper.SetBetweenDate(q, "send_date", startDate, endDate); err != nil {
		factory.Logger(ctx).Error(err)
	}

	var items []SmsMsg
	totalCount, err := q.Limit(limit, offset).FindAndCount(&items)
	if err != nil {
		return 0, nil, err
	}

	return totalCount, items, nil
}

func (SmsMsg) GetUserSmsMsgList(ctx context.Context, phone string, sortBy, order string, offset, limit int) (int64, []SmsMsg, error) {

	query := factory.DB(ctx).Table("sms_msg").Select("sms_msg.*").
		Join("INNER", "map_sms_receiver", "sms_msg.id = map_sms_receiver.sms_id").
		Where("map_sms_receiver.receiver = ?", phone)

	// MIR-1113: 以前は空要素にも "sms_msg." を前置してしまい(例: sortBy="created,"
	// -> "sms_msg.created,sms_msg.")、末尾の不正な列名で並び替え全体が弾かれて
	// いた。QualifySortColumns は空要素を前置せずそのまま残すため、
	// SetSortOrderForTable に置き換える。
	if err := modelhelper.SetSortOrderForTable(query, "sms_msg", sortBy, order); err != nil {
		factory.Logger(ctx).Error(err)
	}

	var items []SmsMsg
	totalCount, err := query.Limit(limit, offset).FindAndCount(&items)
	if err != nil {
		return 0, nil, err
	}

	return totalCount, items, nil
}

// SendCertSms 본인 인증 문자 발송
func (SmsMsg) SendCertSms(ctx context.Context, Code, sender, receiver string) (*SmsMsg, error) {

	filePath := fmt.Sprintf("/template/sms/%s/sms-templates-smscert.txt", serviceutils.GetServiceLocale())
	t, err := snsutils.LoadTextTemplate(filePath)
	if err != nil {
		if err != nil {
			factory.Logger(ctx).WithFields(logrus.Fields{
				"error": err,
			}).Error("Send SMS CERT")
		}
	}

	var tpl bytes.Buffer
	if err := t.Execute(&tpl, SmsCertInfo{
		Code: Code,
	}); err != nil {
		return nil, err
	}

	sendSms := &SmsMsg{
		Title:      "Certification",
		Sender:     sender,
		Status:     REQUEST_SMS,
		SmsType:    SIGN_UP_COMPLETE_SMS,
		Count:      int64(1),
		SmsContent: tpl.String(),
		SendDate:   jsonutils.JSONTime(time.Now()),
	}

	user, err := User{}.GetByEmail(ctx, sender)
	if err != nil {
		return nil, err
	} else if user != nil {
		sendSms.Name = user.Name
	}

	if _, err := sendSms.Create(ctx); err != nil {
		return nil, err
	}

	mapSmsReceiver := &MapSmsReceiver{
		SmsId:    sendSms.Id,
		Receiver: receiver,
	}

	if _, err := mapSmsReceiver.Create(ctx); err != nil {
		return nil, err
	}

	sendSms.MsgId, err = snsutils.SendSMS(ctx, receiver, sendSms)
	if err != nil {
		sendSms.SendDate = jsonutils.JSONTime(time.Now())
		sendSms.Status = FAIL_SMS
		factory.Logger(ctx).WithFields(logrus.Fields{
			"msgId":    sendSms.MsgId,
			"smsType":  sendSms.SmsType,
			"receiver": receiver,
			"content":  sendSms.SmsContent,
			"error":    err,
		}).Error("Fail SMS Send")

	} else {
		sendSms.SendDate = jsonutils.JSONTime(time.Now())
		sendSms.Status = SUCCESS_SMS

		factory.Logger(ctx).WithFields(logrus.Fields{
			"msgId":    sendSms.MsgId,
			"smsType":  sendSms.SmsType,
			"receiver": receiver,
			"content":  sendSms.SmsContent,
		}).Info("Success SMS Send")

	}

	if err := sendSms.Update(ctx); err != nil {
		return nil, err
	}
	return sendSms, nil
}

func (msr *MapSmsReceiver) Create(ctx context.Context) (int64, error) {
	return factory.DB(ctx).Insert(msr)
}

func (msr *MapSmsReceiver) FindBySmsId(ctx context.Context, smsId int64) (int64, []MapSmsReceiver, error) {

	query := factory.DB(ctx).Where("sms_id = ?", smsId)

	var mapSmsReceivers []MapSmsReceiver
	total, err := query.FindAndCount(&mapSmsReceivers)
	if err != nil {
		return 0, nil, err
	}

	return total, mapSmsReceivers, nil
}
