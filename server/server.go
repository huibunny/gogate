package server

import (
	"os"
	"time"

	"github.com/hashicorp/consul/api"
	"github.com/valyala/fasthttp"
	"github.com/wanghongfei/gogate/conf"
	"github.com/wanghongfei/gogate/discovery"
	"github.com/wanghongfei/gogate/perr"
	"github.com/wanghongfei/gogate/redis"
	"github.com/wanghongfei/gogate/server/lb"
	"github.com/wanghongfei/gogate/server/route"
	stat "github.com/wanghongfei/gogate/server/statistics"
	"github.com/wanghongfei/gogate/throttle"
)

type Server struct {
	listenAddr string

	// 负载均衡组件
	lb lb.LoadBalancer

	//// 保存listener引用, 用于关闭server
	//listen 					net.Listener
	//// 是否启用优雅关闭
	//graceShutdown 			bool
	//// 优雅关闭最大等待时间
	//maxWait 				time.Duration
	//wg      				*sync.WaitGroup

	// URI路由组件
	Router *route.Router

	// 过滤器
	preFilters  []*PreFilter
	postFilters []*PostFilter

	// fasthttp对象
	fastServ   *fasthttp.Server
	fastClient *fasthttp.Client

	isStarted bool

	discoveryClient discovery.Client

	// 服务id(string) -> 此服务的限速器对象(*MemoryRateLimiter)
	rateLimiterMap *RateLimiterSyncMap

	trafficStat *stat.TraficStat
}

const (
	// 默认最大连接数
	MAX_CONNECTION = 5000
)

/*
* 创建网关服务对象
*
* PARAMS:
*	- listenAddr: http监听地址, localhost:8080
*
 */
func NewGatewayServer(cfg *conf.GateConfig, configFile, listenAddr string) (*Server, error) {
	maxConn := MAX_CONNECTION
	if cfg.ServerConfig.MaxConnection > maxConn {
		cfg.ServerConfig.MaxConnection = maxConn
	}
	router, err := route.NewRouter(cfg, configFile)
	if nil != err {
		return nil, perr.WrapSystemErrorf(err, "failed to create router")
	}

	// 创建Server对象
	serv := &Server{
		listenAddr: listenAddr,

		lb: &lb.RoundRobinLoadBalancer{},

		Router: router,

		preFilters:  make([]*PreFilter, 0, 3),
		postFilters: make([]*PostFilter, 0, 3),

		//graceShutdown: useGracefullyShutdown,
		//maxWait:       maxWait,
	}

	// 创建FastServer对象
	fastServ := &fasthttp.Server{
		Concurrency:  maxConn,
		Handler:      serv.HandleRequest,
		LogAllErrors: true,
	}
	serv.fastServ = fastServ

	// 创建http client
	serv.fastClient = &fasthttp.Client{
		MaxConnsPerHost: maxConn,
		ReadTimeout:     time.Duration(cfg.ServerConfig.Timeout) * time.Millisecond,
		WriteTimeout:    time.Duration(cfg.ServerConfig.Timeout) * time.Millisecond,
	}

	// 创建每个服务的限速器
	serv.rebuildRateLimiter(cfg)

	// 注册过虑器
	serv.AppendPreFilter(NewPreFilter("service-match-pre-filter", ServiceMatchPreFilter))
	serv.InsertPreFilterBehind("service-match-pre-filter", NewPreFilter("rate-limit-pre-filter", RateLimitPreFilter))
	serv.InsertPreFilterBehind("rate-limit-pre-filter", NewPreFilter("url-rewrite-pre-filter", UrlRewritePreFilter))

	return serv, nil

}

