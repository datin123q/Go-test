package controllers

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/illuminarean/energy-shares-backend-repo/domain"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/apirender"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/factory"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/jwthelper"
	"github.com/labstack/echo/v4"
	"github.com/pangpanglabs/echoswagger/v2"
)

type NoticeMsgApiController struct {
}

func (c NoticeMsgApiController) Init(g echoswagger.ApiGroup) {
	g.GET("", c.GetAll).
		AddParamQueryNested(NoticeSearchInput{}).
		SetDescription("ユーザー通知リスト情報を要求します.").
		SetSummary("ユーザー通知リスト情報要求").
		SetSecurity("Authorization")

	g.GET("/:id", c.GetOne).
		AddParamPath("", "id", "").
		SetDescription("ユーザー通知情報を要求します.").
		SetSummary("ユーザー通知情報要求").
		SetSecurity("Authorization")

	g.POST("/send", c.Create).
		AddParamForm(NoticeSendInput{}, "formData", "", false).
		SetRequestContentType("multipart/form-data").
		AddResponse(http.StatusCreated, "ユーザーニュース情報追加", domain.NoticeMsg{}, nil).
		SetDescription("ユーザー通知情報を追加します.").
		SetSummary("ユーザー通知情報を追加").
		SetSecurity("Authorization")

	g.POST("/send/projects", c.CreateByProject).
		AddParamForm(NoticeSendProjectInput{}, "formData", "", false).
		SetRequestContentType("multipart/form-data").
		AddResponse(http.StatusCreated, "投資会員ニュース情報追加", domain.NoticeMsg{}, nil).
		SetDescription("ユーザー通知情報を追加します.").
		SetSummary("ユーザー通知情報を追加").
		SetSecurity("Authorization")

	g.POST("/send/newsletter", c.CreateByNewsletter).
		AddParamForm(NoticeSendInput{}, "formData", "", false).
		SetRequestContentType("multipart/form-data").
		AddResponse(http.StatusCreated, "購読ユーザー通知情報追加", domain.NoticeMsg{}, nil).
		SetDescription("ニュースレターを許可したユーザーに送信する通知情報を追加します.").
		SetSummary("ニュースレター許可ユーザー通知情報を追加").
		SetSecurity("Authorization")

	g.POST("/send/user-type", c.CreateByUserType).
		AddParamForm(NoticeSendUserTypeInput{}, "formData", "", false).
		SetRequestContentType("multipart/form-data").
		AddResponse(http.StatusCreated, "ユーザータイプ別ニュース情報追加", domain.NoticeMsg{}, nil).
		SetDescription("ユーザータイプ別通知情報を追加します.").
		SetSummary("ユーザータイプ別通知情報を追加").
		SetSecurity("Authorization")
}

func (NoticeMsgApiController) GetAll(c echo.Context) error {

	var v NoticeSearchInput
	if err := c.Bind(&v); err != nil {
		return factory.ErrorParameter.New(err)
	}
	if v.MaxResultCount == 0 {
		v.MaxResultCount = DefaultMaxResultCount
	}

	totalCount, items, err := domain.NoticeMsg{}.GetAll(c.Request().Context(), v.Title, v.StartDate, v.EndDate, v.Sortby, v.Order, v.SkipCount, v.MaxResultCount)
	if err != nil {
		return factory.ErrorDB.New(err)
	}

	return apirender.RenderSucc(c, http.StatusOK, apirender.ArrayResult{
		TotalCount: totalCount,
		Items:      items,
	})
}

func (NoticeMsgApiController) GetOne(c echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return factory.ErrorParameter.New(err)
	}

	_, _, err = jwthelper.TokenToUser(c)
	if err != nil {
		return err
	}

	noticeMsg, err := domain.NoticeMsg{}.GetById(c.Request().Context(), id)
	if err != nil {
		return factory.ErrorDB.New(err)
	}
	if noticeMsg == nil {
		return factory.ErrorNotFound.New(err)
	}

	return apirender.RenderSucc(c, http.StatusOK, noticeMsg)
}

func (NoticeMsgApiController) Create(c echo.Context) error {
	form, err := c.MultipartForm()
	if err != nil {
		return factory.ErrorInvalidForm.New(err)
	}
	var noticeSendInput NoticeSendInput
	if err := c.Bind(&noticeSendInput); err != nil {
		return factory.ErrorParameter.New(err)
	}
	if err := c.Validate(&noticeSendInput); err != nil {
		return factory.ErrorParameter.New(err)
	}

	noticeMsg, err := noticeSendInput.ToModel(domain.NOTICE_DIRECT)
	if err != nil {
		return factory.ErrorParameter.New(err)
	}
	if noticeMsg == nil {
		return factory.ErrorNotFound.New(err)
	}
	if _, err := noticeMsg.Create(c.Request().Context()); err != nil {
		return factory.ErrorDB.New(err)
	}

	for key, value := range form.Value {
		if strings.Contains(key, "user_id_list") {
			userId, err := strconv.ParseInt(value[0], 10, 64)
			if err != nil {
				return factory.ErrorParameter.New(err)
			}
			mapping := domain.UserNoticeMapping{
				MsgId:  noticeMsg.Id,
				UserId: userId,
			}
			if _, err := mapping.Create(c.Request().Context()); err != nil {
				return factory.ErrorDB.New(err)
			}

		}
	}

	return apirender.RenderSucc(c, http.StatusCreated, noticeMsg)
}

