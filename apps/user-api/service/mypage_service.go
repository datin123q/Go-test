package service

import (
	"context"

	"github.com/samber/lo"

	"github.com/illuminarean/energy-shares-backend-repo/service"

	"github.com/illuminarean/energy-shares-backend-repo/domain"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/factory"
	"github.com/illuminarean/energyshares-user-api/service/dto"
)

type MyPageService struct {
	withdrawalService service.IWithdrawalService
}

func NewMyPageService() *MyPageService {
	return &MyPageService{
		withdrawalService: service.NewWithdrawalService(),
	}
}

func (s *MyPageService) GetMyPageById(ctx context.Context, id int64) (*dto.MyPage, error) {
	var err error

	q := factory.DB(ctx)

	var myPage dto.MyPage
	if has, err := q.Table("user").Select("*").Where("user.id = ?", id).Get(&myPage); err != nil {
		return nil, err
	} else if !has {
		return nil, nil
	}

	switch myPage.UserType {
	case domain.PRIVATE:
		pu, err := domain.PrivateUser{}.GetById(ctx, id)
		if err != nil {
			return nil, err
		}
		myPage.User = pu.User
		myPage.UserId = pu.User.Id
		myPage.AuditStatus = pu.AuditStatus
		myPage.Name = pu.FamilyName + pu.FirstName
		myPage.Birthday = pu.Birthday

		myNumber, err := domain.MyNumber{}.GetByUserId(ctx, id)
		if err != nil {
			return nil, err
		}
		myPage.MyNumber = myNumber

	case domain.CORPORATE:
		corp, err := domain.CorporateUser{}.GetById(ctx, id)
		if err != nil {
			return nil, err
		}
		myPage.User = corp.User
		myPage.AuditStatus = corp.AuditStatus
		myPage.MyNumber = corp.MyNumber
		myPage.CorporateProfile = corp.CorporateProfile
		myPage.CertifiedCopyFiles, err = domain.UserFile{}.FindByUserIdAndFileType(ctx, 10, myPage.UserId, domain.CERTIFIED_COPY)
	}

	err = getCorporateUserIdentificationFiles(ctx, &myPage)
	if err != nil {
		return nil, err
	}

	amount, err := getUserShareAmountInfo(ctx, id)
	if err != nil {
		return nil, err
	}
	myPage.UserShareAmountInfo = *amount

	account, err := domain.UserAccount{}.GetByUserId(ctx, id)
	if err != nil {
		return nil, err
	}
	myPage.Balance = lo.FromPtr(account.Balance)

	// 出金リクエスト金額（= 未完了+締切+送金失敗）
	withdrawalTotal, err := s.withdrawalService.GetTotalWithdrawalAmountByUserId(ctx, id)
	if err != nil {
		return nil, err
	}
	myPage.Withdrawal = withdrawalTotal

	// 投資入金進行中の金額、入金が必要な金額を照会
	noDepositRequest, err := domain.DepositRequest{}.GetNoDepositRequestByUserId(ctx, id)
	if err != nil {
		return nil, err
	}
	var totalUnpaidAmount int64
	for _, ndr := range noDepositRequest {
		totalUnpaidAmount += ndr.RequestAmount - ndr.DepositAmount
	}
	myPage.UserTotalUnpaid = totalUnpaidAmount

	// 総資産 = 運用中資産 + 投資完了金 + 投資入金中の金額 + 預け金 + 出金リクエスト金額
	myPage.UserTotalAsset = myPage.UserShareAmountInfo.InvestTotalAmount + lo.FromPtr(account.Balance) + myPage.Withdrawal

	return &myPage, nil
}

func getCorporateUserIdentificationFiles(ctx context.Context, page *dto.MyPage) (err error) {
	page.Identification1FrontFile, err = domain.UserFile{}.FindOneUserIdByAndFileType(ctx, page.UserId, domain.IDENTIFICATION1_FRONT)
	if err != nil {
		return err
	}
	page.Identification1BackFile, err = domain.UserFile{}.FindOneUserIdByAndFileType(ctx, page.UserId, domain.IDENTIFICATION1_BACK)
	if err != nil {
		return err
	}

	page.Identification2FrontFile, err = domain.UserFile{}.FindOneUserIdByAndFileType(ctx, page.UserId, domain.IDENTIFICATION2_FRONT)
	if err != nil {
		return err
	}
	page.Identification2BackFile, err = domain.UserFile{}.FindOneUserIdByAndFileType(ctx, page.UserId, domain.IDENTIFICATION2_BACK)
	if err != nil {
		return err
	}

	return err
}

func getUserShareAmountInfo(ctx context.Context, id int64) (*domain.UserShareAmountInfo, error) {
	amountInfo, err := domain.UserShareAmountInfo{}.GetUserProjectShareAmount(ctx, id)
	if err != nil {
		return nil, err
	}

	requestInvest, err := domain.UserShareAmountInfo{}.SumDepositAmountOfRequestInvest(ctx, id)
	if err != nil {
		return nil, err
	}

	amountInfo.InvestRequestAmount = requestInvest
	// NOTE: 「投資完了金額」は「投資申請中金額」の中に含まれている
	amountInfo.InvestTotalAmount = amountInfo.InvestOperationAmount + amountInfo.InvestRequestAmount
	amountInfo.InvestTotalCount = amountInfo.InvestOperationCount + amountInfo.InvestSuccessCount + amountInfo.InvestRequestCount

	return amountInfo, nil
}
