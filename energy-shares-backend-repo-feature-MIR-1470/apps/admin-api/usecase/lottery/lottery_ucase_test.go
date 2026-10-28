package lottery

import (
	"bytes"
	"context"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/pkg/errors"
	"github.com/samber/lo"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"xorm.io/xorm"
	"xorm.io/xorm/names"

	adomain "github.com/illuminarean/energy-share-core-api/domain"
	cdomain "github.com/illuminarean/energy-shares-backend-repo/clean-arch-domain"
	"github.com/illuminarean/energy-shares-backend-repo/clean-arch-domain/mocks"
	"github.com/illuminarean/energy-shares-backend-repo/domain"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/actorcontext"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/config"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/factory"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/jsonutils"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/jwthelper"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/serviceutils"
)

func newTestMultipartFileHeader(t *testing.T, fieldName, filename, content string) *multipart.FileHeader {
	t.Helper()

	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	part, err := writer.CreateFormFile(fieldName, filename)
	assert.NoError(t, err)
	_, err = part.Write([]byte(content))
	assert.NoError(t, err)
	assert.NoError(t, writer.Close())

	req := httptest.NewRequest(http.MethodPost, "/", &body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	assert.NoError(t, req.ParseMultipartForm(int64(body.Len())+1024))

	fileHeaders := req.MultipartForm.File[fieldName]
	if assert.Len(t, fileHeaders, 1) {
		return fileHeaders[0]
	}
	return nil
}

func newLotteryUploadTestDB(t *testing.T) (*xorm.Engine, context.Context) {
	t.Helper()

	engine, err := xorm.NewEngine("sqlite3", "file::memory:?cache=shared")
	assert.NoError(t, err)
	engine.SetMapper(names.LintGonicMapper)
	t.Cleanup(func() {
		engine.Close()
	})

	_, err = engine.Exec(`
CREATE TABLE lottery_csv_upload_job (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    job_id TEXT NOT NULL,
    project_id INTEGER NOT NULL,
    requested_by TEXT NOT NULL,
    file_name TEXT NOT NULL,
    status TEXT NOT NULL,
    total_row_count INTEGER NOT NULL DEFAULT 0,
    success_row_count INTEGER NOT NULL DEFAULT 0,
    failed_row_count INTEGER NOT NULL DEFAULT 0,
    started_at DATETIME NULL,
    finished_at DATETIME NULL,
    error_message TEXT NULL,
    created DATETIME NULL,
    updated DATETIME NULL,
    deleted DATETIME NULL
);
CREATE TABLE lottery_csv_upload_job_detail (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    job_id INTEGER NOT NULL,
    row_number INTEGER NOT NULL,
    lottery_application_id TEXT NOT NULL,
    expected_lottery_status TEXT NULL,
    expected_applied_at TEXT NULL,
    input_applied_qty INTEGER NULL,
    input_winning_qty INTEGER NOT NULL,
    process_type TEXT NOT NULL,
    status TEXT NOT NULL,
    error_code TEXT NULL,
    error_message TEXT NULL,
    processed_at DATETIME NULL,
    created_share_id TEXT NULL,
    created_deposit_request_id TEXT NULL,
    created DATETIME NULL,
    updated DATETIME NULL,
    deleted DATETIME NULL
);`)
	assert.NoError(t, err)

	ctx := context.WithValue(context.Background(), factory.ConextDBType, engine)
	return engine, ctx
}

func Test_validateWinningProject(t *testing.T) {
	tests := []struct {
		name    string
		project *domain.Project
		wantErr error
	}{
		{
			name: "正常系",
			project: &domain.Project{
				SelectionType: domain.PROJECT_SELECTION_LOTTERY,
				ProjectStatus: domain.END_FUNDING,
			},
			wantErr: nil,
		},
		{
			name:    "projectがnil",
			project: nil,
			wantErr: cdomain.ErrItemNotFound,
		},
		{
			name: "抽選方式でない",
			project: &domain.Project{
				SelectionType: domain.PROJECT_SELECTION_FIRST_COME_FIRST_SERVED,
				ProjectStatus: domain.END_FUNDING,
			},
			wantErr: cdomain.ErrBadRequest,
		},
		{
			name: "募集終了でない",
			project: &domain.Project{
				SelectionType: domain.PROJECT_SELECTION_LOTTERY,
				ProjectStatus: domain.OPEN_FUNDING,
			},
			wantErr: cdomain.ErrBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateWinningProject(tt.project)
			if tt.wantErr == nil {
				assert.NoError(t, err)
				return
			}
			assert.Error(t, err)
			assert.True(t, errors.Is(err, tt.wantErr))
		})
	}
}

func Test_validateWinningApplication(t *testing.T) {
	tests := []struct {
		name        string
		application domain.LotteryApplication
		qty         int64
		wantErr     error
	}{
		{
			name: "正常系",
			application: domain.LotteryApplication{
				Qty: 10,
				Winnings: []domain.Share{
					{Qty: 3},
				},
			},
			qty:     2,
			wantErr: nil,
		},
		{
			name: "キャンセル済み",
			application: domain.LotteryApplication{
				Qty:        10,
				CanceledAt: jsonutils.JSONTime(time.Now()),
			},
			qty:     1,
			wantErr: cdomain.ErrBadRequest,
		},
		{
			name: "qtyが0以下",
			application: domain.LotteryApplication{
				Qty: 10,
			},
			qty:     0,
			wantErr: cdomain.ErrBadRequest,
		},
		{
			name: "申込口数超過",
			application: domain.LotteryApplication{
				Qty: 10,
			},
			qty:     11,
			wantErr: cdomain.ErrBadRequest,
		},
		{
			name: "既当選口数と合算して超過",
			application: domain.LotteryApplication{
				Qty: 10,
				Winnings: []domain.Share{
					{Qty: 8},
				},
			},
			qty:     3,
			wantErr: cdomain.ErrBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateWinningApplication(tt.application, tt.qty)
			if tt.wantErr == nil {
				assert.NoError(t, err)
				return
			}
			assert.Error(t, err)
			assert.True(t, errors.Is(err, tt.wantErr))
		})
	}
}

func Test_validateLosingApplication(t *testing.T) {
	tests := []struct {
		name        string
		application domain.LotteryApplication
		wantErr     error
	}{
		{
			name: "正常系",
			application: domain.LotteryApplication{
				Qty: 10,
				Winnings: []domain.Share{
					{Qty: 3},
				},
			},
			wantErr: nil,
		},
		{
			name: "キャンセル済み",
			application: domain.LotteryApplication{
				Qty:        10,
				CanceledAt: jsonutils.JSONTime(time.Now()),
			},
			wantErr: cdomain.ErrBadRequest,
		},
		{
			name: "全量当選済みで未当選残口数がない",
			application: domain.LotteryApplication{
				Qty: 10,
				Winnings: []domain.Share{
					{Qty: 10},
				},
			},
			wantErr: cdomain.ErrBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateLosingApplication(tt.application)
			if tt.wantErr == nil {
				assert.NoError(t, err)
				return
			}
			assert.Error(t, err)
			assert.True(t, errors.Is(err, tt.wantErr))
		})
	}
}

func Test_validateWinningUser(t *testing.T) {
	tests := []struct {
		name    string
		user    *domain.User
		wantErr error
	}{
		{
			name: "正常系",
			user: &domain.User{
				AuditStatus: domain.COMPLETED_AUDIT,
			},
			wantErr: nil,
		},
		{
			name:    "userがnil",
			user:    nil,
			wantErr: cdomain.ErrItemNotFound,
		},
		{
			name: "BlockInvest",
			user: &domain.User{
				BlockInvest: true,
				AuditStatus: domain.COMPLETED_AUDIT,
			},
			wantErr: cdomain.ErrBadRequest,
		},
		{
			name: "本人確認未完了",
			user: &domain.User{
				AuditStatus: domain.WAIT_AUDIT,
			},
			wantErr: cdomain.ErrBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateWinningUser(tt.user)
			if tt.wantErr == nil {
				assert.NoError(t, err)
				return
			}
			assert.Error(t, err)
			assert.True(t, errors.Is(err, tt.wantErr))
		})
	}
}

func Test_winningQtyTotal(t *testing.T) {
	got := winningQtyTotal(domain.LotteryApplication{
		Winnings: []domain.Share{
			{Qty: 2},
			{Qty: 3},
			{Qty: 5},
		},
	})

	assert.Equal(t, int64(10), got)
}

