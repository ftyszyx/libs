package control

import (
	"fmt"

	"github.com/pkg/errors"

	"github.com/ftyszyx/libs"
	"github.com/ftyszyx/libs/db/mysql"
)

type SqlIO interface {
	BeforeSql(map[string]interface{}) error
	AfterSql(map[string]interface{}, mysql.Params) error
	AddOneRow(int, []string) string //愤青时输出一行
}

func (self *BaseController) BeforeSql(data map[string]interface{}) error {
	return nil
}

func (self *BaseController) AfterSql(data map[string]interface{}, oldinfo mysql.Params) error {
	return nil
}

func (self *BaseController) AddCommonAndReturn(sqlcall SqlIO) {
	err := self.AddCommon(sqlcall)
	if err != nil {
		self.AjaxReturnError(errors.WithStack(err))
	}
	self.AjaxReturnSuccessNull()
}

func (self *BaseController) AddCommon(sqlcall SqlIO) error {
	datacheck := self.model.GetModelStruct()
	err := self.CheckExit(datacheck, self.GetPost(), false)
	if err != nil {
		return err
	}
	adddata := libs.ClearMapByStruct(self.GetPost(), datacheck)
	return self.AddCommonExe(sqlcall, adddata)
}

func (self *BaseController) AddCommonExe(sqlcall SqlIO, adddata map[string]interface{}) error {
	return self.AddCommonTable(sqlcall, adddata, self.model.TableName())
}

func (self *BaseController) AddCommonTable(sqlcall SqlIO, adddata map[string]interface{}, table string) error {
	// o := orm.NewOrm()
	err := sqlcall.BeforeSql(adddata)
	if err != nil {
		return err
	}
	keys, values := mysql.SqlGetInsertInfo(adddata)
	_, err = self.dboper.Raw(fmt.Sprintf("insert into %s (%s) values (%s)", table, keys, values)).Exec()
	if err != nil {
		return err
	}
	err = sqlcall.AfterSql(adddata, nil)
	if err != nil {
		return err
	}
	return nil
}

func (self *BaseController) EditCommonAndReturn(sqlcall SqlIO) {
	err := self.EditCommon(sqlcall)
	if err != nil {
		self.AjaxReturnError(errors.WithStack(err))
	}

	self.AjaxReturnSuccessNull()
}

func (self *BaseController) EditCommon(sqlcall SqlIO) error {
	return self.EditCommonByKey(sqlcall, "id")
}

func (self *BaseController) EditCommonByKey(sqlcall SqlIO, key string) error {
	modelcheck := self.model.GetModelStruct()
	if self.CheckFieldExit(self.GetPost(), key) == false {
		return errors.New("id为空")
	}
	id := self.GetPost()[key].(string)
	err := self.CheckExit(modelcheck, self.GetPost(), true)
	if err != nil {
		return err
	}
	changedata := libs.ClearMapByStruct(self.GetPost(), modelcheck)
	delete(changedata, key)
	if len(changedata) == 0 {
		return errors.New("没有修改")
	}
	return self.UpdateSqlCommon(sqlcall, changedata, key, id)
}

func (self *BaseController) UpdateSqlByIdAndReturn(sqlcall SqlIO, changedata map[string]interface{}, id interface{}) {
	err := self.UpdateSqlCommon(sqlcall, changedata, "id", id)
	if err != nil {
		self.AjaxReturnError(errors.WithStack(err))
	}

	self.AjaxReturnSuccessNull()
}

func (self *BaseController) UpdateSqlById(sqlcall SqlIO, changedata map[string]interface{}, id interface{}) error {
	return self.UpdateSqlCommon(sqlcall, changedata, "id", id)
}

func (self *BaseController) UpdateSqlCommon(sqlcall SqlIO, changedata map[string]interface{}, field string, id interface{}) error {
	err := sqlcall.BeforeSql(changedata)
	if err != nil {
		return nil
	}
	oldinfo := self.model.GetInfoByField(self.dboper, field, id)
	if oldinfo == nil {
		return errors.New("没找到")
	}
	_, err = self.dboper.Raw(fmt.Sprintf("update %s set %s where `%s`=?", self.model.TableName(), mysql.SqlGetKeyValue(changedata, "="), field), id).Exec()
	if err != nil {
		return err
	}
	err = sqlcall.AfterSql(changedata, oldinfo[0])
	if err != nil {
		return err
	}
	return nil
}

func (self *BaseController) DelCommonAndReturn(sqlcall SqlIO) {
	err := self.DelCommon(sqlcall)
	if err != nil {
		self.AjaxReturnError(errors.WithStack(err))
	}
	self.AjaxReturnSuccessNull()
}

func (self *BaseController) DelCommon(sqlcall SqlIO) error {
	self.CheckFieldExitAndReturn(self.GetPost(), "id", "id为空")
	id := self.GetPost()["id"].(string)
	oldinfo := self.model.GetInfoById(self.dboper, id)
	if oldinfo == nil {
		return errors.New("id 不存在")
	}
	err := sqlcall.BeforeSql(oldinfo)
	if err != nil {
		return err
	}
	_, err = self.dboper.Raw(fmt.Sprintf("delete from %s where `id`=?", self.model.TableName()), id).Exec()
	if err != nil {
		return err
	}
	err = sqlcall.AfterSql(oldinfo, nil)
	if err != nil {
		return err
	}
	return nil
}
