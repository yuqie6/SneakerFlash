package model

import "gorm.io/gorm"

const (
	UserRoleUser        = "user"
	UserRoleAdmin       = "admin"
	UserRoleSuperAdmin  = "super_admin"
	UserRoleOpsAdmin    = "ops_admin"
	UserRoleRiskAdmin   = "risk_admin"
	UserRoleCouponAdmin = "coupon_admin"
	UserRoleAuditAdmin  = "audit_admin"
)

const (
	AdminResourceStats    = "stats"
	AdminResourceUsers    = "users"
	AdminResourceOrders   = "orders"
	AdminResourceProducts = "products"
	AdminResourceCoupons  = "coupons"
	AdminResourceRisk     = "risk"
	AdminResourceAudit    = "audit"
)

var adminRolePermissions = map[string][]string{
	UserRoleAdmin: {
		AdminResourceStats,
		AdminResourceUsers,
		AdminResourceOrders,
		AdminResourceProducts,
		AdminResourceCoupons,
		AdminResourceRisk,
		AdminResourceAudit,
	},
	UserRoleSuperAdmin: {
		AdminResourceStats,
		AdminResourceUsers,
		AdminResourceOrders,
		AdminResourceProducts,
		AdminResourceCoupons,
		AdminResourceRisk,
		AdminResourceAudit,
	},
	UserRoleOpsAdmin: {
		AdminResourceStats,
		AdminResourceUsers,
		AdminResourceOrders,
		AdminResourceProducts,
	},
	UserRoleRiskAdmin: {
		AdminResourceStats,
		AdminResourceRisk,
		AdminResourceAudit,
	},
	UserRoleCouponAdmin: {
		AdminResourceStats,
		AdminResourceCoupons,
		AdminResourceAudit,
	},
	UserRoleAuditAdmin: {
		AdminResourceStats,
		AdminResourceAudit,
	},
}

type User struct {
	gorm.Model
	Username        string  `gorm:"type:varchar(50);unique;not null" json:"username"`
	Password        string  `gorm:"type:varchar(100);not null" json:"-"`
	Balance         float64 `gorm:"type:decimal(10,2);default:0;not null" json:"balance"`
	Avatar          string  `gorm:"type:varchar(255);default:''" json:"avatar"`
	TotalSpentCents int64   `gorm:"type:bigint;default:0;not null" json:"total_spent_cents"`
	GrowthLevel     int     `gorm:"type:int;default:1;not null" json:"growth_level"`
	Role            string  `gorm:"type:varchar(20);default:'user';not null" json:"role"`
}

func (User) TableName() string {
	return "users"
}

func IsAdminRole(role string) bool {
	_, ok := adminRolePermissions[NormalizeUserRole(role)]
	return ok
}

func NormalizeUserRole(role string) string {
	if role == "" {
		return UserRoleUser
	}
	return role
}

func PermissionsForRole(role string) []string {
	normalized := NormalizeUserRole(role)
	permissions, ok := adminRolePermissions[normalized]
	if !ok {
		return nil
	}
	cp := make([]string, len(permissions))
	copy(cp, permissions)
	return cp
}

func HasAdminResource(role, resource string) bool {
	for _, permission := range PermissionsForRole(role) {
		if permission == resource {
			return true
		}
	}
	return false
}
