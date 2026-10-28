package controllers

import (
	"fmt"
	"strconv"
	"time"

	"github.com/pkg/errors"

	"github.com/illuminarean/energy-shares-backend-repo/domain"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/jsonutils"
)

type ProjectSearchInput struct {
	SearchInput
	ProjectSearchItem
	StartDate string `query:"start_date" swagger:"desc(開始日登録日)"`
	EndDate   string `query:"end_date" swagger:"desc(終了日登録日)"`
}

type ProjectSearchSimpleOutput struct {
	Id   int64  `json:"id" swagger:"desc(プロジェクトID)"`
	Name string `query:"name" swagger:"desc(プロジェクト名)"`
}

func (p ProjectSearchSimpleOutput) New(project *domain.Project) *ProjectSearchSimpleOutput {
	return &ProjectSearchSimpleOutput{
		Id:   project.Id,
		Name: project.Name,
	}
}

func (p ProjectSearchSimpleOutput) NewList(projectList []domain.Project) []*ProjectSearchSimpleOutput {
	var list []*ProjectSearchSimpleOutput
	for _, v := range projectList {
		fmt.Println("v: ", v.Id, v.Name)
		list = append(list, p.New(&v))
	}
	return list
}

type ProjectSearchItem struct {
	ProjectId         *string `query:"id" swagger:"desc(プロジェクトID)"`
	ProjectName       *string `query:"name" swagger:"desc(プロジェクト名)"`
	ProjectStatus     *string `query:"project_status" swagger:"desc(募集状態)"`
	BlockInvest       *bool   `query:"block_invest" swagger:"desc(投資制限),enum(true|false)"`
	TargetShareAmount *string `query:"target_share_amount" swagger:"desc(目標金)"`
	ProjectType       *string `query:"project_type" swagger:"desc(商品タイプ)"`
	CategoryType      *string `query:"category_type" swagger:"desc(発電タイプ)"`
	OutOfDateProject  bool    `query:"out_of_date_project" swagger:"desc(掲載日を超えた商品)"`
	Posted            *bool   `query:"posted"  swagger:"desc(掲載有無),enum(true|false)"`
}