func TestLotteryUsecase_GetAll_MapsTimestamps(t *testing.T) {
	ctx := context.Background()
	appliedAt := jsonutils.JSONTime(time.Date(2026, 4, 10, 1, 2, 3, 0, time.UTC))
	winningAt := jsonutils.JSONTime(time.Date(2026, 4, 11, 1, 2, 3, 0, time.UTC))
	lostAt := jsonutils.JSONTime(time.Date(2026, 4, 12, 1, 2, 3, 0, time.UTC))
	canceledAt := jsonutils.JSONTime(time.Date(2026, 4, 13, 1, 2, 3, 0, time.UTC))

	lotteryRepo := &mocks.LotteryApplicationRepository{}
	lotteryRepo.On(
		"GetAll",
		mock.Anything,
		mock.Anything,
		mock.Anything,
		mock.Anything,
		mock.Anything,
		mock.Anything,
		mock.Anything,
		mock.Anything,
		mock.Anything,
		mock.Anything,
		mock.Anything,
		mock.Anything,
		mock.Anything,
		mock.Anything,
	).Return(int64(1), []domain.LotteryApplication{
		{
			Id:            "app-001",
			UserId:        1,
			UserName:      "user-1",
			ProjectId:     10,
			ProjectName:   "project-10",
			ProjectStatus: domain.OPEN_FUNDING,
			Qty:           5,
			UnitPrice:     2000,
			Amount:        10000,
			Status:        domain.LOTTERY_STATUS_CANCELED,
			AppliedAt:     appliedAt,
			WinningAt:     winningAt,
			LostAt:        lostAt,
			CanceledAt:    canceledAt,
			Winnings: []domain.Share{
				{Amount: 3000, Qty: 3, Price: 1000, InvestStatus: domain.REQUEST_INVEST},
				{Amount: 2000, Qty: 2, Price: 1000, InvestStatus: domain.CANCEL_INVEST},
			},
		},
	}, nil)
	lotteryRepo.On("GetWinningQtyTotalsByProjectID", mock.Anything, []int64{int64(10)}).
		Return(map[int64]int64{10: 0}, nil)

	u := lotteryUsecase{
		lotteryRepository:         lotteryRepo,
		lotteryHistoryRepository:  &mocks.LotteryHistoryRepository{},
		projectRepository:         &mocks.ProjectRepository{},
		operatorHistoryRepository: &mocks.OperatorHistoryRepository{},
		txManager:                 newLotteryCancelTxManagerMock(),
	}

	got, err := u.GetAll(ctx, adomain.LotteryGetAllRequest{})

	assert.NoError(t, err)
	if assert.Len(t, got.Items, 1) {
		assert.Equal(t, appliedAt, got.Items[0].AppliedAt)
		assert.Equal(t, winningAt, got.Items[0].WinningAt)
		assert.Equal(t, lostAt, got.Items[0].LostAt)
		assert.Equal(t, canceledAt, got.Items[0].CanceledAt)
		assert.Equal(t, int64(2000), got.Items[0].CanceledWinningAmount)
	}
}

func TestLotteryUsecase_GetAll_MapsUserType(t *testing.T) {
	tests := []struct {
		name     string
		userType domain.USER_TYPE
	}{
		{
			name:     "PRIVATE をそのまま返す",
			userType: domain.PRIVATE,
		},
		{
			name:     "CORPORATE をそのまま返す",
			userType: domain.CORPORATE,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()

			lotteryRepo := &mocks.LotteryApplicationRepository{}
			lotteryRepo.On(
				"GetAll",
				mock.Anything,
				mock.Anything,
				mock.Anything,
				mock.Anything,
				mock.Anything,
				mock.Anything,
				mock.Anything,
				mock.Anything,
				mock.Anything,
				mock.Anything,
				mock.Anything,
				mock.Anything,
				mock.Anything,
				mock.Anything,
			).Return(int64(1), []domain.LotteryApplication{
				{
					Id:            "app-001",
					UserId:        1,
					UserType:      tt.userType,
					UserName:      "user-1",
					ProjectId:     10,
					ProjectName:   "project-10",
					ProjectStatus: domain.OPEN_FUNDING,
					Qty:           5,
					UnitPrice:     2000,
					Amount:        10000,
					Status:        domain.LOTTERY_STATUS_APPLIED,
				},
			}, nil)
			lotteryRepo.On("GetWinningQtyTotalsByProjectID", mock.Anything, []int64{int64(10)}).
				Return(map[int64]int64{10: 0}, nil)

			u := lotteryUsecase{
				lotteryRepository:         lotteryRepo,
				lotteryHistoryRepository:  &mocks.LotteryHistoryRepository{},
				projectRepository:         &mocks.ProjectRepository{},
				operatorHistoryRepository: &mocks.OperatorHistoryRepository{},
				txManager:                 newLotteryCancelTxManagerMock(),
			}

			got, err := u.GetAll(ctx, adomain.LotteryGetAllRequest{})

			assert.NoError(t, err)
			if assert.Len(t, got.Items, 1) {
				assert.Equal(t, tt.userType, got.Items[0].UserType)
				assert.Equal(t, int64(1), got.TotalCount)
				assert.Equal(t, int64(10000), got.TotalAmount)
				assert.Equal(t, int64(0), got.Items[0].CanceledWinningAmount)
				assert.Equal(t, domain.OPEN_FUNDING, got.Items[0].ProjectStatus)
				assert.Equal(t, domain.LOTTERY_STATUS_APPLIED, got.Items[0].Status)
			}
		})
	}
}

func newLotteryCancelTxManagerMock() *mocks.TxManager {
	txManager := &mocks.TxManager{}
	txManager.On("WithTx", mock.Anything, mock.Anything).
		Return(func(ctx context.Context, fn func(context.Context) error) error {
			return fn(ctx)
		})
	return txManager
}

