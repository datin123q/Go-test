package controllers

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/illuminarean/energy-share-core-api/service"
	cdomain "github.com/illuminarean/energy-shares-backend-repo/clean-arch-domain"
	"github.com/illuminarean/energy-shares-backend-repo/domain"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/config"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/echoutils"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/factory"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/jsonutils"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/serviceutils"
	"github.com/labstack/echo/v4"
	"github.com/pangpanglabs/echoswagger/v2"
	"github.com/samber/lo"
	"github.com/segmentio/ksuid"
	"github.com/sirupsen/logrus"
)

const LIMITED_AMOUNT int64 = 5000000

type ProjectApiController struct {
	InvestmentService            service.InvestmentService
	userVirtualAccountRepository cdomain.UserVirtualAccountRepository
	sqsClient                    *sqs.Client
	queueURL                     string
}

var (
	getInvestmentWaitersTotalAmount = domain.InvestmentWaiters{}.GetTotalAmountForWaitingStatus
	getShareAmountByUserID          = domain.Share{}.GetAmountByUserId
	checkUserOverAge                = func(ctx context.Context, user *domain.User) error {
		return user.CheckOverAge(ctx)
	}
	getUserByIDForInvestmentWait = func(ctx context.Context, id int64) (*domain.User, error) {
		return domain.User{}.GetById(ctx, id)
	}
	getProjectByIDForInvestmentWait = func(ctx context.Context, id int64) (*domain.Project, error) {
		return domain.Project{}.GetById(ctx, id)
	}
	publishInvestmentWaitRequestEvent = domain.CreateInvestmentWaitRequestEvent
	sendInvestStandbyMail             = func(ctx context.Context, sender, email, name, projectName, requestDate, sharesCount, investmentAmount string) error {
		return domain.Mail{}.InvestStandbyMail(ctx, sender, email, name, projectName, requestDate, sharesCount, investmentAmount)
	}
)

func NewProjectApiController(userVirtualAccountRepository cdomain.UserVirtualAccountRepository) *ProjectApiController {
	return &ProjectApiController{
		userVirtualAccountRepository: userVirtualAccountRepository,
	}
}

