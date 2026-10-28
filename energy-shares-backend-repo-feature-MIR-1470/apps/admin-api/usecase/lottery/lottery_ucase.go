package lottery

import (
	"bytes"
	"context"
	"encoding/csv"
	"fmt"
	"io"
	"mime/multipart"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/pkg/errors"
	"github.com/samber/lo"
	"github.com/segmentio/ksuid"
	"golang.org/x/text/language"
	"golang.org/x/text/message"

	adomain "github.com/illuminarean/energy-share-core-api/domain"
	adminservice "github.com/illuminarean/energy-share-core-api/service"
	cdomain "github.com/illuminarean/energy-shares-backend-repo/clean-arch-domain"
	"github.com/illuminarean/energy-shares-backend-repo/domain"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/actorcontext"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/config"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/factory"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/jsonutils"
)

type lotteryUsecase struct {
	cfg                       *config.Config
	lotteryRepository         cdomain.LotteryApplicationRepository
	lotteryHistoryRepository  cdomain.LotteryHistoryRepository
	projectRepository         cdomain.ProjectRepository
	userRepository            cdomain.UserRepository
	operatorHistoryRepository cdomain.OperatorHistoryRepository
	mailService               cdomain.MailService
	lotteryService            cdomain.LotteryService
	txManager                 factory.TxManager
}

const lotteryListDownloadMaxCount = 10000

var lotteryCSVUploadHeaders = []string{
	"抽選ID",
	"抽選ステータス",
	"投資家名",
	"メールアドレス",
	"ファンド名",
	"申込口数",
	"申込金額",
	"当選総額",
	"当選後キャンセル額",
	"投資可能金額",
	"申込日時",
	"当落",
	"当選させたい金額",
}

var lotteryCSVStatusLabels = map[string]domain.LOTTERY_STATUS{
	"APPLIED":         domain.LOTTERY_STATUS_APPLIED,
	"申込中":             domain.LOTTERY_STATUS_APPLIED,
	"WINNING":         domain.LOTTERY_STATUS_WINNING,
	"当選":              domain.LOTTERY_STATUS_WINNING,
	"PARTIAL_WINNING": domain.LOTTERY_STATUS_PARTIAL_WINNING,
	"一部当選":            domain.LOTTERY_STATUS_PARTIAL_WINNING,
	"LOST":            domain.LOTTERY_STATUS_LOST,
	"落選":              domain.LOTTERY_STATUS_LOST,
	"CANCELED":        domain.LOTTERY_STATUS_CANCELED,
	"キャンセル":           domain.LOTTERY_STATUS_CANCELED,
}

type lotteryCSVUploadRow struct {
	rowNumber             int
	lotteryApplicationID  string
	lotteryStatus         string
	investorName          string
	email                 string
	fundName              string
	appliedQty            string
	appliedAmount         string
	winningAmount         string
	canceledWinningAmount string
	availableAmount       string
	appliedAt             string
	result                string
	desiredWinningAmount  string
}

type lotteryCSVUploadValidatedRow struct {
	rowNumber            int
	lotteryApplicationID string
	expectedStatus       domain.LOTTERY_STATUS
	expectedAppliedAt    string
	inputAppliedQty      int64
	inputWinningQty      int64
	processType          domain.LOTTERY_CSV_UPLOAD_PROCESS_TYPE
	status               domain.LOTTERY_CSV_UPLOAD_DETAIL_STATUS
}

type lotteryCSVUploadValidationError struct {
	rowNumber int
	field     string
	message   string
}

func (e lotteryCSVUploadValidationError) String() string {
	return fmt.Sprintf("row %d [%s]: %s", e.rowNumber, e.field, e.message)
}

func winningQtyTotal(application domain.LotteryApplication) int64 {
	return lo.SumBy(application.Winnings, func(share domain.Share) int64 {
		return share.Qty
	})
}

func hasChanceToWin(status domain.LOTTERY_STATUS) bool {
	return status == domain.LOTTERY_STATUS_APPLIED ||
		status == domain.LOTTERY_STATUS_LOST ||
		status == domain.LOTTERY_STATUS_PARTIAL_WINNING
}

func lotteryWinProposalQtyRangeByApplication(
	application domain.LotteryApplication,
	projectWinningQtyTotal int64,
) (int64, int64) {
	projectRemainingQty := application.TargetShareQty - projectWinningQtyTotal
	if projectRemainingQty < 0 {
		projectRemainingQty = 0
	}

	applicationRemainingQty := application.Qty - winningQtyTotal(application)
	if applicationRemainingQty < 0 {
		applicationRemainingQty = 0
	}

	maxWinningQty := projectRemainingQty
	if applicationRemainingQty < maxWinningQty {
		maxWinningQty = applicationRemainingQty
	}
	if maxWinningQty == 0 {
		return 0, 0
	}

	minWinningQty := application.MinShareQty
	if maxWinningQty < minWinningQty {
		minWinningQty = maxWinningQty
	}

	return maxWinningQty, minWinningQty
}

func canCancelLotteryApplication(application domain.LotteryApplication, projectStatus domain.PROJECT_STATUS) bool {
	return application.CanMarkAsCanceled() &&
		(projectStatus == domain.OPEN_FUNDING || projectStatus == domain.END_FUNDING)
}

func validateWinningProject(project *domain.Project) error {
	if project == nil {
		return errors.Wrap(cdomain.ErrItemNotFound, "project not found")
	}
	if project.SelectionType != domain.PROJECT_SELECTION_LOTTERY {
		return errors.Wrap(cdomain.ErrBadRequest, "project selection type is not lottery")
	}
	if project.ProjectStatus != domain.END_FUNDING {
		return errors.Wrap(cdomain.ErrBadRequest, "project status should be END_FUNDING")
	}
	return nil
}

func validateWinningApplication(application domain.LotteryApplication, qty int64) error {
	if !application.CanceledAt.IsZero() {
		return errors.Wrap(cdomain.ErrBadRequest, "lottery application is canceled")
	}
	if qty <= 0 {
		return errors.Wrap(cdomain.ErrBadRequest, "qty should be greater than zero")
	}
	if qty > application.Qty {
		return errors.Wrap(cdomain.ErrBadRequest, "qty should not exceed lottery application qty")
	}
	if winningQtyTotal(application)+qty > application.Qty {
		return errors.Wrap(cdomain.ErrBadRequest, "winning qty exceeds lottery application qty")
	}
	return nil
}

func validateWinningUser(user *domain.User) error {
	if user == nil {
		return errors.Wrap(cdomain.ErrItemNotFound, "user not found")
	}
	if user.BlockInvest {
		return errors.Wrap(cdomain.ErrBadRequest, "user is blocked from investing")
	}
	if user.AuditStatus != domain.COMPLETED_AUDIT {
		return errors.Wrap(cdomain.ErrBadRequest, "user audit status is not completed")
	}
	return nil
}

func validateLosingApplication(application domain.LotteryApplication) error {
	if !application.CanceledAt.IsZero() {
		return errors.Wrap(cdomain.ErrBadRequest, "lottery application is canceled")
	}
	if application.Qty-winningQtyTotal(application) <= 0 {
		return errors.Wrap(cdomain.ErrBadRequest, "lottery application has no remaining qty to lose")
	}
	return nil
}

