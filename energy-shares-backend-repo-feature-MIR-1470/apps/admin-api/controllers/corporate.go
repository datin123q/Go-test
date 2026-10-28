package controllers

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/illuminarean/energy-shares-backend-repo/service/dto"
	"github.com/pkg/errors"

	"github.com/gofrs/uuid"
	"github.com/labstack/echo/v4"
	"github.com/pangpanglabs/echoswagger/v2"
	"github.com/sirupsen/logrus"

	coreservice "github.com/illuminarean/energy-share-core-api/service"
	"github.com/illuminarean/energy-shares-backend-repo/domain"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/apirender"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/echoutils"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/factory"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/gmoaozorautils"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/jsonutils"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/jwthelper"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/modelhelper"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/s3utils"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/serviceutils"
	"github.com/illuminarean/energy-shares-backend-repo/service"
)

type CorporateApiController struct {
	corporateAdminService  coreservice.ICorporateAdminService
	userInfoHistoryService service.IUserInfoHistoryService
	userAdminService       *coreservice.UserAdminService
}

func NewCorporateApiController() *CorporateApiController {
	return &CorporateApiController{
		corporateAdminService:  coreservice.NewCorporateUserAdminService(),
		userInfoHistoryService: service.NewUserInfoHistoryService(),
		userAdminService:       coreservice.NewUserAdminService(),
	}
}

func (c CorporateApiController) Init(g echoswagger.ApiGroup) {
	g.GET("", c.GetAll,
		echoutils.AllowRoles(domain.ROLE_INVESTOR)).
		AddParamQueryNested(CorporateSearchInput{}).
		SetDescription("法人会員リスト情報をリクエストします.").
		SetSummary("法人会員リスト情報要求").
		SetSecurity("Authorization")

	g.GET("/user-count", c.GetTotalCount,
		echoutils.AllowRoles(domain.ROLE_INVESTOR)).
		SetDescription("法人総会員数をリクエストします.").
		SetSummary("法人総会員数情報要求").
		SetSecurity("Authorization")

	g.GET("/list/download", c.Download,
		echoutils.AllowRoles(domain.ROLE_INVESTOR)).
		AddParamQueryNested(CorporateSearchInput{}).
		SetDescription("法人会員リストをダウンロードします.").
		SetSummary("法人会員リストダウンロード要求").
		SetSecurity("Authorization")

	g.GET("/list/anti-download", c.AntiSocialDownload,
		echoutils.AllowRoles(domain.ROLE_INVESTOR)).
		AddParamQueryNested(AntiSocialCorporateSearchInput{}).
		SetDescription("法人反社会待機中会員リストをダウンロードします.").
		SetSummary("法人反社会待機中会員リストダウンロード要求").
		SetSecurity("Authorization")

	g.GET("/:id", c.GetOne,
		echoutils.AllowRoles(domain.ROLE_INVESTOR)).
		AddParamPath("", "id", "").
		AddResponse(http.StatusOK, "法人会員詳細情報", dto.CorporateUserAdminDetail{}, nil).
		SetDescription("法人会員を照会します.").
		SetSummary("法人会員照会").
		SetSecurity("Authorization")

	g.PUT("/:id/investor-type", c.UpdateInvestorType,
		echoutils.AllowRoles(domain.ROLE_INVESTOR)).
		AddParamPath("", "id", "").
		AddParamBody(InvestorTypeDto{}, "body", "投資家タイプ", true).
		SetSummary("法人会員投資家タイプ修正").
		SetDescription("法人会員投資家タイプ修正").
		SetSecurity("Authorization")

	g.PUT("/:id", c.Update,
		echoutils.AllowRoles(domain.ROLE_INVESTOR, domain.ROLE_OPERATOR)).
		AddParamPath("", "id", "").
		AddParamForm(CorporateUserInput{}, "formData", "", true).
		SetDescription("法人会員情報を修正します.").
		SetSummary("法人会員情報修正").
		SetRequestContentType("multipart/form-data").
		SetSecurity("Authorization")

	g.PUT("/:id/auditstatus", c.UpdateAuditStatus,
		echoutils.AllowRoles(domain.ROLE_INVESTOR)).
		AddParamPath("", "id", "").
		AddParamBody(UserAuditStatusDto{}, "body", "審査状態値", true).
		SetSummary("法人会員加入状態変更処理").
		SetDescription("法人会員加入状態変更処理").
		SetSecurity("Authorization")

	g.PUT("/:id/blockinvest/:block_invest", c.CorporateUpdateBlockInvest,
		echoutils.AllowRoles(domain.ROLE_INVESTOR)).
		AddParamPath("", "id", "").
		AddParamPath("", "block_invest", "").
		SetSummary("会員取引制限変更処理").
		SetDescription("会員取引制限変更処理").
		SetSecurity("Authorization")

	g.DELETE("/:id", c.Delete,
		echoutils.AllowRoles(domain.ROLE_INVESTOR)).
		AddParamPath("", "id", "").
		AddParamBody(PasswordCheckInput{}, "body", "個人会員情報削除", false).
		SetDescription("法人会員情報を削除します.").
		SetSummary("法人会員情報削除(退会)").
		SetSecurity("Authorization")

	// TODO: privateusersとAPIを合わせるのはどうでしょうか?
	g.GET("/signup-status-count", c.GetSignUpStatusCount,
		echoutils.AllowRoles(domain.ROLE_INVESTOR)).
		AddResponse(http.StatusOK, "会員加入状態統計", domain.AuditStatusCount{}, nil).
		SetDescription("会員加入別統計照会").
		SetSummary("会員加入別統計照会").
		SetSecurity("Authorization")
}

