package service

import (
	"database/sql"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/ego-component/egorm"
	"github.com/gotomicro/ego/core/elog"
	"github.com/pkg/errors"

	"github.com/clickvisual/clickvisual/api/internal/invoker"
	"github.com/clickvisual/clickvisual/api/internal/service/inquiry"
	"github.com/clickvisual/clickvisual/api/internal/service/permission"
	"github.com/clickvisual/clickvisual/api/internal/service/permission/pmsplugin"
	"github.com/clickvisual/clickvisual/api/pkg/constx"
	"github.com/clickvisual/clickvisual/api/pkg/model/db"
	"github.com/clickvisual/clickvisual/api/pkg/model/view"
	"github.com/clickvisual/clickvisual/api/pkg/utils"
)

type instanceManager struct {
	dss sync.Map // datasource list
}

func NewInstanceManager() *instanceManager {
	m := &instanceManager{
		dss: sync.Map{},
	}
	datasourceList, _ := db.InstanceList(egorm.Conds{})
	for _, ds := range datasourceList {
		switch ds.Datasource {
		case db.DatasourceMySQL:
			// TODO Not supported at this time
		case db.DatasourceClickHouse:
			// Test connection, storage
			chDb, err := ClickHouseLink(ds.Dsn)
			if err != nil {
				invoker.Logger.Error("ClickHouse", elog.Any("step", "ClickHouseLink"), elog.Any("error", err.Error()))
				continue
			}
			m.dss.Store(ds.DsKey(), inquiry.NewClickHouse(chDb, ds))
		}
	}
	return m
}

func (i *instanceManager) Delete(key string) {
	i.dss.Delete(key)
	return
}

func (i *instanceManager) Add(obj *db.BaseInstance) error {
	switch obj.Datasource {
	case db.DatasourceClickHouse:
		// Test connection, storage
		chDb, err := ClickHouseLink(obj.Dsn)
		if err != nil {
			invoker.Logger.Error("ClickHouse", elog.Any("step", "ClickHouseLink"), elog.Any("error", err.Error()))
			return err
		}
		i.dss.Store(obj.DsKey(), inquiry.NewClickHouse(chDb, obj))
	}
	return nil
}

func (i *instanceManager) Load(id int) (inquiry.Operator, error) {
	instance, err := db.InstanceInfo(invoker.Db, id)
	if err != nil {
		invoker.Logger.Error("instanceManager", elog.Any("id", id), elog.Any("error", err.Error()))
		return nil, err
	}
	obj, ok := i.dss.Load(db.InstanceKey(id))
	if !ok {
		// try again
		if err = i.Add(&instance); err != nil {
			return nil, constx.ErrInstanceObj
		}
		obj, _ = i.dss.Load(db.InstanceKey(id))
	}
	if obj == nil {
		return nil, constx.ErrInstanceObj
	}
	switch instance.Datasource {
	case db.DatasourceClickHouse:
		return obj.(*inquiry.ClickHouse), nil
	}
	return nil, constx.ErrInstanceObj
}

func (i *instanceManager) All() []inquiry.Operator {
	res := make([]inquiry.Operator, 0)
	i.dss.Range(func(key, obj interface{}) bool {
		iid, _ := strconv.Atoi(key.(string))
		instance, _ := db.InstanceInfo(invoker.Db, iid)
		if instance.Datasource == db.DatasourceClickHouse {
			res = append(res, obj.(*inquiry.ClickHouse))
		}
		return true
	})
	return res
}

func ReadAllPermissionTable(uid int, subResource string) []int {
	tables, _ := db.TableList(invoker.Db, egorm.Conds{})
	resArr := make([]int, 0)
	for _, table := range tables {
		if !TableViewIsPermission(uid, table.Database.Iid, table.ID) {
			invoker.Logger.Error("ReadAllPermissionTable",
				elog.Any("uid", uid),
				elog.Any("iid", table.Database.Iid),
				elog.Any("tid", table.ID),
				elog.Any("subResource", subResource))
			continue
		}
		resArr = append(resArr, table.ID)
	}
	return resArr
}

