package domain

import (
	"context"
	"net/http"

	"github.com/illuminarean/energy-shares-backend-repo/pkg/jwthelper"
	"github.com/labstack/echo/v4"
	"github.com/pkg/errors"

	"github.com/illuminarean/energy-shares-backend-repo/domain"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/jsonutils"
)

type UserUsecase interface {
	Login(ctx context.Context, request UserLoginRequest) (UserLoginResponse, error)
	RequestSignUp(ctx context.Context, request UserRequestSignUpRequest) (UserRequestSignUpResponse, error)
	CheckSignupRequestValidity(ctx context.Context, request UserCheckSignupRequestValidityRequest) (UserCheckSignupRequestValidityResponse, error)
	VerifyEmail(ctx context.Context, request UserVerifyEmailRequest) error
	SignUp(ctx context.Context, request UserSignUpRequest) error
	CreatePreSignedUrl(ctx context.Context, request UserCreatePreSignedUrlRequest) (UserCreatePreSignedUrlResponse, error)
}

type UserVerifyEmailChangedRequest struct {
	Email   string `json:"email" validate:"email" format:"email" description:"変更するためのメールアドレス" example:"gildong@email.com"`
	JWTUser jwthelper.User
}

type UserReverifyEmailChangedRequest struct {
	Email   string `json:"email" validate:"email" format:"email" description:"変更するためのメールアドレス" example:"gildong@email.com"`
	JWTUser jwthelper.User
}

type UserLoginRequest struct {
	Email     string `json:"email" description:"メールアドレス" required:"true" validate:"required" example:"gildong@email.com"`
	Password  string `json:"password" description:"パスワード" required:"true" validate:"required" example:"123456"`
	ClientIp  string `json:"-" swaggerignore:"true"`
	UserAgent string `json:"-" swaggerignore:"true"`
}

type UserLoginResponse struct {
	AccessToken                string         `json:"access_token" description:"アクセストークン" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9"`
	RefreshToken               string         `json:"-" swaggerignore:"true"`
	Cookies                    []*http.Cookie `json:"-" swaggerignore:"true"`
	IsFirstTimeLoginAfterAudit bool           `json:"is_first_time_login_after_audit" description:"本登録後初回ログインかどうか" example:"false"`
}

type UserRequestSignUpRequest struct {
	Email             string           `json:"email" required:"true" validate:"required" example:"gildong@email.com" description:"メールアドレス"`
	UserType          domain.USER_TYPE `json:"user_type" required:"true" validate:"required" enum:"PRIVATE,CORPORATE" description:"ユーザータイプ"`
	MarketingCode     string           `json:"marketing_code" description:"マーケティングコード" example:"1234"`
	MarketingCampaign string           `json:"marketing_campaign" description:"マーケティングキャンペーン" example:"test-campaign"`
	ReferralCode      string           `json:"referral_code" description:"紹介コード" example:"1234ABCD"`
}

func (u UserRequestSignUpRequest) Valid(c echo.Context) error {
	if err := c.Validate(&u); err != nil {
		return errors.Wrap(err, "UserRequestSignUpRequest validation error")
	}

	if (u.MarketingCode != "" && u.MarketingCampaign == "") || (u.MarketingCode == "" && u.MarketingCampaign != "") {
		return errors.New("marketing_code and marketing_campaign should be both empty or both not empty")
	}

	return nil
}

func (u UserRequestSignUpRequest) HasMarketing() bool {
	return u.MarketingCode != "" && u.MarketingCampaign != ""
}

type UserRequestSignUpResponse struct {
	RequestId string `json:"request_id" description:"会員登録リクエストID" example:"cb7f1379-8654-4eb5-a39b-1f667c41bb16"`
}

type UserCheckSignupRequestValidityRequest struct {
	RequestId string `param:"request_id" path:"request_id" required:"true" validate:"required"`
}

type UserCheckSignupRequestValidityResponse struct {
	Email    string           `json:"email" description:"メールアドレス" example:"gildong@email.com"`
	UserType domain.USER_TYPE `json:"user_type" description:"ユーザータイプ" enums:"PRIVATE,CORPORATE"`
}

type UserVerifyEmailRequest struct {
	RequestId string `param:"request_id" path:"request_id" required:"true" validate:"required" example:"cb7f1379-8654-4eb5-a39b-1f667c41bb16"`
}

// UserSignUpRequest
// https://illuminarean.atlassian.net/wiki/spaces/ESJ/pages/732758264/1#[USE-B-01]-%ED%9A%8C%EC%9B%90%EA%B0%80%EC%9E%85-%EA%B8%B0%EB%B3%B8%EC%A0%95%EB%B3%B4%EC%9E%85%EB%A0%A5
// 基本情報入力と約款同意
type UserSignUpRequest struct {
	RequestId             string `json:"request_id" description:"登録リクエストID" required:"true" validate:"required" example:"cb7f1379-8654-4eb5-a39b-1f667c41bb16"`
	UserType              string `json:"user_type" description:"ユーザータイプ" required:"true" validate:"required" enum:"PRIVATE,CORPORATE"`
	Email                 string `json:"email" description:"メールアドレス" required:"true" validate:"required" example:"gildong@email.com"`
	Password              string `json:"password" description:"パスワード" required:"true" validate:"required"`
	IsNewsLetter          bool   `json:"is_news_letter" description:"ニュースレター受信同意" example:"true"`
	CorpName              string `json:"corp_name" description:"企業名" example:"illiminrean"`
	ManagerFirstName      string `json:"manager_first_name" description:"担当者名" example:"둘리"`
	ManagerFamilyName     string `json:"manager_family_name" description:"担当者姓" example:"공"`
	ManagerFirstNameKana  string `json:"manager_first_name_kana" description:"担当者名振り仮名" example:"ギルドン" `
	ManagerFamilyNameKana string `json:"manager_family_name_kana" description:"担当者姓振り仮名" example:"ホン"`
	FirstName             string `json:"first_name" description:"名前" example:"길동"`
	FamilyName            string `json:"family_name" description:"姓名" example:"홍"`
	FirstNameKana         string `json:"first_name_kana" description:"名前振り仮名" example:"ギルドン"`
	FamilyNameKana        string `json:"family_name_kana" description:"姓名振り仮名" example:"ホン"`
}

