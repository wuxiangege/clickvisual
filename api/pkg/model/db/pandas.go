package db

import (
	"fmt"
	"time"

	"github.com/ego-component/egorm"
	"github.com/gotomicro/ego/core/elog"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/clickvisual/clickvisual/api/internal/invoker"
)

const (
	CrontabTypNormal int = iota
	CrontabTypSuspended
)

const (
	CrontabStatusWait int = iota
	CrontabStatusPreempt
	CrontabStatusDoing
)

const (
	SourceTypMySQL      = 1
	SourceTypClickHouse = 2
)

const (
	PrimaryMining = 1
	PrimaryShort  = 3
)

const (
	SecondaryAny             = 0
	SecondaryDatabase        = 1
	SecondaryDataIntegration = 2
	SecondaryDataMining      = 3
	SecondaryDashboard       = 4
)

const (
	TertiaryClickHouse   = 10
	TertiaryMySQL        = 11
	TertiaryOfflineSync  = 20
	TertiaryRealTimeSync = 21
)

// 0 No status 2 Executing 3 Abnormal execution 4 Completed
const (
	NodeStatusDefault = 0
	NodeStatusHandler = 2
	NodeStatusError   = 3
	NodeStatusFinish  = 4
)

func (m *BigdataWorkflow) TableName() string {
	return TableNameBigDataWorkflow
}

func (m *BigdataDepend) TableName() string {
	return TableNameBigDataDepend
}

func (m *BigdataCrontab) TableName() string {
	return TableNameBigDataCrontab
}

func (m *BigdataSource) TableName() string {
	return TableNameBigDataSource
}

func (m *BigdataNode) TableName() string {
	return TableNameBigDataNode
}

func (m *BigdataNodeContent) TableName() string {
	return TableNameBigDataNodeContent
}

func (m *BigdataNodeHistory) TableName() string {
	return TableNameBigDataNodeHistory
}

func (m *BigdataNodeResult) TableName() string {
	return TableNameBigDataNodeResult
}

func (m *BigdataFolder) TableName() string {
	return TableNameBigDataFolder
}

type BigdataFolder struct {
	BaseModel

	Uid        int    `gorm:"column:uid;type:int(11)" json:"uid"` // uid of alarm operator
	Iid        int    `gorm:"column:iid;type:int(11)" json:"iid"`
	Name       string `gorm:"column:name;type:varchar(128);NOT NULL" json:"name"` // name of an alarm
	Desc       string `gorm:"column:desc;type:varchar(255);NOT NULL" json:"desc"` // description
	Primary    int    `gorm:"column:primary;type:int(11)" json:"primary"`
	Secondary  int    `gorm:"column:secondary;type:int(11)" json:"secondary"`
	WorkflowId int    `gorm:"column:workflow_id;type:int(11)" json:"workflowId"`
	ParentId   int    `gorm:"column:parent_id;type:int(11)" json:"parentId"`
}

