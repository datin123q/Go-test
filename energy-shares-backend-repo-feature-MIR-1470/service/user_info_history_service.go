package service

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"strings"

	"github.com/sirupsen/logrus"

	"github.com/illuminarean/energy-shares-backend-repo/domain"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/factory"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/jsonutils"
)

type IUserInfoHistoryService interface {
	GetDataForUser(context.Context, int64) (*UserInfoHistoryData, error)
	GetDataForPrivateUser(context.Context, int64) (*UserInfoHistoryData, error)
	GetDataForCorporateUser(context.Context, int64) (*UserInfoHistoryData, error)
	CreateUserInfoHistory(context.Context, UserInfoHistoryData, UserInfoHistoryData, int64, string, string) error
}

var _ IUserInfoHistoryService = (*UserInfoHistoryService)(nil)

type UserInfoHistoryService struct {
	// repository
}

func NewUserInfoHistoryService() IUserInfoHistoryService {
	return &UserInfoHistoryService{}
}

type UserInfoHistoryData struct {
	User                 interface{} // private: PrivateUser, corporate: CorporateUser
	AntiSocialInfo       AntiSocialInfo
	Document             Document
	Account              Account
	CorporateShareHolder []*domain.CorporateShareholder // corporate only
}

type AntiSocialInfo struct {
	AntiSocialCheck     *domain.UserAntisocialCheck   // private & corporate
	AntisocialCheckFile []*domain.AntisocialCheckFile // private & corporate
}

type Document struct {
	Identification1Front []*domain.UserFile // private & corporate
	Identification1Back  []*domain.UserFile // private & corporate
	Identification2Front []*domain.UserFile // private & corporate
	Identification2Back  []*domain.UserFile // private & corporate
	CertifiedCopyFiles   []*domain.UserFile // corporate only
	PoaFiles             []*domain.UserFile // corporate only
}

type Account struct {
	UserAccount     *domain.UserAccount        // private & corporate
	EquityVaAccount *domain.UserVirtualAccount // private & corporate
}

func (s *UserInfoHistoryService) GetDataForUser(ctx context.Context, userId int64) (*UserInfoHistoryData, error) {
	var userData = &UserInfoHistoryData{}
	var err error = nil

	user, err := domain.User{}.GetById(ctx, userId)
	if err != nil {
		return nil, factory.ErrorDB.New(err)
	}
	if user == nil {
		return nil, factory.ErrorNotFound.New(errors.New("user not found"))
	}

	userData.User = user

	return userData, nil
}

func (s *UserInfoHistoryService) GetDataForPrivateUser(ctx context.Context, userId int64) (*UserInfoHistoryData, error) {
	var userData = &UserInfoHistoryData{}
	var err error = nil

	privateUser, err := domain.PrivateUser{}.GetById(ctx, userId)
	if err != nil {
		return nil, factory.ErrorDB.New(err)
	} else if privateUser == nil {
		return nil, factory.ErrorNotFound.New(errors.New("user not found"))
	}

	privateUser.Roles = []string{string(domain.ROLE_USER)}
	userData.User = privateUser

	userData.AntiSocialInfo.AntiSocialCheck, err = domain.UserAntisocialCheck{}.GetByCheckUserId(ctx, userId)
	if err != nil {
		return nil, factory.ErrorDB.New(err)
	}
	userData.AntiSocialInfo.AntisocialCheckFile, err = domain.AntisocialCheckFile{}.FindByCheckId(ctx, 10, userId)
	if err != nil {
		return nil, factory.ErrorDB.New(err)
	}

	userData.Document.Identification1Front, err = domain.UserFile{}.FindByUserIdAndFileType(ctx, 10, userId, domain.IDENTIFICATION1_FRONT)
	if err != nil {
		return nil, factory.ErrorDB.New(err)
	}
	userData.Document.Identification1Back, err = domain.UserFile{}.FindByUserIdAndFileType(ctx, 10, userId, domain.IDENTIFICATION1_BACK)
	if err != nil {
		return nil, factory.ErrorDB.New(err)
	}
	userData.Document.Identification2Front, err = domain.UserFile{}.FindByUserIdAndFileType(ctx, 10, userId, domain.IDENTIFICATION2_FRONT)
	if err != nil {
		return nil, factory.ErrorDB.New(err)
	}
	userData.Document.Identification2Back, err = domain.UserFile{}.FindByUserIdAndFileType(ctx, 10, userId, domain.IDENTIFICATION2_BACK)
	if err != nil {
		return nil, factory.ErrorDB.New(err)
	}

	userData.Account.EquityVaAccount, err = domain.UserVirtualAccount{}.GetVaByUserId(ctx, userId, domain.EQUITY)
	if err != nil {
		return nil, factory.ErrorDB.New(err)
	}
	userData.Account.UserAccount, err = domain.UserAccount{}.GetByUserId(ctx, userId)
	if err != nil {
		return nil, factory.ErrorDB.New(err)
	}

	return userData, nil
}

