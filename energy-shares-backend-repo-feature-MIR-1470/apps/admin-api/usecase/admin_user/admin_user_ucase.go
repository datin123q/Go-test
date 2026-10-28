package admin_user

import (
	"bytes"
	"context"
	"encoding/csv"
	"time"

	"github.com/pkg/errors"
	"github.com/samber/lo"
	"github.com/sirupsen/logrus"

	"github.com/illuminarean/energy-shares-backend-repo/pkg/factory"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/idgenerator"

	adomain "github.com/illuminarean/energy-share-core-api/domain"
	cdomain "github.com/illuminarean/energy-shares-backend-repo/clean-arch-domain"
	"github.com/illuminarean/energy-shares-backend-repo/domain"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/jsonutils"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/modelhelper"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/pwhelper"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/s3utils"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/tool"
)

const (
	MENU_ADMIN_USER_LIST     = "menu.manage_operating_system.mange_administrators_list"
	MENU_ADMIN_USER_DETAIL   = "menu.manage_operating_system.mange_administrators_detail"
	MENU_ADMIN_USER_REGISTER = "menu.manage_operating_system.mange_administrators_register"
	MENU_ADMIN_USER_DOWNLOAD = "menu.manage_operating_system.mange_administrators_list"
)

type adminUserUsecase struct {
	logger                         factory.LoggerFactory
	s3pkg                          s3utils.S3
	pwhelper                       pwhelper.PasswordHelper
	txManager                      factory.TxManager
	idGenerator                    idgenerator.IdGenerator
	adminUserRepository            cdomain.AdminUserRepository
	operatorHistoryRepository      cdomain.OperatorHistoryRepository
	passwordResetRequestRepository cdomain.PasswordResetRequestRepository
	refreshTokenRepository         cdomain.RefreshTokenRepository
	roleRepository                 cdomain.RoleRepository
	loginHistoryRepository         cdomain.LoginHistoryRepository
	adminUserService               cdomain.AdminUserService
	mailService                    cdomain.MailService
	authService                    cdomain.AuthService
}

func New(
	logger factory.LoggerFactory,
	s3pkg s3utils.S3,
	pwhelper pwhelper.PasswordHelper,
	txManager factory.TxManager,
	idGenerator idgenerator.IdGenerator,
	adminUserRepository cdomain.AdminUserRepository,
	operatorHistoryRepository cdomain.OperatorHistoryRepository,
	passwordResetRequestRepository cdomain.PasswordResetRequestRepository,
	loginHistoryRepository cdomain.LoginHistoryRepository,
	refreshTokenRepository cdomain.RefreshTokenRepository,
	roleRepository cdomain.RoleRepository,
	adminUserService cdomain.AdminUserService,
	mailService cdomain.MailService,
	authService cdomain.AuthService,
) adomain.AdminUserUsecase {
	return &adminUserUsecase{
		logger:                         logger,
		s3pkg:                          s3pkg,
		pwhelper:                       pwhelper,
		txManager:                      txManager,
		idGenerator:                    idGenerator,
		adminUserRepository:            adminUserRepository,
		operatorHistoryRepository:      operatorHistoryRepository,
		passwordResetRequestRepository: passwordResetRequestRepository,
		loginHistoryRepository:         loginHistoryRepository,
		adminUserService:               adminUserService,
		refreshTokenRepository:         refreshTokenRepository,
		roleRepository:                 roleRepository,
		mailService:                    mailService,
		authService:                    authService,
	}
}

