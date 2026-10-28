package domain

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/samber/lo"
	"github.com/segmentio/ksuid"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"

	"github.com/illuminarean/energy-shares-backend-repo/pkg/commoncode"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/factory"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/i18nutils"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/jsonutils"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/serviceutils"
)

// 투자 상태
type INVEST_STATUS string

const (
	INVEST_STATUS_CODE               = "INVEST_STATUS"
	REQUEST_INVEST     INVEST_STATUS = "REQUEST_INVEST"   // 투자 요청
	FAIL_INVEST        INVEST_STATUS = "FAIL_INVEST"      // 투자 불성립
	CANCEL_INVEST      INVEST_STATUS = "CANCEL_INVEST"    // 투자 취소
	SUCCESS_INVEST     INVEST_STATUS = "SUCCESS_INVEST"   // 투자 완료
	OPERATION_INVEST   INVEST_STATUS = "OPERATION_INVEST" // 운용중
	REFUND_INVEST      INVEST_STATUS = "REFUND_INVEST"    // 운용 종료

)

// investStatusLabels はINVEST_STATUSの日本語表示名(MIR-1295)。手動での
// 状態変更を拒否したときのエラーメッセージに使う。管理画面のプルダウンの
// 表示名(commoncode経由のi18n変換)とは別に、バックエンド側で組み立てる
// エラーメッセージ専用に持つ。
var investStatusLabels = map[INVEST_STATUS]string{
	REQUEST_INVEST:   "申込中",
	FAIL_INVEST:      "投資不成立",
	CANCEL_INVEST:    "取消",
	SUCCESS_INVEST:   "成立",
	OPERATION_INVEST: "運用中",
	REFUND_INVEST:    "償還済み",
}

// Label はINVEST_STATUSの日本語表示名を返す(MIR-1295)。未知の値は
// 元の文字列をそのまま返す(想定外の値でパニックさせない)。
func (s INVEST_STATUS) Label() string {
	if label, ok := investStatusLabels[s]; ok {
		return label
	}
	return string(s)
}

// newCannotChangeInvestStatusError は、CanManuallyChangeInvestStatusが
// 遷移を拒否したときのエラーを組み立てる(MIR-1295)。domainパッケージは
// clean-arch-domain(cdomain.ErrCannotManuallyChangeInvestStatus)を参照
// できない(循環import)ため、ここではメッセージだけのプレーンなエラーを
// 返す。呼び出し側(usecase層)でcdomain.ErrCannotManuallyChangeInvestStatus
// と組み合わせて包む。
func newCannotChangeInvestStatusError(from, to INVEST_STATUS) error {
	return fmt.Errorf("この出資は%sのため、%sには変更できません", from.Label(), to.Label())
}

func (Share) GetCodeGroup(loc string) commoncode.CodeGroup {

	codes := make([]commoncode.Code, 0, 0)
	codes = append(codes, commoncode.Code{Code: string(REQUEST_INVEST), Name: i18nutils.Translate("commoncode.INVEST_STATUS.REQUEST_INVEST", loc), Order: len(codes) + 1})
	codes = append(codes, commoncode.Code{Code: string(FAIL_INVEST), Name: i18nutils.Translate("commoncode.INVEST_STATUS.FAIL_INVEST", loc), Order: len(codes) + 1})
	codes = append(codes, commoncode.Code{Code: string(CANCEL_INVEST), Name: i18nutils.Translate("commoncode.INVEST_STATUS.CANCEL_INVEST", loc), Order: len(codes) + 1})
	codes = append(codes, commoncode.Code{Code: string(SUCCESS_INVEST), Name: i18nutils.Translate("commoncode.INVEST_STATUS.SUCCESS_INVEST", loc), Order: len(codes) + 1})
	codes = append(codes, commoncode.Code{Code: string(OPERATION_INVEST), Name: i18nutils.Translate("commoncode.INVEST_STATUS.OPERATION_INVEST", loc), Order: len(codes) + 1})
	codes = append(codes, commoncode.Code{Code: string(REFUND_INVEST), Name: i18nutils.Translate("commoncode.INVEST_STATUS.REFUND_INVEST", loc), Order: len(codes) + 1})

	return commoncode.CodeGroup{
		GroupCode: INVEST_STATUS_CODE,
		Codes:     codes,
	}
}

type SHARE_TYPE string

