package mysql

import (
	"context"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/stretchr/testify/assert"

	cdomain "github.com/illuminarean/energy-shares-backend-repo/clean-arch-domain"
	"github.com/illuminarean/energy-shares-backend-repo/domain"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/jsonutils"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/tool"
)

var noticeMsgTestJST = func() *time.Location {
	loc, err := time.LoadLocation("Asia/Tokyo")
	if err != nil {
		panic(err)
	}
	return loc
}()

func Test_noticeMsgRepository_GetAll(t *testing.T) {
	type args struct {
		ctx     context.Context
		options []cdomain.NoticeMsgQueryOption
	}
	tests := []struct {
		name    string
		args    args
		want    int64
		want1   []domain.NoticeMsg
		wantErr bool
	}{
		{
			name: "사용자 알림 목록 조회 (limit 3) => 정상",
			args: args{
				ctx: context.Background(),
				options: []cdomain.NoticeMsgQueryOption{
					cdomain.NoticeMsgQueryOptionFields{}.WithOffset(0),
					cdomain.NoticeMsgQueryOptionFields{}.WithLimit(3),
				},
			},
			want: 12,
			want1: []domain.NoticeMsg{
				{
					Id:             1,
					Title:          "test",
					ContentHtml:    "<p>test</p>",
					NoticeDate:     jsonutils.JSONTime(time.Date(2021, 7, 9, 17, 41, 45, 0, noticeMsgTestJST)),
					Created:        jsonutils.JSONTime{},
					NoticeUserType: domain.NOTICE_NEWSLETTER,
					ProjectId:      0,
				},
				{
					Id:             2,
					Title:          "뉴스레터 수신동의 테스트",
					ContentHtml:    "<p>123</p>",
					NoticeDate:     jsonutils.JSONTime(time.Date(2021, 7, 29, 13, 42, 41, 0, noticeMsgTestJST)),
					Created:        jsonutils.JSONTime{},
					NoticeUserType: domain.NOTICE_NEWSLETTER,
					ProjectId:      0,
				},
				{
					Id:             3,
					Title:          "직접입력 테스트",
					ContentHtml:    "<p>직접입력 테스트</p>",
					NoticeDate:     jsonutils.JSONTime(time.Date(2021, 7, 29, 13, 43, 11, 0, noticeMsgTestJST)),
					Created:        jsonutils.JSONTime{},
					NoticeUserType: domain.NOTICE_DIRECT,
					ProjectId:      0,
				},
			},
			wantErr: false,
		},
		{
			name: "사용자 알림 목록 조회 (limit 3, desc) => 정상",
			args: args{
				ctx: context.Background(),
				options: []cdomain.NoticeMsgQueryOption{
					cdomain.NoticeMsgQueryOptionFields{}.WithSortBy("id"),
					cdomain.NoticeMsgQueryOptionFields{}.WithOrder("desc"),
					cdomain.NoticeMsgQueryOptionFields{}.WithOffset(0),
					cdomain.NoticeMsgQueryOptionFields{}.WithLimit(3),
				},
			},
			want: 12,
			want1: []domain.NoticeMsg{
				{
					Id:             12,
					Title:          "test",
					ContentHtml:    "<p>test</p>",
					NoticeDate:     jsonutils.JSONTime(time.Date(2022, 2, 15, 13, 22, 32, 0, noticeMsgTestJST)),
					Created:        jsonutils.JSONTime{},
					NoticeUserType: domain.NOTICE_DIRECT,
					ProjectId:      0,
				},
				{
					Id:             11,
					Title:          "123",
					ContentHtml:    "<p>123</p>",
					NoticeDate:     jsonutils.JSONTime(time.Date(2021, 9, 23, 9, 56, 46, 0, noticeMsgTestJST)),
					Created:        jsonutils.JSONTime{},
					NoticeUserType: domain.NOTICE_PROJECT,
					ProjectId:      30,
				},
				{
					Id:             10,
					Title:          "test4",
					ContentHtml:    "<p>test4</p>",
					NoticeDate:     jsonutils.JSONTime(time.Date(2021, 8, 18, 14, 55, 7, 0, noticeMsgTestJST)),
					Created:        jsonutils.JSONTime{},
					NoticeUserType: domain.NOTICE_NEWSLETTER,
					ProjectId:      0,
				},
			},
			wantErr: false,
		},
		{
			name: "사용자 알림 목록 조회 (limit 3, desc, title) => 정상",
			args: args{
				ctx: context.Background(),
				options: []cdomain.NoticeMsgQueryOption{
					cdomain.NoticeMsgQueryOptionFields{}.WithSortBy("id"),
					cdomain.NoticeMsgQueryOptionFields{}.WithOrder("desc"),
					cdomain.NoticeMsgQueryOptionFields{}.WithOffset(0),
					cdomain.NoticeMsgQueryOptionFields{}.WithLimit(3),
					cdomain.NoticeMsgQueryOptionFields{}.WithTitle(tool.StringToPointer("test")),
				},
			},
			want: 6,
			want1: []domain.NoticeMsg{
				{
					Id:             12,
					Title:          "test",
					ContentHtml:    "<p>test</p>",
					NoticeDate:     jsonutils.JSONTime(time.Date(2022, 2, 15, 13, 22, 32, 0, noticeMsgTestJST)),
					Created:        jsonutils.JSONTime{},
					NoticeUserType: domain.NOTICE_DIRECT,
					ProjectId:      0,
				},
				{
					Id:             10,
					Title:          "test4",
					ContentHtml:    "<p>test4</p>",
					NoticeDate:     jsonutils.JSONTime(time.Date(2021, 8, 18, 14, 55, 7, 0, noticeMsgTestJST)),
					Created:        jsonutils.JSONTime{},
					NoticeUserType: domain.NOTICE_NEWSLETTER,
					ProjectId:      0,
				},
				{
					Id:             9,
					Title:          "test3",
					ContentHtml:    "<p>test3</p>",
					NoticeDate:     jsonutils.JSONTime(time.Date(2021, 8, 18, 14, 54, 49, 0, noticeMsgTestJST)),
					Created:        jsonutils.JSONTime{},
					NoticeUserType: domain.NOTICE_NEWSLETTER,
					ProjectId:      0,
				},
			},
			wantErr: false,
		},
		{
			name: "사용자 알림 목록 조회 (limit 3, desc, dates) => 정상",
			args: args{
				ctx: context.Background(),
				options: []cdomain.NoticeMsgQueryOption{
					cdomain.NoticeMsgQueryOptionFields{}.WithSortBy("id"),
					cdomain.NoticeMsgQueryOptionFields{}.WithOrder("desc"),
					cdomain.NoticeMsgQueryOptionFields{}.WithOffset(0),
					cdomain.NoticeMsgQueryOptionFields{}.WithLimit(3),
					cdomain.NoticeMsgQueryOptionFields{}.WithStartDate(tool.StringToPointer("2021-08-18 00:00:00")),
					cdomain.NoticeMsgQueryOptionFields{}.WithEndDate(tool.StringToPointer("2021-08-18 23:59:59")),
				},
			},
			want: 4,
			want1: []domain.NoticeMsg{
				{
					Id:             10,
					Title:          "test4",
					ContentHtml:    "<p>test4</p>",
					NoticeDate:     jsonutils.JSONTime(time.Date(2021, 8, 18, 14, 55, 7, 0, noticeMsgTestJST)),
					Created:        jsonutils.JSONTime{},
					NoticeUserType: domain.NOTICE_NEWSLETTER,
					ProjectId:      0,
				},
				{
					Id:             9,
					Title:          "test3",
					ContentHtml:    "<p>test3</p>",
					NoticeDate:     jsonutils.JSONTime(time.Date(2021, 8, 18, 14, 54, 49, 0, noticeMsgTestJST)),
					Created:        jsonutils.JSONTime{},
					NoticeUserType: domain.NOTICE_NEWSLETTER,
					ProjectId:      0,
				},
				{
					Id:             8,
					Title:          "test2",
					ContentHtml:    "<p>test2</p>",
					NoticeDate:     jsonutils.JSONTime(time.Date(2021, 8, 18, 14, 54, 49, 0, noticeMsgTestJST)),
					Created:        jsonutils.JSONTime{},
					NoticeUserType: domain.NOTICE_NEWSLETTER,
					ProjectId:      0,
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := noticeMsgRepository{
				dbFactory: dbFactoryMock{},
			}
			ignoreTime := cmpopts.IgnoreTypes(jsonutils.JSONTime{})

			got, got1, err := repo.GetAll(tt.args.ctx, tt.args.options...)

			assert.NoErrorf(t, err, "noticeMsgRepository.GetAll() error = %v", err)
			assert.Equalf(t, tt.want, got, "noticeMsgRepository.GetAll() got = %v, want %v", got, tt.want)
			assert.Truef(t, cmp.Equal(tt.want1, got1, ignoreTime), "noticeMsgRepository.GetAll() got1 = %v, want %v", got1, tt.want1)
		})
	}
}

func Test_noticeMsgRepository_GetById(t *testing.T) {
	type args struct {
		ctx context.Context
		id  int64
	}
	tests := []struct {
		name    string
		args    args
		want    domain.NoticeMsg
		wantErr bool
	}{
		{
			name: "ID가 1인 사용자 알림 조회 => 정상",
			args: args{
				ctx: context.Background(),
				id:  1,
			},
			want: domain.NoticeMsg{
				Id:              1,
				Title:           "test",
				ContentHtml:     "<p>test</p>",
			NoticeDate:      jsonutils.JSONTime(time.Date(2021, 7, 9, 17, 41, 45, 0, noticeMsgTestJST)),
				NoticeUserCount: 0,
				Created:         jsonutils.JSONTime{},
				NoticeUserType:  domain.NOTICE_NEWSLETTER,
				ProjectId:       0,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := noticeMsgRepository{
				dbFactory: dbFactoryMock{},
			}
			ignoreTime := cmpopts.IgnoreTypes(jsonutils.JSONTime{})

			got, err := repo.GetById(tt.args.ctx, tt.args.id)
			if (err != nil) != tt.wantErr {
				t.Errorf("noticeMsgRepository.GetById() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			assert.Equalf(t, tt.wantErr, err != nil, "noticeMsgRepository.GetById() error = %v, wantErr %v", err, tt.wantErr)
			assert.Truef(t, cmp.Equal(tt.want, got, ignoreTime), "noticeMsgRepository.GetById() got = %v, want %v", got, tt.want)
		})
	}
}

func Test_noticeMsgRepository_Create(t *testing.T) {
	type args struct {
		ctx       context.Context
		noticeMsg domain.NoticeMsg
	}
	tests := []struct {
		name    string
		args    args
		want    int64
		wantErr bool
	}{
		{
			name: "사용자 알림 생성 => 정상",
			args: args{
				ctx: context.Background(),
				noticeMsg: domain.NoticeMsg{
					Title:          "Notice Msg Test",
					ContentHtml:    "<p>Notice Msg Test</p>",
					NoticeDate:     jsonutils.JSONTime(time.Now().Local()),
					NoticeUserType: domain.NOTICE_DIRECT,
					ProjectId:      1,
				},
			},
			want:    13,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := noticeMsgRepository{
				dbFactory: dbFactoryMock{},
			}

			got, err := repo.Create(tt.args.ctx, tt.args.noticeMsg)

			assert.NoErrorf(t, err, "noticeMsgRepository.Create() error = %v", err)
			assert.Equalf(t, tt.want, got, "noticeMsgRepository.Create() got = %v, want %v", got, tt.want)
		})
	}
}
