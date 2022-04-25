/**
 * @Author: alessonhu
 * @Description:
 * @File:  mdm_event_info.go
 * @Version: 1.0.0
 * @Date: 2021/1/13 10:59
 */
package models

import (
	. "mdm/common/qm"
	"time"
)

type MdmEventInfo struct {
	Id            int       `xorm:"not null pk autoincr unique INTEGER"`
	Mid           string    `xorm:"VARCHAR(64)"`
	Udid          string    `xorm:"VARCHAR(64)"`
	RequestType   string    `xorm:"VARCHAR(64)"`
	Uuid          string    `xorm:"VARCHAR(64)"`
	Payloadbase64 string    `xorm:"TEXT"`
	Payloaduuid   string    `xorm:"VARCHAR(64)"`
	State         int       `xorm:"INTEGER"`
	Itime         time.Time `xorm:"created"`
	Utime         time.Time `xorm:"updated"`
}

func InsertMdmEventinfo(eventInfo *MdmEventInfo) (err error) {
	_, err = Db.Insert(eventInfo)
	if err != nil {
		LOG_ERROR(err)
		return
	}
	return
}

func UpdateMdmEventState(eventInfo *MdmEventInfo) (err error) {
	_, err = Db.Exec("update mdm_event_info set state=? where uuid=?", eventInfo.State, eventInfo.Uuid)
	if err != nil {
		LOG_ERROR(err)
		return
	}
	return
}
