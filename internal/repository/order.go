package repository

import (
	"SneakerFlash/internal/model"

	"gorm.io/gorm"
)

type OrderRepo struct {
	db *gorm.DB
}

func NewOrderRepo(db *gorm.DB) *OrderRepo {
	return &OrderRepo{
		db: db,
	}
}

func (r *OrderRepo) Create(order *model.Order) error {
	return r.db.Create(order).Error
}

func (r *OrderRepo) GetByOrderNum(orderNum string) (*model.Order, error) {
	var order model.Order
	if err := r.db.Where("order_num = ?", orderNum).First(&order).Error; err != nil {
		return nil, err
	}

	return &order, nil
}

// 获取用户的订单列表
func (r *OrderRepo) ListByUserID(uid uint, page, pagesize int) ([]model.Order, int64, error) {
	var orders []model.Order
	var total int64

	query := r.db.Model(&model.Order{}).Where("user_id = ?", uid)

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pagesize

	err := query.Offset(offset).Limit(pagesize).Order("created_at desc").Find(&orders).Error

	return orders, total, err
}

// 更新订单的状态(待支付 -> 已支付)
func (r *OrderRepo) UpdateStatus(orderID uint, status model.OrderStatus) error {
	return r.db.Model(&model.Order{}).Where("id = ?", orderID).Update("status", status).Error
}
