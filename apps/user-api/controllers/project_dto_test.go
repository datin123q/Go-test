package controllers

import (
	"testing"
	"time"

	"github.com/illuminarean/energy-shares-backend-repo/domain"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/jsonutils"
	"github.com/samber/lo"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestProjectItemsResponseDto_New_DividendType(t *testing.T) {
	tests := []struct {
		name             string
		project          *domain.Project
		wantDividendType *domain.DIVIDEND_TYPE
	}{
		{
			name: "(成功) DividendType が nil の場合",
			project: &domain.Project{
				Id:           1,
				Name:         "Test Project",
				DividendType: nil,
			},
			wantDividendType: nil,
		},
		{
			name: "(成功) DividendType が INCOME_AND_CAPITAL_GAIN の場合",
			project: &domain.Project{
				Id:           2,
				Name:         "Test Project 2",
				DividendType: lo.ToPtr(domain.INCOME_AND_CAPITAL_GAIN),
			},
			wantDividendType: lo.ToPtr(domain.INCOME_AND_CAPITAL_GAIN),
		},
		{
			name: "(成功) DividendType が INCOME_GAIN の場合",
			project: &domain.Project{
				Id:           3,
				Name:         "Test Project 3",
				DividendType: lo.ToPtr(domain.INCOME_GAIN),
			},
			wantDividendType: lo.ToPtr(domain.INCOME_GAIN),
		},
		{
			name: "(成功) DividendType が CAPITAL_GAIN の場合",
			project: &domain.Project{
				Id:           4,
				Name:         "Test Project 4",
				DividendType: lo.ToPtr(domain.CAPITAL_GAIN),
			},
			wantDividendType: lo.ToPtr(domain.CAPITAL_GAIN),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dto := ProjectItemsResponseDto{}
			got := dto.New(tt.project)

			require.NotNil(t, got)

			if tt.wantDividendType == nil {
				assert.Nil(t, got.DividendType, "DividendType should be nil")
			} else {
				require.NotNil(t, got.DividendType, "DividendType should not be nil")
				assert.Equal(t, *tt.wantDividendType, *got.DividendType)
			}
		})
	}
}

func TestProjectDetailResponseDto_ToProjectPubResponseDto_DividendType(t *testing.T) {
	tests := []struct {
		name             string
		project          *domain.Project
		wantDividendType *domain.DIVIDEND_TYPE
	}{
		{
			name: "(成功) DividendType が nil の場合",
			project: &domain.Project{
				Id:           1,
				Name:         "Test Project",
				DividendType: nil,
			},
			wantDividendType: nil,
		},
		{
			name: "(成功) DividendType が INCOME_AND_CAPITAL_GAIN の場合",
			project: &domain.Project{
				Id:           2,
				Name:         "Test Project 2",
				DividendType: lo.ToPtr(domain.INCOME_AND_CAPITAL_GAIN),
			},
			wantDividendType: lo.ToPtr(domain.INCOME_AND_CAPITAL_GAIN),
		},
		{
			name: "(成功) DividendType が INCOME_GAIN の場合",
			project: &domain.Project{
				Id:           3,
				Name:         "Test Project 3",
				DividendType: lo.ToPtr(domain.INCOME_GAIN),
			},
			wantDividendType: lo.ToPtr(domain.INCOME_GAIN),
		},
		{
			name: "(成功) DividendType が CAPITAL_GAIN の場合",
			project: &domain.Project{
				Id:           4,
				Name:         "Test Project 4",
				DividendType: lo.ToPtr(domain.CAPITAL_GAIN),
			},
			wantDividendType: lo.ToPtr(domain.CAPITAL_GAIN),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dto := ProjectDetailResponseDto{}
			got := dto.ToProjectPubResponseDto(tt.project)

			require.NotNil(t, got)

			if tt.wantDividendType == nil {
				assert.Nil(t, got.DividendType, "DividendType should be nil")
			} else {
				require.NotNil(t, got.DividendType, "DividendType should not be nil")
				assert.Equal(t, *tt.wantDividendType, *got.DividendType)
			}
		})
	}
}

