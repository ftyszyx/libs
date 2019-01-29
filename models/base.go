package models

import (
	"fmt"
	"runtime/debug"
	"strconv"

	"github.com/pkg/errors"

	"github.com/ftyszyx/libs/beego/cache"
	"github.com/ftyszyx/libs/beego/logs"
	"github.com/ftyszyx/libs/db/mysql"
	"github.com/zyx/shop_server/utils"
)

type Model struct {
	tablename string
	cache     cache.Cache
}

func NewModel(tablename string, havecache bool) Model {
	o := new(Model)
	o.tablename = tablename
	if havecache {
		o.cache, _ = cache.NewCache("memory", `{"interval":0}`) //不过期
	}

	return *o
}

type ModelInterface interface {
	Init()
	TableName() string //表名
	InitSqlField(mysql.SqlType) mysql.SqlType
	InitJoinString(mysql.SqlType, bool) mysql.SqlType
	InitField(mysql.SqlType) mysql.SqlType

	Cache() cache.Cache
	ClearCache()
	ClearRowCache(string)
	GetModelStruct() interface{}

	GetFieldName(string) string
	ExportNameProcess(string, interface{}, mysql.Params) (string, error)
	GetInfoAndCache(mysql.DBOperIO, string, bool) mysql.Params
	CheckExit(mysql.DBOperIO, string, interface{}) bool
	GetInfoById(mysql.DBOperIO, interface{}) mysql.Params
	AllExcCommon(mysql.DBOperIO, ModelInterface, AllReqData, int) (error, int, []mysql.Params)
	GetInfoByField(mysql.DBOperIO, string, interface{}) []mysql.Params
	GetNumByField(mysql.DBOperIO, map[string]interface{}) int
	GetInfoByWhere(mysql.DBOperIO, string) ([]mysql.Params, error)
}

//获取所有时请求数据
type AllReqData struct {
	Page   int
	Rownum int
	Order  map[string]interface{}
	And    bool
	Search map[string]interface{}
}

func (self *Model) InitJoinString(sql mysql.SqlType, allfield bool) mysql.SqlType {
	return sql
}
func (self *Model) InitField(sql mysql.SqlType) mysql.SqlType {
	return sql
}

func (self *Model) ExportNameProcess(name string, value interface{}, row mysql.Params) (string, error) {
	if value == nil {
		logs.Info("field %s is nil", name)
		return "", nil
	}
	datastr, ok := value.(string)
	if ok == false {
		return "", errors.New("upload file err:" + name + " not exit")
	}
	return datastr, nil

}

func (self *Model) GetFieldName(name string) string {
	return name
}

func (self *Model) ClearCache() {
	if self.cache != nil {
		self.cache.ClearAll()
	}
}

func (self *Model) Cache() cache.Cache {
	return self.cache
}

func (self *Model) TableName() string {
	return self.tablename
}

func (self *Model) InitSqlField(sql mysql.SqlType) mysql.SqlType {
	return sql
}

func (self *Model) Init() {
	logs.Info("init:%s", self.tablename)
}

func (self *Model) GetModelStruct() interface{} {
	return nil
}

//检查是否存在某个数据
func (self *Model) CheckExitMap(oper mysql.DBOperIO, fieldinfo map[string]interface{}) bool {
	// db := orm.NewOrm()
	var dataList []mysql.Params
	var sqltext mysql.SqlType
	sqltext = &mysql.SqlBuild{}
	sqltext = sqltext.Name(self.TableName())
	num, err := oper.Raw(sqltext.Where(fieldinfo).Find()).Values(&dataList)
	if err == nil && num > 0 {
		return true
	}
	return false
}

//检查是否存在
func (self *Model) CheckExit(oper mysql.DBOperIO, field string, value interface{}) bool {
	data := make(map[string]interface{})
	data[field] = value
	return self.CheckExitMap(oper, data)
}

//获取表里面的一项，默认从内存取，如果内存没有，就从数据库取，并缓存。
func (self *Model) GetInfoAndCache(oper mysql.DBOperIO, uid string, forceUpdate bool) mysql.Params {
	if forceUpdate == false {
		//读旧的
		if self.cache != nil {
			datatemp := self.cache.Get(uid)
			if datatemp != nil {
				info, ok := datatemp.(map[string]interface{})
				if ok {
					// logs.Info("old info")
					return info
				}
			}
		}

	}
	// logs.Info("find info")
	// o := orm.NewOrm()
	var dataList []mysql.Params
	num, err := oper.Raw(fmt.Sprintf(`select * from %s where id=?`, self.TableName()), uid).Values(&dataList)
	if err == nil && num > 0 {
		if self.cache != nil {
			self.cache.Put(uid, dataList[0], 0)
		} else {
			logs.Error("tablename:%s no cache", self.tablename)
		}
		// logs.Info("add info")
		return dataList[0]
	}
	return nil
}

func (self *Model) GetInfoById(oper mysql.DBOperIO, id interface{}) mysql.Params {
	res := self.GetInfoByField(oper, "id", id)
	if res != nil {
		return res[0]
	}
	return nil
}

