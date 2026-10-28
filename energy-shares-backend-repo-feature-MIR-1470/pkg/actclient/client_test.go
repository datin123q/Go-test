package actclient

import (
	"testing"
)

func TestMakeUrl(t *testing.T) {
	type args struct {
		companyName     string
		targetId        string
		marketingCode   string
		signUpRequestId string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "MakeUrl",
			args: args{
				companyName:     "companyName",
				targetId:        "targetId",
				marketingCode:   "marketingCode",
				signUpRequestId: "signUpRequestId",
			},
			want: "https://is.accesstrade.net/cgi-bin/isatV2/companyName/isatWeaselV2.cgi?result_id=targetId&verify=signUpRequestId&rk=marketingCode",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := MakeSignUpRewardUrl(tt.args.companyName, tt.args.targetId, tt.args.marketingCode, tt.args.signUpRequestId); got != tt.want {
				t.Errorf("MakeUrl() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestClient_SignUpReward(t *testing.T) {
	type fields struct {
		CompanyName string
		TargetId    string
	}
	type args struct {
		marketingCode   string
		signUpRequestId string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "SignUpReward",
			fields: fields{
				CompanyName: "energyshares",
				TargetId:    "100",
			},
			args: args{
				marketingCode:   "0100plz8000005-4bde5b64ae3bde19ef41d9b0c5b8b205",
				signUpRequestId: "fd304493-21a3-4c8d-9189-342e501e3809",
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := NewClient()
			if err := c.SignUpReward(tt.args.marketingCode, tt.args.signUpRequestId); (err != nil) != tt.wantErr {
				t.Errorf("SignUpReward() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestCompanyRemover(t *testing.T) {
	type args struct {
		marketingCode string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "CompanyRemover",
			args: args{
				marketingCode: "energyshares_0100plz8000005-4bde5b64ae3bde19ef41d9b0c5b8b205",
			},
			want: "0100plz8000005-4bde5b64ae3bde19ef41d9b0c5b8b205",
		},
		{
			name: "CompanyRemover",
			args: args{
				marketingCode: "0100plz8000005-4bde5b64ae3bde19ef41d9b0c5b8b205",
			},
			want: "0100plz8000005-4bde5b64ae3bde19ef41d9b0c5b8b205",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := CompanyRemover(tt.args.marketingCode); got != tt.want {
				t.Errorf("CompanyRemover() = %v, want %v", got, tt.want)
			}
		})
	}
}