type ProjectInput struct {
	Name          string                        `json:"name" form:"name" swagger:"desc(プロジェクト名)"`
	CategoryType  domain.PROJECT_CATEGORY       `json:"category_type" form:"category_type" validate:"required" swagger:"default(SOLAR),enum(SOLAR|LOGISTICS|FOOD_SERVICE|PRECIOUS_METAL|MEDICAL|RETAIL|REAL_ESTATE|BESS|EVENT|FACTORING|ETC),desc(ファンディングカテゴリ 1:太陽光 2:物流 3:飲食 4:貴金属 5:医療 6:店舗 7:不動産 8:蓄電池 9:イベント 10:ファクタリング 11:その他)"` // ファンディングカテゴリ
	SelectionType domain.PROJECT_SELECTION_TYPE `json:"selection_type" form:"selection_type" swagger:"desc(当選方法),enum(LOTTERY|FIRST_COME_FIRST_SERVED)"`
	Country       domain.PROJECT_COUNTRY_TYPE   `json:"country" form:"country" swagger:"enum(JP|USA) desc(国)" validate:"required"`
	AdminNote     string                        `json:"admin_note" form:"admin_note" swagger:"desc(管理者ノート)"`
	// 未入力を許すため文字列で受ける。*int64 で受けると空文字のバインドに失敗する。
	FundNumber               string                `json:"fund_number" form:"fund_number" swagger:"desc(投資家へ提示するファンド番号。半角数字のみ、未入力可)"`
	ProjectInfoHtml          string                `json:"project_info_html" form:"project_info_html" swagger:"desc(プロジェクト情報／商品内容)"`
	CompanyInfoHtml          string                `json:"company_info_html" form:"company_info_html" swagger:"desc(会社情報／お問い合わせ)"`
	Description              string                `json:"description" form:"description" swagger:"desc(商品説明)"`
	StartAt                  string                `json:"start_at" form:"start_at" swagger:"desc(プロジェクト募集開始日),default(2019-12-12 01:01:01)"` // プロジェクト募集開始日
	EndAt                    string                `json:"end_at" form:"end_at" swagger:"desc(プロジェクト募集終了日),default(2019-12-12 01:01:01)"`     // プロジェクト募集終了日
	PlanedOperationStartDate string                `json:"planed_operation_start_date" form:"planed_operation_start_date" swagger:"desc(予定運用開始日),default(2019-12-12 01:01:01)"`
	PlanedOperationEndDate   string                `json:"planed_operation_end_date" form:"planed_operation_end_date" swagger:"desc(予定運用終了日),default(2019-12-12 01:01:01)"`
	ProjectType              domain.PROJECT_TYPE   `json:"project_type" form:"project_type" validate:"required" swagger:"default(LOAN),enum(LOAN|EQUITY),desc(1:ローン型 2:事業型)"`                                                                                                                                                                                              // Loan / Equity Fund
	ProjectStatus            domain.PROJECT_STATUS `json:"project_status" form:"project_status" validate:"required" swagger:"default(STANDBY_FUNDING),enum(STANDBY_FUNDING|OPEN_FUNDING|END_FUNDING|SUCCESS_FUNDING|FAIL_FUNDING|OPERATION_FUNDING|OPERATION_END_FUNDING|CLOSE_FUNDING|INVALID_FUNDING),desc(募集開始前・募集中・募集終了・成立・不成立・運用中・運用終了・償還完了・無効。CLOSE_FUNDING は償還完了)"` // ファンド状態
	Capacity                 string                `json:"capacity" form:"capacity" swagger:"desc(発電容量・総発電容量)" validate:"required"`                                                                                                                                                                                                                                        // 開発容量（株式・ローンファンド）
	DividendRate             float64               `json:"dividend_rate" form:"dividend_rate" swagger:"desc(配当率)" validate:"required"`                                                                                                                                                                                                                                     // 配当率
	OperationTerm            int                   `json:"operation_term" form:"operation_term" swagger:"desc(運用期間)" validate:"required"`
	CurrentShareAmount       int64                 `json:"current_share_amount" form:"current_share_amount" swagger:"desc(募集済み株式金額)"`   // 募集済み株式金額
	CurrentShareQty          int64                 `json:"current_share_qty" form:"current_share_qty" swagger:"desc(募集済み株式数量)"`         // 募集済み株式数量
	MinShareQty              int64                 `json:"min_share_qty" form:"min_share_qty" swagger:"desc(最小購入数量)"`                   // 最小購入数量
	SharePrice               int64                 `json:"share_price" form:"share_price" swagger:"desc(1口あたりの価格)" validate:"required"` // 1口あたりの価格
	TargetShareQty           int64                 `json:"target_share_qty" form:"target_share_qty" swagger:"desc(総目標株式数)"`             // 目標株式数
	TargetShareAmount        int64                 `json:"target_share_amount" form:"target_share_amount" swagger:"desc(上限募集金額)"`       // 上限募集金額
	TargetMinShareAmount     int64                 `json:"target_min_share_amount" form:"target_min_share_amount" swagger:"desc(目標募集金額)"`
	Posted                   bool                  `json:"posted" form:"posted" swagger:"desc(公開有無)"`
	Spv                      int64                 `json:"spv" form:"spv" swagger:"desc(開発会社 SPV)" validate:"required"`
	BlockInvest              bool                  `json:"block_invest" form:"block_invest" swagger:"desc(True: 制限 / False: 許可（投資制限）)"`
	CautionInfoHtml          string                `json:"caution_info_html" form:"caution_info_html" swagger:"desc(注意事項)"` // 注意事項 HTML
	BankName                 string                `json:"bank_name" form:"bank_name" swagger:"desc(金融機関名)"`
	BranchName               string                `json:"branch_name" form:"branch_name" swagger:"desc(支店名)"`
	BankCode                 string                `json:"bank_code" form:"bank_code" swagger:"desc(銀行コード)"`
	BranchCode               string                `json:"branch_code" form:"branch_code" swagger:"desc(銀行支店コード)"`
	BankAccount              string                `json:"bank_account" form:"bank_account" validate:"omitempty,len=7,number" swagger:"desc(口座番号)"`
	Depositor                string                `json:"depositor" form:"depositor" swagger:"desc(口座名義人)"`
	IsInvestorLimit          bool                  `json:"is_investor_limit" form:"is_investor_limit" swagger:"desc(投資人数制限 on/off)"`
	IsAmountLimit            bool                  `json:"is_amount_limit" form:"is_amount_limit" swagger:"desc(投資金額制限 on/off)"`
	IsClosedFund             bool                  `json:"is_closed_fund" form:"is_closed_fund" swagger:"desc(限定公開ファンドフラグ)"`
	ClosedFundSlug           string                `json:"closed_fund_slug" form:"closed_fund_slug" swagger:"desc(限定公開ファンドURLスラッグ)"`
	ProjectProgressPercent   int                   `json:"project_progress_percent" form:"project_progress_percent" vaild:"range(0|100)" swagger:"desc(プロジェクト進捗率)"`
	IsVisibleInvestmentInfo  bool                  `json:"is_visible_investment_info" form:"is_visible_investment_info" swagger:"desc(投資情報公開有無／初期値は非公開)"`
	SubImageOrder            []int64               `json:"sub_image_order" form:"sub_image_order" swagger:"desc(サブ画像の表示順)"`
	DividendType             *domain.DIVIDEND_TYPE `json:"dividend_type" form:"dividend_type" validate:"omitempty,oneof=INCOME_AND_CAPITAL_GAIN INCOME_GAIN CAPITAL_GAIN" swagger:"enum(INCOME_AND_CAPITAL_GAIN|INCOME_GAIN|CAPITAL_GAIN),desc(分配方式 1:インカムゲイン＋キャピタルゲイン 2:インカムゲイン 3:キャピタルゲイン)"`
	CapitalRaisingEntityId   *int64                `json:"capital_raising_entity_id" form:"capital_raising_entity_id" swagger:"desc(資金調達者ID)"`
	// MIR-1492: 画面を開いた時点の detail_version。int64 で受けると、送られなかった場合
	// (列を足す前に開いた画面・画面以外の呼び出し元)と、一度も保存されていないファンドの
	// 0 を区別できないため文字列で受ける(0 は判定する)。
	DetailVersion  string `json:"detail_version" form:"detail_version" swagger:"desc(ファンド詳細を開いた時点の保存回数。送った場合、その後に別の保存があれば409を返す。未指定なら判定しない)"`
	ForceOverwrite bool   `json:"force_overwrite" form:"force_overwrite" swagger:"desc(trueの場合、開いた後に別の保存があっても上書きする)"`
}

