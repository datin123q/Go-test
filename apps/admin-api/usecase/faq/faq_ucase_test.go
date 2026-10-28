package faq

import (
	"context"
	"testing"
	"time"

	"github.com/gofrs/uuid"

	"github.com/illuminarean/energy-shares-backend-repo/pkg/filehelper"

	"github.com/illuminarean/energy-shares-backend-repo/pkg/s3utils"

	"github.com/samber/lo"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	adomain "github.com/illuminarean/energy-share-core-api/domain"
	cdomain "github.com/illuminarean/energy-shares-backend-repo/clean-arch-domain"
	"github.com/illuminarean/energy-shares-backend-repo/clean-arch-domain/mocks"
	"github.com/illuminarean/energy-shares-backend-repo/domain"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/config"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/jwthelper"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/serviceutils"
)

func Test_faqUsecase_Create(t *testing.T) {
	serviceutils.SetCDN("cdn.dev.energyshares.jp")
	t.Cleanup(func() {
		serviceutils.SetCDN("")
	})

	uid, _ := uuid.FromString("bddad349-d0cc-4a1d-8843-431d54fcd2c7")

	type args struct {
		ctx     context.Context
		request adomain.FaqCreateRequest
	}
	tests := []struct {
		name                      string
		args                      args
		s3pkg                     func() *mocks.S3
		idGenerator               func() *mocks.IdGenerator
		faqRepository             func() *mocks.FaqRepository
		attachmentRepository      func() *mocks.AttachmentRepository
		contentImageRepository    func() *mocks.ContentImageRepository
		operatorHistoryRepository func() *mocks.OperatorHistoryRepository
		wantErr                   bool
	}{
		{
			name: "正常に呼び出された場合",
			args: args{
				ctx: context.Background(),
				request: adomain.FaqCreateRequest{
					Pin:              nil,
					ContentHtml:      "<p>Hello</p>",
					Title:            "重要なお知らせです。",
					Posted:           false,
					CategoryType:     string(domain.FAQ_CATEGORY_ACCOUNT_REGISTRATION),
					ContentImageUrls: []string{"https://es-jp-devv2-public-file-storage.s3.ap-northeast-2.amazonaws.com/content_img_file/2024-09-10/crg1ackldh5j4kkqiueg.jpg"},
					AttachmentUrls:   []string{"https://es-jp-devv2-public-file-storage.s3.ap-northeast-2.amazonaws.com/attachment_file/2024-09-09/crfbl6cldh5hujq0ulq0.jpg"},
					JWTUser: jwthelper.User{
						Id:    12378123,
						Name:  "ホン・ギルドン",
						Email: "gildong@email.com",
					},
					ClientIp: "127.0.0.1",
				},
			},
			idGenerator: func() *mocks.IdGenerator {
				gen := new(mocks.IdGenerator)
				gen.On("GenUuidV4").Return(uid)
				return gen
			},
			faqRepository: func() *mocks.FaqRepository {
				repo := new(mocks.FaqRepository)
				repo.On("Create", context.Background(), domain.Faq{
					Pin:           false,
					ContentHtml:   "<p>Hello</p>",
					Title:         "重要なお知らせです。",
					Posted:        false,
					CategoryType:  domain.FAQ_CATEGORY_ACCOUNT_REGISTRATION,
					CommonId:      uid.String(),
					OwnerId:       12378123,
					UpdateId:      12378123,
					AttachmentCnt: 1,
				}).Return(int64(0), nil)
				return repo
			},
			s3pkg: func() *mocks.S3 {
				s3 := new(mocks.S3)
				s3.On("GetHeadObject", "es-jp-devv2-public-file-storage", "/attachment_file/2024-09-09/crfbl6cldh5hujq0ulq0.jpg").Return(s3utils.S3ObjectInfo{
					FileSize:         123873,
					OriginalFileName: "crfbl6cldh5hujq0ulq0.jpg",
				}, nil)

				s3.On("GetHeadObject", "es-jp-devv2-public-file-storage", "/content_img_file/2024-09-10/crg1ackldh5j4kkqiueg.jpg").Return(s3utils.S3ObjectInfo{
					FileSize:         123773,
					OriginalFileName: "crg1ackldh5j4kkqiueg.jpg",
				}, nil)
				return s3
			},
			attachmentRepository: func() *mocks.AttachmentRepository {
				repo := new(mocks.AttachmentRepository)
				repo.On("Create", context.Background(), domain.Attachment{
					FileName:      "crfbl6cldh5hujq0ulq0",
					OriginalName:  "crfbl6cldh5hujq0ulq0.jpg",
					FileSize:      123873,
					FileExtension: ".jpg",
					FilePath:      "/attachment_file/2024-09-09/",
					FileUrl:       "https://cdn.dev.energyshares.jp/attachment_file/2024-09-09/crfbl6cldh5hujq0ulq0.jpg",
					ResourceId:    uid.String(),
				}).Return(int64(0), nil)
				return repo
			},
			contentImageRepository: func() *mocks.ContentImageRepository {
				repo := new(mocks.ContentImageRepository)
				repo.On("Create", context.Background(), domain.ContentImage{
					FileName:      "crg1ackldh5j4kkqiueg",
					OriginalName:  "crg1ackldh5j4kkqiueg.jpg",
					FileSize:      123773,
					FileExtension: ".jpg",
					FilePath:      "/content_img_file/2024-09-10/",
					FileUrl:       "https://cdn.dev.energyshares.jp/content_img_file/2024-09-10/crg1ackldh5j4kkqiueg.jpg",
					ResourceId:    uid.String(),
				}).Return(int64(0), nil)
				return repo
			},

			operatorHistoryRepository: func() *mocks.OperatorHistoryRepository {
				repo := new(mocks.OperatorHistoryRepository)
				repo.On("Create", context.Background(), domain.OperatorHistory{
					UserId: 12378123,
					Name:   "ホン・ギルドン",
					Email:  "gildong@email.com",
					Menu:   "menu.manage_website.manage_faq_register",
					Ip:     "127.0.0.1",
					State:  domain.REGISTER,
				}).Return(nil)

				return repo
			},
		},
		{
			name: "ピンの数超過によるエラー発生",
			args: args{
				ctx: context.Background(),
				request: adomain.FaqCreateRequest{
					Pin:              lo.ToPtr(true),
					ContentHtml:      "<p>Hello</p>",
					Title:            "重要なお知らせです。",
					Posted:           false,
					CategoryType:     string(domain.FAQ_CATEGORY_ACCOUNT_REGISTRATION),
					ContentImageUrls: []string{"https://es-jp-devv2-public-file-storage.s3.ap-northeast-2.amazonaws.com/content_img_file/2024-09-10/crg1ackldh5j4kkqiueg.jpg"},
					AttachmentUrls:   []string{"https://es-jp-devv2-public-file-storage.s3.ap-northeast-2.amazonaws.com/attachment_file/2024-09-09/crfbl6cldh5hujq0ulq0.jpg"},
					JWTUser: jwthelper.User{
						Id:    12378123,
						Name:  "홍길동",
						Email: "gildong@email.com",
					},
					ClientIp: "127.0.0.1",
				},
			},
			idGenerator: func() *mocks.IdGenerator {
				gen := new(mocks.IdGenerator)
				gen.On("GenUuidV4").Return(uid)
				return gen
			},
			faqRepository: func() *mocks.FaqRepository {
				repo := new(mocks.FaqRepository)
				repo.On("GetAll", context.Background(), mock.MatchedBy(func(field func(fields *cdomain.FaqQueryOptionFields)) bool {
					opt := cdomain.FaqQueryOptionFunc()
					field(&opt)

					return *opt.Pin() == true
				})).Return(int64(4), []domain.Faq{
					{}, {}, {}, // 4건 반환
				}, nil)
				return repo
			},
			s3pkg: func() *mocks.S3 {
				return new(mocks.S3)
			},
			attachmentRepository: func() *mocks.AttachmentRepository {
				return new(mocks.AttachmentRepository)
			},
			contentImageRepository: func() *mocks.ContentImageRepository {
				return new(mocks.ContentImageRepository)
			},

			operatorHistoryRepository: func() *mocks.OperatorHistoryRepository {
				return new(mocks.OperatorHistoryRepository)
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := faqUsecase{
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
				fileHelper:                filehelper.NewFileHelper(),
				s3pkg:                     tt.s3pkg(),
				idGenerator:               tt.idGenerator(),
				faqRepository:             tt.faqRepository(),
				contentImageRepository:    tt.contentImageRepository(),
				attachmentRepository:      tt.attachmentRepository(),
				operatorHistoryRepository: tt.operatorHistoryRepository(),
			}

			err := f.Create(tt.args.ctx, tt.args.request)
			assert.Equalf(t, tt.wantErr, err != nil, "Create() error = %v, wantErr %v", err, tt.wantErr)
		})
	}
}

