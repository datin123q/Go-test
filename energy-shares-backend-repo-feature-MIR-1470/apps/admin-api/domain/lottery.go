package domain

import (
	"context"
	"fmt"
	"mime/multipart"
	"strings"

	"github.com/pkg/errors"

	cdomain "github.com/illuminarean/energy-shares-backend-repo/clean-arch-domain"
	"github.com/illuminarean/energy-shares-backend-repo/domain"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/jsonutils"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/jwthelper"
)

const (
	LotteryCSVUploadMaxFileSizeBytes    int64 = 10 * 1024 * 1024
	LotteryCSVUploadMaxRequestBodyBytes int64 = LotteryCSVUploadMaxFileSizeBytes + 1*1024*1024
	LotteryCSVUploadMaxRows             int   = 10000
)

type LotteryUseCase interface {
	GetAll(ctx context.Context, request LotteryGetAllRequest) (LotteryGetAllResponse, error)
	Win(ctx context.Context, request LotteryWinRequest) error
	Lose(ctx context.Context, request LotteryLoseRequest) error
	Cancel(ctx context.Context, request LotteryCancelRequest) (LotteryCancelResponse, error)
	Download(ctx context.Context, request LotteryDownloadRequest) (LotteryDownloadResponse, error)
	DownloadForBulkLottery(ctx context.Context, request LotteryDownloadForBulkLotteryRequest) (LotteryDownloadResponse, error)
	UploadCSV(ctx context.Context, request LotteryCSVUploadRequest) (LotteryCSVUploadResponse, error)
	ProcessCSVUploadJob(ctx context.Context, jobID string) error
}

type LotteryGetAllRequest struct {
	JWTUser jwthelper.User
	LotteryGetAllQueryParam
}

type LotteryGetAllResponse struct {
	Items       []LotteryItem `json:"items" description:"抽選申込一覧"`
	TotalCount  int64         `json:"total_count" description:"総数" format:"int64" example:"1"`
	TotalAmount int64         `json:"total_amount" description:"総申込金額" format:"int64" example:"10000"`
}

type LotteryCancelRequest struct {
	JWTUser              jwthelper.User
	LotteryApplicationID string `validate:"required" description:"抽選申込ID"`
}

type LotteryCancelResponse struct{}

type LotteryCSVUploadRequest struct {
	JWTUser   jwthelper.User
	ProjectID int64                 `param:"project_id" path:"project_id" description:"プロジェクトID" format:"int64" validate:"required,gt=0"`
	File      *multipart.FileHeader `swaggerignore:"true"`
}

type LotteryCSVUploadResponse struct {
	JobID  string                               `json:"job_id" description:"CSVアップロードジョブID"`
	Status domain.LOTTERY_CSV_UPLOAD_JOB_STATUS `json:"status" description:"CSVアップロードジョブ状態" example:"QUEUED"`
}

type LotteryCSVValidationErrorDetail struct {
	Row     int    `json:"row"`
	Field   string `json:"field"`
	Message string `json:"message"`
}

type LotteryCSVValidationError struct {
	Details []LotteryCSVValidationErrorDetail `json:"details"`
	err     error
}

func NewLotteryCSVValidationError(details []LotteryCSVValidationErrorDetail) *LotteryCSVValidationError {
	lines := make([]string, 0, len(details))
	for _, detail := range details {
		lines = append(lines, fmt.Sprintf("row %d [%s]: %s", detail.Row, detail.Field, detail.Message))
	}

	return &LotteryCSVValidationError{
		Details: details,
		err:     errors.Wrap(cdomain.ErrBadRequest, strings.Join(lines, "\n")),
	}
}

func (e *LotteryCSVValidationError) Error() string {
	if e == nil || e.err == nil {
		return ""
	}
	return e.err.Error()
}

func (e *LotteryCSVValidationError) Unwrap() error {
	if e == nil {
		return nil
	}
	return e.err
}

type LotteryGetAllQueryParam struct {
	cdomain.SearchInput
	UserId          *int64                 `query:"user_id" description:"ユーザーID" format:"int64"`
	UserName        *string                `query:"user_name" description:"ユーザー名" format:"string"`
	ProjectId       *int64                 `query:"project_id" description:"プロジェクトID" format:"int64"`
	ProjectName     *string                `query:"project_name" description:"プロジェクト名" format:"string"`
	AmountMin       *int64                 `query:"amount_min" description:"抽選申込最小金額" format:"int64"`
	AmountMax       *int64                 `query:"amount_max" description:"抽選申込最大金額" format:"int64"`
	RequestDateFrom *string                `query:"request_start_date" description:"抽選申込日（開始）" validate:"omitempty,datetime=2006-01-02 15:04:05" format:"yyyy-MM-dd HH:mm:ss"`
	RequestDateTo   *string                `query:"request_end_date" description:"抽選申込日（終了）" validate:"omitempty,datetime=2006-01-02 15:04:05" format:"yyyy-MM-dd HH:mm:ss"`
	Status          *domain.LOTTERY_STATUS `query:"status" description:"抽選申込状態" validate:"omitempty,oneof=APPLIED WINNING PARTIAL_WINNING LOST CANCELED" enum:"APPLIED,WINNING,PARTIAL_WINNING,LOST,CANCELED"`
}

