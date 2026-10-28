package domain

import (
	"context"
	"fmt"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/pkg/errors"
	"github.com/samber/lo"

	cdomain "github.com/illuminarean/energy-shares-backend-repo/clean-arch-domain"
	"github.com/illuminarean/energy-shares-backend-repo/domain"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/emailutils"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/jsonutils"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/jwthelper"
)

type MailUsecase interface {
	GetAll(ctx context.Context, request MailGetAllRequest) (MailGetAllResponse, error)
	GetOne(ctx context.Context, request MailGetOneRequest) (MailDetailResponse, error)
	GetAllFailed(ctx context.Context, request MailGetAllFailedRequest) (MailGetAllFailedResponse, error)
	GetTemplate(ctx context.Context, request MailGetTemplateRequest) (MailGetTemplateResponse, error)
}

const (
	MAIL_SENT    = "SENT"    // 통합 메일 서비스에서 발송 성공
	MAIL_FAILED  = "FAILED"  // 통합 메일 서비스에서 발송 실패
	MAIL_PENDING = "PENDING" // 통합 메일 서비스에서 발송을 위해 대기중
)

type MailGetAllRequest struct {
	JWTUser  jwthelper.User
	ClientIp string
	MailGetAllQueryParam
	Env string
	cdomain.SearchInput
}

func (m MailGetAllRequest) ToGetAllQuery(env string) emailutils.GetAllQuery {

	// ESJP 에서 사용하는 날짜시간 포맷을 통합 메일 서비스에서 사용하는 날짜시간 포맷으로 변환
	// 포맷 자체는 이미 검증이 되었다고 가정한다.
	toQueryDate := func(datetime string) *string {
		if datetime == "" {
			return nil
		}

		// Parse the input datetime string
		parsedTime, err := jsonutils.ParseDateTimeJST(time.DateTime, datetime)
		if err != nil {
			return nil // 여기서는 에러가 없어야 한다. 이미 검증이 되었다고 가정한다.
		}

		// 통합 메일 서비스 지원 포맷으로 변경
		formattedDate := parsedTime.Format("2006-01-02T15:04:05")

		return &formattedDate
	}

	getAllQuery := emailutils.GetAllQuery{
		ServiceName: "ESJP",
		ServiceEnv:  env,
		Subject:     m.Subject,
		Sender:      m.Sender,
		Receiver:    m.Receiver,
		StartDate:   toQueryDate(m.StartDate),
		EndDate:     toQueryDate(m.EndDate),
		Status:      m.Status,
	}

	// sort
	if m.Sortby != "" && m.Order != "" {
		// 통합 메일 서비스의 쿼리로 변경
		if m.Sortby == "send_date" {
			m.Sortby = "sendDateTime"
		} else if m.Sortby == "created" {
			m.Sortby = "createdDate"
		}
		getAllQuery.Sort = lo.ToPtr(m.Sortby + "," + m.Order)
	}

	// pagination - page는 0부터 시작한다.
	if m.SkipCount >= 0 && m.MaxResultCount > 0 { // 이미 검증은 되었다 판단하지만 혹시 모르니 추가 검증
		getAllQuery.Page = m.SkipCount / m.MaxResultCount
		getAllQuery.Size = m.MaxResultCount
	}

	return getAllQuery

}

func (req MailGetAllRequest) Validate() error {
	validate := validator.New()

	if err := validate.Struct(req.MailGetAllQueryParam); err != nil {
		return errors.Wrap(cdomain.ErrBadRequest, err.Error())
	}

	if err := validate.Struct(req.SearchInput); err != nil {
		return errors.Wrap(cdomain.ErrBadRequest, err.Error())
	}

	validSortbyValues := map[string]bool{
		"id":        true, // id
		"subject":   true, // subject
		"sender":    true, // sender
		"receiver":  true, // receiver
		"status":    true, // status
		"send_date": true, // sendDateTime - 다름
		"created":   true, // createdDate - 다름
	}

	if req.Sortby != "" {
		if !validSortbyValues[req.Sortby] {
			return fmt.Errorf(
				"invalid sortby value: %s. valid values: id, subject, sender, receiver, status, send_date, created",
				req.Sortby)
		}
	}

	return nil
}