func (self *Model) GetInfoByField(oper mysql.DBOperIO, field string, value interface{}) []mysql.Params {
	// o := orm.NewOrm()
	var dataList []mysql.Params
	num, err := oper.Raw(fmt.Sprintf("select * from %s where `%s`=?", self.TableName(), field), value).Values(&dataList)
	if err == nil && num > 0 {
		return dataList
	}
	if err != nil {
		logs.Error("err:%s", err.Error())
	}

	return nil
}

//获取数量
func (self *Model) GetNumByField(oper mysql.DBOperIO, search map[string]interface{}) int {
	// o := orm.NewOrm()
	totalnum := 0
	var dataList []mysql.Params
	var sqltext mysql.SqlType
	sqltext = &mysql.SqlBuild{}
	sqltext = sqltext.Name(self.TableName())
	num, err := oper.Raw(sqltext.Where(search).Count()).Values(&dataList)
	if err == nil && num > 0 {
		totalnum, err = strconv.Atoi(dataList[0][mysql.SQLTotalName].(string))
		if err == nil {
			return totalnum
		}
	}
	if err != nil {
		logs.Error("err:%s", err.Error())
	}

	return 0
}

func (self *Model) GetInfoByWhere(oper mysql.DBOperIO, where string) ([]mysql.Params, error) {
	// o := orm.NewOrm()
	var dataList []mysql.Params
	num, err := oper.Raw(fmt.Sprintf("select * from %s where %s", self.TableName(), where)).Values(&dataList)
	if err == nil && num > 0 {
		return dataList, nil
	}
	if err != nil {

		return nil, errors.WithStack(err)
	}

	return nil, nil
}

//清除缓存
func (self *Model) ClearRowCache(id string) {
	if self.cache != nil {
		self.cache.Delete(id)
	}
}

func (self *Model) AllExcCommon(oper mysql.DBOperIO, model ModelInterface, data AllReqData, gettype int) (error, int, []mysql.Params) {

	var totalnum = 0
	var dataList []mysql.Params
	var sqltext mysql.SqlType
	sqltext = &mysql.SqlBuild{}
	sqltext = sqltext.Name(model.TableName())

	if data.And {
		sqltext = sqltext.Where(data.Search)
	} else {
		sqltext = sqltext.WhereOr(data.Search)
	}
	sqltext = model.InitJoinString(sqltext, false)
	num, err := oper.Raw(sqltext.Count()).Values(&dataList)
	if err == nil && num > 0 {
		totalnum, err = strconv.Atoi(dataList[0][mysql.SQLTotalName].(string))
		if err != nil {
			logs.Error("err:%+v statck:\n %s", err, string(debug.Stack()))
			return errors.WithStack(err), 0, nil
		}
		if gettype == utils.GetAll_type_num {
			return nil, totalnum, nil
		}
		sqltext = sqltext.Order(data.Order)
		sqltext = model.InitJoinString(sqltext, false)
		if data.Page == 0 {
			//不用分页
			sqltext = model.InitJoinString(model.InitField(sqltext), true)

			num, err = oper.Raw(sqltext.Select()).Values(&dataList)
			if err == nil {
				return nil, totalnum, dataList
			} else {
				//logs.Error("err:%s statck:\n %s", err.Error(), string(debug.Stack()))
				return errors.WithStack(err), 0, nil
			}
		} else {
			//用分页
			var start = (data.Page - 1) * data.Rownum
			if totalnum > 1000 {
				//总数很多
				tablealias := sqltext.GetAlias()
				selfidname := "id"
				if tablealias != "" {
					selfidname = tablealias + ".id"
				}
				subsql := sqltext.Limit([]int{start, data.Rownum}).Field(map[string]string{selfidname: "id"}).Select()
				var newsqltext mysql.SqlType
				newsqltext = &mysql.SqlBuild{}
				newsqltext = newsqltext.Name(model.TableName()).Order(data.Order)
				// newsqltext = newsqltext.Name(model.TableName())
				newsqltext = model.InitJoinString(model.InitField(newsqltext), true)
				oldjoinstr := newsqltext.GetJoinStr()
				newsqltext.Join(oldjoinstr + fmt.Sprintf(" INNER join (%s) a ON `a`.`id`=%s ", subsql, mysql.SqlGetKey(selfidname)))
				num, err = oper.Raw(newsqltext.Select()).Values(&dataList)
				if err == nil {
					return nil, totalnum, dataList
				} else {
					//logs.Error("err:%s statck:\n %s", err.Error(), string(debug.Stack()))
					return errors.WithStack(err), 0, nil
				}

			} else {
				//按正常方式
				sqltext = model.InitJoinString(model.InitField(sqltext), true)
				num, err = oper.Raw(sqltext.Limit([]int{start, data.Rownum}).Select()).Values(&dataList)
				if err == nil {
					return nil, totalnum, dataList
				} else {
					//logs.Error("err:%s statck:\n %s", err.Error(), string(debug.Stack()))

					return errors.WithStack(err), 0, nil
				}
			}

		}
	} else {
		return errors.WithStack(err), 0, nil

	}
}