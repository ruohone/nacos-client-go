package nacos_go_client

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/ruohone/nacos-client-go/httpproxy"
	"net/url"
	"sync"
	"sync/atomic"
	"time"
)

var serviceInfoMap = map[string]*DiscoveryServiceInfo{}
var spliter = "@@"
var newCheckSum = ""

var nextIndex = int32(0)
var discoveryOnce sync.Once

type Service struct {
	Name             string  `json:"name"`
	App              string  `json:"app"`
	Group            string  `json:"group"`
	HealthCheckMode  string  `json:"healthCheckMode"`
	ProtectThreshold float32 `json:"protectThreshold"`
}

type Instance struct {
	InstanceId string   `json:"instanceId"`
	Ip         string   `json:"ip"`
	Port       int      `json:"port"`
	Weight     float32  `json:"weight"`
	Valid      bool     `json:"valid"`
	Cluster    Cluster  `json:"cluster"`
	Service    Service  `json:"service"`
	Metadata   Metadata `json:"metadata"`
	Enabled    bool     `json:"enabled"`
}

type DiscoveryServiceInfo struct {
	Dom         string     `json:"dom"`
	Clusters    string     `json:"clusters"`
	Hosts       []Instance `json:"hosts"`
	LastRefTime int        `json:"lastRefTime"`
	Checksum    string     `json:"checksum"`
	Env         string     `json:"env"`
	AllIPs      string     `json:"allIPs"`
}

func NacosDiscovery(addr, dom, clientIp, checksum, env, clusters string) error {
	_, err := discoveryService(addr, dom, clientIp, checksum, env, clusters)
	if err != nil {
		return err
	}

	go discoveryOnce.Do(func() {
		for {
			discoveryService(addr, dom, clientIp, newCheckSum, env, clusters)
			time.Sleep(10 * time.Second)
		}
	})
	return nil
}

func discoveryService(addr, dom, clientIp, checksum, env, clusters string) (*DiscoveryServiceInfo, error) {
	values := url.Values{}
	values.Add("dom", dom)
	values.Add("clientIP", clientIp)
	values.Add("checksum", checksum)
	values.Add("env", env)
	values.Add("clusters", clusters)

	resp := httpproxy.Get(fmt.Sprintf("%s/nacos/v1/ns/api/srvIPXT?%s", addr, values.Encode()))

	str, err := resp.ToString()
	if err != nil {
		fmt.Errorf("discovery service err:%s", err)
		return nil, err
	}

	si := processService(str)
	return si, nil
}

func LoadBalanceGetIp(dsi *DiscoveryServiceInfo) (*Instance, error) {
	si := serviceInfoMap[getKey(dsi)]
	if si == nil || len(si.Hosts) <= 0 {
		return nil, errors.New("not fund service")
	}

	modulo := len(si.Hosts)

	for {
		current := nextIndex
		next := (current + 1) % int32(modulo)
		if atomic.CompareAndSwapInt32(&nextIndex, current, next) && current < int32(modulo) {
			return &si.Hosts[current], nil
		}
	}
}

func processService(j string) *DiscoveryServiceInfo {
	si := &DiscoveryServiceInfo{}
	err := json.Unmarshal([]byte(j), si)
	if err != nil {
		fmt.Errorf("process service json parse err:%s", err)
		return si
	}

	//过滤无效或已下线的host
	validHost := make([]Instance, 0)
	for _, h := range si.Hosts {
		if h.Enabled && h.Valid {
			validHost = append(validHost, h)
		}
	}
	si.Hosts = validHost

	serviceInfoMap[getKey(si)] = si
	newCheckSum = si.Checksum

	return si
}

func getKey(dsi *DiscoveryServiceInfo) string {
	return dsi.Dom + spliter + dsi.Clusters + spliter + dsi.Env + spliter + dsi.AllIPs
}
