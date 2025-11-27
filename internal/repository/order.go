package repository

import (
	"SneakerFlash/internal/model"
	"context"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type OrderRepo struct {
	db *gorm.DB
}

// NewOrderRepo 构建订单仓储。
func NewOrderRepo(db *gorm.DB) *OrderRepo {
	return &OrderRepo{
		db: db,
	}
}

// WithContext 绑定请求上下文，保持仓储日志与上游 request_id 对齐。
func (r *OrderRepo) WithContext(ctx context.Context) *OrderRepo {
	if ctx == nil {
		return r
	}
	return &OrderRepo{db: r.db.WithContext(ctx)}
}

func (r *OrderRepo) Create(order *model.Order) error {
	return r.db.Create(order).Error
}

func (r *OrderRepo) GetByID(id uint) (*model.Order, error) {
	var order model.Order
	if err := r.db.Where("id = ?", id).First(&order).Error; err != nil {
		return nil, err
	}
	return &order, nil
}

// GetByOrderNum 根据订单号查询。
func (r *OrderRepo) GetByOrderNum(orderNum string) (*model.Order, error) {
	var order model.Order
	if err := r.db.Where("order_num = ?", orderNum).First(&order).Error; err != nil {
		return nil, err
	}
	return &order, nil
}

// GetByIDForUpdate 查询订单并加行级锁，适用于支付/优惠券等并发修改场景。
func (r *OrderRepo) GetByIDForUpdate(id uint) (*model.Order, error) {
	var order model.Order
	if err := r.db.Clauses(clause.Locking{Strength: "UPDATE"}).Where("id = ?", id).First(&order).Error; err != nil {
		return nil, err
	}
	return &order, nil
}

// GetByUserAndProduct 查询用户对同一商品的订单，用于幂等创建。
func (r *OrderRepo) GetByUserAndProduct(userID, productID uint) (*model.Order, error) {
	var order model.Order
	if err := r.db.Where("user_id = ? AND product_id = ?", userID, productID).First(&order).Error; err != nil {
		return nil, err
	}
	return &order, nil
}

// ListByUserID 获取用户订单列表，可按状态过滤，按创建时间倒序。
func (r *OrderRepo) ListByUserID(uid uint, status *model.OrderStatus, page, pagesize int) ([]model.Order, int64, error) {
	var orders []model.Order
	var total int64

	query := r.db.Model(&model.Order{}).Where("user_id = ?", uid)
	if status != nil {
		query = query.Where("status = ?", *status)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pagesize

	err := query.Offset(offset).Limit(pagesize).Order("created_at desc").Find(&orders).Error

	return orders, total, err
}

// UpdateStatusIfMatch 仅在当前状态匹配时更新，用于避免重复回调覆盖。
func (r *OrderRepo) UpdateStatusIfMatch(orderID uint, fromStatus, toStatus model.OrderStatus) (int64, error) {
	tx := r.db.Model(&model.Order{}).Where("id = ? AND status = ?", orderID, fromStatus).Update("status", toStatus)
	return tx.RowsAffected, tx.Error
}
