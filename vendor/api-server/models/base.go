/*
 * Licensed Materials - Property of tenxcloud.com
 * (C) Copyright 2016 TenxCloud. All Rights Reserved.
 * 2016-09-24  @author Zhao Shuailong
 */

package models

import (
	"fmt"
	"net"
	"strings"
	"time"

	"github.com/astaxie/beego"
	"github.com/astaxie/beego/orm"
	"github.com/go-sql-driver/mysql"
	"github.com/golang/glog"

	"api-server/models/alert"
	"api-server/models/blobs"
	"api-server/models/cluster"
	"api-server/models/common"
	"api-server/models/integration"
	"api-server/models/labels"
	"api-server/models/team"
	"api-server/models/team2resource"
	"api-server/models/team2user"
	"api-server/models/teamspace"
	"api-server/models/teamspace2user"
	"api-server/models/user"
	"api-server/models/proxy"
)

const (
	// TimeLayout time layout to format timestamp
	TimeLayout = "2006-01-02 15:04:05"
)

var (
	// GetSQLConfig returns mysql config read from app.conf
	// it is defined as a function for test purpose
	GetSQLConfig = func() (driver string, dsn []string, maxIdle, maxConn int, isDebug bool, err error) {
		driver = beego.AppConfig.String("db_driver")
		dbnameStr := beego.AppConfig.String("db_names")
		host := beego.AppConfig.String("db_host")
		port := beego.AppConfig.String("db_port")
		user := beego.AppConfig.String("db_user")
		pass := beego.AppConfig.String("db_password")
		maxIdle, _ = beego.AppConfig.Int("db_max_idle")
		maxConn, _ = beego.AppConfig.Int("db_max_conn")

		isDebug = beego.AppConfig.String("RunMode") != "pro"
		// debug 模式下，允许延迟设置数据库
		if host == "" || user == "" || pass == "" {
			err = fmt.Errorf("MySQL配置有误, host:%s, port:%s, user:%s, pass:%s, if you're running in tests, you can ignore it", host, port, user, pass)
			return
		}

		/*var dbnames []string
		if items := strings.Split(dbnameStr, ","); len(items) > 0 {
			for _, item := range items {
				item = strings.Trim(item, " ")
				if len(item) > 0 {
					dbnames = append(dbnames, item)
				}
			}
		}*/
		dbnameStr = strings.Trim(dbnameStr, "")
		if dbnameStr == "" {
			err = fmt.Errorf("No available databases")
			return
		}
		common.DatabaseDefault = dbnameStr
		// no fault tolerance done here
		/*for idx, name := range dbnames {
			if idx == 0 {
				common.DatabaseDefault = name
			} else if idx == 1 {
				common.Database_2_0 = name
			}
		}*/

		glog.V(4).Infof("db names:%v, host: %s, port:%s\n", dbnameStr, host, port)
		switch driver {
		case "mysql", "":
			driver = "mysql"
			if port == "" {
				port = "3306"
			}
			loc, err := time.LoadLocation("Asia/Shanghai")
			if err != nil {
				loc = time.Local
			}
			c := mysql.Config{
				User:   user,
				Passwd: pass,
				Net:    "tcp",
				Addr:   net.JoinHostPort(host, port),
				Loc:    loc,
				Params: map[string]string{"charset": "utf8"},
			}
			//for _, name := range dbnames {
			c.DBName = dbnameStr
			dsn = append(dsn, c.FormatDSN())
			//}
		case "postgres":
			common.DType = common.DRPostgres
			var configs []string
			configs = append(configs, "user="+user)
			configs = append(configs, "password="+pass)
			configs = append(configs, "host="+host)
			configs = append(configs, "port="+port)
			configs = append(configs, "sslmode=disable")
			configStr := strings.Join(configs, " ")
			//for _, name := range dbnames {
			dsn = append(dsn, configStr+" dbname="+dbnameStr)
			//}
		default:
			err = fmt.Errorf("driver %s is not supported", driver)
			return
		}
		return driver, dsn, maxIdle, maxConn, isDebug, nil
	}
)

func init() {
	driver, dsn, dbmaxIdle, dbmaxConn, isDebug, err := GetSQLConfig()
	if err != nil {
		glog.Errorf("models.init failed to get sql config: %v\n", err)
		if isDebug == false { // production, invalid config causes a panic
			panic(err)
		}
	} else {
		InitializeSQLPool(driver, dsn, dbmaxIdle, dbmaxConn, isDebug)
	}
}

// InitializeSQLPool initializes mysql/postgres connection.
func InitializeSQLPool(driver string, dsn []string, maxIdle, maxConn int, isDebug bool) {
	method := "InitializeSQLPool"
	fmt.Printf("Init sql pool, driver:%s, dsn:%s\n", driver, dsn)
	err := orm.RegisterDataBase("default", driver, dsn[0], maxIdle, maxConn)
	if err != nil {
		glog.Errorln(method, "failed to register database", err)
		panic(err)
	}

	orm.RegisterModel(new(user.UserModel), new(user.UserAccountModel), new(user.InviteCodeModel), new(user.ThirdPartyAccountModel))
	orm.RegisterModel(new(team.TeamModel), new(teamspace.SpaceModel), new(teamspace.SpaceAccountModel))
	orm.RegisterModel(new(team2user.TeamUserModel), new(teamspace2user.SpaceUserModel))
	orm.RegisterModel(new(cluster.ClusterModel))
	orm.RegisterModel(new(team2resource.TeamResourceModel))
	orm.RegisterModel(new(App), new(Permission), new(Role), new(ServicePorts), new(Volume), new(VolumeUsedInfo), new(VolumeLog), new(AppTemplate), new(Snapshot))
	orm.RegisterModel(new(Sec), new(CiFlows), new(Configs), new(License), new(Platform))
	orm.RegisterModel(new(integration.Record), new(user.UsersConsumption), new(user.UsersCharge))
	orm.RegisterModel(new(ResourceQuota))
	orm.RegisterModel(new(EditionRecord))
	orm.RegisterModel(new(BalanceNotifyRules), new(team.Invitation))
	orm.RegisterModel(new(CertificateModel))
	orm.RegisterModel(new(alert.NotifyStrategy), new(alert.NotifyHistory), new(alert.NotifyReceiverGroup), new(alert.NotifyReceiverInvitation))
	orm.RegisterModel(new(labels.LabelModel))
	orm.RegisterModel(new(blobs.Blob))
	orm.RegisterModel(new(ServiceTimeScaleRecord))
	orm.RegisterModel(new(proxy.ProxyGroup),new(proxy.ProxyHost),new(proxy.ProxySpace))
	// orm.RunSyncdb("default", true, true)
	orm.Debug = false
}