func (c *ProjectApiController) Init(g echoswagger.ApiGroup, cfg *config.Config, sqsClient *sqs.Client) {
	c.sqsClient = sqsClient
	c.queueURL = cfg.AWS.INVEST_WORKER_QUEUE_URL

	g.GET("/list", c.GetMainListProjects).
		AddParamQueryNested(ProjectListSearchInput{}).
		AddResponse(http.StatusOK, "ファンド商品一覧 DTO", ProjectItemsResponseDto{}, nil).
		SetDescription("ファンド商品一覧を取得します。").
		SetSummary("ファンド商品一覧取得")

	g.GET("/:id/detail", c.GetOnePublic, jwtFromHeaderMiddleware(cfg.Jwt.Secret)).
		AddParamPath("", "id", "").
		SetDescription("メイン画面のファンドを詳細取得します。").
		AddResponse(http.StatusOK, "ファンド商品詳細 DTO", ProjectDetailResponseDto{}, nil).
		SetSummary("メイン画面ファンド詳細取得")

	g.POST("/:id/invest", c.InvestProjectShares,
		echoutils.AllowRoles(domain.ROLE_USER),
		echoutils.AllowAuditStatuses(domain.COMPLETED_AUDIT)).
		AddParamPath("", "id", "").
		AddParamBody(ShareInvestInput{}, "body", "", true).
		AddResponse(http.StatusAccepted, "投資リクエスト受付結果", InvestProjectSharesResponse{}, nil).
		SetDescription("ファンドへの投資依頼をキューに登録します。").
		SetSummary("ファンド投資（非同期）").
		SetSecurity("Authorization")

	g.POST("/:id/check-invest", c.CheckInvest,
		echoutils.AllowRoles(domain.ROLE_USER),
		echoutils.AllowAuditStatuses(domain.COMPLETED_AUDIT)).
		AddParamPath("", "id", "").
		AddParamBody(ShareCheckInput{}, "body", "", true).
		SetDescription("該当ファンドへの投資が可能かどうかを確認します。").
		SetSummary("ファンド投資可否確認").
		SetSecurity("Authorization")

	g.GET("/:id/catalog", c.GetCategorys,
		echoutils.AllowRoles(domain.ROLE_USER)).
		AddParamPath("", "id", "").
		SetDescription("ファンドカタログをダウンロードします。").
		SetSummary("ファンドカタログダウンロード").
		SetSecurity("Authorization")

	g.POST("/:id/investment-wait-apply", c.InvestmentWaitCreate,
		echoutils.AllowRoles(domain.ROLE_USER),
		echoutils.AllowAuditStatuses(domain.COMPLETED_AUDIT)).
		AddParamBody(InvestmentWaitersInput{}, "body", "", true).
		AddParamPath("", "id", "").
		AddResponse(http.StatusCreated, "投資待機者作成", domain.InvestmentWaiters{}, nil).
		SetDescription("ファンド投資待機者を追加します。").
		SetSummary("ファンド投資待機者追加").
		SetSecurity("Authorization")

	g.GET("/investment-request-check/:request_id", c.InvestmentRequestCheck,
		echoutils.AllowRoles(domain.ROLE_USER)).
		SetSummary("ファンド投資待機者 商品購入コード確認").
		AddParamPath("", "request_id", "").
		SetDescription("ファンド投資待機者の商品購入コードを確認します。").
		SetSecurity("Authorization")

	g.POST("/:id/proposed-invest/:request_id", c.ProposedInvestProjectShares,
		echoutils.AllowRoles(domain.ROLE_USER),
		echoutils.AllowAuditStatuses(domain.COMPLETED_AUDIT)).
		AddParamPath("", "id", "").
		AddParamPath("", "request_id", "").
		AddParamBody(ShareInvestInput{}, "body", "", true).
		AddResponse(http.StatusCreated, "購入完了 Share", domain.Share{}, nil).
		SetDescription("ファンド投資待機者が該当プロジェクトに投資します。").
		SetSummary("ファンド投資待機者によるプロジェクト投資").
		SetSecurity("Authorization")

	g.GET("/:id/preview/:user_id", c.GetOnePreview).
		AddParamPath("", "id", "").
		AddParamPath("", "user_id", "").
		AddResponse(http.StatusOK, "", ProjectDetailResponseDto{}, nil).
		SetDescription("未公開・プレビュー用ファンドを詳細取得します。").
		SetSummary("未公開プレビュー用ファンド詳細取得")

	g.GET("/closed/:slug/detail", c.GetOneClosed, jwtFromHeaderMiddleware(cfg.Jwt.Secret)).
		AddParamPath("", "slug", "").
		SetDescription("限定公開ファンドを詳細取得します。").
		AddResponse(http.StatusOK, "ファンド商品詳細 DTO", ProjectDetailResponseDto{}, nil).
		SetSummary("限定公開ファンド詳細取得")

}

func (ProjectApiController) GetOnePublic(c echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return factory.ErrorParameter.New(err)
	}
	v, err := domain.Project{}.GetByIdForPublic(c.Request().Context(), id, false)
	if err != nil {
		return factory.ErrorDB.New(err)
	}
	if v == nil {
		return factory.ErrorNotFound.New(nil)
	}
	if v.IsClosedFund {
		return factory.ErrorNotFound.New(nil)
	}

	if v.SelectionType == domain.PROJECT_SELECTION_LOTTERY {
		// 抽選では final_funding_amount を申込時に加算しているのでそれをそのまま使う
		v.CurrentShareAmount = v.FinalFundingAmount
	} else if v.ProjectStatus == domain.OPEN_FUNDING || v.ProjectStatus == domain.STANDBY_FUNDING {
		// 募集中は最新の申込金額を常に参照する
		currentShareAmount, err := getCurrentShareAmountForUserView(c.Request().Context(), v.Id)
		if err != nil {
			return factory.ErrorDB.New(err)
		}
		v.CurrentShareAmount = currentShareAmount
	} else if v.FinalFundingAmount > 0 {
		// 募集終了時の申込金額が設定されている場合は優先して設定
		v.CurrentShareAmount = v.FinalFundingAmount
	} else {
		// 募集終了後に申込金額が確定していない場合は最新値を参照
		currentShareAmount, err := getCurrentShareAmountForUserView(c.Request().Context(), v.Id)
		if err != nil {
			return factory.ErrorDB.New(err)
		}
		v.CurrentShareAmount = currentShareAmount
	}

	return renderSucc(c, http.StatusOK, ProjectDetailResponseDto{}.ToProjectPubResponseDto(v))
}