type (
	BigdataNode struct {
		BaseModel

		Uid        int    `gorm:"column:uid;type:int(11)" json:"uid"`
		Iid        int    `gorm:"column:iid;type:int(11)" json:"iid"`
		FolderID   int    `gorm:"column:folder_id;type:int(11)" json:"folderId"`
		Primary    int    `gorm:"column:primary;type:int(11)" json:"primary"`
		Secondary  int    `gorm:"column:secondary;type:int(11)" json:"secondary"`
		Tertiary   int    `gorm:"column:tertiary;type:int(11)" json:"tertiary"`
		WorkflowId int    `gorm:"column:workflow_id;type:int(11)" json:"workflowId"`
		SourceId   int    `gorm:"column:sourceId;type:int(11)" json:"sourceId"`
		Name       string `gorm:"column:name;type:varchar(128);NOT NULL" json:"name"`
		Desc       string `gorm:"column:desc;type:varchar(255);NOT NULL" json:"desc"`
		LockUid    int    `gorm:"column:lock_uid;type:int(11) unsigned" json:"lockUid"`
		LockAt     int64  `gorm:"column:lock_at;type:int(11)" json:"lockAt"`
		Status     int    `gorm:"column:status;type:int(11)" json:"status"` // 0 无状态 1 待执行 2 执行中 3 执行异常 4 执行完成
		UUID       string `gorm:"column:uuid;type:varchar(128)" json:"uuid"`
	}

	BigdataNodeContent struct {
		NodeId          int    `gorm:"column:node_id;type:int(11);uix_node_id,unique" json:"nodeId"`
		Content         string `gorm:"column:content;type:longtext" json:"content"`
		Result          string `gorm:"column:result;type:longtext" json:"result"`
		PreviousContent string `gorm:"column:previous_content;type:longtext" json:"previousContent"`
		Utime           int64  `gorm:"bigint;autoUpdateTime;comment:update time" json:"utime"`
	}

	BigdataNodeHistory struct {
		UUID    string `gorm:"column:uuid;type:varchar(128);uix_uuid,unique" json:"uuid"`
		NodeId  int    `gorm:"column:node_id;type:int(11)" json:"nodeId"`
		Content string `gorm:"column:content;type:longtext" json:"content"`
		Uid     int    `gorm:"column:uid;type:int(11)" json:"uid"`
		Utime   int64  `gorm:"bigint;autoUpdateTime;comment:update time" json:"utime"`
	}

	BigdataNodeResult struct {
		BaseModel

		NodeId       int    `gorm:"column:node_id;type:int(11)" json:"nodeId"`
		Content      string `gorm:"column:content;type:longtext" json:"content"`
		Result       string `gorm:"column:result;type:longtext" json:"result"`
		ExcelProcess string `gorm:"column:excel_process;type:longtext" json:"excelProcess"`
		Uid          int    `gorm:"column:uid;type:int(11)" json:"uid"`
		Cost         int64  `gorm:"column:cost;type:bigint(20)" json:"cost"` // ms
	}
)

type BigdataSource struct {
	BaseModel

	Iid      int    `gorm:"column:iid;type:int(11)" json:"iid"`
	Name     string `gorm:"column:name;type:varchar(128);NOT NULL" json:"name"` // name of an alarm
	Desc     string `gorm:"column:desc;type:varchar(255);NOT NULL" json:"desc"` // description
	URL      string `gorm:"column:url;type:varchar(255);NOT NULL" json:"url"`
	UserName string `gorm:"column:username;type:varchar(255);NOT NULL" json:"username"`
	Password string `gorm:"column:password;type:varchar(255);NOT NULL" json:"password"`
	Typ      int    `gorm:"column:typ;type:int(11)" json:"typ"`
	Uid      int    `gorm:"column:uid;type:int(11)" json:"uid"`
}

type BigdataWorkflow struct {
	BaseModel

	Iid  int    `gorm:"column:iid;type:int(11)" json:"iid"`
	Name string `gorm:"column:name;type:varchar(128);NOT NULL" json:"name"` // name of an alarm
	Desc string `gorm:"column:desc;type:varchar(255);NOT NULL" json:"desc"` // description
	Uid  int    `gorm:"column:uid;type:int(11)" json:"uid"`
}

type BigdataDepend struct {
	Iid                  int     `gorm:"column:iid;type:int(11);index:uix_iid_database_table,unique" json:"iid"`
	Database             string  `gorm:"column:database;type:varchar(128);index:uix_iid_database_table,unique;NOT NULL" json:"database"`
	Table                string  `gorm:"column:table;type:varchar(128);index:uix_iid_database_table,unique;NOT NULL" json:"table"`
	Engine               string  `gorm:"column:engine;type:varchar(128);NOT NULL" json:"engine"`
	DownDepDatabaseTable Strings `gorm:"column:down_dep_database_table;type:text;NOT NULL" json:"downDepDatabaseTable"`
	UpDepDatabaseTable   Strings `gorm:"column:up_dep_database_table;type:text;NOT NULL" json:"upDepDatabaseTable"`
	Rows                 uint64  `gorm:"column:rows;type:bigint(20);default:0;NOT NULL" json:"rows"`
	Bytes                uint64  `gorm:"column:bytes;type:bigint(20);default:0;NOT NULL" json:"bytes"`

	Utime int64 `gorm:"bigint;autoUpdateTime;comment:更新时间" json:"utime"`
}

