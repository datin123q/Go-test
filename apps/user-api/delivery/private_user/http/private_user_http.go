package http

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/pangpanglabs/echoswagger/v2"
	"github.com/pkg/errors"

	cdomain "github.com/illuminarean/energy-shares-backend-repo/clean-arch-domain"
	"github.com/illuminarean/energy-shares-backend-repo/domain"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/actorcontext"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/apirender"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/echoutils"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/factory"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/jwthelper"
	udomain "github.com/illuminarean/energyshares-user-api/domain"
)

type privateUserHttpDelivery struct {
	privateUserUsecase udomain.PrivateUserUsecase
	extractor          jwthelper.UserExtractor
}

func New(
	api echoswagger.ApiRoot,
	privateUserUsecase udomain.PrivateUserUsecase,
	extractor jwthelper.UserExtractor,
) *privateUserHttpDelivery {
	handler := &privateUserHttpDelivery{
		privateUserUsecase: privateUserUsecase,
		extractor:          extractor,
	}

	g := api.Group("v1/private-users", "api/v1/private-users")

	g.GET("/signup/step", handler.GetSignUpStatus,
		echoutils.AllowRoles(domain.ROLE_USER)).
		AddParamQuery("", "status", "", false).
		AddResponse(http.StatusOK, "個人会員加入ステータス照会", udomain.PrvUserGetSignUpStatusResponse{}, nil).
		SetSummary("個人会員加入ステータスと詳細情報照会").
		SetDescription("個人会員加入ステータスと詳細情報照会").
		SetSecurity("Authorization")

	g.PUT("/signup/profiles", handler.RegisterProfile,
		echoutils.AllowRoles(domain.ROLE_USER)).
		AddParamBody(udomain.PrvUserRegisterProfileRequest{}, "body", "個人会員情報", true).
		AddResponse(http.StatusOK, "個人会員情報登録成功", "", nil).
		SetSummary("個人会員情報登録").
		SetDescription("個人会員情報登録").
		SetSecurity("Authorization")

	g.POST("/signup/my-number", handler.RegisterMyNumber,
		echoutils.AllowRoles(domain.ROLE_USER)).
		AddParamBody(udomain.PrvUserRegisterMyNumberRequest{}, "body", "My Number情報", true).
		AddResponse(http.StatusOK, "マイナンバー登録成功", udomain.PrvUserRegisterMyNumberResponse{}, nil).
		SetSummary("個人会員My Number登録").
		SetDescription("個人会員My Number登録").
		SetSecurity("Authorization")

	g.POST("/signup/ekyc", handler.VerifyWithEkyc,
		echoutils.AllowRoles(domain.ROLE_USER),
		echoutils.AllowAuditStatuses(domain.PROVISIONAL_AUDIT, domain.WAIT_IDENTITY_REGISTER_AUDIT, domain.HOLD_AUDIT)).
		AddParamBody(udomain.PrvUserVerifyWithEkycRequest{}, "body", "ユーザー情報", true).
		AddResponse(http.StatusOK, "eKYC連携ID", udomain.PrvUserVerifyWithEkycResponse{}, nil).
		SetSummary("eKYC連携ID発行").
		SetDescription("eKYCで使用する連携IDを発行する").
		SetSecurity("Authorization")

	g.PUT("/signup/ekyc", handler.UpdateEkycStatus,
		echoutils.AllowRoles(domain.ROLE_USER),
		echoutils.AllowAuditStatuses(domain.PROVISIONAL_AUDIT, domain.WAIT_IDENTITY_REGISTER_AUDIT, domain.HOLD_AUDIT)).
		AddParamBody(udomain.PrvUserUpdateEkycStatusRequest{}, "body", "eKYC成功の有無", true).
		AddResponse(http.StatusOK, "eKYCステータス保存成功", "", nil).
		SetSummary("eKYCリクエストステータス保存").
		SetDescription("eKYCリクエストのステータス保存").
		SetSecurity("Authorization")

	g.GET("/signup/ekyc/status", handler.GetEkycStatus).
		AddParamQuery("", "user_id", "ユーザーID", true).
		AddResponse(http.StatusOK, "eKYCリクエスト結果", "結果なし: null | 成功: true | 失敗: false", nil).
		SetSummary("eKYCリクエストステータス照会").
		SetDescription("最も最近作成されたeKYCリクエストのステータス照会")

	g.POST("/signup/ekyc/webhook", handler.CallEkycWebhook).
		AddParamBody(udomain.PrvUserCallEkycWebhookRequest{}, "body", "eKYC最終成功の有無", true).
		AddResponse(http.StatusOK, "eKYC結果通知", "", nil).
		SetSummary("テスト環境（DEV,QA）用eKYC結果通知Webhook").
		SetDescription("テスト環境（DEV,QA）用eKYC結果通知Webhook")

	g.GET("/detail", handler.GetOne,
		echoutils.AllowRoles(domain.ROLE_USER)).
		AddResponse(http.StatusOK, "개인 사용자 상세 정보", udomain.PrvUserGetOneResponse{}, nil).
		SetSummary("개인 회원 상세 정보 조회").
		SetDescription("개인 회원 상세 정보를 조회합니다.").
		SetSecurity("Authorization")

	g.PATCH("", handler.PatchOne,
		echoutils.AllowRoles(domain.ROLE_USER)).
		AddParamBody(udomain.PrvUserPatchOneRequest{}, "body", "個人会員更新", true).
		AddResponse(http.StatusOK, "個人会員更新", "", nil).
		SetSummary("個人会員情報を更新します。").
		SetDescription("個人会員情報を更新します。").
		SetSecurity("Authorization")

	return handler
}

