package http

import (
	"net/http"

	adomain "github.com/illuminarean/energy-share-core-api/domain"
	cdomain "github.com/illuminarean/energy-shares-backend-repo/clean-arch-domain"
	"github.com/illuminarean/energy-shares-backend-repo/domain"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/apirender"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/echoutils"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/factory"
	"github.com/labstack/echo/v4"
	"github.com/pangpanglabs/echoswagger/v2"
	"github.com/pkg/errors"
)

type projectHttpDelivery struct {
	projectUsecase adomain.ProjectUsecase
	logger         factory.LoggerFactory
}

func New(
	api echoswagger.ApiRoot,
	projectUsecase adomain.ProjectUsecase,
	logger factory.LoggerFactory,
) *projectHttpDelivery {
	handler := &projectHttpDelivery{
		projectUsecase: projectUsecase,
		logger:         logger,
	}

	g := api.Group("v1/projects", "api/v1/projects")
	g.GET("/:id/statistics", handler.GetStatistics,
		echoutils.AllowRoles(domain.ROLE_PRODUCT)).
		AddResponse(http.StatusOK, "プロジェクト詳細統計データ取得", adomain.ProjectGetStatisticsResponse{}, nil).
		SetDescription("特定のプロジェクトの統計データを取得します。").
		SetSummary("プロジェクト詳細統計データ取得").
		SetSecurity("Authorization")

	return handler
}

func (phd projectHttpDelivery) GetStatistics(c echo.Context) error {
	ctx := c.Request().Context()

	request := adomain.ProjectGetStatisticsRequest{}
	if err := c.Bind(&request); err != nil {
		return factory.ErrorParameter.New(err)
	}

	if err := c.Validate(&request); err != nil {
		return factory.ErrorParameter.New(err)
	}

	response, err := phd.projectUsecase.GetStatistics(ctx, request)
	switch {
	case err == nil:
		return apirender.RenderSucc(c, http.StatusOK, response)
	case errors.Is(err, cdomain.ErrItemNotFound):
		return factory.ErrorNotFound.New(err)
	default:
		return factory.ErrorDB.New(err)
	}
}
