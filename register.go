package nacos_go_client

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/ruohone/nacos-client-go/httpproxy"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"sync"
	"time"
)

type Cluster struct {
	DefaultCheckPort int           `json:"defaultCheckPort"`
	DefaultPort      int           `json:"defaultPort"`
	HealthChecker    HealthChecker `json:"health_checker"`
	Metadata         Metadata      `json:"metadata"`
	Name             string        `json:"name"`
	UseIPPort4Check  bool          `json:"useIPPort4Check"`
}

type HealthChecker struct {
	Type string `json:"type"`
}

type Metadata struct {
	Name1 string `json:"name1"`
}

type ServiceInfo struct {
	Cluster     Cluster  `json:"cluster"`
	Metadata    Metadata `json:"metadata"`
	Port        string   `json:"port"`
	Enable      string   `json:"enable"`
	Healthy     string   `json:"healthy"`
	Ip          string   `json:"ip"`
	Weight      string   `json:"weight"`
	ServiceName string   `json:"serviceName"`
	Tenant      string   `json:"tenant"`
}

type BeatInfo struct {
	Cluster string `json:"cluster"`
	Dom     string `json:"dom"`
	Ip      string `json:"ip"`
	Port    int    `json:"port"`
}

var registerOnce sync.Once
var stop = false

func NacosServiceRegister(serInfo ServiceInfo, addr string) error {
	err := registerService(serInfo, addr)
	if err != nil {
		return err
	}

	go registerOnce.Do(func() {
		bi := BeatInfo{}
		bi.Dom = serInfo.ServiceName
		bi.Cluster = "DEFAULT"
		bi.Ip = serInfo.Ip
		bi.Port, _ = strconv.Atoi(serInfo.Port)
		for {
			if stop {
				break
			}
			clientBeat(bi, addr)
			time.Sleep(10 * time.Second)
		}
	})

	return nil
}

func registerService(serInfo ServiceInfo, addr string) error {
	cluster, err := json.Marshal(serInfo.Cluster)
	if err != nil {
		return err
	}

	metadata, err := json.Marshal(serInfo.Metadata)
	if err != nil {
		return err
	}

	if serInfo.Tenant == "" {
		serInfo.Tenant = "default"
	}

	valus := url.Values{}
	valus.Add("cluster", string(cluster))
	valus.Add("metadata", string(metadata))
	valus.Add("port", serInfo.Port)
	valus.Add("enable", serInfo.Enable)
	valus.Add("healthy", serInfo.Healthy)
	valus.Add("ip", serInfo.Ip)
	valus.Add("weight", serInfo.Weight)
	valus.Add("serviceName", serInfo.ServiceName)
	valus.Add("tenant", serInfo.Tenant)

	b := &bytes.Buffer{}
	b.WriteString(valus.Encode())
	r, err := http.NewRequest("PUT", fmt.Sprintf("%s/nacos/v1/ns/instance", addr), b)
	if err != nil {
		return err
	}
	r.Header.Add("Content-Type", "application/x-www-form-urlencoded;charset=UTF-8")

	client := http.Client{}
	res, err := client.Do(r)
	if err != nil {
		return err
	}

	if res != nil {
		defer res.Body.Close()
	}

	if res.StatusCode != 200 {
		message, _ := ioutil.ReadAll(res.Body)
		return fmt.Errorf("Status:%v,Message:%s", res.Status, message)
	}
	data, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return err
	}

	if "ok" != string(data) {
		return errors.New("注册服务失败")
	}

	return nil
}

func clientBeat(info BeatInfo, addr string) (int, error) {
	beat, err := json.Marshal(info)
	if err != nil {
		return 0, err
	}

	values := url.Values{}
	values.Add("beat", string(beat))
	values.Add("encoding", "UTF-8")
	values.Add("dom", info.Dom)

	r := httpproxy.NewRequest()
	r.WithHeader("Content-Type", "application/x-www-form-urlencoded;charset=UTF-8")
	req := r.Get(fmt.Sprintf("%s/nacos/v1/ns/api/clientBeat?%s", addr, values.Encode()))
	resp := req.Execute()

	str, err := resp.ToString()
	if err != nil {
		return 0, err
	}

	cbi := make(map[string]int, 1)

	err = json.Unmarshal([]byte(str), &cbi)
	if err != nil {
		return 0, err
	}

	return cbi["clientBeatInterval"], nil
}

func DelRegister(serviceName, ip, cluster string, port int, addr string) error {
	valus := url.Values{}
	valus.Add("cluster", cluster)
	valus.Add("port", strconv.Itoa(port))
	valus.Add("ip", ip)
	valus.Add("serviceName", serviceName)
	valus.Add("encoding", "utf-8")

	r, err := http.NewRequest("DELETE", fmt.Sprintf("%s/nacos/v1/ns/instance?%s", addr, valus.Encode()), nil)
	if err != nil {
		return err
	}
	r.Header.Add("Content-Type", "application/x-www-form-urlencoded;charset=UTF-8")

	client := http.Client{}
	res, err := client.Do(r)
	if err != nil {
		return err
	}

	if res != nil {
		defer res.Body.Close()
	}

	if res.StatusCode != 200 {
		message, _ := ioutil.ReadAll(res.Body)
		return fmt.Errorf("Status:%v,Message:%s", res.Status, message)
	}
	data, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return err
	}

	if "ok" != string(data) {
		return errors.New("删除服务失败")
	}

	stop = true

	return nil
}
