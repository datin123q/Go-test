package controllers

import (
	"encoding/base64"
	"io"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/gofrs/uuid"
	"github.com/labstack/echo/v4"
	"github.com/pangpanglabs/echoswagger/v2"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"

	adminservice "github.com/illuminarean/energy-share-core-api/service"
	"github.com/illuminarean/energy-shares-backend-repo/domain"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/apirender"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/echoutils"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/factory"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/jsonutils"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/jwthelper"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/pathutils"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/s3utils"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/serviceutils"
	"github.com/illuminarean/energy-shares-backend-repo/service"
	"github.com/illuminarean/energy-shares-backend-repo/service/dto"
	userservice "github.com/illuminarean/energyshares-user-api/service"
)

type PrivateUserApiController struct {
	privateUserAdminService adminservice.IPrivateUserAdminService
	userInfoHistoryService  service.IUserInfoHistoryService
	privateUserService      *userservice.PrivateUserService
	userAdminService        *adminservice.UserAdminService
	nexwayService           service.INexwayService
}

func NewPrivateUserApiController() *PrivateUserApiController {
	return &PrivateUserApiController{
		privateUserAdminService: adminservice.NewPrivateUserAdminService(),
		userInfoHistoryService:  service.NewUserInfoHistoryService(),
		privateUserService:      userservice.NewPrivateUserService(),
		userAdminService:        adminservice.NewUserAdminService(),
		nexwayService:           service.NewNexwayService(),
	}
}

func (c *PrivateUserApiController) Init(g echoswagger.ApiGroup) {

	g.GET("", c.GetAll,
		echoutils.AllowRoles(domain.ROLE_INVESTOR)).
		AddParamQueryNested(UserSearchInput{}).
		SetDescription("個人会員リスト情報をリクエストしています。").
		SetSummary("個人会員リスト情報リクエスト").
		AddResponse(http.StatusOK, "OK", dto.PrivateUserItem{}, nil).
		SetSecurity("Authorization")

	g.GET("/user-count", c.GetTotalCount,
		echoutils.AllowRoles(domain.ROLE_INVESTOR)).
		SetDescription("個人総会員数をリクエストしています。").
		SetSummary("個人総会員数情報リクエスト").
		SetSecurity("Authorization")

	g.GET("/list/download", c.Download,
		echoutils.AllowRoles(domain.ROLE_INVESTOR)).
		AddParamQueryNested(UserSearchInput{}).
		SetDescription("個人会員リストをダウンロードします。").
		SetSummary("個人会員リストダウンロードリクエスト").
		SetSecurity("Authorization")

	g.GET("/list/anti-download", c.AntiSocialDownload,
		echoutils.AllowRoles(domain.ROLE_INVESTOR)).
		AddParamQueryNested(AntiSocialUserSearchInput{}).
		SetDescription("個人反社会待機中会員リストをダウンロードします。").
		SetSummary("個人反社会待機中会員リストダウンロードリクエスト").
		SetSecurity("Authorization")

	g.GET("/:id", c.GetOne,
		echoutils.AllowRoles(domain.ROLE_INVESTOR)).
		AddParamPath("", "id", "").
		AddResponse(http.StatusOK, "個人ユーザー詳細情報照会", dto.PrivateUserAdminDetail{}, nil).
		SetDescription("個人会員詳細情報を照会します。").
		SetSummary("個人会員詳細情報照会").
		SetSecurity("Authorization")

	g.PUT("/:id", c.Update,
		echoutils.AllowRoles(domain.ROLE_INVESTOR, domain.ROLE_OPERATOR)).
		AddParamPath("", "id", "").
		AddParamForm(UserInput{}, "formData", "", false).
		AddResponse(http.StatusOK, "OK", dto.PrivateUserAdminDetail{}, nil).
		SetRequestContentType("multipart/form-data").
		SetDescription("個人会員情報を修正します。").
		SetSummary("個人会員情報修正").
		SetSecurity("Authorization")

	g.PUT("/:id/auditstatus", c.UpdateAuditStatus,
		echoutils.AllowRoles(domain.ROLE_INVESTOR)).
		AddParamPath("", "id", "").
		AddParamBody(UserAuditStatusDto{}, "body", "審査状態値", true).
		SetSummary("会員投資認証変更処理").
		SetDescription("会員投資認証変更処理").
		SetSecurity("Authorization")

	g.PUT("/:id/investor-type", c.UpdateInvestorType,
		echoutils.AllowRoles(domain.ROLE_INVESTOR)).
		AddParamPath("", "id", "").
		AddParamBody(InvestorTypeDto{}, "body", "投資家タイプ", true).
		SetSummary("個人会員投資家タイプ修正").
		SetDescription("個人会員投資家タイプ修正").
		SetSecurity("Authorization")

	g.PUT("/:id/blockinvest/:block_invest", c.PrivateUpdateBlockInvest,
		echoutils.AllowRoles(domain.ROLE_INVESTOR)).
		AddParamPath("", "id", "").
		AddParamPath("", "block_invest", "").
		SetSummary("会員取引制限変更処理").
		SetDescription("会員取引制限変更処理").
		SetSecurity("Authorization")

	g.GET("/:id/ekyc", c.GetEkyc,
		echoutils.AllowRoles(domain.ROLE_INVESTOR)).
		AddParamPath("", "id", "").
		AddResponse(http.StatusOK, "eKYC提出書類状態照会", service.EkycResultOutput{}, nil).
		SetDescription("個人会員eKYC情報を照会します。").
		SetSummary("個人会員eKYC情報照会").
		SetSecurity("Authorization")

	// TODO: corporateusers とAPIを統合するのはどうか?
	g.GET("/signup-status-count", c.GetSignUpStatusCount,
		echoutils.AllowRoles(domain.ROLE_INVESTOR)).
		AddResponse(http.StatusOK, "会員加入状態統計", domain.AuditStatusCount{}, nil).
		SetDescription("会員加入別統計照会").
		SetSummary("会員加入別統計照会").
		SetSecurity("Authorization")
}