func TestProjectItemsResponseDto_New_AllFields(t *testing.T) {
	tests := []struct {
		name    string
		project *domain.Project
	}{
		{
			name: "(成功) すべてのフィールドが正しくマッピングされる",
			project: &domain.Project{
				Id:                          1,
				CommonId:                    "common-id-123",
				SpvId:                       10,
				Name:                        "Test Project",
				Description:                 "Description",
				OfferingType:                domain.OFFERING_PUBLIC,
				SelectionType:               domain.PROJECT_SELECTION_LOTTERY,
				StartAt:                     jsonutils.JSONTime{},
				EndAt:                       jsonutils.JSONTime{},
				ProjectType:                 domain.LOAN,
				ProjectStatus:               domain.OPEN_FUNDING,
				Country:                     domain.PROJECT_COUNTRY_JP,
				CategoryType:                domain.SOLAR,
				DividendRate:                5.0,
				OperationTerm:               12,
				CurrentShareAmount:          1000000,
				CurrentShareQty:             100,
				TargetMinShareAmount:        5000000,
				TargetShareQty:              500,
				TargetShareAmount:           10000000,
				Posted:                      true,
				MainImageUrl:                "https://example.com/image.jpg",
				IsWaitInvest:                false,
				IsInvestorLimit:             true,
				Created:                     jsonutils.JSONTime{},
				Updated:                     jsonutils.JSONTime{},
				ProjectProgressPercent:      10,
				RemainingQty:                400,
				CurrentInvestorCount:        50,
				RaisedAmount:                1000000,
				ProjectReportDate:           jsonutils.JSONTime{},
				PublicViewShareAmount:       1000000,
				PublicProgressInvestPercent: 10.5,
				DividendType:                lo.ToPtr(domain.INCOME_AND_CAPITAL_GAIN),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dto := ProjectItemsResponseDto{}
			got := dto.New(tt.project)

			require.NotNil(t, got)
			assert.Equal(t, tt.project.Id, got.Id)
			assert.Equal(t, tt.project.CommonId, got.CommonId)
			assert.Equal(t, tt.project.SpvId, got.SpvId)
			assert.Equal(t, tt.project.Name, got.Name)
			assert.Equal(t, tt.project.Description, got.Description)
			assert.Equal(t, tt.project.OfferingType, got.OfferingType)
			assert.Equal(t, tt.project.SelectionType, got.SelectionType)
			assert.Equal(t, tt.project.ProjectType, got.ProjectType)
			assert.Equal(t, tt.project.ProjectStatus, got.ProjectStatus)
			assert.Equal(t, tt.project.Country, got.Country)
			assert.Equal(t, tt.project.CategoryType, got.CategoryType)
			assert.Equal(t, tt.project.DividendRate, got.DividendRate)
			assert.Equal(t, tt.project.OperationTerm, got.OperationTerm)
			assert.Equal(t, tt.project.Posted, got.Posted)
			assert.Equal(t, tt.project.IsWaitInvest, got.IsWaitInvest)
			assert.Equal(t, tt.project.IsInvestorLimit, got.IsInvestorLimit)
			assert.Equal(t, tt.project.ProjectProgressPercent, got.ProjectProgressPercent)

			// DividendType の確認
			if tt.project.DividendType == nil {
				assert.Nil(t, got.DividendType)
			} else {
				require.NotNil(t, got.DividendType)
				assert.Equal(t, *tt.project.DividendType, *got.DividendType)
			}
		})
	}
}

