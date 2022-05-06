package alarm

import (
	"strconv"

	"github.com/google/uuid"
	"github.com/gotomicro/ego-component/egorm"
	"github.com/gotomicro/ego/core/elog"
	"github.com/spf13/cast"

	"github.com/clickvisual/clickvisual/api/internal/invoker"
	"github.com/clickvisual/clickvisual/api/internal/service"
	"github.com/clickvisual/clickvisual/api/internal/service/event"
	"github.com/clickvisual/clickvisual/api/internal/service/permission"
	"github.com/clickvisual/clickvisual/api/internal/service/permission/pmsplugin"
	"github.com/clickvisual/clickvisual/api/pkg/component/core"
	"github.com/clickvisual/clickvisual/api/pkg/model/db"
	"github.com/clickvisual/clickvisual/api/pkg/model/view"
)

func Create(c *core.Context) {
	var req view.ReqAlarmCreate
	if err := c.Bind(&req); err != nil {
		c.JSONE(1, "invalid parameter: "+err.Error(), nil)
		return
	}
	var tid int
	for _, f := range req.Filters {
		if f.SetOperatorTyp == 0 {
			if tid != 0 {
				c.JSONE(1, "invalid parameter: only one default table allowed", nil)
				return
			}
			tid = f.Tid
		}
	}
	tableInfo, err := db.TableInfo(invoker.Db, tid)
	if err = permission.Manager.CheckNormalPermission(view.ReqPermission{
		UserId:      c.Uid(),
		ObjectType:  pmsplugin.PrefixInstance,
		ObjectIdx:   strconv.Itoa(tableInfo.Database.Iid),
		SubResource: pmsplugin.Alarm,
		Acts:        []string{pmsplugin.ActEdit},
		DomainType:  pmsplugin.PrefixTable,
		DomainId:    strconv.Itoa(tableInfo.ID),
	}); err != nil {
		c.JSONE(1, err.Error(), nil)
		return
	}
	tx := invoker.Db.Begin()
	obj := &db.Alarm{
		Tid:        tid,
		Uuid:       uuid.NewString(),
		Name:       req.Name,
		Desc:       req.Desc,
		Interval:   req.Interval,
		Unit:       req.Unit,
		Tags:       req.Tags,
		ChannelIds: db.Ints(req.ChannelIds),
		Uid:        c.Uid(),
	}
	if err = db.AlarmCreate(tx, obj); err != nil {
		tx.Rollback()
		c.JSONE(1, "alarm create failed 01: "+err.Error(), nil)
		return
	}
	err = service.Alarm.CreateOrUpdate(tx, obj, req)
	if err != nil {
		tx.Rollback()
		c.JSONE(1, "alarm create failed 02: "+err.Error(), nil)
		return
	}
	if err = tx.Commit().Error; err != nil {
		tx.Rollback()
		c.JSONE(1, "alarm create failed 03: "+err.Error(), nil)
		return
	}
	event.Event.AlarmCMDB(c.User(), db.OpnAlarmsCreate, map[string]interface{}{"obj": obj})
	c.JSONOK()
	return
}

func Update(c *core.Context) {
	id := cast.ToInt(c.Param("id"))
	if id == 0 {
		c.JSONE(1, "invalid parameter", nil)
		return
	}
	var (
		req view.ReqAlarmCreate
		err error
	)
	if err = c.Bind(&req); err != nil {
		c.JSONE(1, "invalid parameter: "+err.Error(), nil)
		return
	}

	instanceInfo, tableInfo, alarmInfo, errAlarmInfo := db.GetAlarmTableInstanceInfo(id)
	if err = permission.Manager.CheckNormalPermission(view.ReqPermission{
		UserId:      c.Uid(),
		ObjectType:  pmsplugin.PrefixInstance,
		ObjectIdx:   strconv.Itoa(tableInfo.Database.Iid),
		SubResource: pmsplugin.Alarm,
		Acts:        []string{pmsplugin.ActEdit},
		DomainType:  pmsplugin.PrefixTable,
		DomainId:    strconv.Itoa(tableInfo.ID),
	}); err != nil {
		c.JSONE(1, err.Error(), nil)
		return
	}

	switch req.Status {
	case db.AlarmStatusOpen:
		err = service.Alarm.OpenOperator(id)
	case db.AlarmStatusClose:
		if errAlarmInfo != nil {
			c.JSONE(1, "alarm update failed 02"+errAlarmInfo.Error(), nil)
			return
		}
		op, errInstanceManager := service.InstanceManager.Load(instanceInfo.ID)
		if errInstanceManager != nil {
			c.JSONE(core.CodeErr, errInstanceManager.Error(), nil)
			return
		}
		if err = op.AlertViewDrop(alarmInfo.ViewTableName, tableInfo.Database.Cluster); err != nil {
			c.JSONE(1, "alarm update failed when delete metrics view"+err.Error(), nil)
			return
		}
		if err = service.Alarm.PrometheusRuleDelete(&instanceInfo, &alarmInfo); err != nil {
			c.JSONE(1, "alarm update failed 03: prometheus rule delete failed", nil)
			return
		}
		err = db.AlarmUpdate(invoker.Db, id, map[string]interface{}{"status": db.AlarmStatusClose})
	default:
		err = service.Alarm.Update(c.Uid(), id, req)
	}
	if err != nil {
		c.JSONE(1, "alarm update failed 04"+err.Error(), nil)
		return
	}
	event.Event.AlarmCMDB(c.User(), db.OpnAlarmsUpdate, map[string]interface{}{"req": req})
	c.JSONOK()
}