func (a PrivateUserApiController) GetAll(c echo.Context) error {

	var v UserSearchInput
	if err := c.Bind(&v); err != nil {
		return factory.ErrorParameter.New(err)
	}
	if err := v.ToCondition().Validate(); err != nil {
		return factory.ErrorParameter.New(err)
	}
	if v.MaxResultCount == 0 {
		v.MaxResultCount = DefaultMaxResultCount
	}
	tokenId, email, err := jwthelper.TokenToUser(c)
	if err != nil {
		return factory.ErrorTokenInvalid.New(err)
	}

	// console log
	factory.Logger(c.Request().Context()).WithFields(logrus.Fields{
		"sortby":           v.Sortby,
		"order":            v.Order,
		"max_result_count": v.MaxResultCount,
		"skip_count":       v.SkipCount,
	}).Info("SearchStart")

	totalCount, items, err := a.privateUserAdminService.GetAll(c.Request().Context(), v.Id, v.Name, v.Email, v.InvestorType, v.AuditStatus, v.RequestId, v.Sortby, v.Order, v.SkipCount, v.MaxResultCount, v.StartDate, v.EndDate, v.ToCondition())
	if err != nil {
		return factory.ErrorDB.New(err)
	}

	if items == nil {
		items = make([]*dto.PrivateUserItem, 0)
	}

	err = domain.OperatorHistory{}.Record(c.Request().Context(), tokenId, email,
		"menu.manage_users.individual_investors_list", "", c.RealIP(), domain.LOOKUP)
	if err != nil {
		return factory.ErrorDB.New(err)
	}

	return apirender.RenderSucc(c, http.StatusOK, apirender.ArrayResult{
		TotalCount: totalCount,
		Items:      items,
	})
}

func (PrivateUserApiController) GetTotalCount(c echo.Context) error {

	var v UserSearchInput
	if err := c.Bind(&v); err != nil {
		return factory.ErrorParameter.New(err)
	}
	if err := v.ToCondition().Validate(); err != nil {
		return factory.ErrorParameter.New(err)
	}
	_, _, err := jwthelper.TokenToUser(c)
	if err != nil {
		return factory.ErrorTokenInvalid.New(err)
	}

	totalCount, _, err := domain.PrivateUser{}.GetAll(c.Request().Context(), v.Id, v.Name, v.Email, v.InvestorType, v.AuditStatus, v.RequestId, v.Sortby, v.Order, v.SkipCount, v.MaxResultCount, v.StartDate, v.EndDate, v.ToCondition())
	if err != nil {
		return factory.ErrorDB.New(err)
	}

	return apirender.RenderSucc(c, http.StatusOK, totalCount)
}