func TestLotteryUsecase_Cancel(t *testing.T) {
	type args struct {
		ctx     context.Context
		request adomain.LotteryCancelRequest
	}

	tests := []struct {
		name                string
		args                args
		lotteryRepository   func() *mocks.LotteryApplicationRepository
		projectRepository   func() *mocks.ProjectRepository
		userRepository      func() *mocks.UserRepository
		mailService         func(t *testing.T) *mocks.MailService
		lotteryHistoryRepo  func() *mocks.LotteryHistoryRepository
		operatorHistoryRepo func() *mocks.OperatorHistoryRepository
		want                adomain.LotteryCancelResponse
		wantErr             bool
	}{
		{
			name: "(成功) 抽選申込キャンセル",
			args: args{
				ctx: actorcontext.WithActor(context.Background(), jwthelper.User{Id: 999}),
				request: adomain.LotteryCancelRequest{
					JWTUser:              jwthelper.User{Id: 999},
					LotteryApplicationID: "app-001",
				},
			},
			lotteryRepository: func() *mocks.LotteryApplicationRepository {
				repo := &mocks.LotteryApplicationRepository{}
				repo.On("LockById", mock.Anything, "app-001").
					Return(domain.LotteryApplication{Id: "app-001", ProjectId: 1, Amount: 10000}, nil)
				repo.On("MarkAsCanceled", mock.Anything, "app-001").
					Return(true, nil)
				repo.On("GetById", mock.Anything, "app-001").
					Return(domain.LotteryApplication{
						Id:         "app-001",
						ProjectId:  1,
						UserId:     999,
						Amount:     10000,
						CanceledAt: jsonutils.JSONTime(time.Date(2026, 4, 27, 12, 34, 56, 0, time.UTC)),
					}, nil)
				return repo
			},
			projectRepository: func() *mocks.ProjectRepository {
				repo := &mocks.ProjectRepository{}
				repo.On("LockById", mock.Anything, int64(1)).
					Return(domain.Project{Id: 1, ProjectStatus: domain.OPEN_FUNDING}, nil)
				repo.On("ReduceFinalFundingAmount", mock.Anything, int64(1), int64(10000)).
					Return(nil)
				repo.On("GetById", mock.Anything, int64(1)).
					Return(domain.Project{Id: 1, Name: "Lottery Fund", ProjectStatus: domain.OPEN_FUNDING}, nil)
				return repo
			},
			userRepository: func() *mocks.UserRepository {
				repo := &mocks.UserRepository{}
				repo.On("GetById", mock.Anything, int64(999)).
					Return(domain.User{Id: 999, Name: "Lottery User", Email: "lottery@example.com"}, nil)
				return repo
			},
			mailService: func(t *testing.T) *mocks.MailService {
				service := mocks.NewMailService(t)
				service.On("SendMail", mock.Anything, mock.MatchedBy(func(request cdomain.SendMailRequest) bool {
					return cmp.Equal(request, cdomain.SendMailRequest{
						TemplateType: domain.LOTTERY_APPLICATION_CANCEL_TEMPLATE,
						MailType:     domain.MAIL_SEND_PRIORITY_NORMAL,
						Sender:       "info@example.com",
						Receivers:    []string{"lottery@example.com"},
						Details: map[string]interface{}{
							"ProjectName":       "Lottery Fund",
							"Receiver":          "Lottery User",
							"ApplicationAmount": "10,000",
							"CancelTime":        "2026-04-27 21:34:56",
							"ServiceUrl":        "https://user.example.com",
						},
					})
				})).Return(nil)
				return service
			},
			lotteryHistoryRepo: func() *mocks.LotteryHistoryRepository {
				repo := &mocks.LotteryHistoryRepository{}
				repo.On("Create", mock.Anything, domain.LotteryHistory{
					ApplicationId: "app-001",
					PrevStatus:    domain.LOTTERY_STATUS_APPLIED,
					NewStatus:     domain.LOTTERY_STATUS_CANCELED,
					ActorId:       999,
				}).Return(nil)
				return repo
			},
			operatorHistoryRepo: func() *mocks.OperatorHistoryRepository {
				repo := &mocks.OperatorHistoryRepository{}
				repo.On("Create", mock.Anything, domain.OperatorHistory{
					UserId: 999,
					Menu:   "menu.manage_accounting.project_lottery.list",
					Ip:     "",
					Title:  "lottery cancel application_id=app-001",
					State:  domain.CHANGE,
				}).Return(nil)
				return repo
			},
			want:    adomain.LotteryCancelResponse{},
			wantErr: false,
		},
		{
			name: "(失敗) 抽選申込キャンセル_履歴インサート失敗",
			args: args{
				ctx: actorcontext.WithActor(context.Background(), jwthelper.User{Id: 999}),
				request: adomain.LotteryCancelRequest{
					JWTUser:              jwthelper.User{Id: 999},
					LotteryApplicationID: "app-001",
				},
			},
			lotteryRepository: func() *mocks.LotteryApplicationRepository {
				repo := &mocks.LotteryApplicationRepository{}
				repo.On("LockById", mock.Anything, "app-001").
					Return(domain.LotteryApplication{Id: "app-001", ProjectId: 1}, nil)
				repo.On("MarkAsCanceled", mock.Anything, "app-001").
					Return(true, nil)
				return repo
			},
			projectRepository: func() *mocks.ProjectRepository {
				repo := &mocks.ProjectRepository{}
				repo.On("LockById", mock.Anything, int64(1)).
					Return(domain.Project{Id: 1, ProjectStatus: domain.OPEN_FUNDING}, nil)
				repo.On("ReduceFinalFundingAmount", mock.Anything, int64(1), int64(0)).
					Return(nil)
				return repo
			},
			userRepository: func() *mocks.UserRepository {
				return &mocks.UserRepository{}
			},
			mailService: func(t *testing.T) *mocks.MailService {
				return mocks.NewMailService(t)
			},
			lotteryHistoryRepo: func() *mocks.LotteryHistoryRepository {
				repo := &mocks.LotteryHistoryRepository{}
				repo.On("Create", mock.Anything, domain.LotteryHistory{
					ApplicationId: "app-001",
					PrevStatus:    domain.LOTTERY_STATUS_APPLIED,
					NewStatus:     domain.LOTTERY_STATUS_CANCELED,
					ActorId:       999,
				}).Return(errors.New("insert failed"))
				return repo
			},
			operatorHistoryRepo: func() *mocks.OperatorHistoryRepository {
				repo := &mocks.OperatorHistoryRepository{}
				repo.On("Create", mock.Anything, domain.OperatorHistory{
					UserId: 999,
					Menu:   "menu.manage_accounting.project_lottery.list",
					Ip:     "",
					Title:  "lottery cancel application_id=app-001",
					State:  domain.CHANGE,
				}).Return(nil)
				return repo
			},
			want:    adomain.LotteryCancelResponse{},
			wantErr: true,
		},
		{
			name: "(失敗) 抽選申込キャンセル_対象の抽選申込が見つからない",
			args: args{
				ctx: actorcontext.WithActor(context.Background(), jwthelper.User{Id: 999}),
				request: adomain.LotteryCancelRequest{
					LotteryApplicationID: "app-001",
				},
			},
			lotteryRepository: func() *mocks.LotteryApplicationRepository {
				repo := &mocks.LotteryApplicationRepository{}
				repo.On("LockById", mock.Anything, "app-001").
					Return(domain.LotteryApplication{}, errors.New("not found"))
				return repo
			},
			projectRepository: func() *mocks.ProjectRepository {
				repo := &mocks.ProjectRepository{}
				repo.On("LockById", mock.Anything, int64(1)).
					Return(domain.Project{Id: 1, ProjectStatus: domain.OPEN_FUNDING}, nil)
				return repo
			},
			userRepository: func() *mocks.UserRepository {
				return &mocks.UserRepository{}
			},
			mailService: func(t *testing.T) *mocks.MailService {
				return mocks.NewMailService(t)
			},
			lotteryHistoryRepo: func() *mocks.LotteryHistoryRepository {
				return &mocks.LotteryHistoryRepository{}
			},
			operatorHistoryRepo: func() *mocks.OperatorHistoryRepository {
				return &mocks.OperatorHistoryRepository{}
			},
			want:    adomain.LotteryCancelResponse{},
			wantErr: true,
		},
		{
			name: "(失敗) 抽選申込キャンセル_当選済みはキャンセル不可",
			args: args{
				ctx: actorcontext.WithActor(context.Background(), jwthelper.User{Id: 999}),
				request: adomain.LotteryCancelRequest{
					LotteryApplicationID: "app-002",
				},
			},
			lotteryRepository: func() *mocks.LotteryApplicationRepository {
				repo := &mocks.LotteryApplicationRepository{}
				repo.On("LockById", mock.Anything, "app-002").
					Return(domain.LotteryApplication{
						Id:        "app-002",
						ProjectId: 1,
						Status:    domain.LOTTERY_STATUS_WINNING,
						WinningAt: jsonutils.JSONTime(time.Date(2026, 4, 14, 10, 0, 0, 0, time.UTC)),
					}, nil)
				return repo
			},
			projectRepository: func() *mocks.ProjectRepository {
				repo := &mocks.ProjectRepository{}
				repo.On("LockById", mock.Anything, int64(1)).
					Return(domain.Project{Id: 1, ProjectStatus: domain.OPEN_FUNDING}, nil)
				return repo
			},
			userRepository: func() *mocks.UserRepository {
				return &mocks.UserRepository{}
			},
			mailService: func(t *testing.T) *mocks.MailService {
				return mocks.NewMailService(t)
			},
			lotteryHistoryRepo: func() *mocks.LotteryHistoryRepository {
				return &mocks.LotteryHistoryRepository{}
			},
			operatorHistoryRepo: func() *mocks.OperatorHistoryRepository {
				return &mocks.OperatorHistoryRepository{}
			},
			want:    adomain.LotteryCancelResponse{},
			wantErr: true,
		},
		{
			name: "(失敗) 抽選申込キャンセル_キャンセル済みはキャンセル不可",
			args: args{
				ctx: actorcontext.WithActor(context.Background(), jwthelper.User{Id: 999}),
				request: adomain.LotteryCancelRequest{
					LotteryApplicationID: "app-005",
				},
			},
			lotteryRepository: func() *mocks.LotteryApplicationRepository {
				repo := &mocks.LotteryApplicationRepository{}
				repo.On("LockById", mock.Anything, "app-005").
					Return(domain.LotteryApplication{
						Id:         "app-005",
						ProjectId:  1,
						Status:     domain.LOTTERY_STATUS_CANCELED,
						CanceledAt: jsonutils.JSONTime(time.Date(2026, 4, 14, 10, 0, 0, 0, time.UTC)),
					}, nil)
				return repo
			},
			projectRepository: func() *mocks.ProjectRepository {
				repo := &mocks.ProjectRepository{}
				repo.On("LockById", mock.Anything, int64(1)).
					Return(domain.Project{Id: 1, ProjectStatus: domain.OPEN_FUNDING}, nil)
				return repo
			},
			userRepository: func() *mocks.UserRepository {
				return &mocks.UserRepository{}
			},
			mailService: func(t *testing.T) *mocks.MailService {
				return mocks.NewMailService(t)
			},
			lotteryHistoryRepo: func() *mocks.LotteryHistoryRepository {
				return &mocks.LotteryHistoryRepository{}
			},
			operatorHistoryRepo: func() *mocks.OperatorHistoryRepository {
				return &mocks.OperatorHistoryRepository{}
			},
			want:    adomain.LotteryCancelResponse{},
			wantErr: true,
		},
		{
			name: "(失敗) 抽選申込キャンセル_actor が context に無い",
			args: args{
				ctx: context.Background(),
				request: adomain.LotteryCancelRequest{
					LotteryApplicationID: "app-001",
				},
			},
			lotteryRepository: func() *mocks.LotteryApplicationRepository {
				return &mocks.LotteryApplicationRepository{}
			},
			projectRepository: func() *mocks.ProjectRepository {
				return &mocks.ProjectRepository{}
			},
			userRepository: func() *mocks.UserRepository {
				return &mocks.UserRepository{}
			},
			mailService: func(t *testing.T) *mocks.MailService {
				return mocks.NewMailService(t)
			},
			lotteryHistoryRepo: func() *mocks.LotteryHistoryRepository {
				return &mocks.LotteryHistoryRepository{}
			},
			operatorHistoryRepo: func() *mocks.OperatorHistoryRepository {
				return &mocks.OperatorHistoryRepository{}
			},
			want:    adomain.LotteryCancelResponse{},
			wantErr: true,
		},
		{
			name: "(失敗) 抽選申込キャンセル_更新時にエラーが発生する",
			args: args{
				ctx: actorcontext.WithActor(context.Background(), jwthelper.User{Id: 999}),
				request: adomain.LotteryCancelRequest{
					LotteryApplicationID: "app-001",
				},
			},
			lotteryRepository: func() *mocks.LotteryApplicationRepository {
				repo := &mocks.LotteryApplicationRepository{}
				repo.On("LockById", mock.Anything, "app-001").
					Return(domain.LotteryApplication{Id: "app-001", ProjectId: 1}, nil)
				repo.On("MarkAsCanceled", mock.Anything, "app-001").
					Return(false, errors.New("update failed"))
				return repo
			},
			projectRepository: func() *mocks.ProjectRepository {
				repo := &mocks.ProjectRepository{}
				repo.On("LockById", mock.Anything, int64(1)).
					Return(domain.Project{Id: 1, ProjectStatus: domain.OPEN_FUNDING}, nil)
				return repo
			},
			userRepository: func() *mocks.UserRepository {
				return &mocks.UserRepository{}
			},
			mailService: func(t *testing.T) *mocks.MailService {
				return mocks.NewMailService(t)
			},
			lotteryHistoryRepo: func() *mocks.LotteryHistoryRepository {
				return &mocks.LotteryHistoryRepository{}
			},
			operatorHistoryRepo: func() *mocks.OperatorHistoryRepository {
				return &mocks.OperatorHistoryRepository{}
			},
			want:    adomain.LotteryCancelResponse{},
			wantErr: true,
		},
		{
			name: "(失敗) 抽選申込キャンセル_更新結果が false になる",
			args: args{
				ctx: actorcontext.WithActor(context.Background(), jwthelper.User{Id: 999}),
				request: adomain.LotteryCancelRequest{
					LotteryApplicationID: "app-001",
				},
			},
			lotteryRepository: func() *mocks.LotteryApplicationRepository {
				repo := &mocks.LotteryApplicationRepository{}
				repo.On("LockById", mock.Anything, "app-001").
					Return(domain.LotteryApplication{Id: "app-001", ProjectId: 1}, nil)
				repo.On("MarkAsCanceled", mock.Anything, "app-001").
					Return(false, nil)
				return repo
			},
			projectRepository: func() *mocks.ProjectRepository {
				repo := &mocks.ProjectRepository{}
				repo.On("LockById", mock.Anything, int64(1)).
					Return(domain.Project{Id: 1, ProjectStatus: domain.OPEN_FUNDING}, nil)
				return repo
			},
			userRepository: func() *mocks.UserRepository {
				return &mocks.UserRepository{}
			},
			mailService: func(t *testing.T) *mocks.MailService {
				return mocks.NewMailService(t)
			},
			lotteryHistoryRepo: func() *mocks.LotteryHistoryRepository {
				return &mocks.LotteryHistoryRepository{}
			},
			operatorHistoryRepo: func() *mocks.OperatorHistoryRepository {
				return &mocks.OperatorHistoryRepository{}
			},
			want:    adomain.LotteryCancelResponse{},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u := lotteryUsecase{
				cfg: &config.Config{
					CommonConfig: config.CommonConfig{
						ServiceConfig: struct {
							Mail serviceutils.MailConfig
							Url  serviceutils.UrlConfig
						}{
							Mail: serviceutils.MailConfig{Info: "info@example.com"},
							Url:  serviceutils.UrlConfig{UserServiceDomain: "https://user.example.com"},
						},
					},
				},
				lotteryRepository:         tt.lotteryRepository(),
				lotteryHistoryRepository:  tt.lotteryHistoryRepo(),
				projectRepository:         tt.projectRepository(),
				userRepository:            tt.userRepository(),
				operatorHistoryRepository: tt.operatorHistoryRepo(),
				mailService:               tt.mailService(t),
				txManager:                 newLotteryCancelTxManagerMock(),
			}

			got, err := u.Cancel(tt.args.ctx, tt.args.request)

			assert.Equalf(t, tt.want, got, "Cancel() got = %v, want %v", got, tt.want)
			assert.Equalf(t, tt.wantErr, err != nil, "Cancel() error = %v, wantErr %v", err, tt.wantErr)
		})
	}
}