func (c ProjectApiController) InvestProjectShares(ctx echo.Context) error {
	uid, _, err := tokenToUser(ctx)
	if err != nil {
		return err
	}

	// パスパラメータのproject_idを取得
	id, err := strconv.ParseInt(ctx.Param("id"), 10, 64)
	if err != nil {
		return factory.ErrorParameter.New(err)
	}

	var v ShareInvestInput

	// リクエストボディのバインド
	if err := ctx.Bind(&v); err != nil {
		return factory.ErrorParameter.New(err)
	}

	// バリデーションチェック
	if err := ctx.Validate(&v); err != nil {
		return factory.ErrorParameter.New(err)
	}

	// リクエストボディのuserIdとJWTトークンのuserIdが合致するか検証
	if uid != v.UserId {
		return factory.ErrorNotMatchOwner.New(nil)
	}

	factory.Logger(ctx.Request().Context()).WithFields(logrus.Fields{
		"project_id": v.ProjectId,
		"user_id":    uid,
		"qty":        v.Qty,
	}).Info("Request Invest")

	// プロジェクト情報を取得
	p, err := domain.Project{}.GetById(ctx.Request().Context(), id)
	if err != nil {
		return factory.ErrorParameter.New(err)
	}
	if err := validateInvestProjectSelectionType(p); err != nil {
		return err
	}

	// ユーザー情報を取得
	user, err := domain.User{}.GetById(ctx.Request().Context(), uid)
	if err != nil {
		return factory.ErrorDB.New(err)
	}

	// マイナンバーのステータスを確認
	if user.UserType == domain.PRIVATE {
		m, err := domain.MyNumber{}.CheckMyNumberStatusByUserId(ctx.Request().Context(), user.Id)
		if err != nil {
			return factory.ErrorDB.New(err)
		}
		if !m {
			return factory.ErrorMyNumberNotRegistered.New(fmt.Errorf("fail my number load to user id:%v", user.Id))
		}
	}

	// ブラックリストのステータスを確認
	if user.BlockInvest {
		return factory.ErrorBlockInvest.New(nil)
	}

	amount := p.SharePrice * v.Qty

	// 既存のProjectモデルを変更するのは難しいため、既存のSharesを参照して成功した投資家数を取得
	_, projectShares, err := domain.Share{}.GetByProjectId(ctx.Request().Context(), p.Id)
	if err != nil {
		return factory.ErrorDB.New(err)
	}

	// 成功した投資家はSUCCESS_INVEST, OPERATION_INVEST, REFUND_INVESTで定義される
	successInvestorCount := lo.Reduce(projectShares, func(sum int, share domain.Share, _ int) int {
		if lo.Contains(
			[]domain.INVEST_STATUS{
				domain.SUCCESS_INVEST,
				domain.OPERATION_INVEST,
				domain.REFUND_INVEST,
			}, share.InvestStatus) {
			return sum + 1
		}

		return sum
	}, 0)

	// 投資家数500人未満の制限
	if p.IsInvestorLimit && (successInvestorCount >= domain.TargetInvestorCount) {
		return factory.ErrorOverInvestorCount.New(nil)
	}

	// 投資金額の制限
	if p.IsAmountLimit {

		investedAmount, err := domain.Share{}.GetAmountByUserId(ctx.Request().Context(), p.Id, user.Id)
		if err != nil {
			return factory.ErrorParameter.New(err)
		}

		waitingAmount, err := domain.InvestmentWaiters{}.GetTotalAmountForWaitingStatus(ctx.Request().Context(), p.Id, user.Id)
		if err != nil {
			return factory.ErrorParameter.New(err)
		}

		totalAmount := amount + investedAmount + waitingAmount
		if totalAmount >= LIMITED_AMOUNT {
			return factory.ErrorOverLimitedAmount.New(nil)
		}
	}

	// 年齢確認
	if err := user.CheckOverAge(ctx.Request().Context()); err != nil {
		return factory.ErrorOverAgeInvestor.New(err)
	}

	// 購入数確認
	if !p.CheckShareQty(v.Qty, amount) {
		return factory.ErrorOverLimitedAmount.New(err)
	}

	// キャンセル待ち投資
	if p.IsWaitInvest {
		// キャンセル待ちモードへ移行済みの場合
		return factory.ErrorAlreadyWaiting.New(nil)
	}

	// projectのステータス確認
	if !p.IsInvestmentOpen(false) {
		return factory.ErrorNotOpenFunding.New(err)
	}

	if c.sqsClient == nil || c.queueURL == "" {
		return factory.ErrorSystem.New(fmt.Errorf("SQS client or queue URL is not configured"))
	}

	body, err := json.Marshal(v)
	if err != nil {
		return factory.ErrorSystem.New(fmt.Errorf("failed to marshal investment payload: %w", err))
	}

	input := &sqs.SendMessageInput{
		QueueUrl:    aws.String(c.queueURL),
		MessageBody: aws.String(string(body)),
	}
	if strings.HasSuffix(strings.ToLower(c.queueURL), ".fifo") {
		input.MessageGroupId = aws.String(fmt.Sprintf("project-%d", v.ProjectId))
		input.MessageDeduplicationId = aws.String(ksuid.New().String())
	}

	if _, err := c.sqsClient.SendMessage(ctx.Request().Context(), input); err != nil {
		return factory.ErrorSystem.New(fmt.Errorf("failed to send SQS message: %w", err))
	}

	userVirtualAccount, err := c.userVirtualAccountRepository.GetByUserIdAndLinkedStatus(ctx.Request().Context(), uid, p.ProjectType, domain.ON_VA_ACCOUNT)
	if err != nil {
		return factory.ErrorDB.New(err)
	}

	res := InvestProjectSharesResponse{
		ProjectName:       p.Name,
		BankName:          userVirtualAccount.VaBankName,
		BankAccountNumber: userVirtualAccount.VaAccountNumber,
		BranchName:        userVirtualAccount.VaBranchName,
		Depositor:         userVirtualAccount.VaHolderNameKana,
		Qty:               v.Qty,
		Price:             p.SharePrice,
		Amount:            v.Qty * p.SharePrice,
	}

	return renderSucc(ctx, http.StatusAccepted, res)
}

