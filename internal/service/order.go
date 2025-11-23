package service

import (
	"SneakerFlash/internal/model"
	"SneakerFlash/internal/pkg/utils"
	"SneakerFlash/internal/repository"
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
}

type OrderWithPayment struct {
	Order   *model.Order
	Payment *model.Payment
}

func NewOrderService(db *gorm.DB) *OrderService {
	return &OrderService{
		db:          db,
		orderRepo:   repository.NewOrderRepo(db),
		paymentRepo: repository.NewPaymentRepo(db),
	}
}

// 创建订单并确保存在对应支付单；按 user_id+product_id 幂等，返回订单与支付单
func (s *OrderService) CreateOrderAndInitPayment(userID, productID uint, amountCents int64) (*OrderWithPayment, error) {
	var result OrderWithPayment

	err := s.db.Transaction(func(tx *gorm.DB) error {
		txOrderRepo := repository.NewOrderRepo(tx)
		txPaymentRepo := repository.NewPaymentRepo(tx)

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
		paymentID, genErr := utils.GenSnowflakeID()
		if genErr != nil {
			return genErr
		}
		payment := &model.Payment{
			OrderID:     order.ID,
			PaymentID:   paymentID,
			AmountCents: amountCents,
			Status:      model.PaymentStatusPending,
		}
		payment, err = txPaymentRepo.CreateIfAbsent(payment)
		if err != nil {
			return err
		}

		result = OrderWithPayment{
			Order:   order,
			Payment: payment,
		}
		return nil
	})

	if err != nil {
		return nil, err
	}
	return &result, nil
}

// 查询订单列表（可选状态过滤）
func (s *OrderService) ListOrders(userID uint, status *model.OrderStatus, page, pageSize int) ([]model.Order, int64, error) {
	return s.orderRepo.ListByUserID(userID, status, page, pageSize)
}

// 获取订单详情，包含支付单；校验 user 归属
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

	return &OrderWithPayment{
		Order:   order,
		Payment: payment,
	}, nil
}

// 处理支付结果回调，幂等更新支付单与订单状态
// 防 1: 脏写 2: 并发竟态 3: 重复回调
func (s *OrderService) HandlePaymentResult(paymentID string, targetStatus model.PaymentStatus, notifyData string) (*OrderWithPayment, error) {
	if targetStatus != model.PaymentStatusPaid && targetStatus != model.PaymentStatusFailed && targetStatus != model.PaymentStatusRefunded {
		return nil, fmt.Errorf("%w: %s", ErrUnsupportedPayStatus, targetStatus)
	}

	var result OrderWithPayment

	// 事务防脏写
	// 事务的四大特性
	// 原子性、一致性、隔离性、持久性
	err := s.db.Transaction(func(tx *gorm.DB) error {
		txPaymentRepo := repository.NewPaymentRepo(tx)
		txOrderRepo := repository.NewOrderRepo(tx)

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
