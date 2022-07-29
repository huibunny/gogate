package route

import (
	"io/ioutil"
	"os"

	"github.com/wanghongfei/gogate/conf"
	"github.com/wanghongfei/gogate/perr"
	"gopkg.in/yaml.v2"
)

type Router struct {
	// 配置文件路径
	cfgPath string

	cfg *conf.GateConfig

	// path(string) -> *conf.ServiceInfo
	pathMatcher *PathMatcher

	ServInfos []*conf.ServiceInfo
}

/*
* 创建路由器
*
* PARAMS:
*	- path: 路由配置文件路径
*
 */
func NewRouter(cfg *conf.GateConfig, configFile string) (*Router, error) {
	matcher, servInfos, err := loadRoute(cfg, configFile)
	if nil != err {
		return nil, perr.WrapSystemErrorf(err, "failed to load route info")
	}

	return &Router{
		pathMatcher: matcher,
		cfg:         cfg,
		cfgPath:     configFile,
		ServInfos:   servInfos,
	}, nil
}

/*
* 重新加载路由器
 */
func (r *Router) ReloadRoute() error {
	matcher, servInfos, err := loadRoute(r.cfg, r.cfgPath)
	if nil != err {
		return perr.WrapSystemErrorf(err, "failed to load route info")
	}

	r.ServInfos = servInfos
	r.pathMatcher = matcher

	return nil
}

/*
* 根据uri选择一个最匹配的appId
*
* RETURNS:
*	返回最匹配的conf.ServiceInfo
 */
func (r *Router) Match(reqPath string) *conf.ServiceInfo {

	return r.pathMatcher.Match(reqPath)
}

func loadRoute(cfg *conf.GateConfig, path string) (*PathMatcher, []*conf.ServiceInfo, error) {
	servInfos := make([]*conf.ServiceInfo, 0, 10)
	// 构造 path->serviceId 映射
	// 保存到字典树中
	tree := NewTrieTree()
	// 保存到map中
	routeMap := make(map[string]*conf.ServiceInfo)
	if len(path) > 0 {
		// 打开配置文件
		routeFile, err := os.Open(path)
		if nil != err {
			return nil, nil, perr.WrapSystemErrorf(err, "failed to open file")
		}
		defer routeFile.Close()

		// 读取
		buf, err := ioutil.ReadAll(routeFile)
		if nil != err {
			return nil, nil, err
		}

		// 解析yml
		// ymlMap := make(map[string]*conf.ServiceInfo)
		ymlMap := make(map[string]map[string]*conf.ServiceInfo)
		err = yaml.UnmarshalStrict(buf, &ymlMap)
		if nil != err {
			return nil, nil, err
		}
		for name, info := range ymlMap["services"] {
			// 验证
			err = validateServiceInfo(info)
			if nil != err {
				return nil, nil, perr.WrapSystemErrorf(err, "invalid config for %s", name)
			}

			tree.PutString(info.Prefix, info)
			routeMap[info.Prefix] = info

			servInfos = append(servInfos, info)
		}
	} else {
		for _, info := range cfg.Router {
			err := validateServiceInfo(info)
			if nil != err {
				return nil, nil, perr.WrapSystemErrorf(err, "invalid router config")
			}

			tree.PutString(info.Prefix, info)
			routeMap[info.Prefix] = info

			servInfos = append(servInfos, info)
		}
	}

	matcher := &PathMatcher{
		routeMap:      routeMap,
		routeTrieTree: tree,
	}
	return matcher, servInfos, nil
}

func validateServiceInfo(info *conf.ServiceInfo) error {
	if nil == info {
		return perr.WrapSystemErrorf(nil, "info is empty")
	}

	if len(info.Id) == 0 && len(info.Host) == 0 {
		return perr.WrapSystemErrorf(nil, "id and host are both empty")
	}

	if len(info.Prefix) == 0 {
		return perr.WrapSystemErrorf(nil, "path is empty")
	}

	return nil
}