func (s *UserInfoHistoryService) GetDataForCorporateUser(ctx context.Context, userId int64) (*UserInfoHistoryData, error) {
	var userData = &UserInfoHistoryData{}
	var err error = nil

	corporateUser, err := domain.CorporateUser{}.GetById(ctx, userId)
	if err != nil {
		return nil, factory.ErrorDB.New(err)
	} else if corporateUser == nil {
		return nil, factory.ErrorNotFound.New(errors.New("user not found"))
	}

	userData.User = corporateUser
	userData.AntiSocialInfo.AntiSocialCheck = corporateUser.AntisocialCheck
	userData.Document.Identification1Front = corporateUser.Identification1FrontFile
	userData.Document.Identification1Back = corporateUser.Identification1BackFile
	userData.Document.Identification2Front = corporateUser.Identification2FrontFile
	userData.Document.Identification2Back = corporateUser.Identification2BackFile
	userData.Document.CertifiedCopyFiles = corporateUser.CertifiedCopyFiles
	userData.Document.PoaFiles = corporateUser.PoaFiles
	userData.Account.UserAccount = corporateUser.UserAccount
	userData.Account.EquityVaAccount = corporateUser.EquityVaAccount
	userData.CorporateShareHolder = corporateUser.CorporateShareholder

	userData.AntiSocialInfo.AntisocialCheckFile, err = domain.AntisocialCheckFile{}.FindByCheckId(ctx, 10, userId)
	if err != nil {
		return nil, factory.ErrorDB.New(err)
	}

	return userData, nil
}

func (s *UserInfoHistoryService) CreateUserInfoHistory(
	ctx context.Context, before, after UserInfoHistoryData, processId int64, processName string, processEmail string) error {

	// UserInfoHistoryData.User 가 user, private_user, corporate_user 중에 해당되는 구조체인지 확인 후
	// 각각의 필드에 지정된 i18n tag 값을 저장하기 위한 자료구조
	var userInfoHistory *domain.UserInfoHistory
	var prevMapData map[string]interface{}
	var currMapData map[string]interface{}
	if v, ok := after.User.(*domain.User); ok {
		userInfoHistory = toUserInfoHistory(ctx, v.Id, processId, processName, processEmail)
		prevUser := before.User
		currUser := v

		prevMapUser := convertTagToMap(reflect.ValueOf(prevUser).Elem())
		currMapUser := convertTagToMap(reflect.ValueOf(currUser).Elem())
		prevMapData = mergeMaps(prevMapUser)
		currMapData = mergeMaps(currMapUser)

	} else if v, ok := after.User.(*domain.PrivateUser); ok {
		userInfoHistory = toUserInfoHistory(ctx, v.User.Id, processId, processName, processEmail)
		prevUser := &before.User.(*domain.PrivateUser).User
		currUser := &v.User

		prevMapUser := convertTagToMap(reflect.ValueOf(prevUser).Elem())
		prevMapProfile := convertTagToMap(reflect.ValueOf(&before.User.(*domain.PrivateUser).PrivateProfile).Elem())
		currMapUser := convertTagToMap(reflect.ValueOf(currUser).Elem())
		currMapProfile := convertTagToMap(reflect.ValueOf(&after.User.(*domain.PrivateUser).PrivateProfile).Elem())

		prevMapData = mergeMaps(prevMapUser, prevMapProfile)
		currMapData = mergeMaps(currMapUser, currMapProfile)

	} else if v, ok := after.User.(*domain.CorporateUser); ok {
		userInfoHistory = toUserInfoHistory(ctx, v.User.Id, processId, processName, processEmail)
		prevUser := &before.User.(*domain.CorporateUser).User
		currUser := &v.User

		prevMapUser := convertTagToMap(reflect.ValueOf(prevUser).Elem())
		prevMapProfile := convertTagToMap(reflect.ValueOf(&before.User.(*domain.CorporateUser).CorporateProfile).Elem())
		currMapUser := convertTagToMap(reflect.ValueOf(currUser).Elem())
		currMapProfile := convertTagToMap(reflect.ValueOf(&after.User.(*domain.CorporateUser).CorporateProfile).Elem())

		prevMapData = mergeMaps(prevMapUser, prevMapProfile)
		currMapData = mergeMaps(currMapUser, currMapProfile)
	}

	// 인증상태 비교 후 히스토리를 남김
	compareAuditState(ctx, prevMapData, currMapData, userInfoHistory)

	// 거래제한 상태 비교 후 히스토리를 남김
	compareBlockInvest(ctx, prevMapData, currMapData, userInfoHistory)

	// 내용 상태 비교 후 히스토리를 남김
	compareContents(ctx, prevMapData, currMapData, userInfoHistory)

	// 투자등급 변경 상태 비교 후 히스토리를 남김
	compareInvestment(ctx, prevMapData, currMapData, userInfoHistory)

	// 증빙서류 변경 상태 비교 후 히스토리를 남김
	compareDocument(ctx, before.Document, after.Document, userInfoHistory)

	// 반사회세력 정보 변경 상태 비교 후 히스토리를 남김
	compareAntisocial(ctx, before.AntiSocialInfo, after.AntiSocialInfo, userInfoHistory)

	// 최대주주 정보 변경 상태 비교 후 히스토리를 남김
	compareShareholder(ctx, before, after, userInfoHistory)

	return nil
}

