package database

import (
	"time"

	mapset "github.com/deckarep/golang-set/v2"
	"github.com/ssrlive/proxypool/config"
	"github.com/ssrlive/proxypool/log"
	"github.com/ssrlive/proxypool/pkg/proxy"

	"gorm.io/gorm"
)

// 设置数据库字段，表名为默认为type名的复数。相比于原作者，不使用软删除特性
type Proxy struct {
	ID        uint `gorm:"primarykey"`
	CreatedAt time.Time
	UpdatedAt time.Time
	proxy.Base
	Link       string
	Identifier string `gorm:"unique"`
}

func (p *Proxy) BeforeCreate(tx *gorm.DB) (err error) {
	p.Latency = p.Latency / time.Millisecond
	return
}

func (p *Proxy) AfterFind(tx *gorm.DB) (err error) {
	p.Latency = p.Latency * time.Millisecond
	return
}

func InitTables() {
	if DB == nil {
		err := connect()
		if err != nil {
			return
		}
	}
	// Warnln: 自动迁移仅仅会创建表，缺少列和索引，并且不会改变现有列的类型或删除未使用的列以保护数据。
	// 如更改表的Column请于数据库中操作
	err := DB.AutoMigrate(&Proxy{})
	if err != nil {
		log.Errorln("\n\t\t[db/proxy.go] database migration failed")
		panic(err)
	}

	log.Infoln("current unusable node delete interval is %d day(s)", config.Config.UnusableDeleteInterval)
}

func SaveProxyList(pl proxy.ProxyList) {
	if DB == nil {
		return
	}

	DB.Transaction(func(tx *gorm.DB) error {

		last_unusable_proxy_id_set := mapset.NewSet[string]()
		DB.Model(&Proxy{}).Where("usable = ?", false).Pluck("identifier", &last_unusable_proxy_id_set)

		// Set All Usable to false
		// 这一步也把不在pl中(说明delay test时就出现问题)的且usable=true的节点的usable也改为了false，实现了对所有节点的状态更新(不在pl中的节点单单靠pl是无法完成更新的)
		if err := DB.Model(&Proxy{}).Where("useable = ?", true).Update("useable", "false").Error; err != nil {
			log.Warnln("database: Reset useable to false failed: %s", err.Error())
		}
		// Create or Update proxies
		for i := 0; i < pl.Len(); i++ {
			p := Proxy{
				Base:       *pl[i].BaseInfo(),
				Link:       pl[i].Link(),
				Identifier: pl[i].Identifier(),
			}

			p.Useable = p.Latency < proxy.MaxLantency

			// 如果上一次就是不可用， 这次还是不可用， 则不更新该proxy的信息， 以免影响更新时间戳
			if !p.Useable && last_unusable_proxy_id_set.ContainsOne(p.Identifier) {
				continue
			}

			if err := DB.Create(&p).Error; err != nil {
				// Update with Identifier
				if uperr := DB.Model(&Proxy{}).Where("identifier = ?", p.Identifier).Updates(&Proxy{
					Base: proxy.Base{Useable: p.Useable, Name: p.Name,
						Stat: proxy.Stat{Speed: p.GetSpeed(), Latency: p.GetLatency()}},
				}).Error; uperr != nil {
					log.Warnln("\n\t\tdatabase: Update failed:"+
						"\n\t\tdatabase: When Created item: %s"+
						"\n\t\tdatabase: When Updated item: %s", err.Error(), uperr.Error())
				}
			}
		}
		log.Infoln("database: Updated")
		return nil
	})
}

// Get a proxy list consists of all proxies in database
func GetAllProxies() (proxies proxy.ProxyList) {
	proxies = make(proxy.ProxyList, 0)
	if DB == nil {
		return nil
	}

	proxiesDB := make([]Proxy, 0)
	DB.Select("link").Find(&proxiesDB)

	for _, proxyDB := range proxiesDB {
		if proxiesDB != nil {
			p, err := proxy.ParseProxyFromLink(proxyDB.Link)
			if err == nil && p != nil {
				p.SetUseable(false)
				proxies = append(proxies, p)
			}
		}
	}
	return
}

// Clear proxies unusable more than 1 week
func ClearOldItems() {
	if DB == nil {
		return
	}

	lastWeek := time.Now().Add(-time.Hour * 24 * time.Duration(config.Config.UnusableDeleteInterval))

	if err := DB.Where("updated_at < ? AND useable = ?", lastWeek, false).Delete(&Proxy{}); err != nil {
		var count int64
		DB.Model(&Proxy{}).Where("updated_at < ? AND useable = ?", lastWeek, false).Count(&count)
		if count == 0 {
			log.Infoln("database: Nothing old to sweep") // TODO always this line?
		} else {
			log.Warnln("database: Delete old item failed: %s", err.Error.Error())
		}
	} else {
		log.Infoln("database: Swept old and unusable proxies")
	}
}