func lotteryStatusAfterWinning(application domain.LotteryApplication, additionalWinningQty int64) domain.LOTTERY_STATUS {
	if winningQtyTotal(application)+additionalWinningQty < application.Qty {
		return domain.LOTTERY_STATUS_PARTIAL_WINNING
	}
	return domain.LOTTERY_STATUS_WINNING
}

func lotteryStatusAfterLosing(application domain.LotteryApplication) domain.LOTTERY_STATUS {
	if application.Status == domain.LOTTERY_STATUS_PARTIAL_WINNING {
		return domain.LOTTERY_STATUS_PARTIAL_WINNING
	}
	if application.Status == domain.LOTTERY_STATUS_WINNING {
		return domain.LOTTERY_STATUS_WINNING
	}
	return domain.LOTTERY_STATUS_LOST
}

func (l lotteryUsecase) createLotteryHistory(
	ctx context.Context,
	applicationID string,
	prevStatus domain.LOTTERY_STATUS,
	newStatus domain.LOTTERY_STATUS,
	actorID int64,
) error {
	return l.lotteryHistoryRepository.Create(ctx, domain.LotteryHistory{
		ApplicationId: applicationID,
		PrevStatus:    prevStatus,
		NewStatus:     newStatus,
		ActorId:       actorID,
	})
}

func parseLotteryCSVUploadRows(fileHeader *multipart.FileHeader) ([]lotteryCSVUploadRow, error) {
	if fileHeader == nil {
		return nil, errors.Wrap(cdomain.ErrBadRequest, "csv file is required")
	}

	file, err := fileHeader.Open()
	if err != nil {
		return nil, errors.Wrap(err, "failed to open csv file")
	}
	defer file.Close()

	content, err := io.ReadAll(io.LimitReader(file, adomain.LotteryCSVUploadMaxFileSizeBytes+1))
	if err != nil {
		return nil, errors.Wrap(err, "failed to read csv file")
	}
	if int64(len(content)) > adomain.LotteryCSVUploadMaxFileSizeBytes {
		return nil, adomain.NewLotteryCSVValidationError([]adomain.LotteryCSVValidationErrorDetail{
			{
				Field:   "file",
				Message: "CSVファイルサイズは10MB以下にしてください",
			},
		})
	}
	if !utf8.Valid(content) {
		return nil, errors.Wrap(cdomain.ErrBadRequest, "csv file must be utf-8")
	}
	content = bytes.TrimPrefix(content, []byte{0xEF, 0xBB, 0xBF})

	reader := csv.NewReader(bytes.NewReader(content))
	reader.FieldsPerRecord = -1
	reader.TrimLeadingSpace = false

	records, err := reader.ReadAll()
	if err != nil {
		return nil, errors.Wrap(cdomain.ErrBadRequest, "failed to parse csv file")
	}
	if len(records) == 0 {
		return nil, errors.Wrap(cdomain.ErrBadRequest, "csv file is empty")
	}
	if len(records[0]) != len(lotteryCSVUploadHeaders) {
		return nil, errors.Wrap(cdomain.ErrBadRequest, "csv header column count is invalid")
	}
	for idx, header := range lotteryCSVUploadHeaders {
		if records[0][idx] != header {
			return nil, errors.Wrap(cdomain.ErrBadRequest, "csv header does not match template")
		}
	}
	if len(records) == 1 {
		return nil, errors.Wrap(cdomain.ErrBadRequest, "csv file has no data rows")
	}
	if len(records)-1 > adomain.LotteryCSVUploadMaxRows {
		return nil, adomain.NewLotteryCSVValidationError([]adomain.LotteryCSVValidationErrorDetail{
			{
				Field:   "file",
				Message: "申込件数は10000件以下にしてください",
			},
		})
	}

	rows := make([]lotteryCSVUploadRow, 0, len(records)-1)
	for i, record := range records[1:] {
		rowNumber := i + 2
		if len(record) != len(lotteryCSVUploadHeaders) {
			return nil, errors.Wrapf(cdomain.ErrBadRequest, "csv row %d column count is invalid", rowNumber)
		}

		trimmed := make([]string, len(record))
		isEmptyRow := true
		for j, column := range record {
			trimmed[j] = strings.TrimSpace(column)
			if trimmed[j] != "" {
				isEmptyRow = false
			}
		}
		if isEmptyRow {
			return nil, errors.Wrapf(cdomain.ErrBadRequest, "csv row %d is empty", rowNumber)
		}
		rows = append(rows, lotteryCSVUploadRow{
			rowNumber:             rowNumber,
			lotteryApplicationID:  trimmed[0],
			lotteryStatus:         trimmed[1],
			investorName:          trimmed[2],
			email:                 trimmed[3],
			fundName:              trimmed[4],
			appliedQty:            trimmed[5],
			appliedAmount:         trimmed[6],
			winningAmount:         trimmed[7],
			canceledWinningAmount: trimmed[8],
			availableAmount:       trimmed[9],
			appliedAt:             trimmed[10],
			result:                trimmed[11],
			desiredWinningAmount:  trimmed[12],
		})
	}

	return rows, nil
}

func parseLotteryCSVInteger(value string, allowBlank bool) (*int64, error) {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		if allowBlank {
			return nil, nil
		}
		return nil, errors.New("value is required")
	}
	normalized := strings.ReplaceAll(trimmed, ",", "")
	normalized = strings.TrimSuffix(normalized, "円")
	normalized = strings.TrimSpace(normalized)

	parsed, err := strconv.ParseInt(normalized, 10, 64)
	if err != nil {
		return nil, errors.Wrap(err, "value must be an integer")
	}
	return &parsed, nil
}

func parseLotteryCSVStatus(value string) (domain.LOTTERY_STATUS, bool) {
	status, ok := lotteryCSVStatusLabels[strings.TrimSpace(value)]
	return status, ok
}

func parseLotteryCSVAppliedAt(value string) (time.Time, error) {
	return jsonutils.ParseDateTimeJST("2006/1/2 15:04", value)
}

func formatLotteryCSVAppliedAt(t time.Time) string {
	return t.In(time.FixedZone("Asia/Tokyo", 9*60*60)).Format("2006/1/2 15:04")
}

func escapeCSVFormulaString(value string) string {
	if value == "" {
		return value
	}

	switch value[0] {
	case '=', '+', '-', '@':
		return "'" + value
	default:
		return value
	}
}

func appendLotteryCSVValidationError(
	validationErrors *[]lotteryCSVUploadValidationError,
	rowNumber int,
	field string,
	message string,
) {
	*validationErrors = append(*validationErrors, lotteryCSVUploadValidationError{
		rowNumber: rowNumber,
		field:     field,
		message:   message,
	})
}

func joinLotteryCSVValidationErrors(validationErrors []lotteryCSVUploadValidationError) error {
	if len(validationErrors) == 0 {
		return nil
	}

	details := make([]adomain.LotteryCSVValidationErrorDetail, 0, len(validationErrors))
	for _, validationError := range validationErrors {
		details = append(details, adomain.LotteryCSVValidationErrorDetail{
			Row:     validationError.rowNumber,
			Field:   validationError.field,
			Message: validationError.message,
		})
	}
	return adomain.NewLotteryCSVValidationError(details)
}