func InstanceViewIsPermission(uid, iid int) bool {
	if instanceViewIsPermission(uid, iid, pmsplugin.Log) ||
		instanceViewIsPermission(uid, iid, pmsplugin.Alarm) ||
		instanceViewIsPermission(uid, iid, pmsplugin.Pandas) {
		return true
	}
	return false
}

func instanceViewIsPermission(uid int, iid int, subResource string) bool {
	// check instance permission
	if err := permission.Manager.CheckNormalPermission(view.ReqPermission{
		UserId:      uid,
		ObjectType:  pmsplugin.PrefixInstance,
		ObjectIdx:   strconv.Itoa(iid),
		SubResource: subResource,
		Acts:        []string{pmsplugin.ActView},
	}); err == nil {
		invoker.Logger.Debug("ReadAllPermissionInstance",
			elog.Any("uid", uid),
			elog.Any("step", "InstanceViewIsPermission"),
			elog.Any("iid", iid),
			elog.Any("subResource", subResource))
		return true
	}
	// check databases permission
	conds := egorm.Conds{}
	conds["iid"] = iid
	databases, err := db.DatabaseList(invoker.Db, conds)
	if err != nil {
		invoker.Logger.Error("PmsCheckInstanceRead", elog.String("error", err.Error()))
		return false
	}
	for _, d := range databases {
		if databaseViewIsPermission(uid, iid, d.ID, subResource) {
			return true
		}
	}
	return false
}

func ClickHouseLink(dsn string) (conn *sql.DB, err error) {

	invoker.Logger.Debug("clickhouseDsnConvert", elog.String("dsn", utils.ClickhouseDsnConvert(dsn)))

	conn, err = sql.Open("clickhouse", utils.ClickhouseDsnConvert(dsn))
	if err != nil {
		invoker.Logger.Error("ClickHouse", elog.Any("step", "sql.error"), elog.String("error", err.Error()))
		return
	}
	conn.SetMaxIdleConns(5)
	conn.SetMaxOpenConns(10)
	conn.SetConnMaxLifetime(time.Minute * 3)
	if err = conn.Ping(); err != nil {
		invoker.Logger.Error("ClickHouse", elog.String("step", "notException"), elog.Any("error", err.Error()))
		return
	}
	return
}

func InstanceCreate(req view.ReqCreateInstance) (obj db.BaseInstance, err error) {
	conds := egorm.Conds{}
	conds["datasource"] = req.Datasource
	conds["name"] = req.Name
	checks, err := db.InstanceList(conds)
	if err != nil {
		err = errors.Wrap(err, "create DB failed 01: ")
		return
	}
	invoker.Logger.Debug("InstanceCreate", elog.Any("checks", checks))
	if len(checks) > 0 {
		err = errors.New("data source configuration with duplicate name")
		return
	}
	if req.Mode == inquiry.ModeCluster && len(req.Clusters) == 0 {
		err = errors.New("you need to fill in the cluster information")
		return
	}
	obj = db.BaseInstance{
		Datasource:       req.Datasource,
		Name:             req.Name,
		Dsn:              strings.TrimSpace(req.Dsn),
		RuleStoreType:    req.RuleStoreType,
		FilePath:         req.FilePath,
		Desc:             req.Desc,
		ClusterId:        req.ClusterId,
		Namespace:        req.Namespace,
		Configmap:        req.Configmap,
		PrometheusTarget: req.PrometheusTarget,
		ReplicaStatus:    req.ReplicaStatus,
		Mode:             req.Mode,
		Clusters:         req.Clusters,
	}
	invoker.Logger.Debug("instanceCreate", elog.Any("obj", obj))
	if req.PrometheusTarget != "" {
		if err = Alarm.PrometheusReload(req.PrometheusTarget); err != nil {
			err = errors.Wrap(err, "create DB failed 02:")
			return
		}
	}
	tx := invoker.Db.Begin()
	if err = db.InstanceCreate(tx, &obj); err != nil {
		tx.Rollback()
		err = errors.Wrap(err, "create DB failed 03: ")
		return
	}
	if err = InstanceManager.Add(&obj); err != nil {
		tx.Rollback()
		err = errors.Wrap(err, "DNS configuration exception, database connection failure 01: ")
		return
	}
	if err = tx.Commit().Error; err != nil {
		err = errors.Wrap(err, "DNS configuration exception, database connection failure 02: ")
		return
	}
	return obj, nil
}

