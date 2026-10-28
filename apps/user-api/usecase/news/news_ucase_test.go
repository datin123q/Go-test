package news

import (
	"context"
	"testing"

	"github.com/stretchr/testify/mock"

	cdomain "github.com/illuminarean/energy-shares-backend-repo/clean-arch-domain"
	"github.com/illuminarean/energy-shares-backend-repo/clean-arch-domain/mocks"
	"github.com/illuminarean/energy-shares-backend-repo/domain"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/jwthelper"
	udomain "github.com/illuminarean/energyshares-user-api/domain"
)

func Test_NewsUsecase_GetAll(t *testing.T) {
	type args struct {
		ctx     context.Context
		request udomain.NewsGetAllRequest
	}
	type fields struct {
		newsRepository            cdomain.NewsRepository
		operatorHistoryRepository cdomain.OperatorHistoryRepository
	}

	tests := []struct {
		name    string
		args    args
		fields  fields
		want    int64
		want1   udomain.NewsGetAllResponse
		wantErr bool
	}{
		{
			name: "お知らせ一覧照会_成功",
			args: args{
				ctx: context.Background(),
				request: udomain.NewsGetAllRequest{
					JWTUser: jwthelper.User{
						Id:    12378123,
						Name:  "ホン・ギルドン",
						Email: "gildong@email.com",
					},
					ClientIp: "127.0.0.1",
				},
			},
			fields: fields{
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
						int64(2),
						[]domain.News{
							mockNews1,
							mockNews2,
						}, nil)
					return repo
				}(),
				operatorHistoryRepository: func() *mocks.OperatorHistoryRepository {
					repo := new(mocks.OperatorHistoryRepository)
					repo.On("Create", context.Background(), domain.OperatorHistory{
						UserId: 12378123,
						Name:   "ホン・ギルドン",
						Email:  "gildong@email.com",
						Menu:   "menu.manage_website.manage_news_list",
						Ip:     "127.0.0.1",
						State:  domain.LOOKUP,
					}).Return(nil)
					return repo
				}(),
			},

			want: 2,
			want1: udomain.NewsGetAllResponse{
				TotalCount: 2,
				Items: []domain.News{
					mockNews1,
					mockNews2,
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u := newsUsecase{
				newsRepository:            tt.fields.newsRepository,
				operatorHistoryRepository: tt.fields.operatorHistoryRepository,
			}
			got, err := u.GetAll(tt.args.ctx, tt.args.request)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetAll() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if len(got.Items) != int(tt.want) {
				t.Errorf("GetAll() got = %v, want %v", got, tt.want)
			}
		})
	}
}