func (a *CorporateApiController) GetAll(c echo.Context) error {
	var v CorporateSearchInput
	if err := c.Bind(&v); err != nil {
		return factory.ErrorParameter.New(err)
	}
	if err := v.ToCondition().Validate(); err != nil {
		return factory.ErrorParameter.New(err)
	}
	if v.MaxResultCount == 0 {
		v.MaxResultCount = DefaultMaxResultCount
	}

	// console log
	factory.Logger(c.Request().Context()).WithFields(logrus.Fields{
		"sortby":           v.Sortby,
		"order":            v.Order,
		"max_result_count": v.MaxResultCount,
		"skip_count":       v.SkipCount,
	}).Info("SearchStart")

	totalCount, items, err := a.corporateAdminService.GetAll(c.Request().Context(), v.Id, v.Name, v.Email, v.InvestorType, v.AuditStatus, v.Sortby, v.Order, v.SkipCount, v.MaxResultCount, v.StartDate, v.EndDate, v.ToCondition())

	if err != nil {
		return factory.ErrorDB.New(err)
	}

	userId, email, err := jwthelper.TokenToUser(c)
	if err != nil {
		return factory.ErrorTokenInvalid.New(err)
	}

	err = domain.OperatorHistory{}.Record(c.Request().Context(), userId, email,
		"menu.manage_users.corporate_investors_list", "", c.RealIP(), domain.LOOKUP)
	if err != nil {
		return factory.ErrorDB.New(err)
	}

	return apirender.RenderSucc(c, http.StatusOK, apirender.ArrayResult{
		TotalCount: totalCount,
		Items:      items,
	})
}

func (CorporateApiController) GetTotalCount(c echo.Context) error {
	var v CorporateSearchInput
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

	totalCount, _, err := domain.CorporateUser{}.GetAll(c.Request().Context(), v.Id, v.Name, v.Email, v.InvestorType, v.AuditStatus, v.Sortby, v.Order, v.SkipCount, v.MaxResultCount, v.StartDate, v.EndDate, v.ToCondition())

	if err != nil {
		return factory.ErrorDB.New(err)
	}

	return apirender.RenderSucc(c, http.StatusOK, totalCount)
}

func (a *CorporateApiController) GetOne(c echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return factory.ErrorParameter.New(err)
	}
	userId, email, err := jwthelper.TokenToUser(c)
	if err != nil {
		return factory.ErrorTokenInvalid.New(err)
	}

	v, err := a.corporateAdminService.GetById(c.Request().Context(), id)
	if err != nil {
		return factory.ErrorDB.New(err)
	}
	if v == nil {
		return factory.ErrorNotFound.New(err)
	}

	err = domain.OperatorHistory{}.Record(c.Request().Context(), userId, email,
		"menu.manage_users.corporate_investors_detail", v.Email+" ("+v.Name+")", c.RealIP(), domain.LOOKUP)
	if err != nil {
		return factory.ErrorDB.New(err)
	}

	return apirender.RenderSucc(c, http.StatusOK, v)
}

