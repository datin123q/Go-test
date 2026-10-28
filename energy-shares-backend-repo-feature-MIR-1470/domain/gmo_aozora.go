package domain

import (
	"context"

	"github.com/pkg/errors"
	"xorm.io/xorm"

	"github.com/illuminarean/energy-shares-backend-repo/pkg/cipher"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/commoncode"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/config"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/factory"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/i18nutils"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/jsonutils"
)

// //gmo-aozora
//
//	type ApiRequestLog struct {
//		Id          int64              `json:"id"
//		UserId      int64              `json:"user_id"`
//		AccountId   int64              `json:"account_id"`
//		Url         string             `json:"url"`
//		Request     string             `json:"request" xorm:"TEXT"`
//		Response    string             `json:"response" xorm:"TEXT"`
//		RequestTime jsonutils.JSONTime `json:"request_time"`
//		Created     jsonutils.JSONTime `json:"created" xorm:"created"`
//		Updated     jsonutils.JSONTime `json:"updated" xorm:"updated"`
//	}
const GMO_AOZORA_BANK_NAME = "ＧМＯあおぞらネット銀行"
const (
	// 사용 금지! 삭제 예정! (pkg/cipher 의 값 사용)
	CIPHER_KEY = "wM3b*MLjg9LHLrfXWbC-iR3hbjxz!3n8" //32자만 가능
	ES_JP_
	CIPHER_IVKEY = "AHPQ@a!N7xoBE_Ur" //16자만 가능
)

type DEPOSIT_STATUS string

const (
	DEPOSIT_STATUS_CODE                = "DEPOSIT_STATUS"
	FULL_USE_DEPOSIT    DEPOSIT_STATUS = "FULL_USE_DEPOSIT"    // 전체 사용
	NOT_USE_DEPOSIT     DEPOSIT_STATUS = "NOT_USE_DEPOSIT"     // 미사용
	PARTIAL_USE_DEPOSIT DEPOSIT_STATUS = "PARTIAL_USE_DEPOSIT" // 일부 사용
	REFUND_USE_DEPOSIT  DEPOSIT_STATUS = "REFUND_USE_DEPOSIT"
)

func (VaDeposit) GetDepositCheckStatus(loc string) commoncode.CodeGroup {
	codes := make([]commoncode.Code, 0, 0)
	codes = append(codes, commoncode.Code{Code: string(FULL_USE_DEPOSIT), Name: i18nutils.Translate("commoncode."+DEPOSIT_STATUS_CODE+"."+string(FULL_USE_DEPOSIT), loc), Order: len(codes) + 1})
	codes = append(codes, commoncode.Code{Code: string(NOT_USE_DEPOSIT), Name: i18nutils.Translate("commoncode."+DEPOSIT_STATUS_CODE+"."+string(NOT_USE_DEPOSIT), loc), Order: len(codes) + 1})
	codes = append(codes, commoncode.Code{Code: string(PARTIAL_USE_DEPOSIT), Name: i18nutils.Translate("commoncode."+DEPOSIT_STATUS_CODE+"."+string(PARTIAL_USE_DEPOSIT), loc), Order: len(codes) + 1})

	return commoncode.CodeGroup{
		GroupCode: DEPOSIT_STATUS_CODE,
		Codes:     codes,
	}
}

//
//// 계좌 정도
//type BankAccount struct {
//	Id                     int64              `json:"id"`
//	AccountId              string             `json:"account_id" xorm:"UNIQUE NOT NULL"` // 아조라 뱅크 어카운트 ID
//	AccountNumber          string             `json:"account_number" xorm:"UNIQUE NOT NULL"`
//	BranchCode             string             `json:"branch_code"`
//	BranchName             string             `json:"branch_name"`
//	AccountTypeCode        string             `json:"account_type_code"`
//	AccountTypeName        string             `json:"account_type_name"`
//	PrimaryAccountCode     string             `json:"primary_account_code"`
//	PrimaryAccountCodeName string             `json:"primary_account_code_name"`
//	AccountName            string             `json:"account_name"`
//	AccountNameKana        string             `json:"account_name_kana"`
//	CurrencyCode           string             `json:"currency_code"`
//	CurrencyName           string             `json:"currency_name"`
//	TransferLimitAmount    string             `json:"transfer_limit_amount"`
//	AdminNote              string             `json:"admin_note"`
//	BaseDate               string             `json:"base_date"` //최종 시간
//	BaseTime               string             `json:"base_time"` //조회 시간
//	Created                jsonutils.JSONTime `json:"created" xorm:"created"`
//	Updated                jsonutils.JSONTime `json:"updated" xorm:"updated"`
//}
//
//// 잔고
//type BankBalance struct {
//	Id                   int64              `json:"id"`
//	AccountId            string             `json:"account_id" xorm:"UNIQUE NOT NULL"`
//	AccountTypeName      string             `json:"account_type_name"`
//	AccountTypeCode      string             `json:"account_type_code"`
//	Balance              string             `json:"balance"`
//	WithDrawableAmount   string             `json:"with_drawable_amount"`
//	PreviousDayBalance   string             `json:"previous_day_balance"`
//	PreviousMonthBalance string             `json:"previous_month_balance"`
//	CurrencyCode         string             `json:"currency_code"`
//	CurrencyName         string             `json:"currency_name"`
//	BaseDate             string             `json:"base_date"` //최종 시간
//	BaseTime             string             `json:"base_time"` //조회 시간
//	Created              jsonutils.JSONTime `json:"created" xorm:"created"`
//	Updated              jsonutils.JSONTime `json:"updated" xorm:"updated"`
//}

