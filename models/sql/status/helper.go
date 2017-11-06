/*
 * Licensed Materials - Property of tenxcloud.com
 * (C) Copyright 2016 TenxCloud. All Rights Reserved.
 * 2016-09-26  @author Zhao Shuailong
 */

package status

import (
	"github.com/astaxie/beego/orm"
	"github.com/go-sql-driver/mysql"
	"github.com/lib/pq"
)

// ParseErrorCode consumes an simple error, and produces an error code
func ParseErrorCode(err error) (uint32, error) {
	if err == nil {
		return SQLSuccess, nil
	}

	// check orm error defined by beego
	switch err {
	case orm.ErrNoRows:
		return SQLErrNoRowFound, err
	case orm.ErrMultiRows:
		return SQLErrMultiRows, err
	case orm.ErrArgs:
		return SQLErrInvalidMultiInsert, err
	}

	// check if error is mysql specified error
	if driverErr, ok := err.(*mysql.MySQLError); ok {
		if errCode, ok := mysqlErrorMapping[driverErr.Number]; ok {
			return errCode, driverErr
		}

		// mysql error not included in the mapping
		return SQLErrUnCategoried, driverErr
	}

	if driverErr, ok := err.(*pq.Error); ok {
		if errCode, ok := postgresErrorMapping[driverErr.Code.Name()]; ok {
			return errCode, driverErr
		}

		// postgres error not included in the mapping
		return SQLErrUnCategoried, driverErr
	}

	return SQLErrUnCategoried, err
}
