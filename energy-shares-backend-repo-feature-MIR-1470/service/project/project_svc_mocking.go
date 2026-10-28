package project

import (
	"time"

	"github.com/illuminarean/energy-shares-backend-repo/domain"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/jsonutils"
)

var (
	mockShare1Id                     = "SHARE-1"
	mockSpv1Id                       = int64(500)
	mockProject1Id                   = int64(1000)
	mockOwner1Id                     = int64(2000)
	mockProjectMainImage1Id          = int64(3000)
	mockProjectNews1Id               = int64(5000)
	mockAttachment1Id                = int64(6000)
	mockResource1Id                  = "87cc7b7f-e1ae-4ce1-aa47-97fe7ec9ca3f"
	mockContractBeforePageId         = int64(7000)
	mockContractAfterPageId          = int64(7001)
	mockSilentPartnershipContractId  = int64(7002)
	mockProjectReportId              = int64(7003)
	mockRepaymentDividendSchedule1Id = int64(8000)
	mockProjectInvestmentExecute1Id  = int64(9000)
	mockProjectMilestone1Id          = int64(10000)
	mockProjectViewCount1Id          = int64(11000)
)

var (
	mockProject1 = domain.Project{
		Id:          mockProject1Id,
		SpvId:       mockSpv1Id,
		MainImageId: mockProjectMainImage1Id,
	}

	mockSpv1 = domain.Spv{
		Id: mockSpv1Id,
	}

	mockShare1 = domain.Share{
		Id:          mockShare1Id,
		OwnerId:     mockOwner1Id,
		ProjectId:   mockProject1Id,
		ProjectType: domain.EQUITY,
	}

	mockProjectMainImage1 = domain.ProjectImage{
		Id:        mockProjectMainImage1Id,
		ProjectId: mockProject1Id,
	}

	mockProjectNews1 = domain.ProjectNews{
		Id:        mockProjectNews1Id,
		ProjectId: mockProject1Id,
	}

	mockAttachment1 = domain.Attachment{
		Id:         mockAttachment1Id,
		ResourceId: mockResource1Id,
	}

	mockContractBeforePage = domain.ProjectFile{
		Id:        mockContractBeforePageId,
		ProjectId: mockProject1Id,
		FileType:  domain.CONTRACT_BEFORE_PAGE,
	}

	mockContractAfterPage = domain.ProjectFile{
		Id:        mockContractAfterPageId,
		ProjectId: mockProject1Id,
		FileType:  domain.CONTRACT_AFTER_PAGE,
	}

	mockSilentPartnershipContract = domain.ProjectFile{
		Id:        mockSilentPartnershipContractId,
		ProjectId: mockProject1Id,
		FileType:  domain.SILENT_PARTNERSHIPS,
	}

	mockProjectReport = domain.ProjectFile{
		Id:        mockProjectReportId,
		ProjectId: mockProject1Id,
		FileType:  domain.PROJECT_REPORT,
	}

	mockRepaymentDividendSchedule1 = domain.RepaymentDividendSchedule{
		Id:          mockRepaymentDividendSchedule1Id,
		ReturnType:  domain.REPAYMENT_RETURN,
		ProjectId:   mockProjectReportId,
		ExecuteDate: jsonutils.JSONTime(time.Unix(0, 0)),
	}

	mockProjectInvestmentExecute1 = domain.ProjectInvestmentExecute{
		Id:        mockProjectInvestmentExecute1Id,
		ProjectId: mockProject1Id,
	}

	mockProjectMilestone1 = domain.ProjectMilestone{
		Id:        mockProjectMilestone1Id,
		ProjectId: mockProject1Id,
	}

	mockProjectViewCount1 = domain.ProjectViewCount{
		Id:        mockProjectViewCount1Id,
		ProjectId: mockProject1Id,
		ViewCount: 100,
	}
)