// i18nタグで指定されたフィールド値を保存
func convertTagToMap(val reflect.Value) map[string]interface{} {

	var mapData = make(map[string]interface{}, 0)

	// すべてのフィールドのタグを検査
	for i := 0; i < val.NumField(); i++ {
		typeField := val.Type().Field(i)
		tag := typeField.Tag

		i18nStr := tag.Get("i18n")
		if len(i18nStr) <= 0 {
			continue
		}

		// .で区切って最後のstringをキーとして指定
		strs := strings.Split(i18nStr, ".")
		str := strs[len(strs)-1]
		if len(str) <= 0 {
			continue
		}

		// 各フィールドのデータ値をstringに変換した後、valueに保存
		switch t := val.Field(i).Interface().(type) {
		case string:
			mapData[str] = t
		case *string:
			if t != nil {
				mapData[str] = *t
			} else {
				mapData[str] = ""
			}
		case bool:
			// bool データ型は準備された翻訳コード値で保存
			if str == string(domain.USER_CHANGED_CONTENT_AUDIT_STATUS) || str == string(domain.USER_CHANGED_CONTENT_TRANSACTION) {
				mapData[str] = strconv.FormatBool(t)
			} else if str == string(domain.USER_CHANGED_CONTENT_PROFESSIONAL_INVESTOR) {
				if t == true {
					mapData[str] = string(domain.USER_HISTORY_AGREEMENT_AGREE)
				} else if t == false {
					mapData[str] = string(domain.USER_HISTORY_AGREEMENT_DISAGREE)
				}
			} else if str == string(domain.USER_CHANGED_CONTENT_POSTAL_CODE) || str == string(domain.USER_CHANGED_CONTENT_MANAGER_POSTAL_CODE_STATUS) {
				if t == true {
					mapData[str] = string(domain.POSTAL_CODE_STATE_SENT)
				} else if t == false {
					mapData[str] = string(domain.POSTAL_CODE_STATE_NOTSENT)
				}
			} else {
				if t == true {
					mapData[str] = string(domain.USER_HISTORY_CHECK_YES)
				} else if t == false {
					mapData[str] = string(domain.USER_HISTORY_CHECK_NO)
				}
			}
		case *bool:
			if t != nil {
				if str == string(domain.USER_CHANGED_CONTENT_PRO_INVEST_ASSET) ||
					str == string(domain.USER_CHANGED_CONTENT_PRO_INVEST_FINANCE_ASSET) ||
					str == string(domain.USER_CHANGED_CONTENT_OPERATINGFUNDS) {
					mapData[str] = strconv.FormatBool(*t)
				}
			}
		default:
			if value, ok := val.Field(i).Interface().(jsonutils.JSONTime); ok {
				mapData[str] = value.Time().String()
			} else {
				mapData[str] = fmt.Sprintf("%v", t)
			}
		}
	}

	return mapData
}