type CheckDuplicateFundNumberInput struct {
	FundNumber string `query:"fund_number" swagger:"desc(確認するファンド番号)"`
	ProjectId  int64  `query:"id" swagger:"desc(自身のファンドID。新規登録時は指定しない)"`
}

type CheckDuplicateFundNumberResponseDto struct {
	IsDuplicated bool `json:"is_duplicated" swagger:"desc(他のファンドで使われている場合 true)"`
}

// ExpectedDetailVersion は、この保存で判定に使う「画面を開いた時点の detail_version」を
// 返す(MIR-1492)。loaded が空の呼び出し元(画面以外)と、警告を見た管理者が上書きを
// 選んだ場合(force)は判定しないため nil を返す。
func ExpectedDetailVersion(loaded string, force bool) (*int64, error) {
	if loaded == "" || force {
		return nil, nil
	}

	version, err := strconv.ParseInt(loaded, 10, 64)
	if err != nil {
		return nil, errors.Wrapf(err, "fail to parse detail version: %s", loaded)
	}
	if version < 0 {
		return nil, errors.Errorf("detail version must not be negative: %s", loaded)
	}

	return &version, nil
}

// IsDetailVersionConflicted は、画面を開いた後に別の保存があったかを判定する(MIR-1492)。
// expected は ExpectedDetailVersion の戻り値、current は保存直前に読み直した値。
// 大小は見ず、一致しなければ競合とする(開いた時点の値が今より大きいのは、DB を
// 戻したなど想定外の状態のため、上書きせずに知らせる)。
func IsDetailVersionConflicted(expected *int64, current int64) bool {
	return expected != nil && *expected != current
}