func Test_faqUsecase_Update(t *testing.T) {
	serviceutils.SetCDN("cdn.dev.energyshares.jp")
	t.Cleanup(func() {
		serviceutils.SetCDN("")
	})

	type args struct {
		ctx     context.Context
		request adomain.FaqUpdateRequest
	}
	tests := []struct {
		name                      string
		args                      args
		s3pkg                     func() *mocks.S3
		faqRepository             func() *mocks.FaqRepository
		attachmentRepository      func() *mocks.AttachmentRepository
		contentImageRepository    func() *mocks.ContentImageRepository
		operatorHistoryRepository func() *mocks.OperatorHistoryRepository
		wantErr                   bool
	}{
		{
			name: "正常に呼び出された場合 (添付ファイル/画像がないFAQについて更新)",
			args: args{
				ctx: context.Background(),
				request: adomain.FaqUpdateRequest{
					Id:               mockFaq1.Id,
					Pin:              nil,
					ContentHtml:      "<p>Hello</p>",
					Title:            "重要なお知らせです。",
					Posted:           false,
					CategoryType:     string(domain.FAQ_CATEGORY_ACCOUNT_REGISTRATION),
					ContentImageUrls: []string{"https://es-jp-devv2-public-file-storage.s3.ap-northeast-2.amazonaws.com/content_img_file/2024-09-10/crg1ackldh5j4kkqiueg.jpg"},
					AttachmentUrls:   []string{"https://es-jp-devv2-public-file-storage.s3.ap-northeast-2.amazonaws.com/attachment_file/2024-09-09/crfbl6cldh5hujq0ulq0.jpg"},
					JWTUser: jwthelper.User{
						Id:    12378123,
						Name:  "ホン・ギルドン",
						Email: "gildong@email.com",
					},
					ClientIp: "127.0.0.1",
				},
			},
			faqRepository: func() *mocks.FaqRepository {
				repo := new(mocks.FaqRepository)
				repo.On("GetById", context.Background(), mockFaq1.Id).Return(mockFaq1, nil)
				repo.On("Update", context.Background(), domain.Faq{
					Id:            mockFaq1.Id,
					Pin:           false,
					ContentHtml:   "<p>Hello</p>",
					Title:         "重要なお知らせです。",
					Posted:        false,
					CategoryType:  domain.FAQ_CATEGORY_ACCOUNT_REGISTRATION,
					CommonId:      mockFaq1.CommonId,
					UpdateId:      12378123,
					AttachmentCnt: 1,
				}).Return(nil)
				return repo
			},
			s3pkg: func() *mocks.S3 {
				s3 := new(mocks.S3)
				s3.On("GetHeadObject", "es-jp-devv2-public-file-storage", "/attachment_file/2024-09-09/crfbl6cldh5hujq0ulq0.jpg").Return(s3utils.S3ObjectInfo{
					FileSize:         123873,
					OriginalFileName: "crfbl6cldh5hujq0ulq0.jpg",
				}, nil)

				s3.On("GetHeadObject", "es-jp-devv2-public-file-storage", "/content_img_file/2024-09-10/crg1ackldh5j4kkqiueg.jpg").Return(s3utils.S3ObjectInfo{
					FileSize:         123773,
					OriginalFileName: "crg1ackldh5j4kkqiueg.jpg",
				}, nil)
				return s3
			},
			attachmentRepository: func() *mocks.AttachmentRepository {
				repo := new(mocks.AttachmentRepository)
				repo.On("GetAll", context.Background(), mock.MatchedBy(func(field func(fields *cdomain.AttachmentQueryOptionFields)) bool {
					opt := cdomain.AttachmentQueryOptionFunc()
					field(&opt)
					return *opt.ResourceId() == mockFaq1.CommonId
				})).Return([]domain.Attachment{}, nil)

				repo.On("Create", context.Background(), domain.Attachment{
					FileName:      "crfbl6cldh5hujq0ulq0",
					OriginalName:  "crfbl6cldh5hujq0ulq0.jpg",
					FileSize:      123873,
					FileExtension: ".jpg",
					FilePath:      "/attachment_file/2024-09-09/",
					FileUrl:       "https://cdn.dev.energyshares.jp/attachment_file/2024-09-09/crfbl6cldh5hujq0ulq0.jpg",
					ResourceId:    "bddad349-d0cc-4a1d-8843-431d54fcd2c7",
				}).Return(int64(0), nil)
				return repo
			},
			contentImageRepository: func() *mocks.ContentImageRepository {
				repo := new(mocks.ContentImageRepository)
				repo.On("GetAll", context.Background(), mock.MatchedBy(func(field func(fields *cdomain.ContentImageQueryOptionFields)) bool {
					opt := cdomain.ContentImageQueryOptionFunc()
					field(&opt)
					return *opt.ResourceId() == mockFaq1.CommonId
				})).Return([]domain.ContentImage{}, nil)

				repo.On("Create", context.Background(), domain.ContentImage{
					FileName:      "crg1ackldh5j4kkqiueg",
					OriginalName:  "crg1ackldh5j4kkqiueg.jpg",
					FileSize:      123773,
					FileExtension: ".jpg",
					FilePath:      "/content_img_file/2024-09-10/",
					FileUrl:       "https://cdn.dev.energyshares.jp/content_img_file/2024-09-10/crg1ackldh5j4kkqiueg.jpg",
					ResourceId:    "bddad349-d0cc-4a1d-8843-431d54fcd2c7",
				}).Return(int64(0), nil)
				return repo
			},

			operatorHistoryRepository: func() *mocks.OperatorHistoryRepository {
				repo := new(mocks.OperatorHistoryRepository)
				repo.On("Create", context.Background(), domain.OperatorHistory{
					UserId: 12378123,
					Name:   "ホン・ギルドン",
					Email:  "gildong@email.com",
					Menu:   "menu.manage_website.manage_faq_detail",
					Title:  "重要なお知らせです。",
					Ip:     "127.0.0.1",
					State:  domain.CHANGE,
				}).Return(nil)

				return repo
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := faqUsecase{
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
				fileHelper:                filehelper.NewFileHelper(),
				s3pkg:                     tt.s3pkg(),
				faqRepository:             tt.faqRepository(),
				contentImageRepository:    tt.contentImageRepository(),
				attachmentRepository:      tt.attachmentRepository(),
				operatorHistoryRepository: tt.operatorHistoryRepository(),
			}

			err := f.Update(tt.args.ctx, tt.args.request)
			assert.Equalf(t, tt.wantErr, err != nil, "Create() error = %v, wantErr %v", err, tt.wantErr)
		})
	}
}

