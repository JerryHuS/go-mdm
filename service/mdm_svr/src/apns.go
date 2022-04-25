/**
 * @Author: alessonhu
 * @Description:
 * @File:  apns.go
 * @Version: 1.0.0
 * @Date: 2021/1/5 11:19
 */
package main

import (
	"crypto/tls"
	"github.com/sideshow/apns2"
	"github.com/sideshow/apns2/certificate"
	"github.com/sideshow/apns2/payload"
	"mdm/common/qm"
	"strings"
)

type apnsService struct {
	PushCertData  string
	PushCertPath  string
	PushCertPass  string
	PushCertTopic string
	PushClient    *apns2.Client
}

const (
	MdmPushP12FromDb  = "MdmPushP12"
	MdmPushPassFromDb = "MdmPushPass"
)

func (svc *apnsService) InitPushClient() (err error) {
	var cert tls.Certificate
	qm.LOG_INFO_F("[init push cert] start.")
	defer qm.LOG_INFO_F("[init push cert] end.")

	//调试使用
	if *flConnCertPath != "" && *flConnCertPass != "" {
		cert, err = certificate.FromP12File(svc.PushCertPath, svc.PushCertPass)
		if err != nil {
			qm.LOG_INFO_F("[init push cert] err:%s, certpath:%s, certpass:%s", err.Error(), svc.PushCertPath, svc.PushCertPass)
			return
		}
	} else {
		//从db加载
		if g_app.ApnsSvc.PushCertData == "" || g_app.ApnsSvc.PushCertPass == "" {
			qm.LOG_INFO_F("[init push cert] push cert not prepared, init later")
			return
		} else {

			cert, err = certificate.FromP12Bytes([]byte(g_app.ApnsSvc.PushCertData), g_app.ApnsSvc.PushCertPass)
			if err != nil {
				qm.LOG_ERROR_F("[init push cert] push cert err: %s", err.Error())
				return
			}
		}
	}
	commonNameArr := strings.Split(cert.Leaf.Subject.CommonName, ":")
	svc.PushCertTopic = commonNameArr[len(commonNameArr)-1]
	qm.LOG_INFO("[init push cert] push cert common name is :", cert.Leaf.Subject.CommonName)
	qm.LOG_INFO("[init push cert] push cert topic is :", svc.PushCertTopic)
	svc.PushClient = apns2.NewClient(cert).Production()
	return
}

func (svc *apnsService) Push(token, pushmagic string) {
	if svc.PushClient == nil {
		qm.LOG_ERROR_F("[apns] push client nil, re init...")
		svc.InitPushClient()
		if svc.PushClient == nil {
			qm.LOG_ERROR_F("[apns] push client nil, exit...")
		}
		return
	}
	notif := &apns2.Notification{
		DeviceToken: token,
		Payload:     payload.NewPayload().Mdm(pushmagic),
	}
	res, err := svc.PushClient.Push(notif)
	qm.LOG_DEBUG_F("%+v", res)
	if err != nil {
		qm.LOG_ERROR_F("[apns] push data err, %s", err.Error())
		return
	}
	if !res.Sent() {
		qm.LOG_INFO_F("[apns] apns push fail, token:%s, magic:%s, code:%d, time:%s, reason:%s", token, pushmagic, res.StatusCode, res.Timestamp.String(), res.Reason)
		return
	}
	qm.LOG_INFO_F("[apns] apns push succ, token:%s, magic:%s, code:%d, time:%s", token, pushmagic, res.StatusCode, res.Timestamp.String())
	return
}
