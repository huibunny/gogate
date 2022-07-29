package server

import (
	"github.com/valyala/fasthttp"
	"github.com/wanghongfei/gogate/conf"
	"github.com/wanghongfei/gogate/utils"
)

// 从请求上下文中取出*ServiceInfo
func GetServiceInfoFromUserValue(ctx *fasthttp.RequestCtx, key string) (*conf.ServiceInfo, bool) {
	val := ctx.UserValue(key)
	if nil == val {
		return nil, false
	}

	info, ok := val.(*conf.ServiceInfo)
	if !ok {
		return nil, false
	}

	return info, true
}

// 从请求上下文中取出string
func GetStringFromUserValue(ctx *fasthttp.RequestCtx, key string) string {
	val := ctx.UserValue(key)
	if nil == val {
		return ""
	}

	str, ok := val.(string)
	if !ok {
		return ""
	}

	return str
}

func GetInt64FromUserValue(ctx *fasthttp.RequestCtx, key string) int64 {
	val := ctx.UserValue(key)
	if nil == val {
		return -1
	}

	num, ok := val.(int64)
	if !ok {
		return -1
	}

	return num
}

func GetStopWatchFromUserValue(ctx *fasthttp.RequestCtx) *utils.Stopwatch {
	return ctx.UserValue(STOPWATCH).(*utils.Stopwatch)
}
