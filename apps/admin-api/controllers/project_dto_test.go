package controllers

import (
	"strings"
	"testing"

	"github.com/illuminarean/energy-shares-backend-repo/domain"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/jsonutils"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/validator"
	"github.com/samber/lo"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestProjectInput_ToModel_DividendType(t *testing.T) {
	tests := []struct {
		name             string
		input            ProjectInput
		wantDividendType *domain.DIVIDEND_TYPE
		wantErr          bool
		skipValidation   bool // バリデーションをスキップするかどうか
	}{
		{
			name: "(成功) dividend_type が nil の場合",
			input: ProjectInput{
				Name:                    "Test Project",
				CategoryType:            domain.SOLAR,
				Country:                 domain.PROJECT_COUNTRY_JP,
				ProjectType:             domain.LOAN,
				ProjectStatus:           domain.STANDBY_FUNDING,
				Capacity:                "1000kW",
				DividendRate:            5.0,
				OperationTerm:           12,
				SharePrice:              10000,
				Spv:                     1,
				StartAt:                 "2024-01-01 00:00:00",
				EndAt:                   "2024-12-31 23:59:59",
				DividendType:            nil,
				IsVisibleInvestmentInfo: false,
			},
			wantDividendType: nil,
			wantErr:          false,
		},
		{
			name: "(成功) dividend_type が INCOME_AND_CAPITAL_GAIN の場合",
			input: ProjectInput{
				Name:                    "Test Project",
				CategoryType:            domain.SOLAR,
				Country:                 domain.PROJECT_COUNTRY_JP,
				ProjectType:             domain.LOAN,
				ProjectStatus:           domain.STANDBY_FUNDING,
				Capacity:                "1000kW",
				DividendRate:            5.0,
				OperationTerm:           12,
				SharePrice:              10000,
				Spv:                     1,
				StartAt:                 "2024-01-01 00:00:00",
				EndAt:                   "2024-12-31 23:59:59",
				DividendType:            lo.ToPtr(domain.INCOME_AND_CAPITAL_GAIN),
				IsVisibleInvestmentInfo: false,
			},
			wantDividendType: lo.ToPtr(domain.INCOME_AND_CAPITAL_GAIN),
			wantErr:          false,
		},
		{
			name: "(成功) dividend_type が INCOME_GAIN の場合",
			input: ProjectInput{
				Name:                    "Test Project",
				CategoryType:            domain.SOLAR,
				Country:                 domain.PROJECT_COUNTRY_JP,
				ProjectType:             domain.LOAN,
				ProjectStatus:           domain.STANDBY_FUNDING,
				Capacity:                "1000kW",
				DividendRate:            5.0,
				OperationTerm:           12,
				SharePrice:              10000,
				Spv:                     1,
				StartAt:                 "2024-01-01 00:00:00",
				EndAt:                   "2024-12-31 23:59:59",
				DividendType:            lo.ToPtr(domain.INCOME_GAIN),
				IsVisibleInvestmentInfo: false,
			},
			wantDividendType: lo.ToPtr(domain.INCOME_GAIN),
			wantErr:          false,
		},
		{
			name: "(成功) dividend_type が CAPITAL_GAIN の場合",
			input: ProjectInput{
				Name:                    "Test Project",
				CategoryType:            domain.SOLAR,
				Country:                 domain.PROJECT_COUNTRY_JP,
				ProjectType:             domain.LOAN,
				ProjectStatus:           domain.STANDBY_FUNDING,
				Capacity:                "1000kW",
				DividendRate:            5.0,
				OperationTerm:           12,
				SharePrice:              10000,
				Spv:                     1,
				StartAt:                 "2024-01-01 00:00:00",
				EndAt:                   "2024-12-31 23:59:59",
				DividendType:            lo.ToPtr(domain.CAPITAL_GAIN),
				IsVisibleInvestmentInfo: false,
			},
			wantDividendType: lo.ToPtr(domain.CAPITAL_GAIN),
			wantErr:          false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// バリデーションのテストはスキップ（バリデーションは HTTP レイヤーで行われる）
			got, err := tt.input.ToModel()

			if tt.wantErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.NotNil(t, got)

			if tt.wantDividendType == nil {
				assert.Nil(t, got.DividendType, "DividendType should be nil")
			} else {
				require.NotNil(t, got.DividendType, "DividendType should not be nil")
				assert.Equal(t, *tt.wantDividendType, *got.DividendType)
			}
		})
	}
}