const (
	BASIC_SHARE  SHARE_TYPE = "BASIC"  // 일반 투자로 인한 거래
	MANUAL_SHARE SHARE_TYPE = "MANUAL" // 수동 투자로 인한 거래

)

type Share struct {
	Id                          string             `json:"id" swagger:"desc(Shares ID)" xorm:"unique varchar(255) pk"`
	ProjectId                   int64              `json:"project_id" swagger:"desc(펀드 ID)" xorm:"index"`
	CategoryType                PROJECT_CATEGORY   `json:"category_type" swagger:"desc(카테고리 ID)"`
	ProjectType                 PROJECT_TYPE       `json:"project_type" swagger:"desc(프로젝트 타입)"`
	OwnerId                     int64              `json:"owner_id" swagger:"desc(소유자 ID)" xorm:"not null index"` // 소유자 ID
	UpdateId                    int64              `json:"update_id" swagger:"desc(수정자 ID)" xorm:"not null"`      // 수정자 ID
	Desc                        string             `json:"desc" swagger:"desc(설명)" `
	InvestStatus                INVEST_STATUS      `json:"invest_status" swagger:"desc(설명)"`                                                               // 투자상태
	Amount                      int64              `json:"amount" swagger:"desc(총액)"`                                                                      // 주식 금액
	Qty                         int64              `json:"qty" swagger:"desc(수량)"`                                                                         // 주식 수량
	Price                       int64              `json:"price" swagger:"desc(1구좌 가격)" `                                                                  // 1구좌당 가격
	ShareType                   SHARE_TYPE         `json:"share_type" swagger:"desc(share_type BASIC_SHARE, MANUAL_SHARE)" xorm:"default 'BASIC_SHARE' " ` // 1구좌당 가격
	InvestDate                  jsonutils.JSONTime `json:"invest_date" xorm:"not null"`                                                                    // 투자 요청 일자
	CommitmentDate              jsonutils.JSONTime `json:"commitment_date"`                                                                                // 투자 약정 일자
	Created                     jsonutils.JSONTime `json:"created" xorm:"created"`
	Updated                     jsonutils.JSONTime `json:"updated" xorm:"updated"`
	Deleted                     jsonutils.JSONTime `json:"deleted" xorm:"deleted"`
	ContractBeforePageCheckDate jsonutils.JSONTime `json:"contract_before_page_check_date"  swagger:"desc(계약성립 전 서면 확인 날짜)"`
	ContractAfterPageCheckDate  jsonutils.JSONTime `json:"contract_after_page_check_date"  swagger:"desc(계약성립 전 서면 확인 날짜)"`
	SilentPartnershipsCheckDate jsonutils.JSONTime `json:"silent_partnerships_check_date"  swagger:"desc(계약성립 전 서면 확인 날짜)"`
	ProjectReportCheckDate      jsonutils.JSONTime `json:"project_report_check_date"  swagger:"desc(계약성립 전 서면 확인 날짜)"`
}

func (s Share) TableName() string {
	return "share"
}

func (s Share) SelectAll() string {
	return s.TableName() + ".*"
}

func (s Share) WithInvestStatus(investStatus INVEST_STATUS) Share {
	s.InvestStatus = investStatus
	return s
}

// CanCancelInvest は投資申請中またはクーリングオフ対象の投資完了からキャンセルへ遷移可能かを返す。
func (s Share) CanCancelInvest() bool {
	return s.InvestStatus == REQUEST_INVEST || s.InvestStatus == SUCCESS_INVEST
}

type ShareTypeStatistics struct {
	LoanRatio   float64 `json:"loan_ratio"`
	EquityRatio float64 `json:"equity_ratio"`
}

type UserShareAmountInfo struct {
	InvestTotalAmount int64 `json:"invest_total_amount"`
	InvestTotalCount  int64 `json:"invest_total_count"`

	InvestOperationAmount int64 `json:"invest_operation_amount"`
	InvestOperationCount  int64 `json:"invest_operation_count"`

	InvestSuccessAmount int64 `json:"invest_success_amount"`
	InvestSuccessCount  int64 `json:"invest_success_count"`

	InvestRequestAmount int64 `json:"invest_request_amount"`
	InvestRequestCount  int64 `json:"invest_request_count"`
}

