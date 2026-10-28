package controllers

import (
	"errors"
	"net/http"
	"strconv"
	"time"

	adminservice "github.com/illuminarean/energy-share-core-api/service"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/echoutils"

	"github.com/illuminarean/energy-shares-backend-repo/domain"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/apirender"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/jsonutils"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/jwthelper"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/serviceutils"
	"github.com/illuminarean/energy-shares-backend-repo/service"

	"github.com/labstack/echo/v4"
	"github.com/pangpanglabs/echoswagger/v2"
	"github.com/sirupsen/logrus"

	"github.com/illuminarean/energy-shares-backend-repo/pkg/factory"
)

type UserApiController struct {
	userInfoHistoryService service.IUserInfoHistoryService
	userAdminService       *adminservice.UserAdminService
}

func NewUserApiController() *UserApiController {
	return &UserApiController{
		userInfoHistoryService: service.NewUserInfoHistoryService(),
		userAdminService:       adminservice.NewUserAdminService(),
	}
}

// 会員退会
// パスワード確認

func (c UserApiController) Init(g echoswagger.ApiGroup) {

	g.GET("", c.GetAll,
		echoutils.AllowRoles(domain.ROLE_INVESTOR)).
		AddParamQueryNested(UserSearchInput{}).
		SetDescription("ユーザーリスト情報をリクエストしています。").
		SetSummary("ユーザーリスト情報リクエスト").
		SetSecurity("Authorization")

	g.GET("/:id/smsmsgs", c.GetUserSmsMsgs,
		echoutils.AllowRoles(domain.ROLE_MESSENGER)).
		SetSummary("ユーザーSMSメッセージリスト照会").
		AddParamPath("", "id", "").
		AddParamQueryNested(SearchInput{}).
		SetDescription("ユーザーのSMSメッセージリストを照会します。").
		SetSecurity("Authorization")

	g.POST("/passwordresetrequest", c.PasswordResetRequest,
		echoutils.AllowRoles(domain.ROLE_ADMIN)).
		AddParamBody(PasswordResetRequestInput{}, "body", "パスワードリセットリクエスト", true).
		SetSummary("パスワードリセットリクエスト作成").
		SetDescription("パスワードリセット法人個人同じ").
		SetSecurity("Authorization")

	g.GET("/:id/investments", c.GetInvestmentList,
		echoutils.AllowRoles(domain.ROLE_INVESTOR)).
		SetSummary("ユーザーが投資したすべての投資取引履歴リクエスト").
		AddParamPath("", "id", "").
		AddParamQuery(MyPageSearchInput{}, "", "", true).
		AddResponse(http.StatusOK, "ユーザーが投資したすべての投資取引履歴情報", domain.Investment{}, nil).
		SetSecurity("Authorization").
		SetDescription("ユーザーが投資した投資取引履歴リクエスト")

	g.GET("/:id/investments/download", c.Download,
		echoutils.AllowRoles(domain.ROLE_INVESTOR)).
		SetSummary("ユーザーが投資したすべての投資取引履歴ダウンロードリクエスト").
		AddParamPath("", "id", "").
		AddParamQuery(MyPageSearchInput{}, "", "", true).
		AddResponse(http.StatusOK, "ユーザーが投資したすべての投資取引履歴情報ダウンロード", domain.Investment{}, nil).
		SetSecurity("Authorization").
		SetDescription("ユーザーが投資したすべての投資取引履歴情報ダウンロードリクエスト")

	g.GET("/:id/investments/:share_id", c.GetInvestment,
		echoutils.AllowRoles(domain.ROLE_INVESTOR)).
		SetSummary("投資取引履歴詳細情報リクエスト").
		AddParamPath("", "id", "").
		AddParamPath("", "share_id", "").
		AddResponse(http.StatusOK, "投資取引履歴詳細情報", domain.Investment{}, nil).
		SetSecurity("Authorization").
		SetDescription("投資取引履歴詳細情報リクエスト")

	g.POST("/:id/antisocial-check-update", c.AntisocialCheckStatusUpdate,
		echoutils.AllowRoles(domain.ROLE_INVESTOR)).
		AddParamPath("", "id", "").
		AddParamForm(UserAntisocialCheckDto{}, "body", "", true).
		SetRequestContentType("multipart/form-data").
		SetDescription("反社会勢力チェックAPI ファイル入力名:antisocial_check_file 削除ファイル入力名:antisocial_check_file_delete ").
		SetSecurity("Authorization")

	g.POST("/:id/link-virtual-account", c.NewLinkAccount,
		echoutils.AllowRoles(domain.ROLE_INVESTOR)).
		AddParamBody(LinkVaAccountProjectType{}, "body", "", true).
		AddParamPath("", "id", "").
		AddResponse(http.StatusOK, "新規仮想口座接続成功 ", domain.UserAccount{}, nil).
		SetDescription("新しい仮想口座を接続します。既存の仮想口座は使用不可になります。  ").
		SetSummary("新しい仮想口座を接続します。").
		SetSecurity("Authorization")

	g.GET("/:id/bankinfo", c.GetUserBankInfo,
		echoutils.AllowRoles(domain.ROLE_INVESTOR)).
		AddParamQueryNested(SearchInput{}).
		AddResponse(http.StatusOK, "ユーザー出金先金融機関照会", domain.UserAccountBankInfo{}, nil).
		SetDescription("ユーザー出金先金融機関を照会します。").
		SetSummary("ユーザー出金先金融機関照会").
		SetSecurity("Authorization")

	g.GET("/:id/assets/check", c.CheckRemainingAssets,
		echoutils.AllowRoles(domain.ROLE_INVESTOR)).
		AddParamPath("", "id", "").
		SetSummary("会員資産確認").
		SetDescription("会員の資産が残っているかを確認します（true | false）").
		SetSecurity("Authorization")
}