func Test_lotteryWinProposalQtyRange(t *testing.T) {
	tests := []struct {
		name                   string
		application            domain.LotteryApplication
		projectWinningQtyTotal int64
		wantMax                int64
		wantMin                int64
	}{
		{
			name: "ファンド側残口数が十分なら申込残口数が最大当選口数になる",
			application: domain.LotteryApplication{
				Qty:            30,
				TargetShareQty: 100,
				MinShareQty:    5,
			},
			projectWinningQtyTotal: 0,
			wantMax:                30,
			wantMin:                5,
		},
		{
			name: "部分当選済みなら申込残口数を上限にする",
			application: domain.LotteryApplication{
				Qty:            30,
				TargetShareQty: 100,
				MinShareQty:    5,
				Winnings: []domain.Share{
					{Qty: 10},
				},
			},
			projectWinningQtyTotal: 0,
			wantMax:                20,
			wantMin:                5,
		},
		{
			name: "ファンド側残口数が申込残口数より小さい場合はファンド側残口数を上限にする",
			application: domain.LotteryApplication{
				Qty:            30,
				TargetShareQty: 100,
				MinShareQty:    10,
			},
			projectWinningQtyTotal: 95,
			wantMax:                5,
			wantMin:                5,
		},
		{
			name: "申込残口数が最低投資口数未満なら最小も最大に合わせる",
			application: domain.LotteryApplication{
				Qty:            30,
				TargetShareQty: 100,
				MinShareQty:    10,
				Winnings: []domain.Share{
					{Qty: 25},
				},
			},
			projectWinningQtyTotal: 0,
			wantMax:                5,
			wantMin:                5,
		},
		{
			name: "ファンド側残口数が負になる場合は0に丸める",
			application: domain.LotteryApplication{
				Qty:            30,
				TargetShareQty: 100,
				MinShareQty:    10,
			},
			projectWinningQtyTotal: 120,
			wantMax:                0,
			wantMin:                0,
		},
		{
			name: "申込残口数が負になる場合は0に丸める",
			application: domain.LotteryApplication{
				Qty:            10,
				TargetShareQty: 100,
				MinShareQty:    10,
				Winnings: []domain.Share{
					{Qty: 12},
				},
			},
			projectWinningQtyTotal: 0,
			wantMax:                0,
			wantMin:                0,
		},
		{
			name: "project 情報が未設定なら0を返す",
			application: domain.LotteryApplication{
				Qty: 10,
			},
			projectWinningQtyTotal: 10,
			wantMax:                0,
			wantMin:                0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotMax, gotMin := lotteryWinProposalQtyRangeByApplication(tt.application, tt.projectWinningQtyTotal)
			assert.Equal(t, tt.wantMax, gotMax)
			assert.Equal(t, tt.wantMin, gotMin)
		})
	}
}

func Test_lotteryStatusAfterWinning(t *testing.T) {
	tests := []struct {
		name      string
		app       domain.LotteryApplication
		addQty    int64
		wantState domain.LOTTERY_STATUS
	}{
		{
			name: "全量当選でWINNINGになる",
			app: domain.LotteryApplication{
				Qty: 10,
			},
			addQty:    10,
			wantState: domain.LOTTERY_STATUS_WINNING,
		},
		{
			name: "一部当選でPARTIAL_WINNINGになる",
			app: domain.LotteryApplication{
				Qty: 10,
			},
			addQty:    3,
			wantState: domain.LOTTERY_STATUS_PARTIAL_WINNING,
		},
		{
			name: "既当選がある場合は合算して判定する",
			app: domain.LotteryApplication{
				Qty: 10,
				Winnings: []domain.Share{
					{Qty: 4},
				},
			},
			addQty:    6,
			wantState: domain.LOTTERY_STATUS_WINNING,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.wantState, lotteryStatusAfterWinning(tt.app, tt.addQty))
		})
	}
}

func Test_lotteryStatusAfterLosing(t *testing.T) {
	tests := []struct {
		name      string
		app       domain.LotteryApplication
		wantState domain.LOTTERY_STATUS
	}{
		{
			name: "未当選申込の落選はLOSTになる",
			app: domain.LotteryApplication{
				Status: domain.LOTTERY_STATUS_APPLIED,
			},
			wantState: domain.LOTTERY_STATUS_LOST,
		},
		{
			name: "一部当選済み申込の残口落選はPARTIAL_WINNINGを維持する",
			app: domain.LotteryApplication{
				Status: domain.LOTTERY_STATUS_PARTIAL_WINNING,
			},
			wantState: domain.LOTTERY_STATUS_PARTIAL_WINNING,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.wantState, lotteryStatusAfterLosing(tt.app))
		})
	}
}

func Test_escapeCSVFormulaString(t *testing.T) {
	tests := []struct {
		name  string
		value string
		want  string
	}{
		{name: "equal", value: "=cmd|'/C calc'!A0", want: "'=cmd|'/C calc'!A0"},
		{name: "plus", value: "+SUM(A1:A2)", want: "'+SUM(A1:A2)"},
		{name: "minus", value: "-10+20", want: "'-10+20"},
		{name: "at", value: "@SUM(A1:A2)", want: "'@SUM(A1:A2)"},
		{name: "safe", value: "app-001", want: "app-001"},
		{name: "empty", value: "", want: ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, escapeCSVFormulaString(tt.value))
		})
	}
}

