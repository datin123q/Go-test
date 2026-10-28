package investor_report

import (
	"context"

	"github.com/pkg/errors"

	adomain "github.com/illuminarean/energy-share-core-api/domain"
	cdomain "github.com/illuminarean/energy-shares-backend-repo/clean-arch-domain"
	"github.com/illuminarean/energy-shares-backend-repo/domain"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/s3utils"
)

type investorReportUsecase struct {
	s3pkg                  s3utils.S3
	reportOutputRepository cdomain.ReportOutputRepository
}

func New(
	s3pkg s3utils.S3,
	reportOutputRepository cdomain.ReportOutputRepository,
) adomain.InvestorReportUsecase {
	return &investorReportUsecase{
		s3pkg:                  s3pkg,
		reportOutputRepository: reportOutputRepository,
	}
}

func (u *investorReportUsecase) GetAllPaymentStatement(ctx context.Context, req adomain.InvestorReportGetAllRequest) (adomain.InvestorReportGetAllResponse, error) {
	// 1. 상세한 비즈니스 로직중 하나라는 생각에 프런트엔드로부터 숨김
	// 2. 투자자 지불조서를 위한 질의에 필요한 필터링 정보를 따로 상수로 빼놓아 가독성을 높임
	var (
		submissionType = domain.SUBMISSION_INVESTORS
		reportTypes    = []domain.REPORT_TYPE{domain.REPORT_PAYMENT_STATEMENT}
		groupBy        = "year"
	)

	total, items, err := u.reportOutputRepository.GetAll(ctx,
		cdomain.ReportOutputQueryOptionFunc().WithSubmissionType(submissionType),
		cdomain.ReportOutputQueryOptionFunc().WithReportTypes(reportTypes...),
		cdomain.ReportOutputQueryOptionFunc().WithGroupBy(groupBy),
		cdomain.ReportOutputQueryOptionFunc().WithSortBy(req.Sortby),
		cdomain.ReportOutputQueryOptionFunc().WithOrder(req.Order),
		cdomain.ReportOutputQueryOptionFunc().WithOffset(req.SkipCount),
		cdomain.ReportOutputQueryOptionFunc().WithLimit(req.MaxResultCount),
	)
	if err != nil {
		return adomain.InvestorReportGetAllResponse{}, errors.Wrap(err, "failed to get all payment statement")
	}

	// S3 서비스를 이용하여 파일 URL을 서명된 URL로 변경
	for i := range items {
		singedUrl, err := u.s3pkg.CreateSignedUrl(ctx, items[i].FileUrl)
		if err != nil {
			return adomain.InvestorReportGetAllResponse{}, errors.Wrap(err, "failed to create signed url")
		}
		items[i].FileUrl = singedUrl
	}

	return adomain.InvestorReportGetAllResponse{
		TotalCount: total,
		Items:      items,
	}, nil
}

func (u *investorReportUsecase) GetAllTransactionBalanceReport(ctx context.Context, req adomain.InvestorReportGetAllRequest) (adomain.InvestorReportGetAllResponse, error) {

	var (
		submissionType = domain.SUBMISSION_INVESTORS
		reportTypes    = []domain.REPORT_TYPE{domain.REPORT_TRANSACTION_BALANCE_ADMIN_REPORT}
		groupBy        = "year,quarter"
	)

	total, items, err := u.reportOutputRepository.GetAll(ctx,
		cdomain.ReportOutputQueryOptionFunc().WithSubmissionType(submissionType),
		cdomain.ReportOutputQueryOptionFunc().WithReportTypes(reportTypes...),
		cdomain.ReportOutputQueryOptionFunc().WithGroupBy(groupBy),
		cdomain.ReportOutputQueryOptionFunc().WithSortBy(req.Sortby),
		cdomain.ReportOutputQueryOptionFunc().WithOrder(req.Order),
		cdomain.ReportOutputQueryOptionFunc().WithOffset(req.SkipCount),
		cdomain.ReportOutputQueryOptionFunc().WithLimit(req.MaxResultCount),
	)

	if err != nil {
		return adomain.InvestorReportGetAllResponse{}, errors.Wrap(err, "failed to get all transaction balance report")
	}

	// S3 서비스를 이용하여 파일 URL을 서명된 URL로 변경
	for i := range items {
		singedUrl, err := u.s3pkg.CreateSignedUrl(ctx, items[i].FileUrl)
		if err != nil {
			return adomain.InvestorReportGetAllResponse{}, errors.Wrap(err, "failed to create signed url")
		}
		items[i].FileUrl = singedUrl
	}

	return adomain.InvestorReportGetAllResponse{
		TotalCount: total,
		Items:      items,
	}, nil
}

func (u *investorReportUsecase) OpenPaymentStatement(ctx context.Context, req adomain.InvestorReportOpenPaymentStatementRequest) error {
	var (
		submissionType = domain.SUBMISSION_INVESTORS
		reportTypes    = []domain.REPORT_TYPE{domain.REPORT_PAYMENT_STATEMENT}
	)

	_, items, err := u.reportOutputRepository.GetAll(ctx,
		cdomain.ReportOutputQueryOptionFunc().WithSubmissionType(submissionType),
		cdomain.ReportOutputQueryOptionFunc().WithReportTypes(reportTypes...),
		cdomain.ReportOutputQueryOptionFunc().WithYear(req.Year),
	)

	if err != nil {
		return errors.Wrap(err, "failed to get all payment statement")
	}

	// 모든 지불조서가 이미 공개되어 있는지 확인
	alreadyAllOpened := true
	for _, item := range items {
		if item.IsOpen == false {
			alreadyAllOpened = false
			item.IsOpen = true
			if err := u.reportOutputRepository.Update(ctx, item); err != nil {
				return errors.Wrap(err, "failed to update report output")
			}
		}
	}

	if alreadyAllOpened {
		return cdomain.ErrAlreadyOpenReport
	}
	return nil
}

