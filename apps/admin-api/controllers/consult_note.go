package controllers

import (
	"github.com/pkg/errors"
	"net/http"
	"strconv"

	"github.com/illuminarean/energy-shares-backend-repo/domain"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/apirender"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/echoutils"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/factory"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/jwthelper"
	"github.com/labstack/echo/v4"
	"github.com/pangpanglabs/echoswagger/v2"
)

type ConsultNoteApiController struct{}

// MIR-1408: 権限はルートごとに付ける。グループ(接頭辞 "api")に付けると、/api 配下の存在しない URL まで 403 になる。
func (c ConsultNoteApiController) Init(g echoswagger.ApiGroup) {
	g.GET("/users/:id/consult-notes", c.GetAllByUserId,
		echoutils.AllowRoles(domain.ROLE_INVESTOR)).
		AddParamQueryNested(SearchInput{}).
		AddResponse(http.StatusOK, "ユーザーの相談メモリスト照会", domain.ConsultNote{}, nil).
		SetDescription("ユーザーの相談メモリストを照会します.").SetSummary("ユーザーの相談メモリスト照会").
		SetSecurity("Authorization")

	g.POST("/users/:id/consult-notes", c.Create,
		echoutils.AllowRoles(domain.ROLE_INVESTOR)).
		AddParamBody(ConsultNoteInput{}, "body", "", true).
		AddResponse(http.StatusCreated, "相談メモ追加", domain.ConsultNote{}, nil).
		SetDescription("相談メモを追加します.").SetSummary("相談メモ追加").
		SetSecurity("Authorization")

	g.GET("/users/:id/consult-notes/:note_id", c.GetOne,
		echoutils.AllowRoles(domain.ROLE_INVESTOR)).
		AddParamPath("", "id", "ユーザーID").
		AddParamPath("", "note_id", "相談メモID").
		AddResponse(http.StatusOK, "相談メモ詳細照会成功", domain.ConsultNote{}, nil).
		SetDescription("相談メモ詳細照会").SetSummary("相談メモ詳細照会").
		SetSecurity("Authorization")

	g.PUT("/users/:id/consult-notes/:note_id", c.Update,
		echoutils.AllowRoles(domain.ROLE_INVESTOR)).
		AddParamPath("", "id", "ユーザーID").
		AddParamPath("", "note_id", "相談メモID").
		AddParamBody(ConsultNoteInput{}, "body", "", true).
		AddResponse(http.StatusOK, "相談メモ情報修正", domain.ConsultNote{}, nil).
		SetDescription("相談メモを修正します.").SetSummary("相談メモ修正").
		SetSecurity("Authorization")

	g.DELETE("/users/:id/consult-notes/:note_id", c.Delete,
		echoutils.AllowRoles(domain.ROLE_INVESTOR)).
		AddParamPath("", "id", "ユーザーID").
		AddParamPath("", "note_id", "相談メモID").
		AddResponse(http.StatusOK, "相談メモ情報削除", domain.ConsultNote{}, nil).
		SetDescription("相談メモを削除します.").SetSummary("相談メモ削除").
		SetSecurity("Authorization")
}

func (ConsultNoteApiController) GetAllByUserId(c echo.Context) error {
	var s SearchInput
	if err := c.Bind(&s); err != nil {
		return factory.ErrorParameter.New(err)
	}
	if s.MaxResultCount == 0 {
		s.MaxResultCount = DefaultMaxResultCount
	}

	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return factory.ErrorParameter.New(err)
	}

	totalCount, items, err := domain.ConsultNote{}.GetAllByUserId(c.Request().Context(), id, s.Sortby, s.Order, s.SkipCount, s.MaxResultCount)
	if err != nil {
		return factory.ErrorDB.New(err)
	}

	return apirender.RenderSucc(c, http.StatusOK, apirender.ArrayResult{
		Items:      items,
		TotalCount: totalCount,
	})
}

func (ConsultNoteApiController) Create(c echo.Context) error {

	var cni ConsultNoteInput
	if err := c.Bind(&cni); err != nil {
		return factory.ErrorParameter.New(err)
	}
	if err := c.Validate(&cni); err != nil {
		return factory.ErrorParameter.New(err)
	}

	cn := cni.ToModel()
	if cn == nil {
		return factory.ErrorParameter.New(errors.New("consult note is nil"))
	}

	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return factory.ErrorParameter.New(err)
	}
	cn.UserId = id

	loginId, _, err := jwthelper.TokenToUser(c)
	if err != nil {
		return factory.ErrorTokenInvalid.New(err)
	}

	loginName, err := jwthelper.GetNameFromToken(c)
	if err != nil {
		return factory.ErrorTokenInvalid.New(err)
	}

	cn.OwnerId = loginId
	cn.ModifierId = loginId
	cn.Owner = loginName

	if _, err := cn.Create(c.Request().Context()); err != nil {
		return factory.ErrorDB.New(err)
	}

	return apirender.RenderSucc(c, http.StatusCreated, cn)
}

func (ConsultNoteApiController) GetOne(c echo.Context) error {
	noteId, err := strconv.ParseInt(c.Param("note_id"), 10, 64)
	if err != nil {
		return factory.ErrorParameter.New(err)
	}

	consultNote, err := domain.ConsultNote{}.GetById(c.Request().Context(), noteId)
	if err != nil {
		return factory.ErrorDB.New(err)
	}
	if consultNote == nil {
		return factory.ErrorNotFound.New(err)
	}

	return apirender.RenderSucc(c, http.StatusOK, consultNote)
}

func (ConsultNoteApiController) Update(c echo.Context) error {
	var cni ConsultNoteUpdateInput
	if err := c.Bind(&cni); err != nil {
		return factory.ErrorParameter.New(err)
	}
	if err := c.Validate(&cni); err != nil {
		return factory.ErrorParameter.New(err)
	}

	cn := cni.ToModel()
	if cn == nil {
		return factory.ErrorParameter.New(errors.New("consult note is nil"))
	}

	noteId, err := strconv.ParseInt(c.Param("note_id"), 10, 64)
	if err != nil {
		return factory.ErrorParameter.New(err)
	}
	cn.Id = noteId

	loginId, _, err := jwthelper.TokenToUser(c)
	if err != nil {
		return factory.ErrorTokenInvalid.New(err)
	}

	loginName, err := jwthelper.GetNameFromToken(c)
	if err != nil {
		return factory.ErrorTokenInvalid.New(err)
	}

	cn.ModifierId = loginId
	cn.Owner = loginName

	if _, err := cn.Update(c.Request().Context()); err != nil {
		return factory.ErrorDB.New(err)
	}

	updatedConsultNote, err := domain.ConsultNote{}.GetById(c.Request().Context(), noteId)
	if err != nil {
		return factory.ErrorDB.New(err)
	}
	if updatedConsultNote == nil {
		return factory.ErrorNotFound.New(err)
	}

	return apirender.RenderSucc(c, http.StatusOK, updatedConsultNote)
}

func (ConsultNoteApiController) Delete(c echo.Context) error {
	noteId, err := strconv.ParseInt(c.Param("note_id"), 10, 64)
	if err != nil {
		return factory.ErrorParameter.New(err)
	}

	consultNote, err := domain.ConsultNote{}.GetById(c.Request().Context(), noteId)
	if err != nil {
		return factory.ErrorDB.New(err)
	}
	if consultNote == nil {
		return factory.ErrorNotFound.New(err)
	}

	err = consultNote.Delete(c.Request().Context())
	if err != nil {
		return factory.ErrorDB.New(err)
	}

	return apirender.RenderSucc(c, http.StatusOK, nil)
}