func (p privateUserHttpDelivery) GetSignUpStatus(c echo.Context) error {
	user, err := p.extractor.Extract(c)
	if err != nil {
		return factory.ErrorFailAuthorized.New(err)
	}

	ctx := c.Request().Context()

	// Request DTOの作成及びバインド
	request := udomain.PrvUserGetSignUpStatusRequest{
		JWTUser: user,
	}
	if err := c.Bind(&request); err != nil {
		return factory.ErrorParameter.New(err)
	}
	if err := c.Validate(request); err != nil {
		return factory.ErrorParameter.New(err)
	}

	// Usecaseを呼び出してエラーをハンドル
	response, err := p.privateUserUsecase.GetSignUpStatus(ctx, request)
	switch {
	case err == nil:
		return apirender.RenderSucc(c, http.StatusOK, response)
	case errors.Is(err, cdomain.ErrItemNotFound): // ユーザー照会結果がない場合の画面処理用エラー （参考: ESJP-738）
		return factory.ErrorIllegalRequest.New(err)
	case errors.Is(err, cdomain.ErrBadRequest):
		return factory.ErrorParameter.New(err)
	default:
		return factory.ErrorDB.New(err)
	}
}

func (p privateUserHttpDelivery) RegisterProfile(c echo.Context) error {
	user, err := p.extractor.Extract(c)
	if err != nil {
		return factory.ErrorFailAuthorized.New(err)
	}

	ctx := actorcontext.WithActor(c.Request().Context(), user)

	// Request DTOの作成及びバインド
	request := udomain.PrvUserRegisterProfileRequest{
		JWTUser: user,
	}
	if err := c.Bind(&request); err != nil {
		return factory.ErrorParameter.New(err)
	}
	if err := request.Validate(); err != nil {
		return factory.ErrorParameter.New(err)
	}

	// Usecaseを呼び出してエラーをハンドル
	response, err := p.privateUserUsecase.RegisterProfile(ctx, request)
	switch {
	case err == nil:
		return apirender.RenderSucc(c, http.StatusOK, response)
	case errors.Is(err, cdomain.ErrBadRequest):
		return factory.ErrorParameter.New(err)
	case errors.Is(err, cdomain.ErrProfileIncompleteForEkyc):
		return factory.ErrorProfileIncompleteForEkyc.New(err)
	case errors.Is(err, cdomain.ErrServiceUnavailable):
		return factory.ErrorInvalidAssetAndIncome.New(err)
	case errors.Is(err, cdomain.ErrItemConflict):
		return factory.ErrorParameter.New(err)
	default:
		return factory.ErrorDB.New(err)
	}
}

func (p privateUserHttpDelivery) RegisterMyNumber(c echo.Context) error {
	user, err := p.extractor.Extract(c)
	if err != nil {
		return factory.ErrorFailAuthorized.New(err)
	}

	ctx := actorcontext.WithActor(c.Request().Context(), user)

	// Request DTOの作成及びバインド
	request := udomain.PrvUserRegisterMyNumberRequest{
		JWTUser: user,
	}
	if err := c.Bind(&request); err != nil {
		return factory.ErrorParameter.New(err)
	}
	if err := c.Validate(request); err != nil {
		return factory.ErrorParameter.New(err)
	}

	// Usecaseを呼び出してエラーをハンドル
	response, err := p.privateUserUsecase.RegisterMyNumber(ctx, request)
	switch {
	case err == nil:
		return apirender.RenderSucc(c, http.StatusOK, response)
	case errors.Is(err, cdomain.ErrAlreadyExists):
		return factory.ErrorMyNumberAlreadyExist.New(err)
	case errors.Is(err, cdomain.ErrItemConflict):
		return factory.ErrorParameter.New(err)
	case errors.Is(err, cdomain.ErrServiceUnavailable):
		return factory.ErrorServiceUnavailable.New(err)
	default:
		return factory.ErrorDB.New(err)
	}
}

func (p privateUserHttpDelivery) VerifyWithEkyc(c echo.Context) error {
	user, err := p.extractor.Extract(c)
	if err != nil {
		return factory.ErrorFailAuthorized.New(err)
	}
	ctx := actorcontext.WithActor(c.Request().Context(), user)

	// Request DTOの作成及びバインド
	request := udomain.PrvUserVerifyWithEkycRequest{
		JWTUser: user,
	}
	if err := c.Bind(&request); err != nil {
		return factory.ErrorParameter.New(err)
	}
	if err := c.Validate(request); err != nil {
		return factory.ErrorParameter.New(err)
	}

	// Usecaseを呼び出してエラーをハンドル
	response, err := p.privateUserUsecase.VerifyWithEkyc(ctx, request)
	switch {
	case err == nil:
		return apirender.RenderSucc(c, http.StatusOK, response)
	case errors.Is(err, cdomain.ErrForbidden):
		return factory.ErrorFailAuthorized.New(err)
	case errors.Is(err, cdomain.ErrItemConflict):
		return factory.ErrorParameter.New(err)
	default:
		return factory.ErrorDB.New(err)
	}
}