// 지분 상태 업데이트
func (s *Share) UpdateInvestStatus(ctx context.Context, status INVEST_STATUS, updateId int64, investMailType INVEST_MAIL_TYPE) error {
	factory.Logger(ctx).WithFields(logrus.Fields{
		"project_id":     s.ProjectId,
		"owner_id":       s.OwnerId,
		"invest_id":      s.Id,
		"invest_qty":     s.Qty,
		"current_status": s.InvestStatus,
		"target_status":  status,
		"update_id":      updateId,
	}).Info("Change Invest Status")

	beforeStatus := s.InvestStatus
	s.InvestStatus = status

	dr, err := DepositRequest{}.GetDepositRequestByShareId(ctx, s.Id)
	if err != nil {
		return errors.Wrap(err, "fail to get deposit request by share id")
	}
	if dr == nil {
		return errors.New("Not Found Deposit Request")
	}
	owner, err := User{}.GetById(ctx, s.OwnerId)
	if err != nil {
		return errors.Wrap(err, "fail to get by owner")
	}
	if owner == nil {
		return factory.ErrorNotFound.New(err, "Not Found ")
	}

	if updateId == 0 {
		updateId, err = AdminUser{}.GetSuperAdminId(ctx)
		if err != nil {
			return errors.Wrap(err, "fail to get super admin id")
		}
	}

	p, err := Project{}.GetById(ctx, s.ProjectId)
	if err != nil {
		return factory.ErrorNotFound.New(err, "Not Found Product")
	}

	switch s.InvestStatus {
	case REQUEST_INVEST:
		if err := s.RequestInvest(ctx); err != nil {
			return errors.Wrap(err, "fail to request invest")
		}

		if investMailType == IMMEDIATE || investMailType == NOT_IMMEDIATE {
			// 취소 대기가 아닌 경우에 메일 발송
			// MIR-1251: メール送信はDBのロールバックで取り消せないため、
			// トランザクション内から呼ばれた場合はコミット後まで遅延させる
			// (AfterCommitはトランザクション外なら即時実行するので、既存の
			// 呼び出し元の挙動は変わらない)。RequestInvestMailは残高で文面を
			// 分岐するため、残高の読み取り自体は遅延させずここ(WithdrawBalance
			// より前)で行う。読み取りごとAfterCommitの中へ入れると、AfterCommit
			// が遅延実行される経路(admin-api/user-apiにMIR-1202の
			// factory.WithAfterCommit配線が入った場合)で、同じトランザクション内
			// の後続のWithdrawBalanceが確定した後の残高を読んでしまい、既に
			// 全額投資済みの会員に「入金してください」のメールが送られる。
			userAccount, err := UserAccount{}.GetByUserId(ctx, owner.Id)
			if err != nil {
				// 残高が読めない場合はRequestInvestMailに渡すuserAccountがnilに
				// なる。RequestInvestMailはuserAccount.Balanceへ直接アクセス
				// するため、nilのまま呼ぶとAfterCommit実行時(将来コミット後
				// まで遅延される経路ができた場合はコミット成功後)にnilポインタ
				// 参照でpanicする。メール送信自体を諦め、ここではログのみ残す。
				factory.Logger(ctx).Error(factory.ErrorDB.New(err, "Not Found User Account"))
			} else {
				_ = factory.AfterCommit(ctx, func(commitCtx context.Context) error {
					if err := s.RequestInvestMail(commitCtx, owner, p, dr, userAccount); err != nil {
						factory.Logger(commitCtx).Error(err)
					}
					return nil
				})
			}
		}

		if err := CreateRequestInvestmentEvent(ctx, owner, s, p); err != nil {
			return errors.Wrap(err, "fail to send contract investment event")
		}

	case REFUND_INVEST:
		err := s.RefundInvest(ctx)
		if err != nil {
			return errors.Wrap(err, "fail to refund invest")
		}

	case OPERATION_INVEST:
		err := s.OperationInvest(ctx)
		if err != nil {
			return errors.Wrap(err, "fail to refund invest")
		}

	case SUCCESS_INVEST:
		err := s.SuccessInvest(ctx, dr, updateId, p, owner)
		if err != nil {
			return errors.Wrap(err, "fail to success invest")
		}
		// MIR-1251: 成立メールもコミット後まで遅延させる。balanceMatch経由の
		// 呼び出しでは、この時点でWithdrawBalance自体は既に成功しているが、
		// 呼び出し元のDepositCheckProcessがこの後さらにdeposit_requestの
		// 状態更新(d.Update)を控えており、同じトランザクション内にあるため
		// そちらが失敗するとここまでの変更も含めて全体がロールバックされる。
		// メールを即時送信するとロールバック後も「投資が成立した」という
		// 誤ったメールが会員に届いたまま残ってしまう。
		_ = factory.AfterCommit(ctx, func(commitCtx context.Context) error {
			if err := s.SuccessInvestMail(commitCtx, owner, p, dr, investMailType); err != nil {
				factory.Logger(commitCtx).Error(errors.Wrap(err, "fail to send success invest mail"))
			}
			return nil
		})
		if err := CreateCompletedInvestmentEvent(ctx, s, owner, p, dr); err != nil {
			return errors.Wrap(err, "fail to send completed investment event")
		}

	case CANCEL_INVEST:
		if err = s.CancelInvest(ctx, updateId, dr, p, err); err != nil {
			return errors.Wrap(err, "fail to cancel invest")
		}

		// MIR-1251: 取消メールもコミット後まで遅延させる(理由はSUCCESS_INVESTと同じ)。
		_ = factory.AfterCommit(ctx, func(commitCtx context.Context) error {
			if err := s.CancelInvestMail(commitCtx, owner, p, investMailType); err != nil {
				factory.Logger(commitCtx).Error(errors.Wrap(err, "fail to send cancel invest mail"))
			}
			return nil
		})

		if err = CreateInvestmentEventsByDepositStatus(ctx, s.CommitmentDate.Time(), dr); err != nil {
			return errors.Wrap(err, "fail to send investment events by deposit status")
		}

		// 입금 필요액 자동 해소
		if err = s.ResolveDepositAmount(ctx); err != nil {
			return errors.Wrap(err, "fail to resolve deposit amount")
		}

	case FAIL_INVEST:
		if err = s.FailInvest(ctx, updateId, dr, err); err != nil {
			return errors.Wrap(err, "fail to fail invest")
		}

		if err = CreateInvestmentEventsByDepositStatus(ctx, s.CommitmentDate.Time(), dr); err != nil {
			return errors.Wrap(err, "fail to send investment events by deposit status")
		}

		// 입금 필요액 자동 해소
		if err = s.ResolveDepositAmount(ctx); err != nil {
			return errors.Wrap(err, "fail to resolve deposit amount")
		}

	default:
		return factory.ErrorParameter.New(errors.New("Not Supported Status"))
	}

	_, err = TransactionInfoHistory{}.Create(ctx, string(beforeStatus), string(s.InvestStatus), s.Id, updateId)
	if err != nil {
		return errors.Wrap(err, "fail to create transaction info history")
	}

	return nil
}

