package mysql

import (
	"context"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/stretchr/testify/assert"

	cdomain "github.com/illuminarean/energy-shares-backend-repo/clean-arch-domain"
	"github.com/illuminarean/energy-shares-backend-repo/domain"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/jsonutils"
)

func Test_loginHistoryRepository_GetAll(t *testing.T) {
	type args struct {
		ctx     context.Context
		options []cdomain.LoginHistoryQueryOption
	}
	tests := []struct {
		name    string
		args    args
		want    int64
		want1   []domain.LoginHistory
		wantErr bool
	}{
		{
			name: "get all by default value",
			args: args{
				ctx: context.Background(),
				options: []cdomain.LoginHistoryQueryOption{
					cdomain.LoginHistoryQueryOptionFunc().WithOffset(0),
					cdomain.LoginHistoryQueryOptionFunc().WithLimit(3),
				},
			},
			want: 20,
			want1: []domain.LoginHistory{
				{
					Id:          1,
					UserId:      6,
					Name:        "パク・ジュンヨン",
					Email:       "pjy1418@gmail.com",
					IpAddr:      "59.18.200.132",
					Agent:       "Mozilla/5.0 (Linux; Android 6.0; Nexus 5 Build/MRA58N) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.114 Mobile Safari/537.36",
					LoginResult: domain.SUCCESS_LOGIN,
					LoginTime:   jsonutils.JSONTime(loginTime1),
					Created:     jsonutils.JSONTime(createTime1),
				},
				{
					Id:          2,
					UserId:      110,
					Name:        "やまだ タロ",
					Email:       "private@test.com",
					IpAddr:      "110.10.64.236",
					Agent:       "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.114 Safari/537.36",
					LoginResult: domain.SUCCESS_LOGIN,
					LoginTime:   jsonutils.JSONTime(loginTime2),
					Created:     jsonutils.JSONTime(createTime2),
				},
				{
					Id:          3,
					UserId:      1,
					Name:        "ADMIN",
					Email:       "admin@energyshares.jp",
					IpAddr:      "10.8.1.30",
					Agent:       "Mozilla/5.0 (Linux; Android 6.0; Nexus 5 Build/MRA58N) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.114 Mobile Safari/537.36",
					LoginResult: domain.SUCCESS_LOGIN,
					LoginTime:   jsonutils.JSONTime(loginTime3),
					Created:     jsonutils.JSONTime(createTime3),
				},
			},
			wantErr: false,
		},
		{
			name: "desc sort by id",
			args: args{
				ctx: context.Background(),
				options: []cdomain.LoginHistoryQueryOption{
					cdomain.LoginHistoryQueryOptionFunc().WithSortBy("id"),
					cdomain.LoginHistoryQueryOptionFunc().WithOrder("desc"),
					cdomain.LoginHistoryQueryOptionFunc().WithOffset(0),
					cdomain.LoginHistoryQueryOptionFunc().WithLimit(3),
				},
			},
			want: 20,
			want1: []domain.LoginHistory{
				{
					Id:          20,
					UserId:      110,
					Name:        "やまだ タロ",
					Email:       "private@test.com",
					IpAddr:      "127.0.0.1",
					Agent:       "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.106 Safari/537.36",
					LoginResult: domain.SUCCESS_LOGIN,
					LoginTime:   jsonutils.JSONTime(loginTime20),
					Created:     jsonutils.JSONTime(createTime20),
				},
				{
					Id:          19,
					UserId:      110,
					Name:        "やまだ タロ",
					Email:       "private@test.com",
					IpAddr:      "127.0.0.1",
					Agent:       "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.106 Safari/537.36",
					LoginResult: domain.SUCCESS_LOGIN,
					LoginTime:   jsonutils.JSONTime(loginTime19),
					Created:     jsonutils.JSONTime(createTime19),
				},
				{
					Id:          18,
					UserId:      110,
					Name:        "やまだ タロ",
					Email:       "private@test.com",
					IpAddr:      "127.0.0.1",
					Agent:       "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.106 Safari/537.36",
					LoginResult: domain.SUCCESS_LOGIN,
					LoginTime:   jsonutils.JSONTime(loginTime18),
					Created:     jsonutils.JSONTime(createTime18),
				},
			},
			wantErr: false,
		},
		// {
		// 	name: "desc sort by id and find between the start and end dates",
		// 	args: args{
		// 		ctx: context.Background(),
		// 		options: []cdomain.LoginHistoryQueryOption{
		// 			cdomain.LoginHistoryQueryOptionFunc().WithSortBy("id"),
		// 			cdomain.LoginHistoryQueryOptionFunc().WithOrder("desc"),
		// 			cdomain.LoginHistoryQueryOptionFunc().WithOffset(0),
		// 			cdomain.LoginHistoryQueryOptionFunc().WithStartDate("2021-06-23 20:00:00"),
		// 			cdomain.LoginHistoryQueryOptionFunc().WithEndDate("2021-06-23 21:00:00"),
		// 		},
		// 	},
		// 	want: 5,
		// 	want1: []domain.LoginHistory{
		// 		{
		// 			Id:          15,
		// 			UserId:      110,
		// 			Name:        "やまだ タロ",
		// 			Email:       "private@test.com",
		// 			IpAddr:      "127.0.0.1",
		// 			UserAgent:       "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.106 Safari/537.36",
		// 			LoginResult: domain.SUCCESS_LOGIN,
		// 			LoginTime:   jsonutils.JSONTime(loginTime15),
		// 			Created:     jsonutils.JSONTime(createTime15),
		// 		},
		// 		{
		// 			Id:          14,
		// 			UserId:      110,
		// 			Name:        "やまだ タロ",
		// 			Email:       "private@test.com",
		// 			IpAddr:      "127.0.0.1",
		// 			UserAgent:       "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.106 Safari/537.36",
		// 			LoginResult: domain.SUCCESS_LOGIN,
		// 			LoginTime:   jsonutils.JSONTime(loginTime14),
		// 			Created:     jsonutils.JSONTime(createTime14),
		// 		},
		// 		{
		// 			Id:          13,
		// 			UserId:      108,
		// 			Name:        "ホソン 鄭",
		// 			Email:       "inbq1190@naver.com",
		// 			IpAddr:      "127.0.0.1",
		// 			UserAgent:       "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.106 Safari/537.36",
		// 			LoginResult: domain.SUCCESS_LOGIN,
		// 			LoginTime:   jsonutils.JSONTime(loginTime13),
		// 			Created:     jsonutils.JSONTime(createTime13),
		// 		},
		// 		{
		// 			Id:          12,
		// 			UserId:      108,
		// 			Name:        "ホソン 鄭",
		// 			Email:       "inbq1190@naver.com",
		// 			IpAddr:      "127.0.0.1",
		// 			UserAgent:       "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.106 Safari/537.36",
		// 			LoginResult: domain.SUCCESS_LOGIN,
		// 			LoginTime:   jsonutils.JSONTime(loginTime12),
		// 			Created:     jsonutils.JSONTime(createTime12),
		// 		},
		// 		{
		// 			Id:          11,
		// 			UserId:      108,
		// 			Name:        "ホソン 鄭",
		// 			Email:       "inbq1190@naver.com",
		// 			IpAddr:      "127.0.0.1",
		// 			UserAgent:       "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.106 Safari/537.36",
		// 			LoginResult: domain.SUCCESS_LOGIN,
		// 			LoginTime:   jsonutils.JSONTime(loginTime11),
		// 			Created:     jsonutils.JSONTime(createTime11),
		// 		},
		// 	},
		// 	wantErr: false,
		// },
		{
			name: "find contain the name",
			args: args{
				ctx: context.Background(),
				options: []cdomain.LoginHistoryQueryOption{
					cdomain.LoginHistoryQueryOptionFunc().WithOffset(0),
					cdomain.LoginHistoryQueryOptionFunc().WithLimit(3),
					cdomain.LoginHistoryQueryOptionFunc().WithName("やまだ"),
				},
			},
			want: 12,
			want1: []domain.LoginHistory{
				{
					Id:          2,
					UserId:      110,
					Name:        "やまだ タロ",
					Email:       "private@test.com",
					IpAddr:      "110.10.64.236",
					Agent:       "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.114 Safari/537.36",
					LoginResult: domain.SUCCESS_LOGIN,
					LoginTime:   jsonutils.JSONTime(loginTime2),
					Created:     jsonutils.JSONTime(createTime2),
				},
				{
					Id:          6,
					UserId:      110,
					Name:        "やまだ タロ",
					Email:       "private@test.com",
					IpAddr:      "10.8.1.42",
					Agent:       "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.114 Safari/537.36",
					LoginResult: domain.SUCCESS_LOGIN,
					LoginTime:   jsonutils.JSONTime(loginTime6),
					Created:     jsonutils.JSONTime(createTime6),
				},
				{
					Id:          7,
					UserId:      110,
					Name:        "やまだ タロ",
					Email:       "private@test.com",
					IpAddr:      "10.8.1.42",
					Agent:       "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.114 Safari/537.36",
					LoginResult: domain.SUCCESS_LOGIN,
					LoginTime:   jsonutils.JSONTime(loginTime7),
					Created:     jsonutils.JSONTime(createTime7),
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := loginHistoryMysqlRepository{
				dbFactory: &dbFactoryMock{},
			}
			ignoreTime := cmpopts.IgnoreTypes(jsonutils.JSONTime{})

			got, got1, err := repo.GetAll(tt.args.ctx, tt.args.options...)
			assert.Equalf(t, tt.wantErr, err != nil, "loginHistoryMysqlRepository.GetAll() error = %v, wantErr %v", err, tt.wantErr)
			assert.Equalf(t, int(tt.want), int(got), "loginHistoryMysqlRepository.GetAll() = %v, want %v", got, tt.want)
			assert.Equalf(t, true, cmp.Equal(tt.want1, got1, ignoreTime), "loginHistoryMysqlRepository.GetAll() = %v, want %v", got1, tt.want1)
		})
	}
}

func Test_loginHistoryRepository_Create(t *testing.T) {
	type args struct {
		ctx          context.Context
		loginHistory domain.LoginHistory
	}
	tests := []struct {
		name    string
		args    args
		want    int64
		wantErr bool
	}{
		{
			name: "create login history",
			args: args{
				ctx: context.Background(),
				loginHistory: domain.LoginHistory{
					UserId:      110,
					Name:        "やまだ タロ",
					Email:       "private@test.com",
					IpAddr:      "10.8.1.42",
					Agent:       "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.114 Safari/537.36",
					LoginResult: domain.SUCCESS_LOGIN,
					LoginTime:   jsonutils.JSONTime(time.Now()),
				},
			},
			want:    21,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := loginHistoryMysqlRepository{
				dbFactory: &dbFactoryMock{},
			}

			got, err := repo.Create(tt.args.ctx, tt.args.loginHistory)
			assert.Equalf(t, tt.wantErr, err != nil, "loginHistoryMysqlRepository.Create() error = %v, wantErr %v", err, tt.wantErr)
			assert.Equalf(t, int(tt.want), int(got), "loginHistoryMysqlRepository.Create() = %v, want %v", got, tt.want)
		})
	}
}
