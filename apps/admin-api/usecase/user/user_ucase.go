package user

import (
	"bytes"
	"context"
	"encoding/csv"
	"strconv"
	"time"

	"github.com/pkg/errors"
	"github.com/samber/lo"

	"github.com/illuminarean/energy-shares-backend-repo/domain"

	adomain "github.com/illuminarean/energy-share-core-api/domain"
	cdomain "github.com/illuminarean/energy-shares-backend-repo/clean-arch-domain"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/emailutils"
)

type userUsecase struct {
	emailpkg                emailutils.EmailRequester
	userRepository          cdomain.UserRepository
	compositeUserRepository cdomain.CompositeUserRepository
}

func New(
	emailpkg emailutils.EmailRequester,
	userRepository cdomain.UserRepository,
	compositeUserRepository cdomain.CompositeUserRepository,
) adomain.UserUsecase {
	return &userUsecase{
		emailpkg:                emailpkg,
		userRepository:          userRepository,
		compositeUserRepository: compositeUserRepository,
	}
}

func (u userUsecase) GetSentMails(ctx context.Context, request adomain.UserGetSentMailsRequest) (adomain.MailGetAllResponse, error) {

	user, err := u.userRepository.GetById(ctx, request.UserId)
	if err != nil {
		return adomain.MailGetAllResponse{}, err
	}

	// 변화가 있을 수 있는 부분이라 리펙터링은 추후에 한다.
	// queryParam 작성
	queryParam := emailutils.GetAllQuery{
		ServiceName: "ESJP",
		ServiceEnv:  request.Env,
		Receiver:    lo.ToPtr(user.Email),
	}

	// sort
	if request.Sortby != "" && request.Order != "" {
		// 통합 메일 서비스의 쿼리로 변경
		if request.Sortby == "send_date" {
			request.Sortby = "sendDateTime"
		} else if request.Sortby == "created" {
			request.Sortby = "createdDate"
		}
		queryParam.Sort = lo.ToPtr(request.Sortby + "," + request.Order)
	}

	// pagination - page는 0부터 시작한다.
	if request.SkipCount >= 0 && request.MaxResultCount > 0 { // 이미 검증은 되었다 판단하지만 혹시 모르니 추가 검증
		queryParam.Page = request.SkipCount / request.MaxResultCount
		queryParam.Size = request.MaxResultCount
	}

	response, err := u.emailpkg.GetAllRequest(ctx, queryParam)
	if err != nil {
		return adomain.MailGetAllResponse{}, err
	}

	var mailGetAllResponse adomain.MailGetAllResponse
	if err := mailGetAllResponse.FromEmailListResponse(response); err != nil {
		return mailGetAllResponse, err
	}
	return mailGetAllResponse, nil
}

func (u userUsecase) DownloadUserByFilter(ctx context.Context, request adomain.UserDownloadUserByFilterRequest) (adomain.UserDownloadUserByFilterResponse, error) {

	// 연령대에 따른 생년월일 계산
	today := time.Now()
	var minBirthday, maxBirthday *string
	if request.Age != nil && request.Age.Min != nil {
		minBirthday = lo.ToPtr(time.Date(today.Year()-int(*request.Age.Max), today.Month(), today.Day(), 0, 0, 0, 0, time.Local).String())
	}
	if request.Age != nil && request.Age.Max != nil {
		maxBirthday = lo.ToPtr(time.Date(today.Year()-int(*request.Age.Min), today.Month(), today.Day(), 23, 59, 59, 999999999, time.Local).String())
	}

	// 날짜 범위 추출
	minSignUpRequestDate, maxSignUpRequestDate := formatDateRange(request.SignUpRequestDate)
	minSignUpApprovalDate, maxSignUpApprovalDate := formatDateRange(request.SignUpApprovalDate)
	minLastLoginDate, maxLastLoginDate := formatDateRange(request.LastLoginDate)

	// 숫자 범위 추출
	minTotalAsset, maxTotalAsset := formatNumberRange(request.TotalAsset)
	minInvestOperationAmount, maxInvestOperationAmount := formatNumberRange(request.InvestOperationAmount)
	minDeposit, maxDeposit := formatNumberRange(request.Deposit)
	minInvestSuccessAmount, maxInvestSuccessAmount := formatNumberRange(request.InvestSuccessAmount)
	minInvestRequestAmount, maxInvestRequestAmount := formatNumberRange(request.InvestRequestAmount)
	minUnpaidAmount, maxUnpaidAmount := formatNumberRange(request.UnpaidAmount)

	opts := cdomain.CompositeUserQueryOptionFunc()

	_, compositeUsers, err := u.compositeUserRepository.GetAll(ctx,
		opts.WithEmails(request.Emails...),
		opts.WithAuditStatuses(request.AuditStatuses...),
		opts.WithUserTypes(request.UserTypes...),
		opts.WithNationalities(request.Nationalities...),
		opts.WithGenders(request.Genders...),
		opts.WithJobs(request.Jobs...),
		opts.WithAssets(request.Assets...),
		opts.WithAnnualIncomes(request.AnnualIncomes...),
		opts.WithAllowsNewsletter(request.AllowsNewsletter),
		opts.WithHasMyNumber(request.HasMyNumber),
		opts.WithMinSignUpRequestDate(minSignUpRequestDate),
		opts.WithMaxSignUpRequestDate(maxSignUpRequestDate),
		opts.WithMinSignUpApprovalDate(minSignUpApprovalDate),
		opts.WithMaxSignUpApprovalDate(maxSignUpApprovalDate),
		opts.WithMinLastLoginDate(minLastLoginDate),
		opts.WithMaxLastLoginDate(maxLastLoginDate),
		opts.WithMinBirthday(minBirthday),
		opts.WithMaxBirthday(maxBirthday),
		opts.WithMinTotalAsset(minTotalAsset),
		opts.WithMaxTotalAsset(maxTotalAsset),
		opts.WithMinInvestOperationAmount(minInvestOperationAmount),
		opts.WithMaxInvestOperationAmount(maxInvestOperationAmount),
		opts.WithMinDeposit(minDeposit),
		opts.WithMaxDeposit(maxDeposit),
		opts.WithMinInvestSuccessAmount(minInvestSuccessAmount),
		opts.WithMaxInvestSuccessAmount(maxInvestSuccessAmount),
		opts.WithMinInvestRequestAmount(minInvestRequestAmount),
		opts.WithMaxInvestRequestAmount(maxInvestRequestAmount),
		opts.WithMinUnpaidAmount(minUnpaidAmount),
		opts.WithMaxUnpaidAmount(maxUnpaidAmount),
	)

	if len(request.AuditStatuses) == 0 { // 심사상태 필터링이 없으면 무효 상태는 기본적으로 제외한다
		opts.WithNoAuditStatus(domain.INVALIDITY_AUDIT)
	}

	// 데이터를 엑셀 파일로 변환
	csvData := exportCSV(compositeUsers)
	var buf bytes.Buffer
	writer := csv.NewWriter(&buf)

	if err = writer.WriteAll(csvData); err != nil {
		return adomain.UserDownloadUserByFilterResponse{}, errors.Wrap(err, "failed to write csv data")
	}

	return adomain.UserDownloadUserByFilterResponse{Data: buf.Bytes()}, nil
}

