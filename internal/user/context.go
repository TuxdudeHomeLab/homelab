package user

import "context"

var (
	userInfoKey = ctxKeyUserInfo{}
)

type ctxKeyUserInfo struct{}

func UserInfoFromContext(ctx context.Context) (*UserInfo, bool) {
	user, ok := ctx.Value(userInfoKey).(*UserInfo)
	return user, ok
}

func WithUserInfo(ctx context.Context, user *UserInfo) context.Context {
	return context.WithValue(ctx, userInfoKey, user)
}