// FundNumberMaxDigits はファンド番号の桁数上限。
// BIGINT の範囲(19桁)には収まるが、運用上の上限として 12 桁で制限する。
const FundNumberMaxDigits = 12

// ParseFundNumber は画面から届いたファンド番号を数値へ変換する。
// 未入力は NULL として保存するため nil を返す。
func ParseFundNumber(value string) (*int64, error) {
	if value == "" {
		return nil, nil
	}

	// strconv.ParseInt は符号や前後の空白を許さないが、"+1" は通してしまう。
	// 半角数字のみを受け付けるため、先に文字種を確かめる。
	for _, r := range value {
		if r < '0' || r > '9' {
			return nil, errors.Errorf("fund number must contain digits only: %s", value)
		}
	}

	if len(value) > FundNumberMaxDigits {
		return nil, errors.Errorf("fund number must be %d digits or less: %s", FundNumberMaxDigits, value)
	}

	fundNumber, err := strconv.ParseInt(value, 10, 64)
	if err != nil {
		return nil, errors.Wrapf(err, "fail to parse fund number: %s", value)
	}

	// 0 は使えない。更新時に fund_number を Nullable 指定しており、
	// XORM は参照先が 0 のポインタもゼロ値として NULL に置き換えるため、
	// 保存しても未設定として消えてしまう。
	if fundNumber == 0 {
		return nil, errors.Errorf("fund number must not be zero: %s", value)
	}

	return &fundNumber, nil
}

func (p *ProjectInput) ToModel() (*domain.Project, error) {
	fundNumber, err := ParseFundNumber(p.FundNumber)
	if err != nil {
		return nil, err
	}

	if p.StartAt == "" {
		p.StartAt = "2020-01-01 00:00:00"
	}
	startAt, err := jsonutils.ParseJSONTimeJST(time.DateTime, p.StartAt)
	if err != nil {
		return nil, err
	}

	if p.EndAt == "" {
		p.EndAt = p.StartAt
	}
	endAt, err := jsonutils.ParseJSONTimeJST(time.DateTime, p.EndAt)
	if err != nil {
		return nil, err
	}

	var planedOperationStartDate jsonutils.JSONTime
	if p.PlanedOperationStartDate != "" {
		planedOperationStartDate, err = jsonutils.ParseJSONTimeJST(time.DateTime, p.PlanedOperationStartDate)
		if err != nil {
			return nil, err
		}
	}

	var planedOperationEndDate jsonutils.JSONTime
	if p.PlanedOperationEndDate != "" {
		planedOperationEndDate, err = jsonutils.ParseJSONTimeJST(time.DateTime, p.PlanedOperationEndDate)
		if err != nil {
			return nil, err
		}
	}

	return &domain.Project{
		SpvId:                    p.Spv,
		Name:                     p.Name,
		CategoryType:             p.CategoryType,
		SelectionType:            p.SelectionType,
		AdminNote:                p.AdminNote,
		ProjectInfoHtml:          p.ProjectInfoHtml,
		CompanyInfoHtml:          p.CompanyInfoHtml,
		CautionInfoHtml:          p.CautionInfoHtml,
		Description:              p.Description,
		StartAt:                  startAt,
		EndAt:                    endAt,
		PlanedOperationStartDate: planedOperationStartDate,
		PlanedOperationEndDate:   planedOperationEndDate,

		ProjectType:             p.ProjectType,
		ProjectStatus:           p.ProjectStatus,
		Capacity:                p.Capacity,
		Country:                 p.Country,
		DividendRate:            p.DividendRate,
		OperationTerm:           p.OperationTerm,
		CurrentShareAmount:      p.CurrentShareAmount,
		CurrentShareQty:         p.CurrentShareQty,
		MinShareQty:             p.MinShareQty,
		SharePrice:              p.SharePrice,
		TargetMinShareAmount:    p.TargetMinShareAmount,
		TargetShareQty:          p.TargetShareQty,
		TargetShareAmount:       p.TargetShareAmount,
		Posted:                  p.Posted,
		FundNumber:              fundNumber,
		BlockInvest:             p.BlockInvest,
		BankName:                p.BankName,
		BranchName:              p.BranchName,
		BankCode:                p.BankCode,
		BranchCode:              p.BranchCode,
		BankAccount:             p.BankAccount,
		Depositor:               p.Depositor,
		IsInvestorLimit:         p.IsInvestorLimit,
		IsAmountLimit:           p.IsAmountLimit,
		IsVisibleInvestmentInfo: p.IsVisibleInvestmentInfo,
		IsClosedFund:            p.IsClosedFund,
		ClosedFundSlug:          p.ClosedFundSlug,
		ProjectProgressPercent:  p.ProjectProgressPercent,
		DividendType:            p.DividendType,
		CapitalRaisingEntityId:  p.CapitalRaisingEntityId,
	}, nil
}

