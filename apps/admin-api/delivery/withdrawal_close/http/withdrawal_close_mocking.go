package http

import (
	"time"

	adomain "github.com/illuminarean/energy-share-core-api/domain"
	"github.com/illuminarean/energy-shares-backend-repo/domain"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/jsonutils"
)

var (
	mockWithdrawalCloseGetAllItem1 = adomain.WithdrawalCloseGetAllItem{
		Id:           1,
		Total:        100000,
		Amount:       100000,
		Fee:          0,
		CloserId:     1,
		CloserName:   "ホン ギルドン",
		CloseDate:    jsonutils.JSONTime(time.Date(2024, 12, 19, 0, 0, 0, 0, time.Local)),
		TransferDate: jsonutils.JSONTime(time.Date(2025, 2, 9, 0, 0, 0, 0, time.Local)),
		Status:       domain.WITHDRAWAL_REQUEST_STATUS_COMPLETE,
	}
)