type MailGetAllQueryParam struct {
	Subject   *string `query:"subject" description:"제목" example:"가입을 환영합니다"`
	Sender    *string `query:"sender" description:"발신자" example:"info@energyshares.jp"`
	Receiver  *string `query:"receiver" description:"수신자" example:"customer@exmaple.com"`
	Status    *string `query:"status" description:"메일 발신 상태(SENT, FAILED, PENDING)" example:"SENT" validate:"omitempty,oneof=SENT FAILED PENDING"`
	StartDate string  `query:"start_date" description:"시작일" format:"yyyy-mm-dd HH:mm:ss" example:"2024-10-10 00:00:00"`
	EndDate   string  `query:"end_date" description:"종료일" format:"yyyy-mm-dd HH:mm:ss" example:"2024-10-15 23:59:59"`
}

type MailGetAllResponse struct {
	TotalCount int64          `json:"total_count" description:"메일 개수" example:"10"`
	Items      []MailResponse `json:"items"`
}

type MailResponse struct {
	Id           int64                `json:"id" description:"메일 아이디" example:"100"`
	Subject      string               `json:"subject" description:"메일 제목" example:"[EnergyShares] 会員登録用の認証メールです。"`
	Sender       string               `json:"sender" description:"발신자" example:"info@energyshares.jp"`
	TemplateType domain.TEMPLATE_TYPE `json:"template_type" description:"템플릿 타입" example:"SIGN_UP"`
	MailType     domain.MAIL_TYPE     `json:"mail_type" description:"메일 발신 우선순위(FAST, NORMAL)" example:"FAST"`
	SendDate     jsonutils.JSONTime   `json:"send_date" description:"보낸 날짜" example:"2024-10-10 17:21:23"`
	Count        int64                `json:"count" description:"수신자 수" example:"3"`
	Status       domain.MAIL_STATUS   `json:"status" description:"메일 발신 상태(SENT, FAILED, PENDING)" example:"SENT"`
}

func (r *MailGetAllResponse) FromEmailListResponse(emailListResponse emailutils.EmailListResponse) error {
	items := make([]MailResponse, 0, len(emailListResponse.Content))
	for _, email := range emailListResponse.Content {
		mailType, err := toMailType(email.SendType)
		if err != nil {
			return err
		}
		mailStatus, err := toMailStatus(email.Status)
		if err != nil {
			return err
		}

		item := MailResponse{
			Id:           email.ID,
			Subject:      email.Subject,
			Sender:       email.Sender,
			TemplateType: domain.TEMPLATE_TYPE(email.TemplateName),
			MailType:     mailType,
			SendDate:     email.SendDateTime,
			Count:        int64(len(email.Receivers)),
			Status:       mailStatus,
		}
		items = append(items, item)
	}

	r.TotalCount = emailListResponse.TotalElements
	r.Items = items
	return nil
}

type MailGetOneRequest struct {
	MailId   string `param:"id" path:"id" description:"메일 상세 조회를 위한 메일 ID" example:"1" format:"int64"`
	JWTUser  jwthelper.User
	ClientIp string
}

type MailDetailResponse struct {
	Id           int64                `json:"id" description:"메일 아이디" example:"100"`
	Subject      string               `json:"subject" description:"메일 제목" example:"[EnergyShares] 会員登録用の認証メールです。"`
	Name         string               `json:"name" swagger:"발신자 이름" example:""` // TODO: 발송자 이름 적용 여부 논의 필요
	Sender       string               `json:"sender" description:"발신자 메일" example:"info@energyshares.jp"`
	Receiver     []string             `json:"receiver" description:"수신 메일 목록" example:"[\"recipient1@domain.com\", \"recipient2@domain.com\"]"`
	Content      string               `json:"content" description:"메일 내용" example:"<p>안녕하세요</p>\r\n<p>회원가입을 환영합니다.</p>\r\n<p>감사합니다.</p>"`
	TemplateType domain.TEMPLATE_TYPE `json:"template_type" description:"템플릿 타입" example:"SIGN_UP"`
	MailType     domain.MAIL_TYPE     `json:"mail_type" description:"메일 발신 우선순위(FAST, NORMAL)" example:"FAST"`
	Status       domain.MAIL_STATUS   `json:"status" description:"메일 발신 상태(SENT, FAILED, PENDING)" example:"SENT"`
	Count        int64                `json:"count" description:"수신 수" example:"3"`
	SendDate     jsonutils.JSONTime   `json:"send_date" description:"보낸 날짜" example:"2024-10-10 17:21:23"`
	Updated      jsonutils.JSONTime   `json:"updated" example:"2024-10-10 17:21:23"`
	Created      jsonutils.JSONTime   `json:"created" example:"2024-10-10 17:21:23"`
}