func Test_faqUsecase_GetAll(t *testing.T) {
	type args struct {
		ctx     context.Context
		request adomain.FaqGetAllRequest
	}
	tests := []struct {
		name                      string
		args                      args
		faqRepository             func() *mocks.FaqRepository
		attachmentRepository      func() *mocks.AttachmentRepository
		operatorHistoryRepository func() *mocks.OperatorHistoryRepository
		want                      adomain.FaqGetAllResponse
		wantErr                   bool
	}{
		{
			name: "正常に呼び出された場合",
			args: args{
				ctx: context.Background(),
				request: adomain.FaqGetAllRequest{
					JWTUser: jwthelper.User{
						Id:    12378123,
						Name:  "ホン・ギルドン",
						Email: "gildong@email.com",
					},
					ClientIp: "127.0.0.1",
					FaqGetAllQueryParam: adomain.FaqGetAllQueryParam{
						StartDate:    lo.ToPtr("2024-07-10T00:00:00+09:00"),
						EndDate:      lo.ToPtr("2024-08-15T23:59:59+09:00"),
						Posted:       lo.ToPtr(true),
						CategoryType: lo.ToPtr(string(domain.FAQ_CATEGORY_FUNDS)),
						SearchInput: cdomain.SearchInput{
							SkipCount:      0,
							MaxResultCount: 10,
							Sortby:         "title",
							Order:          "asc",
						},
					},
				},
			},
			faqRepository: func() *mocks.FaqRepository {
				repo := new(mocks.FaqRepository)
				repo.On("GetAll", context.Background(),
					mock.Anything, mock.Anything,
					mock.MatchedBy(func(field func(*cdomain.FaqQueryOptionFields)) bool {
						opt := cdomain.FaqQueryOptionFunc()
						field(&opt)
						return *opt.Posted() == true
					}),
					mock.MatchedBy(func(field func(*cdomain.FaqQueryOptionFields)) bool {
						opt := cdomain.FaqQueryOptionFunc()
						field(&opt)

						expected := lo.Must(time.Parse(time.RFC3339, "2024-07-10T00:00:00+09:00"))
						return opt.StartDate().Unix() == expected.Unix()
					}),

					mock.MatchedBy(func(field func(*cdomain.FaqQueryOptionFields)) bool {
						opt := cdomain.FaqQueryOptionFunc()
						field(&opt)

						expected := lo.Must(time.Parse(time.RFC3339, "2024-08-15T23:59:59+09:00"))
						return opt.EndDate().Unix() == expected.Unix()
					}),
					mock.MatchedBy(func(field func(*cdomain.FaqQueryOptionFields)) bool {
						opt := cdomain.FaqQueryOptionFunc()
						field(&opt)
						return opt.SortBy() == "pin,title"
					}),
					mock.MatchedBy(func(field func(*cdomain.FaqQueryOptionFields)) bool {
						opt := cdomain.FaqQueryOptionFunc()
						field(&opt)
						return opt.Order() == "desc,asc"
					}),
					mock.MatchedBy(func(field func(*cdomain.FaqQueryOptionFields)) bool {
						opt := cdomain.FaqQueryOptionFunc()
						field(&opt)
						return opt.Offset() == 0
					}),
					mock.MatchedBy(func(field func(*cdomain.FaqQueryOptionFields)) bool {
						opt := cdomain.FaqQueryOptionFunc()
						field(&opt)
						return opt.Limit() == 10
					}),
				).Return(int64(1), []domain.Faq{mockFaq1}, nil)
				return repo
			},
			attachmentRepository: func() *mocks.AttachmentRepository {
				repo := new(mocks.AttachmentRepository)
				repo.On("GetAll", context.Background(),
					mock.MatchedBy(func(field func(*cdomain.AttachmentQueryOptionFields)) bool {
						opt := cdomain.AttachmentQueryOptionFunc()
						field(&opt)
						return *opt.ResourceId() == "bddad349-d0cc-4a1d-8843-431d54fcd2c7"
					}),
				).Return([]domain.Attachment{mockAttachment1}, nil)
				return repo
			},
			operatorHistoryRepository: func() *mocks.OperatorHistoryRepository {
				repo := new(mocks.OperatorHistoryRepository)
				repo.On("Create", context.Background(), domain.OperatorHistory{
					UserId: 12378123,
					Name:   "ホン・ギルドン",
					Email:  "gildong@email.com",
					Menu:   "menu.manage_website.manage_faq_list",
					Ip:     "127.0.0.1",
					State:  domain.LOOKUP,
				}).Return(nil)

				return repo
			},

			want: adomain.FaqGetAllResponse{
				TotalCount: 1,
				Items: []domain.Faq{
					mockFaq1.WithAttachments([]domain.Attachment{mockAttachment1}),
				},
			},
			wantErr: false,
		},
		{
			// MIR-1113: sortby/order未指定のとき "pin," / "desc," のように
			// 末尾に空要素が残る。pkg/orm.SortOrder は空要素を列指定なしの
			// プレースホルダとして読み飛ばすため、既定の pin desc がそのまま
			// 効くことを確認する。
			name: "sortby/orderが未指定でも既定のpin descで並ぶ。",
			args: args{
				ctx: context.Background(),
				request: adomain.FaqGetAllRequest{
					JWTUser: jwthelper.User{
						Id:    12378123,
						Name:  "ホン・ギルドン",
						Email: "gildong@email.com",
					},
					ClientIp: "127.0.0.1",
					FaqGetAllQueryParam: adomain.FaqGetAllQueryParam{
						SearchInput: cdomain.SearchInput{
							SkipCount:      0,
							MaxResultCount: 10,
							Sortby:         "",
							Order:          "",
						},
					},
				},
			},
			faqRepository: func() *mocks.FaqRepository {
				repo := new(mocks.FaqRepository)
				repo.On("GetAll", context.Background(),
					mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything,
					mock.MatchedBy(func(field func(*cdomain.FaqQueryOptionFields)) bool {
						opt := cdomain.FaqQueryOptionFunc()
						field(&opt)
						return opt.SortBy() == "pin,"
					}),
					mock.MatchedBy(func(field func(*cdomain.FaqQueryOptionFields)) bool {
						opt := cdomain.FaqQueryOptionFunc()
						field(&opt)
						return opt.Order() == "desc,"
					}),
					mock.MatchedBy(func(field func(*cdomain.FaqQueryOptionFields)) bool {
						opt := cdomain.FaqQueryOptionFunc()
						field(&opt)
						return opt.Offset() == 0
					}),
					mock.MatchedBy(func(field func(*cdomain.FaqQueryOptionFields)) bool {
						opt := cdomain.FaqQueryOptionFunc()
						field(&opt)
						return opt.Limit() == 10
					}),
				).Return(int64(1), []domain.Faq{mockFaq1}, nil)
				return repo
			},
			attachmentRepository: func() *mocks.AttachmentRepository {
				repo := new(mocks.AttachmentRepository)
				repo.On("GetAll", context.Background(),
					mock.MatchedBy(func(field func(*cdomain.AttachmentQueryOptionFields)) bool {
						opt := cdomain.AttachmentQueryOptionFunc()
						field(&opt)
						return *opt.ResourceId() == "bddad349-d0cc-4a1d-8843-431d54fcd2c7"
					}),
				).Return([]domain.Attachment{mockAttachment1}, nil)
				return repo
			},
			operatorHistoryRepository: func() *mocks.OperatorHistoryRepository {
				repo := new(mocks.OperatorHistoryRepository)
				repo.On("Create", context.Background(), domain.OperatorHistory{
					UserId: 12378123,
					Name:   "ホン・ギルドン",
					Email:  "gildong@email.com",
					Menu:   "menu.manage_website.manage_faq_list",
					Ip:     "127.0.0.1",
					State:  domain.LOOKUP,
				}).Return(nil)

				return repo
			},

			want: adomain.FaqGetAllResponse{
				TotalCount: 1,
				Items: []domain.Faq{
					mockFaq1.WithAttachments([]domain.Attachment{mockAttachment1}),
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := faqUsecase{
				faqRepository:             tt.faqRepository(),
				attachmentRepository:      tt.attachmentRepository(),
				operatorHistoryRepository: tt.operatorHistoryRepository(),
			}

			got, err := f.GetAll(tt.args.ctx, tt.args.request)

			assert.NoErrorf(t, err, "GetAll() error = %v, wantErr %v", err, tt.wantErr)
			assert.Equalf(t, tt.want, got, "GetAll() = %v, want %v", got, tt.want)
		})
	}
}

func Test_faqUsecase_GetOne(t *testing.T) {
	type args struct {
		ctx     context.Context
		request adomain.FaqGetOneRequest
	}
	tests := []struct {
		name                      string
		args                      args
		faqRepository             func() *mocks.FaqRepository
		attachmentRepository      func() *mocks.AttachmentRepository
		operatorHistoryRepository func() *mocks.OperatorHistoryRepository
		want                      adomain.FaqGetOneResponse
		wantErr                   bool
	}{
		{
			name: "正常に呼び出された場合",
			args: args{
				ctx: context.Background(),
				request: adomain.FaqGetOneRequest{
					JWTUser: jwthelper.User{
						Id:    12378123,
						Name:  "ホン・ギルドン",
						Email: "gildong@email.com",
					},
					ClientIp: "127.0.0.1",
					Id:       1000,
				},
			},
			faqRepository: func() *mocks.FaqRepository {
				repo := new(mocks.FaqRepository)
				repo.On("GetById", context.Background(), int64(1000)).Return(mockFaq1, nil)
				return repo
			},
			attachmentRepository: func() *mocks.AttachmentRepository {
				repo := new(mocks.AttachmentRepository)
				repo.On("GetAll", context.Background(),
					mock.MatchedBy(func(field func(*cdomain.AttachmentQueryOptionFields)) bool {
						opt := cdomain.AttachmentQueryOptionFunc()
						field(&opt)
						return *opt.ResourceId() == "bddad349-d0cc-4a1d-8843-431d54fcd2c7"
					}),
				).Return([]domain.Attachment{mockAttachment1}, nil)
				return repo
			},
			operatorHistoryRepository: func() *mocks.OperatorHistoryRepository {
				repo := new(mocks.OperatorHistoryRepository)
				repo.On("Create", context.Background(), domain.OperatorHistory{
					UserId: 12378123,
					Name:   "ホン・ギルドン",
					Email:  "gildong@email.com",
					Menu:   "menu.manage_website.manage_faq_detail",
					Ip:     "127.0.0.1",
					State:  domain.LOOKUP,
				}).Return(nil)
				return repo
			},

			want: adomain.FaqGetOneResponse{
				Faq: mockFaq1.WithAttachments([]domain.Attachment{mockAttachment1}),
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := faqUsecase{
				faqRepository:             tt.faqRepository(),
				attachmentRepository:      tt.attachmentRepository(),
				operatorHistoryRepository: tt.operatorHistoryRepository(),
			}
			got, err := f.GetOne(tt.args.ctx, tt.args.request)
			assert.NoErrorf(t, err, "GetOne() error = %v, wantErr %v", err, tt.wantErr)
			assert.Equalf(t, tt.want, got, "GetOne() = %v, want %v", got, tt.want)
		})
	}
}

func Test_faqUsecase_Delete(t *testing.T) {
	type args struct {
		ctx     context.Context
		request adomain.FaqDeleteRequest
	}
	tests := []struct {
		name                      string
		args                      args
		s3pkg                     func() *mocks.S3
		faqRepository             func() *mocks.FaqRepository
		operatorHistoryRepository func() *mocks.OperatorHistoryRepository
		attachmentRepository      func() *mocks.AttachmentRepository
		contentImageRepository    func() *mocks.ContentImageRepository
		wantErr                   bool
	}{
		{
			name: "正常に呼び出された場合",
			args: args{
				ctx: context.Background(),
				request: adomain.FaqDeleteRequest{
					JWTUser: jwthelper.User{
						Id:    12378123,
						Name:  "ホン・ギルドン",
						Email: "gildong@email.com",
					},
					ClientIp: "127.0.0.1",
					Id:       1000,
				},
			},
			faqRepository: func() *mocks.FaqRepository {
				repo := new(mocks.FaqRepository)
				repo.On("GetById", context.Background(), int64(1000)).Return(mockFaq1, nil)
				repo.On("Delete", context.Background(), int64(1000)).Return(nil)
				return repo
			},
			operatorHistoryRepository: func() *mocks.OperatorHistoryRepository {
				repo := new(mocks.OperatorHistoryRepository)
				repo.On("Create", context.Background(), domain.OperatorHistory{
					UserId: 12378123,
					Name:   "ホン・ギルドン",
					Email:  "gildong@email.com",
					Menu:   "menu.manage_website.manage_faq_detail",
					Ip:     "127.0.0.1",
					Title:  "重要なお知らせです。",
					State:  domain.DELETE,
				}).Return(nil)
				return repo
			},
			attachmentRepository: func() *mocks.AttachmentRepository {
				repo := new(mocks.AttachmentRepository)
				repo.On("GetAll", context.Background(),
					mock.MatchedBy(func(field func(*cdomain.AttachmentQueryOptionFields)) bool {
						opt := cdomain.AttachmentQueryOptionFunc()
						field(&opt)
						return *opt.ResourceId() == "bddad349-d0cc-4a1d-8843-431d54fcd2c7"
					}),
				).Return([]domain.Attachment{mockAttachment1}, nil)

				repo.On("DeleteByResourceId", context.Background(), "bddad349-d0cc-4a1d-8843-431d54fcd2c7").Return(nil)

				return repo
			},

			contentImageRepository: func() *mocks.ContentImageRepository {
				repo := new(mocks.ContentImageRepository)
				repo.On("GetAll", context.Background(),
					mock.MatchedBy(func(field func(*cdomain.ContentImageQueryOptionFields)) bool {
						opt := cdomain.ContentImageQueryOptionFunc()
						field(&opt)
						return *opt.ResourceId() == "bddad349-d0cc-4a1d-8843-431d54fcd2c7"
					}),
				).Return([]domain.ContentImage{mockContentImage1}, nil)

				repo.On("DeleteByResourceId", context.Background(), "bddad349-d0cc-4a1d-8843-431d54fcd2c7").Return(nil)
				return repo
			},
			s3pkg: func() *mocks.S3 {
				s3 := new(mocks.S3)
				s3.On("DeleteObject", "public-bucket", "/attachment_file/2702602/bpbmq9297s524a3jhgn1.pdf").Return(nil)
				s3.On("DeleteObject", "public-bucket", "/content_image/2702602/bpbmq9297s524a3jhgn1.jpg").Return(nil)
				return s3
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := faqUsecase{
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
							S3_PUBLIC_BUCKET: "public-bucket",
						},
					},
				},
				faqRepository:             tt.faqRepository(),
				operatorHistoryRepository: tt.operatorHistoryRepository(),
				attachmentRepository:      tt.attachmentRepository(),
				contentImageRepository:    tt.contentImageRepository(),
				s3pkg:                     tt.s3pkg(),
			}
			err := f.Delete(tt.args.ctx, tt.args.request)
			assert.NoErrorf(t, err, "Delete() error = %v, wantErr %v", err, tt.wantErr)
		})
	}
}