func Test_parseLotteryCSVUploadRows(t *testing.T) {
	tests := []struct {
		name      string
		content   string
		wantRows  []lotteryCSVUploadRow
		wantErr   error
		filename  string
		fieldName string
	}{
		{
			name:      "正常系",
			fieldName: "file",
			filename:  "lottery.csv",
			content:   "抽選ID,抽選ステータス,投資家名,メールアドレス,ファンド名,申込口数,申込金額,当選総額,当選後キャンセル額,投資可能金額,申込日時,当落,当選させたい金額\napp-001,申込中,投資家A,a@example.com,ファンドA,10,100000,0,0,100000,2026/5/1 10:00,落選,0\napp-002,申込中,投資家B,b@example.com,ファンドB,5,50000,30000,10000,50000,2026/5/1 10:05,当選,10000\n",
			wantRows: []lotteryCSVUploadRow{
				{rowNumber: 2, lotteryApplicationID: "app-001", lotteryStatus: "申込中", investorName: "投資家A", email: "a@example.com", fundName: "ファンドA", appliedQty: "10", appliedAmount: "100000", winningAmount: "0", canceledWinningAmount: "0", availableAmount: "100000", appliedAt: "2026/5/1 10:00", result: "落選", desiredWinningAmount: "0"},
				{rowNumber: 3, lotteryApplicationID: "app-002", lotteryStatus: "申込中", investorName: "投資家B", email: "b@example.com", fundName: "ファンドB", appliedQty: "5", appliedAmount: "50000", winningAmount: "30000", canceledWinningAmount: "10000", availableAmount: "50000", appliedAt: "2026/5/1 10:05", result: "当選", desiredWinningAmount: "10000"},
			},
		},
		{
			name:      "UTF-8 BOM付き正常系",
			fieldName: "file",
			filename:  "lottery.csv",
			content:   "\ufeff抽選ID,抽選ステータス,投資家名,メールアドレス,ファンド名,申込口数,申込金額,当選総額,当選後キャンセル額,投資可能金額,申込日時,当落,当選させたい金額\napp-001,申込中,投資家A,a@example.com,ファンドA,10,100000,0,0,100000,2026/5/1 10:00,落選,0\n",
			wantRows: []lotteryCSVUploadRow{
				{rowNumber: 2, lotteryApplicationID: "app-001", lotteryStatus: "申込中", investorName: "投資家A", email: "a@example.com", fundName: "ファンドA", appliedQty: "10", appliedAmount: "100000", winningAmount: "0", canceledWinningAmount: "0", availableAmount: "100000", appliedAt: "2026/5/1 10:00", result: "落選", desiredWinningAmount: "0"},
			},
		},
		{
			name:      "ヘッダ不一致",
			fieldName: "file",
			filename:  "lottery.csv",
			content:   "lottery_application_id,status,user_name,email,project_name,applied_qty,applied_amount,winning_amount,available_amount,applied_at,result,desired_winning_amount\napp-001,APPLIED,投資家A,a@example.com,ファンドA,10,100000,0,100000,2026/5/1 10:00,落選,0\n",
			wantErr:   cdomain.ErrBadRequest,
		},
		{
			name:      "空行を含む",
			fieldName: "file",
			filename:  "lottery.csv",
			content:   "抽選ID,抽選ステータス,投資家名,メールアドレス,ファンド名,申込口数,申込金額,当選総額,当選後キャンセル額,投資可能金額,申込日時,当落,当選させたい金額\napp-001,申込中,投資家A,a@example.com,ファンドA,10,100000,0,0,100000,2026/5/1 10:00,落選,0\n,,,,,,,,,,,,\n",
			wantErr:   cdomain.ErrBadRequest,
		},
		{
			name:      "列数不一致",
			fieldName: "file",
			filename:  "lottery.csv",
			content:   "抽選ID,抽選ステータス,投資家名,メールアドレス,ファンド名,申込口数,申込金額,当選総額,当選後キャンセル額,投資可能金額,申込日時,当落,当選させたい金額\napp-001,申込中,投資家A,a@example.com,ファンドA,10,100000,0,0,100000,2026/5/1 10:00,落選\n",
			wantErr:   cdomain.ErrBadRequest,
		},
		{
			name:      "ファイルサイズ上限超過",
			fieldName: "file",
			filename:  "lottery.csv",
			content:   strings.Repeat("a", int(adomain.LotteryCSVUploadMaxFileSizeBytes)+1),
			wantErr:   cdomain.ErrBadRequest,
		},
		{
			name:      "行数上限超過",
			fieldName: "file",
			filename:  "lottery.csv",
			content: func() string {
				var b strings.Builder
				b.WriteString("抽選ID,抽選ステータス,投資家名,メールアドレス,ファンド名,申込口数,申込金額,当選総額,当選後キャンセル額,投資可能金額,申込日時,当落,当選させたい金額\n")
				for i := 0; i < adomain.LotteryCSVUploadMaxRows+1; i++ {
					b.WriteString("app-001,申込中,投資家A,a@example.com,ファンドA,10,100000,0,0,100000,2026/5/1 10:00,落選,0\n")
				}
				return b.String()
			}(),
			wantErr: cdomain.ErrBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fileHeader := newTestMultipartFileHeader(t, tt.fieldName, tt.filename, tt.content)
			got, err := parseLotteryCSVUploadRows(fileHeader)
			if tt.wantErr == nil {
				assert.NoError(t, err)
				assert.Equal(t, tt.wantRows, got)
				return
			}

			assert.Error(t, err)
			assert.True(t, errors.Is(err, tt.wantErr))
			if tt.name == "行数上限超過" {
				var csvValidationErr *adomain.LotteryCSVValidationError
				assert.True(t, errors.As(err, &csvValidationErr))
				if assert.Len(t, csvValidationErr.Details, 1) {
					assert.Equal(t, "file", csvValidationErr.Details[0].Field)
					assert.Equal(t, "申込件数は10000件以下にしてください", csvValidationErr.Details[0].Message)
				}
			}
		})
	}
}

func Test_validateLotteryCSVUploadRows(t *testing.T) {
	lotteryRepo := mocks.NewLotteryApplicationRepository(t)
	userRepo := mocks.NewUserRepository(t)

	appliedAt, err := jsonutils.ParseJSONTimeJST(time.DateTime, "2026-05-01 10:00:00")
	assert.NoError(t, err)

	application := domain.LotteryApplication{
		Id:             "app-001",
		ProjectId:      101,
		UserId:         5001,
		Qty:            10,
		UnitPrice:      10000,
		Amount:         100000,
		AppliedAt:      appliedAt,
		Status:         domain.LOTTERY_STATUS_APPLIED,
		ProjectName:    "ファンドA",
		UserName:       "投資家A",
		TargetShareQty: 100,
		MinShareQty:    1,
	}

	lotteryRepo.On("GetById", mock.Anything, "app-001").Return(application, nil)
	userRepo.On("GetById", mock.Anything, int64(5001)).Return(domain.User{Id: 5001, Email: "a@example.com"}, nil)

	u := lotteryUsecase{
		cfg:                       &config.Config{},
		lotteryRepository:         lotteryRepo,
		lotteryHistoryRepository:  nil,
		projectRepository:         nil,
		userRepository:            userRepo,
		operatorHistoryRepository: nil,
		mailService:               nil,
		txManager:                 nil,
	}

	rows := []lotteryCSVUploadRow{
		{
			rowNumber:            2,
			lotteryApplicationID: "app-001",
			lotteryStatus:        "申込中",
			investorName:         "投資家A",
			email:                "a@example.com",
			fundName:             "ファンドA",
			appliedQty:           "10",
			appliedAmount:        "100000",
			winningAmount:        "0",
			availableAmount:      "100000",
			appliedAt:            "2026/5/1 10:00",
			result:               "",
			desiredWinningAmount: "",
		},
		{
			rowNumber:            3,
			lotteryApplicationID: "app-001",
			lotteryStatus:        "申込中",
			investorName:         "投資家A",
			email:                "a@example.com",
			fundName:             "ファンドA",
			appliedQty:           "10",
			appliedAmount:        "100000",
			winningAmount:        "0",
			availableAmount:      "100000",
			appliedAt:            "2026/5/1 10:00",
			result:               "保留",
			desiredWinningAmount: "0",
		},
	}

	_, err = u.validateLotteryCSVUploadRows(context.Background(), domain.Project{Id: 101}, rows)
	assert.Error(t, err)
	assert.True(t, errors.Is(err, cdomain.ErrBadRequest))
	assert.Contains(t, err.Error(), "row 2 [抽選ID]: 同一CSV内で抽選IDが重複しています")
	assert.Contains(t, err.Error(), "row 3 [当落]: 当落は「当選」「落選」または空欄で入力してください")
}

