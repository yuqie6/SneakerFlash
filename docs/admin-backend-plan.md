# Admin 后端实施方案

> 本文档描述管理后台所需的全部后端改动。前端已就绪，部署后端后即可连通。

---

## 1. Model 变更

### User 新增 `role` 字段

```go
// internal/model/user.go
type User struct {
    gorm.Model
    Username        string  `gorm:"type:varchar(50);unique;not null" json:"username"`
    Password        string  `gorm:"type:varchar(100);not null" json:"-"`
    Balance         float64 `gorm:"type:decimal(10,2);default:0;not null" json:"balance"`
    Avatar          string  `gorm:"type:varchar(255);default:''" json:"avatar"`
    TotalSpentCents int64   `gorm:"type:bigint;default:0;not null" json:"total_spent_cents"`
    GrowthLevel     int     `gorm:"type:int;default:1;not null" json:"growth_level"`
    Role            string  `gorm:"type:varchar(20);default:'user';not null" json:"role"` // 新增
}
```

### Migration

```sql
ALTER TABLE users ADD COLUMN role VARCHAR(20) NOT NULL DEFAULT 'user';
```

### 手动提权

```sql
UPDATE users SET role = 'admin' WHERE username = '<你的管理员用户名>';
```

---

## 2. JWT Claims 扩展

```go
// internal/pkg/utils/jwt.go
type Claims struct {
    UserID    uint   `json:"user_id"`
    Username  string `json:"username"`
    Role      string `json:"role"`       // 新增
    TokenType string `json:"token_type"`
    jwt.RegisteredClaims
}
```

`GenerateToken` 函数签名变更：

```go
func GenerateToken(userID uint, username, role, tokenType string) (string, error)
```

Login handler 调用时传入 `user.Role`。

---

## 3. AdminAuth Middleware

新建 `internal/middlerware/admin.go`：

```go
package middlerware

import (
    "SneakerFlash/internal/pkg/app"
    "SneakerFlash/internal/pkg/e"
    "net/http"
    "github.com/gin-gonic/gin"
)

func AdminAuth() gin.HandlerFunc {
    return func(c *gin.Context) {
        appG := app.Gin{C: c}
        role, exists := c.Get("role")
        if !exists || role != "admin" {
            appG.ErrorMsg(http.StatusForbidden, e.UNAUTHORIZED, "需要管理员权限")
            c.Abort()
            return
        }
        c.Next()
    }
}
```

在 `JWTauth()` 中新增一行：

```go
ctx.Set("role", claims.Role)
```

---

## 4. API 端点规格

### 4.1 Stats

```
GET /admin/stats

Response {
  total_users: int,
  total_orders: int,
  total_revenue_cents: int64,   // 已支付订单关联的 payment.amount_cents 之和
  total_products: int,
  pending_orders: int            // status=0 的订单数
}
```

### 4.2 Users

```
GET /admin/users?page=1&page_size=20

Response (分页) {
  list: [{
    id, username, balance, avatar,
    total_spent_cents, growth_level, role, created_at
  }],
  total, page, page_size
}
```

### 4.3 Orders

```
GET /admin/orders?page=1&page_size=20&status=0|1|2

Response (分页) {
  list: [Order],   // 不过滤 user_id，返回全站订单
  total, page, page_size
}
```

### 4.4 Coupons (CRUD)

```
GET /admin/coupons?page=1&page_size=20
Response (分页): { list: [Coupon], total, page, page_size }

POST /admin/coupons
Body: {
  type: "full_cut" | "discount",
  title: string,
  description: string,
  amount_cents: int64,
  discount_rate: int,
  min_spend_cents: int64,
  valid_from: datetime,
  valid_to: datetime,
  purchasable: bool,
  price_cents: int64,
  status: "active" | "inactive"
}
Response: Coupon

PUT /admin/coupons/:id
Body: 同 POST（部分字段）
Response: Coupon

DELETE /admin/coupons/:id
Response: { message: "ok" }
```

### 4.5 Products

```
GET /admin/products?page=1&page_size=20
Response (分页): { list: [Product], total, page, page_size }
```

全站商品列表，不过滤 user_id。

### 4.6 Risk (黑名单/灰名单)

```
GET    /admin/risk/blacklist       → { ip: string[], user: string[] }
POST   /admin/risk/blacklist       Body: { type: "ip"|"user", value: string }
DELETE /admin/risk/blacklist       Body: { type: "ip"|"user", value: string }

GET    /admin/risk/graylist        → { ip: string[], user: string[] }
POST   /admin/risk/graylist        Body: { type: "ip"|"user", value: string }
DELETE /admin/risk/graylist        Body: { type: "ip"|"user", value: string }
```

Redis key 约定（与现有中间件一致）：

| 名单 | Redis Key |
|------|-----------|
| IP 黑名单 | `risk:ip:black` (SET) |
| 用户黑名单 | `risk:user:black` (SET) |
| IP 灰名单 | `risk:ip:gray` (SET) |
| 用户灰名单 | `risk:user:gray` (SET) |

