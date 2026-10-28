package project

import (
	"context"

	"github.com/illuminarean/energy-shares-backend-repo/pkg/config"

	"github.com/pkg/errors"

	cdomain "github.com/illuminarean/energy-shares-backend-repo/clean-arch-domain"
	"github.com/illuminarean/energy-shares-backend-repo/domain"
)

type projectService struct {
	cfg                                 *config.Config
	shareRepository                     cdomain.ShareRepository
	investmentRepository                cdomain.InvestmentRepository
	projectRepository                   cdomain.ProjectRepository
	projectImageRepository              cdomain.ProjectImageRepository
	projectNewsRepository               cdomain.ProjectNewsRepository
	projectFileRepository               cdomain.ProjectFileRepository
	projectViewCountRepository          cdomain.ProjectViewCountRepository
	projectInvestmentExecuteRepository  cdomain.ProjectInvestmentExecuteRepository
	projectMilestoneRepository          cdomain.ProjectMilestoneRepository
	attachmentRepository                cdomain.AttachmentRepository
	repaymentDividendScheduleRepository cdomain.RepaymentDividendScheduleRepository
	spvRepository                       cdomain.SpvRepository
	mailService                         cdomain.MailService
}

func New(
	cfg *config.Config,
	shareRepository cdomain.ShareRepository,
	investmentRepository cdomain.InvestmentRepository,
	projectRepository cdomain.ProjectRepository,
	projectImageRepository cdomain.ProjectImageRepository,
	projectNewsRepository cdomain.ProjectNewsRepository,
	projectFileRepository cdomain.ProjectFileRepository,
	projectViewCountRepository cdomain.ProjectViewCountRepository,
	projectInvestmentExecuteRepository cdomain.ProjectInvestmentExecuteRepository,
	projectMilestoneRepository cdomain.ProjectMilestoneRepository,
	attachmentRepository cdomain.AttachmentRepository,
	repaymentDividendScheduleRepository cdomain.RepaymentDividendScheduleRepository,
	spvRepository cdomain.SpvRepository,
	mailService cdomain.MailService,
) cdomain.ProjectService {
	return &projectService{
		cfg:                                 cfg,
		shareRepository:                     shareRepository,
		investmentRepository:                investmentRepository,
		projectRepository:                   projectRepository,
		projectImageRepository:              projectImageRepository,
		projectNewsRepository:               projectNewsRepository,
		projectFileRepository:               projectFileRepository,
		projectViewCountRepository:          projectViewCountRepository,
		projectInvestmentExecuteRepository:  projectInvestmentExecuteRepository,
		projectMilestoneRepository:          projectMilestoneRepository,
		attachmentRepository:                attachmentRepository,
		repaymentDividendScheduleRepository: repaymentDividendScheduleRepository,
		spvRepository:                       spvRepository,
		mailService:                         mailService,
	}
}

