package discovery

import (
	"strconv"

	"github.com/hashicorp/consul/api"
	"github.com/wanghongfei/gogate/conf"
	. "github.com/wanghongfei/gogate/conf"
	"github.com/wanghongfei/gogate/perr"
)

type ConsulClient struct {
	// 继承方法
	*periodicalRefreshClient

	client *api.Client

	// 保存服务地址
	// key: 服务名:版本号, 版本号为eureka注册信息中的metadata[version]值
	// val: []*InstanceInfo
	registryMap *InsInfoArrSyncMap
}

func NewConsulClient() (Client, error) {
	cfg := &api.Config{}
	cfg.Address = conf.App.ConsulConfig.Address
	cfg.Scheme = "http"

	c, err := api.NewClient(cfg)
	if nil != err {
		return nil, perr.WrapSystemErrorf(err, "failed to init consule client")
	}

	consulClient := &ConsulClient{client: c}
	consulClient.periodicalRefreshClient = newPeriodicalRefresh(consulClient)

	return consulClient, nil
}

func (c *ConsulClient) GetInternalRegistryStore() *InsInfoArrSyncMap {
	return c.registryMap
}

func (c *ConsulClient) SetInternalRegistryStore(registry *InsInfoArrSyncMap) {
	c.registryMap = registry
}

func (c *ConsulClient) Get(serviceId string) []*InstanceInfo {
	instance, exist := c.registryMap.Get(serviceId)
	if !exist {
		return nil
	}

	return instance
}

func (c *ConsulClient) QueryServices() ([]*InstanceInfo, error) {
	servMap, err := c.client.Agent().Services()
	if nil != err {
		return nil, err
	}

	// 查出所有健康实例
	healthList, _, err := c.client.Health().State(api.HealthPassing, nil)
	if nil != err {
		return nil, perr.WrapSystemErrorf(err, "failed to query consul")
	}

	instances := make([]*InstanceInfo, 0, 10)
	for _, servInfo := range servMap {
		servName := servInfo.Service
		servId := servInfo.ID

		// 查查在healthList中有没有
		isHealth := false
		for _, healthInfo := range healthList {
			if healthInfo.ServiceName == servName && healthInfo.ServiceID == servId {
				isHealth = true
				break
			}
		}

		if !isHealth {
			Log.Warn("following instance is not health, skip; service name: %v, service id: %v", servName, servId)
			continue
		}

		instances = append(
			instances,
			&InstanceInfo{
				ServiceName: servInfo.Service,
				Addr:        servInfo.Address + ":" + strconv.Itoa(servInfo.Port),
				Meta:        servInfo.Meta,
			},
		)
	}

	return instances, nil
}

func (c *ConsulClient) Register() error {
	return perr.WrapSystemErrorf(nil, "not implement yet")
}

func (c *ConsulClient) UnRegister() error {
	return perr.WrapSystemErrorf(nil, "not implement yet")
}
