/**
 * @Author: alessonhu
 * @Description:
 * @File:  checkin.go
 * @Version: 1.0.0
 * @Date: 2020/12/23 14:14
 */
package main

import (
	"encoding/hex"
	"fmt"
	"github.com/gin-gonic/gin"
	"mdm/common/models"
	"mdm/common/qm"
	"net/http"
)

const (
	//checkin state
	CheckStateIn  = 1
	CheckStateOut = -1

	//redis key
	//设备系统信息key mdm_client_mid
	MdmDeviceInfoKey = "mdm_clientinfo_"
	//描述文件未执行队列
	MdmReadyProfileKey = "mdm_ready_proinfo_"
	//描述文件待执行队列
	MdmRunProfileKey = "mdm_run_proinfo_"
	//checkin value:k-v
	CheckKeyUdid      = "udid"
	CheckKeyToken     = "token"
	CheckKeyPushmagic = "pushmagic"
	CheckKeyData      = "data"
	CheckKeyState     = "state"
)

// CheckinRequest represents an MDM checkin command struct.
type CheckinCommand struct {
	// MessageType can be either Authenticate,
	// TokenUpdate or CheckOut
	Mid          string
	MessageType  string
	Topic        string
	UDID         string
	EnrollmentID string
	auth
	update
}

// Authenticate Message Type
type auth struct {
	OSVersion    string
	BuildVersion string
	ProductName  string
	SerialNumber string
	IMEI         string
	MEID         string
	DeviceName   string `plist:"DeviceName,omitempty"`
	Challenge    []byte `plist:"Challenge,omitempty"`
	Model        string `plist:"Model,omitempty"`
	ModelName    string `plist:"ModelName,omitempty"`
}

// TokenUpdate Message Type
type update struct {
	Token                 hexData
	PushMagic             string
	UnlockToken           hexData
	AwaitingConfiguration bool
	userTokenUpdate
}

// data decodes to []byte,
// we can then attach a string method to the type
// Tokens are encoded as Hex Strings
type hexData []byte

func (d hexData) String() string {
	return hex.EncodeToString(d)
}

// TokenUpdate with user keys
type userTokenUpdate struct {
	UserID        string `plist:",omitempty"`
	UserLongName  string `plist:",omitempty"`
	UserShortName string `plist:",omitempty"`
	NotOnConsole  bool   `plist:",omitempty"`
}

func (svc *mdmService) MdmCheckin(c *gin.Context) {
	var cmd CheckinCommand
	body, err := mdmRequestBody(c.Request, &cmd)
	if err != nil {
		qm.LOG_ERROR(err)
		c.String(http.StatusBadRequest, "request err")
		return
	}

	cmd.Mid = c.Query("mid")
	if cmd.Mid == "" {
		c.String(http.StatusBadRequest, "request err")
		return
	}

	qm.LOG_DEBUG_F(string(body))

	var mdmDeviceInfo models.MdmDeviceInfo
	redisData := make(map[string]interface{})
	mdmDeviceInfo.Mid = cmd.Mid
	switch cmd.MessageType {
	case "Authenticate":
		mdmDeviceInfo.OsVersion = cmd.OSVersion
		mdmDeviceInfo.BuildVersion = cmd.BuildVersion
		mdmDeviceInfo.SerialNum = cmd.SerialNumber
		mdmDeviceInfo.Imei = cmd.IMEI
		mdmDeviceInfo.DeviceName = cmd.DeviceName
		mdmDeviceInfo.Model = cmd.Model
		mdmDeviceInfo.ModelName = cmd.ModelName
		mdmDeviceInfo.ProductName = cmd.ProductName
		mdmDeviceInfo.Udid = cmd.UDID
	case "TokenUpdate":
		mdmDeviceInfo.State = CheckStateIn
		mdmDeviceInfo.Token = cmd.Token.String()
		mdmDeviceInfo.Pushmagic = cmd.PushMagic
		mdmDeviceInfo.Udid = cmd.UDID

		redisData[CheckKeyToken] = cmd.Token.String()
		redisData[CheckKeyPushmagic] = cmd.PushMagic
	case "CheckOut":
		mdmDeviceInfo.State = CheckStateOut
		mdmDeviceInfo.Udid = cmd.UDID
		//清空命令队列
		err = g_app.GRedis.Del(MdmReadyProfileKey + cmd.UDID).Err()
		if err != nil {
			qm.LOG_ERROR(err)
		}
		err = g_app.GRedis.Del(MdmRunProfileKey + cmd.UDID).Err()
		if err != nil {
			qm.LOG_ERROR(err)
		}
	}

	err = models.SaveMdmDeviceinfo(&mdmDeviceInfo)
	if err != nil {
		c.String(http.StatusInternalServerError, "")
	}

	redisData[CheckKeyUdid] = cmd.UDID
	redisData[CheckKeyData] = string(body)
	redisData[CheckKeyState] = mdmDeviceInfo.State

	err = g_app.GRedis.HMSet(MdmDeviceInfoKey+cmd.Mid, redisData).Err()
	if err != nil {
		qm.LOG_ERROR_F("put redis err, %s", err.Error())
	}

	c.String(http.StatusOK, string(rspMdmPlist()))
	return
}

func (svc *mdmService) MdmWake(c *gin.Context) {
	var token, magic string
	var ok bool
	mid := c.Request.Header.Get("Mid")
	res, err := g_app.GRedis.HMGet(MdmDeviceInfoKey+mid, "udid", "token", "pushmagic").Result()
	if err != nil {
		qm.LOG_ERROR(err)
	}
	qm.LOG_INFO_F("%+v", res)

	if token, ok = res[1].(string); !ok {
		qm.LOG_ERROR_F("[ apns wake ] token is err")
		c.String(http.StatusInternalServerError, "request err")
		return
	}
	if magic, ok = res[2].(string); !ok {
		qm.LOG_ERROR_F("[ apns wake ] magic is err")
		c.String(http.StatusInternalServerError, "request err")
		return
	}
	g_app.ApnsSvc.Push(token, magic)

	c.String(http.StatusOK, fmt.Sprintf("notify device succ, mid is %s, udid is %s", mid, res[0].(string)))
	return
}
