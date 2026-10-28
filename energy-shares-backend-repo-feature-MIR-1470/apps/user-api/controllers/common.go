package controllers

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/pangpanglabs/echoswagger/v2"

	"github.com/illuminarean/energy-shares-backend-repo/domain"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/commoncode"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/serviceutils"
)

type AppInfo struct {
	Name    string `json:"Name" `
	Version string `json:"Version"`
}

type CommonApiController struct{}

func NewCommonApiController() *CommonApiController {
	return &CommonApiController{}
}

func (c CommonApiController) Init(g echoswagger.ApiGroup) {
	g.GET("/version", c.GetInfo).
		SetSummary("バージョン照会")

	g.GET("/codes", c.GetCommonCodeList).
		AddResponse(http.StatusOK, "", commoncode.CodeGroup{}, nil).
		SetDescription("共通コードを照会します。").
		AddParamHeader("", "Accept-Language", "", false).
		SetSummary("共通コード照会")
}

func (CommonApiController) GetInfo(c echo.Context) error {
	return renderSucc(c, http.StatusOK, AppInfo{
		Name:    serviceutils.GetServiceName(),
		Version: serviceutils.GetServiceVersion(),
	})
}

func (CommonApiController) GetCommonCodeList(c echo.Context) error {
	lang := c.Request().Header.Get("Accept-Language")
	if lang == "" {
		lang = "ja"
	}
	accept := lang

	codes := make([]commoncode.CodeGroup, 0, 0)
	codes = append(codes, domain.FundingRequest{}.GetReqeustStatusCodeGroup(accept))
	codes = append(codes, domain.Faq{}.GetTypeCodeGroup(accept))
	codes = append(codes, domain.Project{}.GetCategoryCodeGroup(accept))
	codes = append(codes, domain.Project{}.GetTypeCodeGroup(accept))
	codes = append(codes, domain.Project{}.GetListCodeGroup(accept))
	codes = append(codes, domain.Project{}.GetStatusCodeGroup(accept))
	codes = append(codes, domain.Project{}.GetProjrctListStatusCodeGroup(accept))
	codes = append(codes, domain.Project{}.GetProjectCountryGroup(accept))
	codes = append(codes, domain.Project{}.GetProjectSelectionTypeCodeGroup(accept))
	codes = append(codes, domain.Project{}.GetProjecDividendTypeGroup(accept))
	codes = append(codes, domain.ProjectMilestone{}.GetMilestoneStatus(accept))
	codes = append(codes, domain.Share{}.GetCodeGroup(accept))
	codes = append(codes, domain.User{}.GetLoginResultCodeGroup(accept))
	codes = append(codes, domain.User{}.GetUserTypeCodeGroup(accept))
	codes = append(codes, domain.User{}.GetLoginTypeCodeGroup(accept))
	codes = append(codes, domain.Mail{}.GetMailTypeCodeGroup(accept))
	codes = append(codes, domain.Mail{}.GetMailStatusCodeGroup(accept))
	codes = append(codes, domain.SmsMsg{}.GetSmsTypeCodeGroup(accept))
	codes = append(codes, domain.SmsMsg{}.GetSmsStatusCodeGroup(accept))
	codes = append(codes, domain.PrivateProfile{}.GetAssetsCodeGroup(accept))
	codes = append(codes, domain.PrivateProfile{}.GetsAnnualincomeCodeGroup(accept))
	codes = append(codes, domain.PrivateProfile{}.GetGenderCodeGroup(accept))
	codes = append(codes, domain.User{}.GetAuditStatusCodeGroup(accept))
	codes = append(codes, domain.User{}.GetInvestorTypeCodeGroup(accept))
	codes = append(codes, domain.User{}.GetPrivateInvestorTypeCodeGroup(accept))
	codes = append(codes, domain.User{}.GetCorporateInvestorTypeCodeGroup(accept))
	codes = append(codes, domain.User{}.GetNationalityCodeGroup(accept))
	codes = append(codes, domain.User{}.GetDepositTypeCodeGroup(accept))
	codes = append(codes, domain.User{}.GetCorpTypeCodeGroup(accept))
	codes = append(codes, domain.User{}.GetUserJobTypeCodeGroup(accept))
	codes = append(codes, domain.User{}.GetUserInvestmentPurposeTypeCodeGroup(accept))
	codes = append(codes, domain.UserInfoHistory{}.GetUserHistoryProcessTypeCodeGroup(accept))
	codes = append(codes, domain.UserInfoHistory{}.GetPostalCodeStateCodeGroup(accept))
	codes = append(codes, domain.UserInfoHistory{}.GetTransactionTypeCodeGroup(accept))
	codes = append(codes, domain.UserInfoHistory{}.GetUserHistoryCheckTypeCodeGroup(accept))
	codes = append(codes, domain.UserInfoHistory{}.GetUserHistoryAgreementTypeCodeGroup(accept))
	codes = append(codes, domain.UserInfoHistory{}.GetUserDocumentTypeCodeGroup(accept))
	codes = append(codes, domain.UserInfoHistory{}.GetUserHistoryTypeCodeGroup(accept))
	codes = append(codes, domain.UserInfoHistory{}.GetUserChangedContentTypeCodeGroup(accept))
	codes = append(codes, domain.ProjectInfoHistory{}.GetProjectPublishTypeCodeGroup(accept))
	codes = append(codes, domain.ProjectInfoHistory{}.GetProjectInvestTypeCodeGroup(accept))
	codes = append(codes, domain.ProjectInfoHistory{}.GetProjectNewsTypeCodeGroup(accept))
	codes = append(codes, domain.ProjectInfoHistory{}.GetProjectHistoryTypeCodeGroup(accept))
	codes = append(codes, domain.ProjectInfoHistory{}.GetProjectChangedContentTypeCodeGroup(accept))
	codes = append(codes, domain.SpvInfoHistory{}.GetSpvHistoryTypeCodeGroup(accept))
	codes = append(codes, domain.SpvInfoHistory{}.GetSpvChangedContentTypeCodeGroup(accept))
	codes = append(codes, domain.PolicyInfoHistory{}.GetPolicyHistoryTypeCodeGroup(accept))
	codes = append(codes, domain.FundingRequestHistory{}.GetFundingHistoryTypeCodeGroup(accept))
	codes = append(codes, domain.FundingRequestHistory{}.GetFundingHistoryContentTypeCodeGroup(accept))
	codes = append(codes, domain.QnaInfoHistory{}.GetQnaChangedContentTypeCodeGroup(accept))
	codes = append(codes, domain.UserVirtualAccount{}.GetVaLinkStatusCodeGroup(accept))
	codes = append(codes, domain.UserRoles{}.GetRoleTypeCodeGroup(accept))
	codes = append(codes, domain.MyNumber{}.GetMyNumberStatusGroup(accept))
	codes = append(codes, domain.Policy{}.GetTypeCodeGroup(accept))
	codes = append(codes, domain.Policy{}.GetViewTypeCodeGroup(accept))
	codes = append(codes, domain.ReportOutput{}.GetQuarterTypeCodeGroup(accept))
	codes = append(codes, domain.ReportOutput{}.GetReportTypeCodeGroup(accept))
	codes = append(codes, domain.SettlementSchedule{}.GetSettlementExecutionStatusCodeGroup(accept))
	codes = append(codes, domain.Settlement{}.GetSettlementStatusCodeGroup(accept))
	codes = append(codes, domain.Settlement{}.GetSettlementTypeCodeGroup(accept))
	codes = append(codes, domain.TransactionInfoHistory{}.GetTransactionHistoryTypeCodeGroup(accept))
	codes = append(codes, domain.TransactionInfoHistory{}.GetTransactionChangedContentTypeCodeGroup(accept))
	codes = append(codes, domain.UserTransaction{}.GetTransactionTypeCodeGroup(accept))
	codes = append(codes, domain.UserTransaction{}.GetTransactionDetailTypeCodeGroup(accept))
	codes = append(codes, domain.DepositMatchData{}.GetDepositMatchStatus(accept))
	codes = append(codes, domain.DepositCheckData{}.GetDepositCheckStatus(accept))
	codes = append(codes, domain.InvestmentWaiters{}.GetTypeCodeGroup(accept))
	codes = append(codes, domain.InvestmentWaiters{}.GetStatusCodeGroup(accept))
	codes = append(codes, domain.InvestmentWaitHistory{}.GetTypeCodeGroup(accept))
	codes = append(codes, domain.OperatorHistory{}.GetOperatorStateCodeGroup(accept))
	codes = append(codes, domain.Contract{}.GetContractStatusGroup(accept))
	codes = append(codes, domain.Withdrawal{}.GetWithdrawalStatusTypeCodeGroup(accept))
	codes = append(codes, domain.EkycRequestLog{}.GetResultStatusCodeGroup(accept))
	codes = append(codes, domain.Qna{}.GetQnaStatusTypeCodeGroup(accept))
	codes = append(codes, domain.LotteryApplication{}.GetStatusCodeGroup(accept))
	return renderSucc(c, http.StatusOK, codes)
}
