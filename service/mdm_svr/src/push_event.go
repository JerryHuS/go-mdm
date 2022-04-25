/**
 * @Author: alessonhu
 * @Description:
 * @File:  push_event.go
 * @Version: 1.0.0
 * @Date: 2021/1/12 15:36
 */
package main

import (
	"github.com/gin-gonic/gin"
	"mdm/common/models"
	"mdm/common/qm"
	"net/http"
	"time"
)

type EventReq struct {
	Mid         string `json:"mid"`
	RequestType string `json:"request_type"`
	PayloadUuid string `json:"payload_uuid"`
	Payload     string `json:"payload"`
}

var awakeChan = make(chan string)

func (svc *mdmService) MdmEvent(c *gin.Context) {
	req := &EventReq{}
	//内容加密,自定义

	qm.LOG_DEBUG_F("req is %+v", req)

	token := c.Request.Header.Get("Token")
	tokenValid := CheckMdmToken(token, req.Mid)
	if !tokenValid {
		qm.LOG_ERROR_F("token is err, token is %s, mid is %s", token, req.Mid)
		c.String(http.StatusOK, "token is error")
		return
	}
	res, err := g_app.GRedis.HMGet(MdmDeviceInfoKey+req.Mid, CheckKeyUdid).Result()
	if err != nil {
		qm.LOG_ERROR(err)
		MakeMdmApiRsp(c, SvrCode, "err server", nil)
		return
	}
	udid, ok := res[0].(string)
	if !ok {
		MakeMdmApiRsp(c, SvrCode, "device don't enroll", nil)
		return
	}
	defer svc.AwakeDevice(err, req.Mid)
	switch req.RequestType {
	case "InstallProfile":
		uuid, profileInfo, err := NewCommandPayload([]byte(req.Payload))
		if err != nil {
			qm.LOG_ERROR(err)
			MakeMdmApiRsp(c, SvrCode, "err server", nil)
			return
		}
		err = g_app.GRedis.RPush(MdmReadyProfileKey+udid, profileInfo).Err()
		if err != nil {
			qm.LOG_ERROR(err)
			MakeMdmApiRsp(c, SvrCode, "err server", nil)
			return
		}
		err = models.InsertMdmEventinfo(&models.MdmEventInfo{
			Mid:         req.Mid,
			Udid:        udid,
			RequestType: "InstallProfile",
			Uuid:        uuid,
			State:       0,
			Payloaduuid: req.PayloadUuid,
		})
		if err != nil {
			qm.LOG_ERROR(err)
			MakeMdmApiRsp(c, SvrCode, "err server", nil)
			return
		}
		MakeMdmApiRsp(c, SuccCode, "", nil)
	default:
		MakeMdmApiRsp(c, ErrCode, "err request_type", nil)
	}
	return
}

func WakeDevicesByTime() {
	//定时检测设备指令队列，如果有堆积，就唤醒设备拉取。
	ticker := time.NewTicker(time.Minute * time.Duration(30))
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			WakeDevices()
		}
	}
}

func WakeDevices() {
	profileKey, err := g_app.GRedis.Keys(MdmReadyProfileKey).Result()
	if err != nil {
		qm.LOG_ERROR(err)
		return
	}
	qm.LOG_INFO_F("profile key not run has %d", len(profileKey))

	return
}

func (svc *mdmService) AwakeDevice(err error, mid string) {
	if err == nil {
		g_app.WakePool.EntryChannel <- &Task{mid: mid}
	}
	return
}

//弄个协程池去接收，apns不知道有没有频率限次，先设为100
type Task struct {
	mid string
}

func (t *Task) DealTask() {
	qm.LOG_INFO_F("[apns start] mid is %s", t.mid)
	defer qm.LOG_INFO_F("[apns end] mid is %s", t.mid)
	var token, magic string
	var ok bool
	res, err := g_app.GRedis.HMGet(MdmDeviceInfoKey+t.mid, "udid", "token", "pushmagic").Result()
	if err != nil {
		qm.LOG_ERROR(err)
	}
	qm.LOG_INFO_F("%+v", res)

	if token, ok = res[1].(string); !ok {
		qm.LOG_ERROR_F("[apns wake] token is err")
		return
	}
	if magic, ok = res[2].(string); !ok {
		qm.LOG_ERROR_F("[apns wake] magic is err")
		return
	}
	g_app.ApnsSvc.Push(token, magic)
}

//定义池类型
type Pool struct {
	//对外接收Task的入口
	EntryChannel chan *Task

	//协程池最大worker数量,限定Goroutine的个数，默认100
	WorkerNum int

	//协程池内部的任务就绪队列
	JobsChannel chan *Task
}

//创建一个协程池
func NewPool(cap int) *Pool {
	p := Pool{
		EntryChannel: make(chan *Task),
		WorkerNum:    cap,
		JobsChannel:  make(chan *Task),
	}

	return &p
}

//协程池创建一个worker并且开始工作
func (p *Pool) worker(work_ID int) {
	//worker不断的从JobsChannel内部任务队列中拿任务
	for task := range p.JobsChannel {
		//如果拿到任务,则执行task任务\
		qm.LOG_INFO_F("[worker] id is %d", work_ID)
		task.DealTask()
	}
}

//让协程池Pool开始工作
func (p *Pool) Run() {
	defer RECOVER_FUNC()
	//1,首先根据协程池的worker数量限定,开启固定数量的Worker,
	//  每一个Worker用一个Goroutine承载
	for i := 0; i < p.WorkerNum; i++ {
		go p.worker(i)
	}

	//2, 从EntryChannel协程池入口取外界传递过来的任务
	//   并且将任务送进JobsChannel中
	for task := range p.EntryChannel {
		p.JobsChannel <- task
	}

	//3, 执行完毕需要关闭JobsChannel
	close(p.JobsChannel)

	//4, 执行完毕需要关闭EntryChannel
	close(p.EntryChannel)
}