---

## 5. Repository 新增方法

### UserRepo

```go
func (r *UserRepo) ListAll(ctx context.Context, page, pageSize int) ([]model.User, int64, error)
func (r *UserRepo) CountAll(ctx context.Context) (int64, error)
```

### CouponRepo

```go
func (r *CouponRepo) ListAll(ctx context.Context, page, pageSize int) ([]model.Coupon, int64, error)
func (r *CouponRepo) Update(ctx context.Context, id uint, updates map[string]any) error
func (r *CouponRepo) Delete(ctx context.Context, id uint) error
```

### ProductRepo

```go
func (r *ProductRepo) ListAll(ctx context.Context, page, pageSize int) ([]model.Product, int64, error)
func (r *ProductRepo) CountAll(ctx context.Context) (int64, error)
```

### OrderRepo（新建或扩展 OrderService）

```go
func ListAllOrders(ctx context.Context, status *model.OrderStatus, page, pageSize int) ([]model.Order, int64, error)
func CountByStatus(ctx context.Context, status model.OrderStatus) (int64, error)
func SumRevenue(ctx context.Context) (int64, error)   // JOIN payments WHERE status='paid'
```

---

## 6. Service 层

新建 `internal/service/admin.go`：

```go
type AdminService struct {
    db          *gorm.DB
    userRepo    *repository.UserRepo
    productRepo *repository.ProductRepo
}

func (s *AdminService) Stats(ctx context.Context) (*AdminStats, error)
func (s *AdminService) ListUsers(ctx context.Context, page, pageSize int) ([]model.User, int64, error)
func (s *AdminService) ListAllOrders(ctx context.Context, status *model.OrderStatus, page, pageSize int) ([]model.Order, int64, error)
func (s *AdminService) ListAllProducts(ctx context.Context, page, pageSize int) ([]model.Product, int64, error)
```

新建 `internal/service/risk.go`：

```go
type RiskService struct {
    rdb *redis.Client
}

func (s *RiskService) ListBlacklist(ctx context.Context) (ips, users []string, err error)
func (s *RiskService) AddBlacklist(ctx context.Context, entryType, value string) error
func (s *RiskService) RemoveBlacklist(ctx context.Context, entryType, value string) error
// 灰名单同理
```

优惠券 CRUD 复用现有 `CouponService`，新增 `ListAll`、`Update`、`Delete` 方法。

---

## 7. Handler 层

新建 `internal/handler/admin.go`：

```go
type AdminHandler struct {
    adminSvc  *service.AdminService
    riskSvc   *service.RiskService
    couponSvc *service.CouponService
}

func (h *AdminHandler) Stats(c *gin.Context)
func (h *AdminHandler) ListUsers(c *gin.Context)
func (h *AdminHandler) ListOrders(c *gin.Context)
func (h *AdminHandler) ListCoupons(c *gin.Context)
func (h *AdminHandler) CreateCoupon(c *gin.Context)
func (h *AdminHandler) UpdateCoupon(c *gin.Context)
func (h *AdminHandler) DeleteCoupon(c *gin.Context)
func (h *AdminHandler) ListProducts(c *gin.Context)
func (h *AdminHandler) ListBlacklist(c *gin.Context)
func (h *AdminHandler) AddBlacklist(c *gin.Context)
func (h *AdminHandler) RemoveBlacklist(c *gin.Context)
func (h *AdminHandler) ListGraylist(c *gin.Context)
func (h *AdminHandler) AddGraylist(c *gin.Context)
func (h *AdminHandler) RemoveGraylist(c *gin.Context)
```

---

## 8. 路由注册

在 `internal/server/http.go` 中 `auth` group 之后：

```go
riskServicer := service.NewRiskService(redis.RDB)
adminServicer := service.NewAdminService(db.DB, userRepo, productRepo)
adminHandler := handler.NewAdminHandler(adminServicer, riskServicer, couponServicer)

admin := api.Group("/admin")
admin.Use(middlerware.JWTauth(), middlerware.AdminAuth())
{
    admin.GET("/stats", adminHandler.Stats)
    admin.GET("/users", adminHandler.ListUsers)
    admin.GET("/orders", adminHandler.ListOrders)
    admin.GET("/coupons", adminHandler.ListCoupons)
    admin.POST("/coupons", adminHandler.CreateCoupon)
    admin.PUT("/coupons/:id", adminHandler.UpdateCoupon)
    admin.DELETE("/coupons/:id", adminHandler.DeleteCoupon)
    admin.GET("/products", adminHandler.ListProducts)
    admin.GET("/risk/blacklist", adminHandler.ListBlacklist)
    admin.POST("/risk/blacklist", adminHandler.AddBlacklist)
    admin.DELETE("/risk/blacklist", adminHandler.RemoveBlacklist)
    admin.GET("/risk/graylist", adminHandler.ListGraylist)
    admin.POST("/risk/graylist", adminHandler.AddGraylist)
    admin.DELETE("/risk/graylist", adminHandler.RemoveGraylist)
}
```
