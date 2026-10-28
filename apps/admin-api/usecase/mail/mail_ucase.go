package mail

import (
	"context"
	"encoding/json"
	"strings"

	"github.com/sirupsen/logrus"

	adomain "github.com/illuminarean/energy-share-core-api/domain"
	cdomain "github.com/illuminarean/energy-shares-backend-repo/clean-arch-domain"
	"github.com/illuminarean/energy-shares-backend-repo/domain"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/config"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/emailutils"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/factory"
)

type mailUsecase struct {
	cfg                  *config.Config
	logger               factory.LoggerFactory
	emailpkg             emailutils.EmailRequester
	mailFailedRepository cdomain.MailFailedRepository
}

func New(
	cfg *config.Config,
	logger factory.LoggerFactory,
	emailpkg emailutils.EmailRequester,
	mailFailedRepository cdomain.MailFailedRepository,
) adomain.MailUsecase {
	return &mailUsecase{
		cfg:                  cfg,
		logger:               logger,
		emailpkg:             emailpkg,
		mailFailedRepository: mailFailedRepository,
	}
}

func (m mailUsecase) GetAll(ctx context.Context, request adomain.MailGetAllRequest) (adomain.MailGetAllResponse, error) {

	queryParam := request.ToGetAllQuery(request.Env)
	response, err := m.emailpkg.GetAllRequest(ctx, queryParam)
	if err != nil {
		return adomain.MailGetAllResponse{}, err
	}

	var mailGetAllResponse adomain.MailGetAllResponse
	if err := mailGetAllResponse.FromEmailListResponse(response); err != nil {
		return mailGetAllResponse, err
	}
	return mailGetAllResponse, nil
}

func (m mailUsecase) GetOne(ctx context.Context, request adomain.MailGetOneRequest) (adomain.MailDetailResponse, error) {

	response, err := m.emailpkg.GetByIdRequest(ctx, request.MailId)
	if err != nil {
		return adomain.MailDetailResponse{}, err
	}

	var mailDetailResponse adomain.MailDetailResponse
	if err := mailDetailResponse.FromEmailResponse(response); err != nil {
		return mailDetailResponse, err
	}

	return mailDetailResponse, nil
}

func (m mailUsecase) GetAllFailed(ctx context.Context, request adomain.MailGetAllFailedRequest) (adomain.MailGetAllFailedResponse, error) {

	total, items, err := m.mailFailedRepository.GetAll(ctx,
		cdomain.MailFailedQueryOptionFunc().WithReceiver(request.Receiver),
		cdomain.MailFailedQueryOptionFunc().WithTemplateName(request.TemplateName),
		cdomain.MailFailedQueryOptionFunc().WithSortBy(request.Sortby),
		cdomain.MailFailedQueryOptionFunc().WithOrder(request.Order),
		cdomain.MailFailedQueryOptionFunc().WithOffset(request.SkipCount),
		cdomain.MailFailedQueryOptionFunc().WithLimit(request.MaxResultCount),
	)

	if err != nil {
		return adomain.MailGetAllFailedResponse{}, err
	}

	mailFailedItems := make([]adomain.MailFailedItem, 0, len(items))
	for i := range items {
		mailFailedItem := adomain.MailFailedItem{
			Id:           items[i].Id,
			TemplateName: items[i].TemplateName,
			Created:      items[i].Created,
		}
		if err := json.Unmarshal(items[i].Receivers, &mailFailedItem.Receivers); err != nil {
			// MIR-1365: 保存されている宛先データ(JSON)が壊れていて読み取れなかった
			// ときに、以前は宛先をそのままエラーログへ出していた。異常時にしか
			// 通らない経路だが、宛先そのものは出さず、バイト長だけにする。
			m.logger.Entry(ctx).WithFields(
				logrus.Fields{
					"id":                    items[i].Id,
					"template_name":         items[i].TemplateName,
					"created":               items[i].Created,
					"receivers_byte_length": len(items[i].Receivers),
				}).Error("get all failed mail request")
			continue
		}
		mailFailedItems = append(mailFailedItems, mailFailedItem)
	}

	return adomain.MailGetAllFailedResponse{
		TotalCount: total,
		Items:      mailFailedItems,
	}, nil
}

func (m mailUsecase) GetTemplate(ctx context.Context, request adomain.MailGetTemplateRequest) (adomain.MailGetTemplateResponse, error) {

	templateName := domain.TEMPLATE_TYPE(request.TemplateType).TemplateName()
	response, err := m.emailpkg.GetTemplate(ctx, templateName)
	if err != nil {
		return adomain.MailGetTemplateResponse{}, err
	}

	var mailGetTemplateResponse adomain.MailGetTemplateResponse
	mailGetTemplateResponse.FromMailTemplate(response)
	mailGetTemplateResponse.Content = strings.ReplaceAll(mailGetTemplateResponse.Content, "{{ServiceUrl}}", m.cfg.ServiceConfig.Url.UserServiceDomain)

	return mailGetTemplateResponse, nil
}
