package domain

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"time"

	"github.com/illuminarean/energy-shares-backend-repo/pkg/factory"
)

// Sha256Hex はGmoAozoraOAuthStateのlookup/state_hash計算に使う共通ヘルパー。
// admin-api(lookup)とjobs-scheduler(lookup・state_hashの両方)の両方から
// 使うため、どちらのアプリからもimportできるdomainパッケージに置く。
func Sha256Hex(s string) string {
	sum := sha256.Sum256([]byte(s))
	return hex.EncodeToString(sum[:])
}

// GmoAozoraOAuthState は、GMOあおぞら銀行のOAuth認可コールバック
// (/gmo-aozora/auth/access-token)の戻り先を検証するための使い捨てstateを
// 保持する(MIR-1298)。stateを照合していなかったため、外から呼んで本番の
// トークンを差し替えられる不具合の修正。
//
// LookupとStateHashは別の値。admin-apiがcrypto/randで作る生値をSとすると、
// Lookup = sha256(S)。ライブラリ(gmo-aozora-api-go)がURLへ実際に載せるstate値
// (sha256(S+salt)、Hと呼ぶ)を、StateHash = sha256(H) として書き戻す。
// StateHashをHそのものにしないのは、DBを読める人(Metabaseの利用者は
// 全権限)が有効期間中のHを読んで、正規のログインへ割り込めないようにする
// ため。2段になっているのはライブラリの実装(ganb.go の getState)に合わせた結果。
type GmoAozoraOAuthState struct {
	Id          int64      `json:"-" xorm:"pk autoincr"`
	Lookup      string     `json:"-" xorm:"lookup"`
	StateHash   *string    `json:"-" xorm:"state_hash"`
	Nonce       *string    `json:"-" xorm:"nonce"`
	AdminUserId int64      `json:"-" xorm:"admin_user_id"`
	ExpiresAt   time.Time  `json:"-" xorm:"expires_at"`
	UsedAt      *time.Time `json:"-" xorm:"used_at"`
	Result      *string    `json:"-" xorm:"result"`
	Created     time.Time  `json:"-" xorm:"created"`
}

func (GmoAozoraOAuthState) TableName() string {
	return "gmo_aozora_oauth_state"
}

const (
	GmoAozoraOAuthResultSuccess = "SUCCESS"
	GmoAozoraOAuthResultFailure = "FAILURE"
)

// Create はadmin-apiのGmoConnectUrlが呼ぶ。lookupはSのsha256。
func (GmoAozoraOAuthState) Create(ctx context.Context, lookup string, adminUserId int64, expiresAt time.Time) error {
	row := GmoAozoraOAuthState{
		Lookup:      lookup,
		AdminUserId: adminUserId,
		ExpiresAt:   expiresAt.UTC(),
	}
	_, err := factory.DB(ctx).Insert(&row)
	return err
}

// GetByLookup は、jobs-schedulerのGetLoginがadmin-api発行のstate(S)から行を
// 引くのに使う。未使用・期限内であることは呼び出し側で確認する。
func (GmoAozoraOAuthState) GetByLookup(ctx context.Context, lookup string) (*GmoAozoraOAuthState, error) {
	var row GmoAozoraOAuthState
	has, err := factory.DB(ctx).Where("lookup = ?", lookup).Get(&row)
	if err != nil {
		return nil, err
	}
	if !has {
		return nil, nil
	}
	return &row, nil
}

// SetStateHashOnce は、ライブラリが実際に発行したURLから取り出したHのハッシュと
// nonceを1回だけ書き戻す。WHERE state_hash IS NULLの条件付きUPDATEなので、
// GetLoginが同じstateで2回呼ばれても2回目は書き戻さない
// (2回書き戻すとnonceが上書きされ、最初のURLでのログインが壊れるため)。
func (GmoAozoraOAuthState) SetStateHashOnce(ctx context.Context, id int64, stateHash, nonce string) (bool, error) {
	res, err := factory.DB(ctx).Exec(
		"UPDATE gmo_aozora_oauth_state SET state_hash = ?, nonce = ? WHERE id = ? AND state_hash IS NULL",
		stateHash, nonce, id,
	)
	if err != nil {
		return false, err
	}
	affected, err := res.RowsAffected()
	if err != nil {
		return false, err
	}
	return affected == 1, nil
}

// ClaimByStateHash は、コールバックが受け取ったstateのハッシュから引いた行を、
// 未使用・期限内であれば1回だけ使用済みにする。更新できた場合のみ行を返す
// (code交換を1回だけ行うゲートを兼ねる。同じstateで2回同時に来ても片方だけが通る)。
func (GmoAozoraOAuthState) ClaimByStateHash(ctx context.Context, stateHash string, now time.Time) (*GmoAozoraOAuthState, error) {
	res, err := factory.DB(ctx).Exec(
		"UPDATE gmo_aozora_oauth_state SET used_at = ? WHERE state_hash = ? AND used_at IS NULL AND expires_at > ?",
		now.UTC(), stateHash, now.UTC(),
	)
	if err != nil {
		return nil, err
	}
	affected, err := res.RowsAffected()
	if err != nil {
		return nil, err
	}
	if affected != 1 {
		return nil, nil
	}
	return GmoAozoraOAuthState{}.GetByStateHash(ctx, stateHash)
}

// GetByStateHash は、同じstateでのリロード(F5・二重送信)時に、既に処理済みの
// 行から結果を読み直すのに使う。
func (GmoAozoraOAuthState) GetByStateHash(ctx context.Context, stateHash string) (*GmoAozoraOAuthState, error) {
	var row GmoAozoraOAuthState
	has, err := factory.DB(ctx).Where("state_hash = ?", stateHash).Get(&row)
	if err != nil {
		return nil, err
	}
	if !has {
		return nil, nil
	}
	return &row, nil
}

// RecordResult はベストエフォートで結果区分を残す。失敗しても認可フロー自体は
// 失敗させない(リロード時の表示用の付随情報のため)。
func (GmoAozoraOAuthState) RecordResult(ctx context.Context, id int64, result string) error {
	_, err := factory.DB(ctx).Exec(
		"UPDATE gmo_aozora_oauth_state SET result = ? WHERE id = ?",
		result, id,
	)
	return err
}