func (a PrivateUserApiController) GetOne(c echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return factory.ErrorParameter.New(err)
	}

	tokenId, email, err := jwthelper.TokenToUser(c)
	if err != nil {
		return err
	}

	user, err := a.privateUserAdminService.GetById(c.Request().Context(), id)
	if err != nil {
		return factory.ErrorDB.New(err)
	} else if user == nil {
		return factory.ErrorNotFound.New(err)
	}

	err = domain.OperatorHistory{}.Record(c.Request().Context(), tokenId, email,
		"menu.manage_users.individual_investors_detail", user.Email+" ("+user.Name+")", c.RealIP(), domain.LOOKUP)
	if err != nil {
		return factory.ErrorDB.New(err)
	}

	return apirender.RenderSucc(c, http.StatusOK, user)
}

func (a *PrivateUserApiController) Update(c echo.Context) error {

	var userInput UserInput

	form, err := c.MultipartForm()
	if err != nil {
		return err
	}

	if err := c.Bind(&userInput); err != nil {
		return factory.ErrorParameter.New(err)
	}

	if err := c.Validate(&userInput); err != nil {
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

	// 権限チェック
	if !(jwthelper.HasROLE(c, domain.ROLE_ADMIN) || adminUserId == id) {
		return factory.ErrorNotHaveRole.New(err)
	}

	user, err := userInput.ToModel()
	if err != nil {
		return factory.ErrorParameter.New(err)
	}

	before, err := a.userInfoHistoryService.GetDataForPrivateUser(c.Request().Context(), id)
	if err != nil {
		return err
	}
	beforeUser := before.User.(*domain.PrivateUser)

	user.User.Id = beforeUser.User.Id
	user.PrivateProfile.UserId = beforeUser.PrivateProfile.UserId
	user.PrivateProfile.Id = beforeUser.PrivateProfile.Id

	// is_newsletter がリクエストに含まれている場合はその値を使用
	if userInput.IsNewsletter != nil {
		user.IsNewsletter = *userInput.IsNewsletter
		if *userInput.IsNewsletter && !beforeUser.IsNewsletter {
			user.NewsletterAgreeAt = jsonutils.JSONTime(time.Now())
		}
	} else {
		user.IsNewsletter = beforeUser.IsNewsletter
	}
	// BlockInvestは別のAPIで修正する情報であるため、原本のままにしておく
	user.BlockInvest = beforeUser.BlockInvest

	restrictPersonalInfoEditsAfterFinalAudit(user, beforeUser)

	if (!beforeUser.PostalCodeStatus) && (user.PostalCodeStatus) {
		user.PostalCodeSendDate = jsonutils.JSONTime(time.Now())
	}

	if err := deleteOwnedUserFiles(c.Request().Context(), form.Value, beforeUser.User.Id,
		logrus.Fields{"user_id": beforeUser.User.Id}); err != nil {
		return err
	}

	for name, value := range form.File {

		fileType := domain.UserFileTypePicker(name)
		if fileType == "" {
			continue
		}

		for _, mFile := range value {
			userFile := &domain.UserFile{
				FileName:      s3utils.MakeS3FileName(),
				FileType:      fileType,
				OriginalName:  mFile.Filename,
				UserId:        user.User.Id,
				OwnerId:       user.User.Id,
				UpdateId:      user.User.Id,
				UserType:      domain.PRIVATE,
				FileSize:      mFile.Size,
				FileExtension: s3utils.GetExtension(mFile.Filename),
				FilePath:      s3utils.MakeS3Path(domain.USER_FILE),
				FileUrl:       "",
			}
			userFile.FileUrl, err = s3utils.UploadShowFile(c.Request().Context(), s3utils.BUCKET_PRIVATE_NAME, userFile, mFile)
			if err != nil {
				return factory.ErrorAWSUpload.New(err)
			}

			if _, err := userFile.Create(c.Request().Context()); err != nil {
				return factory.ErrorDB.New(err)
			}
		}
	}

	if err := user.Update(c.Request().Context()); err != nil {
		return factory.ErrorDB.New(err)
	}

	after, err := a.userInfoHistoryService.GetDataForPrivateUser(c.Request().Context(), id)
	if err != nil {
		return err
	}
	afterUser := after.User.(*domain.PrivateUser)

	a.userInfoHistoryService.CreateUserInfoHistory(c.Request().Context(), *before, *after, adminUserId, adminName, adminEmail)

	userDetail, err := a.privateUserAdminService.GetById(c.Request().Context(), id)
	if err != nil {
		return factory.ErrorDB.New(err)
	}

	err = domain.OperatorHistory{}.Record(c.Request().Context(), adminUserId, adminEmail,
		"menu.manage_users.individual_investors_detail", afterUser.Email+" ("+afterUser.Name+")", c.RealIP(), domain.CHANGE)
	if err != nil {
		return factory.ErrorDB.New(err)
	}

	return apirender.RenderSucc(c, http.StatusOK, userDetail)
}

// restrictPersonalInfoEditsAfterFinalAudit はMIR-1173で追加。国籍・性別・生年月日・
// 電話番号は本登録承認前(AuditStatusがCOMPLETED_AUDIT・INVALIDITY_AUDIT以外)の
// 会員だけ変更できる。画面側は承認完了・無効の会員に対してこれらの入力欄を
// disabledにするが、APIへ直接不正な値を送られた場合に備えてここでも弾く
// (黙って元の値へ戻す。BlockInvestと同じ扱い)。判定にはbeforeUser(更新前の
// AuditStatus)を使う。同じリクエストでAuditStatusを変更する経路はこの
// ハンドラには無いため、before/afterでAuditStatus自体が揺れる心配はない。
func restrictPersonalInfoEditsAfterFinalAudit(user, beforeUser *domain.PrivateUser) {
	if !beforeUser.AuditStatus.IsFinalized() {
		return
	}
	user.PrivateProfile.Nationality = beforeUser.PrivateProfile.Nationality
	user.PrivateProfile.Gender = beforeUser.PrivateProfile.Gender
	user.PrivateProfile.Birthday = beforeUser.PrivateProfile.Birthday
	user.User.MobilePhone = beforeUser.User.MobilePhone
}

func (a PrivateUserApiController) UpdateAuditStatus(c echo.Context) error {
	var v UserAuditStatusDto
	var mail domain.Mail

	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return factory.ErrorParameter.New(err)
	}

	if err := c.Bind(&v); err != nil {
		return factory.ErrorParameter.New(err)
	}
	adminUserId, adminName, adminEmail, err := jwthelper.TokenToUserExtended(c)
	if err != nil {
		return err
	}

	before, err := a.userInfoHistoryService.GetDataForPrivateUser(c.Request().Context(), id)
	if err != nil {
		return err
	}

	user := *before.User.(*domain.PrivateUser)

	switch domain.AUDIT_STATUS(v.AuditStatus) {
	case domain.WAIT_AUDIT, domain.WAIT_ANTISOCIAL_AUDIT, domain.CHECK_ANTISOCIAL_AUDIT, domain.PROVISIONAL_AUDIT, domain.WAIT_FINAL_APPROVAL_AUDIT:
		// TODO: マイナンバー登録復活時にcase文に追加すること
		// domain.CHECK_MY_NUMBER_AUDIT

		user.AuditStatus = domain.AUDIT_STATUS(v.AuditStatus)
		// FIXME: CHECK_MY_NUMBER_AUDIT 로 변경시 기존 승인 됐던 마이넘버는 어떻게 처리할지 (세동님께 공유 완료, 추후 결정하기로 함)

	case domain.HOLD_AUDIT:
		user.AuditStatus = domain.AUDIT_STATUS(v.AuditStatus)
		uId, err := uuid.NewV4()

		if err != nil {
			factory.Logger(c.Request().Context()).WithFields(logrus.Fields{
				"auditStatus": user.AuditStatus,
				"err":         err,
				"userId":      user.User.Id,
			}).Error("Failed to create HOLD_USER check code (uuid)")
			return factory.ErrorSystem.New(err)
		}
		holdUserCheckCode := uId.String()
		user.HoldUserCheckCode = &holdUserCheckCode

	case domain.COMPLETED_AUDIT:
		// 本登録承認完了
		user.User.ApproveSignUp(time.Now())

		if serviceutils.EnvIsProd() {
			// バーチャル口座割り当て
			// NOTE: 本番環境のみ自動割り当て。その他環境は手動で割り当てる。
			err = user.LinkVaAccounts(c.Request().Context(), domain.EQUITY)
			if err != nil {
				logrus.WithFields(logrus.Fields{
					"message": "Failed to link va accounts",
					"error":   err,
					"time":    time.Now().String(),
				}).Error("UpdateAuditStatus")

				return err
			}
		}

		err = mail.CompletedAuditStatusForPrivateMail(c.Request().Context(), serviceutils.GetServiceInfoMail(), user.Email, user.Name)
		if err != nil {
			return factory.ErrorAWSSendMail.New(err)
		}

	case domain.INVALIDITY_AUDIT:
		// 총 자산이 남아 있는 경우 무효화 불가
		totalAsset, err := a.userAdminService.GetUserTotalAsset(c.Request().Context(), user.User.Id)
		if err != nil {
			return factory.ErrorDB.New(errors.New("failed to get user total asset"))
		}
		if totalAsset > 0 {
			return factory.ErrorInvalidAuditStatus.New(errors.New("cannot update to INVALIDITY_AUDIT cause user has asset: " + strconv.FormatInt(totalAsset, 10)))
		}

		user.AuditStatus = domain.AUDIT_STATUS(v.AuditStatus)
		user.InvalidReason = v.InvalidReason
		err = mail.SendAuditInvalidityToUserEmail(c.Request().Context(), serviceutils.GetServiceInfoMail(), user.Email, user.Name, *user.InvalidReason)
		if err != nil {
			factory.Logger(c.Request().Context()).WithFields(logrus.Fields{
				"auditStatus":   user.AuditStatus,
				"err":           err,
				"userId":        user.User.Id,
				"invalidReason": user.InvalidReason,
			}).Error("Failed to send mail (SEND_AUDIT_INVALIDITY_TO_USER_TEMPLATE)")
			return factory.ErrorAWSSendMail.New(err)
		}

	default:
		return factory.ErrorParameter.New(err)
	}

	if err := user.Update(c.Request().Context()); err != nil {
		return factory.ErrorDB.New(err)
	}

	after, err := a.userInfoHistoryService.GetDataForPrivateUser(c.Request().Context(), id)
	if err != nil {
		return err
	}
	afterUser := after.User.(*domain.PrivateUser)

	a.userInfoHistoryService.CreateUserInfoHistory(c.Request().Context(), *before, *after, adminUserId, adminName, adminEmail)

	err = domain.OperatorHistory{}.Record(c.Request().Context(), adminUserId, adminEmail,
		"menu.manage_users.individual_investors_detail", afterUser.Email+" ("+afterUser.Name+")", c.RealIP(), domain.CHANGE)
	if err != nil {
		return factory.ErrorDB.New(err)
	}
	return apirender.RenderSucc(c, http.StatusOK, user)
}

