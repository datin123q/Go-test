package notice_msg

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/samber/lo"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	adomain "github.com/illuminarean/energy-share-core-api/domain"
	cdomain "github.com/illuminarean/energy-shares-backend-repo/clean-arch-domain"
	"github.com/illuminarean/energy-shares-backend-repo/clean-arch-domain/mocks"
	"github.com/illuminarean/energy-shares-backend-repo/domain"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/jsonutils"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/jwthelper"
)

func Test_noticeMsgUsecase_GetAll(t *testing.T) {
	type args struct {
		ctx     context.Context
		request adomain.NoticeMsgGetAllRequest
	}
	tests := []struct {
		name                        string
		args                        args
		noticeMsgRepository         func() *mocks.NoticeMsgRepository
		userNoticeMappingRepository func() *mocks.UserNoticeMappingRepository
		want                        adomain.NoticeMsgGetAllResponse
		wantErr                     bool
	}{
		{
			name: "全お知らせ情報リスト取得 (成功)",
			args: args{
				ctx: context.Background(),
				request: adomain.NoticeMsgGetAllRequest{
					JWTUser: jwthelper.User{
						Id:    12378123,
						Name:  "홍길동",
						Email: "gildong@email.com",
					},
					ClientIp: "127.0.0.1",
					NoticeMsgGetAllQueryParam: adomain.NoticeMsgGetAllQueryParam{
						Title: lo.ToPtr("123"),
						SearchInput: cdomain.SearchInput{
							SkipCount:      0,
							MaxResultCount: 10,
							Sortby:         "id",
							Order:          "asc",
						},
					},
				},
			},
			noticeMsgRepository: func() *mocks.NoticeMsgRepository {
				repo := new(mocks.NoticeMsgRepository)
				repo.On("GetAll", context.Background(),
					mock.AnythingOfType("func(*clean_arch_domain.NoticeMsgQueryOptionFields)"),
					mock.AnythingOfType("func(*clean_arch_domain.NoticeMsgQueryOptionFields)"),
					mock.AnythingOfType("func(*clean_arch_domain.NoticeMsgQueryOptionFields)"),
					mock.AnythingOfType("func(*clean_arch_domain.NoticeMsgQueryOptionFields)"),
					mock.AnythingOfType("func(*clean_arch_domain.NoticeMsgQueryOptionFields)"),
					mock.AnythingOfType("func(*clean_arch_domain.NoticeMsgQueryOptionFields)"),
					mock.AnythingOfType("func(*clean_arch_domain.NoticeMsgQueryOptionFields)"),
				).Return(int64(2), []domain.NoticeMsg{mockNoticeMsg6, mockNoticeMsg11}, nil)
				return repo
			},
			userNoticeMappingRepository: func() *mocks.UserNoticeMappingRepository {
				repo := new(mocks.UserNoticeMappingRepository)
				repo.On("GetAll", context.Background(),
					mock.AnythingOfType("func(*clean_arch_domain.UserNoticeMappingQueryOptionFields)"),
				).Return(int64(1), []domain.UserNoticeMapping{mockUserNoticeMapping4}, nil)
				repo.On("CountByMsgIds", context.Background(),
					[]int64{int64(6), int64(11)},
				).Return(int64(2), []domain.NoticeMsgUserCount{
					{MsgId: 6, UserCount: 0},
					{MsgId: 11, UserCount: mockUserNoticeMapping4_userCount},
				}, nil)
				return repo
			},
			want: adomain.NoticeMsgGetAllResponse{
				TotalCount: 2,
				Items:      []domain.NoticeMsg{mockNoticeMsg6, mockNoticeMsg11_2},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			uc := noticeMsgUsecase{
				noticeMsgRepository:         tt.noticeMsgRepository(),
				userNoticeMappingRepository: tt.userNoticeMappingRepository(),
			}
			ignoreTime := cmpopts.IgnoreTypes(jsonutils.JSONTime{})

			got, err := uc.GetAll(tt.args.ctx, tt.args.request)

			assert.NoErrorf(t, err, "noticeMsgUsecase.GetAll() error = %v, wantErr %v", err, tt.wantErr)
			assert.Equal(t, tt.want.TotalCount, got.TotalCount)
			assert.Truef(t, cmp.Equal(tt.want.Items, got.Items, ignoreTime), "noticeMsgUsecase.GetAll() got = %v, want %v", got.Items, tt.want.Items)
		})
	}
}