func (l lotteryUsecase) validateLotteryCSVUploadRows(
	ctx context.Context,
	project domain.Project,
	rows []lotteryCSVUploadRow,
) ([]lotteryCSVUploadValidatedRow, error) {
	validationErrors := make([]lotteryCSVUploadValidationError, 0)
	validatedRows := make([]lotteryCSVUploadValidatedRow, 0, len(rows))
	applicationIDRows := make(map[string][]int)

	userCache := make(map[int64]domain.User)

	for _, row := range rows {
		applicationIDRows[row.lotteryApplicationID] = append(applicationIDRows[row.lotteryApplicationID], row.rowNumber)
		if row.lotteryApplicationID == "" {
			appendLotteryCSVValidationError(&validationErrors, row.rowNumber, "抽選ID", "抽選IDは必須です")
		}

		switch row.result {
		case "当選", "落選", "":
		default:
			appendLotteryCSVValidationError(&validationErrors, row.rowNumber, "当落", "当落は「当選」「落選」または空欄で入力してください")
		}

		appliedQty, parseAppliedQtyErr := parseLotteryCSVInteger(row.appliedQty, false)
		if parseAppliedQtyErr != nil {
			appendLotteryCSVValidationError(&validationErrors, row.rowNumber, "申込口数", "申込口数は整数で入力してください")
		}

		appliedAmount, parseAppliedAmountErr := parseLotteryCSVInteger(row.appliedAmount, false)
		if parseAppliedAmountErr != nil {
			appendLotteryCSVValidationError(&validationErrors, row.rowNumber, "申込金額", "申込金額は整数で入力してください")
		}

		currentWinningAmount, parseWinningAmountErr := parseLotteryCSVInteger(row.winningAmount, false)
		if parseWinningAmountErr != nil {
			appendLotteryCSVValidationError(&validationErrors, row.rowNumber, "当選総額", "当選総額は整数で入力してください")
		}

		desiredWinningAmount, parseDesiredAmountErr := parseLotteryCSVInteger(row.desiredWinningAmount, true)
		switch row.result {
		case "当選":
			if parseDesiredAmountErr != nil || desiredWinningAmount == nil {
				appendLotteryCSVValidationError(&validationErrors, row.rowNumber, "当選させたい金額", "当選時は当選させたい金額を整数で入力してください")
			} else if *desiredWinningAmount <= 0 {
				appendLotteryCSVValidationError(&validationErrors, row.rowNumber, "当選させたい金額", "当選時は当選させたい金額に0より大きい金額を入力してください")
			}
		case "落選":
			if parseDesiredAmountErr != nil || desiredWinningAmount == nil || *desiredWinningAmount != 0 {
				appendLotteryCSVValidationError(&validationErrors, row.rowNumber, "当選させたい金額", "落選時は当選させたい金額に0を入力してください")
			}
		case "":
			if strings.TrimSpace(row.desiredWinningAmount) != "" {
				appendLotteryCSVValidationError(&validationErrors, row.rowNumber, "当選させたい金額", "当落が空欄の場合は当選させたい金額も空欄にしてください")
			}
		}

		parsedAppliedAt, appliedAtErr := parseLotteryCSVAppliedAt(row.appliedAt)
		if appliedAtErr != nil {
			appendLotteryCSVValidationError(&validationErrors, row.rowNumber, "申込日時", "申込日時はyyyy/M/d HH:mm形式で入力してください")
		}

		application, appErr := l.lotteryRepository.GetById(ctx, row.lotteryApplicationID)
		if appErr != nil {
			if errors.Is(appErr, cdomain.ErrItemNotFound) {
				appendLotteryCSVValidationError(&validationErrors, row.rowNumber, "抽選ID", "対象の抽選申込が存在しません")
				validatedRows = append(validatedRows, lotteryCSVUploadValidatedRow{
					rowNumber:            row.rowNumber,
					lotteryApplicationID: row.lotteryApplicationID,
				})
				continue
			}
			return nil, errors.Wrap(appErr, "failed to get lottery application")
		}
		if application.ProjectId != project.Id {
			appendLotteryCSVValidationError(&validationErrors, row.rowNumber, "抽選ID", "対象プロジェクト配下の申込ではありません")
		}

		user, ok := userCache[application.UserId]
		if !ok {
			fetchedUser, userErr := l.userRepository.GetById(ctx, application.UserId)
			if userErr != nil {
				return nil, errors.Wrap(userErr, "failed to get user")
			}
			user = fetchedUser
			userCache[application.UserId] = user
		}

		existingWinningAmount := int64(0)
		for _, share := range application.Winnings {
			existingWinningAmount += share.Qty * share.Price
		}
		expectedStatus, statusMatched := parseLotteryCSVStatus(row.lotteryStatus)
		if !statusMatched || expectedStatus != application.Status {
			appendLotteryCSVValidationError(&validationErrors, row.rowNumber, "抽選ステータス", "抽選ステータスが現在の申込情報と一致しません")
		}
		if row.investorName != application.UserName {
			appendLotteryCSVValidationError(&validationErrors, row.rowNumber, "投資家名", "投資家名が現在の申込情報と一致しません")
		}
		if row.email != user.Email {
			appendLotteryCSVValidationError(&validationErrors, row.rowNumber, "メールアドレス", "メールアドレスが現在の申込情報と一致しません")
		}
		if row.fundName != application.ProjectName {
			appendLotteryCSVValidationError(&validationErrors, row.rowNumber, "ファンド名", "ファンド名が現在の申込情報と一致しません")
		}
		if appliedQty != nil && *appliedQty != application.Qty {
			appendLotteryCSVValidationError(&validationErrors, row.rowNumber, "申込口数", "申込口数が現在の申込情報と一致しません")
		}
		if appliedAmount != nil && *appliedAmount != application.Amount {
			appendLotteryCSVValidationError(&validationErrors, row.rowNumber, "申込金額", "申込金額が現在の申込情報と一致しません")
		}
		if currentWinningAmount != nil && *currentWinningAmount != existingWinningAmount {
			appendLotteryCSVValidationError(&validationErrors, row.rowNumber, "当選総額", "当選総額が現在の申込情報と一致しません")
		}
		if appliedAtErr == nil && row.appliedAt != formatLotteryCSVAppliedAt(application.AppliedAt.Time()) {
			appendLotteryCSVValidationError(&validationErrors, row.rowNumber, "申込日時", "申込日時が現在の申込情報と一致しません")
		}

		processType := domain.LOTTERY_CSV_UPLOAD_PROCESS_TYPE_SKIP
		status := domain.LOTTERY_CSV_UPLOAD_DETAIL_STATUS_SKIPPED
		inputWinningQty := int64(0)

		switch row.result {
		case "当選":
			processType = domain.LOTTERY_CSV_UPLOAD_PROCESS_TYPE_WIN
			status = domain.LOTTERY_CSV_UPLOAD_DETAIL_STATUS_PENDING

			if desiredWinningAmount != nil && application.UnitPrice > 0 {
				if *desiredWinningAmount%application.UnitPrice != 0 {
					appendLotteryCSVValidationError(&validationErrors, row.rowNumber, "当選させたい金額", "当選させたい金額は1口単価で割り切れる必要があります")
				} else {
					inputWinningQty = *desiredWinningAmount / application.UnitPrice
					remainingQty := application.Qty - winningQtyTotal(application)
					if inputWinningQty > remainingQty {
						appendLotteryCSVValidationError(&validationErrors, row.rowNumber, "当選させたい金額", "今回追加当選口数が未当選残口数を超えています")
					}
				}
			}
		case "落選":
			processType = domain.LOTTERY_CSV_UPLOAD_PROCESS_TYPE_LOSE
			status = domain.LOTTERY_CSV_UPLOAD_DETAIL_STATUS_PENDING

			existingWinningQty := winningQtyTotal(application)
			if existingWinningQty > 0 && existingWinningQty < application.Qty {
				appendLotteryCSVValidationError(&validationErrors, row.rowNumber, "当落", "一部当選済みの申込に対して落選は指定できません")
			}
		}

		inputAppliedQty := int64(0)
		if appliedQty != nil {
			inputAppliedQty = *appliedQty
		}

		validatedRows = append(validatedRows, lotteryCSVUploadValidatedRow{
			rowNumber:            row.rowNumber,
			lotteryApplicationID: row.lotteryApplicationID,
			expectedStatus:       expectedStatus,
			expectedAppliedAt:    formatLotteryCSVAppliedAt(parsedAppliedAt),
			inputAppliedQty:      inputAppliedQty,
			inputWinningQty:      inputWinningQty,
			processType:          processType,
			status:               status,
		})
	}

	for applicationID, rowNumbers := range applicationIDRows {
		if applicationID == "" || len(rowNumbers) <= 1 {
			continue
		}
		for _, rowNumber := range rowNumbers {
			appendLotteryCSVValidationError(&validationErrors, rowNumber, "抽選ID", "同一CSV内で抽選IDが重複しています")
		}
	}

	projectCurrentShareQty := project.TargetShareQty - project.RemainingQty
	totalWinningQty := int64(0)
	totalWinningAmount := int64(0)
	winningRowNumbers := make([]int, 0)
	for _, row := range validatedRows {
		if row.processType != domain.LOTTERY_CSV_UPLOAD_PROCESS_TYPE_WIN || row.status != domain.LOTTERY_CSV_UPLOAD_DETAIL_STATUS_PENDING {
			continue
		}
		totalWinningQty += row.inputWinningQty
		totalWinningAmount += row.inputWinningQty * project.SharePrice
		winningRowNumbers = append(winningRowNumbers, row.rowNumber)
	}

	if totalWinningQty > 0 && projectCurrentShareQty+totalWinningQty > project.TargetShareQty {
		for _, rowNumber := range winningRowNumbers {
			appendLotteryCSVValidationError(&validationErrors, rowNumber, "当選させたい金額", "CSV内の当選指定合計がファンドの募集口数上限を超えています")
		}
	}
	if totalWinningAmount > 0 && project.CurrentShareAmount+totalWinningAmount > project.TargetShareAmount {
		for _, rowNumber := range winningRowNumbers {
			appendLotteryCSVValidationError(&validationErrors, rowNumber, "当選させたい金額", "CSV内の当選指定合計がファンドの募集金額上限を超えています")
		}
	}

	if err := joinLotteryCSVValidationErrors(validationErrors); err != nil {
		return nil, err
	}

	return validatedRows, nil
}