func (u *investorReportUsecase) OpenTransactionBalanceReport(ctx context.Context, req adomain.InvestorReportOpenTransactionBalanceRequest) error {
	var (
		submissionType = domain.SUBMISSION_INVESTORS
		reportTypes    = []domain.REPORT_TYPE{domain.REPORT_TRANSACTION_BALANCE_REPORT, domain.REPORT_TRANSACTION_BALANCE_ADMIN_REPORT}
	)

	_, items, err := u.reportOutputRepository.GetAll(ctx,
		cdomain.ReportOutputQueryOptionFunc().WithSubmissionType(submissionType),
		cdomain.ReportOutputQueryOptionFunc().WithReportTypes(reportTypes...),
		cdomain.ReportOutputQueryOptionFunc().WithYear(req.Year),
		cdomain.ReportOutputQueryOptionFunc().WithQuarter(req.Quarter),
	)

	if err != nil {
		return errors.Wrap(err, "failed to get all payment statement")
	}

	// 모든 거래잔액보고서가 이미 공개되어 있는지 확인
	alreadyAllOpened := true
	for _, item := range items {
		if item.IsOpen == false {
			alreadyAllOpened = false
			item.IsOpen = true

			if err := u.reportOutputRepository.Update(ctx, item); err != nil {
				return errors.Wrap(err, "failed to update report output")
			}
			break
		}
	}

	if alreadyAllOpened {
		return cdomain.ErrAlreadyOpenReport
	}
	return nil
}

func (u *investorReportUsecase) DownloadPaymentStatement(ctx context.Context, req adomain.DownloadPaymentStatementRequest) (adomain.DownloadPaymentStatementResponse, error) {
	var (
		submissionType = domain.SUBMISSION_INVESTORS
		reportTypes    = []domain.REPORT_TYPE{domain.REPORT_PAYMENT_STATEMENT}
		// 최신 파일을 다운로드하기 위해 id로 정렬해서 하나만 가져옴
		sortBy = "id"
		order  = "desc"
		limit  = 1
	)

	_, items, err := u.reportOutputRepository.GetAll(ctx,
		cdomain.ReportOutputQueryOptionFunc().WithSubmissionType(submissionType),
		cdomain.ReportOutputQueryOptionFunc().WithReportTypes(reportTypes...),
		cdomain.ReportOutputQueryOptionFunc().WithYear(req.Year),
		cdomain.ReportOutputQueryOptionFunc().WithSortBy(sortBy),
		cdomain.ReportOutputQueryOptionFunc().WithOrder(order),
		cdomain.ReportOutputQueryOptionFunc().WithLimit(limit),
	)

	if err != nil {
		return adomain.DownloadPaymentStatementResponse{}, errors.Wrap(err, "failed to get all transaction balance report")
	}

	if len(items) == 0 {
		return adomain.DownloadPaymentStatementResponse{}, cdomain.ErrItemNotFound
	}

	// S3 서비스를 이용하여 파일 URL을 서명된 URL로 변경
	singedUrl, err := u.s3pkg.CreateSignedUrl(ctx, items[0].FileUrl)
	if err != nil {
		return adomain.DownloadPaymentStatementResponse{}, errors.Wrap(err, "failed to create signed url")
	}

	return adomain.DownloadPaymentStatementResponse{
		Url: singedUrl,
	}, nil
}

func (u *investorReportUsecase) DownloadTransactionBalanceReport(ctx context.Context, req adomain.DownloadTransactionBalanceReportRequest) (adomain.DownloadTransactionBalanceReportResponse, error) {
	var (
		submissionType = domain.SUBMISSION_INVESTORS
		reportTypes    = []domain.REPORT_TYPE{domain.REPORT_TRANSACTION_BALANCE_ADMIN_REPORT}
		// 최신 파일을 다운로드하기 위해 id로 정렬해서 하나만 가져옴
		sortBy = "id"
		order  = "desc"
		limit  = 1
	)

	_, items, err := u.reportOutputRepository.GetAll(ctx,
		cdomain.ReportOutputQueryOptionFunc().WithSubmissionType(submissionType),
		cdomain.ReportOutputQueryOptionFunc().WithReportTypes(reportTypes...),
		cdomain.ReportOutputQueryOptionFunc().WithYear(req.Year),
		cdomain.ReportOutputQueryOptionFunc().WithQuarter(req.Quarter),
		cdomain.ReportOutputQueryOptionFunc().WithSortBy(sortBy),
		cdomain.ReportOutputQueryOptionFunc().WithOrder(order),
		cdomain.ReportOutputQueryOptionFunc().WithLimit(limit),
	)

	if err != nil {
		return adomain.DownloadTransactionBalanceReportResponse{}, errors.Wrap(err, "failed to get all transaction balance report")
	}

	if len(items) == 0 {
		return adomain.DownloadTransactionBalanceReportResponse{}, cdomain.ErrItemNotFound
	}

	// S3 서비스를 이용하여 파일 URL을 서명된 URL로 변경
	singedUrl, err := u.s3pkg.CreateSignedUrl(ctx, items[0].FileUrl)
	if err != nil {
		return adomain.DownloadTransactionBalanceReportResponse{}, errors.Wrap(err, "failed to create signed url")
	}

	return adomain.DownloadTransactionBalanceReportResponse{
		Url: singedUrl,
	}, nil
}
