package mysql

import (
	"context"
	"time"

	"github.com/illuminarean/energy-shares-backend-repo/domain"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/jsonutils"
	"xorm.io/xorm"
)

var (
	mockPolicyHistory1 = domain.PolicyInfoHistory{
		Id:              1,
		PolicyId:        1,
		SubstanceType:   string(domain.POLICY_HISTORY_CONTENTS),
		Name:            "ADMIN",
		Email:           "admin@energyshares.jp",
		IsPosted:        true,
		Label:           jsonutils.JSONTime(time.Date(2024, 2, 9, 0, 0, 0, 0, time.UTC)),
		PrevContent:     "prev content",
		CurrContent:     "curr content",
		PrevContentHtml: "prev content html",
		CurrContentHtml: "curr content html",
		PrevUrl:         "prev url",
		CurrUrl:         "curr url",
		ContentType:     string(domain.POLICY_CHANGED_CONTENTS),
	}

	mockPolicyHistory2 = domain.PolicyInfoHistory{
		Id:              2,
		PolicyId:        1,
		SubstanceType:   string(domain.POLICY_HISTORY_CONTENTS),
		Name:            "ADMIN",
		Email:           "admin@energyshares.jp",
		IsPosted:        false,
		Label:           jsonutils.JSONTime(time.Time{}),
		PrevContent:     "prev content",
		CurrContent:     "curr content",
		PrevContentHtml: "prev content html",
		CurrContentHtml: "curr content html",
		PrevUrl:         "prev url",
		CurrUrl:         "curr url",
		ContentType:     string(domain.POLICY_CHANGED_CONTENTS),
	}

	mockPolicyHistory3 = domain.PolicyInfoHistory{
		Id:              3,
		PolicyId:        2,
		SubstanceType:   string(domain.POLICY_CHANGED_CONTENTS),
		Name:            "ADMIN",
		Email:           "admin@energyshares.jp",
		IsPosted:        false,
		Label:           jsonutils.JSONTime(time.Time{}),
		PrevContent:     "prev content",
		CurrContent:     "curr content",
		PrevContentHtml: "prev content html",
		CurrContentHtml: "curr content html",
		PrevUrl:         "prev url",
		CurrUrl:         "curr url",
		ContentType:     string(domain.POLICY_CHANGED_CONTENTS),
	}
)

type dbFactoryMock struct{}

func (d dbFactoryMock) EnergySharesDB(ctx context.Context) (xorm.Interface, error) {
	xormEngine, err := xorm.NewEngine("sqlite3", ":memory:")
	if err != nil {
		return nil, err
	}

	xormEngine.ShowSQL(true)

	xormEngine.Sync(new(domain.PolicyInfoHistory))

	xormEngine.Insert(mockPolicyHistory1, mockPolicyHistory2, mockPolicyHistory3)

	return xormEngine, nil
}

func (d dbFactoryMock) MyNumberDB(ctx context.Context) (xorm.Interface, error) {
	panic("implement me")
}
