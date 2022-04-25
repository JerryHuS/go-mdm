/**
 * @Author: alessonhu
 * @Description:
 * @File:  util.go
 * @Version: 1.0.0
 * @Date: 2020/12/28 16:52
 */
package main

import (
	"bytes"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	jwt "github.com/dgrijalva/jwt-go/v4"
	"io/ioutil"
	"math/big"
	"mdm/common/qm"
	"net/http"
	"runtime"
	"time"
)

const (
	ClaimsMid    = "token_mid"
	ClaimsExpire = "token_expire"
	ClaimsTimes  = "toke_times"
)

func RECOVER_FUNC() { // 必须要先声明defer，否则不能捕获到panic异常
	if err := recover(); err != nil {
		qm.LOG_INFO("recover_main start")
		qm.LOG_INFO(err) // 这里的err其实就是panic传入的内容
		var buf [4096]byte
		n := runtime.Stack(buf[:], false)
		qm.LOG_INFO(string(buf[:n]))
		qm.LOG_INFO("recover_main end")
	}
}

//接口鉴权
func CheckAppSign(req *http.Request) bool {
	//nowUnix := time.Now().Unix()
	//srandom, _ := strconv.ParseInt(req.Header.Get("STimestamp"), 10, 64)
	body, _ := ioutil.ReadAll(req.Body)
	//再重新写回请求体body中，ioutil.ReadAll会清空c.Request.Body中的数据
	req.Body = ioutil.NopCloser(bytes.NewBuffer(body))

	qm.LOG_DEBUG_F("SSignature ERR。SSignature|STimestamp|SRandom|SAppId(mid)|Body is ")
	qm.LOG_DEBUG_F("%s|%s|%s|%s|%s", req.Header.Get("SSignature"), req.Header.Get("STimestamp"), req.Header.Get("SRandom"), req.Header.Get("SAppId"), string(body))
	qm.LOG_DEBUG_F("ServerSign is %s", getSha256(req.Header.Get("STimestamp")+req.Header.Get("SRandom")+req.Header.Get("SAppId")+string(body)))

	if req.Header.Get("SSignature") == getSha256(req.Header.Get("STimestamp")+req.Header.Get("SRandom")+req.Header.Get("SAppId")+string(body)) ||
		req.Header.Get("SSignature") == qm.Md5Encode([]byte(req.Header.Get("STimestamp")+req.Header.Get("SRandom")+req.Header.Get("SAppId")+string(body))) {
		return true
	} else {
		qm.LOG_ERROR_F("+++check sign err+++")
		qm.LOG_ERROR_F("SSignature ERR。SSignature|STimestamp|SRandom|SAppId(mid)|Body")
		qm.LOG_ERROR_F("%s|%s|%s|%s|%s", req.Header.Get("SSignature"), req.Header.Get("STimestamp"), req.Header.Get("SRandom"), req.Header.Get("SAppId"), string(body))
		qm.LOG_ERROR_F("ServerSign is %s", getSha256(req.Header.Get("STimestamp")+req.Header.Get("SRandom")+req.Header.Get("SAppId")+string(body)))
		return false
	}
}

func getSha256(data string) string {
	h := sha256.New()
	h.Write([]byte(data))
	return hex.EncodeToString(h.Sum(nil))
}

type JwtSign struct {
	hmacSecret string
}

func (s *JwtSign) CheckSign(tokenString string) (jwt.MapClaims, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// Don't forget to validate the alg is what you expect:
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
		}
		// hmacSampleSecret is a []byte containing your secret, e.g. []byte("my_secret_key")
		return []byte(s.hmacSecret), nil
	})

	if err != nil {
		qm.LOG_ERROR_F("parse token error, token[%s] err[%s]", tokenString, err.Error())
		return nil, err
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		qm.LOG_ERROR_F("invalid token [%s]", tokenString)
		return nil, errors.New("invalid token")
	}
	return claims, nil
}

func (s *JwtSign) GetSign(claims jwt.MapClaims) (tokenString string, err error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	// Sign and get the complete encoded token as a string using the secret
	return token.SignedString([]byte(s.hmacSecret))
}

func CheckMdmToken(token, mid string) bool {
	if token == "" || mid == "" {
		qm.LOG_ERROR_F("[check token err] token is %s, mid is %s", token, mid)
		return false
	}
	claims, err := g_app.JwtSignTool.CheckSign(token)
	if err != nil || claims[ClaimsMid] == nil || claims[ClaimsExpire] == nil || claims[ClaimsTimes] == nil {
		qm.LOG_ERROR_F("[check token err] claims err, %+v", claims)
		return false
	}
	if claims[ClaimsMid].(string) != mid {
		qm.LOG_ERROR_F("[check token err] mid not matched, token mid is %s, req mid is %s", claims[ClaimsMid].(string), mid)
		return false
	}
	if claims[ClaimsExpire].(string) < time.Now().Format("2006-01-02 15:04:05") {
		qm.LOG_ERROR_F("[check token err] token expired, token expired time is %s, mid is %s", claims[ClaimsExpire].(string), claims[ClaimsMid].(string))
		return false
	}
	return true
}

func GetRandomString(len int) string {
	var container string
	var str = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890"
	b := bytes.NewBufferString(str)
	length := b.Len()
	bigInt := big.NewInt(int64(length))
	for i := 0; i < len; i++ {
		randomInt, _ := rand.Int(rand.Reader, bigInt)
		container += string(str[randomInt.Int64()])
	}
	return container
}
