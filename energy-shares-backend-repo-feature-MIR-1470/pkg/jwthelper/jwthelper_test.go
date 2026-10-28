package jwthelper

import (
	"net/http"
	"net/http/httptest"
	"reflect"
	"strconv"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"

	"github.com/illuminarean/energy-shares-backend-repo/domain"
)

func TestTokenToUser(t *testing.T) {
	type args struct {
		c echo.Context
	}
	tests := []struct {
		name    string
		args    args
		want    int64
		want1   string
		wantErr bool
	}{
		{
			name: "SuccessToGetUser",
			args: args{
				c: func() echo.Context {

					tokenMap := map[string]interface{}{
						"iss":   domain.SERVICE_NAME,
						"aud":   domain.SERVICE_NAME,
						"user":  strconv.Itoa(int(1)),
						"email": "admin@energyshares.jp",
						"role":  []string{"ROLE_ADMIN", "ROLE_PRODUCT", "ROLE_OPERATOR", "ROLE_USER", "ROLE_PAYMENT", "ROLE_INVESTOR", "ROLE_MESSENGER", "ROLE_REPORT", "ROLE_SUPER_ADMIN"},
					}

					jwtSigningMethod := jwt.SigningMethodHS256

					expDuration := time.Hour * 24

					claims := jwt.MapClaims{
						"nbf": time.Now().Unix(),
						"exp": time.Now().Add(expDuration).Unix(),
					}
					for k, v := range tokenMap {
						claims[k] = v
					}
					token := jwt.NewWithClaims(jwtSigningMethod, claims)

					req := httptest.NewRequest(http.MethodGet, "/", nil)
					rec := httptest.NewRecorder()
					e := echo.New()
					ctx := e.NewContext(req, rec)
					ctx.Set("user", token)

					return ctx
				}(),
			},
			want:    1,
			want1:   "admin@energyshares.jp",
			wantErr: false,
		},
		{
			name: "Success2ToGetUser",
			args: args{
				c: func() echo.Context {

					tokenMap := map[string]interface{}{
						"iss":   domain.SERVICE_NAME,
						"aud":   domain.SERVICE_NAME,
						"user":  strconv.Itoa(int(2)),
						"email": "user@energyshares.jp",
						"role":  []string{"ROLE_ADMIN", "ROLE_PRODUCT", "ROLE_OPERATOR", "ROLE_USER", "ROLE_PAYMENT", "ROLE_INVESTOR", "ROLE_MESSENGER", "ROLE_REPORT", "ROLE_SUPER_ADMIN"},
					}

					jwtSigningMethod := jwt.SigningMethodHS256

					expDuration := time.Hour * 24

					claims := jwt.MapClaims{
						"nbf": time.Now().Unix(),
						"exp": time.Now().Add(expDuration).Unix(),
					}
					for k, v := range tokenMap {
						claims[k] = v
					}
					token := jwt.NewWithClaims(jwtSigningMethod, claims)

					req := httptest.NewRequest(http.MethodGet, "/", nil)
					rec := httptest.NewRecorder()
					e := echo.New()
					ctx := e.NewContext(req, rec)
					ctx.Set("user", token)

					return ctx
				}(),
			},
			want:    2,
			want1:   "user@energyshares.jp",
			wantErr: false,
		},
		{
			name: "FailedToGetUser",
			args: args{
				c: func() echo.Context {

					req := httptest.NewRequest(http.MethodGet, "/", nil)
					rec := httptest.NewRecorder()
					e := echo.New()
					ctx := e.NewContext(req, rec)

					return ctx
				}(),
			},

			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1, err := TokenToUser(tt.args.c)
			if (err != nil) != tt.wantErr {
				t.Errorf("TokenToUser() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("TokenToUser() got = %v, want %v", got, tt.want)
			}
			if got1 != tt.want1 {
				t.Errorf("TokenToUser() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}

func Test_TokenToUserExtended(t *testing.T) {
	type args struct {
		c echo.Context
	}
	tests := []struct {
		name    string
		args    args
		want    int64
		want1   string
		want2   string
		wantErr bool
	}{
		{
			name: "SuccessToGetUserExtended",
			args: args{
				c: func() echo.Context {

					tokenMap := map[string]interface{}{
						"iss":   domain.SERVICE_NAME,
						"aud":   domain.SERVICE_NAME,
						"user":  strconv.Itoa(int(1)),
						"name":  "admin",
						"email": "admin@energyshares.jp",
						"role":  []string{"ROLE_ADMIN", "ROLE_PRODUCT", "ROLE_OPERATOR", "ROLE_USER", "ROLE_PAYMENT", "ROLE_INVESTOR", "ROLE_MESSENGER", "ROLE_REPORT", "ROLE_SUPER_ADMIN"},
					}

					jwtSigningMethod := jwt.SigningMethodHS256

					expDuration := time.Hour * 24

					claims := jwt.MapClaims{
						"nbf": time.Now().Unix(),
						"exp": time.Now().Add(expDuration).Unix(),
					}
					for k, v := range tokenMap {
						claims[k] = v
					}
					token := jwt.NewWithClaims(jwtSigningMethod, claims)

					req := httptest.NewRequest(http.MethodGet, "/", nil)
					rec := httptest.NewRecorder()
					e := echo.New()
					ctx := e.NewContext(req, rec)
					ctx.Set("user", token)

					return ctx
				}(),
			},
			want:    1,
			want1:   "admin",
			want2:   "admin@energyshares.jp",
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1, got2, err := TokenToUserExtended(tt.args.c)
			if (err != nil) != tt.wantErr {
				t.Errorf("TokenToUserExtended() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("TokenToUserExtended() got = %v, want %v", got, tt.want)
			}
			if got1 != tt.want1 {
				t.Errorf("TokenToUserExtended() got1 = %v, want %v", got1, tt.want1)
			}
			if got2 != tt.want2 {
				t.Errorf("TokenToUserExtended() got2 = %v, want %v", got2, tt.want2)
			}
		})
	}
}

func TestTokenToRoleMap(t *testing.T) {
	type args struct {
		c echo.Context
	}
	tests := []struct {
		name    string
		args    args
		want    map[string]int
		wantErr bool
	}{
		{
			name: "SuccessToGetRoleMap",
			args: args{
				c: func() echo.Context {

					tokenMap := map[string]interface{}{
						"iss":   domain.SERVICE_NAME,
						"aud":   domain.SERVICE_NAME,
						"user":  strconv.Itoa(int(1)),
						"email": "admin@energyshares.jp",
						"role":  []string{"ROLE_ADMIN", "ROLE_PRODUCT", "ROLE_OPERATOR", "ROLE_USER", "ROLE_PAYMENT", "ROLE_INVESTOR", "ROLE_MESSENGER", "ROLE_REPORT", "ROLE_SUPER_ADMIN"},
					}

					jwtSigningMethod := jwt.SigningMethodHS256

					expDuration := time.Hour * 24

					claims := jwt.MapClaims{
						"nbf": time.Now().Unix(),
						"exp": time.Now().Add(expDuration).Unix(),
					}
					for k, v := range tokenMap {
						claims[k] = v
					}
					token := jwt.NewWithClaims(jwtSigningMethod, claims)

					req := httptest.NewRequest(http.MethodGet, "/", nil)
					rec := httptest.NewRecorder()
					e := echo.New()
					ctx := e.NewContext(req, rec)
					ctx.Set("user", token)
					return ctx
				}(),
			},
			want: map[string]int{
				"ROLE_ADMIN":       1,
				"ROLE_PRODUCT":     1,
				"ROLE_OPERATOR":    1,
				"ROLE_USER":        1,
				"ROLE_PAYMENT":     1,
				"ROLE_INVESTOR":    1,
				"ROLE_MESSENGER":   1,
				"ROLE_REPORT":      1,
				"ROLE_SUPER_ADMIN": 1,
			},
			wantErr: false,
		},
		{
			name: "SuccessToGetRoleMap2",
			args: args{
				c: func() echo.Context {

					tokenMap := map[string]interface{}{
						"iss":   domain.SERVICE_NAME,
						"aud":   domain.SERVICE_NAME,
						"user":  strconv.Itoa(int(1)),
						"email": "admin@energyshares.jp",
						"role":  []string{"ROLE_ADMIN", "ROLE_OPERATOR", "ROLE_USER", "ROLE_PAYMENT", "ROLE_INVESTOR", "ROLE_MESSENGER", "ROLE_REPORT", "ROLE_SUPER_ADMIN"},
					}

					jwtSigningMethod := jwt.SigningMethodHS256

					expDuration := time.Hour * 24

					claims := jwt.MapClaims{
						"nbf": time.Now().Unix(),
						"exp": time.Now().Add(expDuration).Unix(),
					}
					for k, v := range tokenMap {
						claims[k] = v
					}
					token := jwt.NewWithClaims(jwtSigningMethod, claims)

					req := httptest.NewRequest(http.MethodGet, "/", nil)
					rec := httptest.NewRecorder()
					e := echo.New()
					ctx := e.NewContext(req, rec)
					ctx.Set("user", token)
					return ctx
				}(),
			},
			want: map[string]int{
				"ROLE_ADMIN":       1,
				"ROLE_OPERATOR":    1,
				"ROLE_USER":        1,
				"ROLE_PAYMENT":     1,
				"ROLE_INVESTOR":    1,
				"ROLE_MESSENGER":   1,
				"ROLE_REPORT":      1,
				"ROLE_SUPER_ADMIN": 1,
			},
			wantErr: false,
		},
		{
			name: "FailToGetRoleMap",
			args: args{
				c: func() echo.Context {

					req := httptest.NewRequest(http.MethodGet, "/", nil)
					rec := httptest.NewRecorder()
					e := echo.New()
					ctx := e.NewContext(req, rec)
					return ctx
				}(),
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := TokenToRoleMap(tt.args.c)
			if (err != nil) != tt.wantErr {
				t.Errorf("TokenToRoleMap() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("TokenToRoleMap() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestValidateUserToken(t *testing.T) {
	type args struct {
		c      echo.Context
		userId int64
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "SuccessToValidateUserToken",
			args: args{

				c: func() echo.Context {

					tokenMap := map[string]interface{}{
						"iss":   domain.SERVICE_NAME,
						"aud":   domain.SERVICE_NAME,
						"user":  strconv.Itoa(int(6)),
						"email": "pjy1418@gmail.com",
						"role":  []string{"ROLE_USER"},
					}

					jwtSigningMethod := jwt.SigningMethodHS256

					expDuration := time.Minute * 15

					claims := jwt.MapClaims{
						"nbf": time.Now().Unix(),
						"exp": time.Now().Add(expDuration).Unix(),
					}
					for k, v := range tokenMap {
						claims[k] = v
					}
					token := jwt.NewWithClaims(jwtSigningMethod, claims)

					req := httptest.NewRequest(http.MethodGet, "/", nil)
					rec := httptest.NewRecorder()
					e := echo.New()
					ctx := e.NewContext(req, rec)
					ctx.Set("user", token)
					return ctx
				}(),
				userId: 6,
			},
			wantErr: false,
		},
		{
			name: "SuccessToValidateUserToken",
			args: args{

				c: func() echo.Context {

					tokenMap := map[string]interface{}{
						"iss":   domain.SERVICE_NAME,
						"aud":   domain.SERVICE_NAME,
						"user":  strconv.Itoa(int(6)),
						"email": "pjy1418@gmail.com",
						"role":  []string{"ROLE_USER"},
					}

					jwtSigningMethod := jwt.SigningMethodHS256

					expDuration := time.Minute * 15

					claims := jwt.MapClaims{
						"nbf": time.Now().Unix(),
						"exp": time.Now().Add(expDuration).Unix(),
					}
					for k, v := range tokenMap {
						claims[k] = v
					}
					token := jwt.NewWithClaims(jwtSigningMethod, claims)

					req := httptest.NewRequest(http.MethodGet, "/", nil)
					rec := httptest.NewRecorder()
					e := echo.New()
					ctx := e.NewContext(req, rec)
					ctx.Set("user", token)
					return ctx
				}(),
				userId: 10,
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := ValidateUserToken(tt.args.c, tt.args.userId); (err != nil) != tt.wantErr {
				t.Errorf("ValidateUserToken() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
