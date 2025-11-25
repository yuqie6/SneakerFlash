package service

import (
	"SneakerFlash/internal/model"
	"SneakerFlash/internal/pkg/utils"
	"SneakerFlash/internal/pkg/vip"
	"SneakerFlash/internal/repository"
	"context"
	"errors"
	"fmt"

	"gorm.io/gorm"
)

var (
	ErrOrderNotFound        = errors.New("订单不存在")
	ErrPaymentNotFound      = errors.New("支付单不存在")
	ErrUnsupportedPayStatus = errors.New("不支持的支付状态")
)

type OrderService struct {
	db          *gorm.DB
	orderRepo   *repository.OrderRepo
	paymentRepo *repository.PaymentRepo
	productRepo *repository.ProductRepo
	userRepo    *repository.UserRepo
	couponSvc   *CouponService
}

type OrderWithPayment struct {
	Order   *model.Order      `json:"order"`
	Payment *model.Payment    `json:"payment,omitempty"`
	Coupon  *model.UserCoupon `json:"coupon,omitempty"`
}

// NewOrderService 构建订单服务，聚合订单/支付/商品仓储用于事务处理。
func NewOrderService(db *gorm.DB, productRepo *repository.ProductRepo, userRepo *repository.UserRepo) *OrderService {
	return &OrderService{
		db:          db,
		orderRepo:   repository.NewOrderRepo(db),
		paymentRepo: repository.NewPaymentRepo(db),
		productRepo: productRepo,
		userRepo:    userRepo,
		couponSvc:   NewCouponService(db),
	}
}

// WithContext 绑定请求上下文，使事务与仓储日志携带 request_id。
func (s *OrderService) WithContext(ctx context.Context) *OrderService {
	if ctx == nil {
		return s
	}
	ctxDB := s.db.WithContext(ctx)
	return &OrderService{
		db:          ctxDB,
		orderRepo:   s.orderRepo.WithContext(ctx),
		paymentRepo: s.paymentRepo.WithContext(ctx),
		productRepo: s.productRepo.WithContext(ctx),
		userRepo:    s.userRepo.WithContext(ctx),
		couponSvc:   s.couponSvc.WithContext(ctx),
	}
}

// CreateOrderAndInitPayment 创建订单并确保支付单存在；按 user_id+product_id 幂等，事务内处理并发重试。
func (s *OrderService) CreateOrderAndInitPayment(userID, productID uint, amountCents int64, couponID *uint) (*OrderWithPayment, error) {
	var result OrderWithPayment

	err := s.db.Transaction(func(tx *gorm.DB) error {
		txOrderRepo := repository.NewOrderRepo(tx)
		txPaymentRepo := repository.NewPaymentRepo(tx)
		txCouponSvc := NewCouponService(tx)

		// 先尝试查已有订单，满足幂等
		order, err := txOrderRepo.GetByUserAndProduct(userID, productID)
		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}

		// 不存在则创建新订单
		if errors.Is(err, gorm.ErrRecordNotFound) {
			orderNum, genErr := utils.GenSnowflakeID()
			if genErr != nil {
				return genErr
			}
			order = &model.Order{
				UserID:    userID,
				ProductID: productID,
				OrderNum:  orderNum,
				Status:    model.OrderStatusUnpaid,
			}
			if createErr := txOrderRepo.Create(order); createErr != nil {
				if errors.Is(createErr, gorm.ErrDuplicatedKey) {
					// 并发创建幂等，回查已有订单
					order, err = txOrderRepo.GetByUserAndProduct(userID, productID)
					if err != nil {
						return err
					}
				} else {
					return createErr
				}
			}
		}

		// 创建或复用支付单（单订单唯一支付单）
		finalAmount := amountCents
		var appliedCoupon *model.UserCoupon
		if couponID != nil {
			uc, _, discounted, cErr := txCouponSvc.ApplyCoupon(userID, *couponID, amountCents)
			if cErr != nil {
				return cErr
			}
			finalAmount = discounted
			appliedCoupon = uc
		}

		paymentID, genErr := utils.GenSnowflakeID()
		if genErr != nil {
			return genErr
		}
		payment := &model.Payment{
			OrderID:     order.ID,
			PaymentID:   paymentID,
			AmountCents: finalAmount,
			Status:      model.PaymentStatusPending,
		}
		payment, err = txPaymentRepo.CreateIfAbsent(payment)
		if err != nil {
			return err
		}
		if appliedCoupon != nil {
			if err := repository.NewUserCouponRepo(tx).MarkUsed(appliedCoupon.ID, order.ID); err != nil {
				return err
			}
		}

		result = OrderWithPayment{
			Order:   order,
			Payment: payment,
			Coupon:  appliedCoupon,
		}
		return nil
	})

	if err != nil {
		return nil, err
	}
	return &result, nil
}

// ListOrders 查询订单列表，可选状态过滤。
func (s *OrderService) ListOrders(userID uint, status *model.OrderStatus, page, pageSize int) ([]model.Order, int64, error) {
	return s.orderRepo.ListByUserID(userID, status, page, pageSize)
}

