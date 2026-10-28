package clean_arch_domain

import (
	"context"

	"github.com/illuminarean/energy-shares-backend-repo/domain"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/jsonutils"
)

type UserService interface {
	// user_svc
	IsDuplicateEmail(ctx context.Context, email string) (bool, error)
	IsDuplicateEmailIncludeDeleted(ctx context.Context, email string) (bool, error)

	// user_password_svc
	ChangePassword(ctx context.Context, user domain.User, password string) error
	IsValidPassword(ctx context.Context, user domain.User, password string) bool
}

type UserRepository interface {
	Create(ctx context.Context, user domain.User) (int64, error)
	Update(ctx context.Context, user domain.User) error
	// UpdateLastLoginDate は last_login_date 列だけを更新する。Update(全列書き戻し)は
	// 読み込みから書き込みまでの間に別処理(eKYC Webhookなど)が書いた変更を
	// 巻き戻してしまうため、ログインではこちらを使う(MIR-1373)。
	UpdateLastLoginDate(ctx context.Context, id int64, lastLoginDate jsonutils.JSONTime) error
	GetAll(ctx context.Context, options ...UserQueryOption) (int64, []domain.User, error)
	GetById(ctx context.Context, id int64) (domain.User, error)
	GetByEmail(ctx context.Context, email string) (domain.User, error)
	GetByEmailIncludeDeleted(ctx context.Context, email string) (domain.User, error)
}

type UserQueryOption = func(*UserQueryOptionFields)

func UserQueryOptionFunc() UserQueryOptionFields {
	return UserQueryOptionFields{}
}

// TODO: not implemented all fields yet
type UserQueryOptionFields struct {
	// id                  *string
	// name                *string
	// email               *string
	investorTypes *[]string
	// auditStatus         *string
	// requestId           *string
	// startDate           string
	// endDate             string
	// investGrade         bool
	// basicInfo           bool
	// withdrawal          bool
	// loanDeposit         bool
	// equityDeposit       bool
	newsletterAgreement *bool
	ids                 []int64
	QueryOptions
}

func (qo UserQueryOptionFields) WithInvestorTypes(investorTypes []string) func(*UserQueryOptionFields) {
	return func(qo *UserQueryOptionFields) {
		qo.investorTypes = &investorTypes
	}
}

func (qo UserQueryOptionFields) WithNewsletterAgreement(newsletterAgreement bool) func(*UserQueryOptionFields) {
	return func(qo *UserQueryOptionFields) {
		qo.newsletterAgreement = &newsletterAgreement
	}
}

func (qo UserQueryOptionFields) WithSortBy(sortBy string) func(*UserQueryOptionFields) {
	return func(qo *UserQueryOptionFields) {
		qo.sortBy = &sortBy
	}
}

func (qo UserQueryOptionFields) WithOrder(order string) func(*UserQueryOptionFields) {
	return func(qo *UserQueryOptionFields) {
		qo.order = &order
	}
}

func (qo UserQueryOptionFields) WithOffset(offset int) func(*UserQueryOptionFields) {
	return func(qo *UserQueryOptionFields) {
		qo.offset = &offset
	}
}

func (qo UserQueryOptionFields) WithLimit(limit int) func(*UserQueryOptionFields) {
	return func(qo *UserQueryOptionFields) {
		qo.limit = &limit
	}
}

func (qo UserQueryOptionFields) WithIds(ids []int64) func(*UserQueryOptionFields) {
	return func(qo *UserQueryOptionFields) {
		qo.ids = ids
	}
}

func (qo UserQueryOptionFields) InvestorType() *[]string {
	return qo.investorTypes
}

func (qo UserQueryOptionFields) NewsletterAgreement() *bool {
	return qo.newsletterAgreement
}

func (qo UserQueryOptionFields) Ids() []int64 {
	return qo.ids
}