func (a CorporateApiController) UpdateInvestorType(c echo.Context) error {
	var v InvestorTypeDto
	if err := c.Bind(&v); err != nil {
		return factory.ErrorParameter.New(err)
	}
	if err := c.Validate(&v); err != nil {
		return factory.ErrorParameter.New(err)
	}

	if v.InvestorType != domain.CORPORATE_INVESTOR && v.InvestorType != domain.CORPORATE_PRO_INVESTOR {
		return factory.ErrorParameter.New(errors.New("invalid investor type"))
	}

	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return factory.ErrorParameter.New(err)
	}

	adminUserId, adminName, adminEmail, err := jwthelper.TokenToUserExtended(c)
	if err != nil {
		return factory.ErrorTokenInvalid.New(err)
	}

	// before
	before, err := a.userInfoHistoryService.GetDataForCorporateUser(c.Request().Context(), id)
	if err != nil {
		return err
	}

	// investor type 변경
	user := *before.User.(*domain.CorporateUser)
	user.InvestorType = v.InvestorType
	if err := user.Update(c.Request().Context()); err != nil {
		return factory.ErrorDB.New(err)
	}

	// after
	after, err := a.userInfoHistoryService.GetDataForCorporateUser(c.Request().Context(), id)
	if err != nil {
		return err
	}
	afterUser := after.User.(*domain.CorporateUser)

	a.userInfoHistoryService.CreateUserInfoHistory(c.Request().Context(), *before, *after, adminUserId, adminName, adminEmail)

	err = domain.OperatorHistory{}.Record(c.Request().Context(), adminUserId, adminEmail,
		"menu.manage_users.corporate_investors_detail", afterUser.Email+" ("+afterUser.Name+")", c.RealIP(), domain.CHANGE)
	if err != nil {
		return factory.ErrorDB.New(err)
	}

	return apirender.RenderSucc(c, http.StatusOK, afterUser)
}

type depositorUpdateVerdict int

const (
	// depositorUpdateAllow はdepositorを検証済み(または未変更で元々正常)としてそのまま
	// 更新してよい状態。
	depositorUpdateAllow depositorUpdateVerdict = iota
	// depositorUpdateReject はdepositorを実際に変更しようとしていて、かつ変更後の値が
	// 使用できない文字種を含む状態。更新を拒否する。
	depositorUpdateReject
	// depositorUpdateWarnUnchanged はdepositorが変更されていないが、既存の値が使用できない
	// 文字種を含んだままの状態。更新自体は許可し、警告ログのみ出す。
	depositorUpdateWarnUnchanged
)

// canonicalDepositor はdepositorの実質的な内容を比較するための正規化を行う。前後の空白
// (改行・タブ・NBSP等を含む、ゼロ幅文字は対象外)を除去したうえでZenkakuToHankakuをかける。
// 全角/半角の表記ゆれ(口座名義人欄はonBlurで半角カナ→全角カナへ変換されるため、フロントが
// 送ってくる値の幅は一定しない)だけでは「変更あり」と判定しないための正規化(MIR-1171)。
func canonicalDepositor(s string) string {
	return gmoaozorautils.ZenkakuToHankaku(strings.TrimSpace(s))
}