// 認証状態比較
func compareAuditState(ctx context.Context, prevData, currData map[string]interface{}, info *domain.UserInfoHistory) {

	var prev, curr string
	if val, ok := prevData[string(domain.USER_CHANGED_CONTENT_AUDIT_STATUS)]; ok {
		prev = val.(string)
	}
	if val, ok := currData[string(domain.USER_CHANGED_CONTENT_AUDIT_STATUS)]; ok {
		curr = val.(string)
	}

	if prev == curr {
		return
	}

	// 認証状態を比較し、値が異なる場合は情報カテゴリーと情報内容のテーブルに変更された値を保存
	info.Id = 0
	createHistory(ctx, info, string(domain.USER_HISTORY_AUDIT))
	domain.UserChangedContent{}.Create(ctx, string(domain.USER_CHANGED_CONTENT_AUDIT_STATUS), info.Id, prev, curr, "", "", "", "")
}

// 取引制限状態の比較
func compareBlockInvest(ctx context.Context, prevData, currData map[string]interface{}, info *domain.UserInfoHistory) {

	var prev, curr string
	if val, ok := prevData[string(domain.USER_CHANGED_CONTENT_TRANSACTION)]; ok {
		prev = val.(string)
	}
	if val, ok := currData[string(domain.USER_CHANGED_CONTENT_TRANSACTION)]; ok {
		curr = val.(string)
	}

	if prev == curr {
		return
	}

	info.Id = 0
	createHistory(ctx, info, string(domain.USER_HISTORY_TRANSACTION))
	if prev == "true" && curr == "false" {
		domain.UserChangedContent{}.Create(ctx, string(domain.USER_CHANGED_CONTENT_TRANSACTION), info.Id, string(domain.CONTRACT_BLOCK), string(domain.CONTRACT_NOTBLOCK), "", "", "", "")
	} else if prev == "false" && curr == "true" {
		domain.UserChangedContent{}.Create(ctx, string(domain.USER_CHANGED_CONTENT_TRANSACTION), info.Id, string(domain.CONTRACT_NOTBLOCK), string(domain.CONTRACT_BLOCK), "", "", "", "")
	}
}

