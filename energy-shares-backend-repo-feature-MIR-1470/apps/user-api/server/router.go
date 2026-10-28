package server

import (
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/pangpanglabs/echoswagger/v2"

	cdomain "github.com/illuminarean/energy-shares-backend-repo/clean-arch-domain"
	"github.com/illuminarean/energy-shares-backend-repo/domain"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/config"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/echoutils"

	"github.com/illuminarean/energyshares-user-api/controllers"
)

func registerRouter(r echoswagger.ApiRoot, cfg *config.Config, sqsClient *sqs.Client, userVirtualAccountRepository cdomain.UserVirtualAccountRepository) {
	r.AddSecurityAPIKey("Authorization", "認証トークン", echoswagger.SecurityInHeader)

	controllers.NewCommonApiController().Init(r.Group("Common", "api"))
	controllers.NewProjectApiController(userVirtualAccountRepository).Init(r.Group("Projects", "api/projects"), cfg, sqsClient)
	controllers.NewUserApiController().Init(r.Group("Users", "api/users"))
	controllers.NewCorporateApiController().Init(r.Group("Corporate Users", "api/corporateusers"))
	controllers.ContactApiController{}.Init(r.Group("Contact", "api/contact"))
	controllers.FundingRequestApiController{}.Init(r.Group("Funding Request", "api/fundingrequests"))
	controllers.QnaApiController{}.Init(r.Group("Q&A", "api/qna",
		echoutils.AllowRoles(domain.ROLE_USER)))
	controllers.AuthApiController{}.Init(r.Group("Authentication", "api/auth"))
	controllers.NewJapanBankApiController().Init(r.Group("Japan Bank List", "api/japan-banks"))
	controllers.NoticeMsgApiController{}.Init(r.Group("Notice-msg ", "api/notice-msgs",
		echoutils.AllowRoles(domain.ROLE_USER)))
	controllers.NewWithdrawalController().Init(r.Group("Withdrawal", "api/withdrawal",
		echoutils.AllowRoles(domain.ROLE_USER),
		echoutils.AllowAuditStatuses(domain.COMPLETED_AUDIT)))
	controllers.NewColumnApiController().Init(r.Group("Column", "api/columns"))
	controllers.NewBlogPostApiController().Init(r.Group("Blog Post", "api/blog-posts"))
	controllers.AdEbisAPIController{}.Init(r.Group("Ad Ebis", "api/adebis"))
}
