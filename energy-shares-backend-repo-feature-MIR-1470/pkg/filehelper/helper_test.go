package filehelper

import (
	"testing"

	"github.com/illuminarean/energy-shares-backend-repo/pkg/s3utils"
	"github.com/stretchr/testify/assert"
)

func Test_ExtractFilePath(t *testing.T) {
	type args struct {
		fileURL string
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{
			name: "(성공) fileURL에서 파일 경로를 추출",
			args: args{
				fileURL: "https://cdn.energy-shares.com/attachment_file/2702602/bpbmq9297s524a3jhgn1.pdf",
			},
			want:    "/attachment_file/2702602/bpbmq9297s524a3jhgn1.pdf",
			wantErr: false,
		},
		{
			name: "(成功) パス形式のURLからバケット名を除いたパスを抽出",
			args: args{
				fileURL: "http://localhost:4566/" + s3utils.BUCKET_PRIVATE_NAME +
					"/user_file/2026-08-06/d9q3qn8b1crh9lkldqm0.png",
			},
			want:    "/user_file/2026-08-06/d9q3qn8b1crh9lkldqm0.png",
			wantErr: false,
		},
		{
			name: "(成功) バケット名と同名のディレクトリはCDNのパスとして残す",
			args: args{
				fileURL: "https://cdn.energy-shares.com/attachment_file/" +
					s3utils.BUCKET_PRIVATE_NAME + "/sample.pdf",
			},
			want:    "/attachment_file/" + s3utils.BUCKET_PRIVATE_NAME + "/sample.pdf",
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := fileHelper{}

			got, err := h.ExtractFilePath(tt.args.fileURL)

			assert.Equalf(t, tt.wantErr, err != nil, "ExtractFilePath() error = %v, wantErr %v", err, tt.wantErr)
			assert.Equalf(t, tt.want, got, "ExtractFilePath() = %v, want %v", got, tt.want)
		})
	}
}

func Test_ExtractFileName(t *testing.T) {
	type args struct {
		filePath string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "(성공) filePath 에서 파일 이름을 추출",
			args: args{
				filePath: "/attachment_file/2702602/bpbmq9297s524a3jhgn1.pdf",
			},
			want: "bpbmq9297s524a3jhgn1",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := fileHelper{}

			got := h.ExtractFileName(tt.args.filePath)

			assert.Equalf(t, tt.want, got, "ExtractFileName() = %v, want %v", got, tt.want)
		})
	}
}

func Test_ExtractFileNames(t *testing.T) {
	type args struct {
		filePath []string
	}
	tests := []struct {
		name string
		args args
		want []string
	}{
		{
			name: "(성공) filePath 목록에서 파일 이름 목록을 추출",
			args: args{
				filePath: []string{
					"/attachment_file/2702602/bpbmq9297s524a3jhgn1.pdf",
					"/attachment_file/2702602/test_file.pdf",
				},
			},
			want: []string{
				"bpbmq9297s524a3jhgn1",
				"test_file",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := fileHelper{}

			got := h.ExtractFileNames(tt.args.filePath)

			assert.Equalf(t, tt.want, got, "ExtractFileNames() = %v, want %v", got, tt.want)
		})
	}
}
