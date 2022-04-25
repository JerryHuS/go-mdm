/**
 * @Author: alessonhu
 * @Description:
 * @File:  statistics.go
 * @Version: 1.0.0
 * @Date: 2020/7/3 9:40
 */
package middleware

import (
	"fmt"
	"github.com/cihub/seelog"
	. "mdm/common/qm"
	"strings"
	"sync"
	"time"
)

var (
	path     = GetMainDiectory()
	statTime = time.Minute
)

type ConnscStat struct {
	AllowCalls        map[string]uint64
	RejectCalls       map[string]uint64
	allowMutex        sync.RWMutex
	rejectMutex       sync.RWMutex
	statLogger        seelog.LoggerInterface
	RequestFlow       map[string]uint64
	RequestFlowMutex  sync.RWMutex
	ResponseFlow      map[string]uint64
	ResponseFlowMutex sync.RWMutex
}

const stat_log_xml string = `
<seelog type="asyncloop" minlevel="info" maxlevel="error">
	<outputs formatid="rolllog">
		<rollingfile formatid="rolllog" type="size" filename="$(file_dir)../../../../logs/$(file_name)_stat.log" maxsize="20971520" maxrolls="2" />
	<filter levels="error">
		<rollingfile formatid="rolllog" type="size" filename="$(file_dir)../../../../logs/error_$(file_name)_stat.log" maxsize="20971520" maxrolls="2" />
	</filter>
	</outputs>
	<formats>
		<format id="rolllog" format="%Date %Time [%l] [%Func] [%File.%Line] %Msg%n"/>
	</formats>
</seelog>
`

func NewConnscStat() (*ConnscStat, error) {
	cs := &ConnscStat{
		AllowCalls:   make(map[string]uint64),
		RejectCalls:  make(map[string]uint64),
		RequestFlow:  make(map[string]uint64),
		ResponseFlow: make(map[string]uint64),
	}
	cs.statLogger, _ = seelog.LoggerFromConfigAsString(parse_xml_conf(stat_log_xml))
	return cs, nil
}

func (cs *ConnscStat) Start() {
	LOG_INFO_F(">>>> %s api statistics count <<<<", path)
	ticker := time.NewTicker(statTime)
	for {
		select {
		case <-ticker.C:
			//统计日志输出
			cs.PrintStats()
		}
	}
}

func (cs *ConnscStat) PrintStats() {
	cs.statLogger.Info(">>> stat start")
	totalCalls := uint64(0)

	cs.allowMutex.Lock()
	for cmd, value := range cs.AllowCalls {
		totalCalls += value
		cs.statLogger.Infof("+++ api [%s] allow calls count: %d", cmd, cs.AllowCalls[cmd])
		cs.AllowCalls[cmd] = 0
	}
	cs.allowMutex.Unlock()

	cs.rejectMutex.Lock()
	for cmd, value := range cs.RejectCalls {
		totalCalls += value
		cs.statLogger.Infof("+++ api [%s] reject calls count: %d", cmd, cs.RejectCalls[cmd])
		cs.RejectCalls[cmd] = 0
	}
	cs.rejectMutex.Unlock()

	cs.statLogger.Infof("+++ api total calls count: %d", totalCalls)

	cs.RequestFlowMutex.Lock()
	totalRequestFlow := uint64(0)
	for cmd, _ := range cs.RequestFlow {
		totalRequestFlow += cs.RequestFlow[cmd]
		cs.statLogger.Infof("+++ api [%s] request flow : %s", cmd, toK(cs.RequestFlow[cmd]))
		cs.RequestFlow[cmd] = 0
	}
	cs.RequestFlowMutex.Unlock()

	cs.statLogger.Infof("+++ all request flow : %s", toK(totalRequestFlow))
	cs.statLogger.Info(">>> stat end")
}

func (cs *ConnscStat) UpdateAllowCount(cmd string) {
	cs.allowMutex.Lock()
	currentCount := cs.AllowCalls[cmd]
	cs.AllowCalls[cmd] = currentCount + 1
	cs.allowMutex.Unlock()
}

func (cs *ConnscStat) UpdateDenyCount(cmd string) {
	cs.rejectMutex.Lock()
	currentCount := cs.RejectCalls[cmd]
	cs.RejectCalls[cmd] = currentCount + 1
	cs.rejectMutex.Unlock()
}

func (cs *ConnscStat) IncrRequestFlow(cmd string, flowbytes uint64) {
	cs.RequestFlowMutex.Lock()
	currentFlow, ok := cs.RequestFlow[cmd]
	if !ok {
		currentFlow = 0
	}
	cs.RequestFlow[cmd] = currentFlow + flowbytes
	cs.RequestFlowMutex.Unlock()
}

func (cs *ConnscStat) IncrResponseFlow(cmd string, flowbytes uint64) {
	cs.ResponseFlowMutex.Lock()
	currentFlow, ok := cs.ResponseFlow[cmd]
	if !ok {
		currentFlow = 0
	}
	cs.ResponseFlow[cmd] = currentFlow + flowbytes
	cs.ResponseFlowMutex.Unlock()
}

func parse_xml_conf(str string) string {
	str = strings.Replace(str, "$(file_dir)", path, -1)
	path = GetExeFileBaseName()
	str = strings.Replace(str, "$(file_name)", path, -1)
	curtime := time.Now().Format("2006_01_02_15_04_05")
	return strings.Replace(str, "$(time)", curtime, -1)
}

func toK(bytes uint64) string {
	return fmt.Sprintf("%.2fK", float64(bytes)/1024)
}

func toH(bytes uint64) string {
	switch {
	case bytes < 1024:
		return fmt.Sprintf("%dB", bytes)
	case bytes < 1024*1024:
		return fmt.Sprintf("%.2fK", float64(bytes)/1024)
	case bytes < 1024*1024*1024:
		return fmt.Sprintf("%.2fM", float64(bytes)/1024/1024)
	default:
		return fmt.Sprintf("%.2fG", float64(bytes)/1024/1024/1024)
	}
}
