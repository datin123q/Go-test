package user

import (
	"context"
	"fmt"
	"testing"
	"time"

	adomain "github.com/illuminarean/energy-share-core-api/domain"
	cdomain "github.com/illuminarean/energy-shares-backend-repo/clean-arch-domain"
	"github.com/illuminarean/energy-shares-backend-repo/clean-arch-domain/mocks"
	"github.com/illuminarean/energy-shares-backend-repo/domain"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/emailutils"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/jsonutils"
	"github.com/samber/lo"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

var userUsecaseTestJST = func() *time.Location {
	loc, err := time.LoadLocation("Asia/Tokyo")
	if err != nil {
		panic(err)
	}
	return loc
}()

func Test_userUsecase_GetSentMails(t *testing.T) {
	type args struct {
		ctx     context.Context
		request adomain.UserGetSentMailsRequest
	}
	tests := []struct {
		name           string
		args           args
		emailpkg       func() emailutils.EmailRequester
		userRepository func() cdomain.UserRepository
		want           adomain.MailGetAllResponse
		wantErr        bool
	}{
		{
			name: "특定のユーザーに送信されたメールの照会",
			args: args{
				ctx: context.Background(),
				request: adomain.UserGetSentMailsRequest{
					UserId: 146,
				},
			},
			userRepository: func() cdomain.UserRepository {
				repo := new(mocks.UserRepository)
				repo.On("GetById", context.Background(), int64(146)).Return(mockUser1, nil)
				return repo
			},

			emailpkg: func() emailutils.EmailRequester {
				emailRequester := new(mocks.EmailRequester)
				emailRequester.On("GetAllRequest", context.Background(),
					mock.AnythingOfType("emailutils.GetAllQuery"),
				).Return(emailutils.EmailListResponse{
					Content: []emailutils.Email{
						mockEmailResponse1, mockEmailResponse2,
					},
					TotalElements: 2,
				}, nil)
				return emailRequester
			},
			want: func() adomain.MailGetAllResponse {
				var response adomain.MailGetAllResponse
				_ = response.FromEmailListResponse(emailutils.EmailListResponse{
					Content: []emailutils.Email{
						mockEmailResponse1, mockEmailResponse2,
					},
					TotalElements: 2,
				})
				return response
			}(),
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			usecase := userUsecase{
				emailpkg:       tt.emailpkg(),
				userRepository: tt.userRepository(),
			}
			got, err := usecase.GetSentMails(tt.args.ctx, tt.args.request)

			assert.NoErrorf(t, err, "userUsecase.GetSentMails() error = %v", err)
			assert.Equalf(t, tt.want, got, "userUsecase.GetSentMails() = %v, want %v", got, tt.want)
		})
	}
}

