package server

import (
	"strings"

	"github.com/wanghongfei/gogate/conf"
	. "github.com/wanghongfei/gogate/conf"
	"github.com/wanghongfei/gogate/discovery"
	"github.com/wanghongfei/gogate/perr"

	"github.com/valyala/fasthttp"
	"github.com/wanghongfei/gogate/utils"
)

const META_VERSION = "version"

// 转发请求到指定微服务
// return:
// Response: 响应对象;
// string: 下游服务名
// error: 错误
func (serv *Server) sendRequest(ctx *fasthttp.RequestCtx, req *fasthttp.Request) (*fasthttp.Response, string, error) {
	// 获取服务信息
	info := ctx.UserValue(ROUTE_INFO).(*conf.ServiceInfo)

	var logRecordName string
	// 需要从注册列表中查询地址
	if info.Id != "" {
		if serv.IsInStaticMode() {
			return nil, "", perr.WrapBizErrorf(nil, "no static address found for this service")
		}

		logRecordName = info.Id

		// 获取Client
		appId := strings.ToUpper(info.Id)

		// 灰度, 选择版本
		version := chooseVersion(info.Canary)

		// 取出指定服务的所有实例
		serviceInstances := serv.discoveryClient.Get(appId)
		if nil == serviceInstances {
			return nil, "", perr.WrapBizErrorf(nil, "no instance %s for service (service is offline)", appId)
		}

		// 按version过滤
		if "" != version {
			serviceInstances = filterWithVersion(serviceInstances, version)
			if 0 == len(serviceInstances) {
				// 此version下没有实例
				return nil, "", perr.WrapBizErrorf(nil, "no instance %s:%s for service", appId, version)
			}
		}

		// 负载均衡
		targetInstance := serv.lb.Choose(serviceInstances)
		// 修改请求的host为目标主机地址
		req.URI().SetHost(targetInstance.Addr)

	} else {
		logRecordName = info.Name

		// 直接使用后面的地址
		hostList := strings.Split(info.Host, ",")

		targetAddr := serv.lb.ChooseByAddresses(hostList)
		req.URI().SetHost(targetAddr)
	}

	// 发请求
	resp := new(fasthttp.Response)
	err := serv.fastClient.Do(req, resp)
	if nil != err {
		return nil, "", perr.WrapSystemErrorf(nil, "failed to send request to downstream service")
	}

	return resp, logRecordName, nil
}

// 过滤出meta里version字段为指定值的实例
func filterWithVersion(instances []*discovery.InstanceInfo, targetVersion string) []*discovery.InstanceInfo {
	result := make([]*discovery.InstanceInfo, 0, 5)

	for _, ins := range instances {
		if ins.Meta[META_VERSION] == targetVersion {
			result = append(result, ins)
		}
	}

	return result
}

func chooseVersion(canaryInfos []*conf.CanaryInfo) string {
	if nil == canaryInfos || len(canaryInfos) == 0 {
		return ""
	}

	var weights []int
	for _, info := range canaryInfos {
		weights = append(weights, info.Weight)
	}

	index := utils.RandomByWeight(weights)
	if -1 == index {
		Log.Warn("random interval returned -1")
		return ""
	}

	return canaryInfos[index].Meta
}