func Test_noticeMsgUsecase_GetOne(t *testing.T) {
	type args struct {
		ctx     context.Context
		request adomain.NoticeMsgGetOneRequest
	}
	tests := []struct {
		name                        string
		args                        args
		noticeMsgRepository         func() *mocks.NoticeMsgRepository
		userNoticeMappingRepository func() *mocks.UserNoticeMappingRepository
		userRepository              func() *mocks.UserRepository
		projectRepository           func() *mocks.ProjectRepository
		want                        adomain.NoticeMsgGetOneResponse
		wantErr                     bool
	}{
		{
			name: "1つのお知らせ情報取得 (成功)",
			args: args{
				ctx: context.Background(),
				request: adomain.NoticeMsgGetOneRequest{
					Id: 1,
					JWTUser: jwthelper.User{
						Id:    12378123,
						Name:  "홍길동",
						Email: "gildong@email.com",
					},
					ClientIp: "127.0.0.1",
				},
			},
			noticeMsgRepository: func() *mocks.NoticeMsgRepository {
				repo := new(mocks.NoticeMsgRepository)
				repo.On("GetById", context.Background(), int64(1)).
					Return(mockNoticeMsg1, nil)
				return repo
			},
			userNoticeMappingRepository: func() *mocks.UserNoticeMappingRepository {
				repo := new(mocks.UserNoticeMappingRepository)
				repo.On("GetAll", context.Background(),
					mock.AnythingOfType("func(*clean_arch_domain.UserNoticeMappingQueryOptionFields)"),
				).Return(int64(1), []domain.UserNoticeMapping{mockUserNoticeMapping1}, nil)
				repo.On("CountByMsgIds", context.Background(),
					[]int64{int64(1)},
				).Return(int64(1), []domain.NoticeMsgUserCount{
					{MsgId: 1, UserCount: mockUserNoticeMapping1_userCount},
				}, nil)
				return repo
			},
			userRepository: func() *mocks.UserRepository {
				repo := new(mocks.UserRepository)
				repo.On("GetAll", context.Background(),
					mock.AnythingOfType("func(*clean_arch_domain.UserQueryOptionFields)"),
				).Return(int64(1), []domain.User{user111}, nil)
				return repo
			},
			projectRepository: func() *mocks.ProjectRepository {
				repo := new(mocks.ProjectRepository)
				repo.On("GetById", context.Background(), int64(1)).
					Return(domain.Project{Id: 0}, nil)
				return repo
			},
			want: func() adomain.NoticeMsgGetOneResponse {
				r := adomain.NoticeMsgGetOneResponse{NoticeMsg: mockNoticeMsg1_2}
				r.Users = []domain.User{user111}
				return r
			}(),
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			uc := noticeMsgUsecase{
				noticeMsgRepository:         tt.noticeMsgRepository(),
				userNoticeMappingRepository: tt.userNoticeMappingRepository(),
				userRepository:              tt.userRepository(),
				projectRepository:           tt.projectRepository(),
			}
			ignoreTime := cmpopts.IgnoreTypes(jsonutils.JSONTime{})

			got, err := uc.GetOne(tt.args.ctx, tt.args.request)

			assert.NoErrorf(t, err, "noticeMsgUsecase.GetOne() error = %v, wantErr %v", err, tt.wantErr)
			assert.Truef(t, cmp.Equal(tt.want.NoticeMsg, got.NoticeMsg, ignoreTime), "noticeMsgUsecase.GetOne() got = %v, want %v", got.NoticeMsg, tt.want.NoticeMsg)
		})
	}
}

