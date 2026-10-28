package controllers

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/pangpanglabs/echoswagger/v2"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"

	"github.com/illuminarean/energy-shares-backend-repo/domain"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/apirender"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/factory"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/jwthelper"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/serviceutils"
	"github.com/illuminarean/energy-shares-backend-repo/service"
)

type MyNumberApiController struct {
	myNumberService service.IMyNumberService
}

func NewMyNumberApiController() *MyNumberApiController {
	return &MyNumberApiController{
		myNumberService: service.NewMyNumberService(),
	}
}

func (c MyNumberApiController) Init(g echoswagger.ApiGroup) {

	g.GET("", c.GetAll).
		AddParamQueryNested(MyNumberSearchInput{}).
		AddResponse(http.StatusOK, "マイナンバーリスト", domain.MyNumber{}, nil).
		SetDescription("マイナンバーリストを照会します。").
		SetSummary("マイナンバーリスト照会").
		SetSecurity("Authorization")

	g.PUT("/:id", c.UpdateStatus).
		AddParamPath("", "id", "").
		AddParamForm(MyNumberStatusInput{}, "body", "マイナンバーステータス変更", true).
		SetSummary("マイナンバーステータスを変更します。").
		SetDescription("マイナンバーステータス変更").
		SetSecurity("Authorization")

	g.GET("/:id/check-duplicates", c.CheckDuplicateMyNumber).
		AddParamPath("", "id", "").
		AddResponse(http.StatusOK, "マイナンバー重複チェック", DuplicatedMyNumberResponse{}, nil).
		SetSummary("マイナンバー重複チェック").
		SetDescription("APPROVE、WAIT_APPROVE状態の重複したマイナンバーがあるかをチェック").
		SetSecurity("Authorization")
}

func (api *MyNumberApiController) GetAll(c echo.Context) error {
	var s MyNumberSearchInput

	if err := c.Bind(&s); err != nil {
		return factory.ErrorParameter.New(err)
	}
	if s.MaxResultCount == 0 {
		s.MaxResultCount = DefaultMaxResultCount
	}
	option := s.ToModel()

	id, email, err := jwthelper.TokenToUser(c)
	if err != nil {
		return err
	}

	if s.MyNumber != nil && len(strings.TrimSpace(*s.MyNumber)) > 0 {
		option.MyNumberIdList, err = api.myNumberService.GetMyNumberIdListByMyNumber(c.Request().Context(), *s.MyNumber)
		if err != nil {
			return errors.Wrap(err, "Failed to get list by my number")
		}
		if len(option.MyNumberIdList) == 0 {
			err = domain.OperatorHistory{}.Record(c.Request().Context(), id, email,
				"menu.manage_mynumber.8_manage", "", c.RealIP(), domain.LOOKUP)
			if err != nil {
				return factory.ErrorDB.New(err)
			}

			return apirender.RenderSucc(c, http.StatusOK, apirender.ArrayResult{
				Items:      []domain.MyNumber{},
				TotalCount: 0,
			})
		}
	}

	var items []domain.MyNumber
	totalCount, items, err := domain.MyNumber{}.GetAll(c.Request().Context(), option, s.Sortby, s.Order, s.SkipCount, s.MaxResultCount)
	if err != nil {
		return factory.ErrorParameter.New(err)
	}

	err = domain.OperatorHistory{}.Record(c.Request().Context(), id, email,
		"menu.manage_mynumber.8_manage", "", c.RealIP(), domain.LOOKUP)
	if err != nil {
		return factory.ErrorDB.New(err)
	}

	return apirender.RenderSucc(c, http.StatusOK, apirender.ArrayResult{
		Items:      items,
		TotalCount: totalCount,
	})
}

// canCompleteAuditOnMyNumberApproveは、マイナンバー承認で本登録完了
// (COMPLETED_AUDIT)にしてよいかを返す(MIR-1276、③)。以前はisEkycだけを見て
// おり、会員の現在の審査ステータスを確認していなかった。反社確認中・反社確認
// 待ち(CHECK_ANTISOCIAL_AUDIT・WAIT_ANTISOCIAL_AUDIT)の会員でも、別件で
// マイナンバーが承認された瞬間に本登録完了まで進んでしまっていた。
// 本登録完了にしてよいのは、eKYCを通過済み(isEkyc)かつ審査ステータスが最終
// 承認待機(WAIT_FINAL_APPROVAL_AUDIT)のときだけ(河内さん判断、JIRA
// MIR-1276コメント13911)。引数だけで完結する純粋関数。
//
// TODO: マイナンバー登録復活後に見直す。CHECK_MY_NUMBER_AUDIT(マイナンバー
// 承認待機、domain/user.goで現在コメントアウト中)が復活すると、その
// ステータスの会員のマイナンバー承認もこの関数を通るが、今はisEkyc &&
// WAIT_FINAL_APPROVAL_AUDITしか許可しないため、そちらを素通りできず
// 後退になる。復活時はCHECK_MY_NUMBER_AUDITも許可対象へ加えること。
func canCompleteAuditOnMyNumberApprove(status domain.AUDIT_STATUS, isEkyc bool) bool {
	return isEkyc && status == domain.WAIT_FINAL_APPROVAL_AUDIT
}

