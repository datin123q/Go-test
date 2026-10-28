package domain

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func Test_ExecuteDateIsAfterToday(t *testing.T) {
	tests := []struct {
		name        string
		executeDate string
		want        bool
		wantErr     bool
	}{
		{
			name:        "DateIsBeforeToday",
			executeDate: "2000-01-01",
			want:        false,
			wantErr:     false,
		},
		{
			name:        "DateIsToday",
			executeDate: time.Now().Local().Format(time.DateOnly),
			want:        false,
			wantErr:     false,
		},
		{
			name:        "DateIsInFuture",
			executeDate: "3000-01-01",
			want:        true,
			wantErr:     false,
		},
		{
			name:        "InvalidDateFormat",
			executeDate: "invalid-date",
			want:        false,
			wantErr:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request := ExecuteInvestmentRequest{
				ExecuteInvestmentBody: ExecuteInvestmentBody{
					ExecuteDate: tt.executeDate,
				},
			}

			got, err := request.ExecuteDateIsAfterToday()
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			if !assert.Equal(t, tt.want, got) {
				t.Error("ExecuteDate: ", tt.executeDate)
			}
		})
	}
}
