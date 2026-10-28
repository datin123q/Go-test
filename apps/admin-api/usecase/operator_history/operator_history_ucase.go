package operator_history

import (
	"bytes"
	"context"
	"encoding/csv"
	"fmt"

	"github.com/pkg/errors"

	adomain "github.com/illuminarean/energy-share-core-api/domain"
	cdomain "github.com/illuminarean/energy-shares-backend-repo/clean-arch-domain"
	"github.com/illuminarean/energy-shares-backend-repo/domain"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/config"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/i18nutils"
)

type operatorHistoryUsecase struct {
	operatorHistoryKsqlRepository cdomain.OperatorHistoryKsqlRepository
	cfg                           *config.Config
	translator                    i18nutils.Translator
}

func New(
	operatorHistoryKsqlRepository cdomain.OperatorHistoryKsqlRepository,
	cfg *config.Config,
	translator i18nutils.Translator,
) adomain.OperatorHistoryUsecase {
	return &operatorHistoryUsecase{
		operatorHistoryKsqlRepository: operatorHistoryKsqlRepository,
		cfg:                           cfg,
		translator:                    translator,
	}
}

func (u operatorHistoryUsecase) GetAll(ctx context.Context, request adomain.OperatorHistoryGetAllRequest) (adomain.OperatorHistoryGetAllResponse, error) {
	options := []cdomain.OperatorHistoryKsqlQueryOption{
		cdomain.OperatorHistoryKsqlQueryOptionFunc().WithStartDate(request.StartDate),
		cdomain.OperatorHistoryKsqlQueryOptionFunc().WithEndDate(request.EndDate),
		cdomain.OperatorHistoryKsqlQueryOptionFunc().WithLimit(request.MaxResultCount),
	}

	count, items, err := u.operatorHistoryKsqlRepository.GetAll(ctx, options...)
	if err != nil {
		return adomain.OperatorHistoryGetAllResponse{}, err
	}

	return adomain.OperatorHistoryGetAllResponse{
		TotalCount: count,
		Items:      items,
	}, nil
}

func (u operatorHistoryUsecase) Download(ctx context.Context, request adomain.OperatorHistoryDownloadRequest) (adomain.OperatorHistoryDownloadResponse, error) {
	options := []cdomain.OperatorHistoryKsqlQueryOption{
		cdomain.OperatorHistoryKsqlQueryOptionFunc().WithStartDate(request.StartDate),
		cdomain.OperatorHistoryKsqlQueryOptionFunc().WithEndDate(request.EndDate),
	}

	_, items, err := u.operatorHistoryKsqlRepository.GetAll(ctx, options...)
	if err != nil {
		return adomain.OperatorHistoryDownloadResponse{}, err
	}

	csvData := u.exportCSV(items)
	var buf bytes.Buffer
	writer := csv.NewWriter(&buf)

	if err = writer.WriteAll(csvData); err != nil {
		return adomain.OperatorHistoryDownloadResponse{}, errors.Wrap(err, "failed to write csv data")
	}

	return adomain.OperatorHistoryDownloadResponse{Data: buf.Bytes()}, nil
}

func (u operatorHistoryUsecase) exportCSV(items []domain.OperatorHistoryRowResponse) cdomain.CSVData {
	csvHeaders := u.translator.Translates(domain.OperatorHistoryCSVHeaders)
	records := [][]string{csvHeaders}

	for _, item := range items {
		date := item.Date
		actor := fmt.Sprintf("%s", item.Actor)
		operation := item.Operation
		log := fmt.Sprintf("%+v", item.Log)
		ip := item.IP

		record := []string{date, actor, operation, log, ip}
		records = append(records, record)
	}

	return records
}

func (u operatorHistoryUsecase) GetMenuCodes(ctx context.Context, req adomain.GetCodeListRequest) (adomain.GetCodeListResponse, error) {
	loc := req.Locale
	codes := domain.OperatorHistory{}.GetMenuCodes(loc)
	return adomain.GetCodeListResponse{Codes: codes}, nil
}

func (u operatorHistoryUsecase) GetLogCodes(ctx context.Context, req adomain.GetCodeListRequest) (adomain.GetCodeListResponse, error) {
	loc := req.Locale
	codes := domain.OperatorHistory{}.GetLogCodeList(loc)
	return adomain.GetCodeListResponse{Codes: codes}, nil
}