func validateInvestProjectSelectionType(project *domain.Project) error {
	if project == nil {
		return factory.ErrorParameter.New(errors.New("project is nil"))
	}
	if project.SelectionType != domain.PROJECT_SELECTION_FIRST_COME_FIRST_SERVED {
		return factory.ErrorParameter.New(errors.New("project selection type is not first come first served"))
	}
	return nil
}

func validateInvestmentWaitProject(project *domain.Project) error {
	if project == nil {
		return factory.ErrorParameter.New(errors.New("project is nil"))
	}
	if !project.IsWaitInvest {
		return factory.ErrorNotOpenFunding.New(errors.New("project is not waiting investment"))
	}
	if !project.IsInvestmentOpen(true) {
		return factory.ErrorNotOpenFunding.New(errors.New("Not OPEN_FUNDING"))
	}
	return nil
}

func validateInvestmentWaitApplication(ctx context.Context, user *domain.User, project *domain.Project) error {
	if user == nil {
		return factory.ErrorParameter.New(errors.New("user is nil"))
	}
	if user.BlockInvest {
		return factory.ErrorBlockInvest.New(nil)
	}
	if err := validateInvestmentWaitProject(project); err != nil {
		return err
	}
	if err := checkUserOverAge(ctx, user); err != nil {
		return factory.ErrorOverAgeInvestor.New(err)
	}
	return nil
}

func validateInvestmentWaitAmountLimit(ctx context.Context, userID int64, project *domain.Project, qty int64) error {
	if project == nil {
		return factory.ErrorParameter.New(errors.New("project is nil"))
	}
	if !project.IsAmountLimit {
		return nil
	}

	amount := project.SharePrice * qty
	investedAmount, err := getShareAmountByUserID(ctx, project.Id, userID)
	if err != nil {
		return factory.ErrorParameter.New(err)
	}

	waitingAmount, err := getInvestmentWaitersTotalAmount(ctx, project.Id, userID)
	if err != nil {
		return factory.ErrorParameter.New(err)
	}

	if amount+investedAmount+waitingAmount >= LIMITED_AMOUNT {
		return factory.ErrorOverLimitedAmount.New(nil)
	}
	return nil
}

