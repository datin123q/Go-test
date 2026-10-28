package operator_history

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	adomain "github.com/illuminarean/energy-share-core-api/domain"
	"github.com/illuminarean/energy-shares-backend-repo/clean-arch-domain/mocks"
	"github.com/illuminarean/energy-shares-backend-repo/domain"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/config"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/i18nutils"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/jwthelper"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/pathutils"
)

func Test_operatorHistoryUsecase_GetAll(t *testing.T) {
	type fields struct {
		operatorHistoryKsqlRepository func() *mocks.OperatorHistoryKsqlRepository
		cfg                           *config.Config
		translator                    i18nutils.Translator
	}
	type args struct {
		ctx     context.Context
		request adomain.OperatorHistoryGetAllRequest
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    adomain.OperatorHistoryGetAllResponse
		wantErr bool
	}{
		{
			name: "管理者ログ照会(GetAll) => 正常応答",
			fields: fields{
				operatorHistoryKsqlRepository: func() *mocks.OperatorHistoryKsqlRepository {
					repo := new(mocks.OperatorHistoryKsqlRepository)
					repo.On("GetAll", context.Background(),
						mock.AnythingOfType("func(*clean_arch_domain.OperatorHistoryKsqlQueryOptionFields)"),
						mock.AnythingOfType("func(*clean_arch_domain.OperatorHistoryKsqlQueryOptionFields)"),
						mock.AnythingOfType("func(*clean_arch_domain.OperatorHistoryKsqlQueryOptionFields)"),
					).Return(int64(1), []domain.OperatorHistoryRowResponse{operatorHistory1}, nil)
					return repo
				},
				cfg: &config.Config{CommonConfig: config.CommonConfig{
					EventStream: struct{ KsqlDB config.KsqlDBConfig }{KsqlDB: config.KsqlDBConfig{
						Resources: config.KsqlDBResources{
							AccessLog: "dev_admin_access_log_filtered",
						},
					}},
				}},
				translator: i18nutils.New(pathutils.GetPathOfAssets()),
			},
			args: args{
				ctx: context.Background(),
				request: adomain.OperatorHistoryGetAllRequest{
					JWTUser: jwthelper.User{
						Id:    1,
						Name:  "Admin",
						Email: "admin@energyshares.jp",
					},
					ClientIp: "192.168.0.2",
				},
			},
			want: adomain.OperatorHistoryGetAllResponse{
				TotalCount: 1,
				Items:      []domain.OperatorHistoryRowResponse{operatorHistory1},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u := operatorHistoryUsecase{
				operatorHistoryKsqlRepository: tt.fields.operatorHistoryKsqlRepository(),
				cfg:                           tt.fields.cfg,
				translator:                    tt.fields.translator,
			}

			got, err := u.GetAll(tt.args.ctx, tt.args.request)

			assert.NoErrorf(t, err, "operatorHistoryUsecase.GetAll() error = %v, wantErr %v", err, tt.wantErr)
			assert.Equal(t, tt.want, got, "operatorHistoryUsecase.GetAll() got = %v, want %v", got, tt.want)
		})
	}
}

