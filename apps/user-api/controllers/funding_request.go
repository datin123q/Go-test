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

type FundingRequestApiController struct{}

func (c FundingRequestApiController) Init(g echoswagger.ApiGroup) {
	g.POST("", c.Create).
		AddParamForm(FundingRequestInput{}, "formData", "", false).
		AddResponse(http.StatusCreated, "ファンディング申請書作成", domain.FundingRequest{}, nil).
		SetRequestContentType("multipart/form-data").
		SetDescription("ファンディング申請書を作成します。").
		SetSummary("ファンディング申請書作成")
}

func (FundingRequestApiController) Create(c echo.Context) error {
	form, err := c.MultipartForm()
	if err != nil {
		return factory.ErrorInvalidForm.New(err)
	}

	var fri FundingRequestInput
	if err := c.Bind(&fri); err != nil {
		return factory.ErrorParameter.New(err)
	}
	if err := c.Validate(&fri); err != nil {
		return factory.ErrorParameter.New(err)
	}

	// HTMLタグをエスケープ
	fri.Sanitize()

	fr, err := fri.ToModel()
	if err != nil {
		return factory.ErrorParameter.New(err)
	}
	if fr == nil {
		return factory.ErrorNotFound.New(err)
	}
	if _, err := fr.Create(c.Request().Context()); err != nil {
		return factory.ErrorDB.New(err)
	}

	fr.Attachments, err = domain.Attachment{}.AddAttachments(c.Request().Context(), fr.CommonId, form)
	if err != nil {
		return factory.ErrorAWSUpload.New(err)
	}

	var mail domain.Mail
	err = mail.FundingRequestMail(c.Request().Context(), serviceutils.GetServiceLoanMail(), fr.Email, fr.Name, jsonutils.JSONTime(time.Now()).ToString())
	if err != nil {
		// MIR-1365: 以前は宛先のメールアドレスをそのままログへ出していた。fundingRequestId
		// から対象を特定できるため、宛先そのものは出さない。
		factory.Logger(c.Request().Context()).WithFields(logrus.Fields{
			"err":              err,
			"fundingRequestId": fr.Id,
			"fundingStatus":    fr.RequestStatus,
		}).Error("fail funding request  mail")
	}

	return renderSucc(c, http.StatusOK, fr)
}