func (UserApiController) GetAll(c echo.Context) error {

	var v UserSearchInput
	if err := c.Bind(&v); err != nil {
		return factory.ErrorParameter.New(err)
	}
	if v.MaxResultCount == 0 {
		v.MaxResultCount = DefaultMaxResultCount
	}

	totalCount, items, err := domain.User{}.GetAll(c.Request().Context(), v.Id, v.Name, v.Email, v.InvestorType, v.AuditStatus, v.Sortby, v.Order, v.SkipCount, v.MaxResultCount, v.StartDate, v.EndDate)
	if err != nil {
		return factory.ErrorDB.New(err)
	}

	return apirender.RenderSucc(c, http.StatusOK, apirender.ArrayResult{
		TotalCount: totalCount,
		Items:      items,
	})
}

// AntisocialCheckStatusUpdate ユーザー反社会勢力結果変更機能
func (u UserApiController) AntisocialCheckStatusUpdate(c echo.Context) error {
	var v UserAntisocialCheckDto
	var mail domain.Mail

	if err := c.Bind(&v); err != nil {
		return factory.ErrorParameter.New(err)
	}
	if err := c.Validate(&v); err != nil {
		return factory.ErrorParameter.New(err)
	}

	form, err := c.MultipartForm()
	if err != nil {
		return factory.ErrorInvalidForm.New(err)
	}
	checkFileDeleteIds, err := parseAntisocialCheckFileDeleteIds(form.Value)
	if err != nil {
		return factory.ErrorParameter.New(err)
	}

	adminUserId, adminName, adminEmail, err := jwthelper.TokenToUserExtended(c)
	if err != nil {
		return err
	}

	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return factory.ErrorParameter.New(err)
	}

	anti, err := domain.UserAntisocialCheck{}.GetByCheckUserId(c.Request().Context(), id)
	if err != nil {
		return factory.ErrorDB.New(err)
	}
	if anti == nil {
		return factory.ErrorNotFound.New(err)
	}
	anti.CheckStatus = v.CheckStatus

	user, err := domain.User{}.GetById(c.Request().Context(), id)
	if err != nil {
		return factory.ErrorDB.New(err)
	}
	if user == nil {
		return factory.ErrorNotFound.New(err)
	}

	if user.UserType == domain.PRIVATE {
		if user.AuditStatus != domain.CHECK_ANTISOCIAL_AUDIT && user.AuditStatus != domain.HOLD_AUDIT {
			return factory.ErrorNotMatchAudit.New(err)
		}
	} else if user.UserType == domain.CORPORATE {
		if user.AuditStatus != domain.CHECK_ANTISOCIAL_AUDIT {
			return factory.ErrorNotMatchAudit.New(err)
		}
	}

	var before *service.UserInfoHistoryData
	if user.UserType == domain.PRIVATE {
		before, err = u.userInfoHistoryService.GetDataForPrivateUser(c.Request().Context(), id)
	} else if user.UserType == domain.CORPORATE {
		before, err = u.userInfoHistoryService.GetDataForCorporateUser(c.Request().Context(), id)
	}
	if err != nil {
		return err
	}

	anti.CheckDate = jsonutils.JSONTime(time.Now())
	anti.Checker = v.Checker

	if anti.CheckStatus == domain.ANTISOCIAL_CHECK_YES { // 承認
		err := user.AcceptAntiSocialCheck(c.Request().Context(), anti.CheckDate)
		if err != nil {
			return factory.ErrorDB.New(err)
		}

		if user.UserType == domain.PRIVATE {
			// 会員ステータスを「最終承認待機」へ更新
			user.AuditStatus = domain.WAIT_FINAL_APPROVAL_AUDIT
		}
	} else if anti.CheckStatus == domain.ANTISOCIAL_CHECK_NO { // 非承認
		err = mail.AntisocialCheckDisapproveEmail(c.Request().Context(), serviceutils.GetServiceInfoMail(), user.Email, user.Name)
		if err != nil {
			factory.Logger(c.Request().Context()).WithFields(logrus.Fields{
				"userType":    user.UserType,
				"auditStatus": user.AuditStatus,
				"userId":      user.Id,
				"err":         err,
			}).Error("Fail Send Disapprove Mail")
		}
		anti.InvalidReason = v.InvalidReason

		// CHECK_ANTISOCIAL_AUDIT から非承認の場合INVALIDITY_AUDITに変更
		user.AuditStatus = domain.INVALIDITY_AUDIT
	}

	if err := user.Update(c.Request().Context()); err != nil {
		return factory.ErrorDB.New(err)
	}

	if err := anti.Update(c.Request().Context()); err != nil {
		return factory.ErrorDB.New(err)
	}

	checkFilesToDelete, err := deleteOwnedAntisocialCheckFiles(c.Request().Context(), checkFileDeleteIds, anti.CheckUserId,
		logrus.Fields{"admin_user_id": adminUserId, "check_user_id": anti.CheckUserId})
	if err != nil {
		return err
	}

	anti.CheckFiles, err = domain.AntisocialCheckFile{}.AddCheckFiles(c.Request().Context(), anti.CheckUserId, form)
	if err != nil {
		return factory.ErrorAWSUpload.New(err)
	}

	var after *service.UserInfoHistoryData
	if user.UserType == domain.PRIVATE {
		after, err = u.userInfoHistoryService.GetDataForPrivateUser(c.Request().Context(), id)
	} else if user.UserType == domain.CORPORATE {
		after, err = u.userInfoHistoryService.GetDataForCorporateUser(c.Request().Context(), id)
	}
	if err != nil {
		return err
	}

	u.userInfoHistoryService.CreateUserInfoHistory(c.Request().Context(), *before, *after, adminUserId, adminName, adminEmail)

	deleteAntisocialCheckFilesAfterCommit(c.Request().Context(), checkFilesToDelete)

	return apirender.RenderSucc(c, http.StatusOK, &anti)
}

