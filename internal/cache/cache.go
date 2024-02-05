package cache

import (
	"time"

	"github.com/patrickmn/go-cache"
	"github.com/ssrlive/proxypool/log"
	"github.com/ssrlive/proxypool/pkg/provider"
	"github.com/ssrlive/proxypool/pkg/proxy"
)

var location, _ = time.LoadLocation("Asia/Shanghai")
var c = cache.New(cache.NoExpiration, 10*time.Minute)

func GetProxies(key string) proxy.ProxyList {
	result, found := c.Get(key)
	if found {
		return result.(proxy.ProxyList) //Get返回的是interface
	}
	return nil
}

func SetProxies(key string, proxies proxy.ProxyList) {
	c.Set(key, proxies, cache.NoExpiration)
}

func SetString(key, value string) {
	c.Set(key, value, cache.NoExpiration)
}

func GetString(key string) string {
	result, found := c.Get(key)
	if found {
		return result.(string)
	}
	return ""
}

func UpdateProxies(proxies proxy.ProxyList) {
	SetProxies("proxies", proxies)

	AllProxiesCount = proxies.Len()
	log.Infoln("AllProxiesCount: %d", AllProxiesCount)
	SSProxiesCount = proxies.TypeLen("ss")
	log.Infoln("SSProxiesCount: %d", SSProxiesCount)
	SSRProxiesCount = proxies.TypeLen("ssr")
	log.Infoln("SSRProxiesCount: %d", SSRProxiesCount)
	VmessProxiesCount = proxies.TypeLen("vmess")
	log.Infoln("VmessProxiesCount: %d", VmessProxiesCount)
	TrojanProxiesCount = proxies.TypeLen("trojan")
	log.Infoln("TrojanProxiesCount: %d", TrojanProxiesCount)
	LastCrawlTime = time.Now().In(location).Format("2006-01-02 15:04:05")

	SetString("clashproxies", provider.Clash{
		Base: provider.Base{
			Proxies: &proxies,
		},
	}.Provide()) // update static string provider

	SetString("surgeproxies", provider.Surge{
		Base: provider.Base{
			Proxies: &proxies,
		},
	}.Provide())

}
