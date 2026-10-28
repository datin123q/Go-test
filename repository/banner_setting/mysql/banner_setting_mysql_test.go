package mysql

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/stretchr/testify/assert"

	"github.com/illuminarean/energy-shares-backend-repo/domain"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/jsonutils"
)

func Test_bannerSettingRepository_Create(t *testing.T) {
	type args struct {
		ctx           context.Context
		bannerSetting domain.BannerSetting
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "배너 설정 생성",
			args: args{
				ctx: context.Background(),
				bannerSetting: domain.BannerSetting{
					SettingKey:   "test key",
					SettingValue: "test value",
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := bannerSettingRepository{
				dbFactory: dbFactoryMock{},
			}

			err := repo.Create(tt.args.ctx, tt.args.bannerSetting)
			if (err != nil) != tt.wantErr {
				t.Errorf("BannerSettingRepository.Create() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}

func Test_bannerSettingRepository_GetBySettingKey(t *testing.T) {
	type args struct {
		ctx        context.Context
		settingKey domain.BANNER_SETTING_KEY
	}
	tests := []struct {
		name    string
		args    args
		want    domain.BannerSetting
		wantErr bool
	}{
		{
			name: "배너 설정 조회",
			args: args{
				ctx:        context.Background(),
				settingKey: domain.ROLLING_SECOND,
			},
			want:    mockBannerSetting1,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := bannerSettingRepository{
				dbFactory: dbFactoryMock{},
			}

			ignoreTime := cmpopts.IgnoreTypes(jsonutils.JSONTime{})

			got, err := repo.GetByKey(tt.args.ctx, tt.args.settingKey)
			if (err != nil) != tt.wantErr {
				t.Errorf("BannerSettingRepository.GetBySettingKey() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			assert.Equalf(t, true, cmp.Equal(tt.want, got, ignoreTime), "BannerSettingRepository.GetBySettingKey() diff = %v",
				cmp.Diff(tt.want, got, ignoreTime))
		})
	}
}

func Test_bannerSettingRepository_Update(t *testing.T) {
	type args struct {
		ctx           context.Context
		bannerSetting domain.BannerSetting
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "배너 설정 수정",
			args: args{
				ctx:           context.Background(),
				bannerSetting: mockBannerSetting1,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := bannerSettingRepository{
				dbFactory: dbFactoryMock{},
			}

			err := repo.Update(tt.args.ctx, tt.args.bannerSetting)
			if (err != nil) != tt.wantErr {
				t.Errorf("BannerSettingRepository.Update() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}

func Test_bannerSettingRepository_Delete(t *testing.T) {
	type args struct {
		ctx        context.Context
		settingKey domain.BANNER_SETTING_KEY
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "배너 설정 삭제",
			args: args{
				ctx:        context.Background(),
				settingKey: domain.ROLLING_SECOND,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := bannerSettingRepository{
				dbFactory: dbFactoryMock{},
			}

			err := repo.DeleteByKey(tt.args.ctx, tt.args.settingKey)
			if (err != nil) != tt.wantErr {
				t.Errorf("BannerSettingRepository.Delete() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}