type BigdataCrontab struct {
	NodeId        int    `gorm:"column:node_id;type:int(11);uix_node_id,unique" json:"nodeId"`
	Desc          string `gorm:"column:desc;type:varchar(255);NOT NULL" json:"desc"` // description
	DutyUid       int    `gorm:"column:duty_uid;type:int(11)" json:"dutyUid"`        // person in charge
	Cron          string `gorm:"column:cron;type:varchar(255);NOT NULL" json:"cron"` // cron expression
	Typ           int    `gorm:"column:typ;type:int(11)" json:"typ"`                 // typ 0 Normal scheduling 1 Suspended scheduling
	Status        int    `gorm:"column:status;type:int(11)" json:"status"`           // status 0 default 1 preempt 2 doing
	Uid           int    `gorm:"column:uid;type:int(11)" json:"uid"`                 // user id
	Args          string `gorm:"args:sql_view;type:text" json:"args"`                // sql_view
	IsRetry       int    `gorm:"column:is_retry;type:tinyint(1)" json:"isRetry"`
	RetryTimes    int    `gorm:"column:retry_times;type:int(11)" json:"retryTimes"`
	RetryInterval int    `gorm:"column:retry_interval;type:int(11)" json:"retryInterval"`
	Ctime         int64  `gorm:"bigint;autoCreateTime;comment:创建时间" json:"ctime"`
	Utime         int64  `gorm:"bigint;autoUpdateTime;comment:更新时间" json:"utime"`
}

func (m *BigdataDepend) Name() string {
	return fmt.Sprintf("%s.%s", m.Database, m.Table)
}

func (m *BigdataDepend) Key() string {
	return fmt.Sprintf("%d.%s.%s", m.Iid, m.Database, m.Table)
}

func DependsInfo(db *gorm.DB, id int) (resp BigdataDepend, err error) {
	var sql = "`id`= ? and dtime = 0"
	var binds = []interface{}{id}
	if err = db.Model(BigdataDepend{}).Where(sql, binds...).First(&resp).Error; err != nil {
		elog.Error("info error", zap.Error(err))
		return
	}
	return
}

func DependsInfoX(conds map[string]interface{}) (resp BigdataDepend, err error) {
	sql, binds := egorm.BuildQuery(conds)
	err = invoker.Db.Table(TableNameBigDataDepend).Where(sql, binds...).First(&resp).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		invoker.Logger.Error("infoX error", zap.Error(err))
		return
	}
	return resp, nil
}

func DependsList(conds egorm.Conds) (resp []*BigdataDepend, err error) {
	sql, binds := egorm.BuildQuery(conds)
	if err = invoker.Db.Model(BigdataDepend{}).Where(sql, binds...).Find(&resp).Error; err != nil {
		elog.Error("list error", zap.Error(err))
		return
	}
	return
}

func DependsUpsList(db *gorm.DB, iid int, database, table string) (resp []*BigdataDepend, err error) {
	var conds = make(map[string]interface{}, 0)
	conds["iid"] = iid
	conds["up_dep_database_table"] = egorm.Cond{
		Op:  "like",
		Val: fmt.Sprintf(`"%s.%s"`, database, table),
	}
	sql, binds := egorm.BuildQuery(conds)
	if err = db.Model(BigdataDepend{}).Where(sql, binds...).Find(&resp).Error; err != nil {
		elog.Error("list error", zap.Error(err))
		return
	}
	return
}

