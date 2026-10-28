package controllers

import (
	"net/http"
	"strconv"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/pangpanglabs/echoswagger/v2"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"

	"github.com/illuminarean/energy-shares-backend-repo/domain"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/factory"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/jsonutils"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/jwthelper"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/serviceutils"
)

type QnaApiController struct{}

func (q QnaApiController) Init(g echoswagger.ApiGroup) {
	g.GET("", q.GetAllByToken).
		AddParamQueryNested(QnaSearchInput{}).
		SetDescription("ユーザーのQ&A一覧情報を取得します。").
		SetSummary("Q&A一覧情報取得").
		SetSecurity("Authorization")

	g.GET("/:id", q.GetOne).
		AddParamPath("", "id", "").
		AddResponse(http.StatusOK, "特定のQ&A情報取得", domain.Qna{}, nil).
		SetDescription("Q&Aを取得します。").
		SetSummary("Q&A取得").
		SetSecurity("Authorization")

	g.POST("", q.Create).
		AddParamForm(QnaInput{}, "body", "", false).
		AddResponse(http.StatusCreated, "Q&A作成", domain.Qna{}, nil).
		SetRequestContentType("multipart/form-data").
		SetDescription("Q&Aを追加します。").
		SetSummary("Q&A追加").
		SetSecurity("Authorization")

	g.PUT("/:id", q.Update).
		AddParamPath("", "id", "").
		AddParamForm(QnaInput{}, "body", "", false).
		AddResponse(http.StatusOK, "Q&A編集", domain.Qna{}, nil).
		SetRequestContentType("multipart/form-data").
		SetDescription("Q&Aを編集します。").
		SetSummary("Q&A編集").
		SetSecurity("Authorization")

	g.DELETE("/:id", q.Delete).
		AddParamPath("", "id", "").
		AddResponse(http.StatusOK, "Q&A削除", nil, nil).
		SetDescription("Q&Aを削除します。").
		SetSummary("Q&A削除").
		SetSecurity("Authorization")
}

func (QnaApiController) GetAllByToken(c echo.Context) error {
	var v QnaSearchInput
	if err := c.Bind(&v); err != nil {
		return factory.ErrorParameter.New(err)
	}
	if v.MaxResultCount == 0 {
		v.MaxResultCount = DefaultMaxResultCount
	}
	if v.UserId == nil {
		return factory.ErrorParameter.New(errors.New("UserId is required"))
	}
	userId, err := strconv.ParseInt(*v.UserId, 10, 64)
	if err != nil {
		return factory.ErrorParameter.New(err)
	}
	if err := jwthelper.ValidateUserToken(c, userId); err != nil {
		return err
	}

	factory.Logger(c.Request().Context()).WithFields(logrus.Fields{
		"sortby":           v.Sortby,
		"order":            v.Order,
		"max_result_count": v.MaxResultCount,
		"skip_count":       v.SkipCount,
	}).Info("SearchStart")

	totalCount, items, err := domain.Qna{}.GetAll(c.Request().Context(), v.Sortby, v.Order, v.SkipCount, v.MaxResultCount, v.UserId, v.Name, v.Title, v.Status, v.StartDate, v.EndDate)
	if err != nil {
		return factory.ErrorDB.New(err)
	}

	return renderSucc(c, http.StatusOK, ArrayResult{
		TotalCount: totalCount,
		Items:      items,
	})
}

func (QnaApiController) GetOne(c echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return factory.ErrorParameter.New(err)
	}

	q, err := domain.Qna{}.GetById(c.Request().Context(), id)
	if err != nil {
		return factory.ErrorDB.New(err)
	}

	if q == nil {
		return factory.ErrorNotFound.New(err)
	}

	if err := jwthelper.ValidateUserToken(c, q.UserId); err != nil {
		return err
	}

	q.Attachments, err = domain.Attachment{}.FindByResourceId(c.Request().Context(), 10, q.CommonId)
	if err != nil {
		return factory.ErrorDB.New(err)
	}

	_, a, err := domain.Answer{}.GetAllAnswerById(c.Request().Context(), id, "created", "desc")
	if err != nil {
		return factory.ErrorDB.New(err)
	} else {
		q.Answers = a
	}

	for _, answer := range q.Answers {
		answer.Attachments, err = domain.Attachment{}.FindByResourceId(c.Request().Context(), 10, answer.CommonId)
		if err != nil {
			return factory.ErrorDB.New(err)
		}
	}

	return renderSucc(c, http.StatusOK, q)
}

