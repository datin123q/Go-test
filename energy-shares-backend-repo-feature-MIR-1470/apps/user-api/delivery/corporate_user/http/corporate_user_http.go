package http

import (
	"net/http"

	"github.com/pkg/errors"
	openapigo "github.com/swaggest/openapi-go"

	cdomain "github.com/illuminarean/energy-shares-backend-repo/clean-arch-domain"
	"github.com/illuminarean/energy-shares-backend-repo/domain"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/actorcontext"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/apirender"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/echoutils"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/factory"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/jwthelper"
	udomain "github.com/illuminarean/energyshares-user-api/domain"
	"github.com/labstack/echo/v4"
)

type corporateUserHttpDelivery struct {
	corporateUserUsecase udomain.CorporateUserUsecase
	extractor            jwthelper.UserExtractor
}

func New(
	api *echoutils.EchoSwagger,
	extractor jwthelper.UserExtractor,
	corporateUserUsecase udomain.CorporateUserUsecase,
) *corporateUserHttpDelivery {
	handler := &corporateUserHttpDelivery{
		extractor:            extractor,
		corporateUserUsecase: corporateUserUsecase,
	}

	g := api.Group("/api/v1/corporate-users", echoutils.AllowRoles(domain.ROLE_USER))

	g.GET("/signup/step", handler.GetSignUpStatus).
		SetTags("法人会員").
		SetSummary("法人会員加入ステータス照会").
		AddRequest(udomain.CorporateUserGetSignUpStatusRequest{}).
		AddResponse(udomain.CorporateUserGetSignUpStatusResponse{}, func(cu *openapigo.ContentUnit) {
			cu.HTTPStatus = http.StatusOK
			cu.Description = "照会成功応答値"
		}).
		AddErrors(
			factory.ErrorFailAuthorized,
			factory.ErrorIllegalRequest,
			factory.ErrorParameter,
			factory.ErrorDB,
		).
		AddSecurity("Authorization")

	g.PUT("/signup/profiles", handler.RegisterProfiles).
		SetTags("法人会員").
		SetSummary("法人会員プロフィール登録").
		AddRequest(udomain.CorporateUserRegisterProfilesRequest{}).
		AddResponse(nil, func(cu *openapigo.ContentUnit) {
			cu.HTTPStatus = http.StatusNoContent
			cu.Description = "登録成功応答値"
		}).
		AddErrors(
			factory.ErrorFailAuthorized,
			factory.ErrorSameCorpExists,
			factory.ErrorParameter,
			factory.ErrorDB,
		).
		AddSecurity("Authorization")

	g.POST("/signup/kyc", handler.VerifyWithKyc).
		SetTags("法人会員").
		SetSummary("法人会員KYC用ファイル登録").
		AddRequest(udomain.CorporateUserVerifyWithKycRequest{}).
		AddResponse(nil, func(cu *openapigo.ContentUnit) {
			cu.HTTPStatus = http.StatusNoContent
			cu.Description = "登録成功応答値"
		}).
		AddErrors(
			factory.ErrorFailAuthorized,
			factory.ErrorParameter,
			factory.ErrorDB,
		).
		AddSecurity("Authorization")

	g.POST("/verify-postal-code", handler.VerifyPostalCode).
		SetTags("法人会員").
		SetSummary("法人会員郵便コード認証").
		AddRequest(udomain.CorporateUserVerifyPostalCodeRequest{}).
		AddResponse(nil, func(cu *openapigo.ContentUnit) {
			cu.HTTPStatus = http.StatusNoContent
			cu.Description = "コード認証成功応答値"
		}).
		AddErrors(
			factory.ErrorFailAuthorized,
			factory.ErrorParameter,
			factory.ErrorInvalidAuditStatus,
			factory.ErrorCodeMismatch,
			factory.ErrorNotFound,
			factory.ErrorDB,
		).
		AddSecurity("Authorization")

	g.GET("/detail", handler.GetOne).
		SetTags("法人会員").
		SetSummary("法人会員情報詳細").
		AddRequest(udomain.CorporateUserGetOneRequest{}).
		AddResponse(udomain.CorporateUserGetOneResponse{}, func(cu *openapigo.ContentUnit) {
			cu.HTTPStatus = http.StatusOK
			cu.Description = "照会成功応答値"
		}).
		AddErrors(
			factory.ErrorFailAuthorized,
			factory.ErrorNotMatchAudit,
			factory.ErrorNotFound,
			factory.ErrorDB,
		).
		AddSecurity("Authorization")

	g.PATCH("", handler.PatchOne).
		SetTags("法人会員").
		SetSummary("法人会員情報修正").
		AddRequest(udomain.CorporateUserPatchOneRequest{}).
		AddResponse(nil, func(cu *openapigo.ContentUnit) {
			cu.HTTPStatus = http.StatusNoContent
			cu.Description = "情報修正成功応答値"
		}).
		AddErrors(
			factory.ErrorFailAuthorized,
			factory.ErrorParameter,
			factory.ErrorInvalidUserType,
			factory.ErrorNotFound,
			factory.ErrorDB,
		).
		AddSecurity("Authorization")

	g.GET("/documents", handler.GetDocuments).
		SetTags("法人会員").
		SetSummary("法人会員が登録したファイルリスト照会").
		AddRequest(udomain.CorporateUserGetDocumentsRequest{}).
		AddResponse(udomain.CorporateUserGetDocumentsResponse{}, func(cu *openapigo.ContentUnit) {
			cu.HTTPStatus = http.StatusOK
			cu.Description = "照会成功応答値"
		}).
		AddErrors(
			factory.ErrorFailAuthorized,
			factory.ErrorDB,
		).
		AddSecurity("Authorization")

	g.PUT("/documents", handler.UpdateDocuments).
		SetTags("法人会員").
		SetSummary("法人会員ファイル登録").
		AddRequest(udomain.CorporateUserUpdateDocumentsRequest{}).
		AddResponse(nil, func(cu *openapigo.ContentUnit) {
			cu.HTTPStatus = http.StatusNoContent
			cu.Description = "ファイル登録成功応答値"
		}).
		AddErrors(
			factory.ErrorFailAuthorized,
			factory.ErrorParameter,
			factory.ErrorDB,
		).
		AddSecurity("Authorization")

	return handler
}