// FailInvest 투자 불성립
func (s *Share) FailInvest(ctx context.Context, updateId int64, dr *DepositRequest, err error) error {
	s.InvestStatus = FAIL_INVEST
	s.UpdateId = updateId
	s.CommitmentDate = jsonutils.JSONTime(time.Now()) // 약정 일자
	if err := dr.DepositCancel(ctx); err != nil {
		return factory.ErrorDB.New(err, "Fail Cancel Deposit Request")
	}
	if err := s.Update(ctx); err != nil {
		return factory.ErrorDB.New(err, "Fail Update Share Request")
	}

	return nil
}

// CancelInvest 투자 취소
func (s *Share) CancelInvest(ctx context.Context, updateId int64, dr *DepositRequest, p *Project, err error) error {
	s.InvestStatus = CANCEL_INVEST
	s.UpdateId = updateId
	if err := dr.DepositCancel(ctx); err != nil {
		return factory.ErrorDB.New(err, "Fail Cancel Deposit Request")
	}
	s.CommitmentDate = jsonutils.JSONTime(time.Now()) // 약정 일자
	if err := s.Update(ctx); err != nil {
		return factory.ErrorDB.New(err, "Fail Update Share Request")
	}
	isUpdated := false
	if s.ShareType == MANUAL_SHARE {
		p.IsManualInvest = false
		isUpdated = true
	}

	if p.IsWaitInvest {
		waiters, err := InvestmentWaiters{}.GetByProjectIdAndStatuses(ctx, p.Id, []INVESTMENT_WAITING_TYPE{WAITING, DEPOSIT_REQUEST, PROPOSAL})
		if err != nil {
			return errors.Wrap(err, "fail to get investment waiters")
		}
		if len(waiters) == 0 {
			// キャンセル後のファンドの募集状況を取得
			checkProject, err := Project{}.GetById(ctx, p.Id)
			if err != nil {
				return errors.Wrap(err, "fail to get project by id")
			}
			if !checkProject.HasReachedTarget() && checkProject.ProjectStatus != END_FUNDING {
				// キャンセル待機者が存在せず、かつ募集条件に到達していない場合はキャンセル待機状態を解除する
				// ただし、募集終了ステータスの場合はキャンセル待ち状態を維持する
				p.IsWaitInvest = false
				isUpdated = true
			}
		}
	}

	if isUpdated {
		if err := p.Update(ctx); err != nil {
			return factory.ErrorDB.New(err, "Fail Update Project Request")
		}
	}

	return err
}

