/**
 * @Author: alessonhu
 * @Description:
 * @File:  mdm_deviceinfo.go
 * @Version: 1.0.0
 * @Date: 2021/1/11 16:09
 */
package models

import (
	"time"
)

type MdmDeviceInfo struct {
	Id           int       `xorm:"not null pk autoincr unique INTEGER"`
	Mid          string    `xorm:"VARCHAR(64)"`
	Udid         string    `xorm:"VARCHAR(64)"`
	Pushmagic    string    `xorm:"VARCHAR(64)"`
	Token        string    `xorm:"VARCHAR(64)"`
	DeviceName   string    `xorm:"VARCHAR(64)"`
	BuildVersion string    `xorm:"VARCHAR(64)"`
	Model        string    `xorm:"VARCHAR(64)"`
	ModelName    string    `xorm:"VARCHAR(64)"`
	OsVersion    string    `xorm:"VARCHAR(64)"`
	ProductName  string    `xorm:"VARCHAR(64)"`
	SerialNum    string    `xorm:"VARCHAR(64)"`
	Topic        string    `xorm:"VARCHAR(64)"`
	Imei         string    `xorm:"VARCHAR(64)"`
	Meid         string    `xorm:"VARCHAR(64)"`
	State        int       `xorm:"INTEGER"`
	Itime        time.Time `xorm:"created"`
	Utime        time.Time `xorm:"updated"`
}

func SaveMdmDeviceinfo(deviceinfo *MdmDeviceInfo) (err error) {
	entity := &MdmDeviceInfo{}
	has, err := Db.Where("udid=?", deviceinfo.Udid).Get(entity)
	if err != nil {
		return
	}
	if has {
		deviceinfo.Id = entity.Id
		_, err = Db.Id(deviceinfo.Id).Update(deviceinfo)
		if err != nil {
			return
		}
	} else {
		_, err = Db.Insert(deviceinfo)
		if err != nil {
			return
		}
	}
	return
}