func New(
	cfg *config.Config,
	lotteryRepository cdomain.LotteryApplicationRepository,
	lotteryHistoryRepository cdomain.LotteryHistoryRepository,
	projectRepository cdomain.ProjectRepository,
	userRepository cdomain.UserRepository,
	operatorHistoryRepository cdomain.OperatorHistoryRepository,
	mailService cdomain.MailService,
	lotteryService cdomain.LotteryService,
	txManager factory.TxManager,
) adomain.LotteryUseCase {
	return &lotteryUsecase{
		cfg:                       cfg,
		lotteryRepository:         lotteryRepository,
		lotteryHistoryRepository:  lotteryHistoryRepository,
		projectRepository:         projectRepository,
		userRepository:            userRepository,
		operatorHistoryRepository: operatorHistoryRepository,
		mailService:               mailService,
		lotteryService:            lotteryService,
		txManager:                 txManager,
	}
}

func (l lotteryUsecase) GetAll(ctx context.Context, request adomain.LotteryGetAllRequest) (adomain.LotteryGetAllResponse, error) {
	var err error

	if request.RequestDateFrom != nil {
		_, err = jsonutils.ParseJSONTimeJST(time.DateTime, lo.FromPtr(request.RequestDateFrom))
		if err != nil {
			return adomain.LotteryGetAllResponse{}, errors.Wrap(err, "failed to parse startDate")
		}
	}
	if request.RequestDateTo != nil {
		_, err = jsonutils.ParseJSONTimeJST(time.DateTime, lo.FromPtr(request.RequestDateTo))
		if err != nil {
			return adomain.LotteryGetAllResponse{}, errors.Wrap(err, "failed to parse endDate")
		}
	}

	opts := cdomain.LotteryApplicationQueryOptionFunc()
	totalCount, lotteries, err := l.lotteryRepository.GetAll(ctx,
		opts.WithSortBy(request.Sortby),
		opts.WithOrder(request.Order),
		opts.WithOffset(request.SkipCount),
		opts.WithLimit(request.MaxResultCount),
		opts.WithProjectId(lo.FromPtr(request.ProjectId)),
		opts.WithUserId(lo.FromPtr(request.UserId)),
		opts.WithUserName(lo.FromPtr(request.UserName)),
		opts.WithProjectName(lo.FromPtr(request.ProjectName)),
		opts.WithAmountMin(lo.FromPtr(request.AmountMin)),
		opts.WithAmountMax(lo.FromPtr(request.AmountMax)),
		opts.WithLTERequestDate(lo.FromPtr(request.RequestDateTo)),
		opts.WithGTERequestDate(lo.FromPtr(request.RequestDateFrom)),
		opts.WithStatus(lo.FromPtr(request.Status)),
	)
	if err != nil {
		return adomain.LotteryGetAllResponse{}, errors.Wrap(err, "failed to get all lotteries")
	}

	projectIDs := lo.Uniq(lo.Map(lotteries, func(lottery domain.LotteryApplication, _ int) int64 {
		return lottery.ProjectId
	}))
	winningQtyTotalsByProject, err := l.lotteryRepository.GetWinningQtyTotalsByProjectID(ctx, projectIDs)
	if err != nil {
		return adomain.LotteryGetAllResponse{}, errors.Wrap(err, "failed to load winning qty totals for lotteries")
	}

	items := lo.Map(lotteries, func(lottery domain.LotteryApplication, _ int) adomain.LotteryItem {
		// 当選総額は複数当選の場合に合計する
		winningSumAmount := int64(0)
		winnings := lo.Map(lottery.Winnings, func(winning domain.Share, _ int) adomain.LotteryWinning {
			winningSumAmount += winning.Qty * winning.Price
			return adomain.LotteryWinning{
				ShareID:     winning.Id,
				ProjectId:   winning.ProjectId,
				UserId:      winning.OwnerId,
				Qty:         winning.Qty,
				UnitPrice:   winning.Price,
				Amount:      winning.Price * winning.Qty,
				WinningedAt: winning.InvestDate,
				ShareStatus: winning.InvestStatus,
			}
		})
		maxWinningQty, minWinningQty := lotteryWinProposalQtyRangeByApplication(
			lottery,
			winningQtyTotalsByProject[lottery.ProjectId],
		)

		return adomain.LotteryItem{
			Id:                       lottery.Id,
			UserId:                   lottery.UserId,
			UserType:                 lottery.UserType,
			UserName:                 lottery.UserName,
			UserEmail:                lottery.UserEmail,
			UserAccountBalance:       lo.FromPtr(lottery.UserAccountBalance),
			ProjectId:                lottery.ProjectId,
			ProjectName:              lottery.ProjectName,
			ProjectStatus:            lottery.ProjectStatus,
			Qty:                      lottery.Qty,
			UnitPrice:                lottery.UnitPrice,
			Amount:                   lottery.Amount,
			Status:                   lottery.Status,
			AppliedAt:                lottery.AppliedAt,
			WinningAt:                lottery.WinningAt,
			LostAt:                   lottery.LostAt,
			CanceledAt:               lottery.CanceledAt,
			LotteryWinProposalMaxQty: maxWinningQty,
			LotteryWinProposalMinQty: minWinningQty,
			WinningAmount:            winningSumAmount,
			CanceledWinningAmount:    lottery.CanceledWinningAmount(),
			Winnings:                 winnings,
		}
	})

	// 現在ページの申込金額合計
	totalCurrentPageLotteryAmount := lo.SumBy(lotteries, func(lottery domain.LotteryApplication) int64 {
		return lottery.Amount
	})

	// return response
	return adomain.LotteryGetAllResponse{
		TotalCount:  totalCount,
		TotalAmount: totalCurrentPageLotteryAmount,
		Items:       items,
	}, nil
}