// SuccessInvest 투자 성공
func (s *Share) SuccessInvest(ctx context.Context, dr *DepositRequest, updateId int64, p *Project, owner *User) error {
	s.InvestStatus = SUCCESS_INVEST
	err := dr.OkDeposit(ctx)
	if err != nil {
		return errors.Wrap(err, "fail to ok deposit")
	}
	s.UpdateId = updateId
	s.CommitmentDate = jsonutils.JSONTime(time.Now()) // 약정 일자
	if err := s.Update(ctx); err != nil {
		return factory.ErrorDB.New(err, "Fail Update Share Request")
	}

	var managerName string
	if owner.UserType == CORPORATE {
		corporateUser, err := CorporateUser{}.GetById(ctx, owner.Id)
		if err == nil && corporateUser.ManagerFamilyName != nil && corporateUser.ManagerFirstName != nil {
			managerName = fmt.Sprintf("担当者 %s%s", *corporateUser.ManagerFamilyName, *corporateUser.ManagerFirstName)
		}
	}
	var address string
	if owner.Address2 != nil {
		address = fmt.Sprintf("%s%s%s %s", owner.Province, owner.City, owner.Address, *owner.Address2)
	} else {
		address = fmt.Sprintf("%s%s%s", owner.Province, owner.City, owner.Address)
	}

	contract := &Contract{
		ShareId:     s.Id,
		ShareQty:    s.Qty,
		ShareAmount: s.Amount,
		Status:      CONTRACT_NONE,
		ProjectId:   s.ProjectId,
		UserName:    owner.Name,
		Address:     address,
		ManagerName: managerName,
		InvestDate:  s.InvestDate,
	}
	_, err = contract.Create(ctx)
	if err != nil {
		return errors.Wrap(err, "fail to create contract")
	}
	return nil
}

// OperationInvest 운용중
func (s *Share) OperationInvest(ctx context.Context) error {
	s.InvestStatus = OPERATION_INVEST
	if err := s.Update(ctx); err != nil {
		return factory.ErrorDB.New(err, "Fail Update Share Request")
	}
	return nil
}

// RefundInvest 투자 상환 완료
func (s *Share) RefundInvest(ctx context.Context) error {
	s.InvestStatus = REFUND_INVEST
	if err := s.Update(ctx); err != nil {
		return factory.ErrorDB.New(err, "Fail Update Share Request")
	}
	return nil
}

// ResolveDepositAmount 현재 투자를 제외한 나머지 투자요청들을 찾아서 입금 해소 처리를 합니다.
func (s *Share) ResolveDepositAmount(ctx context.Context) error {
	if IsFundFailureProcessing(ctx) {
		return nil
	}

	depositRequests, err := DepositRequest{}.GetNoDepositRequestByUserId(ctx, s.OwnerId)
	if err != nil {
		return errors.Wrap(err, "fail to get no deposit request by user id")
	}

	for _, depositRequest := range depositRequests {
		// 입금 해소 로직은 현재 투자를 제외한 나머지 투자들에 대해서 수행합니다.
		if depositRequest.ShareId == s.Id {
			continue
		}

		// 현재 계좌 정보 조회
		account, err := UserAccount{}.GetByUserId(ctx, depositRequest.UserId)
		if err != nil {
			return errors.Wrap(err, "fail to get user account by user id")
		}

		// 입금 해소 처리 (예치금 & 입금필요금액)
		if err = account.DepositCheckProcess(ctx, &depositRequest, NOT_IMMEDIATE); err != nil {
			return errors.Wrap(err, "fail to deposit check process")
		}
	}

	return nil
}

// RequestInvest 투자 요청
func (s *Share) RequestInvest(ctx context.Context) error {
	s.InvestStatus = REQUEST_INVEST
	if err := s.Update(ctx); err != nil {
		return factory.ErrorDB.New(err, "Fail Update Share Request")
	}

	return nil
}

