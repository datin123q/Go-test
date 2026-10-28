package controllers

import (
	"strconv"
	"testing"

	client "github.com/gmoaozora/gmo-aozora-api-go/corporateclient"
	"github.com/stretchr/testify/require"

	"github.com/illuminarean/energy-shares-backend-repo/pkg/feequote"
)

func Test_convertFeeResponseToDTO_Success(t *testing.T) {
	res := client.TransferFeeResponse{
		TotalFee: "300",
		TransferFeeDetails: []client.TransferFeeDetail{
			{ItemId: "1", TransferFee: "100"},
			{ItemId: "2", TransferFee: "200"},
		},
	}
	mapping := map[string]string{"1": "10", "2": "20"}

	dto, err := convertFeeResponseToDTO(res, mapping)
	require.NoError(t, err)
	require.Equal(t, int64(300), dto.TotalFee)
	require.Len(t, dto.Details, 2)

	expect := []feequote.WithdrawalFeeDetail{
		{WithdrawalID: 10, TransferFee: 100},
		{WithdrawalID: 20, TransferFee: 200},
	}
	require.Equal(t, expect, dto.Details)
}

func Test_convertFeeResponseToDTO_MissingMap(t *testing.T) {
	res := client.TransferFeeResponse{
		TotalFee: "100",
		TransferFeeDetails: []client.TransferFeeDetail{
			{ItemId: "1", TransferFee: "100"},
		},
	}
	_, err := convertFeeResponseToDTO(res, map[string]string{})
	require.Error(t, err)
}

func Test_convertFeeResponseToDTO_ParseError(t *testing.T) {
	res := client.TransferFeeResponse{
		TotalFee: "xx",
		TransferFeeDetails: []client.TransferFeeDetail{
			{ItemId: "1", TransferFee: "100"},
		},
	}
	m := map[string]string{"1": "10"}
	_, err := convertFeeResponseToDTO(res, m)
	require.Error(t, err)

	// bad fee
	res.TotalFee = "0"
	res.TransferFeeDetails[0].TransferFee = "abc"
	_, err = convertFeeResponseToDTO(res, m)
	require.Error(t, err)

	// bad withdrawal id parse
	res.TransferFeeDetails[0].TransferFee = "100"
	m["1"] = "not_int"
	_, err = convertFeeResponseToDTO(res, m)
	require.Error(t, err)

	// optional: make sure mapping requires numeric strings
	m["1"] = strconv.FormatInt(10, 10)
	res.TransferFeeDetails[0].ItemId = "missing"
	_, err = convertFeeResponseToDTO(res, m)
	require.Error(t, err)
}
