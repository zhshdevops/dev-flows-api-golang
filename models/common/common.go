package common

var (
	// DType Represents the type of driver
	DType = DRMySQL

	// DatabaseDefault is the default database name
	DatabaseDefault string
	HarborServerUrl string
	CICD_REPO_CLONE_IMAGE string = "qinzhao-harbor/clone-repo:v2.2"
	CLONE_LOCATION = "/app"
)

const (
	CUSTOM_STAGE_TYPE = 5
	DEFAULT_PAGE_SIZE = 10
	Default_push_project = "public"

	STATUS_SUCCESS = 0
	STATUS_FAILED = 1
	STATUS_BUILDING = 2
	STATUS_WAITING = 3
	DEFAULT_PAGE_NUMBER = 1
	Time_NIL = "0001-01-01 00:00:00 +0000 UTC"//时间类型的值
	BUILD_DIR = "/enncloud/build_cache/"
	BUILD_IMAGE_STAGE_TYPE = 3
	//镜像仓库类型：1-为本地仓库 2-为DockerHub 3-为自定义
	CUSTOM_REGISTRY=3
	MANUAL_STOP_LABEL="enn-manual-stop-flag"
)



type DriverType int

// Enum the Database driver

const (
	_ DriverType = iota // int enum type
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