func (l lotteryUsecase) UploadCSV(ctx context.Context, request adomain.LotteryCSVUploadRequest) (adomain.LotteryCSVUploadResponse, error) {
	project, err := l.projectRepository.GetById(ctx, request.ProjectID)
	if err != nil {
		return adomain.LotteryCSVUploadResponse{}, errors.Wrap(err, "failed to get project")
	}
	if err := validateWinningProject(&project); err != nil {
		return adomain.LotteryCSVUploadResponse{}, err
	}

	hasActiveJob, err := (domain.LotteryCSVUploadJob{}).HasActiveJobByProjectID(ctx, request.ProjectID)
	if err != nil {
		return adomain.LotteryCSVUploadResponse{}, errors.Wrap(err, "failed to check active lottery csv upload job")
	}
	if hasActiveJob {
		return adomain.LotteryCSVUploadResponse{}, errors.Wrap(cdomain.ErrBadRequest, "lottery csv upload job is already queued or running")
	}

	rows, err := parseLotteryCSVUploadRows(request.File)
	if err != nil {
		return adomain.LotteryCSVUploadResponse{}, err
	}

	validatedRows, err := l.validateLotteryCSVUploadRows(ctx, project, rows)
	if err != nil {
		return adomain.LotteryCSVUploadResponse{}, err
	}

	jobKSUID, err := ksuid.NewRandom()
	if err != nil {
		return adomain.LotteryCSVUploadResponse{}, errors.Wrap(err, "failed to generate lottery csv upload job id")
	}

	response := adomain.LotteryCSVUploadResponse{
		JobID:  jobKSUID.String(),
		Status: domain.LOTTERY_CSV_UPLOAD_JOB_STATUS_QUEUED,
	}
	requestedBy := request.JWTUser.Email
	if requestedBy == "" {
		requestedBy = request.JWTUser.Name
	}

	err = l.txManager.WithTx(ctx, func(txCtx context.Context) error {
		lockedProject, err := l.projectRepository.LockById(txCtx, request.ProjectID)
		if err != nil {
			return errors.Wrap(err, "failed to lock project")
		}
		if err := validateWinningProject(&lockedProject); err != nil {
			return err
		}

		hasActiveJobInTx, err := (domain.LotteryCSVUploadJob{}).HasActiveJobByProjectID(txCtx, request.ProjectID)
		if err != nil {
			return errors.Wrap(err, "failed to recheck active lottery csv upload job")
		}
		if hasActiveJobInTx {
			return errors.Wrap(cdomain.ErrBadRequest, "lottery csv upload job is already queued or running")
		}

		job := &domain.LotteryCSVUploadJob{
			JobID:           response.JobID,
			ProjectID:       request.ProjectID,
			RequestedBy:     requestedBy,
			FileName:        request.File.Filename,
			Status:          domain.LOTTERY_CSV_UPLOAD_JOB_STATUS_QUEUED,
			TotalRowCount:   len(validatedRows),
			SuccessRowCount: 0,
			FailedRowCount:  0,
		}
		_, err = job.Create(txCtx)
		if err != nil {
			return errors.Wrap(err, "failed to create lottery csv upload job")
		}

		for _, row := range validatedRows {
			if _, err := (&domain.LotteryCSVUploadJobDetail{
				JobID:                 job.Id,
				RowNumber:             row.rowNumber,
				LotteryApplicationID:  row.lotteryApplicationID,
				ExpectedLotteryStatus: &row.expectedStatus,
				ExpectedAppliedAt:     &row.expectedAppliedAt,
				InputAppliedQty:       &row.inputAppliedQty,
				InputWinningQty:       row.inputWinningQty,
				ProcessType:           row.processType,
				Status:                row.status,
			}).Create(txCtx); err != nil {
				return errors.Wrap(err, "failed to create lottery csv upload job detail")
			}
		}

		// バッチ予約は別実装とし、受付時点では QUEUED 登録まで行う。
		return nil
	})
	if err != nil {
		return adomain.LotteryCSVUploadResponse{}, err
	}

	return response, nil
}