func TestProjectItemsResponseDto_NewList(t *testing.T) {
	tests := []struct {
		name        string
		projectList []domain.Project
		wantLen     int
	}{
		{
			name: "(成功) 複数プロジェクトのリスト変換",
			projectList: []domain.Project{
				{
					Id:           1,
					Name:         "Project 1",
					DividendType: lo.ToPtr(domain.INCOME_GAIN),
				},
				{
					Id:           2,
					Name:         "Project 2",
					DividendType: nil,
				},
				{
					Id:           3,
					Name:         "Project 3",
					DividendType: lo.ToPtr(domain.CAPITAL_GAIN),
				},
			},
			wantLen: 3,
		},
		{
			name:        "(成功) 空リストの場合",
			projectList: []domain.Project{},
			wantLen:     0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dto := ProjectItemsResponseDto{}
			got := dto.NewList(tt.projectList)

			if tt.wantLen == 0 {
				// 空リストの場合は nil が返される
				assert.Nil(t, got)
			} else {
				assert.NotNil(t, got)
				assert.Len(t, got, tt.wantLen)
			}

			for i, project := range tt.projectList {
				assert.Equal(t, project.Id, got[i].Id)
				assert.Equal(t, project.Name, got[i].Name)

				if project.DividendType == nil {
					assert.Nil(t, got[i].DividendType)
				} else {
					require.NotNil(t, got[i].DividendType)
					assert.Equal(t, *project.DividendType, *got[i].DividendType)
				}
			}
		})
	}
}

func TestProjectItemsResponseDto_New_PlanedOperationDays(t *testing.T) {
	jst, _ := time.LoadLocation("Asia/Tokyo")
	startDate := jsonutils.JSONTime(time.Date(2025, 4, 1, 0, 0, 0, 0, jst))
	endDate := jsonutils.JSONTime(time.Date(2025, 7, 1, 0, 0, 0, 0, jst))

	tests := []struct {
		name     string
		project  *domain.Project
		wantDays int
	}{
		{
			name: "(成功) 両方の日付が設定されている場合、日数が計算される（初日を含む）",
			project: &domain.Project{
				Id:                       1,
				PlanedOperationStartDate: startDate,
				PlanedOperationEndDate:   endDate,
			},
			wantDays: 92,
		},
		{
			name: "(成功) 開始日が未設定の場合、0日",
			project: &domain.Project{
				Id:                     2,
				PlanedOperationEndDate: endDate,
			},
			wantDays: 0,
		},
		{
			name: "(成功) 終了日が未設定の場合、0日",
			project: &domain.Project{
				Id:                       3,
				PlanedOperationStartDate: startDate,
			},
			wantDays: 0,
		},
		{
			name: "(成功) 両方とも未設定の場合、0日",
			project: &domain.Project{
				Id: 4,
			},
			wantDays: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dto := ProjectItemsResponseDto{}
			got := dto.New(tt.project)

			require.NotNil(t, got)
			assert.Equal(t, tt.wantDays, got.PlanedOperationDays)
			assert.Equal(t, tt.project.PlanedOperationStartDate, got.PlanedOperationStartDate)
			assert.Equal(t, tt.project.PlanedOperationEndDate, got.PlanedOperationEndDate)
		})
	}
}

func TestProjectDetailResponseDto_ToProjectPubResponseDto_PlanedOperationDays(t *testing.T) {
	jst, _ := time.LoadLocation("Asia/Tokyo")
	startDate := jsonutils.JSONTime(time.Date(2025, 1, 1, 0, 0, 0, 0, jst))
	endDate := jsonutils.JSONTime(time.Date(2025, 12, 31, 0, 0, 0, 0, jst))

	tests := []struct {
		name     string
		project  *domain.Project
		wantDays int
	}{
		{
			name: "(成功) 両方の日付が設定されている場合、日数が計算される（初日を含む）",
			project: &domain.Project{
				Id:                       1,
				PlanedOperationStartDate: startDate,
				PlanedOperationEndDate:   endDate,
			},
			wantDays: 365,
		},
		{
			name: "(成功) 両方とも未設定の場合、0日",
			project: &domain.Project{
				Id: 2,
			},
			wantDays: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dto := ProjectDetailResponseDto{}
			got := dto.ToProjectPubResponseDto(tt.project)

			require.NotNil(t, got)
			assert.Equal(t, tt.wantDays, got.PlanedOperationDays)
			assert.Equal(t, tt.project.PlanedOperationStartDate, got.PlanedOperationStartDate)
			assert.Equal(t, tt.project.PlanedOperationEndDate, got.PlanedOperationEndDate)
		})
	}
}
