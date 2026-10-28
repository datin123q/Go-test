package sms_msg

import (
	"bytes"
	"context"
	"time"

	"github.com/pkg/errors"

	cdomain "github.com/illuminarean/energy-shares-backend-repo/clean-arch-domain"
	"github.com/illuminarean/energy-shares-backend-repo/domain"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/factory"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/jsonutils"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/snsutils"
)

type smsMsgService struct {
	loggerFactory            factory.LoggerFactory
	snsPkg                   snsutils.Sns
	mapSmsReceiverRepository cdomain.MapSmsReceiverRepository
	smsMsgRepository         cdomain.SmsMsgRepository
}

func New(
	loggerFactory factory.LoggerFactory,
	snsPkg snsutils.Sns,
	mapSmsReceiverRepository cdomain.MapSmsReceiverRepository,
	smsMsgRepository cdomain.SmsMsgRepository,
) cdomain.SmsMsgService {
	return &smsMsgService{
		loggerFactory:            loggerFactory,
		snsPkg:                   snsPkg,
		mapSmsReceiverRepository: mapSmsReceiverRepository,
		smsMsgRepository:         smsMsgRepository,
	}
}

func (s smsMsgService) SendSms(ctx context.Context, request cdomain.SendSmsRequest) error {
	var err error

	content, err := s.getSmsTemplateContent(ctx, request)
	if err != nil {
		return errors.Wrap(err, "failed to get sms template content")
	}

	smsMsg := domain.SmsMsg{
		Title:      request.Title,
		Sender:     request.Sender,
		Status:     domain.REQUEST_SMS,
		SmsType:    request.SmsType,
		Count:      int64(1),
		SmsContent: content,
		SendDate:   jsonutils.JSONTime(time.Now()),
		Name:       request.SenderName,
	}
	smsMsg.Id, err = s.smsMsgRepository.Create(ctx, smsMsg)
	if err != nil {
		return errors.Wrap(err, "failed to create sms_msg")
	}

	mapSmsReceiver := domain.MapSmsReceiver{
		SmsId:    smsMsg.Id,
		Receiver: request.Receiver,
	}
	if _, err := s.mapSmsReceiverRepository.Create(ctx, mapSmsReceiver); err != nil {
		return errors.Wrap(err, "failed to create map_sms_receiver")
	}

	smsMsg.MsgId, err = s.snsPkg.SendSMS(ctx, request.Receiver, &smsMsg)
	if err != nil {
		smsMsg.Status = domain.FAIL_SMS
		s.loggerFactory.Entry(ctx).
			WithField("method", "SendSMS").
			WithField("sms_id", smsMsg.Id).
			WithError(err).
			Error("failed to send sms")
	} else {
		smsMsg.Status = domain.SUCCESS_SMS
	}

	if err := s.smsMsgRepository.Update(ctx, smsMsg); err != nil {
		return err
	}

	return nil
}

func (s smsMsgService) getSmsTemplateContent(ctx context.Context, request cdomain.SendSmsRequest) (string, error) {

	template, err := s.snsPkg.LoadTextTemplate(ctx, request.Template)
	if err != nil {
		return "", errors.Wrap(err, "failed to load sms template")
	}

	var tpl bytes.Buffer
	if err := template.Execute(&tpl, request.Details); err != nil {
		return "", err
	}

	return tpl.String(), nil
}