// CancelInvestMail 투자 취소 메일 발송
func (s *Share) CancelInvestMail(ctx context.Context, owner *User, p *Project, investMailType INVEST_MAIL_TYPE) error {
	var mail Mail

	// 프로젝트 상태가 무효 상태라면 투자 취소 메일을 보내지 않고 투자 무효 메일을 보냅니다.
	if p.ProjectStatus == INVALID_FUNDING {
		err := mail.InvestCancelDueToInvalidityMail(
			ctx, serviceutils.GetServiceInfoMail(), owner.Email, owner.Name, p.Name,
			s.Created.ToString(), jsonutils.JSONTime(time.Now()).ToString(),
		)
		if err != nil {
			factory.Logger(ctx).Error(factory.ErrorDB.New(err, "Fail invest cancellation due to product invalidity mail send"))
		}
		return nil
	}

	// 투자대기자가 입금기한 내에 입금하지 않았을 때 / 임의로 취소하였을 때
	if investMailType == WAITER_REQUEST {
		err := mail.InvestStandbyCanceledMail(ctx, serviceutils.GetServiceInfoMail(), owner.Email, owner.Name, p.Name)
		if err != nil {
			factory.Logger(ctx).Error(factory.ErrorDB.New(err, "Fail Invest Cancel Mail Send"))
		}
	} else {
		err := mail.InvestCancelMail(
			ctx, serviceutils.GetServiceInfoMail(), owner.Email, owner.Name, p.Name,
			s.Created.ToString(), jsonutils.JSONTime(time.Now()).ToString(),
		)
		if err != nil {
			factory.Logger(ctx).Error(factory.ErrorDB.New(err, "Fail Invest Cancel Mail Send"))
		}
	}
	return nil
}

// SuccessInvestMail 투자 성공 메일 발송
func (s *Share) SuccessInvestMail(ctx context.Context, owner *User, p *Project, dr *DepositRequest, investMailType INVEST_MAIL_TYPE) error {
	var mail Mail
	switch investMailType {
	case IMMEDIATE:
		// 8-7-1, 2A. 투자가능금액>=투자신청금액 발송
		err := mail.InvestImmediateSuccessMail(ctx, serviceutils.GetServiceInfoMail(), owner.Email, owner.Name, p.Name, strconv.Itoa(int(s.Amount)))
		if err != nil {
			factory.Logger(ctx).Error(factory.ErrorAWSSendMail.New(errors.Wrap(err, "Fail Invest Immediate Success Mail Send")))
		}
	case NOT_IMMEDIATE:
		// 8-7 발송
		err := mail.InvestSuccessMail(ctx, serviceutils.GetServiceInfoMail(), owner.Email, owner.Name, p.Name, s.Updated.Time().Format(time.DateTime), dr.DepositDate.ToString(), strconv.FormatInt(dr.DepositAmount, 10))
		if err != nil {
			factory.Logger(ctx).Error(factory.ErrorAWSSendMail.New(errors.Wrap(err, "Fail Invest Success Mail")))
		}
	case WAITER_IMMEDIATE:
		// 4B. 순번이 된 투자가의 예치금이 충분하여 바로 투자완료처리
		err := mail.InvestStandbyImmediateSuccessMail(ctx, serviceutils.GetServiceInfoMail(), owner.Email, owner.Name, p.Name, strconv.Itoa(int(s.Amount)))
		if err != nil {
			factory.Logger(ctx).Error(factory.ErrorAWSSendMail.New(errors.Wrap(err, "Fail Invest Standby Immediate Success Mail")))
		}
	case LOTTERY_MANUAL:
		// 抽選の手動当選は usecase 側でコミット後にメール送信する。
		return nil
	default:
		return errors.New("Not Supported Invest Success Mail Type")
	}
	return nil
}