// createInvestmentWaitApplication は同一ユーザーへの申込を直列化し、
// DBミドルウェアが開始した既存トランザクション上でロック取得・金額再集計・INSERT を行う。
// 新しいトランザクションは開始しない（ネストによるコネクション二重取得・部分コミットを防ぐ）。
func createInvestmentWaitApplication(ctx context.Context, user *domain.User, project *domain.Project, qty int64) (*domain.InvestmentWaiters, error) {
	if err := (domain.User{}).LockById(ctx, user.Id); err != nil {
		return nil, factory.ErrorDB.New(err)
	}
	if err := validateInvestmentWaitAmountLimit(ctx, user.Id, project, qty); err != nil {
		return nil, err
	}

	amount := project.SharePrice * qty
	created := domain.InvestmentWaiters{
		ProjectId:   project.Id,
		UserId:      user.Id,
		Qty:         qty,
		Amount:      amount,
		WaitingDate: jsonutils.JSONTime(time.Now()),
		Status:      domain.WAITING,
	}
	if _, err := created.Create(ctx); err != nil {
		return nil, factory.ErrorDB.New(err)
	}
	return &created, nil
}

// notifyInvestmentWaitCreated は申込後のイベント通知・メール送信を行う。
// 失敗しても申込自体の失敗にはせず、ログ出力のみとする（再送による二重登録を防ぐ）。
func notifyInvestmentWaitCreated(ctx context.Context, user *domain.User, investmentWait *domain.InvestmentWaiters, project *domain.Project) {
	if err := publishInvestmentWaitRequestEvent(ctx, user, investmentWait, project); err != nil {
		factory.Logger(ctx).WithError(err).Error("failed to publish investment wait request event")
	}

	if err := sendInvestStandbyMail(
		ctx,
		serviceutils.GetServiceInfoMail(),
		user.Email,
		user.Name,
		project.Name,
		jsonutils.JSONTime.ToString(investmentWait.WaitingDate),
		strconv.Itoa(int(investmentWait.Qty)),
		strconv.Itoa(int(investmentWait.Amount)),
	); err != nil {
		factory.Logger(ctx).WithError(err).Error("failed to send investment standby mail")
	}
}

// applyInvestmentWaitApplication は申込作成と commit 後通知の登録を行う。
// DB登録に成功すれば通知失敗時もエラーを返さない。
func applyInvestmentWaitApplication(ctx context.Context, user *domain.User, project *domain.Project, qty int64) (*domain.InvestmentWaiters, error) {
	investmentWait, err := createInvestmentWaitApplication(ctx, user, project, qty)
	if err != nil {
		return nil, err
	}
	if err := factory.AfterCommit(ctx, func(afterCommitCtx context.Context) error {
		notifyInvestmentWaitCreated(afterCommitCtx, user, investmentWait, project)
		return nil
	}); err != nil {
		factory.Logger(ctx).WithError(err).Error("failed to register investment wait notification")
	}
	return investmentWait, nil
}

func (ProjectApiController) CheckInvest(c echo.Context) error {
	uid, _, err := tokenToUser(c)
	if err != nil {
		return err
	}

	user, err := domain.User{}.GetById(c.Request().Context(), uid)
	if err != nil {
		return factory.ErrorDB.New(err)
	}

	if user.UserType == domain.PRIVATE {
		m, err := domain.MyNumber{}.CheckMyNumberStatusByUserId(c.Request().Context(), user.Id)
		if err != nil {
			return factory.ErrorDB.New(err)
		}
		if !m {
			return factory.ErrorMyNumberNotRegistered.New(nil)
		}
	}

	if user.BlockInvest {
		return factory.ErrorBlockInvest.New(err)
	}

	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return factory.ErrorParameter.New(err)
	}

	var v ShareCheckInput
	if err := c.Bind(&v); err != nil {
		return factory.ErrorParameter.New(err)
	}
	if err := c.Validate(&v); err != nil {
		return factory.ErrorParameter.New(err)
	}

	factory.Logger(c.Request().Context()).WithFields(logrus.Fields{
		"project_id":     v.ProjectId,
		"user_id":        uid,
		"qty":            v.Qty,
		"is_wait_invest": v.IsWaitInvest,
	}).Info("CheckInvest")

	p, err := domain.Project{}.GetById(c.Request().Context(), id)
	if err != nil {
		return factory.ErrorParameter.New(err)
	}

	if !p.IsInvestmentOpen(v.IsWaitInvest) {
		return factory.ErrorNotOpenFunding.New(errors.New("Not OPEN_FUNDING"))
	}

	amount := p.SharePrice * v.Qty

	if p.IsAmountLimit {

		investedAmount, err := domain.Share{}.GetAmountByUserId(c.Request().Context(), p.Id, user.Id)
		if err != nil {
			return factory.ErrorParameter.New(err)
		}

		waitingAmount, err := domain.InvestmentWaiters{}.GetTotalAmountForWaitingStatus(c.Request().Context(), p.Id, user.Id)
		if err != nil {
			return factory.ErrorParameter.New(err)
		}

		totalAmount := amount + investedAmount + waitingAmount
		if totalAmount >= LIMITED_AMOUNT {
			return factory.ErrorOverLimitedAmount.New(nil)
		}
	}

	if v.IsWaitInvest {
		// 투자대기 체크
		if p.InvestWaitRemainingQty > p.TargetShareQty {
			return factory.ErrorInsufficientQty.New(err)
		}
	} else {
		// 투자 체크
		if !p.CheckShareQty(v.Qty, amount) {
			return factory.ErrorInsufficientQty.New(err)
		}
	}
	// 구매자 나이 검사
	if err := user.CheckOverAge(c.Request().Context()); err != nil {
		return factory.ErrorOverAgeInvestor.New(err)
	}

	return renderSucc(c, http.StatusOK, nil)
}