func List(c *core.Context) {
	req := &db.ReqPage{}
	if err := c.Bind(req); err != nil {
		c.JSONE(1, "invalid parameter", err)
		return
	}
	name := c.Query("name")
	tid, _ := strconv.Atoi(c.Query("tid"))
	did, _ := strconv.Atoi(c.Query("did"))
	status, _ := strconv.Atoi(c.Query("status"))
	query := egorm.Conds{}
	if name != "" {
		query["name"] = egorm.Cond{
			Op:  "like",
			Val: name,
		}
	}
	if status != 0 {
		query["status"] = status
	}
	if tid != 0 {
		query["tid"] = tid
	}
	if did != 0 {
		query["mogo_base_table.did"] = did
		total, list := db.AlarmListByDidPage(query, req)
		c.JSONPage(filterSensitiveInfo(list), core.Pagination{
			Current:  req.Current,
			PageSize: req.PageSize,
			Total:    total,
		})
		return
	}
	total, list := db.AlarmListPage(query, req)
	c.JSONPage(filterSensitiveInfo(list), core.Pagination{
		Current:  req.Current,
		PageSize: req.PageSize,
		Total:    total,
	})
	return
}

func filterSensitiveInfo(respList []*db.Alarm) []*db.Alarm {
	for k := range respList {
		respList[k].User.Password = "*"
	}
	return respList
}

func Info(c *core.Context) {
	id := cast.ToInt(c.Param("id"))
	if id == 0 {
		c.JSONE(1, "invalid parameter", nil)
		return
	}
	instanceInfo, tableInfo, alarmInfo, err := db.GetAlarmTableInstanceInfo(id)
	if err != nil {
		c.JSONE(core.CodeErr, err.Error(), nil)
		return
	}
	if err = permission.Manager.CheckNormalPermission(view.ReqPermission{
		UserId:      c.Uid(),
		ObjectType:  pmsplugin.PrefixInstance,
		ObjectIdx:   strconv.Itoa(tableInfo.Database.Iid),
		SubResource: pmsplugin.Alarm,
		Acts:        []string{pmsplugin.ActView},
		DomainType:  pmsplugin.PrefixTable,
		DomainId:    strconv.Itoa(tableInfo.ID),
	}); err != nil {
		c.JSONE(1, err.Error(), nil)
		return
	}
	conds := egorm.Conds{}
	conds["alarm_id"] = alarmInfo.ID
	filters, err := db.AlarmFilterList(conds)
	if err != nil {
		c.JSONE(core.CodeErr, err.Error(), nil)
		return
	}
	conditions, err := db.AlarmConditionList(conds)
	if err != nil {
		c.JSONE(core.CodeErr, err.Error(), nil)
		return
	}
	user, _ := db.UserInfo(alarmInfo.Uid)

	instanceInfo.Dsn = "*"
	user.Password = "*"

	res := view.RespAlarmInfo{
		Alarm:      alarmInfo,
		Filters:    filters,
		Conditions: conditions,
		User:       user,
		Ctime:      alarmInfo.Ctime,
		Utime:      alarmInfo.Utime,
		Instance:   instanceInfo,
		Table:      tableInfo,
	}
	c.JSONE(core.CodeOK, "succ", res)
	return
}

