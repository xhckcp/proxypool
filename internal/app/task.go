package app

import (
	"fmt"
	"sync"
	"time"

	C "github.com/ssrlive/proxypool/config"
	"github.com/ssrlive/proxypool/internal/cache"
	"github.com/ssrlive/proxypool/internal/database"
	"github.com/ssrlive/proxypool/log"
	"github.com/ssrlive/proxypool/pkg/geoIp"
	"github.com/ssrlive/proxypool/pkg/healthcheck"
	"github.com/ssrlive/proxypool/pkg/provider"
	"github.com/ssrlive/proxypool/pkg/proxy"
)

func CrawlGo() {
	wg := &sync.WaitGroup{}
	var pc = make(chan proxy.Proxy)
	for _, g := range Getters {
		wg.Add(1)
		go g.Get2ChanWG(pc, wg)
	}
	proxies := cache.GetProxies("allproxies")
	dbProxies := database.GetAllProxies()
	// Show last time result when launch
	if proxies == nil && dbProxies != nil {
		cache.SetProxies("proxies", dbProxies)
		cache.LastCrawlTime = "抓取中，已载入上次数据库数据"
		log.Infoln("Database: loaded")
	}
	if dbProxies != nil {
		proxies = dbProxies.UniqAppendProxyList(proxies)
	}
	if proxies == nil {
		proxies = make(proxy.ProxyList, 0)
	}

	go func() {
		wg.Wait()
		close(pc)
	}() // Note: 为何并发？可以一边抓取一边读取而非抓完再读
	// for 用于阻塞goroutine
	for p := range pc { // Note: pc关闭后不能发送数据可以读取剩余数据
		if p != nil {
			proxies = proxies.UniqAppendProxy(p)
		}
	}

	proxies.NameClear()
	proxies = proxies.Derive()
	log.Infoln("CrawlGo unique proxy count: %d", len(proxies))

	// Clean Clash unsupported proxy because health check depends on clash
	proxies = provider.Clash{
		Base: provider.Base{
			Proxies: &proxies,
		},
	}.CleanProxies()
	log.Infoln("CrawlGo clash supported proxy count: %d", len(proxies))

	cache.SetProxies("allproxies", proxies)

	cache.UpdateProxies(proxies)

	// Health Check
	log.Infoln("Now proceed proxy health check...")
	healthcheck.SpeedConn = C.Config.SpeedConnection
	healthcheck.DelayConn = C.Config.HealthCheckConnection
	if C.Config.HealthCheckTimeout > 0 {
		healthcheck.DelayTimeout = time.Second * time.Duration(C.Config.HealthCheckTimeout)
		log.Infoln("CONF: Health check timeout is set to %d seconds", C.Config.HealthCheckTimeout)
	}

	proxies = healthcheck.CleanBadProxiesWithGrpool(proxies)

	// proxies = healthcheck.CleanBadProxies(proxies)

	log.Infoln("CrawlGo clash usable proxy count: %d", len(proxies))

	// Format name like US_01 sorted by country
	proxies.NameAddCounrty().Sort()
	log.Infoln("Proxy rename DONE!")

	// Relay check and rename
	healthcheck.RelayCheck(proxies)
	for i := range proxies {
		if s, ok := proxy.ProxyStats.Find(proxies[i]); ok {
			if s.Relay {
				_, c, e := geoIp.GeoIpDB.Find(s.OutIp)
				if e == nil {
					proxies[i].SetName(fmt.Sprintf("Relay_%s-%s", proxies[i].BaseInfo().Name, c))
				}
			} else if s.Pool {
				proxies[i].SetName(fmt.Sprintf("Pool_%s", proxies[i].BaseInfo().Name))
			}
		}
	}

	proxies.NameAddIndex()

	cache.UpdateProxies(proxies)

	cache.UsefullProxiesCount = proxies.Len()

	log.Infoln("Usablility checking done. Open %s to check", C.Config.HostUrl())

	// 并发测试所有节点的速度
	SpeedTest(proxies)

	// 测速完成之后保存测速数据
	database.SaveProxyList(proxies)
	database.ClearOldItems()

	proxies = proxies.GetUsableProxy()

	cache.UpdateProxies(proxies)
}

// Speed test for new proxies
func speedTestNew(proxies proxy.ProxyList) {
	if C.Config.SpeedTest {
		cache.IsSpeedTest = "已开启"
		if C.Config.SpeedTimeout > 0 {
			healthcheck.SpeedTimeout = time.Second * time.Duration(C.Config.SpeedTimeout)
			log.Infoln("config: Speed test timeout is set to %d seconds", C.Config.SpeedTimeout)
		}
		healthcheck.SpeedTestNew(proxies)
	} else {
		cache.IsSpeedTest = "未开启"
	}
}

// Speed test for all proxies in proxy.ProxyList
func SpeedTest(proxies proxy.ProxyList) {
	if C.Config.SpeedTest {
		cache.IsSpeedTest = "已开启"
		if C.Config.SpeedTimeout > 0 {
			log.Infoln("config: Speed test timeout is set to %d seconds", C.Config.SpeedTimeout)
			healthcheck.SpeedTimeout = time.Second * time.Duration(C.Config.SpeedTimeout)
		}
		healthcheck.SpeedTestAll(proxies)
	} else {
		cache.IsSpeedTest = "未开启"
	}
}
