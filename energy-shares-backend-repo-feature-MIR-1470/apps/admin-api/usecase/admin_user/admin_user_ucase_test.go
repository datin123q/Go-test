package admin_user

import (
	"context"
	"net/http"
	"testing"

	"github.com/gofrs/uuid"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	adomain "github.com/illuminarean/energy-share-core-api/domain"
	cdomain "github.com/illuminarean/energy-shares-backend-repo/clean-arch-domain"
	"github.com/illuminarean/energy-shares-backend-repo/clean-arch-domain/mocks"
	"github.com/illuminarean/energy-shares-backend-repo/domain"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/jsonutils"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/jwthelper"
	"github.com/samber/lo"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func Test_adminUsecase_GetAll(t *testing.T) {
	type args struct {
		ctx     context.Context
		request adomain.AdminUserGetAllRequest
	}

	tests := []struct {
		name                      string
		args                      args
		adminUserRepository       func() *mocks.AdminUserRepository
		operatorHistoryRepository func() *mocks.OperatorHistoryRepository
		want                      adomain.AdminUserGetAllResponse
		wantErr                   bool
	}{
		{
			name: "(成功) 管理者リスト照会",
			args: args{
				ctx: context.Background(),
				request: adomain.AdminUserGetAllRequest{
					JWTUser: jwthelper.User{
						Id:    mockSuperAdmin.Id,
						Email: mockSuperAdmin.Email,
						Name:  mockSuperAdmin.Name,
						Roles: map[string]int{"ROLE_ADMIN": 1, "ROLE_SUPER_ADMIN": 1},
					},
					ClientIp: "clientIp",
					AdminUserGetAllQueryParam: adomain.AdminUserGetAllQueryParam{
						SearchInput: cdomain.SearchInput{
							SkipCount:      0,
							MaxResultCount: 10,
							Order:          "asc",
							Sortby:         "id",
						},
					},
				},
			},
			adminUserRepository: func() *mocks.AdminUserRepository {
				repo := &mocks.AdminUserRepository{}
				repo.On("GetAll", mock.Anything,
					mock.Anything, // name
					mock.Anything, // mobilePhone
					mock.MatchedBy(func(field func(*cdomain.AdminUserQueryOptionFields)) bool {
						opt := cdomain.AdminUserQueryOptionFunc()
						field(&opt)
						return opt.SortBy() == "id"
					}), // SortBy
					mock.MatchedBy(func(field func(*cdomain.AdminUserQueryOptionFields)) bool {
						opt := cdomain.AdminUserQueryOptionFunc()
						field(&opt)
						return opt.Order() == "asc"
					}), // Order
					mock.MatchedBy(func(field func(*cdomain.AdminUserQueryOptionFields)) bool {
						opt := cdomain.AdminUserQueryOptionFunc()
						field(&opt)
						return opt.Offset() == 0
					}), // SkipCount
					mock.MatchedBy(func(field func(*cdomain.AdminUserQueryOptionFields)) bool {
						opt := cdomain.AdminUserQueryOptionFunc()
						field(&opt)
						return opt.Limit() == 10
					}), // MaxResultCount
				).Return(int64(1), []domain.AdminUser{
					mockAdminUser,
				}, nil)
				return repo
			},
			operatorHistoryRepository: func() *mocks.OperatorHistoryRepository {
				repo := &mocks.OperatorHistoryRepository{}
				repo.On("Create", mock.Anything, mock.MatchedBy(func(op domain.OperatorHistory) bool {
					ignoreTypes := cmpopts.IgnoreTypes(map[string]int{}, jsonutils.JSONTime{})
					return cmp.Equal(op, domain.OperatorHistory{
						UserId: mockSuperAdmin.Id,
						Name:   mockSuperAdmin.Name,
						Email:  mockSuperAdmin.Email,
						Menu:   MENU_ADMIN_USER_LIST,
						Ip:     "clientIp",
						State:  domain.LOOKUP,
					}, ignoreTypes)
				})).Return(nil)
				return repo
			},
			want: adomain.AdminUserGetAllResponse{
				TotalCount: int64(1),
				Items: []adomain.AdminUserItem{
					{
						Id:                mockAdminUser.Id,
						CommonId:          mockAdminUser.CommonId,
						Name:              mockAdminUser.Name,
						Email:             mockAdminUser.Email,
						MobilePhone:       mockAdminUser.MobilePhone,
						Roles:             mockAdminUser.Roles,
						LastLoginDate:     mockAdminUser.LastLoginDate,
						IsSuperAdmin:      false,
						ValidityStartDate: mockAdminUser.ValidityStartDate,
						ValidityEndDate:   mockAdminUser.ValidityEndDate,
					},
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u := &adminUserUsecase{
				adminUserRepository:       tt.adminUserRepository(),
				operatorHistoryRepository: tt.operatorHistoryRepository(),
			}

			got, err := u.GetAll(tt.args.ctx, tt.args.request)
			assert.Equalf(t, tt.wantErr, err != nil, "GetAll() error = %v, wantErr %v", err, tt.wantErr)
			assert.Equalf(t, tt.want, got, "GetAll() got = %v, want %v", got, tt.want)
		})
	}
}

func Test_adminUsecase_GetOne(t *testing.T) {
	type args struct {
		ctx     context.Context
		request adomain.AdminUserGetOneRequest
	}

	tests := []struct {
		name                      string
		args                      args
		adminUserRepository       func() *mocks.AdminUserRepository
		operatorHistoryRepository func() *mocks.OperatorHistoryRepository
		want                      adomain.AdminUserGetOneResponse
		wantErr                   bool
	}{
		{
			name: "(成功) 管理者詳細照会",
			args: args{
				ctx: context.Background(),
				request: adomain.AdminUserGetOneRequest{
					JWTUser: jwthelper.User{
						Id:    mockSuperAdmin.Id,
						Email: mockSuperAdmin.Email,
						Name:  mockSuperAdmin.Name,
						Roles: map[string]int{"ROLE_ADMIN": 1, "ROLE_SUPER_ADMIN": 1},
					},
					ClientIp: "clientIp",
					Id:       mockAdminUser.Id,
				},
			},
			adminUserRepository: func() *mocks.AdminUserRepository {
				repo := &mocks.AdminUserRepository{}
				repo.On("GetById", mock.Anything, mockAdminUser.Id).Return(mockAdminUser, nil)
				return repo
			},
			operatorHistoryRepository: func() *mocks.OperatorHistoryRepository {
				repo := &mocks.OperatorHistoryRepository{}
				repo.On("Create", mock.Anything, mock.MatchedBy(func(op domain.OperatorHistory) bool {
					ignoreTypes := cmpopts.IgnoreTypes(map[string]int{}, jsonutils.JSONTime{})
					return cmp.Equal(op, domain.OperatorHistory{
						UserId: mockSuperAdmin.Id,
						Name:   mockSuperAdmin.Name,
						Email:  mockSuperAdmin.Email,
						Menu:   MENU_ADMIN_USER_DETAIL,
						Ip:     "clientIp",
						State:  domain.LOOKUP,
					}, ignoreTypes)
				})).Return(nil)
				return repo
			},
			want: adomain.AdminUserGetOneResponse{
				AdminUserItem: adomain.AdminUserItem{
					Id:                mockAdminUser.Id,
					CommonId:          mockAdminUser.CommonId,
					Name:              mockAdminUser.Name,
					Email:             mockAdminUser.Email,
					MobilePhone:       mockAdminUser.MobilePhone,
					Roles:             mockAdminUser.Roles,
					LastLoginDate:     mockAdminUser.LastLoginDate,
					IsSuperAdmin:      false,
					ValidityStartDate: mockAdminUser.ValidityStartDate,
					ValidityEndDate:   mockAdminUser.ValidityEndDate,
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u := &adminUserUsecase{
				adminUserRepository:       tt.adminUserRepository(),
				operatorHistoryRepository: tt.operatorHistoryRepository(),
			}

			got, err := u.GetOne(tt.args.ctx, tt.args.request)
			assert.Equalf(t, tt.wantErr, err != nil, "GetOne() error = %v, wantErr %v", err, tt.wantErr)
			assert.Equalf(t, tt.want, got, "GetOne() got = %v, want %v", got, tt.want)
		})
	}
}

func Test_adminUserUsecase_Create(t *testing.T) {
	uid, _ := uuid.FromString("bddad349-d0cc-4a1d-8843-431d54fcd2c7")

	type args struct {
		ctx     context.Context
		request adomain.AdminUserCreateRequest
	}

	tests := []struct {
		name                      string
		args                      args
		idGenerator               func() *mocks.IdGenerator
		pwhelper                  func() *mocks.PasswordHelper
		adminUserRepository       func() *mocks.AdminUserRepository
		roleRepository            func() *mocks.RoleRepository
		operatorHistoryRepository func() *mocks.OperatorHistoryRepository
		adminUserService          func() *mocks.AdminUserService
		want                      adomain.AdminUserCreateResponse
		wantErr                   bool
	}{
		{
			name: "(成功) 管理者作成",
			args: args{
				ctx: context.Background(),
				request: adomain.AdminUserCreateRequest{
					JWTUser: jwthelper.User{
						Id:    mockSuperAdmin.Id,
						Email: mockSuperAdmin.Email,
						Name:  mockSuperAdmin.Name,
						Roles: map[string]int{"ROLE_ADMIN": 1, "ROLE_SUPER_ADMIN": 1},
					},
					ClientIp:          "clientIp",
					Email:             "admin-user@example.com",
					Password:          "password1234",
					Name:              "admin-user",
					MobilePhone:       "01011112222",
					Roles:             []string{"ROLE_INVESTOR"},
					ValidityStartDate: "2024-02-09 00:00:00",
					ValidityEndDate:   "2024-12-19 23:59:59",
				},
			},
			idGenerator: func() *mocks.IdGenerator {
				gen := new(mocks.IdGenerator)
				gen.On("GenUuidV4").Return(uid)
				gen.On("GenUuidV5", uid, "admin-user@example.com").Return(uid)
				return gen
			},
			pwhelper: func() *mocks.PasswordHelper {
				helper := &mocks.PasswordHelper{}
				helper.On("GenerateSalt").Return("salt")
				helper.On("EncodePassword", "password1234", "salt").Return("encodedPassword")
				return helper
			},
			adminUserRepository: func() *mocks.AdminUserRepository {
				repo := &mocks.AdminUserRepository{}
				repo.On("GetByEmail", mock.Anything, "admin-user@example.com").Return(domain.AdminUser{}, cdomain.ErrItemNotFound)
				repo.On("Create", mock.Anything, mock.MatchedBy(func(admin domain.AdminUser) bool {
					ignoreTypes := cmpopts.IgnoreTypes(jsonutils.JSONTime{})
					ignoreFields := cmpopts.IgnoreFields(domain.AdminUser{}, "CommonId")
					return cmp.Equal(admin, domain.AdminUser{
						Email:        "admin-user@example.com",
						Salt:         "salt",
						Password:     "encodedPassword",
						Name:         "admin-user",
						MobilePhone:  "01011112222",
						LoginFailCnt: 0,
					}, ignoreTypes, ignoreFields)
				})).Return(domain.AdminUser{
					Id:                int64(123),
					CommonId:          uid.String(),
					Email:             "admin-user@example.com",
					Salt:              "salt",
					Password:          "encodedPassword",
					Name:              "admin-user",
					MobilePhone:       "01011112222",
					LoginFailCnt:      0,
					IsSuperAdmin:      false,
					ValidityStartDate: jsonutils.JSONTime{},
					ValidityEndDate:   jsonutils.JSONTime{},
				}, nil)
				return repo
			},
			roleRepository: func() *mocks.RoleRepository {
				repo := &mocks.RoleRepository{}
				repo.On("Create", mock.Anything, domain.Role{
					UserId:   int64(123),
					RoleCode: string(domain.ROLE_INVESTOR),
				}).Return(int64(1), nil).Once()
				repo.On("Create", mock.Anything, domain.Role{
					UserId:   int64(123),
					RoleCode: string(domain.ROLE_ADMIN),
				}).Return(int64(1), nil).Once()

				return repo
			},
			operatorHistoryRepository: func() *mocks.OperatorHistoryRepository {
				repo := &mocks.OperatorHistoryRepository{}
				repo.On("Create", mock.Anything, mock.MatchedBy(func(op domain.OperatorHistory) bool {
					ignoreTypes := cmpopts.IgnoreTypes(map[string]int{}, jsonutils.JSONTime{})
					return cmp.Equal(op, domain.OperatorHistory{
						UserId: mockSuperAdmin.Id,
						Name:   mockSuperAdmin.Name,
						Email:  mockSuperAdmin.Email,
						Menu:   MENU_ADMIN_USER_REGISTER,
						Ip:     "clientIp",
						State:  domain.REGISTER,
					}, ignoreTypes)
				})).Return(nil)
				return repo
			},
			want: adomain.AdminUserCreateResponse{
				AdminUserItem: adomain.AdminUserItem{
					Id:           int64(123),
					CommonId:     uid.String(),
					Email:        "admin-user@example.com",
					Name:         "admin-user",
					MobilePhone:  "01011112222",
					Roles:        []string{"ROLE_INVESTOR", "ROLE_ADMIN"},
					IsSuperAdmin: false,
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u := &adminUserUsecase{
				idGenerator:               tt.idGenerator(),
				pwhelper:                  tt.pwhelper(),
				adminUserRepository:       tt.adminUserRepository(),
				roleRepository:            tt.roleRepository(),
				operatorHistoryRepository: tt.operatorHistoryRepository(),
			}

			ignoreTime := cmpopts.IgnoreTypes(jsonutils.JSONTime{})

			got, err := u.Create(tt.args.ctx, tt.args.request)
			assert.Equalf(t, tt.wantErr, err != nil, "Create() error = %v, wantErr %v", err, tt.wantErr)
			assert.Equalf(t, cmp.Equal(tt.want, got, ignoreTime), true, "Create() = %+v, want %+v", got, tt.want)
		})
	}
}

func Test_adminUserUsecase_Update(t *testing.T) {
	type args struct {
		ctx     context.Context
		request adomain.AdminUserUpdateRequest
	}

	tests := []struct {
		name                      string
		args                      args
		pwhelper                  func() *mocks.PasswordHelper
		adminUserRepository       func() *mocks.AdminUserRepository
		operatorHistoryRepository func() *mocks.OperatorHistoryRepository
		roleRepository            func() *mocks.RoleRepository
		want                      adomain.AdminUserUpdateResponse
		wantErr                   bool
	}{
		{
			name: "(成功) 管理者修正",
			args: args{
				ctx: context.Background(),
				request: adomain.AdminUserUpdateRequest{
					JWTUser: jwthelper.User{
						Id:    mockSuperAdmin.Id,
						Email: mockSuperAdmin.Email,
						Name:  mockSuperAdmin.Name,
						Roles: map[string]int{"ROLE_ADMIN": 1, "ROLE_SUPER_ADMIN": 1},
					},
					ClientIp:          "clientIp",
					Id:                mockAdminUser.Id,
					Email:             "new-admin-user@example.com",
					Password:          lo.ToPtr("new-password1234"),
					Name:              "new-admin-user",
					MobilePhone:       "01033334444",
					ValidityStartDate: "2024-02-09 00:00:00",
					ValidityEndDate:   "2028-12-19 23:59:59",
				},
			},
			pwhelper: func() *mocks.PasswordHelper {
				helper := &mocks.PasswordHelper{}
				helper.On("GenerateSalt").Return("salt")
				helper.On("EncodePassword", "new-password1234", "salt").Return("encodedPassword")
				return helper
			},
			adminUserRepository: func() *mocks.AdminUserRepository {
				repo := &mocks.AdminUserRepository{}
				repo.On("GetById", mock.Anything, mockAdminUser.Id).Return(mockAdminUser, nil)
				repo.On("GetByEmail", mock.Anything, "new-admin-user@example.com").Return(domain.AdminUser{}, cdomain.ErrItemNotFound)
				repo.On("Update", mock.Anything, mock.MatchedBy(func(au domain.AdminUser) bool {
					ignoreTypes := cmpopts.IgnoreTypes(map[string]int{}, jsonutils.JSONTime{})
					ignoreFields := cmpopts.IgnoreFields(domain.AdminUser{}, "CommonId", "LoginFailCnt", "IsSuperAdmin", "LastLoginIp", "LastLoginAgent", "RolesString")
					return cmp.Equal(au, domain.AdminUser{
						Id:          mockAdminUser.Id,
						Email:       "new-admin-user@example.com",
						Password:    "encodedPassword",
						Name:        "new-admin-user",
						Salt:        "salt",
						MobilePhone: "01033334444",
						Roles:       []string{"ROLE_ADMIN"},
					}, ignoreTypes, ignoreFields)
				})).Return(nil)
				repo.On("Update", mock.Anything, mock.Anything).Return(nil)
				return repo
			},
			operatorHistoryRepository: func() *mocks.OperatorHistoryRepository {
				repo := &mocks.OperatorHistoryRepository{}
				repo.On("Create", mock.Anything, mock.MatchedBy(func(op domain.OperatorHistory) bool {
					ignoreTypes := cmpopts.IgnoreTypes(map[string]int{}, jsonutils.JSONTime{})
					return cmp.Equal(op, domain.OperatorHistory{
						UserId: mockSuperAdmin.Id,
						Name:   mockSuperAdmin.Name,
						Email:  mockSuperAdmin.Email,
						Menu:   MENU_ADMIN_USER_DETAIL,
						Ip:     "clientIp",
						State:  domain.CHANGE,
					}, ignoreTypes)
				})).Return(nil)
				return repo
			},
			roleRepository: func() *mocks.RoleRepository {
				repo := &mocks.RoleRepository{}
				repo.On("DeleteByUserId", mock.Anything, mockAdminUser.Id).Return(nil)
				repo.On("Create", mock.Anything, domain.Role{
					UserId:   mockAdminUser.Id,
					RoleCode: string(domain.ROLE_ADMIN),
				}).Return(int64(1), nil)
				return repo
			},
			want: adomain.AdminUserUpdateResponse{
				AdminUserItem: adomain.AdminUserItem{
					Id:          mockAdminUser.Id,
					CommonId:    mockAdminUser.CommonId,
					Email:       "new-admin-user@example.com",
					Name:        "new-admin-user",
					MobilePhone: "01033334444",
					Roles:       []string{"ROLE_ADMIN"},
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u := &adminUserUsecase{
				pwhelper:                  tt.pwhelper(),
				adminUserRepository:       tt.adminUserRepository(),
				roleRepository:            tt.roleRepository(),
				operatorHistoryRepository: tt.operatorHistoryRepository(),
			}

			ignoreTime := cmpopts.IgnoreTypes(jsonutils.JSONTime{})

			got, err := u.Update(tt.args.ctx, tt.args.request)
			assert.Equalf(t, tt.wantErr, err != nil, "Update() error = %v, wantErr %v", err, tt.wantErr)
			assert.Equalf(t, cmp.Equal(tt.want, got, ignoreTime), true, "Update() = %+v, want %+v", got, tt.want)
		})
	}
}

func Test_adminUserUsecase_DownloadHistories(t *testing.T) {
	type args struct {
		ctx     context.Context
		request adomain.AdminUserDownloadHistoriesRequest
	}

	tests := []struct {
		name                      string
		args                      args
		adminUserRepository       func() *mocks.AdminUserRepository
		operatorHistoryRepository func() *mocks.OperatorHistoryRepository
		adminUserService          func() *mocks.AdminUserService
		want                      adomain.AdminUserDownloadHistoriesResponse
		wantErr                   bool
	}{
		{
			name: "(成功) 管理者履歴ダウンロード",
			args: args{
				ctx: context.Background(),
				request: adomain.AdminUserDownloadHistoriesRequest{
					JWTUser: jwthelper.User{
						Id:    mockSuperAdmin.Id,
						Email: mockSuperAdmin.Email,
						Name:  mockSuperAdmin.Name,
						Roles: map[string]int{"ROLE_ADMIN": 1, "ROLE_SUPER_ADMIN": 1},
					},
					ClientIp: "clientIp",
					SearchInput: cdomain.SearchInput{
						SkipCount:      0,
						MaxResultCount: 10,
						Order:          "asc",
						Sortby:         "id",
					},
				},
			},
			adminUserRepository: func() *mocks.AdminUserRepository {
				repo := &mocks.AdminUserRepository{}
				repo.On("GetAll", mock.Anything,
					mock.Anything, // name
					mock.Anything, // mobilePhone
					mock.MatchedBy(func(field func(*cdomain.AdminUserQueryOptionFields)) bool {
						opt := cdomain.AdminUserQueryOptionFunc()
						field(&opt)
						return opt.SortBy() == "id"
					}), // SortBy
					mock.MatchedBy(func(field func(*cdomain.AdminUserQueryOptionFields)) bool {
						opt := cdomain.AdminUserQueryOptionFunc()
						field(&opt)
						return opt.Order() == "asc"
					}), // Order
					mock.MatchedBy(func(field func(*cdomain.AdminUserQueryOptionFields)) bool {
						opt := cdomain.AdminUserQueryOptionFunc()
						field(&opt)
						return opt.Offset() == 0
					}), // SkipCount
					mock.MatchedBy(func(field func(*cdomain.AdminUserQueryOptionFields)) bool {
						opt := cdomain.AdminUserQueryOptionFunc()
						field(&opt)
						return opt.Limit() == 10
					}), // MaxResultCount
				).Return(int64(1), []domain.AdminUser{
					mockAdminUser,
				}, nil)
				return repo
			},
			operatorHistoryRepository: func() *mocks.OperatorHistoryRepository {
				repo := &mocks.OperatorHistoryRepository{}
				repo.On("Create", mock.Anything, mock.MatchedBy(func(op domain.OperatorHistory) bool {
					ignoreTypes := cmpopts.IgnoreTypes(map[string]int{}, jsonutils.JSONTime{})
					return cmp.Equal(op, domain.OperatorHistory{
						UserId: mockSuperAdmin.Id,
						Name:   mockSuperAdmin.Name,
						Email:  mockSuperAdmin.Email,
						Menu:   MENU_ADMIN_USER_DOWNLOAD,
						Ip:     "clientIp",
						State:  domain.DOWNLOAD,
					}, ignoreTypes)
				})).Return(nil)
				return repo
			},
			adminUserService: func() *mocks.AdminUserService {
				service := &mocks.AdminUserService{}
				service.On("ExportCSV", mock.Anything, []domain.AdminUser{
					mockAdminUser,
				}).Return(cdomain.CSVData{}, nil)
				return service
			},
			want:    adomain.AdminUserDownloadHistoriesResponse{},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u := &adminUserUsecase{
				adminUserRepository:       tt.adminUserRepository(),
				operatorHistoryRepository: tt.operatorHistoryRepository(),
				adminUserService:          tt.adminUserService(),
			}

			_, err := u.DownloadHistories(tt.args.ctx, tt.args.request)
			assert.Equalf(t, tt.wantErr, err != nil, "DownloadHistories() error = %v, wantErr %v", err, tt.wantErr)
		})
	}
}

func Test_adminUserUsecase_Delete(t *testing.T) {
	type args struct {
		ctx     context.Context
		request adomain.AdminUserDeleteRequest
	}

	tests := []struct {
		name                      string
		args                      args
		adminUserRepository       func() *mocks.AdminUserRepository
		roleRepository            func() *mocks.RoleRepository
		operatorHistoryRepository func() *mocks.OperatorHistoryRepository
		want                      adomain.AdminUserDeleteResponse
		wantErr                   bool
	}{
		{
			name: "(成功) 管理者削除",
			args: args{
				ctx: context.Background(),
				request: adomain.AdminUserDeleteRequest{
					JWTUser: jwthelper.User{
						Id:    mockSuperAdmin.Id,
						Email: mockSuperAdmin.Email,
						Name:  mockSuperAdmin.Name,
						Roles: map[string]int{"ROLE_ADMIN": 1, "ROLE_SUPER_ADMIN": 1},
					},
					ClientIp: "clientIp",
					Id:       mockAdminUser.Id,
				},
			},
			adminUserRepository: func() *mocks.AdminUserRepository {
				repo := &mocks.AdminUserRepository{}
				repo.On("GetById", mock.Anything, mockAdminUser.Id).Return(mockAdminUser, nil)
				repo.On("Delete", mock.Anything, mockAdminUser.Id).Return(nil)
				return repo
			},
			roleRepository: func() *mocks.RoleRepository {
				repo := &mocks.RoleRepository{}
				repo.On("DeleteByUserId", mock.Anything, mockAdminUser.Id).Return(nil)
				return repo
			},
			operatorHistoryRepository: func() *mocks.OperatorHistoryRepository {
				repo := &mocks.OperatorHistoryRepository{}
				repo.On("Create", mock.Anything, domain.OperatorHistory{
					UserId: mockSuperAdmin.Id,
					Name:   mockSuperAdmin.Name,
					Email:  mockSuperAdmin.Email,
					Title:  mockAdminUser.Email + " (" + mockAdminUser.Name + ")",
					Menu:   MENU_ADMIN_USER_DETAIL,
					Ip:     "clientIp",
					State:  domain.DELETE,
				}).Return(nil)
				return repo
			},
			want:    adomain.AdminUserDeleteResponse{},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u := &adminUserUsecase{
				adminUserRepository:       tt.adminUserRepository(),
				roleRepository:            tt.roleRepository(),
				operatorHistoryRepository: tt.operatorHistoryRepository(),
			}

			_, err := u.Delete(tt.args.ctx, tt.args.request)
			assert.Equalf(t, tt.wantErr, err != nil, "Delete() error = %v, wantErr %v", err, tt.wantErr)
		})
	}
}

func Test_adminUserUsecase_Login(t *testing.T) {
	type args struct {
		ctx     context.Context
		request adomain.AdminUserLoginRequest
	}

	tests := []struct {
		name                      string
		args                      args
		s3pkg                     func() *mocks.S3
		pwhelper                  func() *mocks.PasswordHelper
		adminUserRepository       func() *mocks.AdminUserRepository
		operatorHistoryRepository func() *mocks.OperatorHistoryRepository
		adminUserService          func() *mocks.AdminUserService
		authService               func() *mocks.AuthService
		want                      adomain.AdminUserLoginResponse
		wantErr                   bool
	}{
		{
			name: "(成功) 管理者ログイン",
			args: args{
				ctx: context.Background(),
				request: adomain.AdminUserLoginRequest{
					Email:     mockSuperAdmin.Email,
					Password:  "password1234",
					ClientIp:  "clientIp",
					UserAgent: "userAgent",
				},
			},
			s3pkg: func() *mocks.S3 {
				s3 := &mocks.S3{}
				s3.On("CreateCookieSign").Return([]*http.Cookie{}, nil)
				return s3
			},
			pwhelper: func() *mocks.PasswordHelper {
				helper := &mocks.PasswordHelper{}
				helper.On("EncodePassword", "password1234", mockSuperAdmin.Salt).Return(mockSuperAdmin.Password)
				helper.On("IsSamePassword", mockSuperAdmin.Password, mockSuperAdmin.Password).Return(true)
				return helper
			},
			adminUserRepository: func() *mocks.AdminUserRepository {
				repo := &mocks.AdminUserRepository{}
				repo.On("GetByEmail", mock.Anything, mockSuperAdmin.Email).Return(mockSuperAdmin, nil)
				repo.On("Update", mock.Anything, mock.MatchedBy(func(au domain.AdminUser) bool {
					ignoreTypes := cmpopts.IgnoreTypes(map[string]int{}, jsonutils.JSONTime{})
					ignoreFields := cmpopts.IgnoreFields(domain.AdminUser{}, "CommonId", "IsSuperAdmin", "LastLoginAgent", "RolesString")
					return cmp.Equal(au, domain.AdminUser{
						Id:           mockSuperAdmin.Id,
						Email:        mockSuperAdmin.Email,
						Password:     mockSuperAdmin.Password,
						Salt:         mockSuperAdmin.Salt,
						Name:         mockSuperAdmin.Name,
						MobilePhone:  mockSuperAdmin.MobilePhone,
						Roles:        mockSuperAdmin.Roles,
						LoginFailCnt: 0,
					}, ignoreTypes, ignoreFields)
				})).Return(nil)
				return repo
			},
			operatorHistoryRepository: func() *mocks.OperatorHistoryRepository {
				repo := &mocks.OperatorHistoryRepository{}
				repo.On("Create", mock.Anything, mock.MatchedBy(func(op domain.OperatorHistory) bool {
					ignoreTypes := cmpopts.IgnoreTypes(map[string]int{}, jsonutils.JSONTime{})
					return cmp.Equal(op, domain.OperatorHistory{
						UserId: mockSuperAdmin.Id,
						Name:   mockSuperAdmin.Name,
						Email:  mockSuperAdmin.Email,
						Menu:   "",
						Ip:     "clientIp",
						Title:  "",
						State:  domain.LOGIN,
					}, ignoreTypes)
				})).Return(nil)
				return repo
			},
			adminUserService: func() *mocks.AdminUserService {
				service := &mocks.AdminUserService{}
				service.On("IsValidPeriod", mock.Anything, mockSuperAdmin).Return(true, nil)
				return service
			},
			authService: func() *mocks.AuthService {
				service := &mocks.AuthService{}
				service.On("CreateAdminAccessToken", mock.Anything, mockSuperAdmin).Return("accessToken", nil)
				service.On("CreateAdminRefreshToken", mock.Anything, cdomain.AdminRefreshTokenInput{
					Admin: mockSuperAdmin,
					IP:    "clientIp",
					Agent: "userAgent",
				}).Return("refreshToken", nil)
				return service
			},
			want: adomain.AdminUserLoginResponse{
				AccessToken:  "accessToken",
				RefreshToken: "refreshToken",
				Cookies:      []*http.Cookie{},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u := &adminUserUsecase{
				s3pkg:                     tt.s3pkg(),
				pwhelper:                  tt.pwhelper(),
				adminUserRepository:       tt.adminUserRepository(),
				operatorHistoryRepository: tt.operatorHistoryRepository(),
				adminUserService:          tt.adminUserService(),
				authService:               tt.authService(),
			}

			got, err := u.Login(tt.args.ctx, tt.args.request)
			assert.Equalf(t, tt.wantErr, err != nil, "Login() error = %v, wantErr %v", err, tt.wantErr)
			assert.Equalf(t, tt.want, got, "Login() = %+v, want %+v", got, tt.want)
		})
	}
}
func Test_adminUserUsecase_Logout(t *testing.T) {
	type args struct {
		ctx     context.Context
		request adomain.AdminUserLogoutRequest
	}

	tests := []struct {
		name                      string
		args                      args
		refreshTokenRepository    func() *mocks.RefreshTokenRepository
		operatorHistoryRepository func() *mocks.OperatorHistoryRepository
		want                      adomain.AdminUserLogoutResponse
		wantErr                   bool
	}{
		{
			name: "(成功) 管理者ログアウト",
			args: args{
				ctx: context.Background(),
				request: adomain.AdminUserLogoutRequest{
					JWTUser: jwthelper.User{
						Id:    mockSuperAdmin.Id,
						Email: mockSuperAdmin.Email,
						Name:  mockSuperAdmin.Name,
						Roles: map[string]int{"ROLE_ADMIN": 1, "ROLE_SUPER_ADMIN": 1},
					},
					ClientIp:     "clientIp",
					UserAgent:    "userAgent",
					RefreshToken: "refreshToken",
				},
			},
			refreshTokenRepository: func() *mocks.RefreshTokenRepository {
				repo := &mocks.RefreshTokenRepository{}
				repo.On("GetByToken", mock.Anything, "refreshToken").Return(domain.RefreshToken{
					Id:       int64(1),
					UserId:   mockSuperAdmin.Id,
					Token:    "refreshToken",
					ExpireIn: int64(100000),
				}, nil)
				repo.On("Update", mock.Anything, domain.RefreshToken{
					Id:       int64(1),
					UserId:   mockSuperAdmin.Id,
					Token:    "refreshToken",
					ExpireIn: int64(0),
				}).Return(nil)
				return repo
			},
			operatorHistoryRepository: func() *mocks.OperatorHistoryRepository {
				repo := &mocks.OperatorHistoryRepository{}
				repo.On("Create", mock.Anything, mock.MatchedBy(func(op domain.OperatorHistory) bool {
					ignoreTypes := cmpopts.IgnoreTypes(map[string]int{}, jsonutils.JSONTime{})
					return cmp.Equal(op, domain.OperatorHistory{
						UserId: mockSuperAdmin.Id,
						Name:   mockSuperAdmin.Name,
						Email:  mockSuperAdmin.Email,
						Menu:   "",
						Ip:     "clientIp",
						Title:  "",
						State:  domain.LOGOUT,
					}, ignoreTypes)
				})).Return(nil)
				return repo
			},
			want:    adomain.AdminUserLogoutResponse{},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u := &adminUserUsecase{
				refreshTokenRepository:    tt.refreshTokenRepository(),
				operatorHistoryRepository: tt.operatorHistoryRepository(),
			}

			got, err := u.Logout(tt.args.ctx, tt.args.request)
			assert.Equalf(t, tt.wantErr, err != nil, "Logout() error = %v, wantErr %v", err, tt.wantErr)
			assert.Equalf(t, tt.want, got, "Logout() = %+v, want %+v", got, tt.want)
		})
	}
}
