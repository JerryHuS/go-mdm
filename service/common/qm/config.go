/**
 * @Author: alessonhu
 * @Description:
 * @File:  config.go
 * @Version: 1.0.0
 * @Date: 2021/3/10 19:29
 */
package qm

import (
	"github.com/BurntSushi/toml"
	"os"
)

var g_config_path string

//
//	获取toml配置类型对象
//
func GetConfigStructEx(path string, v interface{}) bool {
	_, err := toml.DecodeFile(path, v)
	if err != nil {
		return false
	}
	return true
}

//
//	获取toml配置类型对象
//
func GetConfigStruct(v interface{}) bool {
	return GetConfigStructEx(g_config_path, v)
}

func init() {
	g_config_path = GetMainDiectory()
	g_config_path += ".." + string(os.PathSeparator) + ".." + string(os.PathSeparator) + "service.conf"
}
