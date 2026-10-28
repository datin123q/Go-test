package controllers

import (
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/pangpanglabs/echoswagger/v2"
	"github.com/sirupsen/logrus"

	"github.com/illuminarean/energy-shares-backend-repo/domain"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/apirender"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/factory"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/jsonutils"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/jwthelper"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/serviceutils"
)

type FundingRequestApiController struct{}

func (c FundingRequestApiController) Init(g echoswagger.ApiGroup) {
	g.GET("", c.GetAll).
		AddParamQueryNested(FundingRequestSearchInput{}).
		AddResponse(http.StatusOK, "ファンディング申請書一覧", domain.FundingRequest{}, nil).
		SetDescription("ファンディング申請書一覧を取得します。").
		SetSummary("ファンディング申請書一覧取得").
		SetSecurity("Authorization")

	g.GET("/:id", c.GetOne).
		AddParamPath("", "id", "").
		AddResponse(http.StatusOK, "特定のファンディング申請書情報取得", domain.FundingRequest{}, nil).
		SetDescription("ファンディング申請書を取得します。").
		SetSummary("ファンディング申請書取得").
		SetSecurity("Authorization")

	g.PUT("/:id", c.Update).
		AddParamPath("", "id", "").
		AddParamForm(FundingRequestInput{}, "body", "", true).
		AddResponse(http.StatusOK, "特定のファンディング申請書情報更新", domain.FundingRequest{}, nil).
		SetRequestContentType("multipart/form-data").
		SetDescription("ファンディング申請書を更新します。").
		SetSummary("ファンディング申請書更新").
		SetSecurity("Authorization")

	g.GET("/exist-corp-no", c.ExistCorpNo).
		AddParamQueryNested(FundingRequestCorpNoInput{}).
		AddResponse(http.StatusOK, "法人番号重複確認", domain.FundingRequest{}, nil).
		SetDescription("法人番号の重複を確認します。").
		SetSummary("法人番号重複確認").
		SetSecurity("Authorization")

	g.GET("/list/download", c.Download).
		AddParamQueryNested(SearchInput{}).
		AddResponse(http.StatusOK, "ファンディング申請書一覧ダウンロード", domain.FundingRequest{}, nil).
		SetDescription("ファンディング申請書一覧をダウンロードします。").
		SetSummary("ファンディング申請書一覧ダウンロード").
		SetSecurity("Authorization")
}

func (FundingRequestApiController) GetAll(c echo.Context) error {
	var s FundingRequestSearchInput
	if err := c.Bind(&s); err != nil {
		return factory.ErrorParameter.New(err)
	}
	if s.MaxResultCount == 0 {
		s.MaxResultCount = DefaultMaxResultCount
	}

	factory.Logger(c.Request().Context()).WithFields(logrus.Fields{
		"sortby":           s.Sortby,
		"order":            s.Order,
		"max_result_count": s.MaxResultCount,
		"skip_count":       s.SkipCount,
	}).Info("SearchStart")

	var items []domain.FundingRequest
	totalCount, items, err := domain.FundingRequest{}.GetAll(c.Request().Context(),
		s.Sortby, s.Order, s.SkipCount, s.MaxResultCount, s.Id, s.Email, s.Phone, s.RequestStatus, s.StartDate, s.EndDate)
	if err != nil {
		return factory.ErrorParameter.New(err)
	}

	userId, email, err := jwthelper.TokenToUser(c)
	if err != nil {
		return err
	}

	err = domain.OperatorHistory{}.Record(c.Request().Context(), userId, email,
		"menu.manage_product.funding_request_list", "", c.RealIP(), domain.LOOKUP)
	if err != nil {
		return factory.ErrorDB.New(err)
	}

	return apirender.RenderSucc(c, http.StatusOK, apirender.ArrayResult{
		Items:      items,
		TotalCount: totalCount,
	})
}

func (FundingRequestApiController) GetOne(c echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return factory.ErrorParameter.New(err)
	}

	userId, email, err := jwthelper.TokenToUser(c)
	if err != nil {
		return err
	}

	fr, err := domain.FundingRequest{}.GetById(c.Request().Context(), id)
	if err != nil {
		return factory.ErrorDB.New(err)
	}
	if fr == nil {
		return factory.ErrorNotFound.New(err)
	}

	fr.Attachments, err = domain.Attachment{}.FindByResourceId(c.Request().Context(), 10, fr.CommonId)
	if err != nil {
		return factory.ErrorDB.New(err)
	}

	err = domain.OperatorHistory{}.Record(c.Request().Context(), userId, email,
		"menu.manage_product.funding_request_detail", fr.Name, c.RealIP(), domain.LOOKUP)
	if err != nil {
		return factory.ErrorDB.New(err)
	}

	return apirender.RenderSucc(c, http.StatusOK, fr)
}