func (a PrivateUserApiController) UpdateInvestorType(c echo.Context) error {
	var v InvestorTypeDto

	if err := c.Bind(&v); err != nil {
		return factory.ErrorParameter.New(err)
	}

	if v.InvestorType != domain.PRIVATE_INVESTOR && v.InvestorType != domain.PRIVATE_PRO_INVESTOR {
		return factory.ErrorParameter.New(errors.New("invalid investor type"))
	}

	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return factory.ErrorParameter.New(err)
	}

	adminUserId, adminName, adminEmail, err := jwthelper.TokenToUserExtended(c)
	if err != nil {
		return err
	}

	// before
	before, err := a.userInfoHistoryService.GetDataForPrivateUser(c.Request().Context(), id)
	if err != nil {
		return err
	}

	// 投資家タイプ変更
	user := *before.User.(*domain.PrivateUser)
	user.InvestorType = v.InvestorType
	if err := user.Update(c.Request().Context()); err != nil {
		return factory.ErrorDB.New(err)
	}

	// after
	after, err := a.userInfoHistoryService.GetDataForPrivateUser(c.Request().Context(), id)
	if err != nil {
		return err
	}
	afterUser := after.User.(*domain.PrivateUser)

	a.userInfoHistoryService.CreateUserInfoHistory(c.Request().Context(), *before, *after, adminUserId, adminName, adminEmail)

	err = domain.OperatorHistory{}.Record(c.Request().Context(), adminUserId, adminEmail,
		"menu.manage_users.individual_investors_detail", afterUser.Email+" ("+afterUser.Name+")", c.RealIP(), domain.CHANGE)
	if err != nil {
		return factory.ErrorDB.New(err)
	}

	return apirender.RenderSucc(c, http.StatusOK, afterUser)
}

