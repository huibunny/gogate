package main

import (
	"flag"
	"os"
	"os/signal"
	"syscall"

	"github.com/hashicorp/consul/api"
	consulutil "github.com/huibunny/gocore/thirdpart/consul"
	"github.com/wanghongfei/gogate/conf"
	"github.com/wanghongfei/gogate/perr"
	serv "github.com/wanghongfei/gogate/server"
)

func main() {
	// config args, priority: config > consul
	var (
		configFile   = flag.String("config", "", "config file, prior to use.")
		consulAddr   = flag.String("consul", "localhost:8500", "consul server address.")
		consulFolder = flag.String("folder", "", "consul kv folder.")
		serviceName  = flag.String("name", "apigateway", "both microservice name and kv name.")
		listenAddr   = flag.String("listen", ":8080", "listen address.")
	)
	flag.Parse()
	// Configuration
	cfg := &conf.GateConfig{}
	var serviceID string
	var consulClient *api.Client
	var err error
	if len(*configFile) > 0 {
		cfg = serv.InitGogate(*configFile)
	} else if len(*consulAddr) > 0 {
		consulClient, serviceID, _, err = consulutil.RegisterAndCfgConsul(cfg, *consulAddr, *serviceName, *listenAddr, *consulFolder)
		if err != nil {
			return
		}
		conf.InitLog(cfg)
		defer consulutil.DeregisterService(consulClient, serviceID)
	} else {
		conf.Log.Fatalf("no input: config file or consul address not provided!")
		return
	}
	// 初始化
	conf.ValidateGogateConfig(cfg)

	// 构造gogate对象
	server, err := serv.NewGatewayServer(cfg, *configFile, *listenAddr)
	checkErrorExit(err, true)

	conf.Log.Infof("pre filters: %v", server.ExportAllPreFilters())
	conf.Log.Infof("post filters: %v", server.ExportAllPostFilters())

	go func() {
		server.Start(cfg, consulClient)
	}()

	// Waiting signal
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt, syscall.SIGTERM)

	s := <-interrupt
	conf.Log.Info("app - Run - signal: " + s.String())
	err = server.Shutdown()
	checkErrorExit(err, true)
	conf.Log.Info("listener has been closed")

}

func checkErrorExit(err error, exit bool) {
	if nil != err {
		conf.Log.Error(perr.EnvMsg(err))

		if exit {
			os.Exit(1)
		}
	}
}
