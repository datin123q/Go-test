package gmo_aozora

import (
	"context"

	ganblib "github.com/gmoaozora/gmo-aozora-api-go"
	"github.com/pkg/errors"
	"xorm.io/xorm"

	"github.com/illuminarean/energy-shares-backend-repo/domain"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/factory"
)

// DetachedContext は、担当者がポップアップを閉じる・ALBのタイムアウト等で
// 親ctxがキャンセルされても、既に取得済みのトークンを保存する処理を最後まで
// 走らせるためのもの(MIR-1298)。①敵対的検証の指摘: context.WithoutCancel
// だけでは不十分——DB呼び出しは`factory.DB(ctx)`が読む*xorm.Session(HTTP
// ミドルウェアが元のctxを閉じ込めて保持している、xorm@v1.3.9の
// session_raw.goのsession.ctx)を経由するため、渡したctxのWithoutCancelは
// SQL実行には効かない。jobs-scheduler/jobs/job_execution_history.goの
// jobExecutionHistoryContextと同じ手当てで、DBの値をSessionからその親の
// Engineへ差し替える(Engineはリクエストのctxを閉じ込めない)。
func DetachedContext(ctx context.Context) context.Context {
	ctx = context.WithoutCancel(ctx)
	if session, ok := ctx.Value(factory.ConextDBType).(*xorm.Session); ok {
		ctx = context.WithValue(ctx, factory.ConextDBType, session.Engine())
	}
	return ctx
}

// validateToken は、ライブラリ(gmo-aozora-api-go)の応答をそのまま信用しない
// (MIR-1298)。OpenIDGetToken/RefreshTokensは、応答のJSON解析に失敗すると
// errors.Wrap(err, ...)のerrがnilなので空のトークンとエラー無しを返すことが
// あり、RefreshTokensはHTTPの状態コードを見ないため
// {"error":"invalid_grant"}のような応答でも解析自体は成功する。トークンの
// 中身を見て初めて判定できる。
func validateToken(token ganblib.Token) error {
	if token.Error != "" {
		return errors.Errorf("gmo aozora token error: %s", token.Error)
	}
	if token.AccessToken == "" {
		return errors.New("gmo aozora token: access token is empty")
	}
	if token.RefreshToken == "" {
		return errors.New("gmo aozora token: refresh token is empty")
	}
	return nil
}

// updateStoredToken はgmo_aozora_client(id=1)へトークンを保存する。
// GmoAozoraClient.Updateは渡した構造体のAccessToken/RefreshTokenを暗号文へ
// 書き換えるため、再試行のたびにDBから読み直して構造体を作り直す
// (失敗した試行の構造体を使い回すと、既に暗号化された値を二重に暗号化する)。
// 更新件数が0件の場合も失敗として扱う。
func updateStoredToken(ctx context.Context, token ganblib.Token) error {
	const maxAttempts = 3
	var lastErr error
	for attempt := 1; attempt <= maxAttempts; attempt++ {
		client, err := domain.GmoAozoraClient{}.GetByID(ctx, 1)
		if err != nil {
			lastErr = err
			continue
		}
		if client == nil {
			return errors.New("gmo aozora client row not found")
		}
		client.AccessToken = token.AccessToken
		client.RefreshToken = token.RefreshToken

		affected, err := client.Update(ctx)
		if err != nil {
			lastErr = err
			continue
		}
		if affected != 1 {
			lastErr = errors.Errorf("gmo aozora client update affected %d rows, want 1", affected)
			continue
		}
		return nil
	}
	return errors.Wrap(lastErr, "fail client model update")
}

// refreshTokenWithRetryOnRejectionは、銀行がrefresh_tokenを拒否した
// (token.Errorが空でない)場合に、DBから読み直して1回だけ再試行する
// (MIR-1298)。デプロイ中は新旧2つのECSタスクが並行するため、このタスクが
// 読み込んだrefresh_tokenが、既に他タスクの更新で古くなっている場合を
// 救うため。読み直した値が最初に使った値と同じなら、再試行しても無駄なので
// そのまま返す。
func refreshTokenWithRetryOnRejection(ctx context.Context, refreshToken string) (ganblib.Token, error) {
	token, err := callRefreshTokens(refreshToken)
	if err != nil {
		return token, err
	}
	if token.Error == "" {
		return token, nil
	}

	client, dbErr := domain.GmoAozoraClient{}.GetByID(ctx, 1)
	if dbErr != nil || client == nil || client.RefreshToken == "" || client.RefreshToken == refreshToken {
		return token, nil
	}
	return callRefreshTokens(client.RefreshToken)
}

// callRefreshTokensは変数にして、テストが実際の銀行API呼び出しを介さずに
// 差し替えられるようにする(既存のgetAccountInfoClient等と同じ間接呼び出しの方針)。
var callRefreshTokens = func(refreshToken string) (ganblib.Token, error) {
	ganb, err := ganblib.New(ClientId(), ClientSecret(), nonceSave, nonceCheckAndDelete)
	if err != nil {
		return ganblib.Token{}, errors.Wrap(err, "fail create ganb")
	}
	return ganb.RefreshTokens(refreshToken)
}