func (l lotteryUsecase) Win(ctx context.Context, request adomain.LotteryWinRequest) error {
	// まずはトランザクション外で対象申込とファンドの基本条件を確認する。
	application, err := l.lotteryRepository.GetById(ctx, request.LotteryApplicationID)
	if err != nil {
		return errors.Wrap(err, "failed to get lottery application")
	}

	project, err := domain.Project{}.GetById(ctx, application.ProjectId)
	if err != nil {
		return errors.Wrap(err, "failed to get project")
	}
	if err := validateWinningProject(project); err != nil {
		return err
	}

	if err := validateWinningApplication(application, request.Qty); err != nil {
		return err
	}

	// 投資対象ユーザーが存在することだけ事前に確認しておく。
	user, err := domain.User{}.GetById(ctx, application.UserId)
	if err != nil {
		return errors.Wrap(err, "failed to get user")
	}
	if err := validateWinningUser(user); err != nil {
		return err
	}

	var createdShare *domain.Share
	var createdDepositRequest *domain.DepositRequest
	err = l.txManager.WithTx(ctx, func(txCtx context.Context) error {
		// 同一申込への並列当選処理を直列化するため、先に申込行をロックする。
		lockApplication := domain.LotteryApplication{}
		has, err := factory.DB(txCtx).
			SQL("SELECT lottery_application.* FROM lottery_application WHERE lottery_application.id = ? FOR UPDATE", request.LotteryApplicationID).
			Get(&lockApplication)
		if err != nil {
			return errors.Wrap(err, "failed to lock lottery application")
		}
		if !has {
			return errors.Wrap(cdomain.ErrItemNotFound, "lottery application not found")
		}

		// 残募集枠の確認に使う project 行も同一トランザクションでロックする。
		lockedProject := domain.Project{}
		has, err = factory.DB(txCtx).
			SQL("SELECT "+domain.PROJECT_DETAIL_COLUMN_LIST+" FROM project WHERE project.id = ? FOR UPDATE", lockApplication.ProjectId).
			Get(&lockedProject)
		if err != nil {
			return errors.Wrap(err, "failed to lock project")
		}
		if !has {
			return errors.Wrap(cdomain.ErrItemNotFound, "project not found")
		}

		// ロック取得後に必要な付随情報を含む形で再取得し、業務条件を再検証する。
		txApplication, err := l.lotteryRepository.GetById(txCtx, request.LotteryApplicationID)
		if err != nil {
			return errors.Wrap(err, "failed to get lottery application in tx")
		}
		if !txApplication.CanceledAt.IsZero() {
			return errors.Wrap(cdomain.ErrBadRequest, "lottery application is canceled")
		}

		txProject := &lockedProject
		if err := validateWinningProject(txProject); err != nil {
			return err
		}

		txUser, err := domain.User{}.GetById(txCtx, txApplication.UserId)
		if err != nil {
			return errors.Wrap(err, "failed to get user in tx")
		}
		if err := validateWinningUser(txUser); err != nil {
			return err
		}
		if txUser.UserType == domain.PRIVATE {
			hasMyNumber, err := domain.MyNumber{}.CheckMyNumberStatusByUserId(txCtx, txUser.Id)
			if err != nil {
				return errors.Wrap(err, "failed to check my number status")
			}
			if !hasMyNumber {
				return errors.Wrap(cdomain.ErrBadRequest, "my number is not registered")
			}
		}
		if err := txUser.CheckOverAge(txCtx); err != nil {
			return errors.Wrap(cdomain.ErrBadRequest, err.Error())
		}
		if _, err := (domain.UserVirtualAccount{}).GetVaByUserId(txCtx, txUser.Id, txProject.ProjectType); err != nil {
			return errors.Wrap(err, "failed to get linked virtual account")
		}

		// 既当選口数と今回当選口数の合計が申込口数を超えないことを最終確認する。
		txExistingWinningQty := winningQtyTotal(txApplication)
		if txExistingWinningQty+request.Qty > txApplication.Qty {
			return errors.Wrap(cdomain.ErrBadRequest, "winning qty exceeds lottery application qty")
		}

		// current_share_amount/current_share_qty 相当の上限チェックは Project.CheckShareQty に寄せる。
		amount := txProject.SharePrice * request.Qty
		if !txProject.CheckShareQty(request.Qty, amount) {
			return errors.Wrap(cdomain.ErrBadRequest, "insufficient remaining project qty or amount")
		}

		// 当選後の処理は既存の通常投資フローを流用し、share/deposit_request まで一括作成する。
		share, err := adminservice.InvestmentService{}.Invest(
			txCtx,
			txProject,
			txUser,
			request.Qty,
			domain.LOTTERY_MANUAL,
			true, // isWaitInvest --- 抽選当選はキャンセル待ち繰上と同じ入金期限ロジックを適用するため、trueを渡す
			false,
			false, // isWaitInvestFinalPortion --- 抽選は繰上げの分割対象ではないため常にfalse(MIR-1068)
		)
		if err != nil {
			return errors.Wrap(err, "failed to invest by lottery winning")
		}
		createdShare = share

		depositRequest, err := domain.DepositRequest{}.GetDepositRequestByShareId(txCtx, share.Id)
		if err != nil {
			return errors.Wrap(err, "failed to get deposit request in tx")
		}
		createdDepositRequest = depositRequest

		// 抽選申込と生成した share を中間テーブルで紐付ける。
		if _, err := (&domain.LotteryWinning{
			LotteryApplicationId: txApplication.Id,
			ShareId:              share.Id,
		}).Create(txCtx); err != nil {
			return errors.Wrap(err, "failed to create lottery winning")
		}

		// winning_at は初回当選日時として扱い、未設定時のみ更新する。
		if txApplication.WinningAt.IsZero() {
			updated, err := l.lotteryRepository.MarkAsWinning(txCtx, txApplication.Id)
			if err != nil {
				return errors.Wrap(err, "failed to mark lottery application as winning")
			}
			if !updated {
				return errors.Wrap(cdomain.ErrBadRequest, "failed to update lottery application winning status")
			}
		}

		// 手動当選の監査ログはDB更新と同じトランザクションで残す。
		if _, err := (&domain.OperatorHistory{
			UserId: request.JWTUser.Id,
			Name:   request.JWTUser.Name,
			Email:  request.JWTUser.Email,
			Menu:   "menu.manage_accounting.payment_list",
			Ip:     "",
			Title:  "lottery manual win application_id=" + txApplication.Id + " share_id=" + share.Id,
			State:  domain.CHANGE,
		}).Create(txCtx); err != nil {
			return errors.Wrap(err, "failed to create operator history")
		}

		if err := l.createLotteryHistory(
			txCtx,
			txApplication.Id,
			txApplication.Status,
			lotteryStatusAfterWinning(txApplication, request.Qty),
			request.JWTUser.Id,
		); err != nil {
			return errors.Wrap(err, "failed to create lottery history")
		}

		return nil
	})
	if err != nil {
		return err
	}

	// DB確定後は、トランザクション内で確定した share/deposit を使って通知だけ行う。
	if createdShare == nil {
		return errors.Wrap(cdomain.ErrItemNotFound, "share not found after commit")
	}
	if createdDepositRequest == nil {
		return errors.Wrap(cdomain.ErrItemNotFound, "deposit request not found after commit")
	}

	project, err = domain.Project{}.GetById(ctx, application.ProjectId)
	if err != nil {
		factory.Logger(ctx).Errorf("failed to get project after commit for lottery winning mail: application_id=%s project_id=%d err=%v", application.Id, application.ProjectId, err)
		return nil
	}
	if project == nil {
		factory.Logger(ctx).Errorf("project not found after commit for lottery winning mail: application_id=%s project_id=%d", application.Id, application.ProjectId)
		return nil
	}
	user, err = domain.User{}.GetById(ctx, application.UserId)
	if err != nil {
		factory.Logger(ctx).Errorf("failed to get user after commit for lottery winning mail: application_id=%s user_id=%d err=%v", application.Id, application.UserId, err)
		return nil
	}
	if user == nil {
		factory.Logger(ctx).Errorf("user not found after commit for lottery winning mail: application_id=%s user_id=%d", application.Id, application.UserId)
		return nil
	}

	switch createdShare.InvestStatus {
	case domain.SUCCESS_INVEST:
		err := l.mailService.SendMail(ctx, cdomain.SendMailRequest{
			TemplateType: domain.LOTTERY_WINNING_IMMEDIATE_SUCCESS_TEMPLATE,
			Sender:       l.cfg.ServiceConfig.Mail.Info,
			Receivers:    []string{user.Email},
			MailType:     domain.MAIL_SEND_PRIORITY_NORMAL,
			Details: map[string]interface{}{
				"Receiver":          user.Name,
				"ProjectName":       project.Name,
				"WinningAmount":     message.NewPrinter(language.Japanese).Sprintf("%d", createdShare.Amount),
				"ApplicationAmount": message.NewPrinter(language.Japanese).Sprintf("%d", application.Amount),
				"ServiceUrl":        l.cfg.ServiceConfig.Url.UserServiceDomain,
			},
		})
		if err != nil {
			factory.Logger(ctx).Errorf("failed to send lottery winning immediate success mail: %v", err)
		}
	case domain.REQUEST_INVEST:
		err := l.mailService.SendMail(ctx, cdomain.SendMailRequest{
			TemplateType: domain.LOTTERY_WINNING_INSUFFICIENT_BALANCE_TEMPLATE,
			Sender:       l.cfg.ServiceConfig.Mail.Info,
			Receivers:    []string{user.Email},
			MailType:     domain.MAIL_SEND_PRIORITY_NORMAL,
			Details: map[string]interface{}{
				"Receiver":          user.Name,
				"ProjectName":       project.Name,
				"WinningAmount":     message.NewPrinter(language.Japanese).Sprintf("%d", createdShare.Amount),
				"ApplicationAmount": message.NewPrinter(language.Japanese).Sprintf("%d", application.Amount),
				"LackAmount":        message.NewPrinter(language.Japanese).Sprintf("%d", createdDepositRequest.NeedAmount()),
				"PaymentEnd":        jsonutils.JSONTime.ToString(createdDepositRequest.ExpireDepositDate),
				"ServiceUrl":        l.cfg.ServiceConfig.Url.UserServiceDomain,
			},
		})
		if err != nil {
			factory.Logger(ctx).Errorf("failed to send lottery winning insufficient balance mail: %v", err)
		}
	}

	return nil
}

