package formutils

import (
	"mime/multipart"
	"reflect"
	"sort"
	"testing"
)

func TestFromToUrls(t *testing.T) {
	type args struct {
		form multipart.Form
		name string
	}
	tests := []struct {
		name string
		args args
		want []string
	}{
		{
			name: "SuccessToGetUrls",
			args: args{

				name: "image",
				form: func() multipart.Form {
					form := multipart.Form{}
					form.Value = map[string][]string{
						"image1": []string{"https://www.google.com/1"},
						"image2": []string{"https://www.google.com/2"},
						"image3": []string{"https://www.google.com/3"},
					}
					return form
				}(),
			},
			want: []string{"https://www.google.com/1", "https://www.google.com/2", "https://www.google.com/3"},
		},
		{
			name: "SuccessToGetUrls2",
			args: args{

				name: "attachment",
				form: func() multipart.Form {
					form := multipart.Form{}
					form.Value = map[string][]string{
						"attachment1": []string{"https://www.google.com/1"},
						"attachment2": []string{"https://www.google.com/2"},
						"attachment3": []string{"https://www.google.com/4"},
					}
					return form
				}(),
			},
			want: []string{"https://www.google.com/1", "https://www.google.com/2", "https://www.google.com/4"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := FromToUrls(tt.args.form, tt.args.name)
			sort.Strings(got)
			if !reflect.DeepEqual(got, tt.want) {

				t.Errorf("FromToUrls() = %v, want %v", got, tt.want)
			}
		})
	}
}