func (p projectService) GetDetail(ctx context.Context, id int64) (domain.Project, error) {
	project, err := p.projectRepository.GetById(ctx, id)
	if err != nil {
		return domain.Project{}, errors.Wrap(err, "err get project")
	}

	// SPV単件照会
	spv, err := p.spvRepository.GetById(ctx, project.SpvId)
	if err != nil && !errors.Is(err, cdomain.ErrItemNotFound) {
		return domain.Project{}, errors.Wrap(err, "err get spv")
	}

	// メイン画像が存在するかどうかを照会
	mainImage, err := p.projectImageRepository.GetById(ctx, project.MainImageId)
	if err != nil && !errors.Is(err, cdomain.ErrItemNotFound) {
		return domain.Project{}, errors.Wrap(err, "err get main image")
	}

	projectNews, err := p.projectNewsRepository.GetAll(ctx,
		cdomain.ProjectQueryOptionFunc().WithProjectId(project.Id),
		cdomain.ProjectQueryOptionFunc().WithSortBy("created"),
		cdomain.ProjectQueryOptionFunc().WithOrder("asc"),
		cdomain.ProjectQueryOptionFunc().WithLimit(30))
	if err != nil && !errors.Is(err, cdomain.ErrItemNotFound) {
		return domain.Project{}, errors.Wrap(err, "err get project news")
	}

	attachments, err := p.attachmentRepository.GetAll(ctx,
		cdomain.AttachmentQueryOptionFunc().WithResourceId(project.CommonId),
		cdomain.AttachmentQueryOptionFunc().WithLimit(10))

	contractBeforePage, err := p.projectFileRepository.GetByProjectIdAndFileType(ctx, project.Id, domain.CONTRACT_BEFORE_PAGE)
	if err != nil && !errors.Is(err, cdomain.ErrItemNotFound) {
		return domain.Project{}, errors.Wrap(err, "err get contract before page")
	}

	contractAfterPage, err := p.projectFileRepository.GetByProjectIdAndFileType(ctx, project.Id, domain.CONTRACT_AFTER_PAGE)
	if err != nil && !errors.Is(err, cdomain.ErrItemNotFound) {
		return domain.Project{}, errors.Wrap(err, "err get contract after page")
	}

	silentPartnerships, err := p.projectFileRepository.GetByProjectIdAndFileType(ctx, project.Id, domain.SILENT_PARTNERSHIPS)
	if err != nil && !errors.Is(err, cdomain.ErrItemNotFound) {
		return domain.Project{}, errors.Wrap(err, "err get silent partnerships")
	}

	projectReport, err := p.projectFileRepository.GetByProjectIdAndFileType(ctx, project.Id, domain.PROJECT_REPORT)
	if err != nil && !errors.Is(err, cdomain.ErrItemNotFound) {
		return domain.Project{}, errors.Wrap(err, "err get project report")
	}

	_, repaymentDividendSchedules, err := p.repaymentDividendScheduleRepository.GetAll(ctx,
		cdomain.RepaymentDividendQueryOptionFunc().WithProjectId(project.Id))
	if err != nil && !errors.Is(err, cdomain.ErrItemNotFound) {
		return domain.Project{}, errors.Wrap(err, "err get repayment dividend schedule")
	}

	projectInvestmentExecute, err := p.projectInvestmentExecuteRepository.GetByProjectId(ctx, project.Id)
	if err != nil && !errors.Is(err, cdomain.ErrItemNotFound) {
		return domain.Project{}, errors.Wrap(err, "err get project investment execute")
	}

	projectMilestones, err := p.projectMilestoneRepository.GetAll(ctx,
		cdomain.ProjectQueryOptionFunc().WithProjectId(project.Id),
	)
	if err != nil && !errors.Is(err, cdomain.ErrItemNotFound) {
		return domain.Project{}, errors.Wrap(err, "err get project milestones")
	}

	projectViewCount, err := p.projectViewCountRepository.GetByProjectId(ctx, project.Id)
	if errors.Is(err, cdomain.ErrItemNotFound) {
		if err = p.projectViewCountRepository.Create(ctx, domain.ProjectViewCount{ProjectId: project.Id}); err != nil {
			return domain.Project{}, errors.Wrap(err, "err create project view count")
		}
	}

	if err != nil {
		return domain.Project{}, errors.Wrap(err, "err get project view count")
	}

	return project.WithSpv(spv).WithMainImage(mainImage).
		WithProjectNews(projectNews).
		WithAttachments(attachments).
		WithContractBeforePage(contractBeforePage).
		WithContractAfterPage(contractAfterPage).
		WithSilentPartnerships(silentPartnerships).
		WithProjectReport(projectReport).
		WithRepaymentDividendSchedules(repaymentDividendSchedules).
		WithProjectInvestmentExecute(projectInvestmentExecute).
		WithProjectMilestones(projectMilestones).
		WithProjectViewCount(projectViewCount.ViewCount), nil
}

