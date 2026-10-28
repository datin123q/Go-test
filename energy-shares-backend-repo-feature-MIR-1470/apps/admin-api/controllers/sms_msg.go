package controllers

import (
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
	"github.com/pangpanglabs/echoswagger/v2"
	"github.com/pkg/errors"

	"github.com/illuminarean/energy-shares-backend-repo/domain"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/apirender"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/factory"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/jwthelper"
)

type SmsApiController struct {
}

func (c SmsApiController) Init(g echoswagger.ApiGroup) {
	g.GET("", c.GetAll).
		AddParamQueryNested(SmsSearchInput{}).
		AddResponse(http.StatusOK, "メッセージ送信リスト照会", domain.SmsMsg{}, nil).
		SetDescription("送信メッセージリストを照会します。").
		SetSummary("メッセージリストリスト照会").
		SetSecurity("Authorization")

	g.GET("/:id", c.GetOne).
		AddParamPath("", "id", "メッセージID").
		AddResponse(http.StatusOK, "メッセージ送信詳細照会", domain.SmsMsg{}, nil).
		SetDescription("送信メッセージを照会します。").
		SetSummary("メッセージ詳細照会").
		SetSecurity("Authorization")

	g.POST("", c.Create).
		AddParamBody(SmsInput{}, "body", "", false).
		AddResponse(http.StatusCreated, "登録メッセージ", domain.SmsMsg{}, nil).
		SetDescription("メッセージを作成します。").
		SetSummary("メッセージ作成").
		SetSecurity("Authorization")

	g.POST("/member", c.CreateMemberSMS).
		AddParamBody(SmsInput{}, "body", "", false).
		AddResponse(http.StatusCreated, "登録メッセージ", domain.SmsMsg{}, nil).
		SetDescription("メンバーメッセージを作成します。").
		SetSummary("メンバーメッセージ作成").
		SetSecurity("Authorization")
}

// GetAll SMS送信履歴
func (SmsApiController) GetAll(c echo.Context) error {
	var v SmsSearchInput
	if err := c.Bind(&v); err != nil {
		return factory.ErrorParameter.New(err)
	}
	if v.MaxResultCount == 0 {
		v.MaxResultCount = DefaultMaxResultCount
	}

	userId, email, err := jwthelper.TokenToUser(c)
	if err != nil {
		return err
	}

	totalCount, items, err := domain.SmsMsg{}.GetAll(c.Request().Context(), v.Content, v.Name, v.Sender, v.Receiver, v.Status, v.StartDate, v.EndDate, v.Sortby, v.Order, v.SkipCount, v.MaxResultCount)
	if err != nil {
		return factory.ErrorDB.New(err)
	}

	err = domain.OperatorHistory{}.Record(c.Request().Context(), userId, email,
		"menu.manage_mail_sms.manage_sms", "", c.RealIP(), domain.LOOKUP)
	if err != nil {
		return factory.ErrorDB.New(err)
	}

	return apirender.RenderSucc(c, http.StatusOK, apirender.ArrayResult{
		Items:      items,
		TotalCount: totalCount,
	})
}

// Create SMS作成
func (SmsApiController) Create(c echo.Context) error {
	var v SmsInput
	if err := c.Bind(&v); err != nil {
		return factory.ErrorParameter.New(err)
	}
	if err := c.Validate(&v); err != nil {
		return factory.ErrorParameter.New(err)
	}

	sms, err := v.ToModel()
	if err != nil {
		return factory.ErrorParameter.New(err)
	}

	userId, email, err := jwthelper.TokenToUser(c)
	if err != nil {
		return err
	}
	sms.Sender = email

	user, err := domain.User{}.GetByEmail(c.Request().Context(), email)
	if err != nil {
		return err
	}
	if user != nil {
		sms.Name = user.Name
	}

	// create sms
	if err := createSms(c, sms); err != nil {
		return err
	}

	err = domain.OperatorHistory{}.Record(c.Request().Context(), userId, email,
		"menu.manage_mail_sms.manage_sms_manage_detail", sms.Title, c.RealIP(), domain.REGISTER)
	if err != nil {
		return factory.ErrorDB.New(err)
	}

	// send sms
	if err := sms.Send(c.Request().Context()); err != nil {
		return factory.ErrorAWSSendSMS.New(err)
	}

	err = domain.OperatorHistory{}.Record(c.Request().Context(), userId, email,
		"menu.manage_mail_sms.manage_sms_manage_send", sms.Title, c.RealIP(), domain.REGISTER)
	if err != nil {
		return factory.ErrorDB.New(err)
	}

	return apirender.RenderSucc(c, http.StatusOK, sms)
}