func (a PrivateUserApiController) PrivateUpdateBlockInvest(c echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return factory.ErrorParameter.New(err)
	}

	adminUserId, adminName, adminEmail, err := jwthelper.TokenToUserExtended(c)
	if err != nil {
		return err
	}

	blockInvest, err := strconv.ParseBool(c.Param("block_invest"))
	if err != nil {
		return factory.ErrorParameter.New(err)
	}

	before, err := a.userInfoHistoryService.GetDataForPrivateUser(c.Request().Context(), id)
	if err != nil {
		return err
	}

	user := *before.User.(*domain.PrivateUser)
	user.BlockInvest = blockInvest

	if err := user.Update(c.Request().Context()); err != nil {
		return factory.ErrorDB.New(err)
	}

	after, err := a.userInfoHistoryService.GetDataForPrivateUser(c.Request().Context(), id)
	if err != nil {
		return err
	}
	afterUser := after.User.(*domain.PrivateUser)

	a.userInfoHistoryService.CreateUserInfoHistory(c.Request().Context(), *before, *after, adminUserId, adminName, adminEmail)

	err = domain.OperatorHistory{}.Record(c.Request().Context(), adminUserId, adminEmail,
		"menu.manage_users.individual_investors_detail", afterUser.Email+" ("+afterUser.Name+")", c.RealIP(), domain.CHANGE)
	if err != nil {
		return factory.ErrorDB.New(err)
	}

	return apirender.RenderSucc(c, http.StatusOK, afterUser)
}