func (a adminUserUsecase) GetAll(ctx context.Context, request adomain.AdminUserGetAllRequest) (adomain.AdminUserGetAllResponse, error) {
	opts := cdomain.AdminUserQueryOptionFunc()
	totalCount, adminUsers, err := a.adminUserRepository.GetAll(ctx,
		opts.WithName(lo.FromPtr(request.Name)),
		opts.WithMobilePhone(lo.FromPtr(request.MobilePhone)),
		opts.WithSortBy(request.Sortby),
		opts.WithOrder(request.Order),
		opts.WithOffset(request.SkipCount),
		opts.WithLimit(request.MaxResultCount),
	)
	if err != nil {
		return adomain.AdminUserGetAllResponse{}, err
	}

	items := lo.Map(adminUsers, func(v domain.AdminUser, _ int) adomain.AdminUserItem {
		v.IsSuperAdmin = tool.StrSliceContains(v.Roles, string(domain.ROLE_SUPER_ADMIN))

		return adomain.AdminUserItem{
			Id:                v.Id,
			CommonId:          v.CommonId,
			Email:             v.Email,
			Name:              v.Name,
			Roles:             v.Roles,
			LastLoginDate:     v.LastLoginDate,
			MobilePhone:       v.MobilePhone,
			IsSuperAdmin:      v.IsSuperAdmin,
			RoleUpdated:       v.RoleUpdated,
			ValidityStartDate: v.ValidityStartDate,
			ValidityEndDate:   v.ValidityEndDate,
		}
	})

	if err = a.operatorHistoryRepository.Create(ctx, domain.OperatorHistory{
		UserId:  request.JWTUser.Id,
		Name:    request.JWTUser.Name,
		Email:   request.JWTUser.Email,
		Menu:    MENU_ADMIN_USER_LIST,
		Ip:      request.ClientIp,
		State:   domain.LOOKUP,
		Created: jsonutils.JSONTime(time.Now()),
	}); err != nil {
		return adomain.AdminUserGetAllResponse{}, errors.Wrap(err, "failed to create operator history")
	}

	return adomain.AdminUserGetAllResponse{
		TotalCount: totalCount,
		Items:      items,
	}, nil
}

func (a adminUserUsecase) GetOne(ctx context.Context, request adomain.AdminUserGetOneRequest) (adomain.AdminUserGetOneResponse, error) {
	v, err := a.adminUserRepository.GetById(ctx, request.Id)
	if err != nil {
		return adomain.AdminUserGetOneResponse{}, errors.Wrap(err, "failed to get admin by id")
	}

	if err = a.operatorHistoryRepository.Create(ctx, domain.OperatorHistory{
		UserId: request.JWTUser.Id,
		Email:  request.JWTUser.Email,
		Name:   request.JWTUser.Name,
		Menu:   MENU_ADMIN_USER_DETAIL,
		Ip:     request.ClientIp,
		State:  domain.LOOKUP,
	}); err != nil {
		return adomain.AdminUserGetOneResponse{}, errors.Wrap(err, "failed to create operator history")
	}

	return adomain.AdminUserGetOneResponse{
		AdminUserItem: adomain.AdminUserItem{
			Id:                v.Id,
			CommonId:          v.CommonId,
			Email:             v.Email,
			Name:              v.Name,
			Roles:             v.Roles,
			LastLoginDate:     v.LastLoginDate,
			MobilePhone:       v.MobilePhone,
			IsSuperAdmin:      v.IsSuperAdmin,
			RoleUpdated:       v.RoleUpdated,
			ValidityStartDate: v.ValidityStartDate,
			ValidityEndDate:   v.ValidityEndDate,
		},
	}, nil
}