func (ProjectApiController) GetMainListProjects(c echo.Context) error {
	var v ProjectListSearchInput
	if err := c.Bind(&v); err != nil {
		return factory.ErrorParameter.New(err)
	}
	if v.MaxResultCount == 0 {
		v.MaxResultCount = DefaultMaxResultCount
	}

	// console log
	factory.Logger(c.Request().Context()).WithFields(logrus.Fields{
		"listType":         v.ListType,
		"max_result_count": v.MaxResultCount,
		"skip_count":       v.SkipCount,
		"sortby":           v.Sortby,
		"order":            v.Order,
	}).Info("SearchStart")

	totalCount, items, err := domain.Project{}.GetByProjectType(c.Request().Context(), v.ListType,
		v.SkipCount, v.MaxResultCount, v.Sortby, v.Order, v.ProjectName, v.StartDate, v.EndDate)
	if err != nil {
		return factory.ErrorDB.New(err)
	}
	if err := applyCurrentShareAmountForUserView(c.Request().Context(), items); err != nil {
		return factory.ErrorDB.New(err)
	}

	return renderSucc(c, http.StatusOK, ArrayResult{
		TotalCount: totalCount,
		Items:      ProjectItemsResponseDto{}.NewList(items),
	})
}

func (ProjectApiController) GetCategorys(c echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return factory.ErrorParameter.New(err)
	}

	v, err := domain.Project{}.GetById(c.Request().Context(), id)
	if err != nil {
		return factory.ErrorParameter.New(err)
	}
	if v == nil {
		return factory.ErrorNotFound.New(err)
	}

	catalog, err := domain.Project{}.CreateProjectCatalogPDF(c.Request().Context(), v)
	if err != nil {
		return factory.ErrorFailDownloadPDF.New(err)
	}
	w := c.Response().Writer
	w.Header().Set("Content-Type", "application/octet-stream")
	w.Header().Set("Content-Disposition", "attachment; filename="+v.Name+".pdf")
	w.Header().Set("File-Name", v.Name+".pdf")
	w.Header().Set("Content-Transfer-Encoding", "binary")
	w.Header().Set("Expires", "0")

	if _, err := w.Write(catalog.Bytes()); err != nil {
		return factory.ErrorFailDownloadPDF.New(err)
	}
	return renderSucc(c, http.StatusOK, catalog)
}

func (ProjectApiController) InvestmentWaitCreate(c echo.Context) error {
	uid, _, err := tokenToUser(c)
	if err != nil {
		return err
	}
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return factory.ErrorParameter.New(err)
	}

	var v InvestmentWaitersInput

	if err := c.Bind(&v); err != nil {
		return factory.ErrorParameter.New(err)
	}
	if err := c.Validate(&v); err != nil {
		return factory.ErrorParameter.New(err)
	}

	user, err := getUserByIDForInvestmentWait(c.Request().Context(), uid)
	if err != nil {
		return factory.ErrorParameter.New(err)
	}

	if user.UserType == domain.PRIVATE {
		m, err := domain.MyNumber{}.CheckMyNumberStatusByUserId(c.Request().Context(), user.Id)
		if err != nil {
			return factory.ErrorDB.New(err)
		}
		if !m {
			return factory.ErrorMyNumberNotRegistered.New(nil)
		}
	}

	project, err := getProjectByIDForInvestmentWait(c.Request().Context(), id)
	if err != nil {
		return factory.ErrorParameter.New(err)
	}

	if err := validateInvestmentWaitApplication(c.Request().Context(), user, project); err != nil {
		return err
	}

	investmentWait, err := applyInvestmentWaitApplication(c.Request().Context(), user, project, v.Qty)
	if err != nil {
		return err
	}

	return renderSucc(c, http.StatusCreated, investmentWait)
}