// RequestInvestMail 투자 요청 메일 발송 (투자금 부족시)
func (s *Share) RequestInvestMail(ctx context.Context, owner *User, p *Project, dr *DepositRequest, userAccount *UserAccount) error {
	var mail Mail

	// 가상 계좌 정보
	va, err := UserVirtualAccount{}.GetVaByUserId(ctx, owner.Id, p.ProjectType)
	if err != nil {
		factory.Logger(ctx).Error(factory.ErrorAWSSendMail.New(errors.Wrap(err, "Fail to get va by user id")))
	}

	if lo.FromPtr(userAccount.Balance) == 0 {
		err := mail.InvestRequestMail(ctx, serviceutils.GetServiceInfoMail(), owner.Email, owner.Name, p.Name, jsonutils.JSONTime.ToString(s.InvestDate),
			jsonutils.JSONTime.ToString(dr.ExpireDepositDate), strconv.Itoa(int(s.Amount)), dr.DepositBankName, dr.DepositBranchName, dr.DepositAccountNumber,
			va.VaHolderNameKana, p.ProjectType)
		if err != nil {
			factory.Logger(ctx).Error(factory.ErrorAWSSendMail.New(errors.Wrap(err, "Fail Invest Deposit Request Mail Send")))
		}
		return nil
	}

	if dr.NeedAmount() > lo.FromPtr(userAccount.Balance) { // 잔액이 충분하지 않은 경우
		lackAmount := s.Amount - (lo.FromPtr(userAccount.Balance) + dr.DepositAmount)

		// [1.입금 요청]-[1B. 0 < 예치금 < 투자 신청 금액]
		err = mail.InvestInsufficientBalanceMail(ctx, serviceutils.GetServiceInfoMail(), owner.Email, owner.Name, strconv.Itoa(int(lackAmount)),
			p.Name, jsonutils.JSONTime.ToString(s.InvestDate),
			dr.DepositBankName, dr.DepositBranchName, dr.DepositAccountNumber, va.VaHolderNameKana,
			strconv.Itoa(int(s.Amount)), jsonutils.JSONTime.ToString(dr.ExpireDepositDate))
		if err != nil {
			factory.Logger(ctx).Error(factory.ErrorAWSSendMail.New(errors.Wrap(err, "Fail Invest Insufficient Balance Mail Send")))
		}
	}

	return nil
}

// RefundOperationInvestsByProject は、ファンドの運用中(OPERATION_INVEST)の出資を、すべて償還済み
// (REFUND_INVEST)に付け替える(MIR-272)。呼ぶのは、分配償還ジョブが最後の償還(残元本 0)を払ったときと、
// ファンドを償還完了にするとき。ファンドの状態は変えない。
func (Share) RefundOperationInvestsByProject(ctx context.Context, projectId, uid int64) error {
	var items []Share
	q := factory.DB(ctx).Table("share").Where("share.project_id = ?", projectId)
	q = q.And("share.invest_status = ?", OPERATION_INVEST)
	if err := q.Find(&items); err != nil {
		return err
	}
	for i := range items {
		if err := items[i].UpdateInvestStatus(ctx, REFUND_INVEST, uid, NOT_IMMEDIATE); err != nil {
			return errors.Wrap(err, "fail update invest status")
		}
	}
	return nil
}

// ChangeOfInvestStatusByProject 운영 또는 종료시 구좌 상태를 변경하는 함수
func (Share) ChangeOfInvestStatusByProject(ctx context.Context, status PROJECT_STATUS, projectId, uid int64) error {
	var investStatus INVEST_STATUS
	var items []Share
	switch status {
	case OPERATION_FUNDING:
		q := factory.DB(ctx).Table("share").Where("share.project_id = ?", projectId)
		q = q.And("share.invest_status = ?", SUCCESS_INVEST)
		if err := q.Find(&items); err != nil {
			return errors.Wrap(err, "fail to find share by project id")
		}
		investStatus = OPERATION_INVEST
		for i := range items {
			if err := items[i].UpdateInvestStatus(ctx, investStatus, uid, NOT_IMMEDIATE); err != nil {
				return errors.Wrap(err, "fail update invest status")
			}

		}

	case CLOSE_FUNDING:
		// 償還済みへの付け替えは、ファンドの状態とは別の処理(RefundOperationInvestsByProject)。分配償還ジョブが
		// 償還完了の状態を借りて呼んでいたため分けた(MIR-272)。黙って何もしないで通さず、呼び出し元の直し忘れに
		// 気づけるようにエラーにする。
		return errors.New("use RefundOperationInvestsByProject to refund operating investments")
	case FAIL_FUNDING:
		q := factory.DB(ctx).Table("share").Where("share.project_id = ?", projectId)
		q = q.And("share.invest_status IN (?,?)", SUCCESS_INVEST, REQUEST_INVEST)
		if err := q.Find(&items); err != nil {
			return err
		}
		err := UserDeposit{}.FailFundingToRefund(ctx, projectId)
		if err != nil {
			return err
		}
		investStatus = FAIL_INVEST
		for i := range items {
			if err := items[i].UpdateInvestStatus(ctx, investStatus, uid, NOT_IMMEDIATE); err != nil {
				return errors.Wrap(err, "fail update invest status")
			}
		}
	default:
		return nil
	}

	return nil
}

