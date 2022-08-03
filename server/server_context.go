package server

import (
	"errors"

	coreutils "github.com/huibunny/gocore/utils"
	"github.com/valyala/fasthttp"
	"github.com/wanghongfei/gogate/conf"
	"github.com/wanghongfei/gogate/utils"
)

func VerifyToken(ctx *fasthttp.RequestCtx, secret string) error {
	// Get the Basic Authentication credentials
	auth := ctx.Request.Header.Peek("Authorization")
	strAuth := string(auth)
	userName, password, expireTime, _, err := coreutils.ParseToken(strAuth, secret)
	if err != nil {
		err = errors.New("ParseToken returns error: " + err.Error())
	} else {
		if len(userName) > 0 && len(password) > 0 {
			now := coreutils.CurrentTime()
			if expireTime <= now {
				err = errors.New("token expired")
			} else {
				//
			}
		} else {
			err = errors.New("token failure")
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
