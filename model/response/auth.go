package response

// LoginResponse 登录响应。refresh token 只通过 HttpOnly Cookie 下发，不放进响应体。
type LoginResponse struct {
	AccessToken          string    `json:"access_token"`
	AccessTokenExpiresAt int64     `json:"access_token_expires_at"` // 毫秒时间戳
	User                 *UserInfo `json:"user"`
}

// UserInfo 用户信息
type UserInfo struct {
	ID       uint64 `json:"id"`
	Username string `json:"username"`
	Nickname string `json:"nickname"`
	Avatar   string `json:"avatar"`
	Email    string `json:"email"`
	Role     int8   `json:"role"`
}