// Create 지분 생성
func (s *Share) Create(ctx context.Context) (int64, error) {

	ksuid, err := ksuid.NewRandom()
	if err != nil {
		return 0, errors.Wrap(err, "fail to generate ksuid")
	}

	s.Id = ksuid.String()

	return factory.DB(ctx).Insert(s)
}

// GetById 지분 상세 조회
func (Share) GetById(ctx context.Context, id string) (*Share, error) {
	var v Share
	if has, err := factory.DB(ctx).Table("share").
		ID(id).Get(&v); err != nil {
		return nil, err
	} else if !has {
		return nil, nil
	}

	return &v, nil
}

// GetByProjectId 상품 지분 목록 조회
func (Share) GetByProjectId(ctx context.Context, projectId int64) (int64, []Share, error) {
	q := factory.DB(ctx)

	var items []Share
	totalCount, err := q.Where("project_id = ?", projectId).FindAndCount(&items)
	if err != nil {
		return 0, nil, errors.Wrap(err, "fail to find items")
	}
	return totalCount, items, nil
}

// 투자자별 취소 되지 않은 전체 투자금액 조회
func (Share) GetAmountByUserId(ctx context.Context, pid, id int64) (int64, error) {
	q := factory.DB(ctx)
	amount, err := q.Where("project_id = ? AND invest_status != 'CANCEL_INVEST' AND owner_id = ? ", pid, id).Sum(&Share{}, "amount")
	if err != nil {
		return 0, errors.Wrap(err, "fail to get amount by user")
	}
	return int64(amount), nil
}

// 지분 업데이트
func (s *Share) Update(ctx context.Context) (err error) {
	_, err = factory.DB(ctx).ID(s.Id).Update(s)
	return err
}

// 투자자 통계 조회
func (UserShareAmountInfo) GetUserProjectShareAmount(ctx context.Context, userId int64) (*UserShareAmountInfo, error) {
	q := factory.DB(ctx)

	template := "SELECT " +
		"(SELECT IFNULL(SUM(s.amount),0) FROM share s  WHERE  s.invest_status =  \"OPERATION_INVEST\" AND  s.owner_id = %v) AS invest_operation_amount ," +
		"(SELECT IFNULL(SUM(s.amount),0) FROM share s  WHERE  s.invest_status =  \"SUCCESS_INVEST\"  AND s.owner_id = %v) AS invest_success_amount ," +
		"(SELECT COUNT(s.id) FROM share s  WHERE  s.invest_status =  \"OPERATION_INVEST\" AND s.owner_id = %v) AS invest_operation_count ," +
		"(SELECT COUNT(s.id) FROM share s  WHERE  s.invest_status =  \"SUCCESS_INVEST\" AND s.owner_id = %v) AS invest_success_count ," +
		"(SELECT COUNT(s.id) FROM share s  WHERE  s.invest_status =  \"REQUEST_INVEST\" AND s.owner_id = %v) AS invest_request_count" +
		" FROM (SELECT * FROM share WHERE share.owner_id = %v) g;"

	sql := fmt.Sprintf(template, userId, userId, userId, userId, userId, userId)
	var userShareAmountInfo = new(UserShareAmountInfo)
	_, err := q.SQL(sql).Get(userShareAmountInfo)

	if err != nil {
		return nil, errors.Wrap(err, "fail to get user project share amount ")
	}

	return userShareAmountInfo, nil
}

func (UserShareAmountInfo) SumDepositAmountOfRequestInvest(ctx context.Context, userId int64) (int64, error) {
	var amountInfo UserShareAmountInfo

	query := factory.DB(ctx).Table("share").
		Join("INNER", "deposit_request", "share.id = deposit_request.share_id").
		Select("IFNULL(SUM(deposit_request.deposit_amount), 0) as invest_request_amount").
		Where("share.owner_id = ? AND share.invest_status IN (?, ?)", userId, REQUEST_INVEST, SUCCESS_INVEST)

	_, err := query.Get(&amountInfo)
	if err != nil {
		return 0, errors.Wrap(err, "fail to get REQUEST_INVEST deposit amount")
	}

	return amountInfo.InvestRequestAmount, nil
}