func imageToBase64(imagePath string) (string, error) {

	imageFile, err := os.Open(imagePath)
	if err != nil {
		return "", err
	}
	defer imageFile.Close()

	imageData, err := io.ReadAll(imageFile)
	if err != nil {
		return "", err
	}

	base64Encoded := base64.StdEncoding.EncodeToString(imageData)

	return base64Encoded, nil
}

func (a *PrivateUserApiController) GetEkyc(c echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return factory.ErrorParameter.New(err)
	}

	userId, email, err := jwthelper.TokenToUser(c)
	if err != nil {
		return err
	}

	type EkycResultResponse struct {
		FrontImage string `json:"front_image" form:"front_image" swagger:"desc(신분증 앞면 이미지)"`
		BackImage  string `json:"back_image" form:"back_image" swagger:"desc(신분증 뒷면 이미지)"`
	}

	// dev、qa環境は一時画像を出力して終了
	if serviceutils.GetEnv() == "DEV" || serviceutils.GetEnv() == "QA" {
		frontImage := pathutils.GetPathOfAssets() + "/img/ekyc_sample_front.jpg"
		backImage := pathutils.GetPathOfAssets() + "/img/ekyc_sample_back.png"

		encodedFront, err := imageToBase64(frontImage)
		if err != nil { // エラーメッセージを合わせるためにnexwayサービス呼び出し失敗として処理
			return factory.ErrorSystem.New(errors.New("Failed call nexway service"))
		}
		encodedBack, err := imageToBase64(backImage)
		if err != nil { // エラーメッセージを合わせるためにnexwayサービス呼び出し失敗として処理
			return factory.ErrorSystem.New(errors.New("Failed call nexway service"))
		}

		err = domain.OperatorHistory{}.Record(c.Request().Context(), userId, email,
			"menu.manage_users.individual_investors_detail", "", c.RealIP(), domain.LOOKUP)
		if err != nil {
			return factory.ErrorDB.New(err)
		}

		return apirender.RenderSucc(c, http.StatusOK, EkycResultResponse{FrontImage: encodedFront, BackImage: encodedBack})
	}

	// nexway ekyc情報照会
	ekycRequestLog := domain.EkycRequestLog{}
	ekycReq, err := ekycRequestLog.GetLatest(c.Request().Context(), id)
	if err != nil {
		return factory.ErrorDB.New(err)
	}
	if ekycReq == nil {
		return factory.ErrorNotFound.New(errors.New("Failed to find ekyc request log"))
	}

	ekycResult := &service.EkycResultOutput{}
	if ekycReq.ApplicantId != 0 {
		ekycResult, err = a.nexwayService.GetEkycResultByApplicantId(c.Request().Context(), strconv.FormatInt(ekycReq.ApplicantId, 10))
		if err != nil {
			return factory.ErrorSystem.New(errors.Wrap(err, "Failed call nexway service"))
		}
		if ekycResult.ErrorCode != "" {
			return factory.ErrorSystem.New(errors.New("Failed call nexway service"))
		}
	}

	err = domain.OperatorHistory{}.Record(c.Request().Context(), userId, email,
		"menu.manage_users.individual_investors_detail", "", c.RealIP(), domain.LOOKUP)
	if err != nil {
		return factory.ErrorDB.New(err)
	}

	// 前後の画像を返す
	var front string
	for _, image := range ekycResult.ApplicantImages {
		if image.ImageType == "front" {
			front = image.ImageData
			break
		}
	}
	var back string
	for _, image := range ekycResult.ApplicantImages {
		if image.ImageType == "back" {
			back = image.ImageData
			break
		}
	}

	return apirender.RenderSucc(c, http.StatusOK, EkycResultResponse{FrontImage: front, BackImage: back})
}