func Test_lotteryUsecase_ValidateLotteryCSVUploadRows_RejectsWhenProjectAmountLimitExceeded(t *testing.T) {
	appliedAt, err := jsonutils.ParseJSONTimeJST(time.DateTime, "2026-05-01 10:00:00")
	assert.NoError(t, err)

	lotteryRepo := mocks.NewLotteryApplicationRepository(t)
	lotteryRepo.On("GetById", mock.Anything, "app-001").Return(domain.LotteryApplication{
		Id:             "app-001",
		ProjectId:      101,
		ProjectName:    "ファンドA",
		UserId:         5001,
		UserName:       "投資家A",
		Qty:            10,
		UnitPrice:      10000,
		Amount:         100000,
		AppliedAt:      appliedAt,
		Status:         domain.LOTTERY_STATUS_APPLIED,
		TargetShareQty: 100,
		MinShareQty:    1,
	}, nil)
	userRepo := mocks.NewUserRepository(t)
	userRepo.On("GetById", mock.Anything, int64(5001)).Return(domain.User{Id: 5001, Email: "a@example.com"}, nil)

	u := lotteryUsecase{
		cfg:               &config.Config{},
		lotteryRepository: lotteryRepo,
		userRepository:    userRepo,
	}

	rows := []lotteryCSVUploadRow{
		{
			rowNumber:            2,
			lotteryApplicationID: "app-001",
			lotteryStatus:        "申込中",
			investorName:         "投資家A",
			email:                "a@example.com",
			fundName:             "ファンドA",
			appliedQty:           "10",
			appliedAmount:        "100000",
			winningAmount:        "0",
			appliedAt:            "2026/5/1 10:00",
			result:               "当選",
			desiredWinningAmount: "30000",
		},
	}

	_, err = u.validateLotteryCSVUploadRows(context.Background(), domain.Project{
		Id:                 101,
		SharePrice:         10000,
		CurrentShareAmount: 80000,
		TargetShareAmount:  100000,
		TargetShareQty:     100,
		RemainingQty:       50,
	}, rows)
	assert.Error(t, err)
	assert.True(t, errors.Is(err, cdomain.ErrBadRequest))
	if err != nil {
		assert.Contains(t, err.Error(), "CSV内の当選指定合計がファンドの募集金額上限を超えています")
	}
}

func Test_lotteryUsecase_ValidateLotteryCSVUploadRows_RejectsWhenProjectQtyLimitExceeded(t *testing.T) {
	appliedAt, err := jsonutils.ParseJSONTimeJST(time.DateTime, "2026-05-01 10:00:00")
	assert.NoError(t, err)

	lotteryRepo := mocks.NewLotteryApplicationRepository(t)
	lotteryRepo.On("GetById", mock.Anything, "app-001").Return(domain.LotteryApplication{
		Id:             "app-001",
		ProjectId:      101,
		ProjectName:    "ファンドA",
		UserId:         5001,
		UserName:       "投資家A",
		Qty:            10,
		UnitPrice:      1000,
		Amount:         10000,
		AppliedAt:      appliedAt,
		Status:         domain.LOTTERY_STATUS_APPLIED,
		TargetShareQty: 100,
		MinShareQty:    1,
	}, nil)
	userRepo := mocks.NewUserRepository(t)
	userRepo.On("GetById", mock.Anything, int64(5001)).Return(domain.User{Id: 5001, Email: "a@example.com"}, nil)

	u := lotteryUsecase{
		cfg:               &config.Config{},
		lotteryRepository: lotteryRepo,
		userRepository:    userRepo,
	}

	rows := []lotteryCSVUploadRow{
		{
			rowNumber:            2,
			lotteryApplicationID: "app-001",
			lotteryStatus:        "申込中",
			investorName:         "投資家A",
			email:                "a@example.com",
			fundName:             "ファンドA",
			appliedQty:           "10",
			appliedAmount:        "10000",
			winningAmount:        "0",
			appliedAt:            "2026/5/1 10:00",
			result:               "当選",
			desiredWinningAmount: "3000",
		},
	}

	_, err = u.validateLotteryCSVUploadRows(context.Background(), domain.Project{
		Id:                 101,
		SharePrice:         1000,
		CurrentShareAmount: 50000,
		TargetShareAmount:  100000,
		TargetShareQty:     100,
		RemainingQty:       2,
	}, rows)
	assert.Error(t, err)
	assert.True(t, errors.Is(err, cdomain.ErrBadRequest))
	if err != nil {
		assert.Contains(t, err.Error(), "CSV内の当選指定合計がファンドの募集口数上限を超えています")
	}
}

func Test_lotteryUsecase_UploadCSV(t *testing.T) {
	engine, ctx := newLotteryUploadTestDB(t)

	projectRepo := mocks.NewProjectRepository(t)
	lotteryRepo := mocks.NewLotteryApplicationRepository(t)
	userRepo := mocks.NewUserRepository(t)

	appliedAt1, err := jsonutils.ParseJSONTimeJST(time.DateTime, "2026-05-01 10:00:00")
	assert.NoError(t, err)
	appliedAt2, err := jsonutils.ParseJSONTimeJST(time.DateTime, "2026-05-01 11:00:00")
	assert.NoError(t, err)

	projectRepo.On("GetById", mock.Anything, int64(101)).Return(domain.Project{
		Id:                101,
		Name:              "ファンドA",
		SelectionType:     domain.PROJECT_SELECTION_LOTTERY,
		ProjectStatus:     domain.END_FUNDING,
		SharePrice:        10000,
		TargetShareQty:    100,
		RemainingQty:      100,
		TargetShareAmount: 1000000,
	}, nil)
	projectRepo.On("LockById", mock.Anything, int64(101)).Return(domain.Project{
		Id:                101,
		Name:              "ファンドA",
		SelectionType:     domain.PROJECT_SELECTION_LOTTERY,
		ProjectStatus:     domain.END_FUNDING,
		SharePrice:        10000,
		TargetShareQty:    100,
		RemainingQty:      100,
		TargetShareAmount: 1000000,
	}, nil)
	lotteryRepo.On("GetById", mock.Anything, "app-001").Return(domain.LotteryApplication{
		Id:             "app-001",
		ProjectId:      101,
		ProjectName:    "ファンドA",
		UserId:         5001,
		UserName:       "投資家A",
		Qty:            10,
		UnitPrice:      10000,
		Amount:         100000,
		AppliedAt:      appliedAt1,
		Status:         domain.LOTTERY_STATUS_APPLIED,
		TargetShareQty: 100,
		MinShareQty:    1,
	}, nil)
	lotteryRepo.On("GetById", mock.Anything, "app-002").Return(domain.LotteryApplication{
		Id:             "app-002",
		ProjectId:      101,
		ProjectName:    "ファンドA",
		UserId:         5002,
		UserName:       "投資家B",
		Qty:            5,
		UnitPrice:      10000,
		Amount:         50000,
		AppliedAt:      appliedAt2,
		Status:         domain.LOTTERY_STATUS_APPLIED,
		TargetShareQty: 100,
		MinShareQty:    1,
	}, nil)
	userRepo.On("GetById", mock.Anything, int64(5001)).Return(domain.User{Id: 5001, Email: "a@example.com"}, nil)
	userRepo.On("GetById", mock.Anything, int64(5002)).Return(domain.User{Id: 5002, Email: "b@example.com"}, nil)

	u := lotteryUsecase{
		cfg:                       &config.Config{},
		lotteryRepository:         lotteryRepo,
		lotteryHistoryRepository:  nil,
		projectRepository:         projectRepo,
		userRepository:            userRepo,
		operatorHistoryRepository: nil,
		mailService:               nil,
		txManager:                 factory.NewTxManager(engine),
	}

	fileHeader := newTestMultipartFileHeader(t, "file", "lottery.csv", "抽選ID,抽選ステータス,投資家名,メールアドレス,ファンド名,申込口数,申込金額,当選総額,当選後キャンセル額,投資可能金額,申込日時,当落,当選させたい金額\napp-001,申込中,投資家A,a@example.com,ファンドA,10,100000,0,0,999999,2026/5/1 10:00,当選,20000\napp-002,申込中,投資家B,b@example.com,ファンドA,5,50000,0,0,abc,2026/5/1 11:00,,\n")

	response, err := u.UploadCSV(ctx, adomain.LotteryCSVUploadRequest{
		JWTUser: jwthelper.User{
			Id:    1000,
			Email: "operator@email.com",
			Name:  "オペレーター",
		},
		ProjectID: 101,
		File:      fileHeader,
	})
	assert.NoError(t, err)
	assert.NotEmpty(t, response.JobID)
	assert.Equal(t, domain.LOTTERY_CSV_UPLOAD_JOB_STATUS_QUEUED, response.Status)
	if err != nil {
		return
	}

	job, err := (domain.LotteryCSVUploadJob{}).GetByJobID(ctx, response.JobID)
	assert.NoError(t, err)
	if assert.NotNil(t, job) {
		assert.Equal(t, 2, job.TotalRowCount)
		assert.Equal(t, "lottery.csv", job.FileName)
	} else {
		return
	}

	details, err := (domain.LotteryCSVUploadJobDetail{}).GetByJobID(ctx, job.Id)
	assert.NoError(t, err)
	if assert.Len(t, details, 2) {
		assert.Equal(t, domain.LOTTERY_CSV_UPLOAD_PROCESS_TYPE_WIN, details[0].ProcessType)
		assert.Equal(t, domain.LOTTERY_CSV_UPLOAD_DETAIL_STATUS_PENDING, details[0].Status)
		assert.Equal(t, domain.LOTTERY_STATUS_APPLIED, lo.FromPtr(details[0].ExpectedLotteryStatus))
		assert.Equal(t, "2026/5/1 10:00", lo.FromPtr(details[0].ExpectedAppliedAt))
		assert.EqualValues(t, 10, lo.FromPtr(details[0].InputAppliedQty))
		assert.EqualValues(t, 2, details[0].InputWinningQty)

		assert.Equal(t, domain.LOTTERY_CSV_UPLOAD_PROCESS_TYPE_SKIP, details[1].ProcessType)
		assert.Equal(t, domain.LOTTERY_CSV_UPLOAD_DETAIL_STATUS_SKIPPED, details[1].Status)
		assert.EqualValues(t, 0, details[1].InputWinningQty)
	}
}

