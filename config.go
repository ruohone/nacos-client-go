package nacos_go_client

import (
	"bytes"
	"crypto/md5"
	"fmt"
	"github.com/BurntSushi/toml"
	"github.com/ruohone/nacos-client-go/httpproxy"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"
)

var once sync.Once
var contentMd5 = ""

func NacosConfigRegister(addr, dataId, group string, conf interface{}) error {
	err := nacosUpdateConfig(addr, dataId, group, conf)
	go once.Do(func() {
		ch := make(chan int, 1)
		nacosConfigListen(addr, dataId, group, conf, ch)
		for {
			select {
			case v := <-ch:
				if v == 1 {
					nacosConfigListen(addr, dataId, group, conf, ch)
				}
				if v == 0 { //异常延迟5秒请求
					time.Sleep(time.Second * 5)
					nacosConfigListen(addr, dataId, group, conf, ch)
				}

			}
		}
	})
	return err
}

func nacosConfigListen(addr, dataId, group string, conf interface{}, ch chan int) {
	content := bytes.Buffer{}
	content.WriteString(dataId)
	content.WriteString("%02")
	content.WriteString(group)
	content.WriteString("%02")
	content.WriteString(contentMd5)
	content.WriteString("%01")

	v := url.Values{}
	v.Add("Listening-Configs", content.String())

	c := httpproxy.NewClient(&http.Client{
		Timeout: time.Second * 30,
	})

	reqCli := c.NewRequest()
	reqCli = reqCli.WithHeader("Long-Pulling-Timeout", "30000")
	reqCli = reqCli.WithHeader("exConfigInfo", "true")
	reqCli = reqCli.WithHeader("Content-Type", "application/x-www-form-urlencoded;charset=utf-8")

	reqCli = reqCli.WithFormBody(v)

	reqCli = reqCli.Post(fmt.Sprintf("%s/nacos/v1/cs/configs/listener", addr))
	resp := reqCli.Execute()
	str, err := resp.ToString()
	if err != nil {
		fmt.Println(fmt.Sprintf("NacosConfigRegister resp str:%s err:%s", str, err))
		ch <- 0
		return
	}

	if str == "" {
		fmt.Println(fmt.Sprintf("NacosConfigRegister resp str:%s err: str is nil", str))
		ch <- 1
		return
	}

	str = strings.Split(str, "%01")[0]
	strArr := strings.Split(str, "%02")
	if len(strArr) < 2 {
		fmt.Println(fmt.Sprintf("NacosConfigRegister response format err: str = %s", str))
		ch <- 1
		return
	}

	if strArr[0] == dataId && strArr[1] == group {
		nacosUpdateConfig(addr, dataId, group, conf)
		ch <- 1
		return
	}
}

func nacosUpdateConfig(addr, dataId, group string, conf interface{}) error {
	c := httpproxy.NewClient(&http.Client{
		Timeout: 30 * time.Second,
	})

	url := fmt.Sprintf("%s/nacos/v1/cs/configs?dataId=%s&group=%s", addr, dataId, group)
	resp := c.Get(url)

	str, err := resp.ToString()
	if err != nil {
		fmt.Println(fmt.Sprintf("NacosConfigRegister url:%s err:%s", url, err))
		return err
	}

	if str == "" {
		fmt.Println(fmt.Sprintf("NacosConfigRegister url:%s err:str is nil", url))
		return nil
	}

	_, err = toml.Decode(str, conf)
	if err != nil {
		fmt.Println(fmt.Sprintf("NacosConfigRegister url:%s err:%s", url, err))
		return nil
	}

	w := md5.New()
	io.WriteString(w, str)
	contentMd5 = fmt.Sprintf("%x", w.Sum(nil))
	return nil
}

func UpdateConfig(addr, dataId, group string, conf interface{}) error {
	bf := bytes.Buffer{}
	e := toml.NewEncoder(&bf)
	err := e.Encode(conf)
	if err != nil {
		return err
	}

	cli := httpproxy.NewClient(&http.Client{
		Timeout: 30 * time.Second,
	})

	v := url.Values{}
	v.Set("dataId", dataId)
	v.Set("group", group)
	v.Set("content", bf.String())

	res := cli.PostForm(fmt.Sprintf("%s/nacos/v1/cs/configs", addr), v)
	return res.Error()
}
