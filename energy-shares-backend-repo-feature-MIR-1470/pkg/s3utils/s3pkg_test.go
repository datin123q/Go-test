package s3utils

import (
	"mime"
	"path"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_ExtractType(t *testing.T) {
	type args struct {
		fileName string
	}

	tests := []struct {
		name string
		args args
		want string
	}{
		// jpeg
		{
			name: "fileName : gildong.jpeg",
			args: args{
				fileName: "gildong.jpeg",
			},
			want: "image/jpeg",
		},

		// jpg
		{
			name: "fileName : gildong.jpg",
			args: args{
				fileName: "gildong.jpg",
			},
			want: "image/jpeg",
		},
		{
			name: "fileName : go.gildong.jpg",
			args: args{
				fileName: "go.gildong.jpg",
			},
			want: "image/jpeg",
		},
		// png
		{
			name: "fileName : gildong.png",
			args: args{
				fileName: "gildong.png",
			},
			want: "image/png",
		},
		// gif
		{
			name: "fileName : gildong.gif",
			args: args{
				fileName: "gildong.gif",
			},
			want: "image/gif",
		},
		// pdf
		{
			name: "fileName : gildong.pdf",
			args: args{
				fileName: "gildong.pdf",
			},
			want: "application/pdf",
		},
		// octet-stream
		{
			name: "fileName : gildong",
			args: args{
				fileName: "gildong",
			},
			want: "application/octet-stream",
		},
		{
			name: "fileName : gildong",
			args: args{
				fileName: "gildong",
			},
			want: "application/octet-stream",
		},
	}
	for _, tt := range tests {
		contentType := mime.TypeByExtension(path.Ext(tt.args.fileName))

		if contentType == "" {
			contentType = "application/octet-stream"
		}

		assert.Equal(t, tt.want, contentType)
	}
}

func Test_extractFilenameFromContentDisposition(t *testing.T) {
	type args struct {
		contentDisposition string
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{
			name: "Content-Disposition with attachment and valid filename(성공)",
			args: args{
				contentDisposition: "attachment; filename=\"gildong.jpg\"",
			},
			want:    "gildong.jpg",
			wantErr: false,
		},
		{
			name: "Content-Disposition with URL-encoded filename(성공)",
			args: args{
				contentDisposition: "attachment; filename=\"8026_%E1%84%8B%E1%85%A9%E1%84%89%E1%85%A9%E1%84%85%E1%85%A6%E1%84%89%E1%85%A1%E1%86%AB %E1%84%87%E1%85%A9%E1%86%A8%E1%84%89%E1%85%A1%E1%84%87%E1%85%A9%E1%86%AB 4.jpg\"",
			},
			want:    "8026_오소레산 복사본 4.jpg",
			wantErr: false,
		},
		{
			name: "Content-Disposition with inline(실패: no filename)",
			args: args{
				contentDisposition: "inline",
			},
			want:    "",
			wantErr: true,
		},
		{
			name: "Content-Disposition with no filename parameter(실패)",
			args: args{
				contentDisposition: "attachment",
			},
			want:    "",
			wantErr: true,
		},
		{
			name: "Content-Disposition is empty(실패)",
			args: args{
				contentDisposition: "",
			},
			want:    "",
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := extractFilenameFromContentDisposition(tt.args.contentDisposition)
			if tt.wantErr {
				assert.Error(t, err, "Expected an error but got none")
			} else {
				assert.NoError(t, err, "Expected no error but got one")
				assert.Equalf(t, tt.want, got, "extractFilenameFromContentDisposition(%v)", tt.args.contentDisposition)
			}
		})
	}
}
