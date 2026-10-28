package controllers

import (
	"github.com/labstack/echo/v4"
	"xorm.io/xorm"

	"github.com/illuminarean/energy-shares-backend-repo/pkg/factory"
)

const (
	FlashName      = "flash"
	FlashSeparator = ";"
	MaxPinCount    = 4
)

type ApiResult struct {
	Result  interface{}   `json:"result"`
	Success bool          `json:"success"`
	Error   factory.Error `json:"error"`
}

type ArrayResult struct {
	Items      interface{} `json:"items"`
	TotalCount int64       `json:"total_count"`
}

func renderSucc(c echo.Context, status int, result interface{}) error {
	req := c.Request()
	if req.Method == "POST" || req.Method == "PUT" || req.Method == "DELETE" {
		if session, ok := factory.DB(req.Context()).(*xorm.Session); ok {
			err := session.Commit()
			if err != nil {
				return factory.ErrorDB.New(err)
			}
		}
	}

	return c.JSON(status, ApiResult{
		Success: true,
		Result:  result,
	})
}
