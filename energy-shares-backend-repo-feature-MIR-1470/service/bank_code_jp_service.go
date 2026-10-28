package service

import (
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"

	"github.com/illuminarean/energy-shares-backend-repo/pkg/factory"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/serviceutils"
)

type IBankCodeJPService interface {
	GetBanksWithFreeword(string, int64, *string) (*JapanBankResult, error)
	GetBranchesWithFreeword(string, string, int64, *string) (*JapanBankBranchResult, error)
}

var _ IBankCodeJPService = (*BankCodeJPService)(nil)

type BankCodeJPService struct {
}

func NewBankCodeJPService() IBankCodeJPService {
	return &BankCodeJPService{}
}

const BANK_CODE_JP_FREEWORD_API = "https://apis.bankcode-jp.com/v3/freeword/banks"

func (s *BankCodeJPService) GetBanksWithFreeword(keyword string, limit int64, cursor *string) (*JapanBankResult, error) {

	encodedKeyword := url.QueryEscape(keyword)
	urlStr := BANK_CODE_JP_FREEWORD_API + "?freeword=" + encodedKeyword + "&limit=" + strconv.FormatInt(limit, 10)
	if cursor != nil {
		urlStr += "&cursor=" + url.QueryEscape(*cursor)
	}

	body, err := s.callBankCodeJPApi(echo.GET, urlStr)
	if err != nil {
		return nil, err
	}

	logrus.WithFields(logrus.Fields{
		"api":      urlStr,
		"err":      err,
		"time":     time.Now(),
		"response": string(body),
	}).Info("BankCodeJP : GetBanksWithFreeword() : log for test")

	var output JapanBankResult
	err = json.Unmarshal(body, &output)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"api":      urlStr,
			"err":      err,
			"time":     time.Now(),
			"response": string(body),
		}).Error("BankCodeJP : GetBanksWithFreeword() : Failed unmarshal BankCodeJP response")
		return nil, factory.ErrorServiceUnavailable.New(errors.Wrap(err, "Failed unmarshal BankCodeJP response"))
	}

	return &output, nil
}

func (s *BankCodeJPService) GetBranchesWithFreeword(bankCode, keyword string, limit int64, cursor *string) (*JapanBankBranchResult, error) {

	encodedKeyword := url.QueryEscape(keyword)
	urlStr := BANK_CODE_JP_FREEWORD_API + "/" + bankCode + "/branches?freeword=" + encodedKeyword + "&limit=" + strconv.FormatInt(limit, 10)
	if cursor != nil {
		urlStr += "&cursor=" + url.QueryEscape(*cursor)
	}

	body, err := s.callBankCodeJPApi(echo.GET, urlStr)
	if err != nil {
		return nil, err
	}

	var output JapanBankBranchResult
	err = json.Unmarshal(body, &output)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"api":      urlStr,
			"err":      err,
			"time":     time.Now(),
			"response": string(body),
		}).Error("BankCodeJP : GetBranchesWithFreeword() : Failed unmarshal BankCodeJP response")
		return nil, factory.ErrorServiceUnavailable.New(errors.Wrap(err, "Failed unmarshal BankCodeJP response"))
	}

	return &output, nil
}

func (s *BankCodeJPService) callBankCodeJPApi(method, url string) ([]byte, error) {
	client := &http.Client{}

	req, err := http.NewRequest(method, url, nil)
	if err != nil {
		return nil, factory.ErrorServiceUnavailable.New(err)
	}
	req.Header.Set("Content-Type", "application/json; charset=UTF-8")
	req.Header.Set("apikey", serviceutils.GetBankCodeJPApiKey())

	res, err := client.Do(req)
	if err != nil {
		return nil, factory.ErrorServiceUnavailable.New(err)
	}
	body, err := s.readBody(res)
	if err != nil {
		return nil, factory.ErrorServiceUnavailable.New(err)
	}

	return body, nil
}

func (s *BankCodeJPService) readBody(resp *http.Response) ([]byte, error) {
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	resp.Body.Close()

	return body, nil
}

type JapanBank struct {
	Code             string `json:"code"`
	Name             string `json:"name"`
	HalfWidthKana    string `json:"halfWidthKana"`
	FullWidthKana    string `json:"fullWidthKana"`
	Hiragana         string `json:"hiragana"`
	BusinessTypeCode string `json:"businessTypeCode"`
	BusinessType     string `json:"businessType"`
}

type JapanBankResult struct {
	Banks      []JapanBank `json:"banks"`
	Size       int64       `json:"size"`
	Limit      int64       `json:"limit"`
	HasNext    bool        `json:"hasNext"`
	NextCursor *string     `json:"nextCursor"`
	HasPrev    bool        `json:"hasPrev"`
	PrevCursor *string     `json:"prevCursor"`
	Version    string      `json:"version"`
}

type JapanBankBranch struct {
	Code          string `json:"code"`
	Name          string `json:"name"`
	HalfWidthKana string `json:"halfWidthKana"`
	FullWidthKana string `json:"fullWidthKana"`
	Hiragana      string `json:"hiragana"`
}

type JapanBankBranchResult struct {
	Bank       JapanBank         `json:"bank"`
	Branches   []JapanBankBranch `json:"branches"`
	Size       int64             `json:"size"`
	Limit      int64             `json:"limit"`
	HasNext    bool              `json:"hasNext"`
	NextCursor *string           `json:"nextCursor"`
	HasPrev    bool              `json:"hasPrev"`
	PrevCursor *string           `json:"prevCursor"`
	Version    string            `json:"version"`
}
