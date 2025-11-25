package service

import (
	"errors"

	"github.com/go-sql-driver/mysql"
)

// isMySQLDuplicate 判断是否为 MySQL 唯一键冲突。
func isMySQLDuplicate(err error) bool {
	var mysqlErr *mysql.MySQLError
	if errors.As(err, &mysqlErr) {
		return mysqlErr.Number == 1062
	}
	return false
}
