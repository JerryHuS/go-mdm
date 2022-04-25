/**
 * @Author: alessonhu
 * @Description: 管理端变更证书通知
 * @File:  cert.go
 * @Version: 1.0.0
 * @Date: 2021/2/3 15:53
 */
package main

import (
	"crypto/rand"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"mdm/common/qm"
	pkcs12 "software.sslmate.com/src/go-pkcs12"
)

func TransToP12(certCrt, certKey []byte) (pfx []byte, err error) {
	qm.LOG_DEBUG_F("certCrt is \n %s", string(certCrt))
	qm.LOG_DEBUG_F("certKey is \n %s", string(certKey))
	caBlock, _ := pem.Decode(certCrt)
	crt, err := x509.ParseCertificate(caBlock.Bytes)
	if err != nil {
		err = fmt.Errorf("证书解析异常, Error : %v", err)
		qm.LOG_ERROR_F("%v", err)
		return
	}

	keyBlock, _ := pem.Decode(certKey)
	priKey, err := x509.ParsePKCS1PrivateKey(keyBlock.Bytes)
	if err != nil {
		err = fmt.Errorf("证书密钥解析key异常, Error : %v", err)
		qm.LOG_ERROR_F("%v", err)
		return
	}

	certPwd := *flConnCertPass

	pfx, err = pkcs12.Encode(rand.Reader, priKey, crt, nil, certPwd)
	if err != nil {
		err = fmt.Errorf("pem to p12 转换证书异常, Error : %v", err)
		qm.LOG_ERROR_F("%v", err)
		return
	}

	return pfx, err
}
