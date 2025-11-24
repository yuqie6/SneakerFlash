package handler

import (
	"SneakerFlash/internal/model"
	"time"
)

// MessageResponse 用于返回简单提示信息。
type MessageResponse struct {
	Message string `json:"message"`
}

// TokenPairResponse 用于登录和刷新 token 响应。
type TokenPairResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token,omitempty"`
	ExpiresIn    int    `json:"expires_in"`
}

// AccessTokenResponse 仅返回新的 access token。
type AccessTokenResponse struct {
	AccessToken string `json:"access_token"`
	ExpiresIn   int    `json:"expires_in"`
}

// IDResponse 返回资源主键。
type IDResponse struct {
	ID int `json:"id"`
}

// OrderNumResponse 返回订单号。
type OrderNumResponse struct {
	OrderNum string `json:"order_num"`
}

// UploadURLResponse 返回上传后资源地址。
type UploadURLResponse struct {
	URL string `json:"url"`
}

// UserResponse 用户信息输出。
type UserResponse struct {
	ID        uint      `json:"ID"`
	CreatedAt time.Time `json:"CreatedAt"`
	UpdatedAt time.Time `json:"UpdatedAt"`
	Username  string    `json:"username"`
	Balance   float64   `json:"balance"`
	Avatar    string    `json:"avatar"`
}

// PaymentResponse 用于描述支付单输出，避免暴露内部 gorm.Model。
type PaymentResponse struct {
	ID          uint                `json:"ID"`
	CreatedAt   time.Time           `json:"CreatedAt"`
	UpdatedAt   time.Time           `json:"UpdatedAt"`
	OrderID     uint                `json:"order_id"`
	PaymentID   string              `json:"payment_id"`
	AmountCents int64               `json:"amount_cents"`
	Status      model.PaymentStatus `json:"status"`
	NotifyData  string              `json:"notify_data"`
}

// ProductListResponse 商品列表响应。
type ProductListResponse struct {
	Items []model.Product `json:"items"`
	Total int64           `json:"total"`
	Page  int             `json:"page"`
}

// ProductListWithSizeResponse 包含分页大小的商品列表响应。
type ProductListWithSizeResponse struct {
	Items []model.Product `json:"items"`
	Total int64           `json:"total"`
	Page  int             `json:"page"`
	Size  int             `json:"size"`
}

// OrderListResponse 订单列表响应。
type OrderListResponse struct {
	Items []model.Order `json:"items"`
	Total int64         `json:"total"`
	Page  int           `json:"page"`
	Size  int           `json:"size"`
}

// OrderWithPaymentResponse 订单与支付单组合响应。
type OrderWithPaymentResponse struct {
	Order   *model.Order     `json:"order"`
	Payment *PaymentResponse `json:"payment,omitempty"`
}
