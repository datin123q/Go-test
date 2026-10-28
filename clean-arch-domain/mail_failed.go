package clean_arch_domain

import (
	"context"

	"github.com/illuminarean/energy-shares-backend-repo/domain"
)

type MailFailedRepository interface {
	Create(ctx context.Context, mailFailed domain.MailFailed) (int64, error)
	GetAll(ctx context.Context, options ...MailFailedQueryOption) (int64, []domain.MailFailed, error)
}

type MailFailedQueryOption = func(*MailFailedQueryOptionFields)

func MailFailedQueryOptionFunc() MailFailedQueryOptionFields {
	return MailFailedQueryOptionFields{}
}

type MailFailedQueryOptionFields struct {
	receiver     *string
	templateName *string
	QueryOptions
}

func (qo MailFailedQueryOptionFields) WithReceiver(receiver *string) func(*MailFailedQueryOptionFields) {
	return func(qo *MailFailedQueryOptionFields) {
		qo.receiver = receiver
	}
}

func (qo MailFailedQueryOptionFields) WithTemplateName(templateName *string) func(*MailFailedQueryOptionFields) {
	return func(qo *MailFailedQueryOptionFields) {
		qo.templateName = templateName
	}
}

func (qo MailFailedQueryOptionFields) Receiver() *string {
	return qo.receiver
}

func (qo MailFailedQueryOptionFields) TemplateName() *string {
	return qo.templateName
}

func (qo MailFailedQueryOptionFields) WithSortBy(sortBy string) func(*MailFailedQueryOptionFields) {
	return func(qo *MailFailedQueryOptionFields) {
		qo.sortBy = &sortBy
	}
}

func (qo MailFailedQueryOptionFields) WithOrder(order string) func(*MailFailedQueryOptionFields) {
	return func(qo *MailFailedQueryOptionFields) {
		qo.order = &order
	}
}

func (qo MailFailedQueryOptionFields) WithOffset(offset int) func(*MailFailedQueryOptionFields) {
	return func(qo *MailFailedQueryOptionFields) {
		qo.offset = &offset
	}
}

func (qo MailFailedQueryOptionFields) WithLimit(limit int) func(*MailFailedQueryOptionFields) {
	return func(qo *MailFailedQueryOptionFields) {
		qo.limit = &limit
	}
}
