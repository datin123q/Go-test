package jsonutils

import (
	"testing"
)

func TestMarshalString(t *testing.T) {
	type args struct {
		v any
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{
			args: args{
				v: struct {
					A string `json:"a"`
					B string `json:"b"`
				}{
					A: "a",
					B: "b",
				},
			},
			want:    `{"a":"a","b":"b"}`,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := MarshalString(tt.args.v)
			if (err != nil) != tt.wantErr {
				t.Errorf("MarshalString() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("MarshalString() got = %v, want %v", got, tt.want)
			}
		})
	}
}
