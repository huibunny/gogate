package server

import (
	"github.com/valyala/fasthttp"
	. "github.com/wanghongfei/gogate/conf"
)

func ServiceMatchPreFilter(s *Server, ctx *fasthttp.RequestCtx, newRequest *fasthttp.Request) bool {
	uri := GetStringFromUserValue(ctx, REQUEST_PATH)

	if uri == s.checkApi {
		ctx.Response.SetStatusCode(200)
		NewResponse(ctx.UserValue(REQUEST_PATH).(string), "success").Send(ctx)
		return false
	} else {
		servInfo := s.Router.Match(uri)
		if nil == servInfo {
			// 没匹配到
			ctx.Response.SetStatusCode(404)
			NewResponse(ctx.UserValue(REQUEST_PATH).(string), "no match").Send(ctx)
			return false
		} else if IsInWhiteList(servInfo, uri) {
		} else if servInfo.Verify {
			user_id, err := VerifyToken(ctx, s.Secret)
			if err != nil {
				ctx.Response.SetStatusCode(401)
				NewResponse(ctx.UserValue(REQUEST_PATH).(string), "token error").Send(ctx)
				return false
			}
			newRequest.Header.Del(AUTHORIZATION)
			newRequest.Header.Set(USER_ID, user_id)
		}
		ctx.SetUserValue(ROUTE_INFO, servInfo)
		ctx.SetUserValue(SERVICE_NAME, servInfo.Id)
		Log.Debugf("%s matched to %s", uri, servInfo.Id)
	}

	return true
}