func DependsCreateOrUpdate(db *gorm.DB, data *BigdataDepend) (err error) {
	var row BigdataDepend
	conds := egorm.Conds{}
	conds["iid"] = data.Iid
	conds["database"] = data.Database
	conds["table"] = data.Table
	if row, err = DependsInfoX(conds); err != nil {
		return
	}
	if row.Iid == 0 {
		// create
		if err = db.Model(BigdataDepend{}).Create(data).Error; err != nil {
			elog.Error("create error", zap.Error(err))
			return
		}
		return
	}
	// update
	cu := egorm.Conds{}
	cu["engine"] = data.Engine
	cu["down_dep_database_table"] = data.DownDepDatabaseTable
	cu["up_dep_database_table"] = data.UpDepDatabaseTable
	cu["rows"] = data.Rows
	cu["bytes"] = data.Bytes
	return DependsUpdate(db, data.Iid, data.Database, data.Table, cu)
}

func DependsBatchInsert(db *gorm.DB, rows []*BigdataDepend) (err error) {
	if err = db.Model(BigdataDepend{}).CreateInBatches(rows, len(rows)).Error; err != nil {
		elog.Error("batch create error", zap.Error(err))
		return
	}
	return
}

func DependsUpdate(db *gorm.DB, iid int, database, table string, ups map[string]interface{}) (err error) {
	var sql = "`iid`=? and `database`=? and `table` = ?"
	var binds = []interface{}{iid, database, table}
	if err = db.Model(BigdataDepend{}).Where(sql, binds...).Updates(ups).Error; err != nil {
		elog.Error("update error", zap.Error(err))
		return
	}
	return
}

func DependsDeleteTimeout(db *gorm.DB) (err error) {
	if err = db.Where("utime<?", time.Now().Add(-time.Minute*10).Unix()).Model(BigdataDepend{}).Delete(&BigdataDepend{}).Error; err != nil {
		elog.Error("delete error", zap.Error(err))
		return
	}
	return
}

func DependsDeleteAll(db *gorm.DB, iid int) (err error) {
	if err = db.Where("iid=?", iid).Model(BigdataDepend{}).Delete(&BigdataDepend{}).Error; err != nil {
		elog.Error("delete error", zap.Error(err))
		return
	}
	return
}

func CrontabInfo(db *gorm.DB, nodeId int) (resp BigdataCrontab, err error) {
	var sql = "`node_id`= ?"
	var binds = []interface{}{nodeId}
	if err = db.Model(BigdataCrontab{}).Where(sql, binds...).First(&resp).Error; err != nil {
		elog.Error("info error", zap.Error(err))
		return
	}
	return
}

func CrontabList(conds egorm.Conds) (resp []*BigdataCrontab, err error) {
	sql, binds := egorm.BuildQuery(conds)
	if err = invoker.Db.Model(BigdataCrontab{}).Where(sql, binds...).Find(&resp).Error; err != nil {
		elog.Error("list error", zap.Error(err))
		return
	}
	return
}

func CrontabCreate(db *gorm.DB, data *BigdataCrontab) (err error) {
	if err = db.Model(BigdataCrontab{}).Create(data).Error; err != nil {
		elog.Error("create error", zap.Error(err))
		return
	}
	return
}

func CrontabUpdate(db *gorm.DB, nodeId int, ups map[string]interface{}) (err error) {
	var sql = "`node_id`=?"
	var binds = []interface{}{nodeId}
	if err = db.Model(BigdataCrontab{}).Where(sql, binds...).Updates(ups).Error; err != nil {
		elog.Error("update error", zap.Error(err))
		return
	}
	return
}

func CrontabDelete(db *gorm.DB, nodeId int) (err error) {
	if err = db.Where("node_id=?", nodeId).Model(BigdataCrontab{}).Delete(&BigdataCrontab{}).Error; err != nil {
		elog.Error("delete error", zap.Error(err))
		return
	}
	return
}