// evaluateDepositorUpdate はdepositor(受取人名)の更新前後の値から、検証をどう扱うべきかを
// 判定する。全角/半角の表記ゆれだけの違いは「変更なし」として扱う(MIR-1171)。呼び出し側は
// updatedに前後の空白を含まない値(sanitizeDepositorForUpdateの戻り値)を渡すこと。
func evaluateDepositorUpdate(updated, existing string) depositorUpdateVerdict {
	// 既存値が空でない状態から空文字へ変更しようとした場合は、明示的な意図なく
	// depositorを消してしまう事故を防ぐため拒否する。IsValidDepositorCharsetの
	// 正規表現自体は空文字を許容するが(検証対象外の場面で空を許す必要があるため)、
	// このAPI(更新)に限っては空への変更を許可しない。canonicalDepositor同士の比較より
	// 前に判定する。既存値が全角スペースのみ(canonical化すると空文字になる)の場合、
	// 「変更なし」分岐に落ちて existing 自体の検証(全角スペースは正規表現上valid)を
	// 素通りしてしまうため。
	if updated == "" && existing != "" {
		return depositorUpdateReject
	}
	if canonicalDepositor(updated) != canonicalDepositor(existing) {
		if !gmoaozorautils.IsValidDepositorCharset(updated) {
			return depositorUpdateReject
		}
		return depositorUpdateAllow
	}
	if !gmoaozorautils.IsValidDepositorCharset(existing) {
		return depositorUpdateWarnUnchanged
	}
	return depositorUpdateAllow
}

// sanitizeDepositorForUpdate は保存前のdepositorから前後の空白(改行・タブ・NBSP等)を
// 取り除く。取り除かずに evaluateDepositorUpdate へ渡すと、末尾に付いたこれらの文字が
// 比較上は「変更なし」として見逃されたまま、生の値がDBへ保存されてしまう(MIR-1171)。
func sanitizeDepositorForUpdate(updated string) string {
	return strings.TrimSpace(updated)
}

// carryOverPostalCodeApprovalDatesはMIR-1252の修正。CorporateUserInput.ToModel()が
// 組み立てる新規構造体には郵便コード確認の承認日が含まれずゼロ値のままになるため、更新の
// たびに必ず元の値へ戻す。ここを抜けると、片方だけ確認済みの法人会員をこの画面で編集・
// 保存しただけで承認日がゼロでDBへ上書きされ(CorporateUser.UpdateがMustColsで強制書き込みする)、
// その後もう片方のコードを確認しても本登録が完了しなくなる。
// CorporatePostalCodeApprovalDate側はMIR-1252着手前から同じ理由で消えていた既存不具合で、
// 今回まとめて直す。
func carryOverPostalCodeApprovalDates(updated, origin *domain.CorporateUser) {
	updated.CodeApprovalDate = origin.CodeApprovalDate
	updated.CorporatePostalCodeApprovalDate = origin.CorporatePostalCodeApprovalDate
}

