package usecase

import (
	"context"
	udomain "energyshares-invest-worker/domain"
	"energyshares-invest-worker/service"
	"fmt"
	"strconv"
	"time"

	"github.com/illuminarean/energy-shares-backend-repo/domain"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/factory"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/jsonutils"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/serviceutils"
	"github.com/sirupsen/logrus"
)

type investUsecase struct {
	txManager factory.TxManager
}

func NewInvestUsecase(txManager factory.TxManager) udomain.InvestUsecase {
	return &investUsecase{
		txManager: txManager,
	}
}

// ProcessInvestmentRequest 投資リクエストを処理する
func (u investUsecase) ProcessInvestmentRequest(ctx context.Context, request udomain.InvestRequest,
) error {
	return u.txManager.WithTx(ctx, func(txCtx context.Context) error {
		// プロジェクト情報取得
		p, err := domain.Project{}.GetById(txCtx, request.ProjectId)
		if err != nil {
			return factory.ErrorDB.New(err)
		} else if p == nil {
			msg := fmt.Sprintf("project is nil. projectId:%d", request.ProjectId)
			logrus.Warn(msg)
			return nil
		}

		// ユーザー情報取得
		user, err := domain.User{}.GetById(txCtx, request.UserId)
		if err != nil {
			return factory.ErrorDB.New(err)
		} else if user == nil {
			msg := fmt.Sprintf("user is nil. userId:%d", request.UserId)
			logrus.Warn(msg)
			return nil
		} else if user.BlockInvest {
			// 取引制限ユーザー
			msg := fmt.Sprintf("user is blocked to invest. userId:%d", request.UserId)
			logrus.Info(msg)
			return nil
		}

		// 当選チェック
		amount := p.SharePrice * request.Qty
		if p.IsWaitInvest || !p.CheckShareQty(request.Qty, amount) {
			// キャンセル待ちモードへ移行済みもしくは、募集上限に到達している場合はキャンセル待ち登録処理を行う
			// TODO: キャンセル待ちのロジックを共通化できるようにしたい
			if user.UserType == domain.PRIVATE {
				m, err := domain.MyNumber{}.CheckMyNumberStatusByUserId(txCtx, user.Id)
				if err != nil {
					return factory.ErrorDB.New(err)
				}
				if !m {
					return factory.ErrorMyNumberNotRegistered.New(nil)
				}
			}

			amount := p.SharePrice * request.Qty

			investmentWait := domain.InvestmentWaiters{
				ProjectId:   p.Id,
				UserId:      user.Id,
				Qty:         request.Qty,
				Amount:      amount,
				WaitingDate: jsonutils.JSONTime(time.Now()),
				Status:      domain.WAITING,
			}
			if _, err := investmentWait.Create(txCtx); err != nil {
				return factory.ErrorDB.New(err)
			}

			var mail domain.Mail
			if err := mail.InvestStandbyMail(txCtx, serviceutils.GetServiceInfoMail(), user.Email, user.Name, p.Name, jsonutils.JSONTime.ToString(investmentWait.WaitingDate), strconv.Itoa(int(request.Qty)), strconv.Itoa(int(amount))); err != nil {
				return factory.ErrorAWSSendMail.New(err)
			}

			factory.Logger(ctx).WithFields(logrus.Fields{
				"project_id": p.Id,
				"user_id":    user.Id,
				"qty":        request.Qty,
				"amount":     amount,
			}).Info("Success Add to Invest Waitlist")

			return nil
		}

		// 当選処理
		err = service.InvestService{}.Invest(txCtx, p, user, request.Qty, amount, domain.NOT_IMMEDIATE)
		if err != nil {
			return err
		}

		return nil
	})
}