// 内容比較
func compareContents(ctx context.Context, prevData, currData map[string]interface{}, info *domain.UserInfoHistory) {

	// 比較するキー値をスライスに保存
	contentTypes := []string{string(domain.USER_CHANGED_CONTENT_CORP_NAME), string(domain.USER_CHANGED_CONTENT_EMAIL), string(domain.USER_CHANGED_CONTENT_NAME), string(domain.USER_CHANGED_CONTENT_FAMILY_NAME),
		string(domain.USER_CHANGED_CONTENT_NAME2), string(domain.USER_CHANGED_CONTENT_FAMILY_NAME2),
		string(domain.USER_CHANGED_CONTENT_MOBILEPHONE), string(domain.USER_CHANGED_CONTENT_ZIPCODE),
		string(domain.USER_CHANGED_CONTENT_ADDRESS), string(domain.USER_CHANGED_CONTENT_ADDRESS2), string(domain.USER_CHANGED_CONTENT_PROVINCE),
		string(domain.USER_CHANGED_CONTENT_CITY), string(domain.USER_CHANGED_CONTENT_POSTAL_CODE), string(domain.USER_CHANGED_CONTENT_MANAGER_POSTAL_CODE_STATUS), string(domain.USER_CHANGED_CONTENT_FINANCE_NAME),
		string(domain.USER_CHANGED_CONTENT_BANK_NAME), string(domain.USER_CHANGED_CONTENT_DEPOSIT_TYPE), string(domain.USER_CHANGED_CONTENT_BANK_ACCOUNT),
		string(domain.USER_CHANGED_CONTENT_DEPOSITOR), string(domain.USER_CHANGED_CONTENT_NEWSLETTER), string(domain.USER_CHANGED_CONTENT_CEO_FIRSTNAME),
		string(domain.USER_CHANGED_CONTENT_CEO_FIRSTNAME_FURIGANA), string(domain.USER_CHANGED_CONTENT_CEO_LASTNAME), string(domain.USER_CHANGED_CONTENT_CEO_LASTNAME_FURIGANA),
		string(domain.USER_CHANGED_CONTENT_CEO_EMAIL), string(domain.USER_CHANGED_CONTENT_CEO_PHONE), string(domain.USER_CHANGED_CONTENT_CEO_CERTIFICATION),
		string(domain.USER_CHANGED_CONTENT_MANAGER_CERTIFICATION), string(domain.USER_CHANGED_CONTENT_CERTIFICATION),
		string(domain.USER_CHANGED_CONTENT_INCHARGE_FIRSTNAME), string(domain.USER_CHANGED_CONTENT_INCHARGE_FIRSTNAME_FURIGANA), string(domain.USER_CHANGED_CONTENT_INCHARGE_LASTNAME),
		string(domain.USER_CHANGED_CONTENT_INCHARGE_LASTNAME_FURIGANA), string(domain.USER_CHANGED_CONTENT_INCHARGE_EMAIL), string(domain.USER_CHANGED_CONTENT_INCHARGE_PHONE),
		string(domain.USER_CHANGED_CONTENT_INCHARGE_DEPARTMENT), string(domain.USER_CHANGED_CONTENT_INCHARGE_POSITION),
		string(domain.USER_CHANGED_CONTENT_MANAGER_ZIPCODE), string(domain.USER_CHANGED_CONTENT_MANAGER_PROVINCE), string(domain.USER_CHANGED_CONTENT_MANAGER_CITY),
		string(domain.USER_CHANGED_CONTENT_MANAGER_ADDRESS), string(domain.USER_CHANGED_CONTENT_MANAGER_ADDRESS2), string(domain.USER_CHANGED_CONTENT_MANAGER_BIRTHDAY),
		string(domain.USER_CHANGED_CONTENT_ASSET), string(domain.USER_CHANGED_CONTENT_ANNUAL_INCOME), string(domain.USER_CHANGED_CONTENT_PRO_INVEST_ASSET),
		string(domain.USER_CHANGED_CONTENT_PRO_INVEST_FINANCE_ASSET), string(domain.USER_CHANGED_CONTENT_PERSONAL_BIRTHDAY), string(domain.USER_CHANGED_CONTENT_NATIONALITY),
		string(domain.USER_CHANGED_CONTENT_GENDER), string(domain.USER_CHANGED_CONTENT_PURPOSE), string(domain.USER_CHANGED_CONTENT_OPERATINGFUNDS),
		string(domain.USER_CHANGED_CONTENT_PROFESSIONAL_INVESTOR), string(domain.USER_CHANGED_CONTENT_IS_ADDRESS_CHANGE), string(domain.USER_CHANGED_CONTENT_MANAGER_IS_ADDRESS_CHANGE), string(domain.USER_CHANGED_CONTENT_INVALID_REASON)}

	user, err := domain.User{}.GetById(ctx, info.UserId)
	if err != nil {
		return
	}

	info.Id = 0
	for _, content := range contentTypes {

		if prevData[content] == nil {
			continue
		}

		// 個人ユーザーはFirstNameとFamilyNameでnameが自動更新されるため無視します。(法人名のみ有効)
		if content == string(domain.USER_CHANGED_CONTENT_CORP_NAME) && user.UserType == domain.PRIVATE {
			continue
		}

		prev := prevData[content].(string)
		curr := currData[content].(string)

		if prev != curr {
			createHistory(ctx, info, string(domain.USER_HISTORY_CONTENTS))
			domain.UserChangedContent{}.Create(ctx, content, info.Id, prev, curr, "", "", "", "")
		}
	}
}

// 投資等級変更状態の比較
func compareInvestment(ctx context.Context, prevData, currData map[string]interface{}, info *domain.UserInfoHistory) {
	var prev, curr string
	if val, ok := prevData[string(domain.USER_CHANGED_CONTENT_INVEST_GRADE)]; ok {
		prev = val.(string)
	}
	if val, ok := currData[string(domain.USER_CHANGED_CONTENT_INVEST_GRADE)]; ok {
		curr = val.(string)
	}

	if prev == curr {
		return
	}

	info.Id = 0
	createHistory(ctx, info, string(domain.USER_HISTORY_INVEST_GRADE))
	domain.UserChangedContent{}.Create(ctx, string(domain.USER_CHANGED_CONTENT_INVEST_GRADE), info.Id, prev, curr, "", "", "", "")
}

