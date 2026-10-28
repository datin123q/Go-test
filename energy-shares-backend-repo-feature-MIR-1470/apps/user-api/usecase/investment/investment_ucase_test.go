package investment

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/illuminarean/energy-shares-backend-repo/clean-arch-domain/mocks"
	"github.com/illuminarean/energy-shares-backend-repo/domain"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/jsonutils"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/jwthelper"
	udomain "github.com/illuminarean/energyshares-user-api/domain"
)

var investmentTestJST = func() *time.Location {
	loc, err := time.LoadLocation("Asia/Tokyo")
	if err != nil {
		panic(err)
	}
	return loc
}()

func Test_investmentUsecase_GetResult(t *testing.T) {
	type args struct {
		ctx     context.Context
		request udomain.InvestmentGetResultRequest
	}
	tests := []struct {
		name                         string
		args                         args
		shareRepository              func() *mocks.ShareRepository
		projectRepository            func() *mocks.ProjectRepository
		userVirtualAccountRepository func() *mocks.UserVirtualAccountRepository
		want                         udomain.InvestmentGetResultResponse
		wantErr                      bool
	}{
		{
			name: "投資結果照会成功",
			args: args{
				ctx: context.Background(),
				request: udomain.InvestmentGetResultRequest{
					InvestmentId: "INVESTMENT_ID1",
					JWTUser: jwthelper.User{
						Id: 1000,
					},
				},
			},
			shareRepository: func() *mocks.ShareRepository {
				repo := &mocks.ShareRepository{}
				repo.On("GetById", context.Background(), "INVESTMENT_ID1").
					Return(domain.Share{
						Id:         "INVESTMENT_ID1",
						OwnerId:    1000,
						ProjectId:  2000,
						Qty:        5,
						Price:      1000,
						InvestDate: jsonutils.JSONTime(time.Date(2024, 1, 1, 0, 0, 0, 0, investmentTestJST)),
					}, nil)

				return repo
			},

			projectRepository: func() *mocks.ProjectRepository {
				repo := &mocks.ProjectRepository{}
				repo.On("GetById", context.Background(), int64(2000)).
					Return(domain.Project{
						Id:          2000,
						Name:        "太陽光プロジェクト",
						ProjectType: domain.EQUITY,
					}, nil)
				return repo
			},

			userVirtualAccountRepository: func() *mocks.UserVirtualAccountRepository {
				repo := &mocks.UserVirtualAccountRepository{}
				repo.On("GetByUserIdAndLinkedStatus", context.Background(), int64(1000), domain.EQUITY, domain.ON_VA_ACCOUNT).
					Return(domain.UserVirtualAccount{
						UserId:           2000,
						VaBankName:       "ＧМＯあおぞらネット銀行",
						VaAccountNumber:  "5194080",
						VaHolderNameKana: "カ）ハコブネカ）ハコブネ",
					}, nil)
				return repo
			},
			want: udomain.InvestmentGetResultResponse{
				InvestmentId:      "INVESTMENT_ID1",
				ProjectName:       "太陽光プロジェクト",
				InvestmentDate:    jsonutils.JSONTime(time.Date(2024, 1, 1, 0, 0, 0, 0, investmentTestJST)),
				BankName:          "ＧМＯあおぞらネット銀行",
				BankAccountNumber: "5194080",
				Depositor:         "カ）ハコブネカ）ハコブネ",
				Qty:               5,
				Price:             1000,
				Amount:            5000,
			},
			wantErr: false,
		},

		{
			name: "別のユーザーの投資情報を照会しようとする場合 => エラー返却",
			args: args{
				ctx: context.Background(),
				request: udomain.InvestmentGetResultRequest{
					InvestmentId: "INVESTMENT_ID1",
					JWTUser: jwthelper.User{
						Id: 1000,
					},
				},
			},
			shareRepository: func() *mocks.ShareRepository {
				repo := &mocks.ShareRepository{}
				repo.On("GetById", context.Background(), "INVESTMENT_ID1").
					Return(domain.Share{
						Id:         "INVESTMENT_ID1",
						OwnerId:    1001,
						ProjectId:  2000,
						Qty:        5,
						Price:      1000,
						InvestDate: jsonutils.JSONTime(time.Date(2024, 1, 1, 0, 0, 0, 0, investmentTestJST)),
					}, nil)

				return repo
			},

			projectRepository: func() *mocks.ProjectRepository {
				repo := &mocks.ProjectRepository{}
				return repo
			},

			userVirtualAccountRepository: func() *mocks.UserVirtualAccountRepository {
				repo := &mocks.UserVirtualAccountRepository{}
				return repo
			},
			want:    udomain.InvestmentGetResultResponse{},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			usecase := investmentUsecase{
				shareRepository:              tt.shareRepository(),
				projectRepository:            tt.projectRepository(),
				userVirtualAccountRepository: tt.userVirtualAccountRepository(),
			}
			got, err := usecase.GetResult(tt.args.ctx, tt.args.request)

			assert.Equalf(t, tt.wantErr, err != nil, "GetResult() error = %v, wantErr %v", err, tt.wantErr)
			assert.Equalf(t, tt.want, got, "GetResult() = %v, want %v", got, tt.want)
		})
	}
}