func Test_lotteryUsecase_UploadCSV_RejectsWhenActiveJobExists(t *testing.T) {
	engine, ctx := newLotteryUploadTestDB(t)

	_, err := (&domain.LotteryCSVUploadJob{
		JobID:       "existing-job",
		ProjectID:   101,
		RequestedBy: "operator@email.com",
		FileName:    "existing.csv",
		Status:      domain.LOTTERY_CSV_UPLOAD_JOB_STATUS_QUEUED,
	}).Create(ctx)
	assert.NoError(t, err)

	projectRepo := mocks.NewProjectRepository(t)
	projectRepo.On("GetById", mock.Anything, int64(101)).Return(domain.Project{
		Id:            101,
		SelectionType: domain.PROJECT_SELECTION_LOTTERY,
		ProjectStatus: domain.END_FUNDING,
	}, nil)

	u := lotteryUsecase{
		cfg:                       &config.Config{},
		lotteryRepository:         mocks.NewLotteryApplicationRepository(t),
		lotteryHistoryRepository:  nil,
		projectRepository:         projectRepo,
		userRepository:            mocks.NewUserRepository(t),
		operatorHistoryRepository: nil,
		mailService:               nil,
		txManager:                 factory.NewTxManager(engine),
	}

	fileHeader := newTestMultipartFileHeader(t, "file", "lottery.csv", "抽選ID,抽選ステータス,投資家名,メールアドレス,ファンド名,申込口数,申込金額,当選総額,当選後キャンセル額,投資可能金額,申込日時,当落,当選させたい金額\napp-001,申込中,投資家A,a@example.com,ファンドA,10,100000,0,0,100000,2026/5/1 10:00,当選,10000\n")

	_, err = u.UploadCSV(ctx, adomain.LotteryCSVUploadRequest{
		JWTUser:   jwthelper.User{Id: 1000, Email: "operator@email.com"},
		ProjectID: 101,
		File:      fileHeader,
	})
	assert.Error(t, err)
	assert.True(t, errors.Is(err, cdomain.ErrBadRequest))
	assert.Contains(t, err.Error(), "already queued or running")
}

func Test_lotteryUsecase_UploadCSV_UsesPersistedJobAutoIncrementID(t *testing.T) {
	engine, ctx := newLotteryUploadTestDB(t)

	projectRepo := mocks.NewProjectRepository(t)
	lotteryRepo := mocks.NewLotteryApplicationRepository(t)
	userRepo := mocks.NewUserRepository(t)

	appliedAt, err := jsonutils.ParseJSONTimeJST(time.DateTime, "2026-05-01 10:00:00")
	assert.NoError(t, err)

	projectRepo.On("GetById", mock.Anything, int64(101)).Return(domain.Project{
		Id:                101,
		Name:              "ファンドA",
		SelectionType:     domain.PROJECT_SELECTION_LOTTERY,
		ProjectStatus:     domain.END_FUNDING,
		SharePrice:        10000,
		TargetShareQty:    100,
		RemainingQty:      100,
		TargetShareAmount: 1000000,
	}, nil)
	projectRepo.On("LockById", mock.Anything, int64(101)).Return(domain.Project{
		Id:                101,
		Name:              "ファンドA",
		SelectionType:     domain.PROJECT_SELECTION_LOTTERY,
		ProjectStatus:     domain.END_FUNDING,
		SharePrice:        10000,
		TargetShareQty:    100,
		RemainingQty:      100,
		TargetShareAmount: 1000000,
	}, nil)
	lotteryRepo.On("GetById", mock.Anything, "app-001").Return(domain.LotteryApplication{
		Id:             "app-001",
		ProjectId:      101,
		ProjectName:    "ファンドA",
		UserId:         5001,
		UserName:       "投資家A",
		Qty:            10,
		UnitPrice:      10000,
		Amount:         100000,
		AppliedAt:      appliedAt,
		Status:         domain.LOTTERY_STATUS_APPLIED,
		TargetShareQty: 100,
		MinShareQty:    1,
	}, nil)
	userRepo.On("GetById", mock.Anything, int64(5001)).Return(domain.User{Id: 5001, Email: "a@example.com"}, nil)

	u := lotteryUsecase{
		cfg:                       &config.Config{},
		lotteryRepository:         lotteryRepo,
		lotteryHistoryRepository:  nil,
		projectRepository:         projectRepo,
		userRepository:            userRepo,
		operatorHistoryRepository: nil,
		mailService:               nil,
		txManager:                 factory.NewTxManager(engine),
	}

	fileHeader1 := newTestMultipartFileHeader(t, "file", "lottery1.csv", "抽選ID,抽選ステータス,投資家名,メールアドレス,ファンド名,申込口数,申込金額,当選総額,当選後キャンセル額,投資可能金額,申込日時,当落,当選させたい金額\napp-001,申込中,投資家A,a@example.com,ファンドA,10,100000,0,0,100000,2026/5/1 10:00,,\n")
	resp1, err := u.UploadCSV(ctx, adomain.LotteryCSVUploadRequest{
		JWTUser:   jwthelper.User{Id: 1000, Email: "operator@email.com", Name: "オペレーター"},
		ProjectID: 101,
		File:      fileHeader1,
	})
	assert.NoError(t, err)

	job1, err := (domain.LotteryCSVUploadJob{}).GetByJobID(ctx, resp1.JobID)
	assert.NoError(t, err)
	if !assert.NotNil(t, job1) {
		return
	}
	assert.EqualValues(t, 1, job1.Id)

	job1.Status = domain.LOTTERY_CSV_UPLOAD_JOB_STATUS_COMPLETED
	assert.NoError(t, job1.Update(ctx))

	fileHeader2 := newTestMultipartFileHeader(t, "file", "lottery2.csv", "抽選ID,抽選ステータス,投資家名,メールアドレス,ファンド名,申込口数,申込金額,当選総額,当選後キャンセル額,投資可能金額,申込日時,当落,当選させたい金額\napp-001,申込中,投資家A,a@example.com,ファンドA,10,100000,0,0,100000,2026/5/1 10:00,当選,10000\n")
	resp2, err := u.UploadCSV(ctx, adomain.LotteryCSVUploadRequest{
		JWTUser:   jwthelper.User{Id: 1000, Email: "operator@email.com", Name: "オペレーター"},
		ProjectID: 101,
		File:      fileHeader2,
	})
	assert.NoError(t, err)

	job2, err := (domain.LotteryCSVUploadJob{}).GetByJobID(ctx, resp2.JobID)
	assert.NoError(t, err)
	if !assert.NotNil(t, job2) {
		return
	}
	assert.EqualValues(t, 2, job2.Id)

	details, err := (domain.LotteryCSVUploadJobDetail{}).GetByJobID(ctx, job2.Id)
	assert.NoError(t, err)
	if assert.Len(t, details, 1) {
		assert.EqualValues(t, 2, details[0].JobID)
		assert.Equal(t, 2, details[0].RowNumber)
	}
}

func Test_lotteryUsecase_ProcessCSVUploadJob_CompletesWithoutPendingRows(t *testing.T) {
	_, ctx := newLotteryUploadTestDB(t)

	_, err := (&domain.LotteryCSVUploadJob{
		JobID:         "job-no-pending",
		ProjectID:     101,
		RequestedBy:   "operator@example.com",
		FileName:      "lottery.csv",
		Status:        domain.LOTTERY_CSV_UPLOAD_JOB_STATUS_QUEUED,
		TotalRowCount: 1,
	}).Create(ctx)
	assert.NoError(t, err)

	job, err := (domain.LotteryCSVUploadJob{}).GetByJobID(ctx, "job-no-pending")
	assert.NoError(t, err)
	if !assert.NotNil(t, job) {
		return
	}

	_, err = (&domain.LotteryCSVUploadJobDetail{
		JobID:                job.Id,
		RowNumber:            2,
		LotteryApplicationID: "app-skip",
		InputWinningQty:      0,
		ProcessType:          domain.LOTTERY_CSV_UPLOAD_PROCESS_TYPE_SKIP,
		Status:               domain.LOTTERY_CSV_UPLOAD_DETAIL_STATUS_SKIPPED,
	}).Create(ctx)
	assert.NoError(t, err)

	projectRepo := mocks.NewProjectRepository(t)
	projectRepo.On("GetById", mock.Anything, int64(101)).Return(domain.Project{
		Id:            101,
		SelectionType: domain.PROJECT_SELECTION_LOTTERY,
		ProjectStatus: domain.END_FUNDING,
	}, nil)

	u := lotteryUsecase{
		cfg:                       &config.Config{},
		lotteryRepository:         mocks.NewLotteryApplicationRepository(t),
		lotteryHistoryRepository:  nil,
		projectRepository:         projectRepo,
		userRepository:            mocks.NewUserRepository(t),
		operatorHistoryRepository: nil,
		mailService:               nil,
		txManager:                 nil,
	}

	err = u.ProcessCSVUploadJob(ctx, "job-no-pending")
	assert.NoError(t, err)

	job, err = (domain.LotteryCSVUploadJob{}).GetByJobID(ctx, "job-no-pending")
	assert.NoError(t, err)
	if assert.NotNil(t, job) {
		assert.Equal(t, domain.LOTTERY_CSV_UPLOAD_JOB_STATUS_COMPLETED, job.Status)
		assert.Equal(t, 0, job.SuccessRowCount)
		assert.Equal(t, 0, job.FailedRowCount)
		assert.False(t, job.FinishedAt.IsZero())
	}
}