// ドキュメントを比較するための集計メソッド
func compareDocument(ctx context.Context, before, after Document, info *domain.UserInfoHistory) {
	// 本人確認書類1の比較 (前、後)
	compareDocumentFiles(ctx, before.Identification1Front, after.Identification1Front, string(domain.USER_CHANGED_CONTENT_IDENTIFICATION1_FRONT), info)
	compareDocumentFiles(ctx, before.Identification1Back, after.Identification1Back, string(domain.USER_CHANGED_CONTENT_IDENTIFICATION1_BACK), info)
	// 本人確認書類2の比較 (前、後)
	compareDocumentFiles(ctx, before.Identification2Front, after.Identification2Front, string(domain.USER_CHANGED_CONTENT_IDENTIFICATION2_FRONT), info)
	compareDocumentFiles(ctx, before.Identification2Back, after.Identification2Back, string(domain.USER_CHANGED_CONTENT_IDENTIFICATION2_BACK), info)
	// 履歴事項全部証明書の比較
	compareDocumentFiles(ctx, before.CertifiedCopyFiles, after.CertifiedCopyFiles, string(domain.USER_CHANGED_CONTENT_ALL_CERTIFICATE_HISTORY), info)
	// 委任状の比較
	compareDocumentFiles(ctx, before.PoaFiles, after.PoaFiles, string(domain.USER_CHANGED_CONTENT_POWER_OF_ATTORNEY), info)
}

// 該当ドキュメント比較
func compareDocumentFiles(ctx context.Context, before, after []*domain.UserFile, changedContentName string, info *domain.UserInfoHistory) {

	if len(before) <= 0 && len(after) <= 0 {
		return
	}

	info.Id = 0

	// 削除処理 (IDが異なっていても同じ名前が存在する場合は無視)
	for _, pu := range before {

		var find = false
		for _, cu := range after {
			if (pu.OriginalName == cu.OriginalName) || (pu.Id == cu.Id) {
				find = true
				break
			}
		}

		if !find {
			createHistory(ctx, info, string(domain.USER_HISTORY_DOCUMENT))
			domain.UserChangedContent{}.Create(ctx, changedContentName, info.Id, pu.OriginalName, string(domain.USER_HISTORY_DOCUMENT_DEL), "", "", pu.FileUrl, "")
		}
	}

	// 追加処理
	for _, cu := range after {
		var find = false
		for _, pu := range before {
			if cu.Id == pu.Id {
				find = true
				break
			}
		}

		if !find {
			createHistory(ctx, info, string(domain.USER_HISTORY_DOCUMENT))
			domain.UserChangedContent{}.Create(ctx, changedContentName, info.Id, string(domain.USER_HISTORY_DOCUMENT_ADD), cu.OriginalName, "", "", "", cu.FileUrl)
		}
	}
}

// 反社会勢力情報の変更状態比較
func compareAntisocial(ctx context.Context, before, after AntiSocialInfo, info *domain.UserInfoHistory) {

	var prevCheckFiles = before.AntisocialCheckFile
	var currCheckFiles = after.AntisocialCheckFile

	var prevAntisocialCheck = before.AntiSocialCheck
	var currAntisocialCheck = after.AntiSocialCheck

	if prevAntisocialCheck == nil || currAntisocialCheck == nil {
		return
	}

	info.Id = 0

	// 反社会処理者
	compareAntisocialChecker(ctx, prevAntisocialCheck.Checker, currAntisocialCheck.Checker, info)

	// 反社会処理日
	compareAntisocialCheckDate(ctx, prevAntisocialCheck.CheckDate, currAntisocialCheck.CheckDate, info)

	// 反社会非承認理由
	compareAntisocialInvalidReason(ctx, prevAntisocialCheck.InvalidReason, currAntisocialCheck.InvalidReason, info)

	// 反社会処理状態
	compareAntisocialCheckStatus(ctx, prevAntisocialCheck.CheckStatus, currAntisocialCheck.CheckStatus, info)

	// 反社会ファイル比較
	compareAntisocialCheckFiles(ctx, prevCheckFiles, currCheckFiles, info)
}

// 反社会処理者比較
func compareAntisocialChecker(ctx context.Context, prev, curr string, info *domain.UserInfoHistory) {

	if prev != curr {

		createHistory(ctx, info, string(domain.USER_HISTORY_ANTISOCIAL))
		domain.UserChangedContent{}.Create(ctx, string(domain.USER_CHANGED_CONTENT_ANTISOCIAL_CHECKER), info.Id, prev, curr, "", "", "", "")
	}
}

