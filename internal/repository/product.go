package repository

import (
	"SneakerFlash/internal/model"

	"gorm.io/gorm"
)

// 封装商品repo层方法
type ProductRepo struct {
	db *gorm.DB
}

// 构建 productrepo
func NewProductRepo(db *gorm.DB) *ProductRepo {
	return &ProductRepo{
		db: db,
	}
}

// 根据 ID 获取商品的详情
func (r *ProductRepo) GetByID(id uint) (*model.Product, error) {
	var p model.Product

	if err := r.db.First(&p, id).Error; err != nil {
		return nil, err
	}
	return &p, nil
}

func (r *ProductRepo) GetByIDAndUser(id, userID uint) (*model.Product, error) {
	var p model.Product
	if err := r.db.Where("id = ? AND user_id = ?", id, userID).First(&p).Error; err != nil {
		return nil, err
	}
	return &p, nil
}

// 商品库存减少函数
func (r *ProductRepo) ReduceStockDB(id uint) (int64, error) {
	result := r.db.Model(&model.Product{}).
		Where("id = ? AND stock > 0", id).
		Update("stock", gorm.Expr("stock - 1"))
	return result.RowsAffected, result.Error
}

func (r *ProductRepo) Create(product *model.Product) error {
	return r.db.Create(product).Error
}

func (r *ProductRepo) Update(id uint, data map[string]any) error {
	return r.db.Model(&model.Product{}).Where("id = ?", id).Updates(data).Error
}

func (r *ProductRepo) UpdateByUser(id, userID uint, data map[string]any) (int64, error) {
	tx := r.db.Model(&model.Product{}).Where("id = ? AND user_id = ?", id, userID).Updates(data)
	return tx.RowsAffected, tx.Error
}

func (r *ProductRepo) List(page, pageSize int) ([]model.Product, int64, error) {
	var products []model.Product
	var total int64

	if err := r.db.Model(&model.Product{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	err := r.db.Offset(offset).Limit(pageSize).Order("id desc").Find(&products).Error

	return products, total, err
}

func (r *ProductRepo) ListByUserID(userID uint, page, pageSize int) ([]model.Product, int64, error) {
	var products []model.Product
	var total int64
	query := r.db.Model(&model.Product{}).Where("user_id = ?", userID)
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	offset := (page - 1) * pageSize
	err := query.Offset(offset).Limit(pageSize).Order("id desc").Find(&products).Error
	return products, total, err
}

func (r *ProductRepo) Delete(id uint) error {
	return r.db.Delete(&model.Product{}, id).Error
}