func (r *MailDetailResponse) FromEmailResponse(emailResponse emailutils.EmailResponse) error {

	mailType, err := toMailType(emailResponse.SendType)
	if err != nil {
		return err
	}
	mailStatus, err := toMailStatus(emailResponse.Status)
	if err != nil {
		return err
	}

	r.Id = emailResponse.ID
	r.Subject = emailResponse.Subject
	r.Name = "" // TODO: 발송자 이름 적용 여부 논의 필요
	r.Sender = emailResponse.Sender
	r.Receiver = emailResponse.Receivers
	r.Content = emailResponse.Body
	r.TemplateType = domain.TEMPLATE_TYPE(emailResponse.TemplateName)
	r.MailType = mailType
	r.Status = mailStatus
	r.Count = int64(len(emailResponse.Receivers))
	r.SendDate = emailResponse.SendDateTime
	if emailResponse.CreatedDate != nil {
		r.Created = *emailResponse.CreatedDate
	}
	if emailResponse.ModifiedDate != nil {
		r.Updated = *emailResponse.ModifiedDate
	}

	return nil
}

func toMailStatus(status string) (domain.MAIL_STATUS, error) {
	var mailStatus domain.MAIL_STATUS

	switch status {
	case MAIL_SENT:
		mailStatus = domain.MAIL_SENT
	case MAIL_FAILED:
		mailStatus = domain.MAIL_FAILED
	case MAIL_PENDING:
		mailStatus = domain.MAIL_PENDING
	default:
		return mailStatus, errors.New("invalid mail status: " + status)
	}

	return mailStatus, nil
}

func toMailType(sendType string) (domain.MAIL_TYPE, error) {
	var mailType domain.MAIL_TYPE

	switch sendType {
	case "NORMAL":
		mailType = domain.MAIL_SEND_PRIORITY_NORMAL
	case "FAST":
		mailType = domain.MAIL_SEND_PRIORITY_FAST
	default:
		return mailType, errors.New("invalid mail type: " + sendType)
	}

	return mailType, nil
}

type MailGetAllFailedRequest struct {
	cdomain.SearchInput
	MailGetAllFailedQueryParam
	JWTUser  jwthelper.User
	ClientIp string
}

type MailGetAllFailedQueryParam struct {
	TemplateName *string `query:"template_name" description:"템플릿 이름" example:"NORMAL_MAIL"`
	Receiver     *string `query:"receiver" description:"수신자" example:"info@example.com"`
}

type MailGetAllFailedResponse struct {
	TotalCount int64            `json:"total_count" description:"메일 개수" example:"10"`
	Items      []MailFailedItem `json:"items" description:"발신 요청 최종 실패 메일들"`
}

type MailFailedItem struct {
	Id           int64              `json:"id" description:"메일 아이디" example:"100"`
	Receivers    []string           `json:"receivers" description:"수신 메일 목록" example:"[\"recipient1@domain.com\", \"recipient2@domain.com\"]"`
	TemplateName string             `json:"template_name" description:"템플릿 이름, 템플릿 타입에 prefix, suffix가 붙은 값" example:"NORMAL_MAIL"`
	Created      jsonutils.JSONTime `json:"created" description:"생성 일시" example:"2024-10-10 17:21:23"`
}

type MailGetTemplateRequest struct {
	TemplateType string `param:"template_type" path:"template_type" description:"템플릿 타입. default 로 표기하면 NORMAL_MAIL 을 회신" example:"NORMAL_MAIL" format:"string"`
	JWTUser      jwthelper.User
	ClientIp     string
}

type MailGetTemplateResponse struct {
	Subject      string `json:"subject" description:"메일 제목" example:"[EnergyShares] {{Subject}}"`
	Content      string `json:"content" description:"메일 템플릿"`
	TemplateName string `json:"template_name" description:"템플릿 이름. 템플릿 타입에 prefix, suffix가 붙은 값" example:"NORMAL_MAIL"`
}

func (r *MailGetTemplateResponse) FromMailTemplate(response emailutils.MailTemplateResponse) {
	r.Subject = response.Subject
	r.Content = response.Content
	r.TemplateName = response.TemplateName
}
