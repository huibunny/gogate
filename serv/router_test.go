package serv

import (
	"fmt"
	"testing"

)

func TestLoadRoute(t *testing.T) {
	routeMap, err := loadRoute("../route.yml")
	if nil != err {
		t.Error(err)
	}

	routeMap.Range(func(name, info interface{}) bool {
		servInfo := info.(*ServiceInfo)
		fmt.Printf("path = %v, id = %s\n", servInfo.Prefix, servInfo.Id)

		return true
	})
}

func TestRouter_Match(t *testing.T) {
	r, _ := NewRouter("../route.yml")

	result := r.Match("/user")
	fmt.Println(result)
	if "user-service" != result.Id {
		t.Errorf("/user mismatch, %s\n", result)
	}

	result = r.Match("/order")
	fmt.Println(result)
	if "order-service" != result.Id {
		t.Errorf("/order mismatch, %s\n", result)
	}

	result = r.Match("/aaaa")
	if "common-service" != result.Id {
		t.Errorf("/aaaa mismatch, %s\n", result)
	}
	fmt.Println(result)

	result = r.Match("/us")
	if "common-service" != result.Id {
		t.Errorf("/us mismatch, %s\n", result)
	}
	fmt.Println(result)

	result = r.Match("/img")
	if "http://localhost:4444" != result.Host {
		t.Errorf("/img mismatch, %s\n", result)
	}
	fmt.Println(result)
}
