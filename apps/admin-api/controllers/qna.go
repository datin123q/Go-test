package controllers

import (
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/illuminarean/energy-shares-backend-repo/domain"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/apirender"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/jwthelper"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/serviceutils"

	"github.com/labstack/echo/v4"
	"github.com/pangpanglabs/echoswagger/v2"
	"github.com/sirupsen/logrus"

	"github.com/illuminarean/energy-shares-backend-repo/pkg/factory"
)

type QnaApiController struct {
}

func (q QnaApiController) Init(g echoswagger.ApiGroup) {

	g.GET("/list", q.GetAllAdmin).
		AddParamQueryNested(QnaSearchInput{}).
		SetDescription("1:1問い合わせリスト情報をリクエストしています。(admin)").
		SetSummary("1:1問い合わせリスト情報リクエスト(admin)").
		SetSecurity("Authorization")

	g.GET("/one/:id", q.GetOneAdmin).
		AddParamPath("", "id", "").
		AddResponse(http.StatusOK, "特定の1:1問い合わせ情報照会(admin)", domain.Qna{}, nil).
		SetDescription("1:1問い合わせを照会します。(admin)").
		SetSummary("1:1問い合わせ照会(admin)").
		SetSecurity("Authorization")

	g.POST("/:id/answer", q.CreateAnswer).
		AddParamPath("", "id", "").
		AddParamForm(AnswerInput{}, "body", "", false).
		AddResponse(http.StatusCreated, "1:1問い合わせ回答作成", domain.Answer{}, nil).
		SetRequestContentType("multipart/form-data").
		SetDescription("1:1問い合わせ回答を追加します。").
		SetSummary("1:1問い合わせ回答追加").
		SetSecurity("Authorization")

	g.PUT("/:id/answer/:answerId", q.UpdateAnswer).
		AddParamPath("", "id", "").
		AddParamPath("", "answerId", "").
		AddParamForm(AnswerInput{}, "body", "", false).
		AddResponse(http.StatusOK, "1:1問い合わせ回答修正", domain.Answer{}, nil).
		SetRequestContentType("multipart/form-data").
		SetDescription("1:1問い合わせ回答を修正します。").
		SetSummary("1:1問い合わせ回答修正").
		SetSecurity("Authorization")

}

func (QnaApiController) GetAllAdmin(c echo.Context) error {

	var v QnaSearchInput
	if err := c.Bind(&v); err != nil {
		return factory.ErrorParameter.New(err)
	}
	if v.MaxResultCount == 0 {
		v.MaxResultCount = DefaultMaxResultCount
	}

	userId, _, err := jwthelper.TokenToUser(c)
	if err != nil {
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

	err = domain.OperatorHistory{}.Record(c.Request().Context(), userId, "",
		"menu.manage_qna.list", "", c.RealIP(), domain.LOOKUP)
	if err != nil {
		return factory.ErrorDB.New(err)
	}

	return apirender.RenderSucc(c, http.StatusOK, apirender.ArrayResult{
		TotalCount: totalCount,
		Items:      items,
	})
}

func (QnaApiController) GetOneAdmin(c echo.Context) error {

	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return factory.ErrorParameter.New(err)
	}

	userId, _, err := jwthelper.TokenToUser(c)
	if err != nil {
		return err
	}

	q, err := domain.Qna{}.GetById(c.Request().Context(), id)
	if err != nil {
		return factory.ErrorDB.New(err)
	}
	if q == nil {
		return factory.ErrorNotFound.New(err)
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

	// 質問読了処理
	if q.Status == domain.QNA_RECEIVED {
		q.Status = domain.QNA_RESPONSE_PENDING
		if _, err = q.Update(c.Request().Context()); err != nil {
			return factory.ErrorDB.New(err)
		}
	}

	err = domain.OperatorHistory{}.Record(c.Request().Context(), userId, "",
		"menu.manage_qna.detail", q.Title, c.RealIP(), domain.LOOKUP)
	if err != nil {
		return factory.ErrorDB.New(err)
	}

	return apirender.RenderSucc(c, http.StatusOK, q)
}

func (QnaApiController) CreateAnswer(c echo.Context) error {
	userId, userName, userEmail, err := jwthelper.TokenToUserExtended(c)
	if err != nil {
		return factory.ErrorTokenInvalid.New(err)
	}

	form, err := c.MultipartForm()
	if err != nil {
		return factory.ErrorInvalidForm.New(err)
	}

	var ai AnswerInput
	if err := c.Bind(&ai); err != nil {
		return factory.ErrorParameter.New(err)
	}
	if err := c.Validate(&ai); err != nil {
		return factory.ErrorParameter.New(err)
	}

	a, err := ai.ToModel()
	if err != nil {
		return factory.ErrorParameter.New(err)
	}

	actor := domain.ChangelogActor{
		Id:    userId,
		Name:  userName,
		Email: userEmail,
	}
	a.Actor = actor

	qnaId, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return factory.ErrorParameter.New(err)
	}
	a.QnaId = qnaId

	beforeQna, err := domain.Qna{}.GetById(c.Request().Context(), qnaId)
	if err != nil {
		return factory.ErrorDB.New(err)
	} else if beforeQna == nil {
		return factory.ErrorNotFound.New(errors.New("Not Found Q&A"))
	}

	if _, err := a.CreateAnswer(c.Request().Context()); err != nil {
		return factory.ErrorDB.New(err)
	}

	a.Attachments, err = domain.Attachment{}.AddAttachmentsExtended(c.Request().Context(), a.CommonId, form, actor)
	if err != nil {
		return factory.ErrorAWSUpload.New(err)
	}

	qna, err := domain.Qna{}.GetById(c.Request().Context(), a.QnaId)
	if err != nil {
		return factory.ErrorDB.New(err)
	}
	qna.AnswerCount++
	qna.Status = domain.QNA_RESPONSE_COMPLETED // 回答完了
	if _, err = qna.Update(c.Request().Context()); err != nil {
		return factory.ErrorDB.New(err)
	}

	var mail domain.Mail
	err = mail.QnaMail(c.Request().Context(), serviceutils.GetServiceInfoMail(), qna.Email, qna.Name, qna.Title,
		qna.Created.Time().Format(time.DateTime),
		qna.Answers[0].Created.Time().Format(time.DateTime), qna.Id)
	if err != nil {
		// MIR-1365: 宛先・氏名は出さない。Titleも利用者が自由に書ける文字列で
		// 氏名等が含まれうるため出さない。Idから対象を特定できる。
		factory.Logger(c.Request().Context()).WithFields(logrus.Fields{
			"err": err,
			"Id":  qna.Id,
		}).Error("fail Qna mail")
	}

	err = domain.OperatorHistory{}.Record(c.Request().Context(), userId, "",
		"menu.manage_qna.detail", strconv.FormatInt(a.QnaId, 10)+"'s answer ("+strconv.FormatInt(a.Id, 10)+")", c.RealIP(), domain.REGISTER)
	if err != nil {
		return factory.ErrorDB.New(err)
	}

	afterQna, err := domain.Qna{}.GetById(c.Request().Context(), qnaId)
	if err != nil {
		return factory.ErrorDB.New(err)
	}
	_, err = domain.QnaInfoHistory{}.Create(c.Request().Context(), beforeQna, afterQna, userId)
	if err != nil {
		return factory.ErrorDB.New(err)
	}

	return apirender.RenderSucc(c, http.StatusOK, a)
}

