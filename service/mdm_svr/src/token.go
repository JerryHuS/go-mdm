/**
 * @Author: alessonhu
 * @Description:
 * @File:  token.go
 * @Version: 1.0.0
 * @Date: 2021/1/11 15:16
 */
package main

import (
	"fmt"
	jwt "github.com/dgrijalva/jwt-go/v4"
	"github.com/gin-gonic/gin"
	"mdm/common/qm"
	"net/http"
	"time"
)

const (
	SuccCode = 0
	ErrCode  = -1
	SvrCode  = -2
)

type MdmApiRsp struct {
	RetCode    int         `json:"ret"`
	ErrorCode  int         `json:"errorcode"`
	Msg        string      `json:"msg"`
	StackTrace string      `json:"stacktrace"`
	Data       interface{} `json:"data,omitempty"`
}

func MakeMdmApiRsp(c *gin.Context, code int, stacktrace string, data interface{}) {
	var rsp MdmApiRsp
	rsp.ErrorCode = code
	if code != 0 {
		rsp.RetCode = -1
		rsp.Msg = "request err"
	}
	rsp.StackTrace = stacktrace
	rsp.Data = data
	c.JSON(http.StatusOK, rsp)
}

type TokenInfo struct {
	Token      string `json:"token"`
	Expiretime string `json:"expiretime"`
	UseTimes   int    `json:"usetimes"`
}

func (svc *mdmService) MdmToken(c *gin.Context) {
	if !CheckAppSign(c.Request) {
		qm.LOG_ERROR("check sign err")
		MakeMdmApiRsp(c, ErrCode, "SSignature err", nil)
		return
	}
	mid := c.Request.Header.Get("SAppId")
	duration, err := time.ParseDuration(fmt.Sprintf("%ds", g_app.MdmSvc.TokenExpireSecond))
	if err != nil {
		qm.LOG_ERROR_F("duration err,%s", err.Error())
		MakeMdmApiRsp(c, ErrCode, "Expire err", nil)
	}
	expiredtime := time.Now().Add(duration).Format("2006-01-02 15:04:05")

	claims := jwt.MapClaims{
		ClaimsMid:    mid,
		ClaimsExpire: expiredtime,
		ClaimsTimes:  1,
	}
	token, err := g_app.JwtSignTool.GetSign(claims)
	if err != nil {
		qm.LOG_ERROR(err)
		MakeMdmApiRsp(c, ErrCode, err.Error(), nil)
		return
	}
	tokenInfo := &TokenInfo{
		Token:      token,
		Expiretime: expiredtime,
		UseTimes:   1,
	}
	MakeMdmApiRsp(c, SuccCode, "", tokenInfo)
	return
}
