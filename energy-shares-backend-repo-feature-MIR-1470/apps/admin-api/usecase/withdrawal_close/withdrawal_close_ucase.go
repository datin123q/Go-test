package withdrawal_close

import (
	"bytes"
	"context"
	"encoding/csv"

	adomain "github.com/illuminarean/energy-share-core-api/domain"
	cdomain "github.com/illuminarean/energy-shares-backend-repo/clean-arch-domain"
	"github.com/illuminarean/energy-shares-backend-repo/domain"
	"github.com/pkg/errors"
	"github.com/samber/lo"
	"golang.org/x/text/encoding/japanese"
	"golang.org/x/text/transform"
)

const (
	MENU_WITHDRAWAL_DOWNLOAD = "menu.manage_accounting.manage_withdrawal_download"
)

type withdrawalCloseUsecase struct {
	operatorHistoryRepository cdomain.OperatorHistoryRepository
	withdrawalRepository      cdomain.WithdrawalRepository
	withdrawalCloseRepository cdomain.WithdrawalCloseRepository
	withdrawalService         cdomain.WithdrawalService
}

func New(
	operatorHistoryRepository cdomain.OperatorHistoryRepository,
	withdrawalRepository cdomain.WithdrawalRepository,
	withdrawalCloseRepository cdomain.WithdrawalCloseRepository,
	withdrawalService cdomain.WithdrawalService,
) adomain.WithdrawalCloseUsecase {
	return &withdrawalCloseUsecase{
		operatorHistoryRepository: operatorHistoryRepository,
		withdrawalRepository:      withdrawalRepository,
		withdrawalCloseRepository: withdrawalCloseRepository,
		withdrawalService:         withdrawalService,
	}
}

func (u withdrawalCloseUsecase) GetAll(ctx context.Context, request adomain.WithdrawalCloseGetAllRequest) (adomain.WithdrawalCloseGetAllResponse, error) {
	opts := cdomain.WithdrawalCloseQueryOptionFunc()
	totalCount, closedWithdrawals, err := u.withdrawalCloseRepository.GetAll(ctx,
		opts.WithSortBy(request.Sortby),
		opts.WithOrder(request.Order),
		opts.WithOffset(request.SkipCount),
		opts.WithLimit(request.MaxResultCount),
	)
	if err != nil {
		return adomain.WithdrawalCloseGetAllResponse{}, errors.Wrap(err, "failed to get all closed withdrawals")
	}

	items := lo.Map(closedWithdrawals, func(wd domain.WithdrawalCloseDetail, _ int) adomain.WithdrawalCloseGetAllItem {
		return adomain.WithdrawalCloseGetAllItem{
			Id:           wd.Id,
			Total:        wd.Total,
			Amount:       wd.Amount,
			Fee:          wd.Fee,
			CloserId:     wd.CloserId,
			CloserName:   wd.CloserName,
			CloseDate:    wd.CloseDate,
			TransferDate: wd.TransferDate,
			Status:       wd.Status,
		}
	})

	return adomain.WithdrawalCloseGetAllResponse{
		Items:      items,
		TotalCount: totalCount,
	}, nil
}

func (u withdrawalCloseUsecase) GetWithdrawals(ctx context.Context, request adomain.WithdrawalCloseGetWithdrawalsRequest) (adomain.WithdrawalCloseGetWithdrawalsResponse, error) {

	// WithdrawalClose.Id 에 속하는 Withdrawal 목록 조회
	opts := cdomain.WithdrawalQueryOptionFunc()
	totalCount, closedWithdrawals, err := u.withdrawalRepository.GetAll(ctx,
		opts.WithSortBy(request.Sortby),
		opts.WithOrder(request.Order),
		opts.WithOffset(request.SkipCount),
		opts.WithLimit(request.MaxResultCount),
		opts.WithCloseId(lo.ToPtr(request.Id)),
	)
	if err != nil {
		return adomain.WithdrawalCloseGetWithdrawalsResponse{}, errors.Wrap(err, "failed to get all withdrawals")
	}

	items := lo.Map(closedWithdrawals, func(wd domain.WithdrawalDetail, _ int) adomain.WithdrawalCloseGetWithdrawalsItem {
		return adomain.WithdrawalCloseGetWithdrawalsItem{
			Id:           wd.Withdrawal.Id,
			InvestorId:   wd.Withdrawal.InvestorId,
			InvestorName: wd.Withdrawal.InvestorName,
			Amount:       wd.Withdrawal.Amount,
			Fee:          wd.Withdrawal.Fee,
			RequestDate:  wd.Withdrawal.RequestDate,
			CompleteDate: wd.Withdrawal.CompleteDate,
			Status:       wd.Withdrawal.Status,
			Created:      wd.Withdrawal.Created,
			CloseDate:    wd.Withdrawal.CloseDate,
		}
	})

	return adomain.WithdrawalCloseGetWithdrawalsResponse{
		Items:      items,
		TotalCount: totalCount,
	}, nil
}

func (u withdrawalCloseUsecase) Download(ctx context.Context, request adomain.WithdrawalCloseDownloadRequest) (adomain.WithdrawalCloseDownloadResponse, error) {

	// WithdrawalClose.Id 에 속하는 Withdrawal 목록 조회
	opts := cdomain.WithdrawalQueryOptionFunc()
	_, closedWithdrawals, err := u.withdrawalRepository.GetAll(ctx,
		opts.WithSortBy(request.Sortby),
		opts.WithOrder(request.Order),
		opts.WithOffset(request.SkipCount),
		opts.WithLimit(request.MaxResultCount),
		opts.WithCloseId(lo.ToPtr(request.Id)),
	)
	if err != nil {
		return adomain.WithdrawalCloseDownloadResponse{}, errors.Wrap(err, "failed to get all withdrawals")
	}

	// 데이터를 엑셀 파일로 변환
	buf := new(bytes.Buffer)

	// GMO 업로드용 파일은 ShiftJIS 로 인코딩 해달라는 요청이 있었음 (참고: ESJP-171)
	writer := csv.NewWriter(transform.NewWriter(buf, japanese.ShiftJIS.NewEncoder()))
	writer.UseCRLF = true

	csvData := u.withdrawalService.ExportCSV(ctx, closedWithdrawals)
	if err = writer.WriteAll(csvData); err != nil {
		return adomain.WithdrawalCloseDownloadResponse{}, errors.Wrap(err, "failed to write csv")
	}
	writer.Flush()

	// 운영 로그 생성
	u.operatorHistoryRepository.Create(ctx, domain.OperatorHistory{
		UserId: request.JWTUser.Id,
		Email:  request.JWTUser.Email,
		Name:   request.JWTUser.Name,
		Menu:   MENU_WITHDRAWAL_DOWNLOAD,
		Ip:     request.ClientIp,
		State:  domain.DOWNLOAD,
	})

	return adomain.WithdrawalCloseDownloadResponse{Data: buf.Bytes()}, nil
}

func (u withdrawalCloseUsecase) RequestTransfer(ctx context.Context, request adomain.WithdrawalCloseRequestTransferRequest) (adomain.WithdrawalCloseRequestTransferResponse, error) {
	// TODO implement me
	panic("implement me")
}

func (u withdrawalCloseUsecase) CloseAllWithdrawals(ctx context.Context, request adomain.WithdrawalCloseCloseAllWithdrawalsRequest) (adomain.WithdrawalCloseCloseAllWithdrawalsResponse, error) {
	// TODO implement me
	panic("implement me")
}
