package proxy

import (
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/ssrlive/proxypool/pkg/geoIp"
)

type LantencyDataType = time.Duration

var MaxLantency LantencyDataType = math.MaxUint16 * time.Millisecond

/* Base implements interface Proxy. It's the basic proxy struct. Vmess etc extends Base*/
type Base struct {
	Name    string `yaml:"name" json:"name" gorm:"index"`
	Server  string `yaml:"server" json:"server" gorm:"index"`
	Type    string `yaml:"type" json:"type" gorm:"index"`
	Country string `yaml:"country,omitempty" json:"country,omitempty" gorm:"index"`
	Port    int    `yaml:"port" json:"port" gorm:"index"`
	UDP     bool   `yaml:"udp,omitempty" json:"udp,omitempty"`
	Usable  bool   `yaml:"useable,omitempty" json:"usable,omitempty" gorm:"index"`
	// Speed   float64          `yaml:"speed,omitempty" json:"speed,omitempty" gorm:"index"`
	// Latency LantencyDataType `yaml:"lantency,omitempty" json:"lantency,omitempty" gorm:"index"`
	Stat
}

// TypeName() Get specific proxy type
func (b *Base) TypeName() string {
	if b.Type == "" {
		return "unknown"
	}
	return b.Type
}

// SetName() to a proxy
func (b *Base) SetName(name string) {
	b.Name = name
}

func (b *Base) AddToName(name string) {
	b.Name = b.Name + name
}

func (b *Base) AddBeforeName(name string) {
	b.Name = name + b.Name
}

// SetIP() to a proxy
func (b *Base) SetIP(ip string) {
	b.Server = ip
}

// BaseInfo() get basic info struct of a proxy
func (b *Base) BaseInfo() *Base {
	return b
}

// Clone() returns a new basic proxy
func (b *Base) Clone() Base {
	c := *b
	return c
}

// SetUsable() set Base info "Usable" (true or false)
func (b *Base) SetUsable(useable bool) {
	b.Usable = useable
}

// SetUsable() set Base info "Country" (string)
func (b *Base) SetCountry(country string) {
	b.Country = country
}

func (b *Base) IsUsable() bool {
	return b.Usable
}

func (b *Base) UpdateUsable() {
	b.SetUsable(b.GetLatency() < MaxLantency && b.GetSpeed() > 0)
}

func (b *Base) UpdateLatency(latency time.Duration) {
	b.Stat.UpdateLatency(latency)
	b.UpdateUsable()
}

func (b *Base) UpdateSpeed(speed float64) {

	b.Stat.UpdateSpeed(speed)
	b.UpdateUsable()

	if speed > 0 {

		res := strings.Split(b.Name, "|")

		if len(res) > 1 {
			b.Name = res[0]
		}

		b.AddToName(fmt.Sprintf("|%.2fMb/s", speed))
	}
}

type Proxy interface {
	String() string
	ToClash() string
	ToSurge() string
	Link() string
	Identifier() string
	SetName(name string)
	AddToName(name string)
	SetIP(ip string)
	TypeName() string //ss ssr vmess trojan
	BaseInfo() *Base
	Clone() Proxy
	IsUsable() bool
	SetUsable(useable bool)
	SetCountry(country string)
	GetSpeed() float64
	GetLatency() time.Duration
	UpdateSpeed(speed float64)
	UpdateLatency(lantency time.Duration)
}

func ParseProxyFromLink(link string) (p Proxy, err error) {
	if strings.HasPrefix(link, "ssr://") {
		p, err = ParseSSRLink(link)
	} else if strings.HasPrefix(link, "vmess://") {
		p, err = ParseVmessLink(link)
	} else if strings.HasPrefix(link, "ss://") {
		p, err = ParseSSLink(link)
	} else if strings.HasPrefix(link, "trojan://") {
		p, err = ParseTrojanLink(link)
	}
	if err != nil || p == nil {
		return nil, errors.New("link parse failed")
	}
	_, country, err := geoIp.GeoIpDB.Find(p.BaseInfo().Server) // IP库不准
	if err != nil {
		country = "🏁 ZZ"
	}
	p.SetCountry(country)
	// trojan依赖域名？<-这是啥?不管什么情况感觉都不应该替换域名为IP（主要是IP库的质量和节点质量不该挂钩）
	//if p.TypeName() != "trojan" {
	//	p.SetIP(ip)
	//}
	return
}

func ParseProxyFromClashProxy(p map[string]interface{}) (proxy Proxy, err error) {
	p["name"] = ""
	pjson, err := json.Marshal(p)
	if err != nil {
		return nil, err
	}
	switch p["type"].(string) {
	case "ss":
		var proxy Shadowsocks
		err := json.Unmarshal(pjson, &proxy)
		if err != nil {
			return nil, err
		}
		return &proxy, nil
	case "ssr":
		var proxy ShadowsocksR
		err := json.Unmarshal(pjson, &proxy)
		if err != nil {
			return nil, err
		}
		return &proxy, nil
	case "vmess":
		var proxy Vmess
		err := json.Unmarshal(pjson, &proxy)
		if err != nil {
			return nil, err
		}
		return &proxy, nil
	case "trojan":
		var proxy Trojan
		err := json.Unmarshal(pjson, &proxy)
		if err != nil {
			return nil, err
		}
		return &proxy, nil
	}
	return nil, errors.New("clash json parse failed")
}

func GoodNodeThatClashUnsupported(b Proxy) bool {
	switch b.TypeName() {
	case "ss":
		ss := b.(*Shadowsocks)
		if ss == nil {
			return false
		}
		if ss.Cipher == "none" {
			return true
		} else {
			return false
		}
	case "ssr":
		ssr := b.(*ShadowsocksR)
		if ssr == nil {
			return false
		}
		return false
	}
	return false
}
