package mail

import (
	"context"
	"testing"

	"github.com/samber/lo"
	"github.com/stretchr/testify/assert"

	"github.com/stretchr/testify/mock"

	adomain "github.com/illuminarean/energy-share-core-api/domain"
	cdomain "github.com/illuminarean/energy-shares-backend-repo/clean-arch-domain"
	"github.com/illuminarean/energy-shares-backend-repo/clean-arch-domain/mocks"
	"github.com/illuminarean/energy-shares-backend-repo/domain"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/config"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/emailutils"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/jwthelper"
)

func Test_MailUsecase_GetAll(t *testing.T) {

	type args struct {
		ctx     context.Context
		request adomain.MailGetAllRequest
	}
	tests := []struct {
		name         string
		args         args
		emailpkg     func() *mocks.EmailRequester
		wantResponse adomain.MailGetAllResponse
		wantErr      bool
	}{
		{
			name: "統合メールサービスから送信リスト照会_成功",
			args: args{
				ctx: context.Background(),
				request: adomain.MailGetAllRequest{
					JWTUser: jwthelper.User{
						Id:    12378123,
						Name:  "ホン・ギルドン",
						Email: "gildong@email.com",
					},
					ClientIp: "127.0.0.1",
					Env:      "DEV",
					MailGetAllQueryParam: adomain.MailGetAllQueryParam{
						Status: lo.ToPtr(string(domain.MAIL_SENT)),
					},
					SearchInput: cdomain.SearchInput{
						SkipCount:      0,
						MaxResultCount: 10,
					},
				},
			},
			emailpkg: func() *mocks.EmailRequester {
				emailpkg := &mocks.EmailRequester{}
				emailpkg.On("GetAllRequest", context.Background(),
					mock.MatchedBy(func(query emailutils.GetAllQuery) bool {
						if query.ServiceName != "ESJP" ||
							query.ServiceEnv != "DEV" ||
							*query.Status != string(domain.MAIL_SENT) {
							return false
						}
						return true
					}),
				).Return(mockEmailListResponse, nil)
				return emailpkg
			},
			wantResponse: func() adomain.MailGetAllResponse {
				mailGetAllResponse := adomain.MailGetAllResponse{}
				mailGetAllResponse.FromEmailListResponse(mockEmailListResponse)
				return mailGetAllResponse
			}(),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u := mailUsecase{
				emailpkg: tt.emailpkg(),
			}
			got, err := u.GetAll(tt.args.ctx, tt.args.request)
			assert.Equalf(t, tt.wantResponse, got, "GetAll() = %v, want %v", got, tt.wantResponse)
			assert.Equalf(t, tt.wantErr, err != nil, "GetAll() error = %v, wantErr %v", err, tt.wantErr)
		})
	}
}

func Test_MailUsecase_GetOne(t *testing.T) {

	type args struct {
		ctx     context.Context
		request adomain.MailGetOneRequest
	}
	tests := []struct {
		name         string
		args         args
		emailpkg     func() *mocks.EmailRequester
		wantResponse adomain.MailDetailResponse
		wantErr      bool
	}{
		{
			name: "統合メールサービスから送信1件照会_成功",
			args: args{
				ctx: context.Background(),
				request: adomain.MailGetOneRequest{
					JWTUser: jwthelper.User{
						Id:    12378123,
						Name:  "ホン・ギルドン",
						Email: "gildong@email.com",
					},
					ClientIp: "127.0.0.1",
					MailId:   "1001",
				},
			},
			emailpkg: func() *mocks.EmailRequester {
				emailpkg := &mocks.EmailRequester{}
				emailpkg.On("GetByIdRequest", context.Background(), "1001").
					Return(mockEmailResponse, nil)
				return emailpkg
			},
			wantResponse: func() adomain.MailDetailResponse {
				mailGetOneResponse := adomain.MailDetailResponse{}
				mailGetOneResponse.FromEmailResponse(mockEmailResponse)
				return mailGetOneResponse
			}(),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u := mailUsecase{
				emailpkg: tt.emailpkg(),
			}
			got, err := u.GetOne(tt.args.ctx, tt.args.request)
			assert.Equalf(t, tt.wantResponse, got, "GetOne() = %v, want %v", got, tt.wantResponse)
			assert.Equalf(t, tt.wantErr, err != nil, "GetOne() error = %v, wantErr %v", err, tt.wantErr)
		})
	}
}