func (a adminUserUsecase) Create(ctx context.Context, request adomain.AdminUserCreateRequest) (adomain.AdminUserCreateResponse, error) {
	var (
		uid      = a.idGenerator.GenUuidV4()
		commonId = a.idGenerator.GenUuidV5(uid, request.Email).String()
	)

	// 날짜 형식 변환
	validityStartDate, err := jsonutils.ParseJSONTimeJST(time.DateTime, request.ValidityStartDate)
	if err != nil {
		return adomain.AdminUserCreateResponse{}, errors.Wrap(err, "failed to parse validity start date")
	}

	validityEndDate, err := jsonutils.ParseJSONTimeJST(time.DateTime, request.ValidityEndDate)
	if err != nil {
		return adomain.AdminUserCreateResponse{}, errors.Wrap(err, "failed to parse validity end date")
	}

	// 권한 목록에 ROLE_ADMIN 없으면 추가
	if !tool.StrSliceContains(request.Roles, string(domain.ROLE_ADMIN)) {
		request.Roles = append(request.Roles, string(domain.ROLE_ADMIN))
	}

	// email 중복 체크
	_, err = a.adminUserRepository.GetByEmail(ctx, request.Email)
	if err == nil { // 중복된 이메일 있음
		return adomain.AdminUserCreateResponse{}, errors.Wrap(cdomain.ErrAlreadyExists, "email already exists")
	}
	if err != nil && !errors.Is(err, cdomain.ErrItemNotFound) { // DB 에러
		return adomain.AdminUserCreateResponse{}, errors.Wrap(err, "failed to get admin by email")
	}

	// 어드민 생성
	salt := a.pwhelper.GenerateSalt()
	encodePassword := a.pwhelper.EncodePassword(request.Password, salt)
	newAdminUser := domain.AdminUser{
		CommonId:          commonId,
		Email:             request.Email,
		Name:              request.Name,
		Password:          encodePassword,
		Salt:              salt,
		MobilePhone:       request.MobilePhone,
		LoginFailCnt:      0,
		ValidityStartDate: validityStartDate,
		ValidityEndDate:   validityEndDate,
	}

	admin, err := a.adminUserRepository.Create(ctx, newAdminUser)
	if err != nil {
		return adomain.AdminUserCreateResponse{}, errors.Wrap(err, "failed to create admin user")
	}

	// 권한 등록
	admin.Roles = make([]string, 0)
	admin.Roles = append(admin.Roles, request.Roles...)
	admin.RoleUpdated = jsonutils.JSONTime(time.Now())

	if !modelhelper.Find(request.Roles, string(domain.ROLE_ADMIN)) {
		admin.Roles = append(admin.Roles, string(domain.ROLE_ADMIN))
	}

	for _, role := range admin.Roles {
		if _, err := a.roleRepository.Create(ctx, domain.Role{
			UserId:   admin.Id,
			RoleCode: role,
		}); err != nil {
			return adomain.AdminUserCreateResponse{}, errors.Wrap(err, "failed to create role")
		}
	}

	// 운영 로그 생성
	if err = a.operatorHistoryRepository.Create(ctx, domain.OperatorHistory{
		UserId: request.JWTUser.Id,
		Email:  request.JWTUser.Email,
		Name:   request.JWTUser.Name,
		Menu:   MENU_ADMIN_USER_REGISTER,
		Ip:     request.ClientIp,
		State:  domain.REGISTER,
	}); err != nil {
		return adomain.AdminUserCreateResponse{}, errors.Wrap(err, "failed to create operator history")
	}

	return adomain.AdminUserCreateResponse{
		AdminUserItem: adomain.AdminUserItem{
			Id:                admin.Id,
			CommonId:          admin.CommonId,
			Email:             admin.Email,
			Name:              admin.Name,
			Roles:             admin.Roles,
			LastLoginDate:     admin.LastLoginDate,
			MobilePhone:       admin.MobilePhone,
			IsSuperAdmin:      admin.IsSuperAdmin,
			RoleUpdated:       admin.RoleUpdated,
			ValidityStartDate: admin.ValidityStartDate,
			ValidityEndDate:   admin.ValidityEndDate,
		},
	}, nil
}

func (a adminUserUsecase) Delete(ctx context.Context, request adomain.AdminUserDeleteRequest) (adomain.AdminUserDeleteResponse, error) {

	// 자신을 삭제할 수 없다.
	if request.Id == request.JWTUser.Id {
		return adomain.AdminUserDeleteResponse{}, errors.Wrap(cdomain.ErrBadRequest, "You can't delete yourself.")
	}
	// 최고 관리자는 삭제할 수 없다.
	admin, err := a.adminUserRepository.GetById(ctx, request.Id)
	if err != nil {
		return adomain.AdminUserDeleteResponse{}, errors.Wrap(err, "failed to get admin by id")
	} else if admin.IsSuperAdmin {
		return adomain.AdminUserDeleteResponse{}, errors.Wrap(cdomain.ErrBadRequest, "You can't delete super admin.")
	}

	// 삭제할 운영자의 권한을 삭제
	if err := a.roleRepository.DeleteByUserId(ctx, request.Id); err != nil {
		return adomain.AdminUserDeleteResponse{}, errors.Wrap(err, "failed to delete role with admin id")
	}

	// 운영자 삭제
	if err := a.adminUserRepository.Delete(ctx, request.Id); err != nil {
		return adomain.AdminUserDeleteResponse{}, errors.Wrap(err, "failed to delete admin")
	}

	// 운영 로그 생성
	if err = a.operatorHistoryRepository.Create(ctx, domain.OperatorHistory{
		UserId: request.JWTUser.Id,
		Email:  request.JWTUser.Email,
		Name:   request.JWTUser.Name,
		Title:  admin.Email + " (" + admin.Name + ")",
		Menu:   MENU_ADMIN_USER_DETAIL,
		Ip:     request.ClientIp,
		State:  domain.DELETE,
	}); err != nil {
		return adomain.AdminUserDeleteResponse{}, errors.Wrap(err, "failed to create operator history")
	}

	return adomain.AdminUserDeleteResponse{}, nil
}

