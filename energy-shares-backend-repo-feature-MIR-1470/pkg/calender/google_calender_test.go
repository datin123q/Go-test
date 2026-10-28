package calender

import (
	"reflect"
	"testing"
	"time"
)

func Test_CalendarClient_GetPublicHolidays(t *testing.T) {
	t.Skip("Skip TestCalendarClient_GetPublicHolidays")

	startDate, _ := time.Parse(time.RFC3339, "2024-01-01T00:00:00Z")
	endDate, _ := time.Parse(time.RFC3339, "2024-12-31T00:00:00Z")

	type args struct {
		startDate time.Time
		endDate   time.Time
	}
	tests := []struct {
		name    string
		args    args
		want    []GoogleCalenderEvent
		wantErr bool
	}{
		{
			name: "휴일 가져오기 성공",
			args: args{
				startDate: startDate,
				endDate:   endDate,
			},
			want:    []GoogleCalenderEvent{},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := NewCalendarClient()
			got, err := c.GetPublicHolidays(tt.args.startDate, tt.args.endDate)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetPublicHolidays() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetPublicHolidays() got = %v, want %v", got, tt.want)
			}
		})
	}
}
