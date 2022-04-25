/**
 * @Author: alessonhu
 * @Description:
 * @File:  limit_test.go
 * @Version: 1.0.0
 * @Date: 2020/7/3 14:21
 */
package middleware

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"net/http"
	"testing"
	"time"
)

func TestLimit(t *testing.T) {
	connStat, _ := NewConnscStat()
	go connStat.Start()
	r := gin.Default()
	test := r.Group("/demo")
	test.Use(Limit(connStat, "/demo", func(option *LimitOption) {
		//默认为60次每分钟，不传就用默认值
		option.Request = 1
		option.Time = time.Second
	}))
	{
		test.GET("/hello", func(c *gin.Context) {
			c.String(http.StatusOK, "world")
		})
		test.GET("/hello111", func(c *gin.Context) {
			c.String(http.StatusOK, "world111")
		})
	}
	fmt.Println(r.Run(":6443"))
}