func (QnaApiController) Create(c echo.Context) error {
	userId, userName, userEmail, err := jwthelper.TokenToUserExtended(c)
	if err != nil {
		return factory.ErrorTokenInvalid.New(err)
	}

	form, err := c.MultipartForm()
	if err != nil {
		return factory.ErrorInvalidForm.New(err)
	}

	var qi QnaInput
	if err := c.Bind(&qi); err != nil {
		return factory.ErrorParameter.New(err)
	}
	if err := c.Validate(&qi); err != nil {
		return factory.ErrorParameter.New(err)
	}

	// HTMLタグをエスケープ
	qi.Sanitize()

	q, err := qi.ToModel()
	if err != nil {
		return factory.ErrorParameter.New(err)
	}
	if q == nil {
		return factory.ErrorNotFound.New(err)
	}

	actor := domain.ChangelogActor{
		Id:    userId,
		Name:  userName,
		Email: userEmail,
	}
	q.Actor = actor
	q.UserId = userId
	q.Status = domain.QNA_RECEIVED // 作成時の初期値

	if _, err := q.Create(c.Request().Context()); err != nil {
		return factory.ErrorDB.New(err)
	}

	q.Attachments, err = domain.Attachment{}.AddAttachmentsExtended(c.Request().Context(), q.CommonId, form, actor)
	if err != nil {
		return factory.ErrorAWSUpload.New(err)
	}

	var mail domain.Mail
	dt := jsonutils.JSONTime(time.Now()).ToString()

	// ユーザーへの問い合わせ結果メール送信
	err = mail.ContactMail(c.Request().Context(), serviceutils.GetServiceInfoMail(), userEmail, userName, userEmail, dt, q.Content, "", "")
	if err != nil {
		// MIR-1365: 宛先・氏名・件名・内容などの個人情報は出さない。論理削除済み
		// ユーザーのQnA作成はDBへ保存されずq.Idが0のまま進むため、その場合でも
		// 追跡できるようuserIdも出す。
		factory.Logger(c.Request().Context()).WithFields(logrus.Fields{
			"qnaId":           q.Id,
			"userId":          userId,
			"contactDatetime": dt,
			"err":             err,
		}).Error("ErrorFailContactMailSend")
	}

	// サポートへの問い合わせメール送信
	// NOTE: replyToName にもユーザーのメールアドレスを設定する
	err = mail.ContactMail(
		c.Request().Context(), serviceutils.GetServiceInfoMail(), serviceutils.GetServiceSupportMail(),
		userName, userEmail, dt, q.Content, userEmail, userEmail,
	)
	if err != nil {
		// MIR-1365: 同上。
		factory.Logger(c.Request().Context()).WithFields(logrus.Fields{
			"qnaId":           q.Id,
			"userId":          userId,
			"contactDatetime": dt,
			"err":             err,
		}).Error("ErrorFailContactMailSend")
		return factory.ErrorFailContactMailSend.New(err)
	}

	return renderSucc(c, http.StatusOK, q)
}

