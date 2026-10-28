package project

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/illuminarean/energy-shares-backend-repo/pkg/jsonutils"

	"github.com/illuminarean/energy-shares-backend-repo/clean-arch-domain/mocks"

	"github.com/illuminarean/energy-shares-backend-repo/domain"
)

func Test_projectService_GetDetail(t *testing.T) {
	type args struct {
		ctx context.Context
		id  int64
	}
	tests := []struct {
		name                                string
		args                                args
		projectRepository                   func() *mocks.ProjectRepository
		projectImageRepository              func() *mocks.ProjectImageRepository
		projectNewsRepository               func() *mocks.ProjectNewsRepository
		projectFileRepository               func() *mocks.ProjectFileRepository
		projectViewCountRepository          func() *mocks.ProjectViewCountRepository
		projectInvestmentExecuteRepository  func() *mocks.ProjectInvestmentExecuteRepository
		projectMilestoneRepository          func() *mocks.ProjectMilestoneRepository
		attachmentRepository                func() *mocks.AttachmentRepository
		repaymentDividendScheduleRepository func() *mocks.RepaymentDividendScheduleRepository
		spvRepository                       func() *mocks.SpvRepository
		want                                domain.Project
		wantErr                             bool
	}{
		{
			name: "プロジェクトの詳細情報を正常に照会します : プロジェクトID : 1000 => 1件が照会されることを期待します。 ",
			args: args{
				ctx: context.Background(),
				id:  mockProject1Id,
			},
			projectRepository: func() *mocks.ProjectRepository {
				repo := &mocks.ProjectRepository{}
				repo.On("GetById", mock.Anything, mockProject1Id).Return(mockProject1, nil)
				return repo
			},

			spvRepository: func() *mocks.SpvRepository {
				repo := &mocks.SpvRepository{}
				repo.On("GetById", mock.Anything, mockSpv1Id).Return(mockSpv1, nil)
				return repo
			},

			projectImageRepository: func() *mocks.ProjectImageRepository {
				repo := &mocks.ProjectImageRepository{}
				repo.On("GetById", mock.Anything, mockProjectMainImage1Id).Return(mockProjectMainImage1, nil)
				return repo
			},
			projectNewsRepository: func() *mocks.ProjectNewsRepository {
				repo := &mocks.ProjectNewsRepository{}
				repo.On("GetAll", mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
				).Return([]domain.ProjectNews{mockProjectNews1}, nil)
				return repo
			},
			attachmentRepository: func() *mocks.AttachmentRepository {
				repo := &mocks.AttachmentRepository{}
				repo.On("GetAll", mock.Anything,
					mock.Anything,
					mock.Anything,
				).Return([]domain.Attachment{mockAttachment1}, nil)
				return repo
			},

			projectFileRepository: func() *mocks.ProjectFileRepository {
				repo := &mocks.ProjectFileRepository{}
				repo.On("GetByProjectIdAndFileType", mock.Anything, mockProject1Id, domain.CONTRACT_BEFORE_PAGE).
					Return(mockContractBeforePage, nil)
				repo.On("GetByProjectIdAndFileType", mock.Anything, mockProject1Id, domain.CONTRACT_AFTER_PAGE).
					Return(mockContractAfterPage, nil)
				repo.On("GetByProjectIdAndFileType", mock.Anything, mockProject1Id, domain.SILENT_PARTNERSHIPS).
					Return(mockSilentPartnershipContract, nil)
				repo.On("GetByProjectIdAndFileType", mock.Anything, mockProject1Id, domain.PROJECT_REPORT).
					Return(mockProjectReport, nil)

				return repo
			},

			repaymentDividendScheduleRepository: func() *mocks.RepaymentDividendScheduleRepository {
				repo := &mocks.RepaymentDividendScheduleRepository{}
				repo.On("GetAll", mock.Anything, mock.Anything).
					Return(int64(1), domain.RepaymentDividendSchedules{mockRepaymentDividendSchedule1}, nil)
				return repo
			},
			projectInvestmentExecuteRepository: func() *mocks.ProjectInvestmentExecuteRepository {
				repo := &mocks.ProjectInvestmentExecuteRepository{}
				repo.On("GetByProjectId", mock.Anything, mockProject1Id).Return(mockProjectInvestmentExecute1, nil)
				return repo
			},
			projectMilestoneRepository: func() *mocks.ProjectMilestoneRepository {
				repo := &mocks.ProjectMilestoneRepository{}
				repo.On("GetAll", mock.Anything, mock.Anything).
					Return([]domain.ProjectMilestone{mockProjectMilestone1}, nil)
				return repo
			},
			projectViewCountRepository: func() *mocks.ProjectViewCountRepository {
				repo := &mocks.ProjectViewCountRepository{}
				repo.On("GetByProjectId", mock.Anything, mockProject1Id).Return(mockProjectViewCount1, nil)
				return repo
			},

			want: domain.Project{
				Id:                         mockProject1Id,
				Spv:                        &mockSpv1,
				SpvId:                      mockSpv1Id,
				MainImageId:                mockProjectMainImage1Id,
				MainImage:                  &mockProjectMainImage1,
				News:                       []*domain.ProjectNews{&mockProjectNews1},
				Attachments:                []*domain.Attachment{&mockAttachment1},
				ContractBeforePage:         &mockContractBeforePage,
				ContractAfterPage:          &mockContractAfterPage,
				SilentPartnerships:         &mockSilentPartnershipContract,
				ProjectReport:              &mockProjectReport,
				RepaymentDividendSchedules: domain.RepaymentDividendSchedules{mockRepaymentDividendSchedule1},
				ProjectInvestmentExecute:   &mockProjectInvestmentExecute1,
				ProjectMilestones:          []domain.ProjectMilestone{mockProjectMilestone1},
				ViewCount:                  mockProjectViewCount1.ViewCount,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := projectService{
				projectRepository:                   tt.projectRepository(),
				projectImageRepository:              tt.projectImageRepository(),
				projectNewsRepository:               tt.projectNewsRepository(),
				projectFileRepository:               tt.projectFileRepository(),
				projectViewCountRepository:          tt.projectViewCountRepository(),
				projectInvestmentExecuteRepository:  tt.projectInvestmentExecuteRepository(),
				projectMilestoneRepository:          tt.projectMilestoneRepository(),
				attachmentRepository:                tt.attachmentRepository(),
				repaymentDividendScheduleRepository: tt.repaymentDividendScheduleRepository(),
				spvRepository:                       tt.spvRepository(),
			}
			got, err := service.GetDetail(tt.args.ctx, tt.args.id)

			ignoreTime := cmpopts.IgnoreTypes(jsonutils.JSONTime{})

			assert.Equalf(t, tt.wantErr, err != nil, "GetDetail() error = %v, wantErr %v", err, tt.wantErr)
			assert.Truef(t, cmp.Equal(got, tt.want, ignoreTime), "GetDetail() = %v, want %v", got, tt.want)
		})
	}
}
