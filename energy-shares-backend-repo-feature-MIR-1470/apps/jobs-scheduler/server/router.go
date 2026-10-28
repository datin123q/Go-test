package server

import (
	"os"

	"github.com/labstack/echo/v4"
	"github.com/pangpanglabs/echoswagger/v2"

	"energyshares-job-scheduler/controllers"

	"github.com/illuminarean/energy-shares-backend-repo/pkg/swagger"
)

func NewRouter(debug bool) *echo.Echo {
	e := echo.New()
	r := echoswagger.NewNop(e)
	if debug {
		versionInfo := "commit: " + os.Getenv("VERSION_INFO")
		if versionInfo == "" {
			versionInfo = "dev"
		}
		r = echoswagger.New(e, "/doc", &echoswagger.Info{
			Title:       "Energy Shares Japan Jobs Scheduler API",
			Description: "This is API doc for Energy Shares Japan Jobs Scheduler API.\nThe displayed version incorporates the commit date and hash for precise version tracking.",
			Version:     versionInfo,
		})
	}

	controllers.CommonApiControllers{}.Init(r.Group("Common", ""))
	controllers.AozoraApiControllers{}.Init(r.Group("GMO Aozora", "gmo-aozora"))
	controllers.WebHookApiControllers{}.Init(r.Group("GMO Aozora Webhook", "webhook"))
	controllers.AccountApiControllers{}.Init(r.Group("Account Information", "accounts"))
	controllers.BulkTransferApiControllers{}.Init(r.Group("GMO Aozora Transfer", "bulk-transfers"))
	controllers.VaAccountApiControllers{}.Init(r.Group("GMO Aozora Virtual Account", "va-accounts"))
	controllers.EventApiControllers{}.Init(r.Group("EventMessage", "event-message"))
	controllers.NewNexwayControllers().Init(r.Group("Nexway", "nexway"))
	controllers.NewNexwayJpkiControllers().Init(r.Group("Nexway JPKI", "nexway/jpki"))
	controllers.ExecuteApiControllers{}.Init(r.Group("스케쥴 즉시실행(QA팀 테스트용)", "execute"))
	// controllers.NewCleanUpController().Init(r.Group("CleanUpController", "clean-up")) // 1회성 API 구현용

	// if SWAGGER_MODE is GEN_SWAGGER_MODE, generate swagger.json and exit
	baseURL := "job-scheduler.dev.energyshares.jp"
	saveFile := "swagger-job-scheduler-api.json"
	swagger.GenerateSwaggerIfNeeded(r, e, baseURL, saveFile)

	return e
}