func (a CorporateApiController) Update(c echo.Context) error {
	var v CorporateUserInput
	if err := c.Bind(&v); err != nil {
		return factory.ErrorParameter.New(err)
	}
	if err := c.Validate(&v); err != nil {
		return factory.ErrorParameter.New(err)
	}

	adminUserId, adminName, adminEmail, err := jwthelper.TokenToUserExtended(c)
	if err != nil {
		return factory.ErrorTokenInvalid.New(err)
	}

	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return factory.ErrorParameter.New(err)
	}

	// 権限確認
	if !(jwthelper.HasROLE(c, domain.ROLE_ADMIN) || adminUserId == id) {
		return factory.ErrorNotHaveRole.New(err)
	}

	updatedInfo, err := v.ToModel()
	if err != nil {
		return factory.ErrorParameter.New(err)
	}

	before, err := a.userInfoHistoryService.GetDataForCorporateUser(c.Request().Context(), id)
	if err != nil {
		return err
	}
	originInfo := before.User.(*domain.CorporateUser)

	updatedInfo.User.Id = originInfo.User.Id
	updatedInfo.CorporateProfile.Id = originInfo.CorporateProfile.Id
	updatedInfo.CorporateProfile.UserId = originInfo.CorporateProfile.UserId

	// Depositor(受取人名)に漢字が含まれると、振込バッチ全体がGMOあおぞらネット銀行に
	// 拒否される(MIR-1171)。既存70件は漢字を含んだまま登録済みのため、実際に変更しようと
	// した場合だけ文字種を検証する。無条件で検証すると、Depositor以外の項目だけを
	// 直したい更新まで一律で弾いてしまう(このAPIはリクエスト全体を1件のモデルとして
	// 上書きするため、既存の不正値がリクエストにそのまま含まれてくる)。
	updatedInfo.Depositor = sanitizeDepositorForUpdate(updatedInfo.Depositor)
	switch evaluateDepositorUpdate(updatedInfo.Depositor, originInfo.Depositor) {
	case depositorUpdateReject:
		return factory.ErrorParameter.New(errors.New("invalid depositor charset"))
	case depositorUpdateWarnUnchanged:
		logrus.WithField("user_id", id).
			Warn("MIR-1171: depositor contains disallowed characters but is unchanged, skipping validation")
	}

	// is_newsletter がリクエストに含まれている場合はその値を使用
	if v.IsNewsletter != nil {
		updatedInfo.IsNewsletter = *v.IsNewsletter
		if *v.IsNewsletter && !originInfo.IsNewsletter {
			updatedInfo.NewsletterAgreeAt = jsonutils.JSONTime(time.Now())
		}
	} else {
		updatedInfo.IsNewsletter = originInfo.IsNewsletter
	}
	// BlockInvestは他のAPIで修正される情報のため、元のまま維持する
	updatedInfo.BlockInvest = originInfo.BlockInvest
	carryOverPostalCodeApprovalDates(updatedInfo, originInfo)
	if (!originInfo.PostalCodeStatus) && (updatedInfo.PostalCodeStatus) {
		updatedInfo.PostalCodeSendDate = jsonutils.JSONTime(time.Now())
	} else {
		updatedInfo.PostalCodeSendDate = originInfo.PostalCodeSendDate
	}
	if (!originInfo.CorporatePostalStatus) && (updatedInfo.CorporatePostalStatus) {
		updatedInfo.CorporatePostalCodeSendDate = jsonutils.JSONTime(time.Now())
	} else {
		updatedInfo.CorporatePostalCodeSendDate = originInfo.CorporatePostalCodeSendDate
	}

	form, err := c.MultipartForm()
	if err != nil {
		return err
	}
	if err := deleteOwnedUserFiles(c.Request().Context(), form.Value, originInfo.User.Id,
		logrus.Fields{"user_id": originInfo.User.Id}); err != nil {
		return err
	}

	beforeShareholders := originInfo.CorporateShareholder
	shareholder := make([]CorporateShareholderInput, 0)
	for key, values := range form.Value {
		if !strings.Contains(key, "shareholder_list") {
			continue
		}

		ns := strings.NewReplacer("shareholder_list[", "", "]", "")
		ss := strings.Split(ns.Replace(key), ".")

		index, err := strconv.Atoi(ss[0])
		if err != nil {
			continue
		}
		field := ss[1]
		value := values[0]

		for index >= len(shareholder) {
			shareholder = append(shareholder, CorporateShareholderInput{})
		}
		switch field {
		case "family_name":
			shareholder[index].FamilyName = value
		case "first_name":
			shareholder[index].FirstName = value
		case "family_name_kana":
			shareholder[index].FamilyNameKana = value
		case "first_name_kana":
			shareholder[index].FirstNameKana = value
		case "zip_code":
			shareholder[index].ZipCode = value
		case "address":
			shareholder[index].Address = value
		case "address2":
			shareholder[index].Address2 = value
		case "province":
			shareholder[index].Province = value
		case "city":
			shareholder[index].City = value
		case "birthday":
			shareholder[index].Birthday = value
		case "uid":
			shareholder[index].Uid = value
		}
	}

	for _, value := range shareholder {

		value.UserId = updatedInfo.User.Id
		if len(value.Uid) <= 0 {

			if err := c.Validate(&value); err != nil {
				return factory.ErrorParameter.New(err)
			}

			s, err := value.ToModel()
			if err != nil {
				return factory.ErrorParameter.New(err)
			}

			if _, err := s.Create(c.Request().Context()); err != nil {
				return factory.ErrorDB.New(err)
			}

			continue
		}
	}

	for _, bs := range beforeShareholders {

		var find bool
		for _, value := range shareholder {

			value.UserId = updatedInfo.User.Id
			if value.Uid == bs.Uid {
				find = true
				if err := c.Validate(&value); err != nil {
					return factory.ErrorParameter.New(err)
				}

				s, err := value.ToModel()
				if err != nil {
					return factory.ErrorParameter.New(err)
				}

				s.Id = bs.Id
				if err := s.Update(c.Request().Context()); err != nil {
					return factory.ErrorDB.New(err)
				}

				break
			}
		}

		if !find {

			err = domain.CorporateShareholder{}.DeleteByUid(c.Request().Context(), bs.Uid)
			if err != nil {
				return factory.ErrorDB.New(err)
			}
		}
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
				UserId:        updatedInfo.User.Id,
				OwnerId:       updatedInfo.User.Id,
				UpdateId:      updatedInfo.User.Id,
				UserType:      domain.CORPORATE,
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

	if err := updatedInfo.Update(c.Request().Context()); err != nil {
		return factory.ErrorDB.New(err)
	}

	// 4A. 法人会員に郵便認証コードが送信されます: SIGN_UP_AUDIT_ADDRESS_CHECK_FOR_CORPORATE メール送信
	if (!originInfo.PostalCodeStatus || !originInfo.CorporatePostalStatus) && (updatedInfo.PostalCodeStatus && updatedInfo.CorporatePostalStatus) {
		// 両方ともコード送信完了の場合メール送信
		if err := (&domain.Mail{}).SignUpAuditAddressCheckForCorporateEmail(c.Request().Context(), serviceutils.GetServiceInfoMail(), originInfo.Email, originInfo.Name); err != nil {
			return factory.ErrorAWSSendMail.New(err)
		}
	}

	after, err := a.userInfoHistoryService.GetDataForCorporateUser(c.Request().Context(), id)
	if err != nil {
		return err
	}
	afterUser := after.User.(*domain.CorporateUser)
	a.userInfoHistoryService.CreateUserInfoHistory(c.Request().Context(), *before, *after, adminUserId, adminName, adminEmail)

	if c.Path() == "/api/corporateusers/:id" {

		err = domain.OperatorHistory{}.Record(c.Request().Context(), adminUserId, adminEmail,
			"menu.manage_users.corporate_investors_detail", afterUser.Email+" ("+afterUser.Name+")", c.RealIP(), domain.CHANGE)
		if err != nil {
			return factory.ErrorDB.New(err)
		}

	}
	return apirender.RenderSucc(c, http.StatusOK, afterUser)
}

