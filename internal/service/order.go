package service

import (
	"SneakerFlash/internal/model"
	"SneakerFlash/internal/pkg/utils"
	"SneakerFlash/internal/pkg/vip"
	"SneakerFlash/internal/repository"
	"context"
	"errors"
	"fmt"
	"math"

	"gorm.io/gorm"
)

var (
	ErrOrderNotFound        = errors.New("订单不存在")
	ErrPaymentNotFound      = errors.New("支付单不存在")
	ErrUnsupportedPayStatus = errors.New("不支持的支付状态")
	ErrOrderNotPayable      = errors.New("订单状态不可支付")
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
	Order   *model.Order   `json:"order"`
	Payment *model.Payment `json:"payment,omitempty"`
	Coupon  *MyCoupon      `json:"coupon,omitempty"`
}

// OrderPollResult 描述订单轮询结果，兼容异步创建场景。
type OrderPollResult struct {
	Status    PendingOrderStatus `json:"status"`
	OrderNum  string             `json:"order_num"`
	PaymentID string             `json:"payment_id,omitempty"`
	Order     *OrderWithPayment  `json:"order,omitempty"`
	Message   string             `json:"message,omitempty"`
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

func toMyCoupon(uc *model.UserCoupon, c *model.Coupon) *MyCoupon {
	if uc == nil || c == nil {
		return nil
	}
	return &MyCoupon{
		ID:            uc.ID,
		CouponID:      uc.CouponID,
		Type:          c.Type,
		Title:         c.Title,
		Description:   c.Description,
		AmountCents:   c.AmountCents,
		DiscountRate:  c.DiscountRate,
		MinSpendCents: c.MinSpendCents,
		Status:        uc.Status,
		ValidFrom:     uc.ValidFrom,
		ValidTo:       uc.ValidTo,
		ObtainedFrom:  uc.ObtainedFrom,
	}
}

// ApplyCoupon 在待支付订单上应用/更换优惠券；若 couponID 为空则移除已用优惠券。
func (s *OrderService) ApplyCoupon(userID, orderID uint, couponID *uint) (*OrderWithPayment, error) {
	var result OrderWithPayment

	err := s.db.Transaction(func(tx *gorm.DB) error {
		txOrderRepo := repository.NewOrderRepo(tx)
		txPaymentRepo := repository.NewPaymentRepo(tx)
		txProductRepo := repository.NewProductRepo(tx)
		txCouponSvc := NewCouponService(tx)
		txUserCouponRepo := repository.NewUserCouponRepo(tx)

		order, err := txOrderRepo.GetByIDForUpdate(orderID)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return ErrOrderNotFound
			}
			return err
		}
		if order.UserID != userID {
			return ErrOrderNotFound
		}
		if order.Status != model.OrderStatusUnpaid {
			return ErrOrderNotPayable
		}

		product, err := txProductRepo.GetByID(order.ProductID)
		if err != nil {
			return err
		}
		baseAmount := int64(math.Round(product.Price * 100))
		if baseAmount <= 0 {
			return fmt.Errorf("invalid product price: %v", product.Price)
		}

		if _, existingErr := txUserCouponRepo.GetByOrderID(order.ID); existingErr == nil {
			_ = txUserCouponRepo.ReleaseByOrder(order.ID)
		}

		finalAmount := baseAmount
		var appliedUC *model.UserCoupon
		var appliedTpl *model.Coupon
		if couponID != nil {
			uc, tpl, discounted, cErr := txCouponSvc.ApplyCoupon(userID, *couponID, baseAmount)
			if cErr != nil {
				return cErr
			}
			finalAmount = discounted
			appliedUC = uc
			appliedTpl = tpl
		}

		payment, err := txPaymentRepo.GetByOrderID(order.ID)
		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}
		if payment == nil {
			paymentID, genErr := utils.GenSnowflakeID()
			if genErr != nil {
				return genErr
			}
			payment = &model.Payment{
				OrderID:     order.ID,
				PaymentID:   paymentID,
				AmountCents: finalAmount,
				Status:      model.PaymentStatusPending,
			}
			if _, err := txPaymentRepo.CreateIfAbsent(payment); err != nil {
				return err
			}
		} else {
			rows, err := txPaymentRepo.UpdateAmountIfPending(order.ID, finalAmount)
			if err != nil {
				return err
			}
			if rows == 0 && payment.Status != model.PaymentStatusPending {
				return ErrOrderNotPayable
			}
			payment, err = txPaymentRepo.GetByOrderID(order.ID)
			if err != nil {
				return err
			}
		}

		if appliedUC != nil {
			if err := txUserCouponRepo.MarkUsed(appliedUC.ID, order.ID); err != nil {
				return err
			}
		}

		result = OrderWithPayment{
			Order:   order,
			Payment: payment,
			Coupon:  toMyCoupon(appliedUC, appliedTpl),
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
		amountCents := int64(math.Round(product.Price * 100))
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

	var myCoupon *MyCoupon
	if uc, ucErr := s.couponSvc.userCouponRepo.GetByOrderID(order.ID); ucErr == nil && uc != nil {
		if tpl, tplErr := s.couponSvc.couponRepo.GetByID(uc.CouponID); tplErr == nil {
			myCoupon = toMyCoupon(uc, tpl)
		}
	}

	return &OrderWithPayment{
		Order:   order,
		Payment: payment,
		Coupon:  myCoupon,
	}, nil
}

// GetOrderWithPaymentByNum 根据订单号查询订单与支付单，适用于轮询接口。
func (s *OrderService) GetOrderWithPaymentByNum(userID uint, orderNum string) (*OrderWithPayment, error) {
	order, err := s.orderRepo.GetByOrderNum(orderNum)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrOrderNotFound
		}
		return nil, err
	}
	if order.UserID != userID {
		return nil, ErrOrderNotFound
	}
	return s.GetOrderWithPayment(userID, order.ID)
}

// PollOrder 查询订单创建状态，优先读取缓存的 pending/ready 状态，再回源数据库。
func (s *OrderService) PollOrder(userID uint, orderNum string) (*OrderPollResult, error) {
	ctx := context.Background()
	cache, err := getPendingOrder(ctx, orderNum)
	if err == nil && cache != nil {
		switch cache.Status {
		case PendingStatusPending:
			return &OrderPollResult{Status: PendingStatusPending, OrderNum: orderNum, PaymentID: cache.PaymentID}, nil
		case PendingStatusFailed:
			return &OrderPollResult{Status: PendingStatusFailed, OrderNum: orderNum, Message: cache.Message}, nil
		case PendingStatusReady:
			if cache.OrderID > 0 {
				order, getErr := s.GetOrderWithPayment(userID, cache.OrderID)
				if getErr == nil {
					pid := cache.PaymentID
					if pid == "" && order.Payment != nil {
						pid = order.Payment.PaymentID
					}
					return &OrderPollResult{Status: PendingStatusReady, OrderNum: orderNum, PaymentID: pid, Order: order}, nil
				}
			}
		}
	}

	order, getErr := s.GetOrderWithPaymentByNum(userID, orderNum)
	if getErr != nil {
		if errors.Is(getErr, ErrOrderNotFound) {
			return &OrderPollResult{Status: PendingStatusPending, OrderNum: orderNum}, nil
		}
		return nil, getErr
	}

	pid := ""
	if order.Payment != nil {
		pid = order.Payment.PaymentID
	}

	return &OrderPollResult{Status: PendingStatusReady, OrderNum: orderNum, PaymentID: pid, Order: order}, nil
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