func (UserApiController) PasswordResetRequest(c echo.Context) error {
	var v PasswordResetRequestInput

	if err := c.Bind(&v); err != nil {
		return factory.ErrorParameter.New(err)
	}
	if err := c.Validate(&v); err != nil {
		return factory.ErrorParameter.New(err)
	}

	user, err := domain.User{}.GetByEmail(c.Request().Context(), v.Email)
	if err != nil {
		return factory.ErrorParameter.New(err)
	}
	if user == nil {
		return factory.ErrorUserNotExists.New(errors.New("not found email"))
	}
	if user.AuditStatus == domain.INVALIDITY_AUDIT {
		return factory.ErrorInvalidityAudit.New(err)
	}
	if user.Deleted.Unix() > 0 {
		return factory.ErrorDeletedUser.New(err)
	}
	if user.LoginType != domain.MAIL {
		return factory.ErrorSocialUser.New(err)
	}

	prr, err := domain.PasswordResetRequest{}.GetByEmail(c.Request().Context(), v.Email)
	if err != nil {
		return factory.ErrorDB.New(err)
	} else if prr == nil {
		prr, err = v.ToModel()
		if err != nil {
			return factory.ErrorParameter.New(err)
		}

		err = prr.CreatePasswordResetRequest(c.Request().Context(), v.Email)
		if err != nil {
			return factory.ErrorDB.New(err)
		}
	} else if prr != nil {
		prr.SendCount++
		if prr.SendCount > domain.MAX_SENT_PASSWORD_RESET_REQUEST_MAIL {
			return factory.ErrorSentEmail.New(nil)
		}
		_, err = prr.UpdateById(c.Request().Context())
		if err != nil {
			return factory.ErrorDB.New(err)
		}
	}

	var mail domain.Mail
	err = mail.PasswordResetMail(c.Request().Context(), serviceutils.GetServiceInfoMail(), prr.Email, user.Name, prr.Url)
	if err != nil {
		// MIR-1365: 以前はprr.Url(トークンを含むパスワード再設定リンク)もそのまま
		// ログへ出していた。IDから対象を特定できるため、宛先とURLは出さない。
		factory.Logger(c.Request().Context()).WithFields(logrus.Fields{
			"ID": prr.Id,
		}).Error("Fail Password Reset Email Send")
	}

	return apirender.RenderSucc(c, http.StatusOK, true)
}