func (p privateUserHttpDelivery) UpdateEkycStatus(c echo.Context) error {
	user, err := p.extractor.Extract(c)
	if err != nil {
		return factory.ErrorFailAuthorized.New(err)
	}

	ctx := actorcontext.WithActor(c.Request().Context(), user)

	// Request DTOの作成及びバインド
	request := udomain.PrvUserUpdateEkycStatusRequest{
		JWTUser: user,
	}
	if err := c.Bind(&request); err != nil {
		return factory.ErrorParameter.New(err)
	}
	if err := c.Validate(request); err != nil {
		return factory.ErrorParameter.New(err)
	}

	// Usecaseを呼び出してエラーをハンドル
	err = p.privateUserUsecase.UpdateEkycStatus(ctx, request)
	switch {
	case err == nil:
		return apirender.RenderSucc(c, http.StatusOK, nil)
	case errors.Is(err, cdomain.ErrForbidden):
		return factory.ErrorFailAuthorized.New(err)
	case errors.Is(err, cdomain.ErrItemConflict):
		return factory.ErrorParameter.New(err)
	case errors.Is(err, cdomain.ErrBadRequest):
		return factory.ErrorParameter.New(err)
	default:
		return factory.ErrorDB.New(err)
	}
}

func (p privateUserHttpDelivery) GetEkycStatus(c echo.Context) error {
	ctx := c.Request().Context()

	// Request DTOの作成及びバインド
	request := udomain.PrvUserGetEkycStatusRequest{}
	if err := c.Bind(&request); err != nil {
		return factory.ErrorParameter.New(err)
	}
	if err := c.Validate(request); err != nil {
		return factory.ErrorParameter.New(err)
	}

	// Usecaseを呼び出してエラーをハンドル
	response, err := p.privateUserUsecase.GetEkycStatus(ctx, request)
	switch {
	case err == nil:
		return apirender.RenderSucc(c, http.StatusOK, response)
	default:
		return factory.ErrorDB.New(err)
	}
}

func (p privateUserHttpDelivery) CallEkycWebhook(c echo.Context) error {
	ctx := c.Request().Context()

	// Request DTOの作成及びバインド
	request := udomain.PrvUserCallEkycWebhookRequest{}
	if err := c.Bind(&request); err != nil {
		return factory.ErrorParameter.New(err)
	}
	if err := c.Validate(request); err != nil {
		return factory.ErrorParameter.New(err)
	}

	// Usecaseを呼び出してエラーをハンドル
	err := p.privateUserUsecase.CallEkycWebhook(ctx, request)
	switch {
	case err == nil:
		return apirender.RenderSucc(c, http.StatusOK, nil)
	default:
		return factory.ErrorDB.New(err)
	}
}

func (p privateUserHttpDelivery) GetOne(c echo.Context) error {
	user, err := p.extractor.Extract(c)
	if err != nil {
		return factory.ErrorFailAuthorized.New(err)
	}

	ctx := c.Request().Context()

	// Request DTOの作成及びバインド
	request := udomain.PrvUserGetOneRequest{
		JWTUser: user,
	}

	// Usecaseを呼び出してエラーをハンドル
	response, err := p.privateUserUsecase.GetOne(ctx, request)
	switch {
	case err == nil:
		return apirender.RenderSucc(c, http.StatusOK, response)
	case errors.Is(err, cdomain.ErrItemNotFound):
		return factory.ErrorNotFound.New(err)
	default:
		return factory.ErrorDB.New(err)
	}
}

func (u privateUserHttpDelivery) PatchOne(c echo.Context) error {
	user, err := u.extractor.Extract(c)
	if err != nil {
		return factory.ErrorFailAuthorized.New(err)
	}

	ctx := actorcontext.WithActor(c.Request().Context(), user)

	// Request DTOの作成及びバインド
	request := udomain.PrvUserPatchOneRequest{
		JWTUser: user,
	}
	if err := c.Bind(&request); err != nil {
		return factory.ErrorParameter.New(err)
	}
	if err := c.Validate(request); err != nil {
		return factory.ErrorParameter.New(err)
	}

	err = u.privateUserUsecase.PatchOne(ctx, request)
	switch {
	case err == nil:
		return apirender.RenderSucc(c, http.StatusOK, "")
	case errors.Is(err, cdomain.ErrItemNotFound):
		return factory.ErrorNotFound.New(err)
	case errors.Is(err, cdomain.ErrItemConflict):
		return factory.ErrorInvalidUserType.New(err)
	default:
		return factory.ErrorDB.New(err)
	}
}
