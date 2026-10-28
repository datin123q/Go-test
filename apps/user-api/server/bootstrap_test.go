package server

import (
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/require"
	"xorm.io/xorm"

	"github.com/illuminarean/energy-shares-backend-repo/pkg/config"
)

func TestBootstrapDependencyGraph(t *testing.T) {
	t.Setenv("SWAGGER_MODE", "")

	app := NewBootstrap(
		echo.New(),
		&config.Config{},
		&xorm.Engine{},
		&sqs.Client{},
	).Invoke()

	require.NoError(t, app.Err())
}