func expireRefreshTokenCookie(c echo.Context) {
	refreshCookie := new(http.Cookie)
	refreshCookie.Name = "refresh_token"
	refreshCookie.Value = ""
	refreshCookie.HttpOnly = true
	refreshCookie.Secure = true // HTTPSのみで使用するように設定、必要に応じて設定変更
	refreshCookie.Path = "/"
	refreshCookie.Expires = time.Unix(0, 0) // クッキー有効期限を過去に設定して即座に失効

	c.SetCookie(refreshCookie)
}

func (UserApiController) GetInvestmentList(c echo.Context) error {
	var v MyPageSearchInput
	if err := c.Bind(&v); err != nil {
		return factory.ErrorParameter.New(err)
	}
	if v.MaxResultCount == 0 {
		v.MaxResultCount = DefaultMaxResultCount
	}

	factory.Logger(c.Request().Context()).WithFields(logrus.Fields{
		"sortby":           v.Sortby,
		"order":            v.Order,
		"max_result_count": v.MaxResultCount,
		"skip_count":       v.SkipCount,
	}).Info("Investments SearchStart")

	var userId int64

	userId, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return factory.ErrorParameter.New(err)
	}

	totalcount, invests, err := domain.Investment{}.
		GetInvestmentsByOwnerId(c.Request().Context(), userId, v.ProjectType, v.ProjectStatus, v.InvestStatus, v.Sortby, v.Order, v.SkipCount, v.MaxResultCount)
	if err != nil {
		return factory.ErrorDB.New(err)
	}

	return apirender.RenderSucc(c, http.StatusOK, apirender.ArrayResult{
		Items:      invests,
		TotalCount: totalcount,
	})
}

func (UserApiController) GetInvestment(c echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return factory.ErrorParameter.New(err)
	}

	shareId := c.Param("share_id")
	share, err := domain.Share{}.GetById(c.Request().Context(), shareId)
	if err != nil {
		return factory.ErrorDB.New(err)
	}
	if share == nil {
		return factory.ErrorNotFound.New(err)
	}

	if share.OwnerId != id {
		return factory.ErrorParameter.New(errors.New("owner id of share and id are not equal"))
	}

	invest, err := domain.Investment{}.GetById(c.Request().Context(), shareId)
	if err != nil {
		return factory.ErrorDB.New(err)
	}
	if invest == nil {
		return factory.ErrorNotFound.New(err)
	}

	return apirender.RenderSucc(c, http.StatusOK, invest)
}