func (ProjectApiController) InvestmentRequestCheck(c echo.Context) error {
	_, email, err := tokenToUser(c)
	if err != nil {
		return err
	}
	requestId := c.Param("request_id")

	response, err := domain.InvestmentRequest{}.GetByRequestId(c.Request().Context(), requestId)
	if err != nil {
		return err
	}

	// API를 요청한 유저와 투자제안
	if email != response.Email {
		return factory.ErrorUserNotExists.New(err)
	}

	// 투자대기자 정보를 가져옵니다.
	investmentWaiters, err := domain.InvestmentWaiters{}.GetById(c.Request().Context(), response.InvestmentWaitersId)
	if err != nil {
		return err
	}

	// 투자제안 이후에 사용자가 실제로 투자를 진행한경우 (입금여부 상관 x)
	if investmentWaiters.ShareId != "" {
		response.IsExpire = true
	}

	// 투자제안일자가 오늘날짜와 같은지 확인하고 같지 않다면 isExpired를 true로 변경
	if !response.Created.EqualDate(time.Now()) {
		response.IsExpire = true
	}

	return renderSucc(c, http.StatusOK, response)
}

func (c ProjectApiController) ProposedInvestProjectShares(ctx echo.Context) error {
	var v ShareInvestInput

	if err := ctx.Bind(&v); err != nil {
		return factory.ErrorParameter.New(err)
	}
	if err := ctx.Validate(&v); err != nil {
		return factory.ErrorParameter.New(err)
	}

	uid, _, err := tokenToUser(ctx)
	if err != nil {
		return err
	}

	requestId := ctx.Param("request_id")

	investmentRequest, err := domain.InvestmentRequest{}.GetByRequestId(ctx.Request().Context(), requestId)
	if err != nil {
		return err
	}

	id, err := strconv.ParseInt(ctx.Param("id"), 10, 64)
	if err != nil {
		return factory.ErrorParameter.New(err)
	}
	project, err := domain.Project{}.GetById(ctx.Request().Context(), id)
	if err != nil {
		return factory.ErrorDB.New(err)
	}

	user, err := domain.User{}.GetById(ctx.Request().Context(), uid)
	if err != nil {
		return factory.ErrorNotFound.New(err)
	}

	investmentWaiters, err := domain.InvestmentWaiters{}.GetById(ctx.Request().Context(), investmentRequest.InvestmentWaitersId)
	if err != nil {
		return factory.ErrorDB.New(err)
	}

	if investmentWaiters == nil {
		return factory.ErrorNotFound.New(err)
	}

	if project.Id != investmentWaiters.ProjectId {
		return factory.ErrorParameter.New(err)
	}

	share, err := service.InvestmentService{}.WaitInvest(ctx.Request().Context(), user, investmentWaiters, v.Qty, true)
	if err != nil {
		return err
	}

	uva, err := c.userVirtualAccountRepository.GetByUserIdAndLinkedStatus(ctx.Request().Context(), uid, project.ProjectType, domain.ON_VA_ACCOUNT)
	if err != nil {
		return factory.ErrorDB.New(err)
	}

	res := InvestProjectSharesResponse{
		ProjectName:       project.Name,
		BankName:          uva.VaBankName,
		BankAccountNumber: uva.VaAccountNumber,
		BranchName:        uva.VaBranchName,
		Depositor:         uva.VaHolderNameKana,
		Qty:               share.Qty,
		Price:             share.Price,
		Amount:            share.Amount,
	}

	return renderSucc(ctx, http.StatusCreated, res)
}

