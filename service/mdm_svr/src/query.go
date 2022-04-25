/**
 * @Author: alessonhu
 * @Description:
 * @File:  query.go
 * @Version: 1.0.0
 * @Date: 2021/3/1 15:30
 */
package main

import (
	"encoding/json"
	"github.com/gin-gonic/gin"
	"io/ioutil"
	"mdm/common/qm"
	"net/http"
)

type DeviceState struct {
	DeviceState int `json:"device_state"`
}

func (svc *mdmService) QueryDeviceState(c *gin.Context) {
	body, _ := ioutil.ReadAll(c.Request.Body)
	qm.LOG_DEBUG_F("mdmEvent request body = %v", string(body))
	infobyte, err := qm.AesCBCDecryptIv(string(body), []byte("3U7f4yBM4&LiJp9d"), []byte("jzThmvU0yO*&^8GW"))
	if err != nil {
		MakeMdmApiRsp(c, ErrCode, err.Error(), nil)
		return
	}

	unzipdata := qm.DoZlibUnCompress(infobyte)
	req := &EventReq{}

	// 响应信息
	err = json.Unmarshal(unzipdata, req)
	if err != nil {
		MakeMdmApiRsp(c, ErrCode, err.Error(), nil)
		return
	}

	qm.LOG_DEBUG_F("req is %+v", req)

	token := c.Request.Header.Get("Token")
	tokenValid := CheckMdmToken(token, req.Mid)
	if !tokenValid {
		qm.LOG_ERROR_F("token is err, token is %s, mid is %s", token, req.Mid)
		c.String(http.StatusOK, "token is error")
		return
	}

	res, err := g_app.GRedis.HGet(MdmDeviceInfoKey+req.Mid, CheckKeyState).Result()
	if err != nil {
		qm.LOG_ERROR(err)
	}
	deviceState := &DeviceState{DeviceState: 0}
	if res == "1" {
		deviceState.DeviceState = 1
	}
	qm.LOG_DEBUG_F("rsp is %+v", deviceState)
	MakeMdmApiRsp(c, 0, "", deviceState)
	return
}