func (CorporateApiController) Delete(c echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return factory.ErrorParameter.New(err)
	}
	// プライベートユーザーの場合、OperatorHistoryで使用中の該当部分を理解してから適用
	// tokenId, email, err := tokenToUser(c)
	if err != nil {
		return factory.ErrorTokenInvalid.New(err)
	}

	var v PasswordCheckInput
	if err := c.Bind(&v); err != nil {
		return factory.ErrorParameter.New(err)
	}
	if err := c.Validate(&v); err != nil {
		return factory.ErrorParameter.New(err)
	}

	user, err := domain.CorporateUser{}.GetById(c.Request().Context(), id)
	if err != nil {
		return factory.ErrorDB.New(err)
	}
	if user == nil {
		return factory.ErrorNotFound.New(err)
	}
	if modelhelper.Find(user.Roles, "ROLE_SUPER_ADMIN") || user.Email == serviceutils.GetMasterEmail() {
		return factory.ErrorCanNotDeleteSuperUser.New(err)
	}

	if !user.ValidatePassword(v.Password) {
		return factory.ErrorNotMatchPassword.New(err)
	}

	if err = user.Update(c.Request().Context()); err != nil {
		return factory.ErrorDB.New(err)
	}

	err = domain.CorporateUser{}.Delete(c.Request().Context(), user.User.Id)
	if err != nil {
		return factory.ErrorDB.New(err)
	}
	return apirender.RenderSucc(c, http.StatusOK, user)
}