type ProjectAdminResponseDto struct {
	Id                       int64                         `json:"id"`
	CommonId                 string                        `json:"common_id" swagger:"desc(共有ID)"`
	FundNumber               *int64                        `json:"fund_number" swagger:"desc(投資家へ提示するファンド番号)"`
	SpvId                    int64                         `json:"spv_id" swagger:"desc(開発会社ID)"`
	Spv                      *domain.Spv                   `json:"spv" swagger:"desc(開発会社SPVDEVELOPER)"`
	Name                     string                        `json:"name" swagger:"desc(商品名)"`
	CategoryType             domain.PROJECT_CATEGORY       `json:"category_type" swagger:"desc(ファンディングカテゴリ) "` // ファンディングカテゴリ
	SelectionType            domain.PROJECT_SELECTION_TYPE `json:"selection_type" swagger:"desc(当選方法)"`
	AdminNote                string                        `json:"admin_note" swagger:"desc(管理者メモ)"`              // テキスト()
	ProjectInfoHtml          string                        `json:"project_info_html"  swagger:"desc(プロジェクト情報)"`   // プロジェクトコンテンツHtml
	CompanyInfoHtml          string                        `json:"company_info_html"  swagger:"desc(問い合わせ,会社情報)"` // 会社コンテンツHtml
	Description              string                        `json:"description"  swagger:"desc(商品説明)"`             // 登録者?担当者?
	StartAt                  jsonutils.JSONTime            `json:"start_at"  swagger:"desc(募集開始日時)" `
	EndAt                    jsonutils.JSONTime            `json:"end_at"  swagger:"desc(募集終了日時)"`
	PlanedOperationStartDate jsonutils.JSONTime            `json:"planed_operation_start_date" swagger:"desc(予定運用開始日)"`
	PlanedOperationEndDate   jsonutils.JSONTime            `json:"planed_operation_end_date" swagger:"desc(予定運用終了日)"`
	ProjectType              domain.PROJECT_TYPE           `json:"project_type"  swagger:"desc(プロジェクトタイプ1:LOAN 2:EQUITY)"` // ローン株式ファンド
	ProjectStatus            domain.PROJECT_STATUS         `json:"project_status"  swagger:"desc(ファンドの状態。STANDBY_FUNDING 募集開始前・OPEN_FUNDING 募集中・END_FUNDING 募集終了・SUCCESS_FUNDING 成立・FAIL_FUNDING 不成立・OPERATION_FUNDING 運用中・OPERATION_END_FUNDING 運用終了・CLOSE_FUNDING 償還完了・INVALID_FUNDING 無効)"`
	OwnerId                  int64                         `json:"owner_id"  swagger:"desc(登録者)"` // 登録者?担当者?
	UpdateId                 int64                         `json:"update_id"  swagger:"desc(最近修正者)"`
	Country                  domain.PROJECT_COUNTRY_TYPE   `json:"country" form:"country" swagger:"desc(국가)" validate:"required"`
	Capacity                 string                        `json:"capacity" swagger:"desc(개발 용량,KW,formP DC 규모)"` // 개발 용량 주식 대출 펀드

	DividendRate  float64 `json:"dividend_rate"  swagger:"desc(배당율)"`   //
	OperationTerm int     `json:"operation_term" swagger:"desc( 운용기간)"` // 운용기간

	CurrentShareAmount   int64 `json:"current_share_amount" swagger:"desc(모집된 Share 금액)"` // 모집된 주식 금액
	CurrentShareQty      int64 `json:"current_share_qty" swagger:"desc(현재 모집된 Share 수량)"` // 모집된 주식 수량
	MinShareQty          int64 `json:"min_share_qty"  swagger:"desc(최소 구매 수량)"`           // 최소 구매 수량
	SharePrice           int64 `json:"share_price"  swagger:"desc(1구좌당 가격)"`              // 1구좌 당 가격
	TargetMinShareAmount int64 `json:"target_min_share_amount" swagger:"desc(목표 모금액)"`    // 목표 모금액
	TargetShareQty       int64 `json:"target_share_qty"  swagger:"desc(목표 주식수 총 발생 주식)"`  // 목표 주식수
	TargetShareAmount    int64 `json:"target_share_amount"  swagger:"desc(상한 모금액)"`       // 총 대출액
	Posted               bool  `json:"posted" swagger:"desc(펀드 게시여부)"`                    // 펀딩 디스플레이 여부 펀딩 시작후 수정 불가
	ViewCount            int64 `json:"view_count" swagger:"desc(퍼블릭 조회 카운트)"`             // (퍼블릭 조회 카운트)

	BlockInvest             bool                   `json:"block_invest" swagger:"desc(투자 제한 True:투자제한됨 False: 투자 허용)"`
	MainImageId             int64                  `json:"main_image_id" swagger:"desc(プロジェクトメイン画像固有ID)"`
	MainImageUrl            string                 `json:"main_image_url"  swagger:"desc(ファンディングメイン画像URL)"`
	MainImage               *domain.ProjectImage   `json:"main_image" swagger:"desc(プロジェクト画像)"`               // プロジェクトメイン画像
	SubImageInfoList        []*domain.SubImageInfo `json:"sub_image_info_list" swagger:"desc(プロジェクトサブ画像リスト)"` // プロジェクトサブ画像リスト
	Attachments             []*domain.Attachment   `json:"attachments"  swagger:"desc(プロジェクト添付ファイルリスト)"`      // 添付ファイル // サブ資料
	News                    []*domain.ProjectNews  `json:"project_news" swagger:"desc(プロジェクトニュースリスト)"`        // 添付ファイル // サブ資料
	ClosedFundSlug          string                 `json:"closed_fund_slug" swagger:"desc(限定公開ファンドURLスラッグ)"`
	IsWaitInvest            bool                   `json:"is_wait_invest"  swagger:"desc(投資待機状態True:投資待機商品False:投資許可)"`
	IsManualInvest          bool                   `json:"is_manual_invest"  swagger:"desc(手動投資有無)"`
	IsInvestorLimit         bool                   `json:"is_investor_limit" swagger:"desc(投資家数制限有無)"`
	IsAmountLimit           bool                   `json:"is_amount_limit" swagger:"desc(投資金額制限有無)"`
	IsVisibleInvestmentInfo bool                   `json:"is_visible_investment_info" swagger:"desc(投資情報公開有無)"`
	IsClosedFund            bool                   `json:"is_closed_fund" swagger:"desc(限定公開ファンドフラグ)"`
	Created                 jsonutils.JSONTime     `json:"created"`
	Updated                 jsonutils.JSONTime     `json:"updated"`

	TransactionCount            int64                            `json:"transaction_count" swagger:"desc(取引カウント)"`
	CautionInfoHtml             string                           `json:"caution_info_html" swagger:"desc(注意事項)"` // 注意事項Html
	BankName                    string                           `json:"bank_name" swagger:"desc(金融機関名)"`
	BranchName                  string                           `json:"branch_name" swagger:"desc(支店名)"`
	BankCode                    string                           `json:"bank_code" swagger:"desc(銀行コード)"`
	BranchCode                  string                           `json:"branch_code" swagger:"desc(銀行支店コード)"`
	BankAccount                 string                           `json:"bank_account" swagger:"desc(口座番号)"`
	Depositor                   string                           `json:"depositor" swagger:"desc(口座名義)"`
	ContractBeforePage          *domain.ProjectFile              `json:"contract_before_page" swagger:"desc(계약성립 전 서면)"`
	ContractAfterPage           *domain.ProjectFile              `json:"contract_after_page" swagger:"desc(계약성립 시 서면)"`
	SilentPartnerships          *domain.ProjectFile              `json:"silent_partnerships" swagger:"desc(익명조합 계약서)"`
	ProjectReport               *domain.ProjectFile              `json:"project_report" swagger:"desc(펀드 보고서)"`
	ContractBeforePageDate      jsonutils.JSONTime               `json:"contract_before_page_date"  swagger:"desc(契約成立前書面登録日)"`
	ContractAfterPageDate       jsonutils.JSONTime               `json:"contract_after_page_date"  swagger:"desc(契約成立時書面登録日)"`
	SilentPartnershipsDate      jsonutils.JSONTime               `json:"silent_partnerships_date"  swagger:"desc(匿名組合契約書登録日)"`
	ProjectReportDate           jsonutils.JSONTime               `json:"project_report_date"  swagger:"desc(ファンドレポート登録日)"`
	TransactionDate             jsonutils.JSONTime               `json:"transaction_date" swagger:"desc(取引日)"`
	TransactionAmount           int                              `json:"transaction_amount" swagger:"desc(取引募集額)"`
	PublicProgressInvestPercent float64                          `json:"public_progress_invest_percent" swagger:"desc(ユーザープログレスバー用募集額%-投資キャンセル含む)"`
	RemainingQty                int64                            `json:"remaining_qty" swagger:"desc(対象プロジェクト購入可能口数)"`
	CurrentInvestorCount        int64                            `json:"current_investor_count" swagger:"desc(対象プロジェクト投資家人数)"`
	RaisedAmount                int64                            `json:"raised_amount"  swagger:"desc(募集調達金額)"`
	LoanExecute                 *domain.ProjectInvestmentExecute `json:"project_investment_execute" swagger:"desc(ローン実行)"`
	InvestWaitRemainingQty      int64                            `json:"invest_wait_remaining_qty" swagger:"desc(対象プロジェクト投資待機可能口数)"`
	PublicViewShareAmount       int64                            `json:"public_view_share_amount" swagger:"desc(商品ステータス別Share金額)"`
	PublicViewEndAt             jsonutils.JSONTime               `json:"public_view_end_at" swagger:"desc(사용자 페이지 모집마감일 표시용)"`
	ProjectProgressPercent      int                              `json:"project_progress_percent" swagger:"desc(プロジェクト進捗率)"`
	ProjectMilestones           []domain.ProjectMilestone        `json:"project_milestones"  swagger:"desc(プロジェクトマイルストーンリスト)"`
	DividendType                *domain.DIVIDEND_TYPE            `json:"dividend_type" swagger:"desc(分配方式)"`
	CapitalRaisingEntity        *domain.CapitalRaisingEntity     `json:"capital_raising_entity" swagger:"desc(資金調達者)"`
	DetailVersion               int64                            `json:"detail_version" swagger:"desc(ファンド詳細の保存回数。保存時にそのまま送り返す)"`
}