type GmoAozoraClient struct {
	Id           int64              `json:"id" xorm:"pk autoincr"`
	ClientId     string             `json:"client_id" xorm:"client_id"`
	ClientSecret string             `json:"client_secret" xorm:"client_secret"`
	Scope        string             `json:"scope" xorm:"scope"`
	RedirectUri  string             `json:"redirect_uri" xorm:"redirect_uri"`
	AccessToken  string             `json:"access_token" xorm:"access_token"`
	RefreshToken string             `json:"refresh_token" xorm:"refresh_token"`
	Created      jsonutils.JSONTime `json:"created" xorm:"created"`
	Updated      jsonutils.JSONTime `json:"updated" xorm:"updated"`
}

func (GmoAozoraClient) InitClient(db xorm.Interface, cfg config.Config) (*GmoAozoraClient, error) {
	crypto, err := cipher.NewNiceCrypto(CIPHER_KEY, CIPHER_IVKEY)
	if err != nil {
		return nil, errors.Wrap(err, "failed create cipher crypto")
	}

	var client = GmoAozoraClient{}
	has, err := db.Table(&GmoAozoraClient{}).Select("gmo_aozora_client.*").ID(1).Get(&client)

	if err != nil {
		return nil, err
	}
	if !has {
		accessToken, err := crypto.Encrypt("VADN0H49S9Cnvmy34yNLBEvRoSCGkyTQ")
		if err != nil {
			return nil, err
		}

		refreshToken, err := crypto.Encrypt("J9O2z7QjhHYPqE7AvQxBL197WRGsZkgH")
		if err != nil {
			return nil, err
		}
		client = GmoAozoraClient{
			Id:           1,
			ClientId:     cfg.GmoAozoraConfig.ClientId,
			ClientSecret: cfg.GmoAozoraConfig.ClientSecret,
			Scope:        "openid offline_access private:account private:transfer private:bulk-transfer private:virtual-account",
			RedirectUri:  cfg.GmoAozoraConfig.RedirectUri,
			AccessToken:  accessToken,
			RefreshToken: refreshToken,
		}
		_, err2 := db.Insert(&client)
		if err2 != nil {
			return nil, err2
		}
	} else {
		client.ClientId = cfg.GmoAozoraConfig.ClientId
		client.ClientSecret = cfg.GmoAozoraConfig.ClientSecret
		client.RedirectUri = cfg.GmoAozoraConfig.RedirectUri
		client.AccessToken, err = crypto.Decrypt(client.AccessToken)
		if err != nil {
			return nil, errors.Wrap(err, "failed decrypt Access Token")
		}
		client.RefreshToken, err = crypto.Decrypt(client.RefreshToken)
		if err != nil {
			return nil, errors.Wrap(err, "failed decrypt Refresh Token")
		}
	}

	return &client, nil
}

func (c *GmoAozoraClient) Create(ctx context.Context) (int64, error) {
	crypto, err := cipher.NewNiceCrypto(CIPHER_KEY, CIPHER_IVKEY)
	if err != nil {
		return 0, errors.Wrap(err, "failed create cipher crypto")
	}

	c.AccessToken, err = crypto.Encrypt(c.AccessToken)
	if err != nil {
		return 0, errors.Wrap(err, "failed encrypt access token")
	}
	c.RefreshToken, err = crypto.Encrypt(c.RefreshToken)
	if err != nil {
		return 0, errors.Wrap(err, "failed encrypt refresh token")
	}
	return factory.DB(ctx).Insert(c)
}

func (GmoAozoraClient) GetByID(ctx context.Context, id int64) (*GmoAozoraClient, error) {
	var d GmoAozoraClient
	has, err := factory.DB(ctx).ID(id).Get(&d)

	if err != nil {
		return nil, err
	} else if !has {
		return nil, nil
	}

	crypto, err := cipher.NewNiceCrypto(CIPHER_KEY, CIPHER_IVKEY)
	if err != nil {
		return nil, errors.Wrap(err, "failed create cipher crypto")
	}

	d.AccessToken, err = crypto.Decrypt(d.AccessToken)
	if err != nil {
		return nil, errors.Wrap(err, "failed encrypt access token")
	}
	d.RefreshToken, err = crypto.Decrypt(d.RefreshToken)
	if err != nil {
		return nil, errors.Wrap(err, "failed encrypt refresh token")
	}

	return &d, nil
}

// Update は暗号化してから更新し、実際に更新できた件数を返す(MIR-1298)。
// 呼び出し元がaffected==0を「失敗」として扱えるようにするため
// (以前は件数を捨てていた)。
func (d *GmoAozoraClient) Update(ctx context.Context) (int64, error) {

	crypto, err := cipher.NewNiceCrypto(CIPHER_KEY, CIPHER_IVKEY)
	if err != nil {
		return 0, errors.Wrap(err, "failed create cipher crypto")
	}

	d.AccessToken, err = crypto.Encrypt(d.AccessToken)
	if err != nil {
		return 0, errors.Wrap(err, "failed encrypt access token")
	}
	d.RefreshToken, err = crypto.Encrypt(d.RefreshToken)
	if err != nil {
		return 0, errors.Wrap(err, "failed encrypt refresh token")
	}
	return factory.DB(ctx).ID(d.Id).Update(d)
}