func (a PrivateUserApiController) Download(c echo.Context) error {
	var v UserSearchInput
	if err := c.Bind(&v); err != nil {
		return factory.ErrorParameter.New(err)
	}
	if err := v.ToCondition().Validate(); err != nil {
		return factory.ErrorParameter.New(err)
	}
	if v.MaxResultCount == 0 {
		v.MaxResultCount = UserMaxResultCount
	}

	// console log
	factory.Logger(c.Request().Context()).WithFields(logrus.Fields{
		"sortby":           v.Sortby,
		"order":            v.Order,
		"max_result_count": v.MaxResultCount,
		"skip_count":       v.SkipCount,
	}).Info("SearchStart")

	err := a.privateUserAdminService.DownloadCsv(c.Request().Context(), c.Response().Writer, v.Id, v.Name, v.Email, v.InvestorType, v.AuditStatus, v.RequestId, v.Sortby, v.Order, v.SkipCount, v.MaxResultCount, v.StartDate, v.EndDate, v.ToCondition(), v.InvestGrade, v.BasicInfo, v.Withdrawal, v.LoanDeposit, v.EquityDeposit, v.NewsletterAgreement)
	if err != nil {
		return factory.ErrorFailDownloadCSV.New(err)
	}

	return nil
}

func (PrivateUserApiController) AntiSocialDownload(c echo.Context) error {
	var v AntiSocialUserSearchInput
	if err := c.Bind(&v); err != nil {
		return factory.ErrorParameter.New(err)
	}
	if v.MaxResultCount == 0 {
		v.MaxResultCount = UserMaxResultCount
	}

	// console log
	factory.Logger(c.Request().Context()).WithFields(logrus.Fields{
		"sortby":           v.Sortby,
		"order":            v.Order,
		"max_result_count": v.MaxResultCount,
		"skip_count":       v.SkipCount,
	}).Info("SearchStart")

	tokenId, _, err := jwthelper.TokenToUser(c)
	if err != nil {
		return err
	}

	err = domain.PrivateUser{}.AntiSocialWaitUserList(c.Request().Context(), c.Response().Writer, v.SearchKeys, v.SearchValues, v.Sortby, v.Order, v.SkipCount, v.MaxResultCount, v.StartDate, v.EndDate, tokenId)
	if err != nil {
		return factory.ErrorFailDownloadCSV.New(err)
	}

	return nil
}

func (PrivateUserApiController) GetSignUpStatusCount(c echo.Context) error {
	auditStatusCount, err := domain.PrivateUser{}.GetAuditStatusCount(c.Request().Context(), domain.PRIVATE)
	if err != nil {
		return factory.ErrorDB.New(err)
	}
	return apirender.RenderSucc(c, http.StatusOK, auditStatusCount)
}
