package emailutils

import (
	"context"
	"fmt"
	"net/url"
	"strconv"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/pkg/errors"
	"github.com/samber/lo"

	"github.com/illuminarean/energy-shares-backend-repo/pkg/config"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/factory"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/jsonutils"
)

// 統合メールサービス - メール送信リクエストボディ
// AttachmentFileは構造体は定義するが、現在は考慮されていない。
type AttachmentFile struct {
	Name    string `json:"name"`
	Content string `json:"content"`
}

type MailData struct {
	TemplateFields  map[string]interface{} `json:"templateFields"`
	Sender          string                 `json:"sender"`
	Receivers       []string               `json:"receivers"`
	AttachmentFiles []AttachmentFile       `json:"attachmentFiles"`
	SendType        string                 `json:"sendType"`
	Threading       bool                   `json:"threading"`

	// Optional fields
	ReplyTo     string `json:"replyTo,omitempty"`
	ReplyToName string `json:"replyToName,omitempty"`
}

type MailDataOption func(*MailData)

func NewMailData(templateFields map[string]interface{}, sender string, receivers []string, opts ...MailDataOption) *MailData {
	mailData := &MailData{
		TemplateFields:  templateFields,
		Sender:          sender,
		Receivers:       receivers,
		AttachmentFiles: []AttachmentFile{},
		SendType:        "NORMAL",
		Threading:       false,
	}

	for _, opt := range opts {
		opt(mailData)
	}

	return mailData
}

func WithReplyTo(email string) MailDataOption {
	return func(m *MailData) {
		m.ReplyTo = email
	}
}

func WithReplyToName(name string) MailDataOption {
	return func(m *MailData) {
		m.ReplyToName = name
	}
}

type EmailRequest struct {
	TemplateName string     `json:"templateName"`
	MailData     []MailData `json:"mailData"`
	ServiceName  string     `json:"serviceName"`
	ServiceEnv   string     `json:"serviceEnv"`
}

// 통합 메일 서비스 - GetAll 쿼리 파라미터
type GetAllQuery struct {
	ServiceName string
	ServiceEnv  string
	Subject     *string
	Sender      *string
	Receiver    *string
	Status      *string
	StartDate   *string
	EndDate     *string
	Sort        *string
	Page        int
	Size        int
}

func (input GetAllQuery) ToUrlValues() url.Values {
	result := url.Values{}

	result.Add("serviceName", input.ServiceName)
	result.Add("serviceEnv", input.ServiceEnv)

	if input.Subject != nil {
		result.Add("subject", *input.Subject)
	}
	if input.Sender != nil {
		result.Add("sender", *input.Sender)
	}
	if input.Receiver != nil {
		result.Add("receivers", *input.Receiver)
	}
	if input.Status != nil {
		result.Add("status", *input.Status)
	}
	if input.StartDate != nil && input.EndDate != nil {
		result.Add("sendDateTime", *input.StartDate)
		result.Add("sendDateTime", *input.EndDate)
	}
	if input.Sort != nil {
		result.Add("sort", *input.Sort)
	}

	result.Add("page", strconv.Itoa(input.Page))
	result.Add("size", strconv.Itoa(input.Size))

	return result
}

// 통합 메일 서비스 - GetAll 응답 바디
type Email struct {
	ID             int64              `json:"id"`
	Subject        string             `json:"subject"`
	Receivers      []string           `json:"receivers"`
	TemplateFields string             `json:"templateFields"`
	ServiceName    string             `json:"serviceName"`
	ServiceEnv     string             `json:"serviceEnv"`
	Status         string             `json:"status"`
	SendType       string             `json:"sendType"`
	Sender         string             `json:"sender"`
	TemplateName   string             `json:"templateName"`
	SendDateTime   jsonutils.JSONTime `json:"sendDateTime"`
}

type Sort struct {
	Direction    string `json:"direction"`
	Property     string `json:"property"`
	IgnoreCase   bool   `json:"ignoreCase"`
	NullHandling string `json:"nullHandling"`
	Ascending    bool   `json:"ascending"`
	Descending   bool   `json:"descending"`
}

type Pageable struct {
	PageNumber int64  `json:"pageNumber"`
	PageSize   int64  `json:"pageSize"`
	Sort       []Sort `json:"sort"`
	Offset     int64  `json:"offset"`
	Paged      bool   `json:"paged"`
	Unpaged    bool   `json:"unpaged"`
}

type EmailListResponse struct {
	Content          []Email  `json:"content"`
	Pageable         Pageable `json:"pageable"`
	TotalPages       int64    `json:"totalPages"`
	TotalElements    int64    `json:"totalElements"`
	Last             bool     `json:"last"`
	Size             int64    `json:"size"`
	Number           int64    `json:"number"`
	Sort             []Sort   `json:"sort"`
	NumberOfElements int64    `json:"numberOfElements"`
	First            bool     `json:"first"`
	Empty            bool     `json:"empty"`
}

// 통합 메일 서비스 - GetByIdRequest 응답 바디
type EmailResponse struct {
	ID             int64               `json:"id"`
	ServiceName    string              `json:"serviceName"`
	ServiceEnv     string              `json:"serviceEnv"`
	Subject        string              `json:"subject"`
	Sender         string              `json:"sender"`
	Receivers      []string            `json:"receivers"`
	Body           string              `json:"body"`
	TemplateName   string              `json:"templateName"`
	TemplateFields string              `json:"templateFields"`
	SendType       string              `json:"sendType"`
	Status         string              `json:"status"`
	SendDateTime   jsonutils.JSONTime  `json:"sendDateTime"`
	CreatedDate    *jsonutils.JSONTime `json:"createdDate,omitempty"`
	ModifiedDate   *jsonutils.JSONTime `json:"modifiedDate,omitempty"`
}

