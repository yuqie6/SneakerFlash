package cron

import (
	"SneakerFlash/internal/repository"
	"SneakerFlash/internal/service"
	"context"
	"log/slog"
	"time"

	"github.com/robfig/cron/v3"
	"gorm.io/gorm"
)

// VIPCouponCron VIP 月度发券定时任务管理
type VIPCouponCron struct {
	db          *gorm.DB
	cron        *cron.Cron
	userRepo    *repository.UserRepo
	paidVIPRepo *repository.PaidVIPRepo
	couponSvc   *service.CouponService
}

// NewVIPCouponCron 创建 VIP 发券定时任务管理器
func NewVIPCouponCron(db *gorm.DB) *VIPCouponCron {
	return &VIPCouponCron{
		db:          db,
		cron:        cron.New(),
		userRepo:    repository.NewUserRepo(db),
		paidVIPRepo: repository.NewPaidVIPRepo(db),
		couponSvc:   service.NewCouponService(db),
	}
}

// Start 启动定时任务：每月 1 号 00:01 为所有 VIP 用户发放月度优惠券
func (c *VIPCouponCron) Start() {
	// 每月 1 号 0:01 执行
	_, err := c.cron.AddFunc("1 0 1 * *", c.issueMonthlyVIPCoupons)
	if err != nil {
		slog.Error("注册 VIP 月度发券任务失败", slog.Any("err", err))
		return
	}
	c.cron.Start()
	slog.Info("VIP 月度发券定时任务已启动", slog.String("schedule", "每月1号 00:01"))
}

// Stop 停止定时任务
func (c *VIPCouponCron) Stop() {
	c.cron.Stop()
}

// issueMonthlyVIPCoupons 为所有 VIP 用户发放月度优惠券
func (c *VIPCouponCron) issueMonthlyVIPCoupons() {
	ctx := context.Background()
	now := time.Now()
	slog.Info("开始执行 VIP 月度发券任务", slog.Time("time", now))

	var successCount, failCount int

	// 1. 处理成长等级用户（GrowthLevel >= 1）
	users, err := c.userRepo.ListAllWithGrowthLevel(ctx, 1)
	if err != nil {
		slog.Error("查询成长等级用户失败", slog.Any("err", err))
	} else {
		for _, user := range users {
			if err := c.couponSvc.IssueVIPMonthly(ctx, user.ID, user.GrowthLevel); err != nil {
				slog.Warn("成长等级用户发券失败", slog.Uint64("user_id", uint64(user.ID)), slog.Any("err", err))
				failCount++
			} else {
				successCount++
			}
		}
	}

	// 2. 处理付费 VIP 用户（可能等级更高，覆盖发放）
	paidVIPs, err := c.paidVIPRepo.ListActivePaidVIPs(ctx, now)
	if err != nil {
		slog.Error("查询付费 VIP 用户失败", slog.Any("err", err))
	} else {
		for _, pv := range paidVIPs {
			if err := c.couponSvc.IssueVIPMonthly(ctx, pv.UserID, pv.Level); err != nil {
				slog.Warn("付费 VIP 用户发券失败", slog.Uint64("user_id", uint64(pv.UserID)), slog.Any("err", err))
				failCount++
			} else {
				successCount++
			}
		}
	}

	slog.Info("VIP 月度发券任务完成",
		slog.Int("success", successCount),
		slog.Int("fail", failCount),
		slog.Duration("duration", time.Since(now)),
	)
}

// RunOnce 手动执行一次发券任务（用于测试或补发）
func (c *VIPCouponCron) RunOnce() {
	c.issueMonthlyVIPCoupons()
}