func (UserApiController) GetUserSmsMsgs(c echo.Context) error {
	var v SearchInput
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return factory.ErrorDB.New(err)
	}
	if err := c.Bind(&v); err != nil {
		return factory.ErrorParameter.New(err)
	}
	user, err := domain.User{}.GetById(c.Request().Context(), id)
	if err != nil {
		return factory.ErrorDB.New(err)
	}

	if user == nil {
		return factory.ErrorNotFound.New(nil)
	}

	if v.MaxResultCount == 0 {
		v.MaxResultCount = DefaultMaxResultCount
	}

	factory.Logger(c.Request().Context()).WithFields(logrus.Fields{
		"sortby":           v.Sortby,
		"order":            v.Order,
		"max_result_count": v.MaxResultCount,
		"skip_count":       v.SkipCount,
	}).Info("SearchStart")

	totalCount, items, err := domain.SmsMsg{}.GetUserSmsMsgList(c.Request().Context(), user.MobilePhone, v.Sortby, v.Order, v.SkipCount, v.MaxResultCount)
	if err != nil {
		return factory.ErrorDB.New(err)
	}

	return apirender.RenderSucc(c, http.StatusOK, apirender.ArrayResult{
		Items:      items,
		TotalCount: totalCount,
	})
}

func (UserApiController) Download(c echo.Context) error {
	var v MyPageSearchInput
	if err := c.Bind(&v); err != nil {
		return factory.ErrorParameter.New(err)
	}
	if v.MaxResultCount == 0 {
		v.MaxResultCount = DefaultMaxResultCount
	}

	factory.Logger(c.Request().Context()).WithFields(logrus.Fields{
		"sortby":           v.Sortby,
		"order":            v.Order,
		"max_result_count": v.MaxResultCount,
		"skip_count":       v.SkipCount,
	}).Info("SearchStart")

	var userId int64
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return factory.ErrorParameter.New(err)
	}
	userId = id

	err = domain.User{}.DownloadCsv(c.Request().Context(), c.Response().Writer, userId, v.ProjectType, v.ProjectStatus, v.Sortby, v.Order, v.SkipCount, v.MaxResultCount)
	if err != nil {
		return factory.ErrorFailDownloadCSV.New(err)
	}

	return nil
}

func (UserApiController) NewLinkAccount(c echo.Context) error {
	var v LinkVaAccountProjectType
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return factory.ErrorParameter.New(err)
	}

	if err := c.Bind(&v); err != nil {
		return factory.ErrorParameter.New(err)
	}
	if err := c.Validate(&v); err != nil {
		return factory.ErrorParameter.New(err)
	}

	user, err := domain.User{}.GetById(c.Request().Context(), id)
	if err != nil {
		return factory.ErrorParameter.New(err)
	}
	if user == nil {
		return factory.ErrorUserNotExists.New(err)
	}

	err = domain.UserVirtualAccount{}.InvalidAccounts(c.Request().Context(), user.Id, v.ProjectType)
	if err != nil {
		return factory.ErrorDB.New(err)
	}

	va, err := domain.UserVirtualAccount{}.GetVaByUnLinkedAccount(c.Request().Context(), v.ProjectType)
	if err != nil {
		return factory.ErrorDB.New(err)
	}
	err = va.LinkUser(c.Request().Context(), user.Id)
	if err != nil {
		return factory.ErrorDB.New(err)
	}

	var mail domain.Mail
	err = mail.ReissueAccountMail(c.Request().Context(), serviceutils.GetServiceInfoMail(), user.Email, user.Name, va.VaAccountNumber)
	if err != nil {
		return factory.ErrorAWSSendMail.New(err)
	}

	return apirender.RenderSucc(c, http.StatusOK, va)
}

func (UserApiController) GetUserBankInfo(c echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return factory.ErrorParameter.New(err)
	}

	user, err := domain.User{}.GetById(c.Request().Context(), id)
	if err != nil {
		return factory.ErrorParameter.New(err)
	}
	if user == nil {
		return factory.ErrorUserNotExists.New(err)
	}

	bi := domain.UserAccountBankInfo{
		UserId:      user.Id,
		BankAccount: user.BankAccount,
		BankName:    user.BankName,
		BranchCode:  user.BranchCode,
		BranchName:  user.BranchName,
		DepositType: user.DepositType,
	}
	return apirender.RenderSucc(c, http.StatusOK, bi)
}

func (a UserApiController) CheckRemainingAssets(c echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return factory.ErrorParameter.New(errors.New("invalid id"))
	}

	// 총 자산이 남아 있는 경우 무효화 불가
	totalAsset, err := a.userAdminService.GetUserTotalAsset(c.Request().Context(), id)
	if err != nil {
		return factory.ErrorDB.New(errors.New("failed to get user total asset"))
	}

	return apirender.RenderSucc(c, http.StatusOK, totalAsset > 0)
}
