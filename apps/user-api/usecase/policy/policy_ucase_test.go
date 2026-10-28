package policy

import (
	"testing"
	"time"

	cdomain "github.com/illuminarean/energy-shares-backend-repo/clean-arch-domain"
	"github.com/illuminarean/energy-shares-backend-repo/clean-arch-domain/mocks"
	"github.com/illuminarean/energy-shares-backend-repo/domain"
	udomain "github.com/illuminarean/energyshares-user-api/domain"
	"github.com/samber/lo"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"golang.org/x/net/context"
)

func Test_policyUsecase_GetAll(t *testing.T) {
	type args struct {
		ctx     context.Context
		request udomain.PolicyGetAllRequest
	}

	tests := []struct {
		name                 string
		args                 args
		policyRepository     func() *mocks.PolicyRepository
		attachmentRepository func() *mocks.AttachmentRepository
		want                 udomain.PolicyGetAllResponse
		wantErr              bool
	}{
		{
			name: "(成功) 約款全体照会",
			args: args{
				ctx: context.Background(),
				request: udomain.PolicyGetAllRequest{
					PolicyGetAllQueryParam: udomain.PolicyGetAllQueryParam{
						SearchInput: cdomain.SearchInput{
							Sortby:         "id",
							Order:          "asc",
							SkipCount:      0,
							MaxResultCount: 10,
						},
						PolicyViewType: lo.ToPtr(string(domain.POLICY_VIEW_FOOTER)),
					},
				},
			},
			policyRepository: func() *mocks.PolicyRepository {
				repo := &mocks.PolicyRepository{}
				repo.On("GetAll", context.Background(),
					mock.MatchedBy(func(field func(*cdomain.PolicyQueryOptionFields)) bool {
						opt := cdomain.PolicyQueryOptionFunc()
						field(&opt)
						return opt.SortBy() == "id"
					}), // SortBy
					mock.MatchedBy(func(field func(*cdomain.PolicyQueryOptionFields)) bool {
						opt := cdomain.PolicyQueryOptionFunc()
						field(&opt)
						return opt.Order() == "asc"
					}), // Order
					mock.MatchedBy(func(field func(*cdomain.PolicyQueryOptionFields)) bool {
						opt := cdomain.PolicyQueryOptionFunc()
						field(&opt)
						return opt.Offset() == 0
					}), // SkipCount
					mock.MatchedBy(func(field func(*cdomain.PolicyQueryOptionFields)) bool {
						opt := cdomain.PolicyQueryOptionFunc()
						field(&opt)
						return opt.Limit() == 10
					}), // MaxResultCount
					mock.MatchedBy(func(field func(*cdomain.PolicyQueryOptionFields)) bool {
						opt := cdomain.PolicyQueryOptionFunc()
						field(&opt)
						return opt.PolicyType() == nil
					}),
					mock.MatchedBy(func(field func(*cdomain.PolicyQueryOptionFields)) bool {
						opt := cdomain.PolicyQueryOptionFunc()
						field(&opt)
						return *opt.PolicyViewType() == string(domain.POLICY_VIEW_FOOTER)
					}),
				).Return(int64(2), []domain.Policy{mockPolicy1, mockPolicy2}, nil)
				return repo
			},
			attachmentRepository: func() *mocks.AttachmentRepository {
				repo := &mocks.AttachmentRepository{}

				repo.On("GetAll", context.Background(),
					mock.MatchedBy(func(field func(*cdomain.AttachmentQueryOptionFields)) bool {
						opt := cdomain.AttachmentQueryOptionFunc()
						field(&opt)
						return *opt.ResourceId() == mockPolicy1.CommonId
					}),
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
				).Return([]domain.Attachment{mockAttachment1}, nil)

				repo.On("GetAll", context.Background(),
					mock.MatchedBy(func(field func(*cdomain.AttachmentQueryOptionFields)) bool {
						opt := cdomain.AttachmentQueryOptionFunc()
						field(&opt)
						return *opt.ResourceId() == mockPolicy2.CommonId
					}),
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
				).Return([]domain.Attachment{}, nil)

				return repo
			},
			want: udomain.PolicyGetAllResponse{
				TotalCount: int64(2),
				Items: []udomain.PolicyGetAllItem{
					{
						Id:             mockPolicy1.Id,
						PolicyViewType: mockPolicy1.PolicyViewType,
						PolicyType:     mockPolicy1.PolicyType,
						ContentHtml:    mockPolicy1.ContentHtml,
						Attachment:     &mockAttachment1,
						Created:        mockPolicy1.Created,
						Updated:        mockPolicy1.Updated,
					},
					{
						Id:             mockPolicy2.Id,
						PolicyViewType: mockPolicy2.PolicyViewType,
						PolicyType:     mockPolicy2.PolicyType,
						ContentHtml:    mockPolicy2.ContentHtml,
						Attachment:     nil,
						Created:        mockPolicy2.Created,
						Updated:        mockPolicy2.Updated,
					},
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &policyUsecase{
				policyRepository:     tt.policyRepository(),
				attachmentRepository: tt.attachmentRepository(),
			}
			got, err := p.GetAll(tt.args.ctx, tt.args.request)
			assert.Equalf(t, tt.wantErr, err != nil, "policyUsecase.GetAll() error = %v, wantErr %v", err, tt.wantErr)
			assert.Equalf(t, tt.want, got, "policyUsecase.GetAll() = %v, want %v", got, tt.want)

			mockAttachmentRepo := p.attachmentRepository.(*mocks.AttachmentRepository)
			mockAttachmentRepo.AssertNumberOfCalls(t, "GetAll", 2)
		})
	}
}

func Test_policyUsecase_GetAllHistory(t *testing.T) {
	type args struct {
		ctx     context.Context
		request udomain.PolicyGetAllHistoryRequest
	}

	tests := []struct {
		name                    string
		args                    args
		policyRepository        func() *mocks.PolicyRepository
		policyHistoryRepository func() *mocks.PolicyHistoryRepository
		want                    udomain.PolicyGetAllHistoryResponse
		wantErr                 bool
	}{
		{
			name: "(成功) 約款の履歴照会",
			args: args{
				ctx: context.Background(),
				request: udomain.PolicyGetAllHistoryRequest{
					PolicyGetAllHistoryQueryParam: udomain.PolicyGetAllHistoryQueryParam{
						PolicyViewType: lo.ToPtr(string(domain.POLICY_VIEW_FOOTER)),
						PolicyType:     lo.ToPtr(string(domain.POLICY_CAUTION)),
						SearchInput: cdomain.SearchInput{
							Sortby:         "id",
							Order:          "desc",
							SkipCount:      0,
							MaxResultCount: 10,
						},
					},
				},
			},
			policyRepository: func() *mocks.PolicyRepository {
				repo := &mocks.PolicyRepository{}
				repo.On("GetAll", context.Background(),
					mock.MatchedBy(func(field func(*cdomain.PolicyQueryOptionFields)) bool {
						opt := cdomain.PolicyQueryOptionFunc()
						field(&opt)
						return opt.SortBy() == "id"
					}), // SortBy
					mock.MatchedBy(func(field func(*cdomain.PolicyQueryOptionFields)) bool {
						opt := cdomain.PolicyQueryOptionFunc()
						field(&opt)
						return opt.Order() == "desc"
					}), // Order
					mock.MatchedBy(func(field func(*cdomain.PolicyQueryOptionFields)) bool {
						opt := cdomain.PolicyQueryOptionFunc()
						field(&opt)
						return opt.Offset() == 0
					}), // SkipCount
					mock.MatchedBy(func(field func(*cdomain.PolicyQueryOptionFields)) bool {
						opt := cdomain.PolicyQueryOptionFunc()
						field(&opt)
						return opt.Limit() == 1
					}), // MaxResultCount
					mock.MatchedBy(func(field func(*cdomain.PolicyQueryOptionFields)) bool {
						opt := cdomain.PolicyQueryOptionFunc()
						field(&opt)
						return *opt.PolicyType() == string(domain.POLICY_CAUTION)
					}),
					mock.MatchedBy(func(field func(*cdomain.PolicyQueryOptionFields)) bool {
						opt := cdomain.PolicyQueryOptionFunc()
						field(&opt)
						return *opt.PolicyViewType() == string(domain.POLICY_VIEW_FOOTER)
					}),
				).Return(int64(1), []domain.Policy{mockPolicy1}, nil)
				return repo
			},
			policyHistoryRepository: func() *mocks.PolicyHistoryRepository {
				repo := &mocks.PolicyHistoryRepository{}
				repo.On("GetAll", context.Background(),
					mock.MatchedBy(func(field func(*cdomain.PolicyHistoryQueryOptionFields)) bool {
						opt := cdomain.PolicyHistoryQueryOptionFunc()
						field(&opt)
						return lo.FromPtr(opt.PolicyId()) == mockPolicy1.Id
					}), // policyId
					mock.MatchedBy(func(field func(*cdomain.PolicyHistoryQueryOptionFields)) bool {
						opt := cdomain.PolicyHistoryQueryOptionFunc()
						field(&opt)
						return lo.FromPtr(opt.IsPosted()) == true
					}), // IsPosted
					mock.MatchedBy(func(field func(*cdomain.PolicyHistoryQueryOptionFields)) bool {
						opt := cdomain.PolicyHistoryQueryOptionFunc()
						field(&opt)
						return opt.SortBy() == "id"
					}), // SortBy
					mock.MatchedBy(func(field func(*cdomain.PolicyHistoryQueryOptionFields)) bool {
						opt := cdomain.PolicyHistoryQueryOptionFunc()
						field(&opt)
						return opt.Order() == "desc"
					}), // Order
					mock.MatchedBy(func(field func(*cdomain.PolicyHistoryQueryOptionFields)) bool {
						opt := cdomain.PolicyHistoryQueryOptionFunc()
						field(&opt)
						return opt.Offset() == 0
					}), // SkipCount
					mock.MatchedBy(func(field func(*cdomain.PolicyHistoryQueryOptionFields)) bool {
						opt := cdomain.PolicyHistoryQueryOptionFunc()
						field(&opt)
						return opt.Limit() == 10
					}), // MaxResultCount
				).Return(int64(1), []domain.PolicyInfoHistory{mockPolicyHistory1}, nil)
				return repo
			},
			want: udomain.PolicyGetAllHistoryResponse{
				TotalCount: int64(1),
				Items: []udomain.PolicyGetAllHistoryItem{
					{
						Id:             mockPolicyHistory1.Id,
						IsPosted:       mockPolicyHistory1.IsPosted,
						Label:          lo.If(mockPolicyHistory1.Label.IsEmpty(), (*string)(nil)).Else(lo.ToPtr(mockPolicyHistory1.Label.ToFormat(time.DateOnly))),
						PolicyType:     domain.POLICY_CAUTION,
						PolicyViewType: domain.POLICY_VIEW_FOOTER,
						ContentHtml:    mockPolicyHistory1.CurrContentHtml,
						Created:        mockPolicyHistory1.Created,
					},
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &policyUsecase{
				policyRepository:        tt.policyRepository(),
				policyHistoryRepository: tt.policyHistoryRepository(),
			}
			got, err := p.GetAllHistory(tt.args.ctx, tt.args.request)
			assert.Equalf(t, tt.wantErr, err != nil, "policyUsecase.GetAllHistory() error = %v, wantErr %v", err, tt.wantErr)
			assert.Equalf(t, tt.want, got, "policyUsecase.GetAllHistory() = %v, want %v", got, tt.want)
		})
	}
}