func (a CorporateApiController) UpdateAuditStatus(c echo.Context) error {
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
		return factory.ErrorTokenInvalid.New(err)
	}

	before, err := a.userInfoHistoryService.GetDataForCorporateUser(c.Request().Context(), id)
	if err != nil {
		return err
	}

	user := *before.User.(*domain.CorporateUser)

	switch domain.AUDIT_STATUS(v.AuditStatus) {
	case domain.WAIT_AUDIT, domain.WAIT_ANTISOCIAL_AUDIT, domain.CHECK_ANTISOCIAL_AUDIT, domain.PROVISIONAL_AUDIT:
		user.AuditStatus = domain.AUDIT_STATUS(v.AuditStatus)

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

		// TODO: リリース時点では「本登録保留」は使わない。メール送信処理だけコメントアウトしておく。
		// err = mail.HoldAuditStatusForCorporateEmail(c.Request().Context(), serviceutils.GetServiceInfoMail(), user.Email, user.Name)
		// if err != nil {
		// 	factory.Logger(c.Request().Context()).WithFields(logrus.Fields{
		// 		"auditStatus": user.AuditStatus,
		// 		"err":         err,
		// 		"userId":      user.User.Id,
		// 	}).Error("Failed to send mail (HOLD_AUDIT_STATUS_FOR_CORPORATE_TEMPLATE)")
		// 	return factory.ErrorAWSSendMail.New(err)
		// }

	case domain.CHECK_ADDRESS_AUDIT:
		user.AuditStatus = domain.AUDIT_STATUS(v.AuditStatus)

		// 郵便コード発行
		user.PostalCode, err = modelhelper.GetPostalCode()
		user.CorporatePostalCode, err = modelhelper.GetPostalCode()

		// MIR-1252: 新しいコードを発行するので、前回確認済みだった場合の承認日を
		// 残さない。CodeApprovalDate/CorporatePostalCodeApprovalDateをクリアせずに
		// 残すと、一度でも本登録を完了したユーザーをこの操作で住所確認中へ戻した
		// 際、片方の新しいコードだけを確認しただけで(もう片方は古い承認日が
		// 残っているため)CheckAllPostalCodeの完了判定を満たしてしまう。
		user.CodeApprovalDate = jsonutils.JSONTime{}
		user.CorporatePostalCodeApprovalDate = jsonutils.JSONTime{}

	case domain.COMPLETED_AUDIT:
		user.User.ApproveSignUp(time.Now())
		err = mail.CodeAuditCompleteForCorporateEmail(c.Request().Context(), serviceutils.GetServiceInfoMail(), user.Email, user.Name)
		if err != nil {
			factory.Logger(c.Request().Context()).WithFields(logrus.Fields{
				"auditStatus": user.AuditStatus,
				"err":         err,
				"userId":      user.User.Id,
			}).Error("Failed to send mail (CODE_AUDIT_COMPLETE_FOR_CORPORATE_TEMPLATE)")
			return factory.ErrorAWSSendMail.New(err)
		}

		if serviceutils.EnvIsProd() {
			// バーチャル口座割り当て
			// NOTE: 本番環境のみ自動割り当て。その他環境は手動で割り当てる。
			err = user.LinkVaAccounts(c.Request().Context(), domain.EQUITY)
			if err != nil {
				logrus.WithFields(logrus.Fields{
					"message": "Failed to link va accounts",
					"error":   err,
					"time":    time.Now().String(),
				}).Error("UpdateAuditStatus: failed to link va accounts")

				return err
			}
		}

	case domain.INVALIDITY_AUDIT:
		// 総資産が残っている場合は無効化できません
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

	after, err := a.userInfoHistoryService.GetDataForCorporateUser(c.Request().Context(), id)
	if err != nil {
		return err
	}
	afterUser := after.User.(*domain.CorporateUser)

	a.userInfoHistoryService.CreateUserInfoHistory(c.Request().Context(), *before, *after, adminUserId, adminName, adminEmail)

	err = domain.OperatorHistory{}.Record(c.Request().Context(), adminUserId, adminEmail,
		"menu.manage_users.corporate_investors_detail", afterUser.Email+" ("+afterUser.Name+")", c.RealIP(), domain.CHANGE)
	if err != nil {
		return factory.ErrorDB.New(err)
	}

	return apirender.RenderSucc(c, http.StatusOK, afterUser)
}