func (l lotteryUsecase) Lose(ctx context.Context, request adomain.LotteryLoseRequest) error {
	// まずはトランザクション外で対象申込とファンドの基本条件を確認する。
	application, err := l.lotteryRepository.GetById(ctx, request.LotteryApplicationID)
	if err != nil {
		return errors.Wrap(err, "failed to get lottery application")
	}

	project, err := domain.Project{}.GetById(ctx, application.ProjectId)
	if err != nil {
		return errors.Wrap(err, "failed to get project")
	}
	if err := validateWinningProject(project); err != nil {
		return err
	}
	if err := validateLosingApplication(application); err != nil {
		return err
	}

	var shouldSendLostMail bool
	err = l.txManager.WithTx(ctx, func(txCtx context.Context) error {
		// 同一申込への当選/落選操作を直列化するため、先に申込行をロックする。
		lockApplication := domain.LotteryApplication{}
		has, err := factory.DB(txCtx).
			SQL("SELECT lottery_application.* FROM lottery_application WHERE lottery_application.id = ? FOR UPDATE", request.LotteryApplicationID).
			Get(&lockApplication)
		if err != nil {
			return errors.Wrap(err, "failed to lock lottery application")
		}
		if !has {
			return errors.Wrap(cdomain.ErrItemNotFound, "lottery application not found")
		}

		// ファンド条件も同一トランザクションで再検証する。
		lockedProject := domain.Project{}
		has, err = factory.DB(txCtx).
			SQL("SELECT "+domain.PROJECT_DETAIL_COLUMN_LIST+" FROM project WHERE project.id = ? FOR UPDATE", lockApplication.ProjectId).
			Get(&lockedProject)
		if err != nil {
			return errors.Wrap(err, "failed to lock project")
		}
		if !has {
			return errors.Wrap(cdomain.ErrItemNotFound, "project not found")
		}
		if err := validateWinningProject(&lockedProject); err != nil {
			return err
		}

		// ロック取得後に当選情報込みで再取得し、未当選残口数を最終確認する。
		txApplication, err := l.lotteryRepository.GetById(txCtx, request.LotteryApplicationID)
		if err != nil {
			return errors.Wrap(err, "failed to get lottery application in tx")
		}
		if err := validateLosingApplication(txApplication); err != nil {
			return err
		}

		// 初回落選のみ lost_at を設定し、通知もそのタイミングに限定する。
		shouldSendLostMail = txApplication.LostAt.IsZero()
		if shouldSendLostMail {
			// 部分当選済みでも、残口数があれば lost_at を立てて残分の落選確定を表現する。
			updated, err := l.lotteryRepository.MarkAsLost(txCtx, txApplication.Id)
			if err != nil {
				return errors.Wrap(err, "failed to mark lottery application as lost")
			}
			if !updated {
				return errors.Wrap(cdomain.ErrBadRequest, "failed to update lottery application lost status")
			}
		}

		// 手動落選の監査ログはDB更新と同じトランザクションで残す。
		if _, err := (&domain.OperatorHistory{
			UserId: request.JWTUser.Id,
			Name:   request.JWTUser.Name,
			Email:  request.JWTUser.Email,
			Menu:   "menu.manage_accounting.payment_list",
			Ip:     "",
			Title:  "lottery manual lose application_id=" + txApplication.Id,
			State:  domain.CHANGE,
		}).Create(txCtx); err != nil {
			return errors.Wrap(err, "failed to create operator history")
		}

		if err := l.createLotteryHistory(
			txCtx,
			txApplication.Id,
			txApplication.Status,
			lotteryStatusAfterLosing(txApplication),
			request.JWTUser.Id,
		); err != nil {
			return errors.Wrap(err, "failed to create lottery history")
		}

		return nil
	})
	if err != nil {
		return err
	}

	// 落選通知は初回落選時のみ、コミット後に送信する。
	if shouldSendLostMail {
		user, err := domain.User{}.GetById(ctx, application.UserId)
		if err != nil {
			factory.Logger(ctx).Errorf("failed to get user after commit for lottery lost mail: application_id=%s user_id=%d err=%v", application.Id, application.UserId, err)
			return nil
		}
		if user == nil {
			factory.Logger(ctx).Errorf("user not found after commit for lottery lost mail: application_id=%s user_id=%d", application.Id, application.UserId)
			return nil
		}

		err = l.mailService.SendMail(ctx, cdomain.SendMailRequest{
			TemplateType: domain.LOTTERY_LOST_TEMPLATE,
			Sender:       l.cfg.ServiceConfig.Mail.Info,
			Receivers:    []string{user.Email},
			MailType:     domain.MAIL_SEND_PRIORITY_NORMAL,
			Details: map[string]interface{}{
				"Receiver":          user.Name,
				"ProjectName":       project.Name,
				"ApplicationAmount": message.NewPrinter(language.Japanese).Sprintf("%d", application.Amount),
				"ServiceUrl":        l.cfg.ServiceConfig.Url.UserServiceDomain,
			},
		})
		if err != nil {
			factory.Logger(ctx).Errorf("failed to send lottery lost mail: %v", err)
		}
	}

	return nil
}

