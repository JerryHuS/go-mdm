/**
 * @Author: alessonhu
 * @Description:
 * @File:  limit.go
 * @Version: 1.0.0
 * @Date: 2020/7/3 9:39
 */
package middleware

import (
	"github.com/didip/tollbooth"
	"github.com/didip/tollbooth/limiter"
	"github.com/gin-gonic/gin"
	"net/http"
	. "mdm/common/qm"
	"sync"
	"time"
)

// Change the the map to hold values of the type visitor.
var (
	visitors = make(map[string]*visitor)
	mtx      sync.Mutex
)

type ModLimitOption func(option *LimitOption)

type LimitOption struct {
	Request int
	Time    time.Duration
}

type visitor struct {
	limiter  *limiter.Limiter
	lastSeen time.Time
}

// Run a background goroutine to remove old entries from the visitors map.
func init() {
	go cleanupVisitors()
}

func Limit(cs *ConnscStat, routeGroup string, modOption ModLimitOption) gin.HandlerFunc {
	return func(c *gin.Context) {
		//默认每分钟60次请求
		option := LimitOption{
			Request: 60,
			Time:    time.Minute,
		}
		modOption(&option)
		//规则1：恶意ip加黑（预留）
		//规则2：请求频率限制（ip限流，后续可扩展身份校验）
		ip := c.ClientIP()
		//限流器id：ip+group，不同routergroup不同限流器
		lmt_id := ip + routeGroup
		stat_key := c.Request.Method + ":" + c.Request.URL.Path
		lmt := getVisitor(lmt_id, option.Request, option.Time)
		lmt_err := tollbooth.LimitByRequest(lmt, c.Writer, c.Request)
		cs.IncrRequestFlow(stat_key, uint64(c.Request.ContentLength))
		if lmt_err != nil {
			LOG_INFO_F("limit occurs, err is %s, requestnum is %d, requesttime is %s, requestip is %s", lmt_err, option.Request, option.Time, ip)
			cs.UpdateDenyCount(stat_key)
			c.AbortWithStatus(http.StatusForbidden)
		}
		cs.UpdateAllowCount(stat_key)
		c.Next()
	}
}

func addVisitor(id string, burst int, ttl time.Duration) *limiter.Limiter {
	//rate:发放令牌的速率，单位：个每秒
	rate := float64(burst) / float64(ttl/time.Second)
	lmt := tollbooth.NewLimiter(rate, &limiter.ExpirableOptions{DefaultExpirationTTL: ttl})
	//设置令牌桶容量，即令牌桶有效期内最大令牌数
	lmt.SetBurst(burst)
	// 设置最新访问时间
	mtx.Lock()
	visitors[id] = &visitor{lmt, time.Now()}
	mtx.Unlock()
	return lmt
}

func getVisitor(id string, burst int, ttl time.Duration) *limiter.Limiter {
	mtx.Lock()
	v, exists := visitors[id]
	if !exists {
		mtx.Unlock()
		return addVisitor(id, burst, ttl)
	}
	// visitor的最新访问时间
	v.lastSeen = time.Now()
	mtx.Unlock()
	return v.limiter
}

func cleanupVisitors() {
	for {
		time.Sleep(time.Minute)
		mtx.Lock()
		for id, v := range visitors {
			if time.Now().Sub(v.lastSeen) > 5*time.Minute {
				delete(visitors, id)
			}
		}
		mtx.Unlock()
	}
}