// 启动服务器
func (serv *Server) Start(cfg *conf.GateConfig, consulClient *api.Client) error {
	if cfg.Traffic.EnableTrafficRecord {
		serv.trafficStat = stat.NewTrafficStat(1000, 1, stat.NewCsvFileTraficInfoStore(cfg.Traffic.TrafficLogDir))
		serv.trafficStat.StartRecordTrafic()
	}

	serv.isStarted = true

	// 监听端口
	// listen, err := net.Listen("tcp", serv.listenAddr)
	// if nil != err {
	// 	return perr.WrapSystemErrorf(nil, "failed to listen at %s => %w", serv.listenAddr, err)
	// }

	// 是否启用优雅关闭功能
	//if serv.graceShutdown {
	//	serv.wg = new(sync.WaitGroup)
	//}

	// 保存Listener指针
	//serv.listen = listen

	// 初始化consul
	if consulClient != nil {
		var err error
		serv.discoveryClient, err = discovery.NewConsulClient(consulClient)
		if nil != err {
			return err
		}
	} else {
		conf.Log.Infof("no register center enabled, use static mode")
		serv.discoveryClient = discovery.DoNothingClient
	}

	// 启动注册表定时更新
	err := serv.discoveryClient.StartPeriodicalRefresh()
	if nil != err {
		return perr.WrapSystemErrorf(err, "failed to start discovery module")
	}
	// the corresponding fasthttp code
	m := func(ctx *fasthttp.RequestCtx) {
		switch string(ctx.Path()) {
		case "/healthz":
			func(ctx *fasthttp.RequestCtx) {
				conf.Log.Info("receive health check: ", ctx.Request.URI().String(), ".")
			}(ctx)
		default:
			ctx.Error("not found", fasthttp.StatusNotFound)
		}
	}

	// 启动http server
	conf.Log.Infof("start Gogate at %s, pid: %d", serv.listenAddr, os.Getpid())
	return fasthttp.ListenAndServe(serv.listenAddr, m)
}

// 关闭server
func (serv *Server) Shutdown() error {
	serv.isStarted = false
	serv.discoveryClient.UnRegister()

	err := serv.fastServ.Shutdown()
	if nil != err {
		return perr.WrapSystemErrorf(err, "failed to shutdown server")
	}

	return nil
}

// 更新路由配置文件
func (serv *Server) ReloadRoute(cfg *conf.GateConfig) error {
	conf.Log.Info("start reloading route info")
	err := serv.Router.ReloadRoute()
	serv.rebuildRateLimiter(cfg)
	conf.Log.Info("route info reloaded")

	if nil != err {
		return perr.WrapSystemErrorf(err, "failed to reload route")
	}

	return nil
}

func (serv *Server) IsInStaticMode() bool {
	return serv.discoveryClient == discovery.DoNothingClient
}

func (serv *Server) recordTraffic(servName string, success bool) {
	if nil != serv.trafficStat {
		conf.Log.Debug("log traffic for %s", servName)

		info := &stat.TraficInfo{
			ServiceId: servName,
		}
		if success {
			info.SuccessCount = 1
		} else {
			info.FailedCount = 1
		}

		serv.trafficStat.RecordTrafic(info)
	}

}

// 给路由表中的每个服务重新创建限速器;
// 在更新过route.yml配置文件时调用
func (serv *Server) rebuildRateLimiter(cfg *conf.GateConfig) {
	serv.rateLimiterMap = NewRateLimiterSyncMap()

	// 创建每个服务的限速器
	for _, info := range serv.Router.ServInfos {
		if 0 == info.Qps {
			continue
		}

		rl := serv.createRateLimiter(info, cfg)
		if nil != rl {
			serv.rateLimiterMap.Put(info.Id, rl)
			conf.Log.Debugf("done building rateLimiter for %s", info.Id)
		}
	}
}

// 创建限速器对象
// 如果配置文件中设置了使用redis, 则创建RedisRateLimiter, 否则创建MemoryRateLimiter
func (serv *Server) createRateLimiter(info *conf.ServiceInfo, cfg *conf.GateConfig) throttle.RateLimiter {
	enableRedis := cfg.RedisConfig.Enabled
	if !enableRedis {
		return throttle.NewMemoryRateLimiter(info.Qps)
	}

	client := redis.NewRedisClient(cfg.RedisConfig.Addr, 5)
	err := client.Connect()
	if nil != err {
		conf.Log.Warn("failed to create ratelimiter, err = %v", err)
		return nil
	}

	rl, err := throttle.NewRedisRateLimiter(client, cfg.RedisConfig.RateLimiterLua, info.Qps, info.Id)
	if nil != err {
		conf.Log.Warn("failed to create ratelimiter, err = %v", err)
		return nil
	}

	return rl
}