func (l lotteryUsecase) Cancel(ctx context.Context, request adomain.LotteryCancelRequest) (adomain.LotteryCancelResponse, error) {
	actor, ok := actorcontext.ActorFrom(ctx)
	if !ok {
		return adomain.LotteryCancelResponse{}, errors.Wrap(cdomain.ErrBadRequest, "actor is required")
	}

	err := l.txManager.WithTx(ctx, func(txCtx context.Context) error {
		lockedApplication, err := l.lotteryRepository.LockById(txCtx, request.LotteryApplicationID)
		if err != nil {
			return errors.Wrap(err, "failed to lock lottery application")
		}
		lockedApplication.ComputeStatus()

		lockedProject, err := l.projectRepository.LockById(txCtx, lockedApplication.ProjectId)
		if err != nil {
			return errors.Wrap(err, "failed to lock project")
		}
		if !canCancelLotteryApplication(lockedApplication, lockedProject.ProjectStatus) {
			return errors.Wrap(cdomain.ErrBadRequest, "lottery application cannot be canceled")
		}

		success, err := l.lotteryRepository.MarkAsCanceled(txCtx, request.LotteryApplicationID)
		if err != nil {
			return errors.Wrap(err, "failed to cancel lottery application")
		}
		if !success {
			return errors.Wrap(cdomain.ErrBadRequest, "lottery application was not canceled")
		}

		if lockedProject.ProjectStatus == domain.OPEN_FUNDING {
			if err := l.projectRepository.ReduceFinalFundingAmount(txCtx, lockedProject.Id, lockedApplication.Amount); err != nil {
				return errors.Wrap(err, "failed to reduce final funding amount")
			}
		}

		if err := l.operatorHistoryRepository.Create(txCtx, domain.OperatorHistory{
			UserId: request.JWTUser.Id,
			Name:   request.JWTUser.Name,
			Email:  request.JWTUser.Email,
			Menu:   "menu.manage_accounting.project_lottery.list",
			Ip:     "",
			Title:  "lottery cancel application_id=" + request.LotteryApplicationID,
			State:  domain.CHANGE,
		}); err != nil {
			return errors.Wrap(err, "failed to create operator history")
		}

		if err := l.lotteryHistoryRepository.Create(txCtx, domain.LotteryHistory{
			ApplicationId: request.LotteryApplicationID,
			PrevStatus:    lockedApplication.Status,
			NewStatus:     domain.LOTTERY_STATUS_CANCELED,
			ActorId:       actor.Id,
		}); err != nil {
			return errors.Wrap(err, "failed to create lottery history")
		}

		return nil
	})
	if err != nil {
		return adomain.LotteryCancelResponse{}, err
	}

	application, err := l.lotteryRepository.GetById(ctx, request.LotteryApplicationID)
	if err != nil {
		factory.Logger(ctx).Errorf("failed to get canceled lottery application for mail: application_id=%s err=%v", request.LotteryApplicationID, err)
		return adomain.LotteryCancelResponse{}, nil
	}
	user, err := l.userRepository.GetById(ctx, application.UserId)
	if err != nil {
		factory.Logger(ctx).Errorf("failed to get user for lottery cancel mail: application_id=%s user_id=%d err=%v", request.LotteryApplicationID, application.UserId, err)
		return adomain.LotteryCancelResponse{}, nil
	}
	project, err := l.projectRepository.GetById(ctx, application.ProjectId)
	if err != nil {
		factory.Logger(ctx).Errorf("failed to get project for lottery cancel mail: application_id=%s project_id=%d err=%v", request.LotteryApplicationID, application.ProjectId, err)
		return adomain.LotteryCancelResponse{}, nil
	}

	err = l.mailService.SendMail(ctx, cdomain.SendMailRequest{
		TemplateType: domain.LOTTERY_APPLICATION_CANCEL_TEMPLATE,
		Sender:       l.cfg.ServiceConfig.Mail.Info,
		Receivers:    []string{user.Email},
		MailType:     domain.MAIL_SEND_PRIORITY_NORMAL,
		Details: map[string]interface{}{
			"ProjectName":       project.Name,
			"Receiver":          user.Name,
			"ApplicationAmount": message.NewPrinter(language.Japanese).Sprintf("%d", application.Amount),
			"CancelTime":        jsonutils.JSONTime.ToString(application.CanceledAt),
			"ServiceUrl":        l.cfg.ServiceConfig.Url.UserServiceDomain,
		},
	})
	if err != nil {
		factory.Logger(ctx).Errorf("failed to send lottery cancel mail: %v", err)
	}

	return adomain.LotteryCancelResponse{}, nil
}

func (l lotteryUsecase) Download(ctx context.Context, request adomain.LotteryDownloadRequest) (adomain.LotteryDownloadResponse, error) {
	opts := cdomain.LotteryApplicationQueryOptionFunc()
	totalCount, lotteries, err := l.lotteryRepository.GetAll(ctx,
		opts.WithSortBy(request.Sortby),
		opts.WithOrder(request.Order),
		opts.WithOffset(0),
		opts.WithLimit(lotteryListDownloadMaxCount),
		opts.WithUserId(lo.FromPtr(request.UserId)),
		opts.WithUserName(lo.FromPtr(request.UserName)),
		opts.WithProjectId(lo.FromPtr(request.ProjectId)),
		opts.WithProjectName(lo.FromPtr(request.ProjectName)),
		opts.WithAmountMin(lo.FromPtr(request.AmountMin)),
		opts.WithAmountMax(lo.FromPtr(request.AmountMax)),
		opts.WithStatus(lo.FromPtr(request.Status)),
	)
	if err != nil {
		return adomain.LotteryDownloadResponse{}, errors.Wrap(err, "failed to get lottery applications for download")
	}
	if totalCount > lotteryListDownloadMaxCount {
		return adomain.LotteryDownloadResponse{}, errors.Wrap(cdomain.ErrRequestLimitExceeded, "lottery list CSV download result exceeds max count")
	}

	buf := new(bytes.Buffer)
	writer := csv.NewWriter(buf)

	csvData := l.lotteryService.ExportCSV(lotteries)
	if err = writer.WriteAll(csvData); err != nil {
		return adomain.LotteryDownloadResponse{}, errors.Wrap(err, "failed to write csv data")
	}

	if err = l.operatorHistoryRepository.Create(ctx, domain.OperatorHistory{
		UserId: request.JWTUser.Id,
		Name:   request.JWTUser.Name,
		Email:  request.JWTUser.Email,
		Menu:   "menu.manage_accounting.project_lottery.list",
		Ip:     request.ClientIp,
		State:  domain.DOWNLOAD,
	}); err != nil {
		factory.Logger(ctx).Errorf("failed to create operator history for lottery list CSV download: %v", err)
	}

	return adomain.LotteryDownloadResponse{Data: buf.Bytes()}, nil
}

func (l lotteryUsecase) DownloadForBulkLottery(ctx context.Context, request adomain.LotteryDownloadForBulkLotteryRequest) (adomain.LotteryDownloadResponse, error) {
	opts := cdomain.LotteryApplicationQueryOptionFunc()
	_, lotteries, err := l.lotteryRepository.GetAll(ctx,
		opts.WithProjectId(request.ProjectId),
		opts.WithSortBy(request.Sortby),
		opts.WithOrder(request.Order),
	)
	if err != nil {
		return adomain.LotteryDownloadResponse{}, errors.Wrap(err, "failed to get lottery applications for bulk lottery download")
	}

	eligibleLotteries := make([]domain.LotteryApplication, 0, len(lotteries))
	for _, lottery := range lotteries {
		if hasChanceToWin(lottery.Status) {
			eligibleLotteries = append(eligibleLotteries, lottery)
		}
	}

	buf := new(bytes.Buffer)
	writer := csv.NewWriter(buf)

	csvData := l.lotteryService.ExportCSVForBulkLottery(eligibleLotteries)
	if err = writer.WriteAll(csvData); err != nil {
		return adomain.LotteryDownloadResponse{}, errors.Wrap(err, "failed to write bulk lottery csv data")
	}

	if err = l.operatorHistoryRepository.Create(ctx, domain.OperatorHistory{
		UserId: request.JWTUser.Id,
		Name:   request.JWTUser.Name,
		Email:  request.JWTUser.Email,
		Menu:   "menu.manage_accounting.project_lottery.list",
		Ip:     request.ClientIp,
		State:  domain.DOWNLOAD,
	}); err != nil {
		factory.Logger(ctx).Errorf("failed to create operator history for bulk lottery CSV download: %v", err)
	}

	return adomain.LotteryDownloadResponse{Data: buf.Bytes()}, nil
}
