/**
 * @Author: alessonhu
 * @Description:
 * @File:  connect.go
 * @Version: 1.0.0
 * @Date: 2021/1/4 15:51
 */
package main

import (
	"github.com/gin-gonic/gin"
	"mdm/common/models"
	"mdm/common/qm"
	"net/http"
)

const (
	//checkin类型
	MessageTypeAuthenticate = "Authenticate"
	MessageTypeTokenUpdate  = "TokenUpdate"
	MessageTypeCheckOut     = "CheckOut"

	//connect状态
	StatusIdle         = "Idle"
	StatusAcknowledged = "Acknowledged"
	StatusNotNow       = "NotNow"
)

const installApp = `<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple Computer//DTD PLIST 1.0//EN""http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0"><dict>
    <key>Command</key>
    <dict>
        <key>RequestType</key>
        <string>InstalledApplicationList</string>
    </dict>
    <key>CommandUUID</key>
    <string>149e4fd2-0267-4da2-9b58-bf94282dcdb4</string>
</dict>
</plist>`

type ConnectCommand struct {
	// MessageType can be either Idle, Acknowledged, NotNow or Error
	Status        string
	CommandUUID   string
	UDID          string
	MessageResult string
}

func (svc *mdmService) MdmConnect(c *gin.Context) {
	var cmd ConnectCommand
	body, err := mdmRequestBody(c.Request, &cmd)
	if err != nil {
		qm.LOG_ERROR(err)
		c.String(http.StatusOK, string(rspMdmPlist()))
		return
	}

	qm.LOG_DEBUG_F(string(body))

	switch cmd.Status {
	case StatusIdle:
		command, err := g_app.GRedis.LPop(MdmReadyProfileKey + cmd.UDID).Result()
		if err != nil {
			qm.LOG_ERROR(err)
			c.String(http.StatusOK, string(rspMdmPlist()))
			return
		}
		qm.LOG_DEBUG_F("mdm connect rsp info is \n %s", command)
		err = g_app.GRedis.RPush(MdmRunProfileKey+cmd.UDID, command).Err()
		if err != nil {
			qm.LOG_ERROR(err)
		}
		c.String(http.StatusOK, command)
	case StatusAcknowledged:
		//更新上个uuid执行结果
		go models.UpdateMdmEventState(&models.MdmEventInfo{
			Uuid:  cmd.CommandUUID,
			State: 1,
		})
		err = g_app.GRedis.LPop(MdmRunProfileKey + cmd.UDID).Err()
		if err != nil {
			qm.LOG_ERROR(err)
			c.String(http.StatusOK, string(rspMdmPlist()))
			return
		}
		command, err := g_app.GRedis.LPop(MdmReadyProfileKey + cmd.UDID).Result()
		if err != nil {
			qm.LOG_ERROR(err)
			c.String(http.StatusOK, string(rspMdmPlist()))
			return
		}
		qm.LOG_DEBUG_F("mdm connect rsp info is \n %s", command)
		if command != "" {
			c.String(http.StatusOK, command)
		} else {
			c.String(http.StatusOK, string(rspMdmPlist()))
		}
	case StatusNotNow:
		err = g_app.GRedis.LPush(g_app.GRedis.LPop(MdmRunProfileKey + cmd.UDID).String()).Err()
		if err != nil {
			qm.LOG_ERROR(err)
		}
		c.String(http.StatusOK, string(rspMdmPlist()))
	}
	return
}
