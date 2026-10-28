package service

import (
	"context"
	"encoding/csv"
	"net/http"
	"reflect"
	"strconv"
	"time"

	"github.com/samber/lo"

	"github.com/illuminarean/energy-shares-backend-repo/service/dto"

	"github.com/illuminarean/energy-shares-backend-repo/domain"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/csvutils"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/factory"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/i18nutils"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/modelhelper"
)

var _ IPrivateUserAdminService = (*PrivateUserAdminService)(nil)

type IPrivateUserAdminService interface {
	GetAll(ctx context.Context, id *string, name, email, investorType, auditStatus, requestId *string, sortby, order string, offset, limit int, startDate, endDate string, profileCond domain.ProfileSearchCondition) (int64, []*dto.PrivateUserItem, error)
	GetById(ctx context.Context, id int64) (*dto.PrivateUserAdminDetail, error)
	DownloadCsv(ctx context.Context, w http.ResponseWriter, id, name, email, investorType, auditStatus, requestId *string, sortby, order string, offset, limit int, startDate, endDate string, profileCond domain.ProfileSearchCondition, investGrade, basicInfo, withdrawal, loanDeposit, equityDeposit, newsletterAgreement bool) error
}

type PrivateUserAdminService struct{}

func NewPrivateUserAdminService() IPrivateUserAdminService {
	return &PrivateUserAdminService{}
}

func (s *PrivateUserAdminService) GetAll(ctx context.Context, id *string, name, email, investorType, auditStatus, requestId *string, sortby, order string, offset, limit int, startDate, endDate string, profileCond domain.ProfileSearchCondition) (int64, []*dto.PrivateUserItem, error) {
	var users []domain.PrivateUser
	totalCount, users, err := domain.PrivateUser{}.GetAll(ctx, id, name, email, investorType, auditStatus, requestId, sortby, order, offset, limit, startDate, endDate, profileCond)
	if err != nil {
		return 0, nil, err
	}
	items := make([]*dto.PrivateUserItem, 0)
	for i := range users {
		items = append(items, dto.PrivateUserItem{}.New(&users[i]))
	}
	return totalCount, items, nil
}

func (s *PrivateUserAdminService) GetById(ctx context.Context, id int64) (*dto.PrivateUserAdminDetail, error) {
	var err error = nil

	item, err := domain.PrivateUser{}.GetById(ctx, id)
	if err != nil {
		return nil, err
	}

	equity, err := domain.UserVirtualAccount{}.GetVaByUserId(ctx, id, domain.EQUITY)
	if err != nil {
		return nil, err
	}

	antisocialCheck, err := domain.UserAntisocialCheck{}.GetByCheckUserId(ctx, id)
	if err != nil {
		return nil, err
	}

	account, err := domain.UserAccount{}.GetByUserId(ctx, item.User.Id)
	if err != nil {
		return nil, err
	}

	myNumber, err := domain.MyNumber{}.GetByUserId(ctx, item.User.Id)
	if err != nil {
		return nil, err
	}

	ekycReqLog := domain.EkycRequestLog{}
	ekycResult, err := ekycReqLog.GetLatest(ctx, item.User.Id)
	if err != nil {
		return nil, err
	}

	var user dto.PrivateUserAdminDetail
	user = *dto.PrivateUserAdminDetail{}.New(item, equity, antisocialCheck, account, myNumber, ekycResult)

	err = FindPrivateUserIdentificationFiles(ctx, &user)
	if err != nil {
		return nil, err
	}

	// 重複ユーザーチェック
	duplicates, err := domain.PrivateUser{}.FindDuplicates(ctx, item)
	if err != nil {
		return nil, err
	}
	if len(duplicates) > 0 {
		dups := make([]dto.DuplicateUser, len(duplicates))
		for i, d := range duplicates {
			dups[i] = dto.DuplicateUser{
				CheckedUserID:   d.CheckedUserID,
				DuplicateUserID: d.DuplicateUserID,
				DuplicateReason: d.DuplicateReason,
			}
		}
		user.DuplicateUsers = dups
	} else {
		user.DuplicateUsers = make([]dto.DuplicateUser, 0)
	}

	user.Roles = []string{string(domain.ROLE_USER)}
	if user.EkycResult == nil { // eKYC 결과가 없는 경우
		user.HasEkycImage = false
	} else if user.EkycResult.Created.Time().AddDate(0, 0, 90).Before(time.Now()) { // 90일 이상 지난 경우
		user.HasEkycImage = false
	} else {
		user.HasEkycImage = true
	}

	return &user, nil
}