// GetOrderWithPayment 获取订单详情（含支付单），同时校验用户归属并补偿缺失的支付单。
func (s *OrderService) GetOrderWithPayment(userID, orderID uint) (*OrderWithPayment, error) {
	order, err := s.orderRepo.GetByID(orderID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrOrderNotFound
		}
		return nil, err
	}
	if order.UserID != userID {
		return nil, ErrOrderNotFound
	}

	payment, err := s.paymentRepo.GetByOrderID(orderID)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	if payment == nil && errors.Is(err, gorm.ErrRecordNotFound) {
		// 补偿创建支付单，避免页面缺少 payment_id
		product, pErr := s.productRepo.GetByID(order.ProductID)
		if pErr != nil {
			return nil, pErr
		}
		amountCents := int64(product.Price * 100)
		paymentID, genErr := utils.GenSnowflakeID()
		if genErr != nil {
			return nil, genErr
		}
		newPayment := &model.Payment{
			OrderID:     order.ID,
			PaymentID:   paymentID,
			AmountCents: amountCents,
			Status:      model.PaymentStatusPending,
		}
		if _, cErr := s.paymentRepo.CreateIfAbsent(newPayment); cErr != nil {
			return nil, cErr
		}
		payment = newPayment
	}

	return &OrderWithPayment{
		Order:   order,
		Payment: payment,
	}, nil
}

// HandlePaymentResult 幂等处理支付回调：乐观锁更新支付单，条件更新订单状态，并在支付成功时刷新缓存库存。
func (s *OrderService) HandlePaymentResult(paymentID string, targetStatus model.PaymentStatus, notifyData string) (*OrderWithPayment, error) {
	if targetStatus != model.PaymentStatusPaid && targetStatus != model.PaymentStatusFailed && targetStatus != model.PaymentStatusRefunded {
		return nil, fmt.Errorf("%w: %s", ErrUnsupportedPayStatus, targetStatus)
	}

	var result OrderWithPayment
	if paymentID == "" {
		return nil, ErrPaymentNotFound
	}

	// 事务防脏写
	// 事务的四大特性
	// 原子性、一致性、隔离性、持久性
	err := s.db.Transaction(func(tx *gorm.DB) error {
		txPaymentRepo := repository.NewPaymentRepo(tx)
		txOrderRepo := repository.NewOrderRepo(tx)
		txProductRepo := repository.NewProductRepo(tx)
		txUserCouponRepo := repository.NewUserCouponRepo(tx)

		payment, err := txPaymentRepo.GetByPaymentID(paymentID)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return ErrPaymentNotFound
			}
			return err
		}

		// 乐观锁更新支付状态, 防守护并发竟态与重复回调
		rows, err := txPaymentRepo.UpdateStatusByPaymentIDIfMatch(paymentID, model.PaymentStatusPending, targetStatus, notifyData)
		if err != nil {
			return err
		}
		if rows == 0 {
			// 幂等命中或已处理，返回当前状态
			updated, getErr := txPaymentRepo.GetByPaymentID(paymentID)
			if getErr != nil {
				return getErr
			}
			order, getErr := txOrderRepo.GetByID(payment.OrderID)
			if getErr != nil {
				return getErr
			}
			result = OrderWithPayment{Order: order, Payment: updated}
			return nil
		}

		// 支付状态变更成功后，尝试更新订单状态
		orderStatus := model.OrderStatusFailed
		if targetStatus == model.PaymentStatusPaid {
			orderStatus = model.OrderStatusPaid
		}
		if _, err := txOrderRepo.UpdateStatusIfMatch(payment.OrderID, model.OrderStatusUnpaid, orderStatus); err != nil {
			return err
		}
		order, err := txOrderRepo.GetByID(payment.OrderID)
		if err != nil {
			return err
		}
		updatedPayment, err := txPaymentRepo.GetByPaymentID(paymentID)
		if err != nil {
			return err
		}
		if targetStatus == model.PaymentStatusPaid {
			if product, pErr := txProductRepo.GetByID(order.ProductID); pErr == nil {
				// 异步刷新缓存库存
				refreshStockCacheAsync(product.ID, product.Stock)
				go invalidateProductInfoCache(product.ID)
			}
			// 成长值累积：按支付金额计算成长等级
			txUserRepo := repository.NewUserRepo(tx)
			user, uErr := txUserRepo.GetByIDForUpdate(order.UserID)
			if uErr != nil {
				return uErr
			}
			newTotal := user.TotalSpentCents + payment.AmountCents
			newLevel := vip.CalcGrowthLevel(newTotal)
			if uErr := txUserRepo.UpdateGrowth(order.UserID, newTotal, newLevel); uErr != nil {
				return uErr
			}
		} else {
			// 支付失败/退款则释放已占用的优惠券，避免用户券被锁死
			if releaseErr := txUserCouponRepo.ReleaseByOrder(order.ID); releaseErr != nil {
				return releaseErr
			}
		}
		result = OrderWithPayment{
			Order:   order,
			Payment: updatedPayment,
		}
		return nil
	})

	if err != nil {
		return nil, err
	}
	return &result, nil
}
