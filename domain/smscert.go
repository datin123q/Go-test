package domain

import (
	"context"
	"errors"
	"time"

	"github.com/illuminarean/energy-shares-backend-repo/pkg/factory"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/jsonutils"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/serviceutils"
)

type VERIFICATION_REQUEST_STATUS string

var (
	ENABLE  VERIFICATION_REQUEST_STATUS = "enable"
	DISABLE VERIFICATION_REQUEST_STATUS = "disable"
)

type MOBILE_PHONE_TYPE string

var (
	DEFAULT_PHONE MOBILE_PHONE_TYPE = "DEFAULT"
	CEO_PHONE     MOBILE_PHONE_TYPE = "CEO"
	MANAGER_PHONE MOBILE_PHONE_TYPE = "MANAGER"
)

type MobilePhoneVerificationRequest struct { // 문자 정보
	Id          int64                       `json:"id"`
	RequestId   string                      `json:"request_id" xorm:"not null UUID"`
	UserId      int64                       `json:"user_id"`
	Code        string                      `json:"code"`
	MobilePhone string                      `json:"mobile_phone"`
	Status      VERIFICATION_REQUEST_STATUS `json:"status"`
	FirstReq    bool                        `json:"first_req" swagger:"desc(최초 발송인지 여부)"`
	PhoneType   MOBILE_PHONE_TYPE           `json:"phone_type" swagger:"desc(전화번호 종류(기본, 대표, 거래 담당자))"`
	Expire      jsonutils.JSONTime          `json:"expire" swagger:"desc(만료 일자)"`
	Created     jsonutils.JSONTime          `json:"created" xorm:"created"`
}

func (p *MobilePhoneVerificationRequest) CreateMobilePhoneVerificationRequest(ctx context.Context) (int64, error) {
	return factory.DB(ctx).Insert(p)
}

func (p *MobilePhoneVerificationRequest) UpdateMobilePhoneVerificationRequest(ctx context.Context) (int64, error) {
	return factory.DB(ctx).ID(p.Id).Update(p)
}

func (MobilePhoneVerificationRequest) GetByRequestId(ctx context.Context, requestId string) (*MobilePhoneVerificationRequest, error) {
	var v MobilePhoneVerificationRequest
	if has, err := factory.DB(ctx).Where("request_id = ?", requestId).Get(&v); err != nil {
		return nil, err
	} else if !has {
		return nil, nil
	}
	return &v, nil
}

// GetNotExpireByUserIdAndMobilePhone 유효한 문자 정보 조회
func (p *MobilePhoneVerificationRequest) GetNotExpireByUserIdAndMobilePhone(ctx context.Context, userId int64, phoneType MOBILE_PHONE_TYPE, mobilePhone string) (*MobilePhoneVerificationRequest, error) {
	var v MobilePhoneVerificationRequest
	has, err := factory.DB(ctx).Where("user_id = ? and phone_type = ? and mobile_phone = ? and expire > ?",
		userId, phoneType, mobilePhone, time.Now().Format(time.DateTime)).Get(&v)
	if err != nil {
		return nil, err
	} else if !has {
		return nil, nil
	}

	return &v, nil
}

// CountFirstReqByUserIdInToday 오늘 최초 발송된 문자 개수
func (p *MobilePhoneVerificationRequest) CountFirstReqByUserIdInToday(ctx context.Context, userId int64, phoneType MOBILE_PHONE_TYPE) (int64, error) {
	var today = time.Now().Format(time.DateOnly)

	return factory.DB(ctx).Where("user_id = ? and phone_type = ? and first_req = true and created > ?",
		userId, phoneType, today).Count(p)
}

// CountResendReqAfterFirstReqByUserId 최초 발송 이후 재발송된 문자 개수
func (p *MobilePhoneVerificationRequest) CountResendReqAfterFirstReqByUserId(ctx context.Context, userId int64, phoneType MOBILE_PHONE_TYPE) (int64, error) {
	return factory.DB(ctx).Where("user_id = ? and phone_type = ? and first_req = false and id > (select max(id) from mobile_phone_verification_request where user_id = ? and first_req = true)",
		userId, phoneType, userId).Count(p)
}

// FixedSmsCode は、staging などでSMSを受け取らずに認証を試すための固定のコード。fixedSmsCodeEnabled のときだけ効く
// (テストから参照するため公開している)。
const FixedSmsCode = "830325"

// fixedSmsCodeEnabled は、固定のコードを使える環境かを返す。MIR-1420: 本番では使えない。環境の名前が想定と違っても
// 本番で効かないよう、開発環境・stagingと明示された名前のときだけ許す(config.IsDevelopment と同じ範囲)。
func fixedSmsCodeEnabled() bool {
	return serviceutils.EnvIsLocal() || serviceutils.EnvIsDev() || serviceutils.EnvIsQa() || serviceutils.EnvIsStg()
}

func (p *MobilePhoneVerificationRequest) VerifyRequest(ctx context.Context, userId int64, requestId, code string, verifyTime time.Time) (bool, error) {
	if code == FixedSmsCode && fixedSmsCodeEnabled() {
		return true, nil
	}
	if p.UserId != userId {
		return false, errors.New("Not Match User")
	}

	if p.Code != code {
		return false, errors.New("Not Match Code")
	}

	if p.RequestId != requestId {
		return false, errors.New("Not Match request Id")
	}

	// if time.Time(p.Expire).After(verifyTime) {
	//	return false, errors.New("Expire Code")
	// }
	if _, err := p.UpdateMobilePhoneVerificationRequest(ctx); err != nil {
		return true, nil
	}

	return true, nil

}
