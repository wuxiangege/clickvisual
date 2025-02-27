package service

import (
	"sort"
	"strconv"

	"github.com/ego-component/egorm"
	"github.com/gotomicro/ego/core/elog"

	"github.com/clickvisual/clickvisual/api/internal/invoker"
	"github.com/clickvisual/clickvisual/api/internal/service/permission"
	"github.com/clickvisual/clickvisual/api/internal/service/permission/pmsplugin"
	"github.com/clickvisual/clickvisual/api/pkg/model/db"
	"github.com/clickvisual/clickvisual/api/pkg/model/view"
)

func DatabaseListFilterPms(uid int) (res []view.RespDatabaseSimple, err error) {
	res = make([]view.RespDatabaseSimple, 0)
	dMap := make(map[int]view.RespDatabaseSimple)
	// Fill in all database information and verify related permissions
	ds, _ := db.DatabaseList(invoker.Db, egorm.Conds{})
	for _, d := range ds {
		if !DatabaseViewIsPermission(uid, d.Iid, d.ID) {
			continue
		}
		dMap[d.ID] = view.RespDatabaseSimple{
			Id:           d.ID,
			Iid:          d.Iid,
			DatabaseName: d.Name,
			IsCreateByCV: d.IsCreateByCV,
			Desc:         d.Desc,
			Cluster:      d.Cluster,
			Tables:       make([]view.RespTableSimple, 0),
		}
	}
	ts, err := db.TableList(invoker.Db, egorm.Conds{})
	if err != nil {
		return
	}
	for _, row := range ts {
		item, ok := dMap[row.Database.ID]
		if !ok {
			continue
		}
		if !TableViewIsPermission(uid, row.Database.Iid, row.ID) {
			continue
		}
		respTableSimple := view.RespTableSimple{
			Id:         row.ID,
			Did:        row.Database.ID,
			TableName:  row.Name,
			CreateType: row.CreateType,
			Desc:       row.Desc,
		}
		item.Tables = append(item.Tables, respTableSimple)
		dMap[row.Database.ID] = item
	}
	for _, v := range dMap {
		res = append(res, v)
	}
	sort.Slice(res, func(i, j int) bool {
		return res[i].DatabaseName < res[j].DatabaseName
	})
	return
}

func DatabaseViewIsPermission(uid, iid, did int) bool {
	if databaseViewIsPermission(uid, iid, did, pmsplugin.Log) ||
		databaseViewIsPermission(uid, iid, did, pmsplugin.Alarm) ||
		databaseViewIsPermission(uid, iid, did, pmsplugin.Pandas) {
		return true
	}
	return false
}

func databaseViewIsPermission(uid, iid, did int, subResource string) bool {
	// check database permission
	if err := permission.Manager.CheckNormalPermission(view.ReqPermission{
		UserId:      uid,
		ObjectType:  pmsplugin.PrefixInstance,
		ObjectIdx:   strconv.Itoa(iid),
		SubResource: subResource,
		Acts:        []string{pmsplugin.ActView},
		DomainType:  pmsplugin.PrefixDatabase,
		DomainId:    strconv.Itoa(did),
	}); err == nil {
		invoker.Logger.Debug("ReadAllPermissionInstance",
			elog.Any("uid", uid),
			elog.Any("step", "DatabaseViewIsPermission"),
			elog.Any("iid", iid),
			elog.Any("did", did),
			elog.Any("subResource", subResource))
		return true
	}
	// check databases permission
	conds := egorm.Conds{}
	conds["did"] = did
	tables, err := db.TableList(invoker.Db, conds)
	if err != nil {
		invoker.Logger.Error("PmsCheckInstanceRead", elog.String("error", err.Error()))
		return false
	}
	for _, t := range tables {
		if tableViewIsPermission(uid, iid, t.ID, subResource) {
			return true
		}
	}
	return false
}
