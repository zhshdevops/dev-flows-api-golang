/*
 * Licensed Materials - Property of tenxcloud.com
 * (C) Copyright 2016 TenxCloud. All Rights Reserved.
 * 2016-11-01  @author zhao shuailong
 */

package common

var (
	// DType Represents the type of driver
	DType = DRMySQL

	// DatabaseDefault is the default database name
	DatabaseDefault string
)

type DriverType int

// Enum the Database driver
const (
	_          DriverType = iota // int enum type
	DRMySQL                      // mysql
	DRSqlite                     // sqlite
	DROracle                     // oracle
	DRPostgres                   // pgsql
	DRTiDB                       // TiDB
)

func (t DriverType) String() string {
	switch t {
	case DRMySQL:
		return "mysql"
	case DRPostgres:
		return "postgres"
	case DROracle:
		return "oracle"
	case DRSqlite:
		return "sqlite"
	default:
		return ""
	}
}

// ListMeta describes list of objects, i.e. holds information about pagination options set for the
// list.
type ListMeta struct {
	// Total number of items on the list. Used for pagination.
	Total int `orm:"column(total)" json:"total"`
}