func TestProjectAdminResponseDto_New_DividendType(t *testing.T) {
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
			dto := ProjectAdminResponseDto{}
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

func TestProjectMilestoneInput_ToModel(t *testing.T) {
	tests := []struct {
		name    string
		input   ProjectMilestoneInput
		wantErr bool
	}{
		{
			name: "(成功) 正常なマイルストーン",
			input: ProjectMilestoneInput{
				MilestoneStatus:     "MILESTONE_READY",
				MilestoneInfo:       "着工準備",
				MilestoneAssignment: 10,
				MilestoneNote:       "初期段階",
				MilestoneNo:         1,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.input.ToModel()

			if tt.wantErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.NotNil(t, got)
			assert.Equal(t, domain.MILESTONE_STATUS(tt.input.MilestoneStatus), got.MilestoneStatus)
			assert.Equal(t, tt.input.MilestoneInfo, got.MilestoneInfo)
			assert.Equal(t, tt.input.MilestoneAssignment, got.MilestoneAssignment)
			assert.Equal(t, tt.input.MilestoneNote, got.MilestoneNote)
			assert.Equal(t, tt.input.MilestoneNo, got.MilestoneNo)
		})
	}
}

func TestProjectSearchSimpleOutput_New(t *testing.T) {
	tests := []struct {
		name     string
		project  *domain.Project
		wantId   int64
		wantName string
	}{
		{
			name: "(成功) プロジェクト情報の変換",
			project: &domain.Project{
				Id:   100,
				Name: "Sample Project",
			},
			wantId:   100,
			wantName: "Sample Project",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output := ProjectSearchSimpleOutput{}
			got := output.New(tt.project)

			require.NotNil(t, got)
			assert.Equal(t, tt.wantId, got.Id)
			assert.Equal(t, tt.wantName, got.Name)
		})
	}
}

func TestProjectSearchSimpleOutput_NewList(t *testing.T) {
	tests := []struct {
		name        string
		projectList []domain.Project
		wantLen     int
	}{
		{
			name: "(成功) 複数プロジェクトのリスト変換",
			projectList: []domain.Project{
				{Id: 1, Name: "Project 1"},
				{Id: 2, Name: "Project 2"},
				{Id: 3, Name: "Project 3"},
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
			output := ProjectSearchSimpleOutput{}
			got := output.NewList(tt.projectList)

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
			}
		})
	}
}

