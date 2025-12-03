package repository

import (
	"SneakerFlash/internal/model"
	"context"

	"gorm.io/gorm"
)

type ProductRepo struct {
	db *gorm.DB
}

// NewProductRepo 构建商品仓储。
func NewProductRepo(db *gorm.DB) *ProductRepo {
	return &ProductRepo{
		db: db,
	}
}

// GetByID 根据 ID 获取商品详情。
func (r *ProductRepo) GetByID(ctx context.Context, id uint) (*model.Product, error) {
	var p model.Product

	if err := r.db.WithContext(ctx).First(&p, id).Error; err != nil {
		return nil, err
	}
	return &p, nil
}

func (r *ProductRepo) GetByIDAndUser(ctx context.Context, id, userID uint) (*model.Product, error) {
	var p model.Product
	if err := r.db.WithContext(ctx).Where("id = ? AND user_id = ?", id, userID).First(&p).Error; err != nil {
		return nil, err
	}
	return &p, nil
}

// ReduceStockDB 原子扣减库存，确保剩余库存大于 0 时才减。
func (r *ProductRepo) ReduceStockDB(ctx context.Context, id uint) (int64, error) {
	result := r.db.WithContext(ctx).Model(&model.Product{}).
		Where("id = ? AND stock > 0", id).
		Update("stock", gorm.Expr("stock - 1"))
	return result.RowsAffected, result.Error
}

// ReduceStockDBBatch 批量扣减库存，确保剩余库存 >= count 时才扣减。
func (r *ProductRepo) ReduceStockDBBatch(ctx context.Context, id uint, count int) (int64, error) {
	result := r.db.WithContext(ctx).Model(&model.Product{}).
		Where("id = ? AND stock >= ?", id, count).
		Update("stock", gorm.Expr("stock - ?", count))
	return result.RowsAffected, result.Error
}

// Create 插入新商品。
func (r *ProductRepo) Create(ctx context.Context, product *model.Product) error {
	return r.db.WithContext(ctx).Create(product).Error
}

// Update 按 ID 更新商品，调用方需确保归属。
func (r *ProductRepo) Update(ctx context.Context, id uint, data map[string]any) error {
	return r.db.WithContext(ctx).Model(&model.Product{}).Where("id = ?", id).Updates(data).Error
}

// UpdateByUser 限定创建者更新，返回受影响行数用于判断是否存在/越权。
func (r *ProductRepo) UpdateByUser(ctx context.Context, id, userID uint, data map[string]any) (int64, error) {
	tx := r.db.WithContext(ctx).Model(&model.Product{}).Where("id = ? AND user_id = ?", id, userID).Updates(data)
	return tx.RowsAffected, tx.Error
}

// List 分页查询商品列表，按 id 倒序。
func (r *ProductRepo) List(ctx context.Context, page, pageSize int) ([]model.Product, int64, error) {
	var products []model.Product
	var total int64

	db := r.db.WithContext(ctx)
	if err := db.Model(&model.Product{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	err := db.Offset(offset).Limit(pageSize).Order("id desc").Find(&products).Error

	return products, total, err
}

// ListByUserID 查询指定用户的商品列表。
func (r *ProductRepo) ListByUserID(ctx context.Context, userID uint, page, pageSize int) ([]model.Product, int64, error) {
	var products []model.Product
	var total int64
	query := r.db.WithContext(ctx).Model(&model.Product{}).Where("user_id = ?", userID)
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	offset := (page - 1) * pageSize
	err := query.Offset(offset).Limit(pageSize).Order("id desc").Find(&products).Error
	return products, total, err
}

// Delete 软删除商品。
func (r *ProductRepo) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&model.Product{}, id).Error
}