func (api *MyNumberApiController) UpdateStatus(c echo.Context) error {
	var myNumberStatus MyNumberStatusInput

	if err := c.Bind(&myNumberStatus); err != nil {
		return factory.ErrorParameter.New(err)
	}
	if err := c.Validate(&myNumberStatus); err != nil {
		return factory.ErrorParameter.New(err)
	}

	operatorId, email, err := jwthelper.TokenToUser(c)
	if err != nil {
		return err
	}

	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return factory.ErrorParameter.New(err)
	}

	op, err := domain.AdminUser{}.GetById(c.Request().Context(), operatorId)
	if err != nil {
		return factory.ErrorDB.New(err)
	} else if op == nil {
		return factory.ErrorNotFound.New(nil)
	}

	before, err := domain.MyNumber{}.GetById(c.Request().Context(), id)
	if err != nil {
		return factory.ErrorParameter.New(err)
	}
	if before == nil {
		return factory.ErrorNotFound.New(err)
	}

	myNumber, err := myNumberStatus.ToModel()
	if err != nil {
		return factory.ErrorParameter.New(err)
	}

	myNumber.Id = before.Id
	myNumber.UserId = before.UserId
	myNumber.MyNumber = before.MyNumber
	myNumber.ManagerName = op.Name

	user, err := domain.User{}.GetById(c.Request().Context(), myNumber.UserId)
	if err != nil {
		return factory.ErrorDB.New(err)
	} else if user == nil {
		return factory.ErrorNotFound.New(nil)
	}

	if myNumber.Status == domain.APPROVE {
		dupes, err := api.myNumberService.GetDuplicatedMyNumbersInApproveAndWait(c.Request().Context(), myNumber.Id)
		if err != nil {
			return factory.ErrorSystem.New(err)
		}
		if len(dupes) > 0 {
			return factory.ErrorMyNumberAlreadyExist.New(nil)
		}

		if canCompleteAuditOnMyNumberApprove(user.AuditStatus, user.IsEkyc) {
			user.ApproveSignUp(time.Now())

			// 본등록 완료 메일 발송
			err = (&domain.Mail{}).CompletedAuditStatusForPrivateMail(c.Request().Context(), serviceutils.GetServiceInfoMail(), user.Email, user.Name)
			if err != nil {
				return err
			}
		} else if user.IsEkyc {
			// MIR-1276: 以前はIsEkycだけを見て、会員の現在の審査ステータスを
			// 確認せずに本登録完了(COMPLETED_AUDIT)へ進めていた。反社確認中・
			// 反社確認待ち(CHECK_ANTISOCIAL_AUDIT・WAIT_ANTISOCIAL_AUDIT)の
			// 会員でも、別件でマイナンバーが承認された瞬間に本登録完了まで
			// 進んでしまっていた。マイナンバーの承認自体は成功として扱い
			// (myNumber.Updateは下で行う)、本登録完了にしなかったことだけ
			// 記録する。
			factory.Logger(c.Request().Context()).WithFields(logrus.Fields{
				"user_id":      user.Id,
				"audit_status": user.AuditStatus,
				"my_number_id": myNumber.Id,
			}).Warn("MyNumberUpdateStatus: Skip audit completion: audit status is not WAIT_FINAL_APPROVAL_AUDIT")
		}
		if err := user.Update(c.Request().Context()); err != nil {
			return factory.ErrorDB.New(err)
		}

	} else if myNumber.Status == domain.DISAPPROVED {
		err := domain.LookUpMyNumber{}.DeleteByMyNumberId(c.Request().Context(), myNumber.Id)
		if err != nil {
			return factory.ErrorDB.New(err)
		}

		err = (&domain.Mail{}).MyNumberResult(c.Request().Context(), serviceutils.GetServiceInfoMail(), user.Email, user.Name)
		if err != nil {
			return factory.ErrorAWSSendMail.New(err)
		}
	}

	if err := myNumber.Update(c.Request().Context()); err != nil {
		return factory.ErrorDB.New(err)
	}

	err = domain.OperatorHistory{}.Record(c.Request().Context(), operatorId, email,
		"menu.manage_mynumber.8_manage", "", c.RealIP(), domain.CHANGE)
	if err != nil {
		return factory.ErrorDB.New(err)
	}

	return apirender.RenderSucc(c, http.StatusOK, myNumber)
}

func (api *MyNumberApiController) CheckDuplicateMyNumber(c echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return factory.ErrorParameter.New(err)
	}

	dupes, err := api.myNumberService.GetDuplicatedMyNumbersInApproveAndWait(c.Request().Context(), id)
	if err != nil {
		return factory.ErrorSystem.New(err)
	}

	return apirender.RenderSucc(c, http.StatusOK, DuplicatedMyNumberResponse{}.New(dupes))
}