func (receiver ProjectAdminResponseDto) New(p *domain.Project) *ProjectAdminResponseDto {
	return &ProjectAdminResponseDto{
		Id:                          p.Id,
		CommonId:                    p.CommonId,
		FundNumber:                  p.FundNumber,
		SpvId:                       p.SpvId,
		Spv:                         p.Spv,
		Name:                        p.Name,
		CategoryType:                p.CategoryType,
		SelectionType:               p.SelectionType,
		AdminNote:                   p.AdminNote,
		ProjectInfoHtml:             p.ProjectInfoHtml,
		CompanyInfoHtml:             p.CompanyInfoHtml,
		Description:                 p.Description,
		StartAt:                     p.StartAt,
		EndAt:                       p.EndAt,
		PlanedOperationStartDate:    p.PlanedOperationStartDate,
		PlanedOperationEndDate:      p.PlanedOperationEndDate,
		ProjectType:                 p.ProjectType,
		ProjectStatus:               p.ProjectStatus,
		OwnerId:                     p.OwnerId,
		UpdateId:                    p.UpdateId,
		Capacity:                    p.Capacity,
		Country:                     p.Country,
		DividendRate:                p.DividendRate,
		OperationTerm:               p.OperationTerm,
		CurrentShareAmount:          p.CurrentShareAmount,
		CurrentShareQty:             p.CurrentShareQty,
		MinShareQty:                 p.MinShareQty,
		SharePrice:                  p.SharePrice,
		TargetMinShareAmount:        p.TargetMinShareAmount,
		TargetShareQty:              p.TargetShareQty,
		TargetShareAmount:           p.TargetShareAmount,
		Posted:                      p.Posted,
		ViewCount:                   p.ViewCount,
		BlockInvest:                 p.BlockInvest,
		MainImageId:                 p.MainImageId,
		MainImageUrl:                p.MainImageUrl,
		MainImage:                   p.MainImage,
		SubImageInfoList:            p.SubImageInfoList,
		Attachments:                 p.Attachments,
		News:                        p.News,
		ClosedFundSlug:              p.ClosedFundSlug,
		IsWaitInvest:                p.IsWaitInvest,
		IsManualInvest:              p.IsManualInvest,
		IsInvestorLimit:             p.IsInvestorLimit,
		IsAmountLimit:               p.IsAmountLimit,
		IsVisibleInvestmentInfo:     p.IsVisibleInvestmentInfo,
		IsClosedFund:                p.IsClosedFund,
		Created:                     p.Created,
		Updated:                     p.Updated,
		DetailVersion:               p.DetailVersion,
		TransactionCount:            p.TransactionCount,
		CautionInfoHtml:             p.CautionInfoHtml,
		BankName:                    p.BankName,
		BranchName:                  p.BranchName,
		BankCode:                    p.BankCode,
		BranchCode:                  p.BranchCode,
		BankAccount:                 p.BankAccount,
		Depositor:                   p.Depositor,
		ContractBeforePage:          p.ContractBeforePage,
		ContractAfterPage:           p.ContractAfterPage,
		SilentPartnerships:          p.SilentPartnerships,
		ProjectReport:               p.ProjectReport,
		ContractBeforePageDate:      p.ContractBeforePageDate,
		ContractAfterPageDate:       p.ContractAfterPageDate,
		SilentPartnershipsDate:      p.SilentPartnershipsDate,
		ProjectReportDate:           p.ProjectReportDate,
		TransactionDate:             p.TransactionDate,
		TransactionAmount:           p.TransactionAmount,
		PublicProgressInvestPercent: p.PublicProgressInvestPercent,
		RemainingQty:                p.RemainingQty,
		CurrentInvestorCount:        p.CurrentInvestorCount,
		RaisedAmount:                p.RaisedAmount,
		LoanExecute:                 p.ProjectInvestmentExecute,
		InvestWaitRemainingQty:      p.InvestWaitRemainingQty,
		PublicViewShareAmount:       p.PublicViewShareAmount,
		PublicViewEndAt:             p.EndAt,
		ProjectProgressPercent:      p.ProjectProgressPercent,
		ProjectMilestones:           p.ProjectMilestones,
		DividendType:                p.DividendType,
		CapitalRaisingEntity:        p.CapitalRaisingEntity,
	}
}

