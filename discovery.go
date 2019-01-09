package nacos_go_client

import (
	"encoding/json"
	"fmt"
	"github.com/ruohone/nacos-client-go/httpproxy"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

var serviceInfoMap = map[string]*DiscoveryServiceInfo{}
var spliter = "@@"

type Service struct {
	Name             string  `json:"name"`
	App              string  `json:"app"`
	Group            string  `json:"group"`
	HealthCheckMode  string  `json:"healthCheckMode"`
	ProtectThreshold float32 `json:"protectThreshold"`
}

type Instance struct {
	InstanceId string  `json:"instanceId"`
	Ip         string  `json:"ip"`
	Port       int     `json:"port"`
	Weight     float32 `json:"weight"`
	Valid      bool    `json:"valid"`
	Cluster    Cluster `json:"cluster"`
	Service    Service `json:"service"`
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

func DiscoveryService(addr, dom, clientIp, checksum, udpPort, env, clusters string) {
	c := httpproxy.NewClient(&http.Client{
		Timeout: time.Second * 30,
	})

	values := url.Values{}
	values.Add("dom", dom)
	values.Add("clientIP", clientIp)
	values.Add("checksum", checksum)
	values.Add("udpPort", udpPort)
	values.Add("env", env)
	values.Add("clusters", clusters)
	values.Add("allIPs", strconv.FormatBool(true))

	resp := c.Get(fmt.Sprintf("%s/nacos/v1/ns/api/srvIPXT?%s", addr, values.Encode()))

	str, err := resp.ToString()
	if err != nil {
		fmt.Println(fmt.Sprintf("discovery service err:%s", err))
	}
	fmt.Println(fmt.Sprintf(str))

	si := processService(str)
	fmt.Println(si)
}

func processService(j string) *DiscoveryServiceInfo {
	si := &DiscoveryServiceInfo{}
	err := json.Unmarshal([]byte(j), si)
	if err != nil {
		fmt.Println(fmt.Sprintf("process service json parse err:%s", err))
		return si
	}
	serviceInfoMap[getKey(si)] = si
	//
	//oldService,ok := serviceInfoMap[getKey(si)]
	//if !ok || oldService==nil{
	//	serviceInfoMap[getKey(si)] = si
	//	return nil
	//}
	//
	//if oldService.LastRefTime > si.LastRefTime {
	//	logkit.Debugf("out of date data received,old-t:%d,new-t:%d",oldService.LastRefTime,si.LastRefTime)
	//	return nil
	//}
	//
	//serviceInfoMap[getKey(si)] = si
	//
	//oldHostMap := make(map[string]*DiscoveryServiceInfo)
	//
	//for _,h:= range oldService.Hosts {
	//	oldService[]
	//}
	//
	//

	return si
}

func getKey(dsi *DiscoveryServiceInfo) string {
	return dsi.Dom + spliter + dsi.Clusters + spliter + dsi.Env + spliter + dsi.AllIPs
}
