package conf

import (
	"errors"
	"io/ioutil"
	"os"

	"gopkg.in/yaml.v2"
)

type GateConfig struct {
	Version string `yaml:"version"`

	ServerConfig *ServerConfig `yaml:"server"`

	ConsulConfig *ConsulConfig `yaml:"consul"`

	RedisConfig *RedisConfig `yaml:"redis"`

	EurekaConfig *EurekaConfig `yaml:"eureka"`

	Traffic *TrafficConfig `yaml:"traffic"`

	Log struct {
		ConsoleOnly bool   `yaml:"console-only"`
		FilePattern string `yaml:"file-pattern"`
		FileLink    string `yaml:"file-link"`
		Directory   string `yaml:"directory"`
	} `yaml:"log"`

	Router []*ServiceInfo `yaml:"router"`
}

type ServerConfig struct {
	AppName       string `yaml:"appName"`
	MaxConnection int    `yaml:"maxConnection"`
	// 请求超时时间, ms
	Timeout int `yaml:"timeout"`
	// token secret
	Secret string `yaml:"secret"`
}

// Consul -.
type ConsulConfig struct {
	CheckApi string `yaml:"checkapi"`
	Interval string `yaml:"interval"`
	Timeout  string `yaml:"timeout"`
}

type EurekaConfig struct {
	Enable            bool   `yaml:"enable"`
	ConfigFile        string `yaml:"configFile"`
	EvictionDuration  uint   `yaml:"evictionDuration"`
	HeartbeatInterval int    `yaml:"heartbeatInterval"`
}

type TrafficConfig struct {
	EnableTrafficRecord bool   `yaml:"enableTrafficRecord"`
	TrafficLogDir       string `yaml:"trafficLogDir"`
}

type RedisConfig struct {
	Enabled        bool
	Addr           string
	RateLimiterLua string `yaml:"rateLimiterLua"`
}

type ServiceInfo struct {
	Id          string
	Prefix      string
	Host        string
	Name        string
	StripPrefix bool `yaml:"strip-prefix"`
	Qps         int
	Verify      bool `yaml:"verify"`

	Canary []*CanaryInfo
}

func (info *ServiceInfo) String() string {
	return "prefix = " + info.Prefix + ", id = " + info.Id + ", host = " + info.Host
}

type CanaryInfo struct {
	Meta   string
	Weight int
}

func LoadConfig(filename string) *GateConfig {
	f, err := os.Open(filename)
	if nil != err {
		Log.Error(err)
		panic(err)
	}
	defer f.Close()

	buf, _ := ioutil.ReadAll(f)

	config := new(GateConfig)
	err = yaml.Unmarshal(buf, config)
	if nil != err {
		Log.Error(err)
		panic(err)
	}

	// validateGogateConfig(config)

	return config
}

func InitLog(cfg *GateConfig) {
	initRotateLog(cfg)
}

func ValidateGogateConfig(config *GateConfig) error {
	if nil == config {
		return errors.New("config is nil")
	}

	servCfg := config.ServerConfig
	if servCfg.AppName == "" {
		servCfg.AppName = "gogate"
	}

	if servCfg.MaxConnection == 0 {
		servCfg.MaxConnection = 1000
	}

	if servCfg.Timeout == 0 {
		servCfg.Timeout = 3000
	}

	trafficCfg := config.Traffic
	if trafficCfg.EnableTrafficRecord {
		if trafficCfg.TrafficLogDir == "" {
			trafficCfg.TrafficLogDir = "/tmp"
		}
	}

	rdConfig := config.RedisConfig
	if rdConfig.Enabled {
		if rdConfig.Addr == "" {
			rdConfig.Addr = "127.0.0.1:6379"
		}
	}

	return nil
}
