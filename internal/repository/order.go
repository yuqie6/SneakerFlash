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

func (r *OrderRepo) Create(ctx context.Context, order *model.Order) error {
	return r.db.WithContext(ctx).Create(order).Error
}

func (r *OrderRepo) GetByID(ctx context.Context, id uint) (*model.Order, error) {
	var order model.Order
	if err := r.db.WithContext(ctx).Where("id = ?", id).First(&order).Error; err != nil {
		return nil, err
	}
	return &order, nil
}

// GetByOrderNum 根据订单号查询。
func (r *OrderRepo) GetByOrderNum(ctx context.Context, orderNum string) (*model.Order, error) {
	var order model.Order
	if err := r.db.WithContext(ctx).Where("order_num = ?", orderNum).First(&order).Error; err != nil {
		return nil, err
	}
	return &order, nil
}

// GetByIDForUpdate 查询订单并加行级锁，适用于支付/优惠券等并发修改场景。
func (r *OrderRepo) GetByIDForUpdate(ctx context.Context, id uint) (*model.Order, error) {
	var order model.Order
	if err := r.db.WithContext(ctx).Clauses(clause.Locking{Strength: "UPDATE"}).Where("id = ?", id).First(&order).Error; err != nil {
		return nil, err
	}
	return &order, nil
}

// GetByUserAndProduct 查询用户对同一商品的订单，用于幂等创建。
func (r *OrderRepo) GetByUserAndProduct(ctx context.Context, userID, productID uint) (*model.Order, error) {
	var order model.Order
	if err := r.db.WithContext(ctx).Where("user_id = ? AND product_id = ?", userID, productID).First(&order).Error; err != nil {
		return nil, err
	}
	return &order, nil
}

// ListByUserID 获取用户订单列表，可按状态过滤，按创建时间倒序。
func (r *OrderRepo) ListByUserID(ctx context.Context, uid uint, status *model.OrderStatus, page, pagesize int) ([]model.Order, int64, error) {
	var orders []model.Order
	var total int64

	query := r.db.WithContext(ctx).Model(&model.Order{}).Where("user_id = ?", uid)
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
func (r *OrderRepo) UpdateStatusIfMatch(ctx context.Context, orderID uint, fromStatus, toStatus model.OrderStatus) (int64, error) {
	tx := r.db.WithContext(ctx).Model(&model.Order{}).Where("id = ? AND status = ?", orderID, fromStatus).Update("status", toStatus)
	return tx.RowsAffected, tx.Error
}
