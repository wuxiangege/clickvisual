package rtsync

import (
	"fmt"

	"github.com/gotomicro/ego/core/elog"

	"github.com/clickvisual/clickvisual/api/internal/invoker"
	"github.com/clickvisual/clickvisual/api/internal/service/bigdata/source"
	"github.com/clickvisual/clickvisual/api/internal/service/inquiry"
	"github.com/clickvisual/clickvisual/api/pkg/model/db"
	"github.com/clickvisual/clickvisual/api/pkg/model/view"
)

type ClickHouse2MySQL struct {
	iid          int
	nodeId       int
	sc           *view.SyncContent
	involvedSQLs map[string]string
}

// Stop
// Drop Materialized View -> MySQL engine
func (c *ClickHouse2MySQL) Stop() error {
	var (
		ins db.BaseInstance
		err error
	)
	if ins, err = db.InstanceInfo(invoker.Db, c.iid); err != nil {
		invoker.Logger.Error("ClickHouse2MySQL", elog.String("step", "instanceInfo"), elog.String("error", err.Error()))
		return err
	}
	if err = dropMaterialView(ins, c.nodeId, c.sc); err != nil {
		return err
	}
	return nil
}

// Run
// Materialized View -> MySQL engine
// CREATE DATABASE [IF NOT EXISTS] db_name [ON CLUSTER cluster]
// ENGINE = MySQL('host:port', 'database', 'user', 'password')
func (c *ClickHouse2MySQL) Run() (map[string]string, error) {
	var (
		ins db.BaseInstance
		err error
	)
	c.involvedSQLs = make(map[string]string)
	if ins, err = db.InstanceInfo(invoker.Db, c.iid); err != nil {
		invoker.Logger.Error("ClickHouse2MySQL", elog.String("step", "instanceInfo"), elog.String("error", err.Error()))
		return c.involvedSQLs, err
	}
	if err = c.mysqlEngineDatabase(ins, c.sc); err != nil {
		invoker.Logger.Error("ClickHouse2MySQL", elog.String("step", "mysqlEngineDatabase"), elog.Any("involvedSQLs", c.involvedSQLs), elog.String("error", err.Error()))
		return c.involvedSQLs, err
	}
	if err = c.execTargetSQL(c.sc.Target.TargetBefore); err != nil {
		invoker.Logger.Error("ClickHouse2MySQL", elog.String("step", "TargetBefore"), elog.String("error", err.Error()))
		return c.involvedSQLs, err
	}
	if err = c.materializedView(ins); err != nil {
		invoker.Logger.Error("ClickHouse2MySQL", elog.String("step", "c2mMaterialView"), elog.Any("involvedSQLs", c.involvedSQLs), elog.String("error", err.Error()))
		return c.involvedSQLs, err
	}
	if err = c.execTargetSQL(c.sc.Target.TargetAfter); err != nil {
		invoker.Logger.Error("ClickHouse2MySQL", elog.String("step", "TargetAfter"), elog.String("error", err.Error()))
		return c.involvedSQLs, err
	}
	return c.involvedSQLs, err
}

func (c *ClickHouse2MySQL) materializedView(ins db.BaseInstance) error {
	viewClusterInfo := materialView(c.sc)
	if ins.Mode == inquiry.ModeCluster {
		viewClusterInfo = fmt.Sprintf("%s ON CLUSTER '%s'", viewClusterInfo, c.sc.Cluster())
	}
	// Deletes the materialized view from the last execution
	if err := dropMaterialView(ins, c.nodeId, c.sc); err != nil {
		return err
	}
	viewClusterInfo = fmt.Sprintf("%s TO `%s`.`%s` AS", viewClusterInfo, mysqlEngineDatabaseName(c.sc), c.sc.Target.Table)
	sourceTableName := fmt.Sprintf("`%s`.`%s`", c.sc.Source.Database, c.sc.Source.Table)
	completeSQL := fmt.Sprintf("CREATE MATERIALIZED VIEW %s SELECT %s FROM %s WHERE %s",
		viewClusterInfo, mapping(c.sc.Mapping), sourceTableName, where(c.sc.Source.SourceFilter))

	invoker.Logger.Debug("ClickHouse2MySQL", elog.String("step", "c2mMaterialView"), elog.String("completeSQL", completeSQL))

	c.involvedSQLs["c2mMaterialView"] = completeSQL

	return source.Instantiate(&source.Source{
		DSN: ins.Dsn,
		Typ: db.SourceTypClickHouse,
	}).Exec(completeSQL)
}

func (c *ClickHouse2MySQL) execTargetSQL(sql string) error {
	if sql == "" {
		return nil
	}
	mysqlTarget, err := db.SourceInfo(invoker.Db, c.sc.Target.SourceId)
	if err != nil {
		return err
	}
	return source.Instantiate(&source.Source{
		URL:      mysqlTarget.URL,
		UserName: mysqlTarget.UserName,
		Password: mysqlTarget.Password,
		Typ:      db.SourceTypMySQL,
	}).Exec(sql)
}

func (c *ClickHouse2MySQL) mysqlEngineDatabase(ins db.BaseInstance, sc *view.SyncContent) (err error) {
	// 创建在 clickhouse 中的表是否对用户可见？如果不可见，涉及集群操作，默认采用第一集群？
	dbNameClusterInfo := mysqlEngineDatabaseName(sc)
	if ins.Mode == inquiry.ModeCluster {
		dbNameClusterInfo = fmt.Sprintf("`%s` ON CLUSTER '%s'", dbNameClusterInfo, sc.Cluster())
	}
	s, err := db.SourceInfo(invoker.Db, sc.Target.SourceId)
	if err != nil {
		return
	}
	completeSQL := fmt.Sprintf("CREATE DATABASE IF NOT EXISTS %s ENGINE = MySQL('%s', '%s', '%s', '%s');",
		dbNameClusterInfo,
		s.URL,
		sc.Target.Database,
		s.UserName,
		s.Password)
	invoker.Logger.Debug("ClickHouse2MySQL", elog.String("step", "mysqlEngineDatabase"), elog.String("completeSQL", completeSQL))
	c.involvedSQLs["mysqlEngineDatabase"] = completeSQL
	return source.Instantiate(&source.Source{
		DSN: ins.Dsn,
		Typ: db.SourceTypClickHouse,
	}).Exec(completeSQL)
}