func TestProjectAdminResponseDto_New_AllFields(t *testing.T) {
	tests := []struct {
		name    string
		project *domain.Project
	}{
		{
			name: "(成功) すべてのフィールドが正しくマッピングされる",
			project: &domain.Project{
				Id:                      1,
				CommonId:                "common-id-123",
				SpvId:                   10,
				Name:                    "Test Project",
				CategoryType:            domain.SOLAR,
				SelectionType:           domain.PROJECT_SELECTION_LOTTERY,
				AdminNote:               "Admin note",
				ProjectInfoHtml:         "<p>Project info</p>",
				CompanyInfoHtml:         "<p>Company info</p>",
				Description:             "Description",
				StartAt:                 jsonutils.JSONTime{},
				EndAt:                   jsonutils.JSONTime{},
				ProjectType:             domain.LOAN,
				ProjectStatus:           domain.OPEN_FUNDING,
				OwnerId:                 100,
				UpdateId:                101,
				Capacity:                "1000kW",
				Country:                 domain.PROJECT_COUNTRY_JP,
				DividendRate:            5.0,
				OperationTerm:           12,
				CurrentShareAmount:      1000000,
				CurrentShareQty:         100,
				MinShareQty:             1,
				SharePrice:              10000,
				TargetMinShareAmount:    5000000,
				TargetShareQty:          500,
				TargetShareAmount:       10000000,
				Posted:                  true,
				ViewCount:               1000,
				BlockInvest:             false,
				MainImageId:             1,
				MainImageUrl:            "https://example.com/image.jpg",
				IsWaitInvest:            false,
				IsManualInvest:          false,
				IsInvestorLimit:         true,
				IsAmountLimit:           false,
				IsVisibleInvestmentInfo: true,
				TransactionCount:        50,
				CautionInfoHtml:         "<p>Caution</p>",
				BankName:                "Test Bank",
				BranchName:              "Test Branch",
				BankCode:                "1234",
				BranchCode:              "567",
				BankAccount:             "1234567890",
				Depositor:               "Test Depositor",
				RemainingQty:            400,
				CurrentInvestorCount:    50,
				RaisedAmount:            1000000,
				InvestWaitRemainingQty:  400,
				PublicViewShareAmount:   1000000,
				ProjectProgressPercent:  10,
				DividendType:            lo.ToPtr(domain.INCOME_AND_CAPITAL_GAIN),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dto := ProjectAdminResponseDto{}
			got := dto.New(tt.project)

			require.NotNil(t, got)
			assert.Equal(t, tt.project.Id, got.Id)
			assert.Equal(t, tt.project.CommonId, got.CommonId)
			assert.Equal(t, tt.project.SpvId, got.SpvId)
			assert.Equal(t, tt.project.Name, got.Name)
			assert.Equal(t, tt.project.CategoryType, got.CategoryType)
			assert.Equal(t, tt.project.SelectionType, got.SelectionType)
			assert.Equal(t, tt.project.AdminNote, got.AdminNote)
			assert.Equal(t, tt.project.ProjectType, got.ProjectType)
			assert.Equal(t, tt.project.ProjectStatus, got.ProjectStatus)
			assert.Equal(t, tt.project.Country, got.Country)
			assert.Equal(t, tt.project.DividendRate, got.DividendRate)
			assert.Equal(t, tt.project.OperationTerm, got.OperationTerm)
			assert.Equal(t, tt.project.Posted, got.Posted)
			assert.Equal(t, tt.project.BlockInvest, got.BlockInvest)
			assert.Equal(t, tt.project.IsInvestorLimit, got.IsInvestorLimit)
			assert.Equal(t, tt.project.IsAmountLimit, got.IsAmountLimit)
			assert.Equal(t, tt.project.IsVisibleInvestmentInfo, got.IsVisibleInvestmentInfo)
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

func TestParseFundNumber(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    *int64
		wantErr bool
	}{
		{
			name:  "(成功) 未入力は nil を返す",
			input: "",
			want:  nil,
		},
		{
			name:  "(成功) 半角数字を数値へ変換する",
			input: "12345",
			want:  lo.ToPtr(int64(12345)),
		},
		{
			name:  "(成功) 上限の12桁を受け付ける",
			input: "999999999999",
			want:  lo.ToPtr(int64(999999999999)),
		},
		{
			name:  "(成功) 先頭ゼロは数値として解釈する",
			input: "000123",
			want:  lo.ToPtr(int64(123)),
		},
		{
			// 更新時に Nullable 指定しており、0 は NULL に置き換わって消えるため。
			name:    "(失敗) 0 は受け付けない",
			input:   "0",
			wantErr: true,
		},
		{
			name:    "(失敗) ゼロ埋めだけも受け付けない",
			input:   "000",
			wantErr: true,
		},
		{
			name:    "(失敗) 13桁は桁数超過",
			input:   "1234567890123",
			wantErr: true,
		},
		{
			name:    "(失敗) 全角数字は受け付けない",
			input:   "１２３",
			wantErr: true,
		},
		{
			name:    "(失敗) 英字混じりは受け付けない",
			input:   "12a34",
			wantErr: true,
		},
		{
			name:    "(失敗) 符号付きは受け付けない",
			input:   "+123",
			wantErr: true,
		},
		{
			name:    "(失敗) 負数は受け付けない",
			input:   "-123",
			wantErr: true,
		},
		{
			name:    "(失敗) 空白混じりは受け付けない",
			input:   " 123",
			wantErr: true,
		},
		{
			name:    "(失敗) 小数は受け付けない",
			input:   "1.5",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseFundNumber(tt.input)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, got)
				return
			}

			require.NoError(t, err)
			if tt.want == nil {
				assert.Nil(t, got)
				return
			}
			require.NotNil(t, got)
			assert.Equal(t, *tt.want, *got)
		})
	}
}

