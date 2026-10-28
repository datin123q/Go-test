package investment

import (
	"context"
	"fmt"
	"time"

	"github.com/illuminarean/energy-shares-backend-repo/pkg/s3utils"

	"github.com/illuminarean/energy-shares-backend-repo/pkg/i18nutils"

	"github.com/samber/lo"

	"github.com/pkg/errors"

	cdomain "github.com/illuminarean/energy-shares-backend-repo/clean-arch-domain"
	"github.com/illuminarean/energy-shares-backend-repo/domain"
)

var investmentCSVColumns = []string{
	"model.mypage.invest_number", // 取引番号
	"model.mypage.invest_status", // 取引状態
	"model.mypage.name",          // 投資家名
	"model.mypage.spv_name",      // 開発会社名
	"model.mypage.project_type",  // 投資種別
	"model.mypage.project_name",  // 商品名
	"model.mypage.pay",           // 取引金額
	"model.mypage.invest_date",   // 取引日時
}

type investmentService struct {
	translator                   i18nutils.Translator
	s3pkg                        s3utils.S3
	investmentRepository         cdomain.InvestmentRepository
	contractRepository           cdomain.ContractRepository
	userRepository               cdomain.UserRepository
	userVirtualAccountRepository cdomain.UserVirtualAccountRepository
	depositRequestRepository     cdomain.DepositRequestRepository
	repaymentDividendSchedule    cdomain.RepaymentDividendScheduleRepository
}

func New(
	translator i18nutils.Translator,
	s3pkg s3utils.S3,
	investmentRepository cdomain.InvestmentRepository,
	contractRepository cdomain.ContractRepository,
	userRepository cdomain.UserRepository,
	userVirtualAccountRepository cdomain.UserVirtualAccountRepository,
	depositRequestRepository cdomain.DepositRequestRepository,
	repaymentDividendSchedule cdomain.RepaymentDividendScheduleRepository,
) cdomain.InvestmentService {
	return &investmentService{
		translator:                   translator,
		s3pkg:                        s3pkg,
		investmentRepository:         investmentRepository,
		contractRepository:           contractRepository,
		userRepository:               userRepository,
		userVirtualAccountRepository: userVirtualAccountRepository,
		depositRequestRepository:     depositRequestRepository,
		repaymentDividendSchedule:    repaymentDividendSchedule,
	}
}

func (i investmentService) GetDetail(ctx context.Context, investmentId string) (domain.InvestmentDetail, error) {
	investment, err := i.investmentRepository.GetById(ctx, investmentId)
	if err != nil {
		return domain.InvestmentDetail{}, errors.Wrap(err, "err get investment")
	}

	share := investment.Share

	depositRequest, err := i.depositRequestRepository.GetDepositRequestByShareId(ctx, share.Id)
	if err != nil {
		return domain.InvestmentDetail{}, errors.Wrap(err, "err get deposit request")
	}

	userVirtualAccount, err := i.userVirtualAccountRepository.GetByUserIdAndLinkedStatus(ctx, share.OwnerId, share.ProjectType, domain.ON_VA_ACCOUNT)
	if err != nil {
		return domain.InvestmentDetail{}, errors.Wrap(err, "err get user virtual account")
	}

	user, err := i.userRepository.GetById(ctx, share.OwnerId)
	if err != nil {
		return domain.InvestmentDetail{}, errors.Wrap(err, "err get user")
	}

	contracts, err := i.contractRepository.GetAll(ctx, cdomain.ContractQueryOptionFunc().WithShareId(share.Id))
	if err != nil {
		return domain.InvestmentDetail{}, errors.Wrap(err, "err get contract")
	}

	contract := contracts.First()

	contractAfterPageSignedUrl, err := i.s3pkg.CreateSignedUrl(ctx, contract.ContractAfterPageUrl)
	if err != nil {
		return domain.InvestmentDetail{}, errors.Wrap(err, "err create signed url")
	}

	_, repaymentDividendSchedules, err := i.repaymentDividendSchedule.GetAll(ctx,
		cdomain.RepaymentDividendQueryOptionFunc().WithLimit(1),
		cdomain.RepaymentDividendQueryOptionFunc().WithProjectId(share.ProjectId),
		cdomain.RepaymentDividendQueryOptionFunc().WithReturnType(domain.REPAYMENT_RETURN),
		cdomain.RepaymentDividendQueryOptionFunc().WithOrder("desc"),
		cdomain.RepaymentDividendQueryOptionFunc().WithSortBy("execute_date"),
	)
	if err != nil {
		return domain.InvestmentDetail{}, errors.Wrap(err, "err get repayment dividend schedule")
	}

	// 유저 관련 데이터 맵핑
	investment.BankName = user.BankName
	investment.BankAccountNumber = user.BankAccount
	investment.Depositor = user.Depositor

	// 계약 관련 데이터 맵핑
	investment.ContractStatus = string(contract.Status)
	investment.ContractAfterPageUrl = lo.ToPtr(contractAfterPageSignedUrl)

	return domain.InvestmentDetail{
		Investment:        investment,
		DepositRequest:    depositRequest,
		VaAccount:         userVirtualAccount,
		LastRepaymentDate: lo.ToPtr(repaymentDividendSchedules.First().ExecuteDate.ToString()),
	}, nil
}

func (i investmentService) ExportCSV(items []domain.Investment) cdomain.CSVData {
	csvHeaders := i.translator.Translates(investmentCSVColumns)

	records := [][]string{csvHeaders}

	for _, item := range items {
		record := []string{
			item.Share.Id,
			i.translator.Translate("commoncode.INVEST_STATUS." + string(item.InvestStatus)),
			item.UserName,
			item.SpvName,
			i.translator.Translate("commoncode.PROJECT_TYPE." + string(item.Share.ProjectType)),
			item.Project.Name,
			fmt.Sprintf("%v", item.Qty*item.Price),
			item.InvestDate.ToFormat(time.DateTime),
		}

		records = append(records, record)
	}

	return records
}
