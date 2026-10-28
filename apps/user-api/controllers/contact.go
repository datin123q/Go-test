package controllers

import (
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/pangpanglabs/echoswagger/v2"
	"github.com/sirupsen/logrus"

	"github.com/illuminarean/energy-shares-backend-repo/domain"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/factory"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/jsonutils"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/serviceutils"
)

type ContactApiController struct{}

func (c ContactApiController) Init(g echoswagger.ApiGroup) {
	g.POST("/send-contact-mail", c.SendContactMail).
		AddParamBody(ContactInput{}, "body", "", false).
		AddResponse(http.StatusCreated, "お問い合わせ登録およびメール送信", domain.Contact{}, nil).
		SetSummary("お問い合わせ登録およびメール送信").
		SetDescription("ユーザーがお問い合わせメールを送信すると、管理者にお問い合わせメールが送信されます。")
}

func (ContactApiController) SendContactMail(c echo.Context) error {
	var v ContactInput
	var mail domain.Mail

	if err := c.Bind(&v); err != nil {
		return factory.ErrorParameter.New(err)
	}
	if err := c.Validate(&v); err != nil {
		return factory.ErrorParameter.New(err)
	}

	// HTMLタグをエスケープ
	v.Sanitize()

	dt := jsonutils.JSONTime(time.Now()).ToString()

	contact, err := v.ToModel()
	if err != nil {
		return factory.ErrorParameter.New(err)
	}
	if contact == nil {
		return factory.ErrorNotFound.New(err)
	}

	if _, err := contact.Create(c.Request().Context()); err != nil {
		return factory.ErrorDB.New(err)
	}

	// ユーザーへの問い合わせ結果メール送信
	err = mail.ContactMail(c.Request().Context(), serviceutils.GetServiceInfoMail(), v.Email, v.Name, v.Email, dt, v.Contents, "", "")
	if err != nil {
		// MIR-1365: 宛先・氏名・件名・内容などの個人情報は出さない。contactIdから
		// 対象を特定できる。
		factory.Logger(c.Request().Context()).WithFields(logrus.Fields{
			"contactId":       contact.Id,
			"contactDatetime": dt,
			"err":             err,
		}).Error("ErrorFailContactMailSend")
	}

	// サポートへの問い合わせメール送信
	// NOTE: replyToName にもユーザーのメールアドレスを設定する
	err = mail.ContactMail(
		c.Request().Context(), serviceutils.GetServiceInfoMail(), serviceutils.GetServiceSupportMail(),
		v.Name, v.Email, dt, v.Contents, v.Email, v.Email,
	)
	if err != nil {
		// MIR-1365: 同上。
		factory.Logger(c.Request().Context()).WithFields(logrus.Fields{
			"contactId":       contact.Id,
			"contactDatetime": dt,
			"err":             err,
		}).Error("ErrorFailContactMailSend")
		return factory.ErrorFailContactMailSend.New(err)
	}

	return renderSucc(c, http.StatusOK, nil)
}
