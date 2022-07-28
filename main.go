package main

import (
	"flag"
	"os"

	consulutil "github.com/huibunny/gocore/thirdpart/consul"
	"github.com/huibunny/gocore/utils"
	"github.com/wanghongfei/gogate/conf"
	. "github.com/wanghongfei/gogate/conf"
	"github.com/wanghongfei/gogate/perr"
	serv "github.com/wanghongfei/gogate/server"
)

func main() {
	// config args, priority: config > consul
	var (
		configFile     = flag.String("config", "", "config file, prior to use.")
		consulAddr     = flag.String("consul", "localhost:8500", "consul server address.")
		consulInterval = flag.String("interval", "3", "consul health check interval, seconds.")
		consulTimeout  = flag.String("timeout", "3", "consul health check timeout, seconds.")
		consulFolder   = flag.String("folder", "", "consul kv folder.")
		serviceName    = flag.String("name", "apigateway", "both microservice name and kv name.")
		listenAddr     = flag.String("listen", ":8080", "listen address.")
	)
	flag.Parse()
	host, port := utils.GetHostPort(*listenAddr)
	// Configuration
	cfg := &GateConfig{}
	var err error
	if len(*configFile) > 0 {
		cfg = serv.InitGogate(*configFile)
	} else if len(*consulAddr) > 0 {
		consulClient, serviceID, err := consulutil.RegisterAndCfgConsul(cfg, *consulAddr, *serviceName, host, port,
			*consulInterval, *consulTimeout, *consulFolder)
		if err != nil {
			Log.Fatalf("fail to register consul: %v.", err)
		}
		defer consulutil.DeregisterService(consulClient, serviceID)
	} else {
		Log.Fatalf("no input: config file or consul address not provided!")
		return
	}
	// 初始化
	conf.ValidateGogateConfig(cfg)

	// 构造gogate对象
	server, err := serv.NewGatewayServer(
		App.ServerConfig.Host,
		App.ServerConfig.Port,
		App.EurekaConfig.RouteFile,
		App.ServerConfig.MaxConnection,
	)
	checkErrorExit(err, true)

	Log.Infof("pre filters: %v", server.ExportAllPreFilters())
	Log.Infof("post filters: %v", server.ExportAllPostFilters())

	// 启动服务器
	err = server.Start()
	checkErrorExit(err, true)
	Log.Info("listener has been closed")

	// 等待优雅关闭
	err = server.Shutdown()
	checkErrorExit(err, false)
}

func checkErrorExit(err error, exit bool) {
	if nil != err {
		Log.Error(perr.EnvMsg(err))

		if exit {
			os.Exit(1)
		}
	}
}
