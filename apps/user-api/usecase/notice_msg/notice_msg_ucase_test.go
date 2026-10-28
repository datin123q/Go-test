package notice_msg

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/illuminarean/energy-shares-backend-repo/clean-arch-domain/mocks"
	"github.com/illuminarean/energy-shares-backend-repo/domain"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/jsonutils"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/jwthelper"
	udomain "github.com/illuminarean/energyshares-user-api/domain"
)

func Test_noticeMsgUsecase_GetAll(t *testing.T) {
	type fields struct {
		noticeMsgRepository         func() *mocks.NoticeMsgRepository
		noticeMsgReadRepository     func() *mocks.NoticeMsgReadRepository
		userNoticeMappingRepository func() *mocks.UserNoticeMappingRepository
	}
	type args struct {
		ctx     context.Context
		request udomain.NoticeMsgGetAllRequest
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    udomain.NoticeMsgGetAllResponse
		wantErr bool
	}{
		{
			name: "ユーザー通知情報を要求 => 正常",
			fields: fields{
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
						mock.AnythingOfType("func(*clean_arch_domain.NoticeMsgQueryOptionFields)"),
					).Return(int64(4), []domain.NoticeMsg{mockNoticeMsg1, mockNoticeMsg2, mockNoticeMsg4, mockNoticeMsg11}, nil)
					return repo
				},
				noticeMsgReadRepository: func() *mocks.NoticeMsgReadRepository {
					repo := new(mocks.NoticeMsgReadRepository)
					repo.On("GetAll", context.Background(),
						mock.AnythingOfType("func(*clean_arch_domain.NoticeMsgReadQueryOptionFields)"),
					).Return(int64(1), []domain.NoticeMsgRead{mockNoticeMsgRead1}, nil)
					return repo
				},
				userNoticeMappingRepository: func() *mocks.UserNoticeMappingRepository {
					repo := new(mocks.UserNoticeMappingRepository)
					repo.On("GetAll", context.Background(),
						mock.AnythingOfType("func(*clean_arch_domain.UserNoticeMappingQueryOptionFields)"),
					).Return(int64(4), []domain.UserNoticeMapping{mockUserNoticeMapping1, mockUserNoticeMapping2, mockUserNoticeMapping3, mockUserNoticeMapping4}, nil)
					repo.On("CountByMsgIds", context.Background(),
						[]int64{int64(1), int64(2), int64(4), int64(11)},
					).Return(int64(4), []domain.NoticeMsgUserCount{
						{MsgId: int64(1), UserCount: mockUserNoticeMapping1_userCount},
						{MsgId: int64(2), UserCount: mockUserNoticeMapping2_userCount},
						{MsgId: int64(4), UserCount: mockUserNoticeMapping3_userCount},
						{MsgId: int64(11), UserCount: mockUserNoticeMapping4_userCount},
					}, nil)
					return repo
				},
			},
			args: args{
				ctx: context.Background(),
				request: udomain.NoticeMsgGetAllRequest{
					JWTUser: jwthelper.User{
						Id:    111,
						Name:  "ホン・ギルドン",
						Email: "gildong@email.com",
					},
					ClientIp: "127.0.0.1",
				},
			},
			want: udomain.NoticeMsgGetAllResponse{
				TotalCount: 4,
				Items:      []domain.NoticeMsg{mockNoticeMsg1_2, mockNoticeMsg2_2, mockNoticeMsg4_2, mockNoticeMsg11_2},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			uc := noticeMsgUsecase{
				noticeMsgRepository:         tt.fields.noticeMsgRepository(),
				noticeMsgReadRepository:     tt.fields.noticeMsgReadRepository(),
				userNoticeMappingRepository: tt.fields.userNoticeMappingRepository(),
			}
			ignoreTime := cmpopts.IgnoreTypes(jsonutils.JSONTime{})

			got, err := uc.GetAll(tt.args.ctx, tt.args.request)

			assert.NoErrorf(t, err, "GetAll() error = %v, wantErr %v", err, tt.wantErr)
			assert.Truef(t, cmp.Equal(got, tt.want, ignoreTime), "GetAll() got = %v, want %v", got, tt.want)
		})
	}
}