func WorkflowInfo(db *gorm.DB, id int) (resp BigdataWorkflow, err error) {
	var sql = "`id`= ? and dtime = 0"
	var binds = []interface{}{id}
	if err = db.Model(BigdataWorkflow{}).Where(sql, binds...).First(&resp).Error; err != nil {
		elog.Error("info error", zap.Error(err))
		return
	}
	return
}

func WorkflowList(conds egorm.Conds) (resp []*BigdataWorkflow, err error) {
	sql, binds := egorm.BuildQuery(conds)
	if err = invoker.Db.Model(BigdataWorkflow{}).Where(sql, binds...).Find(&resp).Error; err != nil {
		elog.Error("list error", zap.Error(err))
		return
	}
	return
}

func WorkflowCreate(db *gorm.DB, data *BigdataWorkflow) (err error) {
	if err = db.Model(BigdataWorkflow{}).Create(data).Error; err != nil {
		elog.Error("create error", zap.Error(err))
		return
	}
	return
}

func WorkflowUpdate(db *gorm.DB, id int, ups map[string]interface{}) (err error) {
	var sql = "`id`=?"
	var binds = []interface{}{id}
	if err = db.Model(BigdataWorkflow{}).Where(sql, binds...).Updates(ups).Error; err != nil {
		elog.Error("update error", zap.Error(err))
		return
	}
	return
}

func WorkflowDelete(db *gorm.DB, id int) (err error) {
	if err = db.Model(BigdataWorkflow{}).Delete(&BigdataWorkflow{}, id).Error; err != nil {
		elog.Error("delete error", zap.Error(err))
		return
	}
	return
}

func SourceInfo(db *gorm.DB, id int) (resp BigdataSource, err error) {
	var sql = "`id`= ? and dtime = 0"
	var binds = []interface{}{id}
	if err = db.Model(BigdataSource{}).Where(sql, binds...).First(&resp).Error; err != nil {
		elog.Error("info error", zap.Error(err))
		return
	}
	return
}

func SourceList(conds egorm.Conds) (resp []*BigdataSource, err error) {
	sql, binds := egorm.BuildQuery(conds)
	if err = invoker.Db.Model(BigdataSource{}).Where(sql, binds...).Find(&resp).Error; err != nil {
		elog.Error("list error", zap.Error(err))
		return
	}
	return
}

func SourceCreate(db *gorm.DB, data *BigdataSource) (err error) {
	if err = db.Model(BigdataSource{}).Create(data).Error; err != nil {
		elog.Error("create error", zap.Error(err))
		return
	}
	return
}

func SourceUpdate(db *gorm.DB, id int, ups map[string]interface{}) (err error) {
	var sql = "`id`=?"
	var binds = []interface{}{id}
	if err = db.Model(BigdataSource{}).Where(sql, binds...).Updates(ups).Error; err != nil {
		elog.Error("update error", zap.Error(err))
		return
	}
	return
}

func SourceDelete(db *gorm.DB, id int) (err error) {
	if err = db.Model(BigdataSource{}).Delete(&BigdataSource{}, id).Error; err != nil {
		elog.Error("delete error", zap.Error(err))
		return
	}
	return
}

func NodeInfo(db *gorm.DB, id int) (resp BigdataNode, err error) {
	var sql = "`id`= ? and dtime = 0"
	var binds = []interface{}{id}
	if err = db.Model(BigdataNode{}).Where(sql, binds...).First(&resp).Error; err != nil {
		elog.Error("info error", zap.Error(err))
		return
	}
	return
}

func NodeList(conds egorm.Conds) (resp []*BigdataNode, err error) {
	sql, binds := egorm.BuildQuery(conds)
	if err = invoker.Db.Model(BigdataNode{}).Where(sql, binds...).Find(&resp).Error; err != nil {
		elog.Error("list error", zap.Error(err))
		return
	}
	return
}

func NodeCreate(db *gorm.DB, data *BigdataNode) (err error) {
	if err = db.Model(BigdataNode{}).Create(data).Error; err != nil {
		elog.Error("create error", zap.Error(err))
		return
	}
	return
}