func (NoticeMsgApiController) CreateByProject(c echo.Context) error {
	var noticeSendProjectInput NoticeSendProjectInput
	if err := c.Bind(&noticeSendProjectInput); err != nil {
		return factory.ErrorParameter.New(err)
	}
	if err := c.Validate(&noticeSendProjectInput); err != nil {
		return factory.ErrorParameter.New(err)
	}

	p, err := domain.Project{}.GetById(c.Request().Context(), noticeSendProjectInput.ProjectId)
	if err != nil {
		return factory.ErrorDB.New(err)
	}
	if p == nil {
		return factory.ErrorNotFound.New(nil)
	}

	noticeMsg, err := noticeSendProjectInput.ToModel(domain.NOTICE_PROJECT)
	if err != nil {
		return factory.ErrorParameter.New(err)
	}
	if noticeMsg == nil {
		return factory.ErrorNotFound.New(err)
	}
	if _, err := noticeMsg.Create(c.Request().Context()); err != nil {
		return factory.ErrorDB.New(err)
	}

	investors, err2 := p.GetInvestors(c.Request().Context())
	if err2 != nil {
		return factory.ErrorDB.New(err)
	}

	if len(investors) == 0 {
		return factory.ErrorNotFound.New(nil)
	}

	for _, value := range investors {
		mapping := domain.UserNoticeMapping{
			MsgId:  noticeMsg.Id,
			UserId: value.Id,
		}
		if _, err := mapping.Create(c.Request().Context()); err != nil {
			return factory.ErrorDB.New(err)
		}
	}

	return apirender.RenderSucc(c, http.StatusCreated, noticeMsg)
}

func (NoticeMsgApiController) CreateByNewsletter(c echo.Context) error {
	var noticeSendInput NoticeSendInput
	if err := c.Bind(&noticeSendInput); err != nil {
		return factory.ErrorParameter.New(err)
	}
	if err := c.Validate(&noticeSendInput); err != nil {
		return factory.ErrorParameter.New(err)
	}

	users, err := domain.User{}.GetAllNewsletterAllowed(c.Request().Context())
	if err != nil {
		return factory.ErrorDB.New(err)
	}
	if users == nil {
		return factory.ErrorNotFound.New(nil)
	}

	noticeMsg, err := noticeSendInput.ToModel(domain.NOTICE_NEWSLETTER)
	if err != nil {
		return factory.ErrorParameter.New(err)
	}
	if noticeMsg == nil {
		return factory.ErrorNotFound.New(err)
	}
	if _, err := noticeMsg.Create(c.Request().Context()); err != nil {
		return factory.ErrorDB.New(err)
	}

	for _, value := range users {
		mapping := domain.UserNoticeMapping{
			MsgId:  noticeMsg.Id,
			UserId: value.Id,
		}
		if _, err := mapping.Create(c.Request().Context()); err != nil {
			return factory.ErrorDB.New(err)
		}
	}

	return apirender.RenderSucc(c, http.StatusCreated, noticeMsg)
}

func (NoticeMsgApiController) CreateByUserType(c echo.Context) error {
	var noticeSendUserTypeInput NoticeSendUserTypeInput
	if err := c.Bind(&noticeSendUserTypeInput); err != nil {
		return factory.ErrorParameter.New(err)
	}
	if err := c.Validate(&noticeSendUserTypeInput); err != nil {
		return factory.ErrorParameter.New(err)
	}

	userType := strings.Split(noticeSendUserTypeInput.UserType, ",")

	users, err := domain.User{}.GetByUserType(c.Request().Context(), userType)
	if err != nil {
		return factory.ErrorDB.New(err)
	}
	if users == nil {
		return factory.ErrorNotFound.New(nil)
	}

	noticeMsg, err := noticeSendUserTypeInput.ToModel(domain.NOTICE_USER)
	if err != nil {
		return factory.ErrorParameter.New(err)
	}
	if noticeMsg == nil {
		return factory.ErrorNotFound.New(err)
	}
	if _, err := noticeMsg.Create(c.Request().Context()); err != nil {
		return factory.ErrorDB.New(err)
	}

	for _, value := range users {
		mapping := domain.UserNoticeMapping{
			MsgId:  noticeMsg.Id,
			UserId: value.Id,
		}
		if _, err := mapping.Create(c.Request().Context()); err != nil {
			return factory.ErrorDB.New(err)
		}
	}

	return apirender.RenderSucc(c, http.StatusCreated, noticeMsg)
}
