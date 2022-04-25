/**
 * @Author: alessonhu
 * @Description:
 * @File:  main.go
 * @Version: 1.0.0
 * @Date: 2020/12/23 14:13
 */
package main

import (
	"flag"
	"github.com/Coccodrillo/singleprocess"
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis"
	"io/ioutil"
	"mdm/common/middleware"
	"mdm/common/models"
	"mdm/common/qm"
	"net/http"
)

type CConfigServer struct {
	MdmSvrAddr    string `toml:"mdm_server_ipport_server"`
	MdmExposePort string `toml:"mdm_server_expose_port"`
	HttpsCert     string
	HttpsKey      string
	HttpsAddr     string
}

type MdmConfig struct {
	TlsCertPath       string `toml:"tls_cert_path"`
	TlsKeyPath        string `toml:"tls_key_path"`
	TokenExpireSecond int    `toml:"token_expire_second"`
	JwtSecret         string `toml:"jwt_secret"`
}

type configInfo struct {
	Svc         *mdmService
	GRedis      *redis.Client
	ApnsSvc     *apnsService
	JwtSignTool *JwtSign
	WakePool    *Pool
	Db          models.CDBConfigData `toml:"db"`
	Server      CConfigServer        `toml:"server"`
	MdmSvc      MdmConfig            `toml:"mdm_svr"`
}

const (
	GRedisMdm = 7
	GPoolNum  = 100
)

var (
	flOrg          = flag.String("orgName", "ioa", "Mdm push organization.")
	flConnDomain   = flag.String("connDomain", "", "Mdm server address, such as https://testngn.ioatest.net:28800")
	flConnCertPath = flag.String("connCert", "", "Connect cert path, such as ./svr.svr")
	flConnKeyPath  = flag.String("connKey", "", "Connect key path, such as ./svr.key")
	flConnCertPass = flag.String("connPass", "123456", "Connect cert password.")
	flPushCertPath = flag.String("pushPath", "", "Mdm push cert path, such as ./keystore.p12")
	flPushCertPass = flag.String("pushPass", "keystore", "Mdm push cert password.")

	flRedisAddr = flag.String("redisAddr", "", "Redis Addr, such as 127.0.0.1:16379")
	flRedisPass = flag.String("redisPass", "", "Redis Password.")
	flSvrAddr   = flag.String("svrAddr", ":28800", "Svr :Port, such as :28800")
)

var g_app configInfo

func main() {
	defer qm.LOG_FLUSH()
	defer RECOVER_FUNC()

	if singleprocess.IsAnotherInstanceRunning() {
		qm.LOG_ERROR("process is already running")
		return
	}

	flag.Parse()
	printFlInfo()

	g_app.Db.RedisIPPort = *flRedisAddr
	g_app.Db.RedisPassword = *flRedisPass
	g_app.Server.MdmSvrAddr = *flSvrAddr

	if !qm.GetConfigStruct(&g_app) {
		qm.LOG_ERROR("init config failed")
		//return
	}

	var err error

	if err = models.InitDB(g_app.Db.SourceName); err != nil {
		qm.LOG_ERROR_F("connect to db failed:%s", err.Error())
		//return
	}
	g_app.GRedis, err = models.InitRedis(g_app.Db.RedisIPPort, g_app.Db.RedisPassword, GRedisMdm)
	if err != nil {
		qm.LOG_ERROR("connect to redis failed:%s", err.Error())
		return
	}

	g_app.ApnsSvc = &apnsService{
		PushCertPath: *flPushCertPath,
		PushCertPass: *flPushCertPass,
	}

	g_app.JwtSignTool = &JwtSign{hmacSecret: g_app.MdmSvc.JwtSecret}
	g_app.WakePool = NewPool(GPoolNum)
	go g_app.WakePool.Run()

	go g_app.ApnsSvc.InitPushClient()

	gin.SetMode(gin.ReleaseMode)
	gin.DefaultWriter = new(models.GinLogger)
	gin.DefaultErrorWriter = new(models.GinLogger)
	gin.DisableConsoleColor()
	r := gin.Default()

	var tlsP12 []byte
	if *flConnCertPath != "" && *flConnKeyPath != "" {
		tlsCert, err := ioutil.ReadFile(*flConnCertPath)
		if err != nil {
			qm.LOG_ERROR(err)
			return
		}
		tlsKey, err := ioutil.ReadFile(*flConnKeyPath)
		if err != nil {
			qm.LOG_ERROR(err)
			return
		}
		tlsP12, err = TransToP12(tlsCert, tlsKey)
		if err != nil {
			qm.LOG_ERROR(err)
			return
		}
	}
	g_app.Svc = &mdmService{
		ORGANIZATION: *flOrg,
		SVRURL:       *flConnDomain,
		TLSCert:      tlsP12,
		TLSPass:      *flConnCertPass,
	}

	//go UpdateCertInfo()
	go WakeDevicesByTime()

	connStat, _ := middleware.NewConnscStat()
	go connStat.Start()

	mdmGroup := r.Group("/api/public/mdm")
	mdmGroup.Use(middleware.Limit(connStat, "/api/public/mdm", func(option *middleware.LimitOption) {
		//限流和统计,使用默认配置
	}))
	{
		mdmGroup.GET("/ping", func(c *gin.Context) {
			c.String(http.StatusOK, "pong")
		})
		mdmGroup.GET("/gettoken", g_app.Svc.MdmToken)
		mdmGroup.GET("/enroll", g_app.Svc.MdmEnroll)
		mdmGroup.PUT("/checkin", g_app.Svc.MdmCheckin)
		mdmGroup.PUT("/connect", g_app.Svc.MdmConnect)
		mdmGroup.POST("/pushevent", g_app.Svc.MdmEvent)
		mdmGroup.POST("/querystate", g_app.Svc.QueryDeviceState)
		//for debug
		mdmGroup.GET("/activate", g_app.Svc.MdmWake)
	}
	qm.LOG_INFO_F("run err is %v", r.RunTLS(g_app.Server.MdmSvrAddr, *flConnCertPath, *flConnKeyPath))

}

func printFlInfo() {
	qm.LOG_INFO_F("[flag info print start]")
	qm.LOG_INFO_F("%+v", *flOrg)
	qm.LOG_INFO_F("%+v", *flConnCertPath)
	qm.LOG_INFO_F("%+v", *flConnCertPass)
	qm.LOG_INFO_F("%+v", *flPushCertPath)
	qm.LOG_INFO_F("%+v", *flPushCertPass)
	qm.LOG_INFO_F("[flag info print end]")
}