func Test_noticeMsgUsecase_Create(t *testing.T) {
	type args struct {
		ctx     context.Context
		request adomain.NoticeMsgCreateRequest
	}
	tests := []struct {
		name                        string
		args                        args
		noticeMsgRepository         func() *mocks.NoticeMsgRepository
		userNoticeMappingRepository func() *mocks.UserNoticeMappingRepository
		wantErr                     bool
	}{
		{
			name: "お知らせ情報作成 (成功)",
			args: args{
				ctx: context.Background(),
				request: adomain.NoticeMsgCreateRequest{
					UserIds: []int64{111},
					JWTUser: jwthelper.User{
						Id:    12378123,
						Name:  "홍길동",
						Email: "gildong@email.com",
					},
					ClientIp: "127.0.0.1",
					NoticeMsgCreateRequestBase: adomain.NoticeMsgCreateRequestBase{
						Title:       "example title",
						ContentHtml: "<p>something</p>",
					},
				},
			},
			noticeMsgRepository: func() *mocks.NoticeMsgRepository {
				repo := new(mocks.NoticeMsgRepository)
				repo.On("Create", context.Background(),
					mock.AnythingOfType("domain.NoticeMsg"),
				).Return(int64(1), nil)
				return repo
			},
			userNoticeMappingRepository: func() *mocks.UserNoticeMappingRepository {
				repo := new(mocks.UserNoticeMappingRepository)
				repo.On("Create", context.Background(),
					mock.AnythingOfType("domain.UserNoticeMapping"),
				).Return(int64(1), nil)
				return repo
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			uc := noticeMsgUsecase{
				noticeMsgRepository:         tt.noticeMsgRepository(),
				userNoticeMappingRepository: tt.userNoticeMappingRepository(),
			}

			err := uc.Create(tt.args.ctx, tt.args.request)

			assert.NoErrorf(t, err, "noticeMsgUsecase.Create() error = %v, wantErr %v", err, tt.wantErr)
		})
	}
}

func Test_noticeMsgUsecase_CreateByProject(t *testing.T) {
	type args struct {
		ctx     context.Context
		request adomain.NoticeMsgCreateByProjectRequest
	}
	tests := []struct {
		name                        string
		args                        args
		noticeMsgRepository         func() *mocks.NoticeMsgRepository
		userNoticeMappingRepository func() *mocks.UserNoticeMappingRepository
		projectRepository           func() *mocks.ProjectRepository
		sharesRepository            func() *mocks.ShareRepository
		wantErr                     bool
	}{
		{
			name: "プロジェクト別お知らせ情報作成 (成功)",
			args: args{
				ctx: context.Background(),
				request: adomain.NoticeMsgCreateByProjectRequest{
					ProjectId: 1,
					JWTUser: jwthelper.User{
						Id:    12378123,
						Name:  "홍길동",
						Email: "gildong@email.com",
					},
					ClientIp: "127.0.0.1",
					NoticeMsgCreateRequestBase: adomain.NoticeMsgCreateRequestBase{
						Title:       "example title",
						ContentHtml: "<p>something</p>",
					},
				},
			},
			noticeMsgRepository: func() *mocks.NoticeMsgRepository {
				repo := new(mocks.NoticeMsgRepository)
				repo.On("Create", context.Background(),
					mock.AnythingOfType("domain.NoticeMsg"),
				).Return(int64(1), nil)
				return repo
			},
			userNoticeMappingRepository: func() *mocks.UserNoticeMappingRepository {
				repo := new(mocks.UserNoticeMappingRepository)
				repo.On("Create", context.Background(),
					mock.AnythingOfType("domain.UserNoticeMapping"),
				).Return(int64(1), nil)
				return repo
			},
			projectRepository: func() *mocks.ProjectRepository {
				repo := new(mocks.ProjectRepository)
				repo.On("GetById", context.Background(), int64(1)).
					Return(domain.Project{Id: 0}, nil)
				return repo
			},
			sharesRepository: func() *mocks.ShareRepository {
				repo := new(mocks.ShareRepository)
				repo.On("GetInvestor", context.Background(),
					mock.AnythingOfType("func(*clean_arch_domain.ShareQueryOptionFields)"),
					mock.AnythingOfType("func(*clean_arch_domain.ShareQueryOptionFields)"),
				).Return([]domain.User{user107}, nil)
				return repo
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			uc := noticeMsgUsecase{
				noticeMsgRepository:         tt.noticeMsgRepository(),
				userNoticeMappingRepository: tt.userNoticeMappingRepository(),
				projectRepository:           tt.projectRepository(),
				shareRepository:             tt.sharesRepository(),
			}

			err := uc.CreateByProject(tt.args.ctx, tt.args.request)

			assert.NoErrorf(t, err, "noticeMsgUsecase.CreateByProject() error = %v, wantErr %v", err, tt.wantErr)
		})
	}
}