type LotteryItem struct {
	Id                       string                `json:"id" description:"抽選申込ID" format:"string"`
	ProjectId                int64                 `json:"project_id" description:"プロジェクトID" format:"int64"`
	UserId                   int64                 `json:"user_id" description:"ユーザーID" format:"int64"`
	UserType                 domain.USER_TYPE      `json:"user_type" description:"会員区分" format:"string" enum:"PRIVATE,CORPORATE"`
	Status                   domain.LOTTERY_STATUS `json:"status" description:"抽選申込状態" example:"APPLIED, WINNING, PARTIAL_WINNING, LOST, CANCELED"`
	Qty                      int64                 `json:"qty" description:"申込数量" format:"int64"`
	UnitPrice                int64                 `json:"unit_price" description:"申込単価" format:"int64"`
	Amount                   int64                 `json:"amount" description:"申込金額" format:"int64"`
	ProjectName              string                `json:"project_name" description:"プロジェクト名"`
	ProjectStatus            domain.PROJECT_STATUS `json:"project_status" description:"プロジェクト状態" example:"DRAFT, OPEN, CLOSE, SUCCESS, FAILURE"`
	UserName                 string                `json:"user_name" description:"ユーザー名"`
	UserEmail                string                `json:"user_email" description:"ユーザーメールアドレス"`
	UserAccountBalance       int64                 `json:"user_account_balance" description:"投資可能金額" format:"int64"`
	LotteryWinProposalMaxQty int64                 `json:"lottery_win_proposal_max_qty" description:"当該申込に対する最大当選口数" format:"int64"`
	LotteryWinProposalMinQty int64                 `json:"lottery_win_proposal_min_qty" description:"当該申込に対する最小当選口数" format:"int64"`
	Winnings                 []LotteryWinning      `json:"winnings,omitempty" description:"当選内容"`
	WinningAmount            int64                 `json:"winning_amount,omitempty" description:"当選総額" format:"int64"`
	CanceledWinningAmount    int64                 `json:"canceled_winning_amount" description:"当選後キャンセル額" format:"int64"`

	AppliedAt  jsonutils.JSONTime `json:"applied_at" description:"申込日時" format:"date-time"`
	WinningAt  jsonutils.JSONTime `json:"winning_at,omitempty" description:"当選日時" format:"date-time"`
	LostAt     jsonutils.JSONTime `json:"lost_at,omitempty" description:"落選日時" format:"date-time"`
	CanceledAt jsonutils.JSONTime `json:"canceled_at,omitempty" description:"取消日時" format:"date-time"`
}

type LotteryWinning struct {
	ShareID     string               `json:"investment_id,omitempty" description:"投資ID" format:"string"`
	ProjectId   int64                `json:"project_id" description:"プロジェクトID" format:"int64"`
	UserId      int64                `json:"user_id" description:"ユーザーID" format:"int64"`
	Qty         int64                `json:"qty" description:"当選数量" format:"int64"`
	UnitPrice   int64                `json:"unit_price" description:"当選単価" format:"int64"`
	Amount      int64                `json:"amount" description:"当選金額" format:"int64"`
	WinningedAt jsonutils.JSONTime   `json:"winninged_at,omitempty" description:"当選日時" format:"date-time"`
	ShareStatus domain.INVEST_STATUS `json:"share_status" description:"当選状態" example:"REQUEST_INVEST,FAIL_INVEST,CANCEL_INVEST,SUCCESS_INVEST,OPERATION_INVEST,REFUND_INVEST"`
}

type LotteryWinRequest struct {
	JWTUser              jwthelper.User
	LotteryApplicationID string  `param:"lottery_application_id" path:"lottery_application_id" description:"抽選申込ID" format:"string" validate:"required"`
	Qty                  int64   `json:"qty" description:"当選口数" format:"int64" validate:"required,gt=0"`
	Comment              *string `json:"comment,omitempty" description:"補足コメント" format:"string"`
}

type LotteryLoseRequest struct {
	JWTUser              jwthelper.User
	LotteryApplicationID string  `param:"lottery_application_id" path:"lottery_application_id" description:"抽選申込ID" format:"string" validate:"required"`
	Comment              *string `json:"comment,omitempty" description:"補足コメント" format:"string"`
}

type LotteryDownloadRequest struct {
	JWTUser  jwthelper.User
	ClientIp string
	LotteryDownloadQueryParam
}

type LotteryDownloadForBulkLotteryRequest struct {
	JWTUser   jwthelper.User
	ClientIp  string
	ProjectId int64 `param:"id" path:"id" description:"プロジェクトID" format:"int64" validate:"required,gt=0"`
	LotteryDownloadForBulkLotteryQueryParam
}

type LotteryDownloadResponse struct {
	Data []byte
}

type LotteryDownloadQueryParam struct {
	Sortby      string                 `query:"sortby" description:"並び替え基準"`
	Order       string                 `query:"order" description:"並び順" enum:"asc,desc"`
	UserId      *int64                 `query:"user_id" description:"ユーザーID" format:"int64"`
	UserName    *string                `query:"user_name" description:"ユーザー名" format:"string"`
	ProjectId   *int64                 `query:"project_id" description:"プロジェクトID" format:"int64"`
	ProjectName *string                `query:"project_name" description:"プロジェクト名" format:"string"`
	AmountMin   *int64                 `query:"amount_min" description:"抽選申込最小金額" format:"int64"`
	AmountMax   *int64                 `query:"amount_max" description:"抽選申込最大金額" format:"int64"`
	Status      *domain.LOTTERY_STATUS `query:"status" description:"抽選申込状態" validate:"omitempty,oneof=APPLIED WINNING PARTIAL_WINNING LOST CANCELED" enum:"APPLIED,WINNING,PARTIAL_WINNING,LOST,CANCELED"`
}

type LotteryDownloadForBulkLotteryQueryParam struct {
	Sortby string `query:"sortby" description:"並び替え基準"`
	Order  string `query:"order" description:"並び順" enum:"asc,desc"`
}
