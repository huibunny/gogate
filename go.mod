module github.com/wanghongfei/gogate

go 1.13

replace golang.org/x/net => github.com/golang/net v0.0.0-20190404232315-eb5bcb51f2a3

replace golang.org/x/text => github.com/golang/text v0.3.0

replace golang.org/x/crypto => github.com/golang/crypto v0.0.0-20190411191339-88737f569e3a

replace golang.org/x/sys => github.com/golang/sys v0.0.0-20190412213103-97732733099d

require go.uber.org/zap v1.17.0

require (
	github.com/bwmarrin/snowflake v0.3.0
	github.com/fastly/go-utils v0.0.0-20180712184237-d95a45783239 // indirect
	github.com/golang-jwt/jwt v3.2.2+incompatible // indirect
	github.com/hashicorp/consul/api v1.12.0
	github.com/huibunny/gocore v0.0.0-20220803033418-0dbfa9abb3e7
	github.com/jehiah/go-strftime v0.0.0-20171201141054-1d33003b3869 // indirect
	github.com/jonboulle/clockwork v0.1.0 // indirect
	github.com/lestrrat-go/file-rotatelogs v2.3.0+incompatible
	github.com/lestrrat-go/strftime v1.0.1 // indirect
	github.com/mediocregopher/radix.v2 v0.0.0-20181115013041-b67df6e626f9
	github.com/pborman/uuid v1.2.1 // indirect
	github.com/spf13/viper v1.12.0 // indirect
	github.com/tebeka/strftime v0.1.3 // indirect
	github.com/valyala/fasthttp v1.9.0
	github.com/wanghongfei/go-eureka-client v1.1.0
	gopkg.in/yaml.v2 v2.4.0
)