func (QnaApiController) Update(c echo.Context) error {
	ctx := c.Request().Context()
	userId, userName, userEmail, err := jwthelper.TokenToUserExtended(c)
	if err != nil {
		return factory.ErrorTokenInvalid.New(err)
	}

	var qi QnaInput
	if err := c.Bind(&qi); err != nil {
		return factory.ErrorParameter.New(err)
	}
	if err := c.Validate(&qi); err != nil {
		return factory.ErrorParameter.New(err)
	}

	// HTMLタグをエスケープ
	qi.Sanitize()

	actor := domain.ChangelogActor{
		Id:    userId,
		Name:  userName,
		Email: userEmail,
	}

	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return factory.ErrorParameter.New(err)
	}

	beforeQna, err := domain.Qna{}.GetById(ctx, id)
	if err != nil {
		return factory.ErrorDB.New(err)
	}
	if beforeQna == nil {
		return factory.ErrorNotFound.New(errors.New("qna is not found"))
	}

	if err := jwthelper.ValidateUserToken(c, beforeQna.UserId); err != nil {
		return err
	}

	form, err := c.MultipartForm()
	if err != nil {
		return factory.ErrorInvalidForm.New(err)
	}

	// MIR-1403: 何かを書き換える前に入力を全部確かめる。紐づけ先はリクエストのcommon_idではなく、更新前のQnAの
	// common_id(DBの値)だけを使う。削除の指定に、このQnAの添付でないID(他人の添付・規約などの添付・存在しない
	// ID)が1つでもあれば400で何も変えない(403は画面が強制ログアウトするため使わない。他人の添付と存在しないIDを
	// 同じ扱いにして、IDがあるかを外から見分けられないようにしている)。
	commonId := beforeQna.CommonId
	deleteIds, err := domain.ParseAttachmentDeleteIds(form.Value)
	if err != nil {
		return factory.ErrorParameter.New(err)
	}
	if commonId == "" && (len(deleteIds) > 0 || len(form.File) > 0) {
		return factory.ErrorParameter.New(errors.New("qna has no common_id"))
	}
	owned, err := domain.Attachment{}.FindAllByResourceIdUnscoped(ctx, commonId)
	if err != nil {
		return factory.ErrorDB.New(err)
	}
	toDelete, unknown := domain.ClassifyAttachmentDeletes(deleteIds, owned)
	if len(unknown) > 0 {
		factory.Logger(ctx).WithFields(logrus.Fields{"qna_id": beforeQna.Id, "attachment_ids": unknown}).
			Warn("attachments_delete contains attachments that do not belong to the qna")
		return factory.ErrorParameter.New(errors.New("invalid attachments_delete"))
	}

	q := &domain.Qna{Id: beforeQna.Id, Title: qi.Title, Content: qi.Content, Actor: actor}
	if _, err := q.UpdateByUser(ctx); err != nil {
		return factory.ErrorDB.New(err)
	}

	for _, a := range toDelete {
		if err := (domain.Attachment{}).DeleteByIdAndResourceId(ctx, a.Id, commonId); err != nil {
			return factory.ErrorDB.New(err)
		}
	}

	attachments, err := domain.Attachment{}.AddAttachmentsExtended(ctx, commonId, form, actor)
	if err != nil {
		return factory.ErrorAWSUpload.New(err)
	}

	// S3のファイルは、DBの書き込みが全部済んだあとで、確定後に消す。
	domain.DeleteAttachmentFilesAfterCommit(ctx, toDelete)

	res := *beforeQna
	res.Title = q.Title
	res.Content = q.Content
	res.Read = false
	res.Actor = actor
	res.Attachments = attachments
	return renderSucc(c, http.StatusOK, &res)
}

func (QnaApiController) Delete(c echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return factory.ErrorParameter.New(err)
	}

	q, err := domain.Qna{}.GetById(c.Request().Context(), id)
	if err != nil {
		return factory.ErrorDB.New(err)
	} else if q == nil {
		return factory.ErrorNotFound.New(errors.New("qna is not found"))
	}

	if err := jwthelper.ValidateUserToken(c, q.UserId); err != nil {
		return err
	}

	err = domain.Qna{}.Delete(c.Request().Context(), id)
	if err != nil {
		return factory.ErrorDB.New(err)
	}

	return renderSucc(c, http.StatusOK, nil)
}
