package news

import (
	"context"
	"testing"

	"github.com/gofrs/uuid"
	"github.com/samber/lo"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	adomain "github.com/illuminarean/energy-share-core-api/domain"
	cdomain "github.com/illuminarean/energy-shares-backend-repo/clean-arch-domain"
	"github.com/illuminarean/energy-shares-backend-repo/clean-arch-domain/mocks"
	"github.com/illuminarean/energy-shares-backend-repo/domain"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/config"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/filehelper"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/jwthelper"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/s3utils"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/serviceutils"
)

func Test_NewsUsecase_GetAll(t *testing.T) {

	type args struct {
		ctx     context.Context
		request adomain.NewsGetAllRequest
	}
	tests := []struct {
		name                      string
		args                      args
		newsRepository            func() *mocks.NewsRepository
		attachmentRepository      func() *mocks.AttachmentRepository
		operatorHistoryRepository func() *mocks.OperatorHistoryRepository
		wantTotal                 int64
		wantResponse              adomain.NewsGetAllResponse
		wantErr                   bool
	}{
		{
			name: "お知らせリスト照会_成功",
			args: args{
				ctx: context.Background(),
				request: adomain.NewsGetAllRequest{
					JWTUser: jwthelper.User{
						Id:    12378123,
						Name:  "홍길동",
						Email: "gildong@email.com",
					},
					ClientIp: "127.0.0.1",
					NewsGetAllQueryParam: adomain.NewsGetAllQueryParam{
						Title:     "get all",
						StartDate: "2024-07-10T00:00:00+09:00",
						EndDate:   "2024-08-15T23:59:59+09:00",
						Posted:    lo.ToPtr(true),
						SearchInput: cdomain.SearchInput{
							SkipCount:      0,
							MaxResultCount: 10,
							Sortby:         "title",
							Order:          "asc",
						},
					},
				},
			},
			newsRepository: func() *mocks.NewsRepository {
				repo := &mocks.NewsRepository{}
				repo.On("GetAll", context.Background(),
					mock.AnythingOfType("func(*clean_arch_domain.NewsQueryOptionFields)"),
					mock.AnythingOfType("func(*clean_arch_domain.NewsQueryOptionFields)"),
					mock.AnythingOfType("func(*clean_arch_domain.NewsQueryOptionFields)"),
					mock.AnythingOfType("func(*clean_arch_domain.NewsQueryOptionFields)"),
					mock.AnythingOfType("func(*clean_arch_domain.NewsQueryOptionFields)"),
					mock.AnythingOfType("func(*clean_arch_domain.NewsQueryOptionFields)"),
					mock.AnythingOfType("func(*clean_arch_domain.NewsQueryOptionFields)"),
					mock.AnythingOfType("func(*clean_arch_domain.NewsQueryOptionFields)"),
					mock.AnythingOfType("func(*clean_arch_domain.NewsQueryOptionFields)"),
				).Return(
					int64(1),
					[]domain.News{mockNews1}, nil)
				return repo
			},
			attachmentRepository: func() *mocks.AttachmentRepository {
				repo := &mocks.AttachmentRepository{}
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
					Name:   "홍길동",
					Email:  "gildong@email.com",
					Menu:   "menu.manage_website.manage_news_list",
					Ip:     "127.0.0.1",
					State:  domain.LOOKUP,
				}).Return(nil)
				return repo
			},
			wantTotal: 1,
			wantResponse: adomain.NewsGetAllResponse{
				TotalCount: 1,
				Items: []domain.News{
					mockNews1.WithAttachments([]domain.Attachment{mockAttachment1}),
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u := newsUsecase{
				newsRepository:            tt.newsRepository(),
				attachmentRepository:      tt.attachmentRepository(),
				operatorHistoryRepository: tt.operatorHistoryRepository(),
			}
			got, err := u.GetAll(tt.args.ctx, tt.args.request)

			assert.Equalf(t, tt.wantTotal, got.TotalCount, "GetAll() total = %v, want %v", got.TotalCount, tt.wantTotal)
			assert.Equalf(t, tt.wantResponse, got, "GetAll() = %v, want %v", got, tt.wantResponse)
			assert.Equalf(t, tt.wantErr, err != nil, "GetAll() error = %v, wantErr %v", err, tt.wantErr)
		})
	}
}

