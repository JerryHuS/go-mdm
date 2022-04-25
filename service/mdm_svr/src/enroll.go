/**
 * @Author: alessonhu
 * @Description:
 * @File:  enroll.go
 * @Version: 1.0.0
 * @Date: 2020/12/28 20:33
 */
package main

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"mdm/common/qm"
	"net/http"
)

func (svc *mdmService) MdmEnroll(c *gin.Context) {
	mid := c.Request.Header.Get("Mid")
	/*
		token := c.Request.Header.Get("Token")
		tokenValid := CheckMdmToken(token, mid)
		if !tokenValid {
			qm.LOG_ERROR_F("token is err, token is %s, mid is %s", token, mid)
			c.String(http.StatusOK, "token is error")
			return
		}
	*/

	if len(svc.TLSCert) == 0 || svc.SVRURL == "" || svc.SVRURL == fmt.Sprintf("https://:%s", g_app.Server.MdmExposePort) || g_app.ApnsSvc.PushCertTopic == "" {
		qm.LOG_ERROR_F("tls cert or domain not ready")
		qm.LOG_ERROR_F("len(svc.TLSCert) is %d, svc.SVRURL is %s, g_app.ApnsSvc.PushCertTopic is %s", len(svc.TLSCert), svc.SVRURL, g_app.ApnsSvc.PushCertTopic)
		c.String(http.StatusInternalServerError, "server not ready")
		return
	}

	profile, err := svc.MakeEnrollmentProfile(mid)
	if err != nil {
		qm.LOG_ERROR(err)
		c.String(http.StatusOK, "server err")
		return
	}

	config, err := profileOrPayloadToMobileconfig(profile)
	if err != nil {
		qm.LOG_ERROR(err)
		c.String(http.StatusOK, "server err")
		return
	}

	qm.LOG_DEBUG_F(string(config))

	c.Header("Content-Type", "application/x-apple-aspen-config")
	c.String(http.StatusOK, string(config))
	return
}
