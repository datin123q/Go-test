package domain

import (
	"context"

	"github.com/illuminarean/energy-shares-backend-repo/domain"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/jwthelper"
	"github.com/labstack/echo/v4"
	"github.com/pkg/errors"

	"github.com/illuminarean/energy-shares-backend-repo/pkg/jsonutils"
)

type LotteryUsecase interface {
	Application(ctx context.Context, request LotteryApplicationRequest) (LotteryApplicationResponse, error)
	CheckApplication(ctx context.Context, request LotteryCheckApplicationRequest) (LotteryCheckApplicationResponse, error)
	Cancel(ctx context.Context, request LotteryCancelRequest) (LotteryCancelResponse, error)
	GetApplications(ctx context.Context, request LotteryGetApplicationsRequest) (LotteryGetApplicationsResponse, error)
}

type LotteryApplicationRequest struct {
	JWTUser   jwthelper.User
	ProjectId int64 `json:"project_id" description:"プロジェクトID" required:"true" validate:"required" example:"1"`
	UserID    int64 `json:"user_id" swaggerignore:"true" validate:"required" example:"1"`
	Qty       int64 `json:"qty" description:"申込数量" required:"true" validate:"required,gt=0" example:"100"`
}

type LotteryCheckApplicationRequest struct {
	JWTUser   jwthelper.User
	ProjectId int64 `json:"project_id" description:"プロジェクトID" required:"true" validate:"required" example:"1"`
	UserID    int64 `json:"user_id" swaggerignore:"true" validate:"required" example:"1"`
}

type LotteryGetApplicationsRequest struct {
	JWTUser jwthelper.User
	UserID  int64 `param:"user_id" path:"user_id" required:"true" validate:"required,gt=0" description:"ユーザーID" example:"1"`
	Limit   int   `query:"limit" description:"取得件数" validate:"omitempty,gte=0,lte=100" example:"10"`
	Offset  int   `query:"offset" description:"取得開始位置" validate:"omitempty,gte=0" example:"0"`
}

func (u LotteryApplicationRequest) ValidUser(c echo.Context) error {
	if u.JWTUser.Id != u.UserID {
		return errors.New("不正なリクエストです")
	}

	return nil
}

func (u LotteryCheckApplicationRequest) ValidUser(c echo.Context) error {
	if u.JWTUser.Id != u.UserID {
		return errors.New("不正なリクエストです")
	}

	return nil
}

func (u LotteryGetApplicationsRequest) ValidUser(c echo.Context) error {
	if u.JWTUser.Id != u.UserID {
		return errors.New("不正なリクエストです")
	}

	return nil
}

type LotteryApplicationResponse struct {
	ProjectName         string             `json:"project_name" swagger:"desc(プロジェクト名)"`
	BankName            string             `json:"bank_name" swagger:"desc(銀行名)"`
	BranchName          string             `json:"branch_name" swagger:"desc(支店名)"`
	BankAccountNumber   string             `json:"bank_account_number" swagger:"desc(口座番号)"`
	Depositor           string             `json:"depositor" swagger:"desc(預金者)"`
	Qty                 int64              `json:"qty" swagger:"desc(投資数量)"`
	Price               int64              `json:"price" swagger:"desc(投資単価)"`
	Amount              int64              `json:"amount" swagger:"desc(投資金額)"`
	PlanedSelectionDate jsonutils.JSONTime `json:"planed_selection_date" swagger:"desc(抽選予定日)"`
}

type LotteryCheckApplicationResponse struct {
	HasApplied bool `json:"has_applied" swagger:"desc(既に応募しているかどうか)"`
}

type LotteryCancelRequest struct {
	JWTUser              jwthelper.User
	LotteryApplicationID string `validate:"required" description:"抽選申込ID"`
}

type LotteryCancelResponse struct{}

type LotteryGetApplicationsItem struct {
	Id                string                `json:"id" description:"抽選申込ID" example:"ddeb27fb-d9a0-4624-be4d-4615062daed4"`
	ProjectId         int64                 `json:"project_id" description:"ファンドID" example:"501"`
	ProjectName       string                `json:"project_name" description:"ファンド名" example:"九州太陽光ファンド第1号"`
	ApplicationDate   jsonutils.JSONTime    `json:"application_date" description:"申込日時"`
	LotteryStatus     domain.LOTTERY_STATUS `json:"lottery_status" description:"抽選状態" enums:"APPLIED,WINNING,PARTIAL_WINNING,LOST,CANCELED"`
	ApplicationAmount int64                 `json:"application_amount" description:"申込金額" example:"300000"`
	WinningAmount     int64                 `json:"winning_amount" description:"当選金額" example:"0"`
	Cancelable        bool                  `json:"cancelable" description:"キャンセル可能かどうか" example:"true"`
}

type LotteryGetApplicationsResponse struct {
	Items      []LotteryGetApplicationsItem `json:"items" description:"抽選申込一覧"`
	TotalCount int64                        `json:"total_count" description:"総件数" example:"24"`
	Limit      int                          `json:"limit" description:"取得件数" example:"10"`
	Offset     int                          `json:"offset" description:"取得開始位置" example:"0"`
}

