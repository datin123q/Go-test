package actorcontext

import (
	"context"

	"github.com/illuminarean/energy-shares-backend-repo/pkg/jwthelper"
)

type actorKey struct{}

// WithActor は JWT ユーザー（アクター）をコンテキストに格納し、
// 下位レイヤー（リポジトリデコレーターなど）が操作の実行者を取得できるようにします。
func WithActor(ctx context.Context, user jwthelper.User) context.Context {
	return context.WithValue(ctx, actorKey{}, user)
}

// ActorFrom は WithActor で格納されたアクターを取得します。
// アクターが設定されていない場合は (ゼロ値の User, false) を返します。
func ActorFrom(ctx context.Context) (jwthelper.User, bool) {
	user, ok := ctx.Value(actorKey{}).(jwthelper.User)
	return user, ok
}