func Test_NewsUsecase_GetOne(t *testing.T) {
	type args struct {
		ctx     context.Context
		request adomain.NewsGetOneRequest
	}

	tests := []struct {
		name                      string
		args                      args
		newsRepository            func() *mocks.NewsRepository
		attachmentRepository      func() *mocks.AttachmentRepository
		operatorHistoryRepository func() *mocks.OperatorHistoryRepository
		want                      adomain.NewsGetOneResponse
		wantErr                   bool
	}{
		{
			name: "お知らせ詳細照会_成功",
			args: args{
				ctx: context.Background(),
				request: adomain.NewsGetOneRequest{
					JWTUser: jwthelper.User{
						Id:    12378123,
						Name:  "홍길동",
						Email: "gildong@email.com",
					},
					ClientIp: "127.0.0.1",
					NewsId:   1000,
				},
			},
			newsRepository: func() *mocks.NewsRepository {
				repo := &mocks.NewsRepository{}
				repo.On("GetById", context.Background(), int64(1000)).
					Return(mockNews1, nil)
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
					Name:   "홍길동",
					Email:  "gildong@email.com",
					Menu:   "menu.manage_website.manage_news_detail",
					Ip:     "127.0.0.1",
					State:  domain.LOOKUP,
				}).Return(nil)
				return repo
			},

			want: adomain.NewsGetOneResponse{
				News: mockNews1.WithAttachments([]domain.Attachment{mockAttachment1}),
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u := newsUsecase{
				newsRepository:            tt.newsRepository(),
				attachmentRepository:      tt.attachmentRepository(),
				operatorHistoryRepository: tt.operatorHistoryRepository(),
			}
			got, err := u.GetOne(tt.args.ctx, tt.args.request)

			assert.Equalf(t, tt.wantErr, err != nil, "GetOne() error = %v, wantErr %v", err, tt.wantErr)
			assert.Equalf(t, tt.want, got, "GetOne() = %v, want %v", got, tt.want)
		})
	}
}

func Test_NewsUsecase_Delete(t *testing.T) {
	type args struct {
		ctx     context.Context
		request adomain.NewsDeleteRequest
	}

	tests := []struct {
		name                      string
		args                      args
		s3pkg                     func() *mocks.S3
		newsRepository            func() *mocks.NewsRepository
		operatorHistoryRepository func() *mocks.OperatorHistoryRepository
		attachmentRepository      func() *mocks.AttachmentRepository
		contentImageRepository    func() *mocks.ContentImageRepository
		wantErr                   bool
	}{
		{
			name: "お知らせ削除_成功",
			args: args{
				ctx: context.Background(),
				request: adomain.NewsDeleteRequest{
					JWTUser: jwthelper.User{
						Id:    12378123,
						Name:  "홍길동",
						Email: "gildong@email.com",
					},
					ClientIp: "127.0.0.1",
					NewsId:   mockNews1.Id,
				},
			},
			newsRepository: func() *mocks.NewsRepository {
				repo := &mocks.NewsRepository{}
				repo.On("GetById", context.Background(), mockNews1.Id).
					Return(mockNews1, nil)
				repo.On("Delete", context.Background(), mockNews1.Id).
					Return(nil)
				return repo
			},
			attachmentRepository: func() *mocks.AttachmentRepository {
				repo := &mocks.AttachmentRepository{}
				repo.On("GetAll", context.Background(),
					mock.MatchedBy(func(field func(*cdomain.AttachmentQueryOptionFields)) bool {
						opt := cdomain.AttachmentQueryOptionFunc()
						field(&opt)
						return *opt.ResourceId() == "bddad349-d0cc-4a1d-8843-431d54fcd2c7"
					}),
				).Return([]domain.Attachment{mockAttachment1}, nil)
				repo.On("DeleteByResourceId", context.Background(),
					mockNews1.CommonId,
				).Return(nil)
				return repo
			},
			operatorHistoryRepository: func() *mocks.OperatorHistoryRepository {
				repo := new(mocks.OperatorHistoryRepository)
				repo.On("Create", context.Background(), domain.OperatorHistory{
					UserId: 12378123,
					Name:   "홍길동",
					Email:  "gildong@email.com",
					Menu:   "menu.manage_website.manage_news_detail",
					Ip:     "127.0.0.1",
					State:  domain.DELETE,
				}).Return(nil)
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
			u := newsUsecase{
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
				s3pkg:                     tt.s3pkg(),
				newsRepository:            tt.newsRepository(),
				attachmentRepository:      tt.attachmentRepository(),
				operatorHistoryRepository: tt.operatorHistoryRepository(),
			}
			err := u.Delete(tt.args.ctx, tt.args.request)

			assert.Equalf(t, tt.wantErr, err != nil, "Delete() error = %v, wantErr %v", err, tt.wantErr)
		})
	}
}