func NodeUpdate(db *gorm.DB, id int, ups map[string]interface{}) (err error) {
	var sql = "`id`=?"
	var binds = []interface{}{id}
	if err = db.Model(BigdataNode{}).Where(sql, binds...).Updates(ups).Error; err != nil {
		elog.Error("update error", zap.Error(err))
		return
	}
	return
}

func NodeDelete(db *gorm.DB, id int) (err error) {
	if err = db.Model(BigdataNode{}).Delete(&BigdataNode{}, id).Error; err != nil {
		elog.Error("delete error", zap.Error(err))
		return
	}
	return
}

func NodeContentInfo(db *gorm.DB, id int) (resp BigdataNodeContent, err error) {
	var sql = "`node_id`= ?"
	var binds = []interface{}{id}
	if err = db.Model(BigdataNodeContent{}).Where(sql, binds...).First(&resp).Error; err != nil {
		elog.Error("info error", zap.Error(err))
		return
	}
	return
}

func NodeContentCreate(db *gorm.DB, data *BigdataNodeContent) (err error) {
	if err = db.Model(BigdataNodeContent{}).Create(data).Error; err != nil {
		elog.Error("create error", zap.Error(err))
		return
	}
	return
}

func NodeContentUpdate(db *gorm.DB, id int, ups map[string]interface{}) (err error) {
	var sql = "`node_id`=?"
	var binds = []interface{}{id}
	if err = db.Model(BigdataNodeContent{}).Where(sql, binds...).Updates(ups).Error; err != nil {
		elog.Error("update error", zap.Error(err))
		return
	}
	return
}

func NodeContentDelete(db *gorm.DB, id int) (err error) {
	if err = db.Model(BigdataNodeContent{}).Where("node_id=?", id).Unscoped().Delete(&BigdataNodeContent{}).Error; err != nil {
		elog.Error("delete error", zap.Error(err))
		return
	}
	return
}

func NodeHistoryInfo(db *gorm.DB, uuid string) (resp BigdataNodeHistory, err error) {
	var sql = "`uuid`= ?"
	var binds = []interface{}{uuid}
	if err = db.Model(BigdataNodeHistory{}).Where(sql, binds...).First(&resp).Error; err != nil {
		elog.Error("info error", zap.Error(err))
		return
	}
	return
}

func NodeHistoryListPage(conds egorm.Conds, reqList *ReqPage) (total int64, respList []*BigdataNodeHistory) {
	respList = make([]*BigdataNodeHistory, 0)
	if reqList.PageSize == 0 {
		reqList.PageSize = 10
	}
	if reqList.Current == 0 {
		reqList.Current = 1
	}
	sql, binds := egorm.BuildQuery(conds)
	db := invoker.Db.Select("uuid, utime, uid").Model(BigdataNodeHistory{}).Where(sql, binds...).Order("utime desc")
	db.Count(&total)
	db.Offset((reqList.Current - 1) * reqList.PageSize).Limit(reqList.PageSize).Find(&respList)
	return
}

func NodeHistoryCreate(db *gorm.DB, data *BigdataNodeHistory) (err error) {
	if err = db.Model(BigdataNodeHistory{}).Create(data).Error; err != nil {
		elog.Error("create error", zap.Error(err))
		return
	}
	return
}

func NodeResultInfo(db *gorm.DB, id int) (resp BigdataNodeResult, err error) {
	var sql = "`id`= ? and dtime = 0"
	var binds = []interface{}{id}
	if err = db.Model(BigdataNodeResult{}).Where(sql, binds...).First(&resp).Error; err != nil {
		elog.Error("info error", zap.Error(err))
		return
	}
	return
}

func NodeResultList(conds egorm.Conds) (resp []*BigdataNodeResult, err error) {
	sql, binds := egorm.BuildQuery(conds)
	if err = invoker.Db.Model(BigdataNodeResult{}).Where(sql, binds...).Find(&resp).Error; err != nil {
		elog.Error("list error", zap.Error(err))
		return
	}
	return
}

