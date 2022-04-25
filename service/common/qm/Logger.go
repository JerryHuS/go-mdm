package qm

import (
	"github.com/cihub/seelog"
	"os"
	"runtime"
	"strings"
	"time"
)

const default_log_xml string = `<seelog type="sync" minlevel="trace" maxlevel="error">
		<outputs formatid="rolllog">
			<console/>
			<rollingfile formatid="rolllog" type="size" filename="./$(file_name).log" maxsize="10485760" maxrolls="5" />
		</outputs>
		<formats>
			<format id="rolllog" format="%Date %Time [%LEVEL] [%Func] [%File.%Line] %Msg%n"/>
		</formats>
	</seelog>`

var g_log seelog.LoggerInterface

func LOG_INFO(v ...interface{}) {
	g_log.Info(v)
}

func LOG_TRACE(v ...interface{}) {
	g_log.Trace(v)
}

func LOG_DEBUG(v ...interface{}) {
	g_log.Debug(v)
}

func LOG_WARN(v ...interface{}) {
	g_log.Warn(v)
}

func LOG_ERROR(v ...interface{}) {
	g_log.Error(v)
}

func LOG_INFO_F(format string, v ...interface{}) {
	g_log.Infof(format, v...)
}

func LOG_TRACE_F(format string, v ...interface{}) {
	g_log.Tracef(format, v...)
}

func LOG_DEBUG_F(format string, v ...interface{}) {
	g_log.Debugf(format, v...)
}

func LOG_WARN_F(format string, v ...interface{}) {
	g_log.Warnf(format, v...)
}

func LOG_ERROR_F(format string, v ...interface{}) {
	g_log.Errorf(format, v...)
}

func parse_xml_conf(str string) string {
	path := GetMainDiectory()
	str = strings.Replace(str, "$(file_dir)", path, -1)
	path = GetExeFileBaseName()
	str = strings.Replace(str, "$(file_name)", path, -1)
	curtime := time.Now().Format("2006_01_02_15_04_05")
	return strings.Replace(str, "$(time)", curtime, -1)
}

func GetDefaultLogger() seelog.LoggerInterface {
	//
	//	优先使用processname_log.xml
	//	其次使用../../service_log.xml
	//	如果不存在配置文件，则使用内置日志格式
	//
	var str string
	var path string

	path = GetMainDiectory() + GetExeFileBaseName() + "_log.xml"
	if PathFileExists(path) {
		str = ReadFileAsString(path)
	}

	if len(str) == 0 {
		path = GetMainPath(".." + string(os.PathSeparator) + ".." + string(os.PathSeparator) + "service_log.xml")
		if PathFileExists(path) {
			str = ReadFileAsString(path)
		}
		if len(str) == 0 {
			str = default_log_xml
		}
	}

	log, _ := seelog.LoggerFromConfigAsString(parse_xml_conf(str))

	if log == nil {
		log = seelog.Default
	}

	return log
}

func getStack() []byte {
	buf := make([]byte, 1024)
	for {
		n := runtime.Stack(buf, false)
		if n < len(buf) {
			return buf[:n]
		}
		buf = make([]byte, 2*len(buf))
	}
}

func LOG_FLUSH() {
	// if err := recover(); err != nil {
	// 	LOG_ERROR(string(getStack()))
	// }
	g_log.Flush()
}

func init() {
	g_log = GetDefaultLogger()
	g_log.SetAdditionalStackDepth(1)
}