func TestProjectInput_ToModel_FundNumber(t *testing.T) {
	baseInput := func() ProjectInput {
		return ProjectInput{
			Name:          "Test Project",
			CategoryType:  domain.SOLAR,
			Country:       domain.PROJECT_COUNTRY_JP,
			ProjectType:   domain.LOAN,
			ProjectStatus: domain.STANDBY_FUNDING,
			Capacity:      "1000kW",
			DividendRate:  5.0,
			OperationTerm: 12,
			SharePrice:    10000,
			Spv:           1,
		}
	}

	t.Run("(成功) 未入力は nil のまま保存される", func(t *testing.T) {
		input := baseInput()
		input.FundNumber = ""

		got, err := input.ToModel()

		require.NoError(t, err)
		require.NotNil(t, got)
		assert.Nil(t, got.FundNumber)
	})

	t.Run("(成功) 入力値が数値として保存される", func(t *testing.T) {
		input := baseInput()
		input.FundNumber = "202600123"

		got, err := input.ToModel()

		require.NoError(t, err)
		require.NotNil(t, got)
		require.NotNil(t, got.FundNumber)
		assert.Equal(t, int64(202600123), *got.FundNumber)
	})

	t.Run("(失敗) 数字以外が含まれるとエラーになる", func(t *testing.T) {
		input := baseInput()
		input.FundNumber = "ABC"

		got, err := input.ToModel()

		assert.Error(t, err)
		assert.Nil(t, got)
	})
}

// 口座番号は金融機関を問わず7桁。ゆうちょも通帳の番号(8桁)ではなく振込用の
// 7桁で登録する。ファンドは口座が未設定のまま編集することがあるため、
// 空は通し、入力があるときだけ桁数を求める。
//
// ProjectInput は他にも必須項目が多いため、構造体全体の可否ではなく
// 口座番号のエラーが出たかどうかで判定する。
func TestProjectInput_BankAccountValidation(t *testing.T) {
	tests := []struct {
		name        string
		bankAccount string
		wantErr     bool
	}{
		{
			name:        "(成功) 7桁は通る",
			bankAccount: "1234567",
			wantErr:     false,
		},
		{
			name:        "(成功) 口座未設定を想定し空は通る",
			bankAccount: "",
			wantErr:     false,
		},
		{
			name:        "(失敗) ゆうちょの通帳番号8桁は弾く",
			bankAccount: "12345678",
			wantErr:     true,
		},
		{
			name:        "(失敗) 数字以外を含むものは弾く",
			bankAccount: "123456a",
			wantErr:     true,
		},
	}

	cv := validator.NewCustomValidator()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			input := ProjectInput{BankAccount: tt.bankAccount}

			err := cv.Validate(&input)

			hasBankAccountError := err != nil &&
				strings.Contains(err.Error(), "ProjectInput.BankAccount")

			assert.Equal(t, tt.wantErr, hasBankAccountError)
		})
	}
}

func TestExpectedDetailVersion(t *testing.T) {
	tests := []struct {
		name    string
		loaded  string
		force   bool
		want    *int64
		wantErr bool
	}{
		{name: "(判定する) 開いた時点の値", loaded: "3", want: lo.ToPtr(int64(3))},
		{name: "(判定する) 列を足してから一度も保存されていないファンド(0)", loaded: "0", want: lo.ToPtr(int64(0))},
		{name: "(判定しない) 版を送らない呼び出し元", loaded: ""},
		{name: "(判定しない) 上書きを確認済み", loaded: "3", force: true},
		{name: "(判定しない) 上書きを確認済みなら、値が不正でもエラーにしない", loaded: "abc", force: true},
		{name: "(エラー) 数値でない", loaded: "abc", wantErr: true},
		{name: "(エラー) 負の値", loaded: "-1", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ExpectedDetailVersion(tt.loaded, tt.force)
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestIsDetailVersionConflicted(t *testing.T) {
	tests := []struct {
		name     string
		expected *int64
		current  int64
		want     bool
	}{
		{name: "(競合なし) 開いた時点から誰も保存していない", expected: lo.ToPtr(int64(3)), current: 3},
		{name: "(競合) 開いた後に別の保存があった", expected: lo.ToPtr(int64(3)), current: 4, want: true},
		{name: "(競合) 開いた時点の値が今より大きい(想定外の状態は上書きせず知らせる)", expected: lo.ToPtr(int64(5)), current: 3, want: true},
		{name: "(競合なし) 判定しない場合(nil)", expected: nil, current: 4},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, IsDetailVersionConflicted(tt.expected, tt.current))
		})
	}
}
