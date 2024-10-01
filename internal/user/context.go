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

func MustUserInfo(ctx context.Context) *UserInfo {
	if u, found := UserInfoFromContext(ctx); found {
		return u
	}
	log(ctx).Fatalf("User info not found in context")
	return nil
}

func WithUserInfo(ctx context.Context, user *UserInfo) context.Context {
	return context.WithValue(ctx, userInfoKey, user)
}