func Test_lotteryUsecase_ProcessCSVUploadJob_FailsOnRevalidation(t *testing.T) {
	_, ctx := newLotteryUploadTestDB(t)

	_, err := (&domain.LotteryCSVUploadJob{
		JobID:         "job-revalidate-fail",
		ProjectID:     101,
		RequestedBy:   "operator@example.com",
		FileName:      "lottery.csv",
		Status:        domain.LOTTERY_CSV_UPLOAD_JOB_STATUS_QUEUED,
		TotalRowCount: 1,
	}).Create(ctx)
	assert.NoError(t, err)

	job, err := (domain.LotteryCSVUploadJob{}).GetByJobID(ctx, "job-revalidate-fail")
	assert.NoError(t, err)
	if !assert.NotNil(t, job) {
		return
	}

	appliedQty := int64(10)
	_, err = (&domain.LotteryCSVUploadJobDetail{
		JobID:                 job.Id,
		RowNumber:             2,
		LotteryApplicationID:  "app-001",
		ExpectedLotteryStatus: lo.ToPtr(domain.LOTTERY_STATUS_APPLIED),
		InputAppliedQty:       &appliedQty,
		InputWinningQty:       2,
		ProcessType:           domain.LOTTERY_CSV_UPLOAD_PROCESS_TYPE_WIN,
		Status:                domain.LOTTERY_CSV_UPLOAD_DETAIL_STATUS_PENDING,
	}).Create(ctx)
	assert.NoError(t, err)

	projectRepo := mocks.NewProjectRepository(t)
	projectRepo.On("GetById", mock.Anything, int64(101)).Return(domain.Project{
		Id:            101,
		SelectionType: domain.PROJECT_SELECTION_LOTTERY,
		ProjectStatus: domain.END_FUNDING,
	}, nil)

	lotteryRepo := mocks.NewLotteryApplicationRepository(t)
	lotteryRepo.On("GetById", mock.Anything, "app-001").Return(domain.LotteryApplication{
		Id:             "app-001",
		ProjectId:      101,
		Qty:            1,
		Amount:         10000,
		Status:         domain.LOTTERY_STATUS_APPLIED,
		TargetShareQty: 100,
		MinShareQty:    1,
	}, nil)

	u := lotteryUsecase{
		cfg:                       &config.Config{},
		lotteryRepository:         lotteryRepo,
		lotteryHistoryRepository:  nil,
		projectRepository:         projectRepo,
		userRepository:            mocks.NewUserRepository(t),
		operatorHistoryRepository: nil,
		mailService:               nil,
		txManager:                 nil,
	}

	err = u.ProcessCSVUploadJob(ctx, "job-revalidate-fail")
	assert.NoError(t, err)

	job, err = (domain.LotteryCSVUploadJob{}).GetByJobID(ctx, "job-revalidate-fail")
	assert.NoError(t, err)
	if assert.NotNil(t, job) {
		assert.Equal(t, domain.LOTTERY_CSV_UPLOAD_JOB_STATUS_FAILED, job.Status)
		assert.Equal(t, 0, job.SuccessRowCount)
		assert.Equal(t, 1, job.FailedRowCount)
		assert.NotNil(t, job.ErrorMessage)
		assert.Contains(t, lo.FromPtr(job.ErrorMessage), "row 2")
	}

	details, err := (domain.LotteryCSVUploadJobDetail{}).GetByJobID(ctx, job.Id)
	assert.NoError(t, err)
	if assert.Len(t, details, 1) {
		assert.Equal(t, domain.LOTTERY_CSV_UPLOAD_DETAIL_STATUS_FAILED, details[0].Status)
		assert.NotNil(t, details[0].ErrorMessage)
		assert.Contains(t, lo.FromPtr(details[0].ErrorMessage), "applied qty does not match current lottery application qty")
		assert.False(t, details[0].ProcessedAt.IsZero())
	}
}

func Test_lotteryUsecase_ProcessCSVUploadJob_FailsWhenExpectedStatusChanged(t *testing.T) {
	_, ctx := newLotteryUploadTestDB(t)

	_, err := (&domain.LotteryCSVUploadJob{
		JobID:         "job-status-mismatch",
		ProjectID:     101,
		RequestedBy:   "operator@example.com",
		FileName:      "lottery.csv",
		Status:        domain.LOTTERY_CSV_UPLOAD_JOB_STATUS_QUEUED,
		TotalRowCount: 1,
	}).Create(ctx)
	assert.NoError(t, err)

	job, err := (domain.LotteryCSVUploadJob{}).GetByJobID(ctx, "job-status-mismatch")
	assert.NoError(t, err)
	if !assert.NotNil(t, job) {
		return
	}

	appliedQty := int64(10)
	_, err = (&domain.LotteryCSVUploadJobDetail{
		JobID:                 job.Id,
		RowNumber:             2,
		LotteryApplicationID:  "app-001",
		ExpectedLotteryStatus: lo.ToPtr(domain.LOTTERY_STATUS_APPLIED),
		InputAppliedQty:       &appliedQty,
		InputWinningQty:       0,
		ProcessType:           domain.LOTTERY_CSV_UPLOAD_PROCESS_TYPE_LOSE,
		Status:                domain.LOTTERY_CSV_UPLOAD_DETAIL_STATUS_PENDING,
	}).Create(ctx)
	assert.NoError(t, err)

	projectRepo := mocks.NewProjectRepository(t)
	projectRepo.On("GetById", mock.Anything, int64(101)).Return(domain.Project{
		Id:            101,
		SelectionType: domain.PROJECT_SELECTION_LOTTERY,
		ProjectStatus: domain.END_FUNDING,
	}, nil)

	lotteryRepo := mocks.NewLotteryApplicationRepository(t)
	lotteryRepo.On("GetById", mock.Anything, "app-001").Return(domain.LotteryApplication{
		Id:             "app-001",
		ProjectId:      101,
		Qty:            10,
		Amount:         100000,
		Status:         domain.LOTTERY_STATUS_WINNING,
		TargetShareQty: 100,
		MinShareQty:    1,
	}, nil)

	u := lotteryUsecase{
		cfg:                       &config.Config{},
		lotteryRepository:         lotteryRepo,
		lotteryHistoryRepository:  nil,
		projectRepository:         projectRepo,
		userRepository:            mocks.NewUserRepository(t),
		operatorHistoryRepository: nil,
		mailService:               nil,
		txManager:                 nil,
	}

	err = u.ProcessCSVUploadJob(ctx, "job-status-mismatch")
	assert.NoError(t, err)

	details, err := (domain.LotteryCSVUploadJobDetail{}).GetByJobID(ctx, job.Id)
	assert.NoError(t, err)
	if assert.Len(t, details, 1) {
		assert.Equal(t, domain.LOTTERY_CSV_UPLOAD_DETAIL_STATUS_FAILED, details[0].Status)
		assert.NotNil(t, details[0].ErrorMessage)
		assert.Contains(t, lo.FromPtr(details[0].ErrorMessage), "lottery status does not match current lottery application status")
	}
}

func Test_lotteryUsecase_ProcessCSVUploadJob_NoOpsWhenAlreadyRunning(t *testing.T) {
	_, ctx := newLotteryUploadTestDB(t)

	startedAt := jsonutils.JSONTime(time.Now().Add(-time.Minute))
	_, err := (&domain.LotteryCSVUploadJob{
		JobID:         "job-already-running",
		ProjectID:     101,
		RequestedBy:   "operator@example.com",
		FileName:      "lottery.csv",
		Status:        domain.LOTTERY_CSV_UPLOAD_JOB_STATUS_RUNNING,
		StartedAt:     startedAt,
		TotalRowCount: 1,
	}).Create(ctx)
	assert.NoError(t, err)

	u := lotteryUsecase{
		cfg:                       &config.Config{},
		lotteryRepository:         mocks.NewLotteryApplicationRepository(t),
		lotteryHistoryRepository:  nil,
		projectRepository:         mocks.NewProjectRepository(t),
		userRepository:            mocks.NewUserRepository(t),
		operatorHistoryRepository: nil,
		mailService:               nil,
		txManager:                 nil,
	}

	err = u.ProcessCSVUploadJob(ctx, "job-already-running")
	assert.NoError(t, err)

	job, err := (domain.LotteryCSVUploadJob{}).GetByJobID(ctx, "job-already-running")
	assert.NoError(t, err)
	if assert.NotNil(t, job) {
		assert.Equal(t, domain.LOTTERY_CSV_UPLOAD_JOB_STATUS_RUNNING, job.Status)
		assert.True(t, job.FinishedAt.IsZero())
	}
}