// 反社会処理日比較
func compareAntisocialCheckDate(ctx context.Context, prev, curr jsonutils.JSONTime, info *domain.UserInfoHistory) {

	if prev != curr {

		createHistory(ctx, info, string(domain.USER_HISTORY_ANTISOCIAL))
		domain.UserChangedContent{}.Create(ctx, string(domain.USER_CHANGED_CONTENT_ANTISOCIAL_CHECK_DATE), info.Id, prev.ToString(), curr.ToString(), "", "", "", "")
	}
}

// 反社会非承認理由比較
func compareAntisocialInvalidReason(ctx context.Context, prev, curr string, info *domain.UserInfoHistory) {

	if prev != curr {

		createHistory(ctx, info, string(domain.USER_HISTORY_ANTISOCIAL))
		domain.UserChangedContent{}.Create(ctx, string(domain.USER_CHANGED_CONTENT_ANTISOCIAL_CHECK_DISAPPROVE_INVALID_REASON), info.Id, prev, curr, "", "", "", "")
	}
}

// 反社会処理状態比較
func compareAntisocialCheckStatus(ctx context.Context, prev, curr domain.ANTISOCIAL_CHECK_STATUS, info *domain.UserInfoHistory) {
	if prev != curr {
		var prevStatus string
		var currStatus string

		switch prev {
		case domain.ANTISOCIAL_CHECK_NONE:
			prevStatus = "NONE"
		case domain.ANTISOCIAL_CHECK_YES:
			prevStatus = string(domain.USER_CHANGED_CONTENT_ANTISOCIAL_CHECK_APPROVE)
		case domain.ANTISOCIAL_CHECK_NO:
			prevStatus = string(domain.USER_CHANGED_CONTENT_ANTISOCIAL_CHECK_DISAPPROVE)
		default:
			logrus.Warnf("unknown antisocial check status: %v", prev)
			return
		}

		switch curr {
		case domain.ANTISOCIAL_CHECK_NONE:
			currStatus = "NONE"
		case domain.ANTISOCIAL_CHECK_YES:
			currStatus = string(domain.USER_CHANGED_CONTENT_ANTISOCIAL_CHECK_APPROVE)
		case domain.ANTISOCIAL_CHECK_NO:
			currStatus = string(domain.USER_CHANGED_CONTENT_ANTISOCIAL_CHECK_DISAPPROVE)
		default:
			logrus.Warnf("unknown antisocial check status: %v", curr)
			return
		}

		createHistory(ctx, info, string(domain.USER_HISTORY_ANTISOCIAL))
		domain.UserChangedContent{}.Create(ctx, string(domain.USER_CHANGED_CONTENT_ANTISOCIAL_CHECK_STATUS), info.Id, prevStatus, currStatus, "", "", "", "")
	}
}

// 反社会ファイル比較
func compareAntisocialCheckFiles(ctx context.Context, prev, curr []*domain.AntisocialCheckFile, info *domain.UserInfoHistory) {

	if len(prev) <= 0 && len(curr) <= 0 {
		return
	}

	// 削除処理 (IDが異なっていても同じ名前が存在する場合は無視)
	for _, pu := range prev {

		var find = false
		for _, cu := range curr {
			if (pu.OriginalName == cu.OriginalName) || (pu.Id == cu.Id) {
				find = true
				break
			}
		}

		if !find {
			createHistory(ctx, info, string(domain.USER_HISTORY_ANTISOCIAL))
			domain.UserChangedContent{}.Create(ctx, string(domain.USER_CHANGED_CONTENT_ANTISOCIAL_CHECK_DOCUMENT), info.Id, pu.OriginalName, string(domain.USER_CHANGED_CONTENT_DEL), "", "", pu.FileUrl, "")
		}
	}

	// 追加処理
	for _, cu := range curr {
		var find = false
		for _, pu := range prev {
			if cu.Id == pu.Id {
				find = true
				break
			}
		}

		if !find {
			createHistory(ctx, info, string(domain.USER_HISTORY_ANTISOCIAL))
			domain.UserChangedContent{}.Create(ctx, string(domain.USER_CHANGED_CONTENT_ANTISOCIAL_CHECK_DOCUMENT), info.Id, string(domain.USER_CHANGED_CONTENT_ADD), cu.OriginalName, "", "", "", cu.FileUrl)
		}
	}
}