func (a adminUserUsecase) Update(ctx context.Context, request adomain.AdminUserUpdateRequest) (adomain.AdminUserUpdateResponse, error) {
	// 날짜 형식 변환
	validityStartDate, err := jsonutils.ParseJSONTimeJST(time.DateTime, request.ValidityStartDate)
	if err != nil {
		return adomain.AdminUserUpdateResponse{}, errors.Wrap(err, "failed to parse validity start date")
	}

	validityEndDate, err := jsonutils.ParseJSONTimeJST(time.DateTime, request.ValidityEndDate)
	if err != nil {
		return adomain.AdminUserUpdateResponse{}, errors.Wrap(err, "failed to parse validity end date")
	}

	// 운영자 정보 조회
	user, err := a.adminUserRepository.GetById(ctx, request.Id)
	if err != nil {
		return adomain.AdminUserUpdateResponse{}, errors.Wrap(err, "failed to get admin by id")
	}

	// email 이 변경되는 경우 중복 체크
	if user.Email != request.Email {
		_, err := a.adminUserRepository.GetByEmail(ctx, request.Email)
		if err == nil { // 중복된 이메일 있음
			return adomain.AdminUserUpdateResponse{}, errors.Wrap(cdomain.ErrAlreadyExists, "email already exists")
		}
		if err != nil && !errors.Is(err, cdomain.ErrItemNotFound) {
			return adomain.AdminUserUpdateResponse{}, errors.Wrap(err, "failed to get admin by email")
		}
	}

	// 최고 관리자: 비밀번호 만료 불가능, 메일 변경 불가능
	// 일반 관리자: 비밀번호 만료 가능, 메일 변경 가능
	if !user.IsSuperAdmin {
		user.Email = request.Email
		user.ValidityStartDate = validityStartDate
		user.ValidityEndDate = validityEndDate

	}
	user.Name = request.Name
	user.MobilePhone = request.MobilePhone

	// 인풋에 비밀번호 값이 있는 경우 비밀번호 수정
	if lo.FromPtr(request.Password) != "" {
		salt := a.pwhelper.GenerateSalt()
		inputPassword := a.pwhelper.EncodePassword(lo.FromPtr(request.Password), salt)

		user.Password = inputPassword
		user.Salt = salt
	}

	// 최고 관리자: 권한 변경 불가능
	// 일반 관리자: 권한 변경 가능
	if !tool.StrSliceContains(user.Roles, string(domain.ROLE_SUPER_ADMIN)) {
		// 기존 권한 삭제 후 새로운 권한 할당
		err := a.roleRepository.DeleteByUserId(ctx, user.Id)
		if err != nil {
			return adomain.AdminUserUpdateResponse{}, errors.Wrap(err, "failed to delete role with admin id")
		}
		user.Roles = make([]string, 0)
		user.Roles = append(user.Roles, request.Roles...)
		user.RoleUpdated = jsonutils.JSONTime(time.Now())

		if !modelhelper.Find(request.Roles, string(domain.ROLE_ADMIN)) {
			user.Roles = append(user.Roles, string(domain.ROLE_ADMIN))
		}

		for _, role := range user.Roles {
			if _, err := a.roleRepository.Create(ctx, domain.Role{
				UserId:   user.Id,
				RoleCode: role,
			}); err != nil {
				return adomain.AdminUserUpdateResponse{}, errors.Wrap(err, "failed to create role")
			}
		}
	}

	// 운영자 정보 업데이트
	if err = a.adminUserRepository.Update(ctx, user); err != nil {
		return adomain.AdminUserUpdateResponse{}, errors.Wrap(err, "failed to update admin user")
	}
	// 운영 로그 생성
	if err = a.operatorHistoryRepository.Create(ctx, domain.OperatorHistory{
		UserId: request.JWTUser.Id,
		Email:  request.JWTUser.Email,
		Name:   request.JWTUser.Name,
		Menu:   MENU_ADMIN_USER_DETAIL,
		Ip:     request.ClientIp,
		State:  domain.CHANGE,
	}); err != nil {
		return adomain.AdminUserUpdateResponse{}, errors.Wrap(err, "failed to create operator history")
	}

	return adomain.AdminUserUpdateResponse{
		AdminUserItem: adomain.AdminUserItem{
			Id:                user.Id,
			CommonId:          user.CommonId,
			Email:             user.Email,
			Name:              user.Name,
			Roles:             user.Roles,
			LastLoginDate:     user.LastLoginDate,
			MobilePhone:       user.MobilePhone,
			IsSuperAdmin:      user.IsSuperAdmin,
			RoleUpdated:       user.RoleUpdated,
			ValidityStartDate: user.ValidityStartDate,
			ValidityEndDate:   user.ValidityEndDate,
		},
	}, nil
}

