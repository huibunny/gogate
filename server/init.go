package server

import (
	"github.com/wanghongfei/gogate/conf"
	. "github.com/wanghongfei/gogate/conf"
)

func InitGogate(gogateConfigFile string) *conf.GateConfig {
	cfg := LoadConfig(gogateConfigFile)
	InitLog(cfg)

	return cfg
}
