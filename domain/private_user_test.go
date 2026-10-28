package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/illuminarean/energy-shares-backend-repo/pkg/commoncode"
)

func TestPrivateProfile_GetAssetsCodeGroup(t *testing.T) {
	t.Skip("Skip TestPrivateProfile_GetAssetsCodeGroup")

	type args struct {
		loc string
	}
	tests := []struct {
		name string
		args args
		want commoncode.CodeGroup
	}{
		{
			name: "GetAssetsCodeGroup test",
			args: args{
				loc: "ja",
			},
			want: commoncode.CodeGroup{
				GroupCode: "ASSETS_STATUS",
				Codes: []commoncode.Code{
					{
						Code:  "ASSET1",
						Name:  "150万円未満",
						Order: 1,
					},
					{
						Code:  "ASSET2",
						Name:  "150万円以上 ~ 300万円未満",
						Order: 2,
					},
					{
						Code:  "ASSET3",
						Name:  "300万円以上 ~ 500万円未満",
						Order: 3,
					},
					{
						Code:  "ASSET4",
						Name:  "500万円以上 ~ 700万円未満",
						Order: 4,
					},
					{
						Code:  "ASSET5",
						Name:  "700万円以上 ~ 1,000万円未満",
						Order: 5,
					},
					{
						Code:  "ASSET6",
						Name:  "1,000万円以上 ~ 3,000万円未満",
						Order: 6,
					},
					{
						Code:  "ASSET7",
						Name:  "3,000万円以上",
						Order: 7,
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			group := PrivateProfile{}.GetAssetsCodeGroup(tt.args.loc)
			assert.Equal(t, tt.want, group, "GetAssetsCodeGroup(%v)", tt.args.loc)
		})
	}
}

func TestPrivateProfile_GetsAnnualincomeCodeGroup(t *testing.T) {
	t.Skip("Skip TestPrivateProfile_GetsAnnualincomeCodeGroup")

	type args struct {
		loc string
	}
	tests := []struct {
		name string
		args args
		want commoncode.CodeGroup
	}{
		{
			name: "GetAnnualincomeCodeGroup test",
			args: args{
				loc: "ja",
			},
			want: commoncode.CodeGroup{
				GroupCode: "ANNUALINCOME_STATUS",
				Codes: []commoncode.Code{
					{
						Code:  "ANNUALINCOME1",
						Name:  "300万円未満",
						Order: 1,
					},
					{
						Code:  "ANNUALINCOME2",
						Name:  "300万円以上 ~ 500万円未満",
						Order: 2,
					},
					{
						Code:  "ANNUALINCOME3",
						Name:  "500万円以上 ~ 700万円未満",
						Order: 3,
					},
					{
						Code:  "ANNUALINCOME4",
						Name:  "700万円以上 ~ 1,000万円未満",
						Order: 4,
					},
					{
						Code:  "ANNUALINCOME5",
						Name:  "1,000万円以上 ~ 3,000万円未満",
						Order: 5,
					},
					{
						Code:  "ANNUALINCOME6",
						Name:  "3,000万円以上",
						Order: 6,
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			group := PrivateProfile{}.GetsAnnualincomeCodeGroup(tt.args.loc)
			assert.Equal(t, tt.want, group, "GetsAnnualincomeCodeGroup(%v)", tt.args.loc)
		})
	}
}

func TestPrivateProfile_GetsAnnualincomeCodeGroup_ja(t *testing.T) {
	t.Skip("Skip TestPrivateProfile_GetsAnnualincomeCodeGroup_ja")

	type args struct {
		loc string
	}
	tests := []struct {
		name string
		args args
		want commoncode.CodeGroup
	}{
		{
			name: "GetAnnualincomeCodeGroup test",
			args: args{
				loc: "ja",
			},
			want: commoncode.CodeGroup{
				GroupCode: "ANNUALINCOME_STATUS",
				Codes: []commoncode.Code{
					{
						Code:  "ANNUALINCOME1",
						Name:  "300万円未満",
						Order: 1,
					},
					{
						Code:  "ANNUALINCOME2",
						Name:  "300万円以上 ~ 500万円未満",
						Order: 2,
					},
					{
						Code:  "ANNUALINCOME3",
						Name:  "500万円以上 ~ 700万円未満",
						Order: 3,
					},
					{
						Code:  "ANNUALINCOME4",
						Name:  "700万円以上 ~ 1,000万円未満",
						Order: 4,
					},
					{
						Code:  "ANNUALINCOME5",
						Name:  "1,000万円以上 ~ 3,000万円未満",
						Order: 5,
					},
					{
						Code:  "ANNUALINCOME6",
						Name:  "3,000万円以上",
						Order: 6,
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			group := PrivateProfile{}.GetsAnnualincomeCodeGroup(tt.args.loc)
			assert.Equal(t, tt.want, group, "GetsAnnualincomeCodeGroup(%v)", tt.args.loc)
		})
	}
}

func TestPrivateUser_IsInvestmentSuitability(t *testing.T) {
	t.Skip("Skip TestPrivateUser_IsInvestmentSuitability")

	type args struct {
		asset        ASSETS_STATUS
		annualincome ANNUALINCOME_STATUS
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "failed IsPrivateSuitability ",
			args: args{
				asset:        ASSET1,
				annualincome: ANNUALINCOME1,
			},
			want: false,
		},
		{
			name: "success1 IsPrivateSuitability ",
			args: args{
				asset:        ASSET1,
				annualincome: ANNUALINCOME2,
			},
			want: true,
		},
		{
			name: "success2 InvestmentSuitability ",
			args: args{
				asset:        ASSET2,
				annualincome: ANNUALINCOME1,
			},
			want: true,
		},
		{
			name: "success3 IsPrivateSuitability ",
			args: args{
				asset:        ASSET2,
				annualincome: ANNUALINCOME2,
			},
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			assert.Equalf(t, tt.want, IsPrivateSuitability(tt.args.asset, tt.args.annualincome), "IsPrivateSuitability(%v, %v)", tt.args.asset, tt.args.annualincome)
		})
	}
}
