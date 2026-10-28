package server

import (
	"github.com/pangpanglabs/echoswagger/v2"

	"github.com/illuminarean/energy-share-core-api/controllers"
	"github.com/illuminarean/energy-shares-backend-repo/domain"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/echoutils"
)

func registerRouter(r echoswagger.ApiRoot) {

	r.AddSecurityAPIKey("Authorization", "인증 토큰", echoswagger.SecurityInHeader)

	controllers.NewCommonApiController().Init(r.Group("Common", "api"))
	controllers.NewStorageApiController().Init(r.Group("Storage", "api/storage"))
	controllers.ProjectApiController{}.Init(r.Group("Projects", "api/projects",
		echoutils.AllowRoles(domain.ROLE_PRODUCT)))
	controllers.NewUserApiController().Init(r.Group("Users", "api/users"))
	controllers.NewUserAccountApiController().Init(r.Group("User Accounts", "/api/user-accounts",
		echoutils.AllowRoles(domain.ROLE_PAYMENT)))
	controllers.NewPrivateUserApiController().Init(r.Group("Private Users", "api/privateusers"))
	controllers.NewCorporateApiController().Init(r.Group("Corporate Users", "api/corporateusers"))
	controllers.SpvApiController{}.Init(r.Group("SPV Developer", "api/spvs"))
	controllers.ContactApiController{}.Init(r.Group("Contact", "api/contact"))
	controllers.FundingRequestApiController{}.Init(r.Group("Funding Request", "api/fundingrequests",
		echoutils.AllowRoles(domain.ROLE_PRODUCT)))
	controllers.ConsultNoteApiController{}.Init(r.Group("Consult Note", "api"))
	controllers.LoginHistoryApiController{}.Init(r.Group("Login History", "api/loginhistories",
		echoutils.AllowRoles(domain.ROLE_INVESTOR)))
	controllers.OperatorHistoryApiController{}.Init(r.Group("Operator History", "api/operatorhistories"))
	controllers.MailApiController{}.Init(r.Group("Mail", "api/mails",
		echoutils.AllowRoles(domain.ROLE_MESSENGER)))
	controllers.SmsApiController{}.Init(r.Group("SMS Message", "api/smsmsg",
		echoutils.AllowRoles(domain.ROLE_MESSENGER)))
	controllers.QnaApiController{}.Init(r.Group("Q&A", "api/qna",
		echoutils.AllowRoles(domain.ROLE_INVESTOR)))
	controllers.DashboardApiController{}.Init(r.Group("Dashboard", "api/dashboards"))
	controllers.AuthApiController{}.Init(r.Group("Authentication", "api/auth"))
	controllers.UserInfoHistoryApiController{}.Init(r.Group("User Information History", "api/userinfohistories",
		echoutils.AllowRoles(domain.ROLE_INVESTOR)))
	controllers.FundingRequestHistoryApiController{}.Init(r.Group("Funding Request History", "api/fundingrequesthistories"))
	controllers.SpvInfoHistoryApiController{}.Init(r.Group("Spv Information History", "api/spvinfohistories"))
	controllers.ProjectInfoHistoryApiController{}.Init(r.Group("Project Information History", "api/projectinfohistories"))
	controllers.QnaInfoHistoryApiController{}.Init(r.Group("QnA Information History", "api/qnainfohistories",
		echoutils.AllowRoles(domain.ROLE_INVESTOR)))
	controllers.UserVirtaulAccountApiController{}.Init(r.Group("User Virtual Account", "api/user-virtual-accounts"))
	controllers.NewJapanBankApiController().Init(r.Group("Japan Bank List", "api/japan-banks"))
	controllers.UserTransactionApiController{}.Init(r.Group("User Transactions", "api/user-transactions",
		echoutils.AllowRoles(domain.ROLE_ADMIN)))
	controllers.DepositMatchDataApiController{}.Init(r.Group("Deposit Match Data ", "api/deposit-match-data",
		echoutils.AllowRoles(domain.ROLE_PAYMENT)))
	controllers.DepositCheckApiController{}.Init(r.Group("Deposit Check Data ", "api/deposit-check-data"))
	controllers.DepositRequestLackHistoryApiController{}.Init(r.Group("Deposit Request Lack History ", "api/deposit-request-lack-history"))
	controllers.NoticeMsgApiController{}.Init(r.Group("Notice-msg ", "api/notice-msgs",
		echoutils.AllowRoles(domain.ROLE_INVESTOR)))
	controllers.NewMyNumberApiController().Init(r.Group("My Number", "api/my-number",
		echoutils.AllowRoles(domain.ROLE_MYNUMBER)))
	controllers.NewInvestorReportOutputApiController().Init(r.Group("Investor Report Output", "api/investor-report-output",
		echoutils.AllowRoles(domain.ROLE_REPORT)))
	controllers.TaxOfficeReportOutputApiController{}.Init(r.Group("Tax-Office Report Output", "api/tax-office-report-output",
		echoutils.AllowRoles(domain.ROLE_REPORT)))
	controllers.AuditResponseReportOutputApiController{}.Init(r.Group("Audit Response Report Output", "api/audit-response-report-output",
		echoutils.AllowRoles(domain.ROLE_REPORT)))
	controllers.JobApiController{}.Init(r.Group("Jobs", "api/jobs"))
	controllers.NewWithdrawalController().Init(r.Group("withdrawal", "api/withdrawal",
		echoutils.AllowRoles(domain.ROLE_PAYMENT)))
	controllers.TrustInfoApiController{}.Init(r.Group("Trust Information", "api/trust-infos",
		echoutils.AllowRoles(domain.ROLE_PAYMENT)))
	controllers.TrustInfoHistoryApiController{}.Init(r.Group("Trust Information History", "api/trust-info-history"))
	controllers.NewColumnApiController().Init(r.Group("Column", "api/columns"))
	controllers.CapitalRaisingEntityApiController{}.Init(r.Group("CapitalRaisingEntity", "api/capital-raising-entity",
		echoutils.AllowRoles(domain.ROLE_PRODUCT)))
}