func Delete(c *core.Context) {
	id := cast.ToInt(c.Param("id"))
	if id == 0 {
		c.JSONE(1, "invalid parameter", nil)
		return
	}
	instanceInfo, tableInfo, alarmInfo, err := db.GetAlarmTableInstanceInfo(id)
	if err != nil {
		c.JSONE(1, "alarm failed to delete 01: "+err.Error(), nil)
		return
	}
	if err = permission.Manager.CheckNormalPermission(view.ReqPermission{
		UserId:      c.Uid(),
		ObjectType:  pmsplugin.PrefixInstance,
		ObjectIdx:   strconv.Itoa(tableInfo.Database.Iid),
		SubResource: pmsplugin.Alarm,
		Acts:        []string{pmsplugin.ActDelete},
		DomainType:  pmsplugin.PrefixTable,
		DomainId:    strconv.Itoa(tableInfo.ID),
	}); err != nil {
		c.JSONE(1, err.Error(), nil)
		return
	}
	tx := invoker.Db.Begin()
	if err = db.AlarmDelete(tx, id); err != nil {
		c.JSONE(1, "alarm failed to delete 02: "+err.Error(), nil)
		return
	}
	// filter
	if err = db.AlarmFilterDeleteBatch(tx, id); err != nil {
		tx.Rollback()
		c.JSONE(1, "alarm failed to delete 03: "+err.Error(), nil)
		return
	}
	// condition
	if err = db.AlarmConditionDeleteBatch(tx, id); err != nil {
		tx.Rollback()
		c.JSONE(1, "alarm failed to delete 04: "+err.Error(), nil)
		return
	}
	if err = service.Alarm.PrometheusRuleDelete(&instanceInfo, &alarmInfo); err != nil {
		tx.Rollback()
		c.JSONE(1, "alarm failed to delete 05: "+err.Error(), nil)
		return
	}
	op, err := service.InstanceManager.Load(tableInfo.Database.Iid)
	if err != nil {
		c.JSONE(core.CodeErr, err.Error(), nil)
		return
	}
	if err = op.AlertViewDrop(alarmInfo.ViewTableName, tableInfo.Database.Cluster); err != nil {
		tx.Rollback()
		c.JSONE(1, "alarm failed to delete 06: "+err.Error(), nil)
		return
	}
	if err = tx.Commit().Error; err != nil {
		tx.Rollback()
		c.JSONE(1, "alarm failed to delete 07"+err.Error(), nil)
		return
	}
	event.Event.AlarmCMDB(c.User(), db.OpnAlarmsDelete, map[string]interface{}{"alarmInfo": alarmInfo})
	c.JSONOK()
}

func HistoryList(c *core.Context) {
	var req view.ReqAlarmHistoryList
	if err := c.Bind(&req); err != nil {
		c.JSONE(1, "invalid parameter: "+err.Error(), nil)
		return
	}
	invoker.Logger.Debug("history", elog.Any("req", req))
	conds := egorm.Conds{}
	if req.AlarmId != 0 {
		conds["alarm_id"] = req.AlarmId
	}
	if req.StartTime != 0 {
		conds["ctime"] = egorm.Cond{Op: ">", Val: req.StartTime}
	}
	if req.EndTime != 0 {
		conds["ctime"] = egorm.Cond{Op: "<", Val: req.EndTime}
	}
	total, list := db.AlarmHistoryPage(conds, &db.ReqPage{
		Current:  req.Current,
		PageSize: req.PageSize,
	})
	conds["is_pushed"] = 1
	succ, _ := db.AlarmHistoryPage(conds, &db.ReqPage{
		Current:  req.Current,
		PageSize: req.PageSize,
	})
	c.JSONPage(view.RespAlarmHistoryList{
		Total: total,
		Succ:  succ,
		List:  list,
	}, core.Pagination{
		Current:  req.Current,
		PageSize: req.PageSize,
		Total:    total,
	})
	return
}

func HistoryInfo(c *core.Context) {
	id := cast.ToInt(c.Param("id"))
	if id == 0 {
		c.JSONE(1, "invalid parameter", nil)
		return
	}
	res, err := db.AlarmHistoryInfo(invoker.Db, id)
	if err != nil {
		c.JSONE(core.CodeErr, err.Error(), nil)
		return
	}
	c.JSONE(core.CodeOK, "succ", res)
	return
}
