package testing

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"mime/multipart"
	"net/http/httptest"
	"net/textproto"
	"net/url"
	"strings"

	"xorm.io/xorm"

	"github.com/hillfolk/goutils/echomiddleware"

	"github.com/pkg/errors"

	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"
)

const (
	_boundary = "test-boundary"
)

type Header struct {
	Headers   string
	MIMETypes string
}

type File struct {
	Filename string
	File     []byte
}

type Request struct {
	Path        string
	Method      string
	Body        string            // {"key1": "value1", "key2","value2"}
	Params      map[string]string // "outfitID" : "1004"
	QueryParams map[string][]string
	Header      Header        // {echo.HeaderContentType: echo.MIMEApplicationJSON}
	Claims      jwt.MapClaims // {"id": "1033422"}
	Multipart   Multipart
}

type Multipart struct {
	MultipartFiles map[string][]File
	MultipartForm  map[string]string
}

type Params struct {
	Names  []string
	Values []string
}

func (r Request) params() Params {
	var paramNames, paramValues []string
	for paramName, paramValue := range r.Params {
		paramNames = append(paramNames, paramName)
		paramValues = append(paramValues, paramValue)
	}
	return Params{
		Names:  paramNames,
		Values: paramValues,
	}
}

func (r Request) urlValues() url.Values {
	urlValues := make(url.Values)

	for key, val := range r.QueryParams {
		for index, element := range val {
			if index != 0 {
				urlValues.Add(key, element)
				continue
			}
			urlValues.Set(key, element)
		}
	}
	return urlValues
}

func (r Request) path() string {
	if len(r.urlValues()) != 0 {
		r.Path += fmt.Sprintf("/?%v", r.urlValues().Encode())
	}

	return r.Path
}

func (r Request) body() io.Reader {
	if r.Header.MIMETypes == echo.MIMEMultipartForm {
		return io.NopCloser(r.multipartWriter())
	}

	return strings.NewReader(r.Body)
}

func (r Request) write(writer *multipart.Writer) error {
	var err error
	for key, files := range r.Multipart.MultipartFiles {
		for _, file := range files {
			var part io.Writer

			header := make(textproto.MIMEHeader)
			header.Set(echo.HeaderContentDisposition,
				fmt.Sprintf(`form-data; name="%s"; filename="%s"`, key, file.Filename))
			header.Set(echo.HeaderContentType, echo.MIMEMultipartForm)

			part, err = writer.CreatePart(header)
			if err != nil {
				return errors.Wrap(err, "failed to create part")
			}
			if _, err = part.Write(file.File); err != nil {
				return errors.Wrap(err, "failed to write file")
			}
		}
	}
	for key, value := range r.Multipart.MultipartForm {
		if err = writer.WriteField(key, value); err != nil {
			return errors.Wrap(err, "failed to write field")
		}
	}
	return nil
}

func (r Request) multipartWriter() *bytes.Buffer {
	var buf bytes.Buffer

	writer := multipart.NewWriter(&buf)
	writer.SetBoundary(_boundary)

	r.write(writer)
	writer.Close()

	return &buf
}

func (r Request) mimeType() string {
	if r.Header.MIMETypes == echo.MIMEMultipartForm {
		return fmt.Sprintf("%v; boundary=%v", echo.MIMEMultipartForm, _boundary)
	}

	return r.Header.MIMETypes
}

func RequestContext(request Request) func(e *echo.Echo) echo.Context {
	return func(e *echo.Echo) echo.Context {

		newRequest := httptest.NewRequest(request.Method, request.path(), request.body())
		newRequest.Header.Set(request.Header.Headers, request.mimeType())
		newRequest = newRequest.WithContext(
			context.WithValue(newRequest.Context(), echomiddleware.ContextDBName, &xorm.Session{})) // 고의로 TypeAssertion 실패

		ctx := e.NewContext(newRequest, httptest.NewRecorder())

		params := request.params()

		ctx.Set("user", &jwt.Token{Claims: request.Claims})
		ctx.SetPath(request.path())        // Query Params 설정
		ctx.SetParamNames(params.Names...) // Path Params 설정
		ctx.SetParamValues(params.Values...)

		return ctx
	}
}