func Test_noticeMsgUsecase_CreateByNewsletter(t *testing.T) {
	type args struct {
		ctx     context.Context
		request adomain.NoticeMsgCreateByNewsletterRequest
	}
	tests := []struct {
		name                        string
		args                        args
		noticeMsgRepository         func() *mocks.NoticeMsgRepository
		userNoticeMappingRepository func() *mocks.UserNoticeMappingRepository
		userRepository              func() *mocks.UserRepository
		wantErr                     bool
	}{
		{
			name: "ニュースレター別お知らせ情報作成 (成功)",
			args: args{
				ctx: context.Background(),
				request: adomain.NoticeMsgCreateByNewsletterRequest{
					JWTUser: jwthelper.User{
						Id:    12378123,
						Name:  "홍길동",
						Email: "gildong@email.com",
					},
					ClientIp: "127.0.0.1",
					NoticeMsgCreateRequestBase: adomain.NoticeMsgCreateRequestBase{
						Title:       "example title",
						ContentHtml: "<p>something</p>",
					},
				},
			},
			noticeMsgRepository: func() *mocks.NoticeMsgRepository {
				repo := new(mocks.NoticeMsgRepository)
				repo.On("Create", context.Background(),
					mock.AnythingOfType("domain.NoticeMsg"),
				).Return(int64(1), nil)
				return repo
			},
			userNoticeMappingRepository: func() *mocks.UserNoticeMappingRepository {
				repo := new(mocks.UserNoticeMappingRepository)
				repo.On("Create", context.Background(),
					mock.AnythingOfType("domain.UserNoticeMapping"),
				).Return(int64(1), nil)
				return repo
			},
			userRepository: func() *mocks.UserRepository {
				repo := new(mocks.UserRepository)
				repo.On("GetAll", context.Background(),
					mock.AnythingOfType("func(*clean_arch_domain.UserQueryOptionFields)"),
				).Return(int64(1), []domain.User{user107}, nil)
				return repo
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			uc := noticeMsgUsecase{
				noticeMsgRepository:         tt.noticeMsgRepository(),
				userNoticeMappingRepository: tt.userNoticeMappingRepository(),
				userRepository:              tt.userRepository(),
			}

			err := uc.CreateByNewsletter(tt.args.ctx, tt.args.request)

			assert.NoErrorf(t, err, "noticeMsgUsecase.CreateByNewsletter() error = %v, wantErr %v", err, tt.wantErr)
		})
	}
}

func Test_NoticeMsgUsecase_CreateByUserType(t *testing.T) {
	type args struct {
		ctx     context.Context
		request adomain.NoticeMsgCreateByUserTypeRequest
	}
	tests := []struct {
		name                        string
		args                        args
		noticeMsgRepository         func() *mocks.NoticeMsgRepository
		userNoticeMappingRepository func() *mocks.UserNoticeMappingRepository
		userRepository              func() *mocks.UserRepository
		wantErr                     bool
	}{
		{
			name: "ユーザータイプ別お知らせ情報作成 (成功)",
			args: args{
				ctx: context.Background(),
				request: adomain.NoticeMsgCreateByUserTypeRequest{
					JWTUser: jwthelper.User{
						Id:    12378123,
						Name:  "홍길동",
						Email: "gildong@email.com",
					},
					ClientIp: "127.0.0.1",
					NoticeMsgCreateRequestBase: adomain.NoticeMsgCreateRequestBase{
						Title:       "example title",
						ContentHtml: "<p>something</p>",
					},
				},
			},
			noticeMsgRepository: func() *mocks.NoticeMsgRepository {
				repo := new(mocks.NoticeMsgRepository)
				repo.On("Create", context.Background(),
					mock.AnythingOfType("domain.NoticeMsg"),
				).Return(int64(1), nil)
				return repo
			},
			userNoticeMappingRepository: func() *mocks.UserNoticeMappingRepository {
				repo := new(mocks.UserNoticeMappingRepository)
				repo.On("Create", context.Background(),
					mock.AnythingOfType("domain.UserNoticeMapping"),
				).Return(int64(1), nil)
				return repo
			},
			userRepository: func() *mocks.UserRepository {
				repo := new(mocks.UserRepository)
				repo.On("GetAll", context.Background(),
					mock.AnythingOfType("func(*clean_arch_domain.UserQueryOptionFields)"),
				).Return(int64(1), []domain.User{user107}, nil)
				return repo
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			uc := noticeMsgUsecase{
				noticeMsgRepository:         tt.noticeMsgRepository(),
				userNoticeMappingRepository: tt.userNoticeMappingRepository(),
				userRepository:              tt.userRepository(),
			}

			err := uc.CreateByUserType(tt.args.ctx, tt.args.request)

			assert.NoErrorf(t, err, "noticeMsgUsecase.CreateByUserType() error = %v, wantErr %v", err, tt.wantErr)
		})
	}
}
