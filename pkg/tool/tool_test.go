package tool

import (
	"reflect"
	"testing"
)

func TestStringToPointer(t *testing.T) {
	type args struct {
		str string
	}
	tests := []struct {
		name string
		args args
		want *string
	}{
		{
			name: "StringToPointer Test",
			args: args{
				str: "test",
			},
			want: func() *string {
				str := "test"
				return &str
			}(),
		},
		{
			name: "StringToPointer Test",
			args: args{
				str: "test",
			},
			want: func() *string {
				str := "test"
				return &str
			}(),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := StringToPointer(tt.args.str); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("StringToPointer() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestBoolToPointer(t *testing.T) {

	type args struct {
		flag bool
	}
	tests := []struct {
		name string
		args args
		want *bool
	}{
		{
			name: "false bool to pointer",
			args: args{
				flag: false,
			},
			want: func() *bool {
				flag := false
				return &flag
			}(),
		},
		{
			name: "true bool to pointer",
			args: args{
				flag: true,
			},
			want: func() *bool {
				flag := true
				return &flag
			}(),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := BoolToPointer(tt.args.flag); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("BoolToPointer() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNoInjection(t *testing.T) {
	type args struct {
		i interface{}
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "struct success test",
			args: args{
				i: struct{}{},
			},
			want: false,
		},
		{
			name: "true test",
			args: args{
				i: nil,
			},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NoInjection(tt.args.i); got != tt.want {
				t.Errorf("IsNil() = %v, want %v", got, tt.want)
			}
		})
	}
}