func formatDateRange(dateRange *adomain.DateRange) (minDate, maxDate *string) {
	if dateRange != nil {
		if !dateRange.StartDate.IsEmpty() {
			minDate = lo.ToPtr(dateRange.StartDate.ToFormat(time.DateTime))
		}
		if !dateRange.EndDate.IsEmpty() {
			maxDate = lo.ToPtr(dateRange.EndDate.ToFormat(time.DateTime))
		}
	}
	return minDate, maxDate
}

func formatNumberRange(numberRange *adomain.NumberRange) (min, max *int64) {
	if numberRange != nil {
		min = numberRange.Min
		max = numberRange.Max
	}
	return min, max
}

func exportCSV(items []cdomain.CompositeUser) cdomain.CSVData {

	records := [][]string{cdomain.CompositeUserCSVHeader}

	for _, item := range items {
		record := []string{
			item.Name,
			item.Email,
			lo.If(item.AuditStatus == nil, "").ElseF(func() string {
				return (*item.AuditStatus).DescriptionJapan()
			}),
			string(lo.FromPtr(item.Nationality)),
			string(lo.FromPtr(item.Gender)),
			lo.If(item.Birthday.IsEmpty(), "").ElseF(func() string {
				return item.Birthday.ToFormat(time.DateOnly)
			}),
			lo.If(item.Job == nil, "").ElseF(func() string {
				return (*item.Job).DescriptionJapan()
			}),
			lo.If(item.SignUpRequestDate.IsEmpty(), "").ElseF(func() string {
				return item.SignUpRequestDate.ToFormat(time.DateTime)
			}),
			lo.If(item.SignUpApprovalDate.IsEmpty(), "").ElseF(func() string {
				return item.SignUpApprovalDate.ToFormat(time.DateTime)
			}),
			lo.If(lo.FromPtr(item.AllowsNewsletter), "同意").Else("非同意"),
			lo.If(item.LastLoginDate.IsEmpty(), "").ElseF(func() string {
				return item.LastLoginDate.ToFormat(time.DateTime)
			}),
			lo.If(item.MyNumberStatus != nil && *item.MyNumberStatus == domain.APPROVE, "承認").Else("非承認"),
			lo.If(item.Asset == nil, "").ElseF(func() string {
				return (*item.Asset).DescriptionJapan()
			}),
			lo.If(item.AnnualIncome == nil, "").ElseF(func() string {
				return (*item.AnnualIncome).DescriptionJapan()
			}),
			strconv.FormatInt(lo.FromPtr(item.TotalAsset), 10),
			strconv.FormatInt(lo.FromPtr(item.Deposit), 10),
			strconv.FormatInt(lo.FromPtr(item.InvestOperationAmount), 10),
			strconv.FormatInt(lo.FromPtr(item.InvestSuccessAmount), 10),
			strconv.FormatInt(lo.FromPtr(item.InvestRequestAmount), 10),
			strconv.FormatInt(lo.FromPtr(item.UnpaidAmount), 10),
			string(lo.FromPtr(item.UserType)),
		}
		records = append(records, record)
	}

	return records
}