func (a adminUserUsecase) DownloadHistories(ctx context.Context, request adomain.AdminUserDownloadHistoriesRequest) (adomain.AdminUserDownloadHistoriesResponse, error) {
	opts := cdomain.AdminUserQueryOptionFunc()
	_, adminUsers, err := a.adminUserRepository.GetAll(ctx,
		opts.WithName(lo.FromPtr(request.Name)),
		opts.WithMobilePhone(lo.FromPtr(request.MobilePhone)),
		opts.WithSortBy(request.Sortby),
		opts.WithOrder(request.Order),
		opts.WithOffset(request.SkipCount),
		opts.WithLimit(request.MaxResultCount),
	)
	if err != nil {
		return adomain.AdminUserDownloadHistoriesResponse{}, errors.Wrap(err, "failed to get all admin users")
	}

	// 데이터를 엑셀 파일로 변환
	buf := new(bytes.Buffer)
	w := csv.NewWriter(buf)

	csvData := a.adminUserService.ExportCSV(ctx, adminUsers)
	if err = w.WriteAll(csvData); err != nil {
		return adomain.AdminUserDownloadHistoriesResponse{}, errors.Wrap(err, "failed to write csv")
	}
	w.Flush()

	// 운영 로그 생성
	a.operatorHistoryRepository.Create(ctx, domain.OperatorHistory{
		UserId: request.JWTUser.Id,
		Email:  request.JWTUser.Email,
		Name:   request.JWTUser.Name,
		Menu:   MENU_ADMIN_USER_DOWNLOAD,
		Ip:     request.ClientIp,
		State:  domain.DOWNLOAD,
	})

	return adomain.AdminUserDownloadHistoriesResponse{Data: buf.Bytes()}, nil
}