type MailTemplateResponse struct {
	Subject      string `json:"subject"`
	Content      string `json:"content"`
	TemplateName string `json:"templateName"`
}

type emailpkg struct {
	cfg           *config.Config
	loggerFactory factory.LoggerFactory
	restyClient   *resty.Client
}

// メール受信者が1名の場合のみを想定して実装している。統合メールサービスは複数の受信者への送信が可能である。2024-07-04
type EmailRequester interface {
	SendEmailRequest(ctx context.Context, request EmailRequest) ([]string, error)
	GetAllRequest(ctx context.Context, query GetAllQuery) (EmailListResponse, error)
	GetByIdRequest(ctx context.Context, id string) (EmailResponse, error)
	GetTemplate(ctx context.Context, templateName string) (MailTemplateResponse, error)
}

// MIR-1327: resty.ClientのClone()は内部の*http.Clientをポインタのまま
// 引き継ぐため、DIコンテナが全機能に共有するresty.NewをCloneしてSetTimeoutすると、
// 元のクライアントを使う他の機能(eKYC・egift等)の打ち切り時間まで書き換わって
// しまう。resty.New()で完全に独立したクライアントを作ることで、この機能の
// タイムアウト設定が他へ漏れないようにする。
func New(cfg *config.Config, loggerFactory factory.LoggerFactory) EmailRequester {
	return &emailpkg{
		cfg:           cfg,
		loggerFactory: loggerFactory,
		restyClient: resty.New().
			SetHeader("Content-Type", "application/json").
			SetBaseURL(cfg.EmailService.Url).
			SetTimeout(10 * time.Second),
	}
}

const (
	SEND_REQUEST_PATH = "/api/v1/mails/send-request"
	GET_REQUEST_PATH  = "/api/v1/mail-logs"
	GET_TEMPLATE_PATH = "/api/v1/mails/template"
)

func (e emailpkg) SendEmailRequest(ctx context.Context, request EmailRequest) ([]string, error) {
	var messageIds []int64

	e.loggerFactory.Entry(ctx).
		WithTime(time.Now()).
		WithField("templateName", request.TemplateName).
		Info("SendEmailRequest")

	resp, err := e.restyClient.R().
		SetBody(request).
		SetResult(&messageIds). // 結果をmessageIdsに格納する。
		Post(SEND_REQUEST_PATH)

	if err != nil {
		return nil, errors.Wrap(err, "error sending email request")
	}

	e.loggerFactory.Entry(ctx).
		WithTime(time.Now()).
		WithField("status code", resp.StatusCode()).
		WithField("responseBody", resp.String()).
		WithField("responseTime", resp.Time()).
		Info("SendEmailRequest")

	if resp.IsError() {
		return nil, fmt.Errorf("received non-2xx response: %s", resp.Status())
	}

	if len(messageIds) == 0 {
		return nil, errors.New("received empty messageIds")
	}

	msgIds := lo.Map(messageIds, func(v int64, _ int) string {
		return strconv.FormatInt(v, 10)
	})

	return msgIds, nil
}

func (e emailpkg) GetAllRequest(ctx context.Context, query GetAllQuery) (EmailListResponse, error) {
	var emailListResponse EmailListResponse

	resp, err := e.restyClient.R().
		SetQueryParamsFromValues(query.ToUrlValues()).
		SetResult(&emailListResponse). // 結果をmessageIdsに格納する。
		Get(GET_REQUEST_PATH)

	if err != nil {
		return emailListResponse, errors.Wrap(err, "error sending request for email list")
	}

	if resp.IsError() {
		return emailListResponse, fmt.Errorf("received non-2xx response: %s", resp.Status())
	}

	return emailListResponse, nil
}

func (e emailpkg) GetByIdRequest(ctx context.Context, id string) (EmailResponse, error) {
	var emailResponse EmailResponse

	// path에 id를 넣어준다.
	getByIdPath := fmt.Sprintf("%s/%s", GET_REQUEST_PATH, url.PathEscape(id))

	resp, err := e.restyClient.R().
		SetResult(&emailResponse). // 結果をmessageIdsに格納する。
		Get(getByIdPath)

	if err != nil {
		return emailResponse, errors.Wrap(err, "error sending request for an email")
	}

	if resp.IsError() {
		return emailResponse, fmt.Errorf("received non-2xx response: %s", resp.Status())
	}

	return emailResponse, nil
}

func (e emailpkg) GetTemplate(ctx context.Context, templateName string) (MailTemplateResponse, error) {
	var mailTemplateResponse MailTemplateResponse

	resp, err := e.restyClient.R().
		SetQueryParam("templateName", templateName).
		SetResult(&mailTemplateResponse).
		Get(GET_TEMPLATE_PATH)

	if err != nil {
		return mailTemplateResponse, errors.Wrapf(err, "error sending request for getting template: %s", templateName)
	}

	if resp.IsError() {
		return mailTemplateResponse, fmt.Errorf("received non-2xx response: %s", resp.Status())
	}

	return mailTemplateResponse, nil
}
