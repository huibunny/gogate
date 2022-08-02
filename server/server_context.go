package server

import (
	"bytes"
	"errors"

	"github.com/valyala/fasthttp"
	"github.com/wanghongfei/gogate/conf"
	"github.com/wanghongfei/gogate/utils"
)

func VerifyToken(ctx *fasthttp.RequestCtx, secret string) error {
	// Get the Basic Authentication credentials
	auth := ctx.Request.Header.Peek("Authorization")
	strAuth := string(auth)
	userInfo, err := ParseToken(token, secret)
	if err != nil {
		err = errors.New("ParseToken returns error: %v.", err)
	} else {
		userName := userInfo["username"]
		password := userInfo["password"]
		createTime := userInfo["create_time"]
		tokenExpire := userInfo["token_expire"]
		if len(userInfo["username"]) > 0 && len(userInfo["password"]) > 0 {
			//
		} else {
			//
		}
	}

	return err
}

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