func (QnaApiController) UpdateAnswer(c echo.Context) error {
	userId, userName, userEmail, err := jwthelper.TokenToUserExtended(c)
	if err != nil {
		return factory.ErrorTokenInvalid.New(err)
	}

	var ai AnswerInput
	if err := c.Bind(&ai); err != nil {
		return factory.ErrorParameter.New(err)
	}
	if err := c.Validate(&ai); err != nil {
		return factory.ErrorParameter.New(err)
	}

	a, err := ai.ToModel()
	if err != nil {
		return factory.ErrorParameter.New(err)
	}

	actor := domain.ChangelogActor{
		Id:    userId,
		Name:  userName,
		Email: userEmail,
	}
	a.Actor = actor

	qnaId, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return factory.ErrorParameter.New(err)
	}
	a.QnaId = qnaId

	beforeQna, err := domain.Qna{}.GetById(c.Request().Context(), qnaId)
	if err != nil {
		return factory.ErrorDB.New(err)
	} else if beforeQna == nil {
		return factory.ErrorNotFound.New(errors.New("Not Found Q&A"))
	}

	id, err := strconv.ParseInt(c.Param("answerId"), 10, 64)
	if err != nil {
		return factory.ErrorParameter.New(err)
	}
	a.Id = id

	// MIR-1403: 回答は、このQnAの回答の中から探す(別のQnAの回答のIDを渡されても書き換えない)。管理画面は回答の
	// common_idを送らないため、削除の判定と新しい添付の紐づけ先には、DBの回答のcommon_idを使う(以前は空のまま
	// 保存され、回答の編集で付けた添付がどこにも表示されなかった)。
	beforeAnswer := findAnswerById(beforeQna.Answers, id)
	if beforeAnswer == nil {
		return factory.ErrorNotFound.New(errors.New("Not Found Answer"))
	}
	a.CommonId = beforeAnswer.CommonId

	form, err := c.MultipartForm()
	if err != nil {
		return factory.ErrorInvalidForm.New(err)
	}
	if a.CommonId == "" && len(form.File) > 0 {
		return factory.ErrorParameter.New(errors.New("answer has no common_id"))
	}

	attachmentFilesToDelete, err := deleteOwnedAttachments(c.Request().Context(), form.Value, a.CommonId,
		logrus.Fields{"user": userId, "qna_id": qnaId, "answer_id": id})
	if err != nil {
		return err
	}

	a.Attachments, err = domain.Attachment{}.AddAttachmentsExtended(c.Request().Context(), a.CommonId, form, actor)
	if err != nil {
		return factory.ErrorAWSUpload.New(err)
	}

	if _, err := a.UpdateAnswer(c.Request().Context()); err != nil {
		return factory.ErrorDB.New(err)
	}

	err = domain.OperatorHistory{}.Record(c.Request().Context(), userId, "",
		"menu.manage_qna.detail", strconv.FormatInt(a.QnaId, 10)+"'answer ("+strconv.FormatInt(a.Id, 10)+")", c.RealIP(), domain.CHANGE)
	if err != nil {
		return factory.ErrorDB.New(err)
	}

	afterQna, err := domain.Qna{}.GetById(c.Request().Context(), qnaId)
	if err != nil {
		return factory.ErrorDB.New(err)
	}
	_, err = domain.QnaInfoHistory{}.Create(c.Request().Context(), beforeQna, afterQna, userId)
	if err != nil {
		return factory.ErrorDB.New(err)
	}

	domain.DeleteAttachmentFilesAfterCommit(c.Request().Context(), attachmentFilesToDelete)

	return apirender.RenderSucc(c, http.StatusOK, a)
}

func findAnswerById(answers []*domain.Answer, id int64) *domain.Answer {
	for _, a := range answers {
		if a != nil && a.Id == id {
			return a
		}
	}
	return nil
}