func NodeResultCreate(db *gorm.DB, data *BigdataNodeResult) (err error) {
	if err = db.Model(BigdataNodeResult{}).Create(data).Error; err != nil {
		elog.Error("create error", zap.Error(err))
		return
	}
	return
}

func NodeResultUpdate(db *gorm.DB, id int, ups map[string]interface{}) (err error) {
	var sql = "`id`=?"
	var binds = []interface{}{id}
	if err = db.Model(BigdataNodeResult{}).Where(sql, binds...).Updates(ups).Error; err != nil {
		elog.Error("update error", zap.Error(err))
		return
	}
	return
}

func NodeResultDelete(db *gorm.DB, id int) (err error) {
	if err = db.Model(BigdataNodeResult{}).Delete(&BigdataNodeResult{}, id).Error; err != nil {
		elog.Error("delete error", zap.Error(err))
		return
	}
	return
}

func NodeResultDelete30Days() {
	expire := time.Hour * 24 * 30
	if err := invoker.Db.Model(BigdataNodeResult{}).Where("ctime<?", time.Now().Add(-expire).Unix()).Unscoped().Delete(&BigdataNodeResult{}).Error; err != nil {
		elog.Error("delete error", zap.Error(err))
		return
	}
}

func NodeResultListPage(conds egorm.Conds, reqList *ReqPage) (total int64, respList []*BigdataNodeResult) {
	respList = make([]*BigdataNodeResult, 0)
	if reqList.PageSize == 0 {
		reqList.PageSize = 10
	}
	if reqList.Current == 0 {
		reqList.Current = 1
	}
	sql, binds := egorm.BuildQuery(conds)
	db := invoker.Db.Select("id, ctime, uid, cost").Model(BigdataNodeResult{}).Where(sql, binds...).Order("id desc")
	db.Count(&total)
	db.Offset((reqList.Current - 1) * reqList.PageSize).Limit(reqList.PageSize).Find(&respList)
	return
}

func FolderInfo(db *gorm.DB, id int) (resp BigdataFolder, err error) {
	var sql = "`id`= ? and dtime = 0"
	var binds = []interface{}{id}
	if err = db.Model(BigdataFolder{}).Where(sql, binds...).First(&resp).Error; err != nil {
		elog.Error("release info error", zap.Error(err))
		return
	}
	return
}

func FolderList(conds egorm.Conds) (resp []*BigdataFolder, err error) {
	sql, binds := egorm.BuildQuery(conds)
	if err = invoker.Db.Model(BigdataFolder{}).Where(sql, binds...).Find(&resp).Error; err != nil {
		elog.Error("Deployment list error", zap.Error(err))
		return
	}
	return
}

func FolderCreate(db *gorm.DB, data *BigdataFolder) (err error) {
	if err = db.Model(BigdataFolder{}).Create(data).Error; err != nil {
		elog.Error("create error", zap.Error(err))
		return
	}
	return
}

func FolderUpdate(db *gorm.DB, id int, ups map[string]interface{}) (err error) {
	var sql = "`id`=?"
	var binds = []interface{}{id}
	if err = db.Model(BigdataFolder{}).Where(sql, binds...).Updates(ups).Error; err != nil {
		elog.Error("update error", zap.Error(err))
		return
	}
	return
}

func FolderDeleteBatch(db *gorm.DB, tid int) (err error) {
	if err = db.Model(BigdataFolder{}).Where("`tid`=?", tid).Unscoped().Delete(&BigdataFolder{}).Error; err != nil {
		elog.Error("release delete error", zap.Error(err))
		return
	}
	return
}

func FolderDelete(db *gorm.DB, id int) (err error) {
	if err = db.Model(BigdataFolder{}).Delete(&BigdataFolder{}, id).Error; err != nil {
		elog.Error("release delete error", zap.Error(err))
		return
	}
	return
}