func (u corporateUserHttpDelivery) GetSignUpStatus(c echo.Context) error {
	ctx := c.Request().Context()

	// ユーザー情報抽出
	user, err := u.extractor.Extract(c)
	if err != nil {
		return factory.ErrorFailAuthorized.New(err)
	}

	// Request DTO作成
	request := udomain.CorporateUserGetSignUpStatusRequest{User: user}

	// Request DTOバインド
	if err = c.Bind(&request); err != nil {
		return factory.ErrorParameter.New(err)
	}

	// Request DTO妥当性検証
	if err = c.Validate(&request); err != nil {
		return factory.ErrorParameter.New(err)
	}

	response, err := u.corporateUserUsecase.GetSignUpStatus(ctx, request)

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

func (u corporateUserHttpDelivery) RegisterProfiles(c echo.Context) error {
	// ユーザー情報抽出
	user, err := u.extractor.Extract(c)
	if err != nil {
		return factory.ErrorFailAuthorized.New(err)
	}

	ctx := actorcontext.WithActor(c.Request().Context(), user)

	// Request DTO作成
	request := udomain.CorporateUserRegisterProfilesRequest{User: user}

	// Request DTOバインド
	if err = c.Bind(&request); err != nil {
		return factory.ErrorParameter.New(err)
	}

	// Request DTO妥当性検証
	if err = request.Validate(); err != nil {
		return factory.ErrorParameter.New(err)
	}

	err = u.corporateUserUsecase.RegisterProfiles(ctx, request)
	switch {
	case err == nil:
		return apirender.RenderSucc(c, http.StatusNoContent, "")
	case errors.Is(err, cdomain.ErrAlreadyExists):
		return factory.ErrorSameCorpExists.New(err)
	case errors.Is(err, cdomain.ErrItemConflict):
		return factory.ErrorParameter.New(err) // 無効な会員登録段階リクエスト
	case errors.Is(err, cdomain.ErrBadRequest):
		return factory.ErrorParameter.New(err) // 無効なSignUpStatusリクエスト
	default:
		return factory.ErrorDB.New(err)
	}
}

func (u corporateUserHttpDelivery) VerifyWithKyc(c echo.Context) error {
	// Request DTO作成
	user, err := u.extractor.Extract(c)
	if err != nil {
		return factory.ErrorFailAuthorized.New(err)
	}
	ctx := actorcontext.WithActor(c.Request().Context(), user)
	request := udomain.CorporateUserVerifyWithKycRequest{
		JWTUser: user,
	}

	// Request DTOバインド及び妥当性検証
	if err := c.Bind(&request); err != nil {
		return factory.ErrorParameter.New(err)
	}
	if err = request.Validate(); err != nil {
		return factory.ErrorParameter.New(err)
	}

	err = u.corporateUserUsecase.VerifyWithKyc(ctx, request)
	switch {
	case err == nil:
		return apirender.RenderSucc(c, http.StatusNoContent, nil)
	case errors.Is(err, cdomain.ErrItemConflict):
		return factory.ErrorParameter.New(err) // 無効な会員登録段階リクエスト
	case errors.Is(err, cdomain.ErrBadRequest):
		return factory.ErrorParameter.New(err) // 登録対象の書類が無い
	default:
		return factory.ErrorDB.New(err)
	}
}

func (u corporateUserHttpDelivery) VerifyPostalCode(c echo.Context) error {
	// ユーザー情報抽出
	user, err := u.extractor.Extract(c)
	if err != nil {
		return factory.ErrorFailAuthorized.New(err)
	}
	ctx := actorcontext.WithActor(c.Request().Context(), user)

	// Request DTO作成
	request := udomain.CorporateUserVerifyPostalCodeRequest{User: user}

	// Request DTOバインド
	if err = c.Bind(&request); err != nil {
		return factory.ErrorParameter.New(err)
	}

	// Request DTO妥当性検証
	if err = c.Validate(&request); err != nil {
		return factory.ErrorParameter.New(err)
	}

	err = u.corporateUserUsecase.VerifyPostalCode(ctx, request)
	switch {
	case err == nil:
		return apirender.RenderSucc(c, http.StatusNoContent, "")
	case errors.Is(err, cdomain.ErrItemConflict):
		return factory.ErrorInvalidAuditStatus.New(err)
	case errors.Is(err, cdomain.ErrBadRequest):
		return factory.ErrorCodeMismatch.New(err)
	case errors.Is(err, cdomain.ErrItemNotFound):
		return factory.ErrorNotFound.New(err)
	default:
		return factory.ErrorDB.New(err)
	}
}

func (u corporateUserHttpDelivery) GetOne(c echo.Context) error {
	ctx := c.Request().Context()

	// ユーザー情報抽出
	user, err := u.extractor.Extract(c)
	if err != nil {
		return factory.ErrorFailAuthorized.New(err)
	}

	// Request DTO作成
	request := udomain.CorporateUserGetOneRequest{User: user}

	response, err := u.corporateUserUsecase.GetOne(ctx, request)
	switch {
	case err == nil:
		return apirender.RenderSucc(c, http.StatusOK, response)
	case errors.Is(err, cdomain.ErrItemNotFound):
		return factory.ErrorNotFound.New(err)
	default:
		return factory.ErrorDB.New(err)
	}
}

func (u corporateUserHttpDelivery) PatchOne(c echo.Context) error {
	ctx := c.Request().Context()

	// ユーザー情報抽出
	user, err := u.extractor.Extract(c)
	if err != nil {
		return factory.ErrorFailAuthorized.New(err)
	}

	// Request DTO作成
	request := udomain.CorporateUserPatchOneRequest{
		JwtUser: user,
	}

	// Request DTOバインド
	if err = c.Bind(&request); err != nil {
		return factory.ErrorParameter.New(err)
	}

	// Request DTO妥当性検証
	if err = c.Validate(&request); err != nil {
		return factory.ErrorParameter.New(err)
	}

	err = u.corporateUserUsecase.PatchOne(ctx, request)
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

func (u corporateUserHttpDelivery) GetDocuments(c echo.Context) error {
	ctx := c.Request().Context()

	// ユーザー情報抽出
	user, err := u.extractor.Extract(c)
	if err != nil {
		return factory.ErrorFailAuthorized.New(err)
	}

	// Request DTO作成
	request := udomain.CorporateUserGetDocumentsRequest{JwtUser: user}

	response, err := u.corporateUserUsecase.GetDocuments(ctx, request)
	switch {
	case err == nil:
		return apirender.RenderSucc(c, http.StatusOK, response)
	default:
		return factory.ErrorDB.New(err)
	}
}

func (u corporateUserHttpDelivery) UpdateDocuments(c echo.Context) error {
	ctx := c.Request().Context()

	// ユーザー情報抽出
	user, err := u.extractor.Extract(c)
	if err != nil {
		return factory.ErrorFailAuthorized.New(err)
	}

	// Request DTO作成
	request := udomain.CorporateUserUpdateDocumentsRequest{JwtUser: user}

	// Request DTOバインド
	if err = c.Bind(&request); err != nil {
		return factory.ErrorParameter.New(err)
	}

	// Request DTO妥当性検証
	if err = c.Validate(&request); err != nil {
		return factory.ErrorParameter.New(err)
	}

	err = u.corporateUserUsecase.UpdateDocuments(ctx, request)
	switch {
	case err == nil:
		return apirender.RenderSucc(c, http.StatusOK, "")
	case errors.Is(err, cdomain.ErrBadRequest):
		return factory.ErrorParameter.New(err)
	case errors.Is(err, cdomain.ErrItemNotFound):
		return factory.ErrorNotFound.New(err)
	case errors.Is(err, cdomain.ErrForbidden):
		return factory.ErrorNotMatchOwner.New(err)
	default:
		return factory.ErrorDB.New(err)
	}
}