type ProjectMilestoneInput struct {
	MilestoneStatus     string `json:"milestone_status" form:"milestone_status" validate:"oneof=MILESTONE_READY MILESTONE_GOING MILESTONE_COMPLETED"  swagger:"desc(マイルストーン状態)"`
	MilestoneInfo       string `json:"milestone_info" form:"milestone_info" validate:"min=0,max=30" swagger:"desc(マイルストーン内容)"`
	MilestoneAssignment int    `json:"milestone_assignment" form:"milestone_assignment" validate:"numeric,min=0,max=100" swagger:"desc(割当進捗度)"`
	MilestoneNote       string `json:"milestone_note" form:"milestone_note" validate:"min=0,max=100" swagger:"desc(マイルストーンノート)"`
	MilestoneNo         int    `json:"milestone_no" form:"milestone_no" validate:"numeric" swagger:"desc(進行状況順序)"`
}

func (p *ProjectMilestoneInput) ToModel() (*domain.ProjectMilestone, error) {
	return &domain.ProjectMilestone{
		MilestoneStatus:     domain.MILESTONE_STATUS(p.MilestoneStatus),
		MilestoneNo:         p.MilestoneNo,
		MilestoneInfo:       p.MilestoneInfo,
		MilestoneAssignment: p.MilestoneAssignment,
		MilestoneNote:       p.MilestoneNote,
	}, nil
}
