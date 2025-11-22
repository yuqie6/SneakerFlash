package utils

import (
	"fmt"
	"sync"

	"github.com/bwmarrin/snowflake"
)

var (
	node *snowflake.Node
	once sync.Once
)

// 初始化雪花节点
func InitSnowflake(machineID int64) (err error) {
	once.Do(func() {
		node, err = snowflake.NewNode(machineID)
	})
	if err != nil {
		return fmt.Errorf("雪花算法初始化失败: %w", err)
	}
	return nil
}

// 生成雪花 id
func GenSnowflakeID() (string, error) {
	if node == nil {
		return "", fmt.Errorf("雪花算法未初始化")
	}
	return node.Generate().String(), nil
}

// 生成 int64 类型 的 id
func GenSnowflakeIDInt64() (int64, error) {
	if node == nil {
		return 0, fmt.Errorf("雪花算法未初始化")
	}
	return node.Generate().Int64(), nil
}