func (ProjectApiController) GetOnePreview(c echo.Context) error {
	userId, err := strconv.ParseInt(c.Param("user_id"), 10, 64)
	if err != nil {
		return factory.ErrorParameter.New(err)
	}

	d, err := domain.AdminUser{}.GetById(c.Request().Context(), userId)
	if err != nil {
		return factory.ErrorDB.New(err)
	}
	if d == nil {
		return factory.ErrorNotFound.New(err)
	}

	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return factory.ErrorParameter.New(err)
	}

	v, err := domain.Project{}.GetByIdForPublic(c.Request().Context(), id, true)
	if err != nil {
		return factory.ErrorDB.New(err)
	}
	if v == nil {
		return factory.ErrorNotFound.New(nil)
	}

	dto := ProjectDetailResponseDto{}.ToProjectPubResponseDto(v)

	return renderSucc(c, http.StatusOK, dto)
}

func applyCurrentShareAmountForUserView(ctx context.Context, projects []domain.Project) error {
	if len(projects) == 0 {
		return nil
	}

	projectIds := make([]int64, 0, len(projects))
	for i := range projects {
		if projects[i].SelectionType != domain.PROJECT_SELECTION_LOTTERY {
			projectIds = append(projectIds, projects[i].Id)
		}
	}

	amounts, err := domain.Project{}.GetCurrentShareAmountsForUserView(ctx, projectIds)
	if err != nil {
		return err
	}

	for i := range projects {
		if projects[i].SelectionType == domain.PROJECT_SELECTION_LOTTERY {
			// 抽選では final_funding_amount を申込時に加算しているのでそれをそのまま使う
			projects[i].CurrentShareAmount = projects[i].FinalFundingAmount
		} else if projects[i].ProjectStatus == domain.OPEN_FUNDING || projects[i].ProjectStatus == domain.STANDBY_FUNDING {
			// 募集中は最新の申込金額を常に参照する
			if amount, ok := amounts[projects[i].Id]; ok {
				projects[i].CurrentShareAmount = amount
			} else {
				projects[i].CurrentShareAmount = 0
			}
		} else if projects[i].FinalFundingAmount > 0 {
			// 募集終了時の申込金額が設定されている場合は優先して設定
			projects[i].CurrentShareAmount = projects[i].FinalFundingAmount
		} else {
			// 募集終了後に申込金額が確定していない場合は最新値を参照
			if amount, ok := amounts[projects[i].Id]; ok {
				projects[i].CurrentShareAmount = amount
			} else {
				projects[i].CurrentShareAmount = 0
			}
		}
	}

	return nil
}

func getCurrentShareAmountForUserView(ctx context.Context, projectId int64) (int64, error) {
	amounts, err := domain.Project{}.GetCurrentShareAmountsForUserView(ctx, []int64{projectId})
	if err != nil {
		return 0, err
	}
	return amounts[projectId], nil
}

func (ProjectApiController) GetOneClosed(c echo.Context) error {
	slug := c.Param("slug")
	v, err := domain.Project{}.GetByClosedSlugForClosed(c.Request().Context(), slug, false)
	if err != nil {
		return factory.ErrorDB.New(err)
	}
	if v == nil {
		return factory.ErrorNotFound.New(nil)
	}
	if v.SelectionType == domain.PROJECT_SELECTION_LOTTERY {
		// 抽選では final_funding_amount を申込時に加算しているのでそれをそのまま使う
		v.CurrentShareAmount = v.FinalFundingAmount
	} else if v.ProjectStatus == domain.OPEN_FUNDING || v.ProjectStatus == domain.STANDBY_FUNDING {
		// 募集中は最新の申込金額を常に参照する
		currentShareAmount, err := getCurrentShareAmountForUserView(c.Request().Context(), v.Id)
		if err != nil {
			return factory.ErrorDB.New(err)
		}
		v.CurrentShareAmount = currentShareAmount
	} else if v.FinalFundingAmount > 0 {
		// 募集終了時の申込金額が設定されている場合は優先して設定
		v.CurrentShareAmount = v.FinalFundingAmount
	} else {
		// 募集終了後に申込金額が確定していない場合は最新値を参照
		currentShareAmount, err := getCurrentShareAmountForUserView(c.Request().Context(), v.Id)
		if err != nil {
			return factory.ErrorDB.New(err)
		}
		v.CurrentShareAmount = currentShareAmount
	}

	return renderSucc(c, http.StatusOK, ProjectDetailResponseDto{}.ToProjectPubResponseDto(v))
}