func Test_MailUsecase_GetAllFailed(t *testing.T) {

	type args struct {
		ctx     context.Context
		request adomain.MailGetAllFailedRequest
	}
	tests := []struct {
		name                 string
		args                 args
		mailFailedRepository func() *mocks.MailFailedRepository
		wantResponse         adomain.MailGetAllFailedResponse
		wantErr              bool
	}{
		{
			name: "送信要求失敗リスト照会_成功",
			args: args{
				ctx: context.Background(),
				request: adomain.MailGetAllFailedRequest{
					JWTUser: jwthelper.User{
						Id:    12378123,
						Name:  "ホン・ギルドン",
						Email: "gildong@email.com",
					},
					ClientIp: "127.0.0.1",
					MailGetAllFailedQueryParam: adomain.MailGetAllFailedQueryParam{
						TemplateName: lo.ToPtr("NORMAL_MAIL"),
					},
					SearchInput: cdomain.SearchInput{
						SkipCount:      0,
						MaxResultCount: 10,
					},
				},
			},
			mailFailedRepository: func() *mocks.MailFailedRepository {
				mailFailedRepository := &mocks.MailFailedRepository{}
				mailFailedRepository.On("GetAll", context.Background(),
					mock.AnythingOfType("func(*clean_arch_domain.MailFailedQueryOptionFields)"),
					mock.AnythingOfType("func(*clean_arch_domain.MailFailedQueryOptionFields)"),
					mock.AnythingOfType("func(*clean_arch_domain.MailFailedQueryOptionFields)"),
					mock.AnythingOfType("func(*clean_arch_domain.MailFailedQueryOptionFields)"),
					mock.AnythingOfType("func(*clean_arch_domain.MailFailedQueryOptionFields)"),
					mock.AnythingOfType("func(*clean_arch_domain.MailFailedQueryOptionFields)"),
				).Return(int64(1), []domain.MailFailed{mockMailFailed}, nil)
				return mailFailedRepository
			},
			wantResponse: adomain.MailGetAllFailedResponse{
				TotalCount: int64(1),
				Items:      []adomain.MailFailedItem{mockMailFailedItem},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u := mailUsecase{
				mailFailedRepository: tt.mailFailedRepository(),
			}
			got, err := u.GetAllFailed(tt.args.ctx, tt.args.request)
			assert.Equalf(t, tt.wantResponse, got, "GetAllFailed() = %v, want %v", got, tt.wantResponse)
			assert.Equalf(t, tt.wantErr, err != nil, "GetAllFailed() error = %v, wantErr %v", err, tt.wantErr)
		})
	}
}

func Test_MailUsecase_GetTemplate(t *testing.T) {

	type args struct {
		ctx     context.Context
		request adomain.MailGetTemplateRequest
	}
	tests := []struct {
		name         string
		args         args
		emailpkg     func() *mocks.EmailRequester
		wantResponse adomain.MailGetTemplateResponse
		wantErr      bool
	}{
		{
			name: "統合メールサービスからメールテンプレート1件照会_成功",
			args: args{
				ctx: context.Background(),
				request: adomain.MailGetTemplateRequest{
					JWTUser: jwthelper.User{
						Id:    12378123,
						Name:  "ホン・ギルドン",
						Email: "gildong@email.com",
					},
					ClientIp:     "127.0.0.1",
					TemplateType: "NORMAL_MAIL",
				},
			},
			emailpkg: func() *mocks.EmailRequester {
				emailpkg := &mocks.EmailRequester{}
				// 最後に環境変数が付く必要があるが、テストでは不可能。例) NORMAL_MAIL
				emailpkg.On("GetTemplate", context.Background(), "NORMAL_MAIL").
					Return(mockEmailTemplateResponse, nil)
				return emailpkg
			},
			wantResponse: func() adomain.MailGetTemplateResponse {
				mailGetOneResponse := adomain.MailGetTemplateResponse{}
				mailGetOneResponse.FromMailTemplate(mockEmailTemplateResponse)
				return mailGetOneResponse
			}(),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &config.Config{}
			cfg.CommonConfig.ServiceConfig.Url.UserServiceDomain = "http://localhost:8080"

			u := mailUsecase{
				cfg:      cfg,
				emailpkg: tt.emailpkg(),
			}
			got, err := u.GetTemplate(tt.args.ctx, tt.args.request)
			assert.Equalf(t, tt.wantResponse, got, "GetTemplate() = %v, want %v", got, tt.wantResponse)
			assert.Equalf(t, tt.wantErr, err != nil, "GetTemplate() error = %v, wantErr %v", err, tt.wantErr)
		})
	}
}