func (u UserSignUpRequest) IsUserTypePrivate() bool {
	return u.UserType == string(domain.PRIVATE)
}

func (u UserSignUpRequest) Valid(c echo.Context) error {
	if err := c.Validate(&u); err != nil {
		return errors.Wrap(err, "UserSignUpRequest validation error")
	}

	if u.UserType == string(domain.CORPORATE) && u.CorpName == "" {
		return errors.New("corp_name is required")
	}

	// 法人会員の場合、マネージャー名、姓が完全に必須
	if u.UserType == string(domain.CORPORATE) && (u.ManagerFirstName == "" || u.ManagerFamilyName == "") {
		return errors.New("manager_first_name, manager_family_name are required")
	}

	// 法人会員の場合、マネージャー名ふりがな、姓ふりがなが完全に必須
	if u.UserType == string(domain.CORPORATE) && (u.ManagerFirstNameKana == "" || u.ManagerFamilyNameKana == "") {
		return errors.New("manager_first_name_kana, manager_family_name_kana are required")
	}

	// FirstName, FamilyName, FirstNameKana, FamilyNameKana は全て必須
	if u.UserType == string(domain.PRIVATE) && (u.FirstName == "" || u.FamilyName == "" || u.FirstNameKana == "" || u.FamilyNameKana == "") {
		return errors.New("first_name, family_name, first_name_kana, family_name_kana are required")
	}

	return nil

}

type UserCreatePreSignedUrlRequest struct {
	FileName string `json:"file_name" required:"true" validate:"required" description:"ファイル名" example:"test.jpg"`
}

type UserCreatePreSignedUrlResponse struct {
	PreSignedUrl       string             `json:"pre_signed_url" description:"署名済みURL" example:"https://es-jp-devv2-public-file-storage.s3.ap-northeast-2.amazonaws.com/attachment_file/2024-09-09/crfbl6cldh5hujq0ulq0.jpg?X-Amz-Algorithm=AWS4-HMAC-SHA256&X-Amz-Credential=ASIAX3VRTD32MJW2LNN4%2F20240909%2Fap-northeast-2%2Fs3%2Faws4_request&X-Amz-Date=20240909T090633Z&X-Amz-Expires=1800&X-Amz-Security-Token=IQoJb3JpZ2luX2VjEDEaDmFwLW5vcnRoZWFzdC0yIkYwRAIgJ6IYFoha%2F5rncjLWPtfaMIRuX06TrV%2BWFx3zpXSYqOECIGLiFJ9yG9e%2BKTkpnYg3bxQbCENRKDILaxC%2F5TPVTpv1Ku4BCFoQBRoMNTQwNDY0NTg2NDg0Igy15W%2F6hkjthh13v%2F4qywHWmD2sop8LWdE9Ln3myQxJVhOyQNJUvDAGkw%2BUKfcIN%2FBXDcqspAXx6VoUmR7GwHQh60WteRBtirqhRh8ZxSOgCoHnrw48hhCFg%2BiaOkipop6EHRe6d6%2B6UGmie4QdZvjf4R4WWkiI31Qqjo84FQyVrNRPHK5Ig%2FVrnQqRqvSbRblVZCGYB83S6F4qNcHxBDIjOOTweAXxEZpPoO5PLS8dAJq4E5nuMlfA3QdRr2JadvARmel0YUp%2B0LXTo33pkDwZNAy3OABoBkOaEDC79Pq2BjqZATp5J2nnJUkj574d1r7Bt6QpgVOlDD39CNm%2BLRwTbfD95T1lYXD6%2F%2BW6%2F%2F05QaRRWuTqHhrF3st3qy1zssEA%2FdBry%2FIbJqAO1KB9jke9i274zEt1vh%2FFuxn0MevBmYdATloB31cFfK0JcewBjceZPs6ntVH06YPHkItb77e3R3AqcIV1AxSWElTMWCiUUaPlWXYfzERGsXklZA%3D%3D&X-Amz-SignedHeaders=content-disposition%3Bcontent-type%3Bhost&X-Amz-Signature=92c51d623d8079b4c6748693ae3e1131e7e811e3d0f29de4eb3c7855d1ceab99"`
	ExpiredAt          jsonutils.JSONTime `json:"expired_at" description:"有効期限" example:"2024-09-09 18:36:33"`
	ContentType        string             `json:"content_type" description:"コンテンツタイプ" example:"image/jpeg"`
	ContentDisposition string             `json:"content_disposition" description:"コンテンツ表示" example:"attachment; filename=\"test.jpg\""`
}