// CreateMemberSMS SMS作成
func (SmsApiController) CreateMemberSMS(c echo.Context) error {
	var v SmsInput
	if err := c.Bind(&v); err != nil {
		return factory.ErrorParameter.New(err)
	}
	if err := c.Validate(&v); err != nil {
		return factory.ErrorParameter.New(err)
	}

	users, err := domain.User{}.GetByUserType(c.Request().Context(), v.ReceiverNumbers)
	if err != nil {
		return factory.ErrorDB.New(err)
	}

	var receiverNumbers []string
	for _, user := range users {
		if user.MobilePhone != "" {
			receiverNumbers = append(receiverNumbers, user.MobilePhone)
		}
	}
	v.ReceiverNumbers = receiverNumbers

	sms, err := v.ToModel()
	if err != nil {
		return factory.ErrorParameter.New(err)
	}

	userId, email, err := jwthelper.TokenToUser(c)
	if err != nil {
		return err
	}
	sms.Sender = email

	// create sms
	if err := createSms(c, sms); err != nil {
		return err
	}

	err = domain.OperatorHistory{}.Record(c.Request().Context(), userId, email,
		"menu.manage_mail_sms.manage_sms_manage_detail", sms.Title, c.RealIP(), domain.REGISTER)
	if err != nil {
		return factory.ErrorDB.New(err)
	}

	// send sms
	if err := sms.Send(c.Request().Context()); err != nil {
		return factory.ErrorAWSSendSMS.New(err)
	}
	err = domain.OperatorHistory{}.Record(c.Request().Context(), userId, email,
		"menu.manage_mail_sms.manage_sms_manage_send", sms.Title, c.RealIP(), domain.REGISTER)
	if err != nil {
		return factory.ErrorDB.New(err)
	}

	return apirender.RenderSucc(c, http.StatusOK, sms)
}

func (SmsApiController) GetOne(c echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return factory.ErrorParameter.New(err)
	}

	userId, email, err := jwthelper.TokenToUser(c)
	if err != nil {
		return err
	}

	sms, err := domain.SmsMsg{}.GetById(c.Request().Context(), id)
	if err != nil {
		return factory.ErrorDB.New(err)
	}
	if sms == nil {
		return factory.ErrorNotFound.New(nil)
	}

	mapSmsReceiver := &domain.MapSmsReceiver{}
	_, mapSmsReceiverList, err := mapSmsReceiver.FindBySmsId(c.Request().Context(), id)
	if err != nil {
		return factory.ErrorDB.New(errors.Wrap(err, "fail to find MAP_SMS_RECEIVER"))
	}

	var receiverList []string
	for _, receiver := range mapSmsReceiverList {
		receiverList = append(receiverList, receiver.Receiver)
	}
	sms.Receiver = receiverList

	err = domain.OperatorHistory{}.Record(c.Request().Context(), userId, email,
		"menu.manage_mail_sms.manage_sms_manage_detail", sms.Title, c.RealIP(), domain.LOOKUP)
	if err != nil {
		return factory.ErrorDB.New(err)
	}

	return apirender.RenderSucc(c, http.StatusOK, sms)
}

func createSms(c echo.Context, smsMsg *domain.SmsMsg) error {
	if _, err := smsMsg.Create(c.Request().Context()); err != nil {
		return factory.ErrorDB.New(err)
	}

	for _, receiver := range smsMsg.Receiver {
		mapSmsReceiver := &domain.MapSmsReceiver{
			SmsId:    smsMsg.Id,
			Receiver: receiver,
		}

		if _, err := mapSmsReceiver.Create(c.Request().Context()); err != nil {
			return factory.ErrorDB.New(errors.Wrap(err, "fail to insert MAP_SMS_RECEIVER"))
		}
	}

	return nil
}
