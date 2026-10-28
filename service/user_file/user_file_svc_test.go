package user_file

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/config"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/filehelper"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/s3utils"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	cdomain "github.com/illuminarean/energy-shares-backend-repo/clean-arch-domain"
	"github.com/illuminarean/energy-shares-backend-repo/clean-arch-domain/mocks"
	"github.com/illuminarean/energy-shares-backend-repo/domain"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/jsonutils"
)

func TestUserFileService_CreateUserFile(t *testing.T) {
	type args struct {
		ctx      context.Context
		userFile domain.UserFile
	}
	tests := []struct {
		name               string
		args               args
		s3pkg              func() *mocks.S3
		userFileRepository func() *mocks.UserFileRepository
		want               domain.UserFile
		wantErr            error
	}{
		{
			name: "(성공) UserFile 모델 생성 -> 정상적으로 생성되기를 기대합니다.",
			args: args{
				ctx:      context.Background(),
				userFile: mockUserFile1,
			},
			s3pkg: func() *mocks.S3 {
				s3pkg := &mocks.S3{}
				s3pkg.On("GetHeadObject", "private-bucket", mockUserFile1.GetKey()).
					Return(s3utils.S3ObjectInfo{
						FileSize:         mockUserFile1.FileSize,
						OriginalFileName: mockUserFile1.OriginalName,
					}, nil)
				return s3pkg
			},
			userFileRepository: func() *mocks.UserFileRepository {
				userRepository := &mocks.UserFileRepository{}
				userRepository.On("Create", context.Background(), mockUserFile1).Return(nil)
				return userRepository
			},
			want:    mockUserFile1,
			wantErr: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := userFileService{
				cfg: &config.Config{
					CommonConfig: config.CommonConfig{
						AWS: struct {
							Region                  string
							SES_Region              string
							SNS_Region              string
							CDN                     string
							S3_PRIVATE_BUCKET       string
							S3_PUBLIC_BUCKET        string
							S3_EKYC_IMAGES_BUCKET   string
							SES_ARN                 string
							SNS_ARN                 string
							PUBLIC_KEY_ID           string
							PRIVATE_KEY_PATH        string
							BASE_ENDPOINT           string
							INVEST_WORKER_QUEUE_URL string
						}{
							S3_PRIVATE_BUCKET: "private-bucket",
						},
					},
				},
				s3pkg:              tt.s3pkg(),
				fileHelper:         filehelper.NewFileHelper(),
				userFileRepository: tt.userFileRepository(),
			}

			got, err := service.CreateUserFile(tt.args.ctx, tt.args.userFile)
			if err != nil && !errors.Is(err, tt.wantErr) {
				t.Errorf("userFileService.CreateUserFile() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			ignoreTime := cmpopts.IgnoreTypes(jsonutils.JSONTime{})
			ignoreFields := cmpopts.IgnoreFields(domain.UserFile{}, "FileUrl")
			if diff := cmp.Diff(tt.want, got, ignoreTime, ignoreFields); diff != "" {
				t.Errorf("userFileService.CreateUserFile() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func Test_userFileService_GetPrivateUserIdentificationFiles(t *testing.T) {
	type args struct {
		ctx    context.Context
		userId int64
	}
	tests := []struct {
		name               string
		args               args
		userFileRepository func() *mocks.UserFileRepository
		want               cdomain.PrivateUserIdentificationFiles
		wantErr            error
	}{
		{
			name: "(성공) UserFile 모델 조회 (Id:1000) -> 정상적으로 조회되기를 기대합니다.",
			args: args{
				ctx:    context.Background(),
				userId: mockUserFile1.UserId,
			},
			userFileRepository: func() *mocks.UserFileRepository {
				userRepository := &mocks.UserFileRepository{}
				userRepository.On("GetAll", context.Background(),
					mock.AnythingOfType("func(*clean_arch_domain.UserFileQueryOptionFields)"),
					mock.AnythingOfType("func(*clean_arch_domain.UserFileQueryOptionFields)"),
					mock.AnythingOfType("func(*clean_arch_domain.UserFileQueryOptionFields)"),
					mock.AnythingOfType("func(*clean_arch_domain.UserFileQueryOptionFields)"),
					mock.AnythingOfType("func(*clean_arch_domain.UserFileQueryOptionFields)")).
					Return(int64(1), mockUserFiles, nil)
				return userRepository
			},
			want:    mockPrivateUserIdentificationFiles,
			wantErr: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := userFileService{
				userFileRepository: tt.userFileRepository(),
			}

			got, err := service.GetPrivateUserIdentificationFiles(tt.args.ctx, tt.args.userId)
			if err != nil && !errors.Is(err, tt.wantErr) {
				t.Errorf("userFileService.GetPrivateUserIdentificationFiles() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			ignoreTime := cmpopts.IgnoreTypes(jsonutils.JSONTime{})
			ignoreFields := cmpopts.IgnoreFields(domain.UserFile{}, "FileUrl")
			if diff := cmp.Diff(tt.want, got, ignoreTime, ignoreFields); diff != "" {
				t.Errorf("userFileService.GetPrivateUserIdentificationFiles() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func Test_userFileService_GetCorporateUserIdentificationFiles(t *testing.T) {
	type args struct {
		ctx    context.Context
		userId int64
	}
	tests := []struct {
		name               string
		args               args
		userFileRepository func() *mocks.UserFileRepository
		want               cdomain.CorporateUserIdentificationFiles
		wantErr            error
	}{
		{
			name: "(성공) UserFile 모델 조회 (Id:1000) -> 정상적으로 조회되기를 기대합니다.",
			args: args{
				ctx:    context.Background(),
				userId: mockUserFile1.UserId,
			},
			userFileRepository: func() *mocks.UserFileRepository {
				userRepository := &mocks.UserFileRepository{}
				userRepository.On("GetAll", context.Background(),
					mock.AnythingOfType("func(*clean_arch_domain.UserFileQueryOptionFields)"),
					mock.MatchedBy(func(f func(*cdomain.UserFileQueryOptionFields)) bool {
						opt := cdomain.UserFileQueryOptionFunc()
						f(&opt)
						return *opt.FileType() == string(domain.IDENTIFICATION1_FRONT)
					}),
					mock.AnythingOfType("func(*clean_arch_domain.UserFileQueryOptionFields)"),
					mock.AnythingOfType("func(*clean_arch_domain.UserFileQueryOptionFields)"),
					mock.AnythingOfType("func(*clean_arch_domain.UserFileQueryOptionFields)"),
				).Return(int64(1), domain.UserFiles{mockIdentification1FrontFile}, nil).
					Once()
				userRepository.On("GetAll", context.Background(),
					mock.AnythingOfType("func(*clean_arch_domain.UserFileQueryOptionFields)"),
					mock.MatchedBy(func(f func(*cdomain.UserFileQueryOptionFields)) bool {
						opt := cdomain.UserFileQueryOptionFunc()
						f(&opt)
						return *opt.FileType() == string(domain.IDENTIFICATION1_BACK)
					}),
					mock.AnythingOfType("func(*clean_arch_domain.UserFileQueryOptionFields)"),
					mock.AnythingOfType("func(*clean_arch_domain.UserFileQueryOptionFields)"),
					mock.AnythingOfType("func(*clean_arch_domain.UserFileQueryOptionFields)"),
				).Return(int64(1), domain.UserFiles{mockIdentification1BackFile}, nil).
					Once()
				userRepository.On("GetAll", context.Background(),
					mock.AnythingOfType("func(*clean_arch_domain.UserFileQueryOptionFields)"),
					mock.MatchedBy(func(f func(*cdomain.UserFileQueryOptionFields)) bool {
						opt := cdomain.UserFileQueryOptionFunc()
						f(&opt)
						return *opt.FileType() == string(domain.IDENTIFICATION2_FRONT)
					}),
					mock.AnythingOfType("func(*clean_arch_domain.UserFileQueryOptionFields)"),
					mock.AnythingOfType("func(*clean_arch_domain.UserFileQueryOptionFields)"),
					mock.AnythingOfType("func(*clean_arch_domain.UserFileQueryOptionFields)"),
				).Return(int64(1), domain.UserFiles{mockIdentification2FrontFile}, nil).
					Once()
				userRepository.On("GetAll", context.Background(),
					mock.AnythingOfType("func(*clean_arch_domain.UserFileQueryOptionFields)"),
					mock.MatchedBy(func(f func(*cdomain.UserFileQueryOptionFields)) bool {
						opt := cdomain.UserFileQueryOptionFunc()
						f(&opt)
						return *opt.FileType() == string(domain.IDENTIFICATION2_BACK)
					}),
					mock.AnythingOfType("func(*clean_arch_domain.UserFileQueryOptionFields)"),
					mock.AnythingOfType("func(*clean_arch_domain.UserFileQueryOptionFields)"),
					mock.AnythingOfType("func(*clean_arch_domain.UserFileQueryOptionFields)"),
				).Return(int64(1), domain.UserFiles{mockIdentification2BAckFile}, nil).
					Once()
				userRepository.On("GetAll", context.Background(),
					mock.AnythingOfType("func(*clean_arch_domain.UserFileQueryOptionFields)"),
					mock.MatchedBy(func(f func(*cdomain.UserFileQueryOptionFields)) bool {
						opt := cdomain.UserFileQueryOptionFunc()
						f(&opt)
						return *opt.FileType() == string(domain.CERTIFIED_COPY)
					}),
					mock.AnythingOfType("func(*clean_arch_domain.UserFileQueryOptionFields)"),
					mock.AnythingOfType("func(*clean_arch_domain.UserFileQueryOptionFields)"),
					mock.AnythingOfType("func(*clean_arch_domain.UserFileQueryOptionFields)"),
				).Return(int64(1), domain.UserFiles{mockCertifiedCopyFile}, nil).
					Once()
				userRepository.On("GetAll", context.Background(),
					mock.AnythingOfType("func(*clean_arch_domain.UserFileQueryOptionFields)"),
					mock.MatchedBy(func(f func(*cdomain.UserFileQueryOptionFields)) bool {
						opt := cdomain.UserFileQueryOptionFunc()
						f(&opt)
						return *opt.FileType() == string(domain.POA)
					}),
					mock.AnythingOfType("func(*clean_arch_domain.UserFileQueryOptionFields)"),
					mock.AnythingOfType("func(*clean_arch_domain.UserFileQueryOptionFields)"),
					mock.AnythingOfType("func(*clean_arch_domain.UserFileQueryOptionFields)"),
				).Return(int64(1), domain.UserFiles{mockPOAFile}, nil).
					Once()
				userRepository.On("GetAll", context.Background(),
					mock.AnythingOfType("func(*clean_arch_domain.UserFileQueryOptionFields)"),
					mock.MatchedBy(func(f func(*cdomain.UserFileQueryOptionFields)) bool {
						opt := cdomain.UserFileQueryOptionFunc()
						f(&opt)
						return *opt.FileType() == string(domain.AOI)
					}),
					mock.AnythingOfType("func(*clean_arch_domain.UserFileQueryOptionFields)"),
					mock.AnythingOfType("func(*clean_arch_domain.UserFileQueryOptionFields)"),
					mock.AnythingOfType("func(*clean_arch_domain.UserFileQueryOptionFields)"),
				).Return(int64(1), domain.UserFiles{mockAOIFile}, nil).
					Once()
				userRepository.On("GetAll", context.Background(),
					mock.AnythingOfType("func(*clean_arch_domain.UserFileQueryOptionFields)"),
					mock.MatchedBy(func(f func(*cdomain.UserFileQueryOptionFields)) bool {
						opt := cdomain.UserFileQueryOptionFunc()
						f(&opt)
						return *opt.FileType() == string(domain.SHAREHOLDERS)
					}),
					mock.AnythingOfType("func(*clean_arch_domain.UserFileQueryOptionFields)"),
					mock.AnythingOfType("func(*clean_arch_domain.UserFileQueryOptionFields)"),
					mock.AnythingOfType("func(*clean_arch_domain.UserFileQueryOptionFields)"),
				).Return(int64(1), domain.UserFiles{mockShareholdersFile}, nil).
					Once()
				userRepository.On("GetAll", context.Background(),
					mock.AnythingOfType("func(*clean_arch_domain.UserFileQueryOptionFields)"),
					mock.MatchedBy(func(f func(*cdomain.UserFileQueryOptionFields)) bool {
						opt := cdomain.UserFileQueryOptionFunc()
						f(&opt)
						return *opt.FileType() == string(domain.USER_FILE_OTHERS)
					}),
					mock.AnythingOfType("func(*clean_arch_domain.UserFileQueryOptionFields)"),
					mock.AnythingOfType("func(*clean_arch_domain.UserFileQueryOptionFields)"),
					mock.AnythingOfType("func(*clean_arch_domain.UserFileQueryOptionFields)"),
				).Return(int64(1), domain.UserFiles{mockOthersFile}, nil).
					Once()
				return userRepository
			},
			want:    mockCorporateUserIdentificationFiles,
			wantErr: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := userFileService{
				userFileRepository: tt.userFileRepository(),
			}

			got, err := service.GetCorporateUserIdentificationFiles(tt.args.ctx, tt.args.userId)
			assert.Equalf(t, tt.wantErr, err, "userFileService.GetCorporateUserIdentificationFiles() error = %v, wantErr %v", err, tt.wantErr)

			ignoreTime := cmpopts.IgnoreTypes(jsonutils.JSONTime{})
			ignoreFields := cmpopts.IgnoreFields(domain.UserFile{}, "FileUrl")
			if diff := cmp.Diff(tt.want, got, ignoreTime, ignoreFields); diff != "" {
				t.Errorf("userFileService.GetCorporateUserIdentificationFiles() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}