func Test_userUsecase_DownloadUserByFilter(t *testing.T) {
	type args struct {
		ctx     context.Context
		request adomain.UserDownloadUserByFilterRequest
	}
	tests := []struct {
		name                    string
		args                    args
		compositeUserRepository func() cdomain.CompositeUserRepository
		want                    adomain.UserDownloadUserByFilterResponse
		wantErr                 bool
	}{
		{
			name: "ユーザーフィルターでユーザーをダウンロード",
			args: args{
				ctx: context.Background(),
				request: adomain.UserDownloadUserByFilterRequest{
					AuditStatuses: []domain.AUDIT_STATUS{domain.PROVISIONAL_AUDIT, domain.COMPLETED_AUDIT},
					TotalAsset: &adomain.NumberRange{
						Min: lo.ToPtr(int64(5000)),
						Max: lo.ToPtr(int64(10000)),
					},
					AllowsNewsletter: lo.ToPtr(true),
					LastLoginDate: &adomain.DateRange{
						StartDate: jsonutils.JSONTime(time.Date(2023, 12, 1, 0, 0, 0, 0, lo.Must(time.LoadLocation("Asia/Tokyo")))),
						EndDate:   jsonutils.JSONTime(time.Date(2023, 12, 31, 23, 59, 59, 0, lo.Must(time.LoadLocation("Asia/Tokyo")))),
					},
				},
			},
			compositeUserRepository: func() cdomain.CompositeUserRepository {
				repo := new(mocks.CompositeUserRepository)
				repo.On("GetAll", context.Background(),
					mock.Anything, // Receiver
					mock.MatchedBy(func(field func(fields *cdomain.CompositeUserQueryOptionFields)) bool {
						opt := cdomain.CompositeUserQueryOptionFunc()
						field(&opt)

						want := []string{"PROVISIONAL_AUDIT", "COMPLETED_AUDIT"}
						got := lo.Map(opt.AuditStatuses(), func(status interface{}, _ int) string {
							return fmt.Sprint(status) // statusを文字列に変換
						})

						return lo.EveryBy(want, func(status string) bool {
							return lo.Contains(got, status)
						})
					}), // AuditStatuses
					mock.Anything, // UserTypes
					mock.Anything, // Nationalities
					mock.Anything, // Genders
					mock.Anything, // Jobs
					mock.Anything, // Assets
					mock.Anything, // AnnualIncomes
					mock.MatchedBy(func(field func(fields *cdomain.CompositeUserQueryOptionFields)) bool {
						opt := cdomain.CompositeUserQueryOptionFunc()
						field(&opt)
						return *opt.AllowsNewsletter() == true
					}), // AllowsNewsletter
					mock.Anything, // HasMyNumber
					mock.Anything, // MinSignUpRequestDate
					mock.Anything, // MaxSignUpRequestDate
					mock.Anything, // MinSignUpApprovalDate
					mock.Anything, // MaxSignUpApprovalDate
					mock.MatchedBy(func(field func(fields *cdomain.CompositeUserQueryOptionFields)) bool {
						opt := cdomain.CompositeUserQueryOptionFunc()
						field(&opt)
						return lo.FromPtr(opt.MinLastLoginDate()) == "2023-12-01 00:00:00"
					}), // MinLastLoginDate
					mock.MatchedBy(func(field func(fields *cdomain.CompositeUserQueryOptionFields)) bool {
						opt := cdomain.CompositeUserQueryOptionFunc()
						field(&opt)
						return lo.FromPtr(opt.MaxLastLoginDate()) == "2023-12-31 23:59:59"
					}), // MaxLastLoginDate
					mock.Anything, // MinBirthday
					mock.Anything, // MaxBirthday
					mock.MatchedBy(func(field func(fields *cdomain.CompositeUserQueryOptionFields)) bool {
						opt := cdomain.CompositeUserQueryOptionFunc()
						field(&opt)
						return *opt.MinTotalAsset() == int64(5000)
					}), // MinTotalAsset
					mock.MatchedBy(func(field func(fields *cdomain.CompositeUserQueryOptionFields)) bool {
						opt := cdomain.CompositeUserQueryOptionFunc()
						field(&opt)
						return *opt.MaxTotalAsset() == int64(10000)
					}), // MaxTotalAsset
					mock.Anything, // MinInvestOperationAmount
					mock.Anything, // MaxInvestOperationAmount
					mock.Anything, // MinDeposit
					mock.Anything, // MaxDeposit
					mock.Anything, // MinInvestSuccessAmount
					mock.Anything, // MaxInvestSuccessAmount
					mock.Anything, // MinInvestRequestAmount
					mock.Anything, // MaxInvestRequestAmount
					mock.Anything, // MinUnpaidAmount
					mock.Anything, // MaxUnpaidAmount
				).Return(int64(1),
					[]cdomain.CompositeUser{
						{
							Id:                    1,
							Email:                 "user1@test.com",
							Name:                  "user1",
							AuditStatus:           lo.ToPtr(domain.COMPLETED_AUDIT),
							UserType:              lo.ToPtr(domain.PRIVATE),
							Nationality:           lo.ToPtr(domain.JP),
							Gender:                lo.ToPtr(domain.MALE),
							Job:                   lo.ToPtr(domain.USER_UNLISTED_COMPANY_EMPLOYEE),
							Asset:                 lo.ToPtr(domain.ASSET5),
							AnnualIncome:          lo.ToPtr(domain.ANNUALINCOME5),
							AllowsNewsletter:      lo.ToPtr(true),
							MyNumberStatus:        lo.ToPtr(domain.APPROVE),
							Birthday:              lo.ToPtr(jsonutils.JSONTime(time.Date(1990, 12, 19, 0, 0, 0, 0, userUsecaseTestJST))),
							SignUpRequestDate:     lo.ToPtr(jsonutils.JSONTime(time.Date(2024, 3, 1, 0, 0, 0, 0, userUsecaseTestJST))),
							SignUpApprovalDate:    lo.ToPtr(jsonutils.JSONTime(time.Date(2024, 3, 1, 0, 0, 0, 0, userUsecaseTestJST))),
							LastLoginDate:         lo.ToPtr(jsonutils.JSONTime(time.Date(2024, 10, 15, 0, 0, 0, 0, userUsecaseTestJST))),
							TotalAsset:            lo.ToPtr(int64(7500)),
							InvestOperationAmount: lo.ToPtr(int64(1000)),
							Deposit:               lo.ToPtr(int64(3000)),
							InvestSuccessAmount:   lo.ToPtr(int64(2000)),
							InvestRequestAmount:   lo.ToPtr(int64(1000)),
							UnpaidAmount:          lo.ToPtr(int64(1000)),
						},
					}, nil)
				return repo
			},
			want:    adomain.UserDownloadUserByFilterResponse{},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			usecase := userUsecase{
				compositeUserRepository: tt.compositeUserRepository(),
			}
			_, err := usecase.DownloadUserByFilter(tt.args.ctx, tt.args.request)

			assert.NoErrorf(t, err, "userUsecase.DownloadUserByFilter() error = %v", err)
		})
	}
}