func (FundingRequestApiController) Update(c echo.Context) error {
	ctx := c.Request().Context()

	var fri FundingRequestInput
	if err := c.Bind(&fri); err != nil {
		return factory.ErrorParameter.New(err)
	}
	if err := c.Validate(&fri); err != nil {
		return factory.ErrorParameter.New(err)
	}

	userId, email, err := jwthelper.TokenToUser(c)
	if err != nil {
		return err
	}

	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return factory.ErrorParameter.New(err)
	}

	before, err := domain.FundingRequest{}.GetById(ctx, id)
	if err != nil {
		return factory.ErrorParameter.New(err)
	}
	if before == nil {
		return factory.ErrorNotFound.New(errors.New("funding request is not found"))
	}

	fr, err := fri.ToModel()
	if err != nil {
		return factory.ErrorParameter.New(err)
	}
	fr.Id = id

	form, err := c.MultipartForm()
	if err != nil {
		return factory.ErrorDB.New(err)
	}

	attachmentFilesToDelete, err := deleteOwnedAttachments(ctx, form.Value, before.CommonId,
		logrus.Fields{"user": userId, "funding_request_id": id})
	if err != nil {
		return err
	}
	fr.Attachments, err = domain.Attachment{}.AddAttachments(ctx, before.CommonId, form)
	if err != nil {
		return factory.ErrorAWSUpload.New(err)
	}

	if err := fr.Update(ctx); err != nil {
		return factory.ErrorDB.New(err)
	}

	err = domain.OperatorHistory{}.Record(ctx, userId, email,
		"menu.manage_product.funding_request", fr.Name, c.RealIP(), domain.CHANGE)
	if err != nil {
		return factory.ErrorDB.New(err)
	}

	after, err := domain.FundingRequest{}.GetById(ctx, fr.Id)
	if err != nil {
		return factory.ErrorDB.New(err) // データ修正後、照会ができない場合、その部分はDBエラー
	}
	err = domain.FundingRequestHistory{}.Create(ctx, before, after, userId)
	if err != nil {
		return factory.ErrorDB.New(err)
	}

	if before.RequestStatus != after.RequestStatus {
		switch after.RequestStatus {
		case domain.FAIL_REQUEST:
			{
				var mail domain.Mail
				err = mail.FundingFailMail(ctx, serviceutils.GetServiceLoanMail(), fr.Email, fr.Name, jsonutils.JSONTime(time.Now()).ToString())
				if err != nil {
					// MIR-1365: 以前は宛先のメールアドレスをそのままログへ出していた。after.Id
					// から対象を特定できるため、宛先そのものは出さない。
					factory.Logger(ctx).WithFields(logrus.Fields{
						"err":              err,
						"fundingRequestId": after.Id,
						"fundingStatus":    after.RequestStatus,
					}).Error("fail funding request  mail")
				}
			}

		case domain.PASS_REQUEST:
			{
				var mail domain.Mail
				err = mail.FundingSuccessMail(ctx, serviceutils.GetServiceLoanMail(), fr.Email, fr.Name, jsonutils.JSONTime(time.Now()).ToString())
				if err != nil {
					// MIR-1365: 同上。宛先そのものは出さない。
					factory.Logger(ctx).WithFields(logrus.Fields{
						"err":              err,
						"fundingRequestId": after.Id,
						"fundingStatus":    after.RequestStatus,
					}).Error("fail funding request  mail")
				}

				err = domain.OperatorHistory{}.Record(ctx, userId, email,
					"commoncode.LOG_CONTENT_TYPE.FUNDING_REQUEST_ACCEPT", fr.Name, c.RealIP(), domain.CHANGE)
				if err != nil {
					return factory.ErrorDB.New(err)
				}
			}

		}
	}

	// 上のswitchでも、承認の操作履歴やメールの再送用の行(mail_retry)をDBへ書くので、S3の削除は、DBの書き込みが
	// 全部済んだこの位置で登録する。
	domain.DeleteAttachmentFilesAfterCommit(ctx, attachmentFilesToDelete)

	return apirender.RenderSucc(c, http.StatusOK, fr)
}

func (FundingRequestApiController) ExistCorpNo(c echo.Context) error {
	var input FundingRequestCorpNoInput
	if err := c.Bind(&input); err != nil {
		return factory.ErrorParameter.New(err)
	}
	if err := c.Validate(&input); err != nil {
		return factory.ErrorParameter.New(err)
	}

	fr, err := domain.FundingRequest{}.GetByCorpNo(c.Request().Context(), input.CorpNo)
	if err != nil {
		return factory.ErrorDB.New(err)
	}
	if fr != nil {
		return apirender.RenderSucc(c, http.StatusOK, true)
	}

	return apirender.RenderSucc(c, http.StatusOK, false)
}

func (FundingRequestApiController) Download(c echo.Context) error {
	var v FundingRequestSearchInput
	if err := c.Bind(&v); err != nil {
		return factory.ErrorParameter.New(err)
	}
	if v.MaxResultCount == 0 {
		v.MaxResultCount = DefaultMaxResultCount
	}
	userId, email, err := jwthelper.TokenToUser(c)
	if err != nil {
		return factory.ErrorTokenInvalid.New(err)
	}

	err = domain.FundingRequest{}.DownloadCsv(c.Request().Context(), c.Response().Writer,
		v.Sortby, v.Order, v.SkipCount, v.MaxResultCount, v.Id, v.Email, v.Phone, v.RequestStatus, v.StartDate, v.EndDate)
	if err != nil {
		return factory.ErrorFailDownloadCSV.New(err)
	}

	err = domain.OperatorHistory{}.Record(c.Request().Context(), userId, email,
		"menu.manage_product.funding_request_list", "", c.RealIP(), domain.DOWNLOAD)
	if err != nil {
		return factory.ErrorDB.New(err)
	}

	return nil
}