func (a adminUserUsecase) Login(ctx context.Context, request adomain.AdminUserLoginRequest) (adomain.AdminUserLoginResponse, error) {
	// 이메일로 어드민 조회
	adminUser, err := a.adminUserRepository.GetByEmail(ctx, request.Email)
	if err != nil {
		return adomain.AdminUserLoginResponse{}, errors.Wrap(err, "failed to get admin user by email")
	}

	// 비밀번호 유효기간 검사
	valid, err := a.adminUserService.IsValidPeriod(ctx, adminUser)
	if err != nil {
		return adomain.AdminUserLoginResponse{}, errors.Wrap(err, "failed to check validity period")
	}
	if !valid {
		return adomain.AdminUserLoginResponse{}, cdomain.ErrAlreadyExpired
	}

	// 비밀번호 검사
	inputPassword := a.pwhelper.EncodePassword(request.Password, adminUser.Salt)
	if !a.pwhelper.IsSamePassword(adminUser.Password, inputPassword) {
		// 비밀번호 불일치 - 로그인 실패 카운트 업
		if err := a.txManager.WithTx(ctx, func(txCtx context.Context) error {
			adminUser.LoginFailCnt++
			if err := a.adminUserRepository.Update(txCtx, adminUser); err != nil {
				return err
			}
			return nil
		}); err != nil {
			return adomain.AdminUserLoginResponse{}, errors.Wrap(err, "failed to update login fail count")
		}

		// 비밀번호 불일치 - 로그
		a.logger.Entry(ctx).
			WithFields(logrus.Fields{
				"adminUserId": adminUser.Id,
				"ip":          request.ClientIp,
				"agent":       request.UserAgent,
				"failCount":   adminUser.LoginFailCnt,
			}).
			Info("admin user login failed")
		return adomain.AdminUserLoginResponse{}, cdomain.ErrBadRequest
	}

	// 비밀번호 일치 - Access Token 생성
	accessToken, err := a.authService.CreateAdminAccessToken(ctx, adminUser)
	if err != nil {
		return adomain.AdminUserLoginResponse{}, errors.Wrap(err, "failed to create access token")
	}

	// 비밀번호 일치 - Refresh Token 생성
	refreshToken, err := a.authService.CreateAdminRefreshToken(ctx, cdomain.AdminRefreshTokenInput{
		Admin: adminUser,
		IP:    request.ClientIp,
		Agent: request.UserAgent,
	})
	if err != nil {
		return adomain.AdminUserLoginResponse{}, errors.Wrap(err, "failed to create refresh token")
	}

	// 비밀번호 일치 - CDN 캐시 제어를 위한 쿠키 생성
	cookies, err := a.s3pkg.CreateCookieSign()
	if err != nil {
		return adomain.AdminUserLoginResponse{}, errors.Wrap(err, "failed to get cookie sign")
	}

	// 비밀번호 일치 - 어드민 로그인 정보 업데이트
	adminUser.LoginFailCnt = 0
	adminUser.LastLoginDate = jsonutils.JSONTime(time.Now())
	if err := a.adminUserRepository.Update(ctx, adminUser); err != nil {
		return adomain.AdminUserLoginResponse{}, errors.Wrap(err, "failed to update admin user")
	}

	// 운영 로그 생성
	if err := a.operatorHistoryRepository.Create(ctx, domain.OperatorHistory{
		UserId: adminUser.Id,
		Name:   adminUser.Name,
		Email:  adminUser.Email,
		Menu:   "",
		Ip:     request.ClientIp,
		Title:  "",
		State:  domain.LOGIN,
	}); err != nil {
		return adomain.AdminUserLoginResponse{}, errors.Wrap(err, "failed to create operator history")
	}

	return adomain.AdminUserLoginResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		Cookies:      cookies,
	}, nil

}

func (a adminUserUsecase) Logout(ctx context.Context, request adomain.AdminUserLogoutRequest) (adomain.AdminUserLogoutResponse, error) {
	// TODO : Blacklist 이용한 로그아웃 처리 & 프론트에서 로그아웃이 재대로 처리 되지 않는것으로 보임

	// Refresh Token 조회
	refreshToken, err := a.refreshTokenRepository.GetByToken(ctx, request.RefreshToken)
	if err != nil {
		return adomain.AdminUserLogoutResponse{}, errors.Wrap(err, "failed to get refresh token from database")
	}

	// Refresh Token 만료 처리
	refreshToken.ExpireIn = time.Unix(0, 0).Unix()
	if err := a.refreshTokenRepository.Update(ctx, refreshToken); err != nil {
		return adomain.AdminUserLogoutResponse{}, errors.Wrap(err, "failed to update refresh token to expire")
	}

	// 운영 로그 생성
	if err := a.operatorHistoryRepository.Create(ctx, domain.OperatorHistory{
		UserId: request.JWTUser.Id,
		Name:   request.JWTUser.Name,
		Email:  request.JWTUser.Email,
		Menu:   "",
		Ip:     request.ClientIp,
		Title:  "",
		State:  domain.LOGOUT,
	}); err != nil {
		return adomain.AdminUserLogoutResponse{}, errors.Wrap(err, "failed to create operator history")
	}

	return adomain.AdminUserLogoutResponse{}, nil
}
