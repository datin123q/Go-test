package policy

import (
	"context"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	adomain "github.com/illuminarean/energy-share-core-api/domain"
	cdomain "github.com/illuminarean/energy-shares-backend-repo/clean-arch-domain"
	"github.com/illuminarean/energy-shares-backend-repo/clean-arch-domain/mocks"
	"github.com/illuminarean/energy-shares-backend-repo/domain"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/config"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/filehelper"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/jsonutils"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/s3utils"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/serviceutils"
	"github.com/samber/lo"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func Test_policyUsecase_Update(t *testing.T) {
	serviceutils.SetCDN("cdn.dev.energyshares.jp")
	t.Cleanup(func() {
		serviceutils.SetCDN("")
	})

	type args struct {
		ctx     context.Context
		request adomain.PolicyUpdateRequest
	}

	var newPolicy domain.Policy
	newPolicy.Id = mockPolicy1.Id
	newPolicy.PolicyViewType = domain.POLICY_VIEW_JOIN
	newPolicy.ContentHtml = "content"
	newPolicy.ModifierId = mockModifier.Id
	newPolicy.ModifierEmail = mockModifier.Email

	tests := []struct {
		name                 string
		args                 args
		s3pkg                func() *mocks.S3
		policyRepository     func() *mocks.PolicyRepository
		attachmentRepository func() *mocks.AttachmentRepository
		policyService        func() *mocks.PolicyService
		want                 adomain.PolicyUpdateResponse
		wantErr              bool
	}{
		{
			name: "(成功) 約款修正",
			args: args{
				ctx: context.Background(),
				request: adomain.PolicyUpdateRequest{
					JWTUser:        mockModifier,
					Id:             1,
					PolicyViewType: domain.POLICY_VIEW_JOIN,
					ContentHtml:    newPolicy.ContentHtml,
					AttachmentUrls: []string{mockNewAttachment.FileUrl},
				},
			},
			policyRepository: func() *mocks.PolicyRepository {
				repo := &mocks.PolicyRepository{}
				repo.On("GetById", context.Background(), mockPolicy1.Id).Return(mockPolicy1, nil)
				repo.On("Update", context.Background(), newPolicy).Return(nil)
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
				repo.On("Delete", context.Background(), mockPolicy1.Attachment.Id).Return(nil)
				repo.On("Create", context.Background(), mockNewAttachment).Return(int64(3), nil)
				mockNewAttachment.Id = int64(3)
				newPolicy.Attachment = &mockNewAttachment
				return repo
			},
			s3pkg: func() *mocks.S3 {
				s3 := new(mocks.S3)
				s3.On("GetHeadObject", "es-jp-devv2-public-file-storage", mockNewAttachment.FilePath+mockNewAttachment.OriginalName).Return(s3utils.S3ObjectInfo{
					FileSize:         1234,
					OriginalFileName: "original_file_name.jpg",
				}, nil)
				return s3
			},
			policyService: func() *mocks.PolicyService {
				service := new(mocks.PolicyService)
				service.On("CreatePolicyHistory", context.Background(), cdomain.PolicyHistoryInput{
					Before:   mockPolicy1,
					After:    newPolicy,
					Modifier: mockModifier,
				}).Return(nil)
				return service
			},
			want: adomain.PolicyUpdateResponse{
				Id:             newPolicy.Id,
				CommonId:       mockPolicy1.CommonId,
				OwnerId:        mockPolicy1.OwnerId,
				ModifierId:     newPolicy.ModifierId,
				PolicyViewType: newPolicy.PolicyViewType,
				PolicyType:     mockPolicy1.PolicyType,
				ContentHtml:    newPolicy.ContentHtml,
				ModifierEmail:  newPolicy.ModifierEmail,
				Attachment:     &mockNewAttachment,
				Created:        mockPolicy1.Created,
				Updated:        newPolicy.Updated,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &policyUsecase{
				cfg: &config.Config{
					CommonConfig: config.CommonConfig{
						AWS: struct {
							Region                  string
							SES_Region              string
							SNS_Region              string
							CDN                     string
							S3_PRIVATE_BUCKET       string
							S3_PUBLIC_BUCKET        string
							S3_EKYC_IMAGES_BUCKET   string
							SES_ARN                 string
							SNS_ARN                 string
							PUBLIC_KEY_ID           string
							PRIVATE_KEY_PATH        string
							BASE_ENDPOINT           string
							INVEST_WORKER_QUEUE_URL string
						}{
							S3_PUBLIC_BUCKET: "es-jp-devv2-public-file-storage",
							CDN:              "cdn.dev.energyshares.jp",
						},
					},
				},
				fileHelper:           filehelper.NewFileHelper(),
				policyRepository:     tt.policyRepository(),
				attachmentRepository: tt.attachmentRepository(),
				policyService:        tt.policyService(),
				s3pkg:                tt.s3pkg(),
			}
			got, err := p.Update(tt.args.ctx, tt.args.request)
			assert.Equalf(t, tt.wantErr, err != nil, "policyUsecase.Update() error = %v, wantErr %v", err, tt.wantErr)
			assert.Equalf(t, tt.want, got, "policyUsecase.Update() = %v, want %v", got, tt.want)
		})
	}
}

func Test_policyUsecase_GetAll(t *testing.T) {
	type args struct {
		ctx     context.Context
		request adomain.PolicyGetAllRequest
	}

	tests := []struct {
		name             string
		args             args
		policyRepository func() cdomain.PolicyRepository
		want             adomain.PolicyGetAllResponse
		wantErr          bool
	}{
		{
			name: "(成功) 約款全体照会",
			args: args{
				ctx: context.Background(),
				request: adomain.PolicyGetAllRequest{
					JWTUser: mockModifier,
					PolicyGetAllQueryParam: adomain.PolicyGetAllQueryParam{
						SearchInput: cdomain.SearchInput{
							Sortby:         "id",
							Order:          "desc",
							SkipCount:      0,
							MaxResultCount: 10,
						},
						PolicyType:     lo.ToPtr(string(domain.POLICY_CAUTION)),
						PolicyViewType: lo.ToPtr(string(domain.POLICY_VIEW_FOOTER)),
					},
				},
			},
			policyRepository: func() cdomain.PolicyRepository {
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
						return opt.Limit() == 10
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
				).Return(int64(1), []domain.Policy{mockPolicy2}, nil)
				return repo
			},
			want: adomain.PolicyGetAllResponse{
				TotalCount: int64(1),
				Items: []adomain.PolicyGetAllItem{
					{
						Id:             mockPolicy2.Id,
						CommonId:       mockPolicy2.CommonId,
						OwnerId:        mockPolicy2.OwnerId,
						ModifierId:     mockPolicy2.ModifierId,
						PolicyViewType: mockPolicy2.PolicyViewType,
						PolicyType:     mockPolicy2.PolicyType,
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
				policyRepository: tt.policyRepository(),
			}
			got, err := p.GetAll(tt.args.ctx, tt.args.request)
			assert.Equalf(t, tt.wantErr, err != nil, "policyUsecase.GetAll() error = %v, wantErr %v", err, tt.wantErr)
			assert.Equalf(t, tt.want, got, "policyUsecase.GetAll() = %v, want %v", got, tt.want)
		})
	}
}

func Test_policyUsecase_GetOne(t *testing.T) {
	type args struct {
		ctx     context.Context
		request adomain.PolicyGetOneRequest
	}

	tests := []struct {
		name                 string
		args                 args
		policyRepository     func() cdomain.PolicyRepository
		attachmentRepository func() cdomain.AttachmentRepository
		want                 adomain.PolicyGetOneResponse
		wantErr              bool
	}{
		{
			name: "(成功) 約款単件照会 - 添付ファイル有り",
			args: args{
				ctx: context.Background(),
				request: adomain.PolicyGetOneRequest{
					JWTUser: mockModifier,
					Id:      1,
				},
			},
			policyRepository: func() cdomain.PolicyRepository {
				repo := &mocks.PolicyRepository{}
				repo.On("GetById", context.Background(), mockPolicy1.Id).Return(mockPolicy1, nil)
				return repo
			},
			attachmentRepository: func() cdomain.AttachmentRepository {
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
				return repo
			},
			want: adomain.PolicyGetOneResponse{
				Id:             mockPolicy1.Id,
				CommonId:       mockPolicy1.CommonId,
				OwnerId:        mockPolicy1.OwnerId,
				ModifierId:     mockPolicy1.ModifierId,
				PolicyViewType: mockPolicy1.PolicyViewType,
				PolicyType:     mockPolicy1.PolicyType,
				ContentHtml:    mockPolicy1.ContentHtml,
				ModifierEmail:  mockPolicy1.ModifierEmail,
				Attachment:     &mockAttachment1,
				Created:        mockPolicy1.Created,
				Updated:        mockPolicy1.Updated,
			},
			wantErr: false,
		},
		{name: "(成功) 約款単件照会 - 添付ファイル無し",
			args: args{
				ctx: context.Background(),
				request: adomain.PolicyGetOneRequest{
					JWTUser: mockModifier,
					Id:      2,
				},
			},
			policyRepository: func() cdomain.PolicyRepository {
				repo := &mocks.PolicyRepository{}
				repo.On("GetById", context.Background(), mockPolicy2.Id).Return(mockPolicy2, nil)
				return repo
			},
			attachmentRepository: func() cdomain.AttachmentRepository {
				repo := &mocks.AttachmentRepository{}
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
			want: adomain.PolicyGetOneResponse{
				Id:             mockPolicy2.Id,
				CommonId:       mockPolicy2.CommonId,
				OwnerId:        mockPolicy2.OwnerId,
				ModifierId:     mockPolicy2.ModifierId,
				PolicyViewType: mockPolicy2.PolicyViewType,
				PolicyType:     mockPolicy2.PolicyType,
				ContentHtml:    mockPolicy2.ContentHtml,
				ModifierEmail:  mockPolicy2.ModifierEmail,
				Attachment:     nil,
				Created:        mockPolicy2.Created,
				Updated:        mockPolicy2.Updated,
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
			got, err := p.GetOne(tt.args.ctx, tt.args.request)
			assert.Equalf(t, tt.wantErr, err != nil, "policyUsecase.GetOne() error = %v, wantErr %v", err, tt.wantErr)
			assert.Equalf(t, tt.want, got, "policyUsecase.GetOne() = %v, want %v", got, tt.want)
		})
	}
}

func Test_policyUsecase_GetAllHistory(t *testing.T) {
	type args struct {
		ctx     context.Context
		request adomain.PolicyGetAllHistoryRequest
	}

	tests := []struct {
		name                    string
		args                    args
		policyHistoryRepository func() cdomain.PolicyHistoryRepository
		want                    adomain.PolicyGetAllHistoryResponse
		wantErr                 bool
	}{
		{
			name: "(成功) 約款の履歴照会",
			args: args{
				ctx: context.Background(),
				request: adomain.PolicyGetAllHistoryRequest{
					JWTUser: mockModifier,
					Id:      int64(1),
					PolicyGetAllHistoryQueryParam: adomain.PolicyGetAllHistoryQueryParam{
						SearchInput: cdomain.SearchInput{
							Sortby:         "id",
							Order:          "desc",
							SkipCount:      0,
							MaxResultCount: 10,
						},
						IsPosted: lo.ToPtr(true),
					},
				},
			},
			policyHistoryRepository: func() cdomain.PolicyHistoryRepository {
				repo := &mocks.PolicyHistoryRepository{}
				repo.On("GetAll", context.Background(),
					mock.MatchedBy(func(field func(*cdomain.PolicyHistoryQueryOptionFields)) bool {
						opt := cdomain.PolicyHistoryQueryOptionFunc()
						field(&opt)
						return lo.FromPtr(opt.PolicyId()) == int64(1)
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
			want: adomain.PolicyGetAllHistoryResponse{
				TotalCount: int64(1),
				Items: []adomain.PolicyHistoryItem{
					{
						Id:              mockPolicyHistory1.Id,
						PolicyId:        mockPolicyHistory1.PolicyId,
						SubstanceType:   mockPolicyHistory1.SubstanceType,
						ContentType:     mockPolicyHistory1.ContentType,
						Name:            mockPolicyHistory1.Name,
						Email:           mockPolicyHistory1.Email,
						IsPosted:        true,
						Label:           lo.If(mockPolicyHistory1.Label.IsEmpty(), (*string)(nil)).Else(lo.ToPtr(mockPolicyHistory1.Label.ToFormat(time.DateOnly))),
						PrevContent:     mockPolicyHistory1.PrevContent,
						PrevContentHtml: mockPolicyHistory1.PrevContentHtml,
						PrevUrl:         mockPolicyHistory1.PrevUrl,
						CurrContent:     mockPolicyHistory1.CurrContent,
						CurrContentHtml: mockPolicyHistory1.CurrContentHtml,
						CurrUrl:         mockPolicyHistory1.CurrUrl,
						Created:         mockPolicyHistory1.Created,
					},
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &policyUsecase{
				policyHistoryRepository: tt.policyHistoryRepository(),
			}
			got, err := p.GetAllHistory(tt.args.ctx, tt.args.request)
			assert.Equalf(t, tt.wantErr, err != nil, "policyUsecase.GetAllHistory() error = %v, wantErr %v", err, tt.wantErr)
			assert.Equalf(t, tt.want, got, "policyUsecase.GetAllHistory() = %v, want %v", got, tt.want)
		})
	}
}

func Test_policyUsecase_UpdateHistory(t *testing.T) {
	type args struct {
		ctx     context.Context
		request adomain.PolicyUpdateHistoryRequest
	}

	tests := []struct {
		name                    string
		args                    args
		policyHistoryRepository func() cdomain.PolicyHistoryRepository
		want                    adomain.PolicyUpdateHistoryResponse
		wantErr                 bool
	}{
		{
			name: "約款履歴修正成功 - 掲載○、ラベル有り",
			args: args{
				ctx: context.Background(),
				request: adomain.PolicyUpdateHistoryRequest{
					JWTUser:   mockModifier,
					Id:        1,
					HistoryId: 2,
					Label:     lo.ToPtr("2024-03-14"),
					IsPosted:  true,
				},
			},
			policyHistoryRepository: func() cdomain.PolicyHistoryRepository {
				repo := &mocks.PolicyHistoryRepository{}

				repo.On("GetAll", context.Background(),
					mock.MatchedBy(func(field func(*cdomain.PolicyHistoryQueryOptionFields)) bool {
						opt := cdomain.PolicyHistoryQueryOptionFunc()
						field(&opt)
						return lo.FromPtr(opt.PolicyId()) == int64(1)
					}), // policyId
					mock.MatchedBy(func(field func(*cdomain.PolicyHistoryQueryOptionFields)) bool {
						opt := cdomain.PolicyHistoryQueryOptionFunc()
						field(&opt)
						return lo.FromPtr(opt.IsPosted()) == true
					}), // IsPosted
					mock.MatchedBy(func(field func(*cdomain.PolicyHistoryQueryOptionFields)) bool {
						opt := cdomain.PolicyHistoryQueryOptionFunc()
						field(&opt)
						return lo.FromPtr(opt.Label()) == "2024-03-14"
					}), // label
				).Return(int64(1), []domain.PolicyInfoHistory{mockPolicyHistory3}, nil)

				mockPolicyHistory3.IsPosted = false
				repo.On("Update", context.Background(), mock.MatchedBy(func(history domain.PolicyInfoHistory) bool {
					ignoreFields := cmpopts.IgnoreFields(domain.PolicyInfoHistory{}, "Label", "Created")
					return cmp.Equal(mockPolicyHistory3, history, ignoreFields) &&
						mockPolicyHistory3.Label.ToFormat(time.DateOnly) == history.Label.ToFormat(time.DateOnly)
				})).Return(nil)

				repo.On("GetById", context.Background(), mockPolicyHistory2.Id).Return(mockPolicyHistory2, nil)

				newPolicyHistory := mockPolicyHistory2
				newPolicyHistory.Label = jsonutils.JSONTime(time.Date(2024, 3, 14, 0, 0, 0, 0, time.UTC))
				newPolicyHistory.IsPosted = true
				repo.On("Update", context.Background(), mock.MatchedBy(func(history domain.PolicyInfoHistory) bool {
					ignoreFields := cmpopts.IgnoreFields(domain.PolicyInfoHistory{}, "Label", "Created")
					return cmp.Equal(newPolicyHistory, history, ignoreFields) &&
						newPolicyHistory.Label.ToFormat(time.DateOnly) == history.Label.ToFormat(time.DateOnly)
				})).Return(nil)

				return repo
			},
			want: adomain.PolicyUpdateHistoryResponse{
				PolicyHistoryItem: adomain.PolicyHistoryItem{
					Id:              mockPolicyHistory2.Id,
					PolicyId:        mockPolicyHistory2.PolicyId,
					SubstanceType:   mockPolicyHistory2.SubstanceType,
					ContentType:     mockPolicyHistory2.ContentType,
					Name:            mockPolicyHistory2.Name,
					Email:           mockPolicyHistory2.Email,
					IsPosted:        true,
					Label:           lo.ToPtr("2024-03-14"),
					PrevContent:     mockPolicyHistory2.PrevContent,
					PrevContentHtml: mockPolicyHistory2.PrevContentHtml,
					PrevUrl:         mockPolicyHistory2.PrevUrl,
					CurrContent:     mockPolicyHistory2.CurrContent,
					CurrContentHtml: mockPolicyHistory2.CurrContentHtml,
					CurrUrl:         mockPolicyHistory2.CurrUrl,
					Created:         mockPolicyHistory2.Created,
				},
			},
			wantErr: false,
		},
		{
			name: "約款履歴修正成功 - 掲載✕、ラベル有り",
			args: args{
				ctx: context.Background(),
				request: adomain.PolicyUpdateHistoryRequest{
					JWTUser:   mockModifier,
					Id:        1,
					HistoryId: 2,
					Label:     lo.ToPtr("2024-02-09"),
					IsPosted:  false,
				},
			},
			policyHistoryRepository: func() cdomain.PolicyHistoryRepository {
				repo := &mocks.PolicyHistoryRepository{}

				repo.On("GetById", context.Background(), mockPolicyHistory2.Id).Return(mockPolicyHistory2, nil)

				newPolicyHistory := mockPolicyHistory2
				newPolicyHistory.Label = jsonutils.JSONTime(time.Date(2024, 2, 9, 0, 0, 0, 0, time.UTC))
				newPolicyHistory.IsPosted = false
				repo.On("Update", context.Background(), mock.MatchedBy(func(history domain.PolicyInfoHistory) bool {
					ignoreFields := cmpopts.IgnoreFields(domain.PolicyInfoHistory{}, "Label", "Created")
					return cmp.Equal(newPolicyHistory, history, ignoreFields) &&
						newPolicyHistory.Label.ToFormat(time.DateOnly) == history.Label.ToFormat(time.DateOnly)
				})).Return(nil)

				return repo
			},
			want: adomain.PolicyUpdateHistoryResponse{
				PolicyHistoryItem: adomain.PolicyHistoryItem{
					Id:              mockPolicyHistory2.Id,
					PolicyId:        mockPolicyHistory2.PolicyId,
					SubstanceType:   mockPolicyHistory2.SubstanceType,
					ContentType:     mockPolicyHistory2.ContentType,
					Name:            mockPolicyHistory2.Name,
					Email:           mockPolicyHistory2.Email,
					IsPosted:        false,
					Label:           lo.ToPtr("2024-02-09"),
					PrevContent:     mockPolicyHistory2.PrevContent,
					PrevContentHtml: mockPolicyHistory2.PrevContentHtml,
					PrevUrl:         mockPolicyHistory2.PrevUrl,
					CurrContent:     mockPolicyHistory2.CurrContent,
					CurrContentHtml: mockPolicyHistory2.CurrContentHtml,
					CurrUrl:         mockPolicyHistory2.CurrUrl,
					Created:         mockPolicyHistory2.Created,
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &policyUsecase{
				policyHistoryRepository: tt.policyHistoryRepository(),
			}

			ignoreFields := cmpopts.IgnoreFields(adomain.PolicyUpdateHistoryResponse{}, "Label", "Created")

			got, err := p.UpdateHistory(tt.args.ctx, tt.args.request)
			assert.Equalf(t, tt.wantErr, err != nil, "policyUsecase.UpdateHistory() error = %v, wantErr %v", err, tt.wantErr)
			assert.Equalf(t, cmp.Equal(tt.want, got, ignoreFields), true, "policyUsecase.UpdateHistory() = %v, want %v", got, tt.want)
			assert.Equal(t, lo.FromPtr(tt.want.Label), lo.FromPtr(got.Label))
		})
	}
}

func Test_policyUsecase_Delete(t *testing.T) {
	type args struct {
		ctx     context.Context
		request adomain.PolicyDeleteRequest
	}

	tests := []struct {
		name             string
		args             args
		policyRepository func() cdomain.PolicyRepository
		wantErr          bool
	}{
		{
			name: "(成功) 約款削除",
			args: args{
				ctx: context.Background(),
				request: adomain.PolicyDeleteRequest{
					JWTUser: mockModifier,
					Id:      1,
				},
			},
			policyRepository: func() cdomain.PolicyRepository {
				repo := &mocks.PolicyRepository{}
				repo.On("GetById", context.Background(), mockPolicy1.Id).Return(mockPolicy1, nil)
				newPolicy := mockPolicy1
				newPolicy.ModifierId = mockModifier.Id
				newPolicy.ModifierEmail = mockModifier.Email
				repo.On("Update", context.Background(), newPolicy).Return(nil)
				repo.On("Delete", context.Background(), mockPolicy1.Id).Return(nil)
				return repo
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &policyUsecase{
				policyRepository: tt.policyRepository(),
			}
			if err := p.Delete(tt.args.ctx, tt.args.request); (err != nil) != tt.wantErr {
				t.Errorf("policyUsecase.Delete() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_policyUsecase_CheckLabelDuplicate(t *testing.T) {
	type args struct {
		ctx     context.Context
		request adomain.PolicyCheckLabelDuplicateRequest
	}

	tests := []struct {
		name                    string
		args                    args
		policyHistoryRepository func() cdomain.PolicyHistoryRepository
		want                    adomain.PolicyCheckLabelDuplicateResponse
		wantErr                 bool
	}{
		{
			name: "約款ラベル重複確認正常応答",
			args: args{
				ctx: context.Background(),
				request: adomain.PolicyCheckLabelDuplicateRequest{
					JWTUser: mockModifier,
					Id:      1,
					Label:   lo.ToPtr("2024-02-09"),
				},
			},
			policyHistoryRepository: func() cdomain.PolicyHistoryRepository {
				repo := &mocks.PolicyHistoryRepository{}
				repo.On("GetAll", context.Background(),
					mock.MatchedBy(func(field func(*cdomain.PolicyHistoryQueryOptionFields)) bool {
						opt := cdomain.PolicyHistoryQueryOptionFunc()
						field(&opt)
						return lo.FromPtr(opt.PolicyId()) == int64(1)
					}), // policyId
					mock.MatchedBy(func(field func(*cdomain.PolicyHistoryQueryOptionFields)) bool {
						opt := cdomain.PolicyHistoryQueryOptionFunc()
						field(&opt)
						return lo.FromPtr(opt.IsPosted()) == true
					}), // IsPosted
					mock.MatchedBy(func(field func(*cdomain.PolicyHistoryQueryOptionFields)) bool {
						opt := cdomain.PolicyHistoryQueryOptionFunc()
						field(&opt)
						return lo.FromPtr(opt.Label()) == "2024-02-09"
					}), // label
				).Return(int64(1), []domain.PolicyInfoHistory{mockPolicyHistory1}, nil)
				return repo
			},
			want: adomain.PolicyCheckLabelDuplicateResponse{
				IsDuplicate: true,
				PolicyHistory: &adomain.PolicyHistoryItem{
					Id:              mockPolicyHistory1.Id,
					PolicyId:        mockPolicyHistory1.PolicyId,
					SubstanceType:   mockPolicyHistory1.SubstanceType,
					ContentType:     mockPolicyHistory1.ContentType,
					Name:            mockPolicyHistory1.Name,
					Email:           mockPolicyHistory1.Email,
					IsPosted:        true,
					Label:           lo.ToPtr(mockPolicyHistory1.Label.ToFormat(time.DateOnly)),
					PrevContent:     mockPolicyHistory1.PrevContent,
					PrevContentHtml: mockPolicyHistory1.PrevContentHtml,
					PrevUrl:         mockPolicyHistory1.PrevUrl,
					CurrContent:     mockPolicyHistory1.CurrContent,
					CurrContentHtml: mockPolicyHistory1.CurrContentHtml,
					CurrUrl:         mockPolicyHistory1.CurrUrl,
					Created:         mockPolicyHistory1.Created,
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &policyUsecase{
				policyHistoryRepository: tt.policyHistoryRepository(),
			}
			got, err := p.CheckLabelDuplicate(tt.args.ctx, tt.args.request)
			assert.Equalf(t, tt.wantErr, err != nil, "policyUsecase.CheckLabelDuplicate() error = %v, wantErr %v", err, tt.wantErr)
			assert.Equalf(t, tt.want, got, "policyUsecase.CheckLabelDuplicate() = %v, want %v", got, tt.want)
		})
	}
}
