package service

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"strings"

	"github.com/redis/go-redis/v9"
)

var ErrRiskEntryTypeInvalid = errors.New("风控名单类型无效")

type RiskService struct {
	rdb *redis.Client
}

func NewRiskService(rdb *redis.Client) *RiskService {
	return &RiskService{rdb: rdb}
}

func (s *RiskService) ListBlacklist(ctx context.Context) (ips, users []string, err error) {
	return s.list(ctx, "black")
}

func (s *RiskService) AddBlacklist(ctx context.Context, entryType, value string) error {
	return s.add(ctx, "black", entryType, value)
}

func (s *RiskService) RemoveBlacklist(ctx context.Context, entryType, value string) error {
	return s.remove(ctx, "black", entryType, value)
}

func (s *RiskService) ListGraylist(ctx context.Context) (ips, users []string, err error) {
	return s.list(ctx, "gray")
}

func (s *RiskService) AddGraylist(ctx context.Context, entryType, value string) error {
	return s.add(ctx, "gray", entryType, value)
}

func (s *RiskService) RemoveGraylist(ctx context.Context, entryType, value string) error {
	return s.remove(ctx, "gray", entryType, value)
}

func (s *RiskService) list(ctx context.Context, listType string) (ips, users []string, err error) {
	if s.rdb == nil {
		return nil, nil, fmt.Errorf("redis is nil")
	}
	ipKey, userKey, err := riskKeys(listType)
	if err != nil {
		return nil, nil, err
	}

	ips, err = s.rdb.SMembers(ctx, ipKey).Result()
	if err != nil {
		return nil, nil, err
	}
	users, err = s.rdb.SMembers(ctx, userKey).Result()
	if err != nil {
		return nil, nil, err
	}

	sort.Strings(ips)
	sort.Strings(users)
	return ips, users, nil
}

func (s *RiskService) add(ctx context.Context, listType, entryType, value string) error {
	if s.rdb == nil {
		return fmt.Errorf("redis is nil")
	}
	value = strings.TrimSpace(value)
	if value == "" {
		return fmt.Errorf("风控名单值不能为空")
	}
	key, err := riskEntryKey(listType, entryType)
	if err != nil {
		return err
	}
	return s.rdb.SAdd(ctx, key, value).Err()
}

func (s *RiskService) remove(ctx context.Context, listType, entryType, value string) error {
	if s.rdb == nil {
		return fmt.Errorf("redis is nil")
	}
	value = strings.TrimSpace(value)
	if value == "" {
		return fmt.Errorf("风控名单值不能为空")
	}
	key, err := riskEntryKey(listType, entryType)
	if err != nil {
		return err
	}
	return s.rdb.SRem(ctx, key, value).Err()
}

func riskKeys(listType string) (string, string, error) {
	switch listType {
	case "black":
		return "risk:ip:black", "risk:user:black", nil
	case "gray":
		return "risk:ip:gray", "risk:user:gray", nil
	default:
		return "", "", ErrRiskEntryTypeInvalid
	}
}

func riskEntryKey(listType, entryType string) (string, error) {
	ipKey, userKey, err := riskKeys(listType)
	if err != nil {
		return "", err
	}
	switch strings.TrimSpace(entryType) {
	case "ip":
		return ipKey, nil
	case "user":
		return userKey, nil
	default:
		return "", ErrRiskEntryTypeInvalid
	}
}