func FindPrivateUserIdentificationFiles(ctx context.Context, user *dto.PrivateUserAdminDetail) (err error) {
	user.Identification1FrontFile, err = domain.UserFile{}.FindByUserIdAndFileType(ctx, 1, user.Id, domain.IDENTIFICATION1_FRONT)
	if err != nil {
		return err
	}
	user.Identification1BackFile, err = domain.UserFile{}.FindByUserIdAndFileType(ctx, 1, user.Id, domain.IDENTIFICATION1_BACK)
	if err != nil {
		return err
	}

	user.Identification2FrontFile, err = domain.UserFile{}.FindByUserIdAndFileType(ctx, 1, user.Id, domain.IDENTIFICATION2_FRONT)
	if err != nil {
		return err
	}
	user.Identification2BackFile, err = domain.UserFile{}.FindByUserIdAndFileType(ctx, 1, user.Id, domain.IDENTIFICATION2_BACK)
	if err != nil {
		return err
	}
	user.EkycIdentificationFiles, err = domain.UserFile{}.FindByUserIdAndFileType(ctx, 10, user.Id, domain.EKYC_IDENTIFICATION)
	if err != nil {
		return err
	}
	return err
}

func (s *PrivateUserAdminService) DownloadCsv(ctx context.Context, w http.ResponseWriter, id, name, email, investorType, auditStatus, requestId *string, sortby, order string, offset, limit int, startDate, endDate string, profileCond domain.ProfileSearchCondition, investGrade, basicInfo, withdrawal, loanDeposit, equityDeposit, newsletterAgreement bool) error {

	_, userList, err := domain.PrivateUser{}.GetAll(ctx, id, name, email, investorType, auditStatus, requestId, sortby, order, offset, limit, startDate, endDate, profileCond)
	if err != nil {
		return err
	}
	items := make([]*dto.PrivateUserCsv, 0)
	for i := range userList {
		eva, err := domain.UserVirtualAccount{}.GetVaByUserId(ctx, userList[i].User.Id, domain.EQUITY)
		if err != nil {
			return nil
		} else if eva == nil {
			eva = &domain.UserVirtualAccount{}
		}

		items = append(items, dto.PrivateUserCsv{}.New(&userList[i], eva))
	}

	records := s.PrepareAndReturnCsv(items, investGrade, basicInfo, withdrawal, loanDeposit, equityDeposit, newsletterAgreement)
	csvutils.WriteHeader(w, "energyshares_indi_member_list.csv")

	return csv.NewWriter(w).WriteAll(records)
}