func Test_newsUsecase_Create(t *testing.T) {
	serviceutils.SetCDN("cdn.dev.energyshares.jp")
	t.Cleanup(func() {
		serviceutils.SetCDN("")
	})

	uid, _ := uuid.FromString("bddad349-d0cc-4a1d-8843-431d54fcd2c7")

	type args struct {
		ctx     context.Context
		request adomain.NewsCreateRequest
	}
	tests := []struct {
		name                      string
		args                      args
		s3pkg                     func() *mocks.S3
		idGenerator               func() *mocks.IdGenerator
		newsRepository            func() *mocks.NewsRepository
		attachmentRepository      func() *mocks.AttachmentRepository
		contentImageRepository    func() *mocks.ContentImageRepository
		operatorHistoryRepository func() *mocks.OperatorHistoryRepository
		wantErr                   bool
	}{
		{
			name: "正常に呼び出された場合",
			args: args{
				ctx: context.Background(),
				request: adomain.NewsCreateRequest{
					Pin:              nil,
					ContentHtml:      "<p>Hello</p>",
					Title:            "중요 공지사항입니다.",
					Posted:           false,
					ContentImageUrls: []string{"https://es-jp-devv2-public-file-storage.s3.ap-northeast-2.amazonaws.com/content_img_file/2024-09-10/crg1ackldh5j4kkqiueg.jpg"},
					AttachmentUrls:   []string{"https://es-jp-devv2-public-file-storage.s3.ap-northeast-2.amazonaws.com/attachment_file/2024-09-09/crfbl6cldh5hujq0ulq0.jpg"},
					SendMailStatus:   false,
					JWTUser: jwthelper.User{
						Id:    12378123,
						Name:  "홍길동",
						Email: "gildong@email.com",
					},
					ClientIp: "127.0.0.1",
				},
			},
			newsRepository: func() *mocks.NewsRepository {
				repo := new(mocks.NewsRepository)
				repo.On("Create", context.Background(), domain.News{
					Pin:           false,
					ContentHtml:   "<p>Hello</p>",
					Title:         "중요 공지사항입니다.",
					Posted:        false,
					CommonId:      uid.String(),
					OwnerId:       12378123,
					UpdateId:      12378123,
					AttachmentCnt: 1,
				}).Return(int64(0), nil)
				return repo
			},
			s3pkg: func() *mocks.S3 {
				s3 := new(mocks.S3)
				s3.On("GetHeadObject", "es-jp-devv2-public-file-storage", "attachment_file/2024-09-09/crfbl6cldh5hujq0ulq0.jpg").Return(s3utils.S3ObjectInfo{
					FileSize:         123873,
					OriginalFileName: "crfbl6cldh5hujq0ulq0.jpg",
				}, nil)

				s3.On("GetHeadObject", "es-jp-devv2-public-file-storage", "content_img_file/2024-09-10/crg1ackldh5j4kkqiueg.jpg").Return(s3utils.S3ObjectInfo{
					FileSize:         123773,
					OriginalFileName: "crg1ackldh5j4kkqiueg.jpg",
				}, nil)
				return s3
			},
			idGenerator: func() *mocks.IdGenerator {
				gen := new(mocks.IdGenerator)
				gen.On("GenUuidV4").Return(uid)
				return gen
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
					Name:   "홍길동",
					Email:  "gildong@email.com",
					Menu:   "menu.manage_website.manage_news_register",
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
				request: adomain.NewsCreateRequest{
					Pin:              lo.ToPtr(true),
					ContentHtml:      "<p>Hello</p>",
					Title:            "중요 공지사항입니다.",
					Posted:           false,
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
			newsRepository: func() *mocks.NewsRepository {
				repo := new(mocks.NewsRepository)
				repo.On("GetAll", context.Background(), mock.MatchedBy(func(field func(fields *cdomain.NewsQueryOptionFields)) bool {
					opt := cdomain.NewsQueryOptionFunc()
					field(&opt)

					return *opt.Pin() == true
				})).Return(int64(4), []domain.News{
					{}, {}, {}, // 4件返却
				}, nil)
				return repo
			},
			s3pkg: func() *mocks.S3 {
				return new(mocks.S3)
			},
			idGenerator: func() *mocks.IdGenerator {
				gen := new(mocks.IdGenerator)
				gen.On("GenUuidV4").Return(uid)
				return gen
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
			f := newsUsecase{
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
				newsRepository:            tt.newsRepository(),
				contentImageRepository:    tt.contentImageRepository(),
				attachmentRepository:      tt.attachmentRepository(),
				operatorHistoryRepository: tt.operatorHistoryRepository(),
			}

			err := f.Create(tt.args.ctx, tt.args.request)
			assert.Equalf(t, tt.wantErr, err != nil, "Create() error = %v, wantErr %v", err, tt.wantErr)
		})
	}
}

func Test_newsUsecase_Update(t *testing.T) {
	serviceutils.SetCDN("cdn.dev.energyshares.jp")
	t.Cleanup(func() {
		serviceutils.SetCDN("")
	})

	type args struct {
		ctx     context.Context
		request adomain.NewsUpdateRequest
	}
	tests := []struct {
		name                      string
		args                      args
		s3pkg                     func() *mocks.S3
		newsRepository            func() *mocks.NewsRepository
		attachmentRepository      func() *mocks.AttachmentRepository
		contentImageRepository    func() *mocks.ContentImageRepository
		operatorHistoryRepository func() *mocks.OperatorHistoryRepository
		wantErr                   bool
	}{
		{
			name: "正常に呼び出された場合 (添付ファイル/画像がないnewsについて更新)",
			args: args{
				ctx: context.Background(),
				request: adomain.NewsUpdateRequest{
					Id:               mockNews1.Id,
					Pin:              nil,
					ContentHtml:      "<p>Hello</p>",
					Title:            "중요 공지사항입니다.",
					Posted:           false,
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
			newsRepository: func() *mocks.NewsRepository {
				repo := new(mocks.NewsRepository)
				repo.On("GetById", context.Background(), mockNews1.Id).Return(mockNews1, nil)
				repo.On("Update", context.Background(), domain.News{
					Id:            mockNews1.Id,
					Pin:           false,
					ContentHtml:   "<p>Hello</p>",
					Title:         "중요 공지사항입니다.",
					Posted:        false,
					CommonId:      mockNews1.CommonId,
					OwnerId:       12378123,
					UpdateId:      12378123,
					AttachmentCnt: 1,
				}).Return(nil)
				return repo
			},
			s3pkg: func() *mocks.S3 {
				s3 := new(mocks.S3)
				s3.On("GetHeadObject", "es-jp-devv2-public-file-storage", "attachment_file/2024-09-09/crfbl6cldh5hujq0ulq0.jpg").Return(s3utils.S3ObjectInfo{
					FileSize:         123873,
					OriginalFileName: "crfbl6cldh5hujq0ulq0.jpg",
				}, nil)

				s3.On("GetHeadObject", "es-jp-devv2-public-file-storage", "content_img_file/2024-09-10/crg1ackldh5j4kkqiueg.jpg").Return(s3utils.S3ObjectInfo{
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
					return *opt.ResourceId() == mockNews1.CommonId
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
					return *opt.ResourceId() == mockNews1.CommonId
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
					Name:   "홍길동",
					Email:  "gildong@email.com",
					Menu:   "menu.manage_website.manage_news_detail",
					Title:  "중요 공지사항입니다.",
					Ip:     "127.0.0.1",
					State:  domain.CHANGE,
				}).Return(nil)

				return repo
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := newsUsecase{
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
				newsRepository:            tt.newsRepository(),
				contentImageRepository:    tt.contentImageRepository(),
				attachmentRepository:      tt.attachmentRepository(),
				operatorHistoryRepository: tt.operatorHistoryRepository(),
			}

			err := f.Update(tt.args.ctx, tt.args.request)
			assert.Equalf(t, tt.wantErr, err != nil, "Create() error = %v, wantErr %v", err, tt.wantErr)
		})
	}
}