func (a CorporateApiController) CorporateUpdateBlockInvest(c echo.Context) error {
	adminUserId, adminName, adminEmail, err := jwthelper.TokenToUserExtended(c)
	if err != nil {
		return factory.ErrorTokenInvalid.New(err)
	}
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return factory.ErrorParameter.New(err)
	}

	blockInvest, err := strconv.ParseBool(c.Param("block_invest"))
	if err != nil {
		return factory.ErrorParameter.New(err)
	}

	before, err := a.userInfoHistoryService.GetDataForCorporateUser(c.Request().Context(), id)
	if err != nil {
		return err
	}

	user := *before.User.(*domain.CorporateUser)
	user.BlockInvest = blockInvest

	if err := user.Update(c.Request().Context()); err != nil {
		return factory.ErrorDB.New(err)
	}

	after, err := a.userInfoHistoryService.GetDataForCorporateUser(c.Request().Context(), id)
	if err != nil {
		return err
	}
	afterUser := after.User.(*domain.CorporateUser)

	a.userInfoHistoryService.CreateUserInfoHistory(c.Request().Context(), *before, *after, adminUserId, adminName, adminEmail)

	err = domain.OperatorHistory{}.Record(c.Request().Context(), adminUserId, adminEmail,
		"menu.manage_users.corporate_investors_detail", afterUser.Email+" ("+afterUser.Name+")", c.RealIP(), domain.CHANGE)
	if err != nil {
		return factory.ErrorDB.New(err)
	}

	return apirender.RenderSucc(c, http.StatusOK, afterUser)
}

func (CorporateApiController) AntiSocialDownload(c echo.Context) error {
	var v AntiSocialCorporateSearchInput
	if err := c.Bind(&v); err != nil {
		return factory.ErrorParameter.New(err)
	}
	if v.MaxResultCount == 0 {
		v.MaxResultCount = UserMaxResultCount
	}
	userId, email, err := jwthelper.TokenToUser(c)
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

	err = domain.CorporateUser{}.AntiSocialWaitUserList(c.Request().Context(), c.Response().Writer, v.SearchKeys, v.SearchValues, v.Sortby, v.Order, v.SkipCount, v.MaxResultCount, v.StartDate, v.EndDate, userId)
	if err != nil {
		return factory.ErrorFailDownloadCSV.New(err)
	}

	err = domain.OperatorHistory{}.Record(c.Request().Context(), userId, email,
		"menu.manage_users.corporate_investors_list", "", c.RealIP(), domain.DOWNLOAD)
	if err != nil {
		return factory.ErrorDB.New(err)
	}

	return nil
}

func (CorporateApiController) Download(c echo.Context) error {
	var v CorporateSearchInput
	if err := c.Bind(&v); err != nil {
		return factory.ErrorParameter.New(err)
	}
	if err := v.ToCondition().Validate(); err != nil {
		return factory.ErrorParameter.New(err)
	}
	if v.MaxResultCount == 0 {
		v.MaxResultCount = UserMaxResultCount
	}
	userId, email, err := jwthelper.TokenToUser(c)
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

	err = domain.CorporateUser{}.DownloadCsv(c.Request().Context(), c.Response().Writer, v.Id, v.Name, v.Email, v.InvestorType, v.AuditStatus, v.Sortby, v.Order, v.SkipCount, v.MaxResultCount, v.StartDate, v.EndDate, v.ToCondition(), v.InvestGrade, v.BasicInfo, v.Withdrawal, v.LoanDeposit, v.EquityDeposit, v.NewsletterAgreement)
	if err != nil {
		return factory.ErrorFailDownloadCSV.New(err)
	}

	err = domain.OperatorHistory{}.Record(c.Request().Context(), userId, email,
		"menu.manage_users.corporate_investors_list", "", c.RealIP(), domain.DOWNLOAD)
	if err != nil {
		return factory.ErrorDB.New(err)
	}

	return nil
}

func (CorporateApiController) GetSignUpStatusCount(c echo.Context) error {
	auditStatusCount, err := domain.CorporateUser{}.GetAuditStatusCount(c.Request().Context(), domain.CORPORATE)
	if err != nil {
		return factory.ErrorDB.New(err)
	}
	return apirender.RenderSucc(c, http.StatusOK, auditStatusCount)

}