// 最大株主比較
func compareShareholder(ctx context.Context, before, after UserInfoHistoryData, info *domain.UserInfoHistory) {
	prevShareholders := before.CorporateShareHolder
	currShareholders := after.CorporateShareHolder

	if len(prevShareholders) <= 0 || len(currShareholders) <= 0 {
		return
	}

	// 最大株主の比較キーをスライスデータ構造に保存
	contentTypes := []string{string(domain.USER_CHANGED_CONTENT_NAME), string(domain.USER_CHANGED_CONTENT_FAMILY_NAME),
		string(domain.USER_CHANGED_CONTENT_NAME2), string(domain.USER_CHANGED_CONTENT_FAMILY_NAME2),
		string(domain.USER_CHANGED_CONTENT_BIRTHDAY), string(domain.USER_CHANGED_CONTENT_ZIPCODE),
		string(domain.USER_CHANGED_CONTENT_ADDRESS), string(domain.USER_CHANGED_CONTENT_ADDRESS2),
		string(domain.USER_CHANGED_CONTENT_PROVINCE), string(domain.USER_CHANGED_CONTENT_CITY)}

	info.Id = 0

	// 削除と修正処理
	for _, ps := range prevShareholders {

		var find = false
		prevMapData := convertTagToMap(reflect.ValueOf(ps).Elem())

		for _, cs := range currShareholders {
			if ps.Id != cs.Id {
				continue
			}

			find = true
			currMapData := convertTagToMap(reflect.ValueOf(cs).Elem())

			for _, content := range contentTypes {

				if prevMapData[content] == nil {
					continue
				}

				prev := prevMapData[content].(string)
				curr := currMapData[content].(string)

				if prev != curr {
					createHistory(ctx, info, string(domain.USER_HISTORY_SHAREHOLDER))
					domain.UserChangedContent{}.Create(ctx, content, info.Id, prev, curr+" ("+ps.Uid+")", "", "", "", "")
				}
			}

			break

		}

		// 変更後に変更前の最大株主が見つからない場合は削除
		if !find {
			for _, content := range contentTypes {

				if prevMapData[content] == nil {
					continue
				}

				prev := prevMapData[content].(string)

				createHistory(ctx, info, string(domain.USER_HISTORY_SHAREHOLDER))
				domain.UserChangedContent{}.Create(ctx, content, info.Id, string(domain.USER_CHANGED_CONTENT_DEL), prev+" ("+ps.Uid+")", "", "", "", "")
			}
		}
	}

	// 追加処理
	for _, cs := range currShareholders {

		var find = false
		for _, ps := range prevShareholders {
			if cs.Id == ps.Id {
				find = true
				break
			}
		}

		// 変更後の最大株主が変更前の最大株主から見つからない場合は、最大株主を追加
		if !find {
			currMapData := convertTagToMap(reflect.ValueOf(cs).Elem())
			for _, content := range contentTypes {

				if currMapData[content] == nil {
					continue
				}

				curr := currMapData[content].(string)

				createHistory(ctx, info, string(domain.USER_HISTORY_SHAREHOLDER))
				domain.UserChangedContent{}.Create(ctx, content, info.Id, string(domain.USER_CHANGED_CONTENT_ADD), curr+" ("+cs.Uid+")", "", "", "", "")
			}
		}
	}
}

func createHistory(ctx context.Context, info *domain.UserInfoHistory, substanceType string) {
	if info.Id > 0 {
		return
	}

	info.SubstanceType = substanceType
	_, err := info.Create(ctx)
	if err != nil {
		logrus.Error("failed to save user's information history")
	}
}

func toUserInfoHistory(ctx context.Context, userId, processId int64, name, email string) *domain.UserInfoHistory {

	var userInfoHistory domain.UserInfoHistory
	userInfoHistory.UserId = userId
	userInfoHistory.ProcessUserId = processId
	if userInfoHistory.UserId != userInfoHistory.ProcessUserId {
		userInfoHistory.Name = name
		userInfoHistory.Email = email
		userInfoHistory.ProcessMethod = domain.USER_HISTORY_PROCESS_OPERATOR
	} else {
		userInfoHistory.Name = name
		userInfoHistory.Email = email
		userInfoHistory.ProcessMethod = domain.USER_HISTORY_PROCESS_MEMBER
	}

	return &userInfoHistory
}

func mergeMaps(maps ...map[string]interface{}) map[string]interface{} {

	result := make(map[string]interface{})
	for _, m := range maps {
		for k, v := range m {
			result[k] = v
		}
	}
	return result
}
