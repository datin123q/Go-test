package echoutils

import (
	"net/http"
	"reflect"
	"strconv"
	"strings"

	"github.com/illuminarean/energy-shares-backend-repo/pkg/factory"

	"github.com/illuminarean/energy-shares-backend-repo/pkg/jsonutils"

	"github.com/labstack/echo/v4"
	"github.com/swaggest/openapi-go"
	"github.com/swaggest/openapi-go/openapi31"
)

type EchoSwagger struct {
	e          *echo.Echo
	openapi    *openapi31.Reflector
	operations []*echoMethodSwagger
}

type echoGroupSwagger struct {
	group     *echo.Group
	openapi   *openapi31.Reflector
	groupPath string
	parent    *EchoSwagger
}

type echoMethodSwagger struct {
	operation openapi.OperationContext
}

func NewEchoSwagger(e *echo.Echo, openapi *openapi31.Reflector) *EchoSwagger {
	return &EchoSwagger{
		e:          e,
		openapi:    openapi,
		operations: make([]*echoMethodSwagger, 0),
	}
}

func (swagger *EchoSwagger) Group(path string, m ...echo.MiddlewareFunc) *echoGroupSwagger {
	return &echoGroupSwagger{
		group:     swagger.e.Group(path, m...),
		openapi:   swagger.openapi,
		groupPath: path,
		parent:    swagger,
	}
}

func (es *echoGroupSwagger) GET(path string, handler echo.HandlerFunc, m ...echo.MiddlewareFunc) *echoMethodSwagger {
	fullPath := es.groupPath + path
	es.group.GET(path, handler, m...)

	operation, _ := es.openapi.NewOperationContext(http.MethodGet, convertPathVariables(fullPath))

	ems := &echoMethodSwagger{
		operation: operation,
	}

	es.parent.operations = append(es.parent.operations, ems)
	return ems
}

func (es *echoGroupSwagger) POST(path string, handler echo.HandlerFunc, m ...echo.MiddlewareFunc) *echoMethodSwagger {
	fullPath := es.groupPath + path
	es.group.POST(path, handler, m...)

	operation, _ := es.openapi.NewOperationContext(http.MethodPost, convertPathVariables(fullPath))

	ems := &echoMethodSwagger{
		operation: operation,
	}

	es.parent.operations = append(es.parent.operations, ems)
	return ems
}

func (es *echoGroupSwagger) PUT(path string, handler echo.HandlerFunc, m ...echo.MiddlewareFunc) *echoMethodSwagger {
	fullPath := es.groupPath + path
	es.group.PUT(path, handler, m...)

	operation, _ := es.openapi.NewOperationContext(http.MethodPut, convertPathVariables(fullPath))

	ems := &echoMethodSwagger{
		operation: operation,
	}

	es.parent.operations = append(es.parent.operations, ems)
	return ems
}

func (es *echoGroupSwagger) PATCH(path string, handler echo.HandlerFunc, m ...echo.MiddlewareFunc) *echoMethodSwagger {
	fullPath := es.groupPath + path
	es.group.PATCH(path, handler, m...)

	operation, _ := es.openapi.NewOperationContext(http.MethodPatch, convertPathVariables(fullPath))

	ems := &echoMethodSwagger{
		operation: operation,
	}

	es.parent.operations = append(es.parent.operations, ems)
	return ems
}

func (es *echoGroupSwagger) DELETE(path string, handler echo.HandlerFunc, m ...echo.MiddlewareFunc) *echoMethodSwagger {
	fullPath := es.groupPath + path
	es.group.DELETE(path, handler, m...)

	operation, _ := es.openapi.NewOperationContext(http.MethodDelete, convertPathVariables(fullPath))

	ems := &echoMethodSwagger{
		operation: operation,
	}

	es.parent.operations = append(es.parent.operations, ems)
	return ems
}

func (em *echoMethodSwagger) SetTags(tags ...string) *echoMethodSwagger {
	em.operation.SetTags(tags...)
	return em
}

func (em *echoMethodSwagger) SetSummary(summary string) *echoMethodSwagger {
	em.operation.SetSummary(summary)
	return em
}

func (em *echoMethodSwagger) SetDescription(description string) *echoMethodSwagger {
	em.operation.SetDescription(description)
	return em
}

func (em *echoMethodSwagger) AddRequest(request interface{}, options ...openapi.ContentOption) *echoMethodSwagger {
	em.operation.AddReqStructure(request, options...)

	return em
}

func (em *echoMethodSwagger) AddResponse(response interface{}, options ...openapi.ContentOption) *echoMethodSwagger {
	em.operation.AddRespStructure(response, options...)
	return em
}

func (em *echoMethodSwagger) AddErrors(errors ...factory.ErrorTemplate) *echoMethodSwagger {
	for _, err := range errors {
		// 動的に構造体を生成
		responseType := reflect.StructOf([]reflect.StructField{
			{
				Name: "Code",
				Type: reflect.TypeOf(""),
				Tag:  reflect.StructTag(`json:"code" example:"` + strconv.Itoa(err.Code) + `"`),
			},
			{
				Name: "Message",
				Type: reflect.TypeOf(""),
				Tag:  reflect.StructTag(`json:"message" example:"` + err.Message + `"`),
			},
		})

		// 構造体インスタンスを生成
		response := reflect.New(responseType).Elem()
		response.FieldByName("Code").SetString(strconv.Itoa(err.Code))
		response.FieldByName("Message").SetString(err.Message)

		// AddResponseを呼び出す
		em.AddResponse(response.Interface(), func(cu *openapi.ContentUnit) {
			cu.HTTPStatus = err.Status()
			cu.Description = err.Message
		})
	}
	return em
}

func (em *echoMethodSwagger) AddSecurity(security string) *echoMethodSwagger {
	em.operation.AddSecurity(security)
	return em
}

func (em *echoMethodSwagger) SetDeprecated() *echoMethodSwagger {
	em.operation.SetIsDeprecated(true)
	return em
}

func (swagger *EchoSwagger) RegisterOpenAPI() {
	swagger.openapi.AddTypeMapping(jsonutils.JSONTime{}, "string")

	for _, op := range swagger.operations {
		err := swagger.openapi.AddOperation(op.operation)

		if err != nil {
			panic(err)
		}

	}
	swagger.operations = make([]*echoMethodSwagger, 0) // Clear the operations after registering
}

func convertPathVariables(path string) string {
	parts := strings.Split(path, "/")
	for i, part := range parts {
		if strings.HasPrefix(part, ":") {
			parts[i] = "{" + strings.TrimPrefix(part, ":") + "}"
		}
	}
	return strings.Join(parts, "/")
}