func Test_noticeMsgUsecase_GetOne(t *testing.T) {
	type fields struct {
		noticeMsgRepository         func() *mocks.NoticeMsgRepository
		noticeMsgReadRepository     func() *mocks.NoticeMsgReadRepository
		userNoticeMappingRepository func() *mocks.UserNoticeMappingRepository
	}
	type args struct {
		ctx     context.Context
		request udomain.NoticeMsgGetOneRequest
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    udomain.NoticeMsgGetOneResponse
		wantErr bool
	}{
		{
			name: "ユーザー通知情報を要求 => 正常",
			fields: fields{
				noticeMsgRepository: func() *mocks.NoticeMsgRepository {
					repo := new(mocks.NoticeMsgRepository)
					repo.On("GetById", context.Background(), int64(1)).
						Return(mockNoticeMsg1, nil)
					return repo
				},
				noticeMsgReadRepository: func() *mocks.NoticeMsgReadRepository {
					repo := new(mocks.NoticeMsgReadRepository)
					repo.On("GetAll", context.Background(),
						mock.AnythingOfType("func(*clean_arch_domain.NoticeMsgReadQueryOptionFields)"),
						mock.AnythingOfType("func(*clean_arch_domain.NoticeMsgReadQueryOptionFields)"),
					).Return(int64(1), []domain.NoticeMsgRead{mockNoticeMsgRead1}, nil)
					repo.On("Create", context.Background(),
						mock.AnythingOfType("domain.NoticeMsgRead"),
					).Return(int64(3), nil)
					repo.On("Update", context.Background(),
						mock.AnythingOfType("domain.NoticeMsgRead"),
					).Return(nil)
					return repo
				},
				userNoticeMappingRepository: func() *mocks.UserNoticeMappingRepository {
					repo := new(mocks.UserNoticeMappingRepository)
					repo.On("GetAll", context.Background(),
						mock.AnythingOfType("func(*clean_arch_domain.UserNoticeMappingQueryOptionFields)"),
						mock.AnythingOfType("func(*clean_arch_domain.UserNoticeMappingQueryOptionFields)"),
					).Return(int64(1), []domain.UserNoticeMapping{mockUserNoticeMapping1}, nil)
					repo.On("CountByMsgIds", context.Background(),
						[]int64{int64(1)},
					).Return(int64(1), []domain.NoticeMsgUserCount{{
						MsgId:     int64(1),
						UserCount: mockUserNoticeMapping1_userCount,
					}}, nil)
					return repo
				},
			},
			args: args{
				ctx: context.Background(),
				request: udomain.NoticeMsgGetOneRequest{
					Id: 1,
					JWTUser: jwthelper.User{
						Id:    111,
						Name:  "ホン・ギルドン",
						Email: "gildong@email.com",
					},
					ClientIp: "127.0.0.1",
				},
			},
			want:    udomain.NoticeMsgGetOneResponse{NoticeMsg: mockNoticeMsg1_2},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			uc := noticeMsgUsecase{
				noticeMsgRepository:         tt.fields.noticeMsgRepository(),
				noticeMsgReadRepository:     tt.fields.noticeMsgReadRepository(),
				userNoticeMappingRepository: tt.fields.userNoticeMappingRepository(),
			}
			ignoreTime := cmpopts.IgnoreTypes(jsonutils.JSONTime{})

			got, err := uc.GetOne(tt.args.ctx, tt.args.request)
			assert.NoErrorf(t, err, "GetOne() error = %v, wantErr %v", err, tt.wantErr)
			assert.Truef(t, cmp.Equal(got, tt.want, ignoreTime), "GetOne() got = %v, want %v", got, tt.want)
		})
	}
}
