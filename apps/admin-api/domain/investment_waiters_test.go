package domain

import (
	"encoding/json"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	rootdomain "github.com/illuminarean/energy-shares-backend-repo/domain"
)

func TestInvestmentWaitersResponse_StatusSwaggerEnum(t *testing.T) {
	field, ok := reflect.TypeOf(InvestmentWaitersResponse{}).FieldByName("Status")
	require.True(t, ok)

	assert.Contains(t, field.Tag.Get("swagger"), "enum(WAITING|END_FUNDING|CANCELLED|WINNING|WINNING_AFTER_SUCCESS|WINNING_AFTER_CANCELED|WINNING_AFTER_FAILED|PROPOSAL)")
}

func TestInvestmentWaitersResponse_FromInvestmentWaiters_ShareId(t *testing.T) {
	tests := []struct {
		name        string
		shareId     string
		wantShareId interface{}
	}{
		{
			name:        "Share作成済みの場合はshare_idを返す",
			shareId:     "share-1",
			wantShareId: "share-1",
		},
		{
			name:        "Share未作成の場合はshare_idをnullで返す",
			wantShareId: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			response := (InvestmentWaitersResponse{}).FromInvestmentWaiters(rootdomain.InvestmentWaiters{
				ShareId: tt.shareId,
			})
			data, err := json.Marshal(response)
			require.NoError(t, err)

			var body map[string]interface{}
			require.NoError(t, json.Unmarshal(data, &body))
			assert.Equal(t, tt.wantShareId, body["share_id"])
		})
	}
}