func (s *PrivateUserAdminService) PrepareAndReturnCsv(items []*dto.PrivateUserCsv, investGrade, basicInfo, withdrawal, loanDeposit, equityDeposit, newsletterAgreement bool) [][]string {

	type private struct {
		Name     string `json:"name" i18n:"model.user.name"`            // 이름
		Email    string `json:"email" i18n:"model.user.email_id"`       // メールアドレス(ID)
		Birthday string `json:"birthday" i18n:"model.private.birthday"` // 생년월일
	}

	var privates []private
	for _, item := range items {
		p := private{
			Name:     item.Name,
			Email:    item.Email,
			Birthday: item.Birthday.ToFormat("20060102"),
		}

		privates = append(privates, p)
	}

	cols := factory.ReflectJsonFields(reflect.ValueOf(&private{}).Elem())
	if len(cols) <= 0 {
		return nil
	}

	// 투자 등급
	type invest_grade struct {
		InvestorType string `json:"investor_type" i18n:"model.user.investor_type"` // 투자 등급
	}

	var igs []invest_grade
	if investGrade {
		cols = append(cols, factory.ReflectJsonFields(reflect.ValueOf(&invest_grade{}).Elem())...)

		for _, item := range items {
			ig := invest_grade{
				InvestorType: i18nutils.DefaultTranslate("commoncode.INVESTOR_TYPE." + string(item.InvestorType)),
			}

			igs = append(igs, ig)
		}
	}

	// 기본 정보
	type basic_info struct {
		Name_Furi          string `json:"name_furi" i18n:"model.user.name_furi"`             // 이름 후리가나
		Age                string `json:"age" i18n:"model.user.age"`                         // 나이
		Nationality        string `json:"nationality" i18n:"model.user.nationality"`         // 국적
		Gender             string `json:"gender" i18n:"model.user.gender"`                   // 성별
		PhoneNumber        string `json:"phone_number" i18n:"model.user.mobile_phone"`       // 전화번호
		PhoneCertified     string `json:"phone_certified" i18n:"model.user.phone_certified"` // 전화번호 인증일
		ZipCode            string `json:"zip_code" i18n:"model.user.zip_code"`
		Address            string `json:"address" i18n:"model.user.address"`                             // 주소
		Job                string `json:"job" i18n:"model.user.job"`                                     // 직업
		SignUpRequest      string `json:"sign_up_request" i18n:"model.user.sign_up_request"`             // 회원가입 신청일
		SignUpResponse     string `json:"sign_up_response" i18n:"model.user.sign_up_response"`           // 회원가입 승인일
		MarketingAgree     string `json:"marketing_agree" i18n:"model.user.marketing_agree"`             // 마케팅 동의
		LastConnection     string `json:"last_connection" i18n:"model.user.last_login_date"`             // 최근 접속일
		PostalCode         string `json:"postal_code" i18n:"model.user.postal_code"`                     // 본인확인코드
		CodeCertified      string `json:"code_certified" i18n:"model.user.code_certified"`               // 본인확인코드 인증일
		PostalCodeState    string `json:"postal_code_state" i18n:"model.user.postal_code_state"`         // 우편물 발송상태
		PostalCodeSendDate string `json:"postal_code_send_date" i18n:"model.user.postal_code_send_date"` // 우편물 발송일
		Id                 string `json:"id" i18n:"model.user.uid"`                                      // 유니크 ID
		EkycId             string `json:"request_id" i18n:"model.user.ekyc_id"`                          // eKYC ID
	}
	var bis []basic_info
	if basicInfo {
		cols = append(cols, factory.ReflectJsonFields(reflect.ValueOf(&basic_info{}).Elem())...)

		for _, item := range items {
			bi := basic_info{
				Name_Furi:   item.FamilyNameKana + "　" + item.FirstNameKana,
				Age:         strconv.Itoa(modelhelper.YearDiff(item.Birthday.Time(), time.Now())),
				Nationality: i18nutils.DefaultTranslate("commoncode.NATIONALITY_CODE." + string(item.Nationality)),
				Gender:      i18nutils.DefaultTranslate("commoncode.GENDER" + "." + string(item.Gender)),
				PhoneNumber: item.MobilePhone,
				ZipCode:     item.ZipCode,
				Address:     item.Province + " " + item.City + " " + item.Address + " " + lo.FromPtr(item.Address2),
				Job:         i18nutils.DefaultTranslate("commoncode.USER_JOB_TYPE." + string(item.Job)),
				PostalCode:  item.PostalCode,
				Id:          strconv.Itoa(int(item.Id)),
				EkycId:      item.RequestId,
			}

			if item.SmsCertAt.Unix() > 0 {
				bi.PhoneCertified = item.SmsCertAt.ToString()
			}

			if item.IsNewsletter {
				bi.MarketingAgree = i18nutils.DefaultTranslate("model.user.agree")
			} else {
				bi.MarketingAgree = i18nutils.DefaultTranslate("model.user.disagree")
			}

			if item.NewsletterAgreeAt.Unix() > 0 {
				bi.MarketingAgree += " : " + item.NewsletterAgreeAt.ToString()
			}

			if item.PostalCodeStatus {
				bi.PostalCodeState = i18nutils.DefaultTranslate("commoncode.POSTAL_CODE_STATE.POSTAL_CODE_STATE_SENT")
			} else {
				bi.PostalCodeState = i18nutils.DefaultTranslate("commoncode.POSTAL_CODE_STATE.POSTAL_CODE_STATE_NOTSENT")
			}

			if item.CodeApprovalDate.Unix() > 0 {
				bi.CodeCertified = item.CodeApprovalDate.ToString()
			}

			if item.PostalCodeSendDate.Unix() > 0 {
				bi.PostalCodeSendDate = item.PostalCodeSendDate.ToString()
			}

			if item.SignUpDate.Unix() > 0 {
				bi.SignUpRequest = item.SignUpDate.ToString()
			}

			if item.SignUpApprovalDate.Unix() > 0 {
				bi.SignUpResponse = item.SignUpApprovalDate.ToString()
			}

			if item.LastLoginDate.Unix() > 0 {
				bi.LastConnection = item.LastLoginDate.ToString()
			}

			bis = append(bis, bi)
		}
	}

	// 출금계좌
	type withdrawal_info struct {
		BankName    string `json:"bank_name" i18n:"model.user.withdrawal_bank_name"`       // 출금계좌 금융기관명
		BranchName  string `json:"branch_name" i18n:"model.user.withdrawal_branch_name"`   // 출금계좌 지점명
		DepositType string `json:"deposit_type" i18n:"model.user.withdrawal_deposit_type"` // 출금계좌 예금종류
		BankAccount string `json:"bank_account" i18n:"model.user.withdrawal_bank_account"` // 출금계좌 계좌번호
		Depositor   string `json:"depositor" i18n:"model.user.withdrawal_depositor"`       // 출금계좌 예금주
	}
	var wis []withdrawal_info
	if withdrawal {
		cols = append(cols, factory.ReflectJsonFields(reflect.ValueOf(&withdrawal_info{}).Elem())...)

		for _, item := range items {
			wi := withdrawal_info{
				BankName:    item.BankName,
				BranchName:  item.BranchName,
				DepositType: i18nutils.DefaultTranslate("commoncode.DEPOSIT_TYPE_CODE." + string(item.DepositType)),
				BankAccount: item.BankAccount,
				Depositor:   item.Depositor,
			}

			wis = append(wis, wi)
		}
	}

	// 사업형 가상계좌
	type equity_deposit_info struct {
		BankName    string `json:"bank_name" i18n:"model.user.equity_bank_name"`       // 사업형 입금계좌 금융기관명
		BranchName  string `json:"branch_name" i18n:"model.user.equity_branch_name"`   // 사업형 입금계좌 지점명
		Depositor   string `json:"depositor" i18n:"model.user.equity_depositor"`       // 사업형 입금계좌 예금주
		BankAccount string `json:"bank_account" i18n:"model.user.equity_bank_account"` // 사업형 입금계좌 계좌번호
	}
	var edis []equity_deposit_info
	if equityDeposit {
		cols = append(cols, factory.ReflectJsonFields(reflect.ValueOf(&equity_deposit_info{}).Elem())...)

		for _, item := range items {
			edi := equity_deposit_info{
				BankName:    item.EquityVaAccount.VaBankName,
				BranchName:  item.EquityVaAccount.VaBranchName,
				Depositor:   item.EquityVaAccount.VaHolderNameKana,
				BankAccount: item.EquityVaAccount.VaAccountNumber,
			}

			edis = append(edis, edi)
		}
	}

	// 뉴스레터 수신동의일
	type newsletter_agreement struct {
		AgreementDate string `json:"agreement_date" i18n:"model.user.agreement_date"` // 뉴스레터 수신동의일
	}
	var nas []newsletter_agreement
	if newsletterAgreement {
		cols = append(cols, factory.ReflectJsonFields(reflect.ValueOf(&newsletter_agreement{}).Elem())...)

		for _, item := range items {
			var na newsletter_agreement
			if item.NewsletterAgreeAt.Unix() > 0 && item.IsNewsletter {
				na.AgreementDate = item.NewsletterAgreeAt.ToString()
			}

			nas = append(nas, na)
		}
	}

	records := [][]string{}
	records = append(records, cols)

	for index, _ := range privates {

		var datas []string
		datas = append(datas, factory.DataWriteToCsv(reflect.ValueOf(&privates[index]).Elem())...)

		if len(igs) > 0 {
			datas = append(datas, factory.DataWriteToCsv(reflect.ValueOf(&igs[index]).Elem())...)
		}

		if len(bis) > 0 {
			datas = append(datas, factory.DataWriteToCsv(reflect.ValueOf(&bis[index]).Elem())...)
		}

		if len(wis) > 0 {
			datas = append(datas, factory.DataWriteToCsv(reflect.ValueOf(&wis[index]).Elem())...)
		}

		if len(edis) > 0 {
			datas = append(datas, factory.DataWriteToCsv(reflect.ValueOf(&edis[index]).Elem())...)
		}

		if len(nas) > 0 {
			datas = append(datas, factory.DataWriteToCsv(reflect.ValueOf(&nas[index]).Elem())...)
		}

		records = append(records, datas)
	}

	return records
}