func Test_operatorHistoryUsecase_Download(t *testing.T) {
	type fields struct {
		operatorHistoryKsqlRepository func() *mocks.OperatorHistoryKsqlRepository
		cfg                           *config.Config
		translator                    i18nutils.Translator
	}
	type args struct {
		ctx     context.Context
		request adomain.OperatorHistoryDownloadRequest
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    adomain.OperatorHistoryDownloadResponse
		wantErr bool
	}{
		{
			name: "管理者ログダウンロード(Download)呼び出しに対するbytes返却 => 正常応答",
			fields: fields{
				operatorHistoryKsqlRepository: func() *mocks.OperatorHistoryKsqlRepository {
					repo := new(mocks.OperatorHistoryKsqlRepository)
					repo.On("GetAll", context.Background(),
						mock.AnythingOfType("func(*clean_arch_domain.OperatorHistoryKsqlQueryOptionFields)"),
						mock.AnythingOfType("func(*clean_arch_domain.OperatorHistoryKsqlQueryOptionFields)"),
					).Return(int64(1), []domain.OperatorHistoryRowResponse{operatorHistory1}, nil)
					return repo
				},
				cfg: &config.Config{CommonConfig: config.CommonConfig{
					EventStream: struct{ KsqlDB config.KsqlDBConfig }{KsqlDB: config.KsqlDBConfig{
						Resources: config.KsqlDBResources{
							AccessLog: "dev_admin_access_log_filtered",
						},
					}},
				}},
				translator: i18nutils.New(pathutils.GetPathOfAssets()),
			},
			args: args{
				ctx: context.Background(),
				request: adomain.OperatorHistoryDownloadRequest{
					JWTUser: jwthelper.User{
						Id:    1,
						Name:  "Admin",
						Email: "admin@energyshares.jp",
					},
					ClientIp: "192.168.0.2",
				},
			},
			want: adomain.OperatorHistoryDownloadResponse{
				Data: []byte{
					// 処理日時,運用者名(メールアドレス),ログ種別,ログ,作業内容,IP
					// 2025-01-03 12:13:33,"{Id:1,Name:admin,Email:admin@energyshares.jp}",commoncode.LOG_CONTENT_TYPE.LOGIN,"{LogType:request,RequestID:e72c9dd2-c844-11ef-8b8e-f6f0d9467449,RemoteIP:::1,Host:localhost:8080,Operation:/api/v1/admin-users/login,URI:/api/v1/admin-users/login,Method:POST,Referer:,UserAgent:HTTPie/3.2.4,Status:200,Latency:51.242708ms,BytesIn:74,BytesOut:563,Date:2025-01-03 12:13:33.234758 +0900 KST}",::1
					229, 135, 166, 231, 144, 134, 230, 151, 165, 230, 153, 130, 44, 233, 129, 139, 231, 148, 168, 232, 128, 133, 229, 144, 141, 40, 227, 131, 161, 227, 131, 188, 227, 131, 171, 227, 130, 162, 227, 131, 137, 227, 131, 172, 227, 130, 185, 41, 44, 227, 131, 173, 227, 130, 176, 231, 168, 174, 229, 136, 165, 44, 227, 131, 173, 227, 130, 176, 44, 228, 189, 156, 230, 165, 173, 229, 134, 133, 229, 174, 185, 44, 73, 80, 10, 50, 48, 50, 53, 45, 48, 49, 45, 48, 51, 32, 49, 50, 58, 49, 51, 58, 51, 51, 44, 34, 123, 73, 100, 58, 49, 44, 78, 97, 109, 101, 58, 97, 100, 109, 105, 110, 44, 69, 109, 97, 105, 108, 58, 97, 100, 109, 105, 110, 64, 101, 110, 101, 114, 103, 121, 115, 104, 97, 114, 101, 115, 46, 106, 112, 125, 34, 44, 99, 111, 109, 109, 111, 110, 99, 111, 100, 101, 46, 76, 79, 71, 95, 67, 79, 78, 84, 69, 78, 84, 95, 84, 89, 80, 69, 46, 76, 79, 71, 73, 78, 44, 34, 123, 76, 111, 103, 84, 121, 112, 101, 58, 114, 101, 113, 117, 101, 115, 116, 44, 82, 101, 113, 117, 101, 115, 116, 73, 68, 58, 101, 55, 50, 99, 57, 100, 100, 50, 45, 99, 56, 52, 52, 45, 49, 49, 101, 102, 45, 56, 98, 56, 101, 45, 102, 54, 102, 48, 100, 57, 52, 54, 55, 52, 52, 57, 44, 82, 101, 109, 111, 116, 101, 73, 80, 58, 58, 58, 49, 44, 72, 111, 115, 116, 58, 108, 111, 99, 97, 108, 104, 111, 115, 116, 58, 56, 48, 56, 48, 44, 79, 112, 101, 114, 97, 116, 105, 111, 110, 58, 47, 97, 112, 105, 47, 118, 49, 47, 97, 100, 109, 105, 110, 45, 117, 115, 101, 114, 115, 47, 108, 111, 103, 105, 110, 44, 85, 82, 73, 58, 47, 97, 112, 105, 47, 118, 49, 47, 97, 100, 109, 105, 110, 45, 117, 115, 101, 114, 115, 47, 108, 111, 103, 105, 110, 44, 77, 101, 116, 104, 111, 100, 58, 80, 79, 83, 84, 44, 82, 101, 102, 101, 114, 101, 114, 58, 44, 85, 115, 101, 114, 65, 103, 101, 110, 116, 58, 72, 84, 84, 80, 105, 101, 47, 51, 46, 50, 46, 52, 44, 83, 116, 97, 116, 117, 115, 58, 50, 48, 48, 44, 76, 97, 116, 101, 110, 99, 121, 58, 53, 49, 46, 50, 52, 50, 55, 48, 56, 109, 115, 44, 66, 121, 116, 101, 115, 73, 110, 58, 55, 52, 44, 66, 121, 116, 101, 115, 79, 117, 116, 58, 53, 54, 51, 44, 68, 97, 116, 101, 58, 50, 48, 50, 53, 45, 48, 49, 45, 48, 51, 32, 49, 50, 58, 49, 51, 58, 51, 51, 46, 50, 51, 52, 55, 53, 56, 32, 43, 48, 57, 48, 48, 32, 75, 83, 84, 125, 34, 44, 58, 58, 49, 10,
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u := operatorHistoryUsecase{
				operatorHistoryKsqlRepository: tt.fields.operatorHistoryKsqlRepository(),
				cfg:                           tt.fields.cfg,
				translator:                    tt.fields.translator,
			}

			_, err := u.Download(tt.args.ctx, tt.args.request)

			assert.NoError(t, err, "operatorHistoryUsecase.Download() error = %v, wantErr %v", err, tt.wantErr)
		})
	}
}