func DatabaseCreate(req db.BaseDatabase) (out db.BaseDatabase, err error) {
	op, err := InstanceManager.Load(req.Iid)
	if err != nil {
		return
	}
	tx := invoker.Db.Begin()
	if err = db.DatabaseCreate(tx, &req); err != nil {
		err = errors.Wrap(err, "create failed 01:")
		return
	}
	err = op.DatabaseCreate(req.Name, req.Cluster)
	if err != nil {
		tx.Rollback()
		err = errors.Wrap(err, "create failed 02: ")
		return
	}
	if err = tx.Commit().Error; err != nil {
		tx.Rollback()
		err = errors.Wrap(err, "create failed 03: ")
		return
	}
	return req, nil
}

func TableCreate(uid int, databaseInfo db.BaseDatabase, param view.ReqTableCreate) (tableInfo db.BaseTable, err error) {
	op, err := InstanceManager.Load(databaseInfo.Iid)
	if err != nil {
		return
	}
	s, d, v, a, err := op.TableCreate(databaseInfo.ID, databaseInfo, param)
	if err != nil {
		err = errors.Wrap(err, "create failed 01:")
		return
	}
	tableInfo = db.BaseTable{
		Did:            databaseInfo.ID,
		Name:           param.TableName,
		Typ:            param.Typ,
		Days:           param.Days,
		Brokers:        param.Brokers,
		Topic:          param.Topics,
		Desc:           param.Desc,
		SqlData:        d,
		SqlStream:      s,
		SqlView:        v,
		SqlDistributed: a,
		TimeField:      db.TimeFieldSecond,
		CreateType:     inquiry.TableCreateTypeCV,
		Uid:            uid,
	}
	err = db.TableCreate(invoker.Db, &tableInfo)
	if err != nil {
		err = errors.Wrap(err, "create failed 02:")
		return
	}
	return tableInfo, nil
}

func AnalysisFieldsUpdate(tid int, data []view.IndexItem) (err error) {
	var (
		addMap map[string]*db.BaseIndex
		delMap map[string]*db.BaseIndex
		newMap map[string]*db.BaseIndex
	)
	// check repeat
	repeatMap := make(map[string]interface{})
	for _, r := range data {
		if r.Typ == 3 {
			err = errors.New("param error: json type 3 should not in params:" + r.Field)
			return
		}
		key := r.Field
		if r.RootName != "" {
			key = r.RootName + "." + r.Field
		}
		if _, ok := repeatMap[key]; ok {
			err = errors.New("param error: repeat index field name:" + r.Field)
			return
		}
		repeatMap[key] = struct{}{}
	}
	req := view.ReqCreateIndex{
		Tid:  tid,
		Data: data,
	}
	req.Tid = tid
	addMap, delMap, newMap, err = Index.Diff(req)
	if err != nil {
		return
	}
	invoker.Logger.Debug("IndexUpdate", elog.Any("addMap", addMap), elog.Any("delMap", delMap))
	err = Index.Sync(req, addMap, delMap, newMap)
	if err != nil {
		return
	}
	return nil
}

func InstanceFilterPms(uid int) (res []view.RespInstanceSimple, err error) {
	dArr, err := DatabaseListFilterPms(uid)
	if err != nil {
		return
	}
	res = make([]view.RespInstanceSimple, 0)
	iMap := make(map[int]view.RespInstanceSimple)
	// Fill in all database information and verify related permissions
	is, _ := db.InstanceList(egorm.Conds{})
	for _, i := range is {
		if !InstanceViewIsPermission(uid, i.ID) {
			continue
		}
		iMap[i.ID] = view.RespInstanceSimple{
			Id:           i.ID,
			InstanceName: i.Name,
			Desc:         i.Desc,
			Databases:    make([]view.RespDatabaseSimple, 0),
		}
	}
	for _, d := range dArr {
		// exist
		item, ok := iMap[d.Iid]
		if !ok {
			continue
		}
		item.Databases = append(item.Databases, d)
		iMap[d.Iid] = item
	}
	for _, v := range iMap {
		res = append(res, v)
	}
	sort.Slice(res, func(i, j int) bool {
		return res[i].InstanceName < res[j].InstanceName
	})
	return
}
