package nacos_go_client

import (
	"code.jiecaojingxuan.com/gobase/httpclient"
	"code.jiecaojingxuan.com/gobase/logkit"
	"crypto/md5"
	"fmt"
	"github.com/BurntSushi/toml"
	"io"
	"net/http"
	"testing"
	"time"
)

func TestConf(t *testing.T) {
	//c := &config{}
	//
	//NacosConfigRegister("http://10.210.29.210:8848", "test.toml", "DEFAULT_GROUP", &c)
	//
	//c.Taskurl = "hehrerer"
	//
	//err := UpdateConfig("http://demo.nacos.qiangshouapp.com:8848", "testgg", "DEFAULT_GROUP", c)
	//
	//logkit.Error(err.Error())
	//
	//logkit.Infof("%s",c.Taskurl)

	metadata := Metadata{}
	metadata.Name = "ag"
	hc := HealthChecker{}
	hc.Type = "TCP"

	clu := Cluster{}
	clu.Metadata = metadata
	clu.HealthChecker = hc
	clu.Name = "DEFAULT"
	clu.UseIPPort4Check = true
	clu.DefaultCheckPort = 80
	clu.DefaultPort = 80

	serInfo := ServiceInfo{}
	serInfo.Metadata = metadata
	serInfo.Cluster = clu
	serInfo.Port = "18081"
	serInfo.Enable = "true"
	serInfo.Healthy = "true"
	serInfo.Ip = "172.16.187.60"
	serInfo.Weight = "1.0"
	serInfo.ServiceName = "service-provider-go"
	serInfo.Tenant = "ty"

	//err := NacosServiceRegister(serInfo, "http://localhost:8848")
	//if err != nil {
	//	logkit.Error(err.Error())
	//}

	DiscoveryService("http://localhost:8848", "service-provider-go", serInfo.Ip, "", "0", "", "")

	//err := RegisterService(serInfo,"http://localhost:8848")
	//if err!= nil {
	//	logkit.Error(err.Error())
	//}
	//
	//bi := BeatInfo{}
	//bi.Dom = "service-provider-go"
	//bi.Cluster = "DEFAULT"
	//bi.Ip="172.16.187.60"
	//bi.Port = 18081

	//err = DelRegister("service-provider-go",serInfo.Ip,"DEFAULT",18081,"http://localhost:8848")
	//if err!= nil {
	//	logkit.Error(err.Error())
	//}
	//
	//go func() {
	//	for {
	//		err = ClientBeat(bi,"http://localhost:8848")
	//		if err!= nil {
	//			logkit.Error(err.Error())
	//		}
	//
	//		time.Sleep(time.Second)
	//	}
	//}()

	//time.Sleep(time.Second*70)
	//
	////err = ClientBeat(bi,"http://localhost:8848")
	////if err!= nil {
	////	logkit.Error(err.Error())
	////}
	//
	//time.Sleep(time.Second)
	//
	//time.Sleep(time.Hour)
	//r := gin.New()

}

type config struct {
	Etcd        string    `toml:"etcd"`
	Redis       Redis     `toml:"redis"`
	Mysql       Mysql     `toml:"mysql"`
	HttpPort    int       `toml:"http_port"`
	Achieve     Achieve   `toml:"achieve"`
	Product     Product   `toml:"product"`
	AddLifeUrl  string    `toml:"add_life_url"`
	PushMsgUrl  string    `toml:"push_url"`
	Taskurl     string    `toml:"task_url"` //任务首页
	AppURI      string    `toml:"app_url"`  // app首页
	AddMoneyUrl string    `toml:"add_money_url"`
	RedisLive   RedisConf `toml:"redis_live_dur"`
}
type RedisConf struct {
	Addr     string `toml:"addr"`
	Password string `toml:"password"`
}

type Redis struct {
	Achieve string `toml:"achieve"`
}

type Mysql struct {
	Achieve string `toml:"achieve"`
}

type Achieve struct {
	SupplyCoin   int    `toml:"supply_coin"`
	SupplyMaxDay int    `toml:"supply_max_day"` //最大补签天数
	SignHomeLink string `toma:"sign_home_link"`
}

type Product struct {
	TemplatePath string `toml:"path"`
	TemplateName string `toml:"name"`
}

func ttt(addr, dataId, group string, conf interface{}) error {
	c := httpclient.NewClient(&http.Client{
		Timeout: 30 * time.Second,
	})

	url := fmt.Sprintf("%s/nacos/v1/cs/configs?dataId=%s&group=%s", addr, dataId, group)
	resp := c.Get(url)

	str, err := resp.ToString()
	if err != nil {
		logkit.Errorf("NacosConfigRegister url:%s err:%s", url, err)
		return err
	}

	if str == "" {
		logkit.Errorf("NacosConfigRegister url:%s err:str is nil", url)
		return nil
	}

	_, err = toml.Decode(str, conf)
	if err != nil {
		logkit.Errorf("NacosConfigRegister url:%s err:%s", url, err)
		return nil
	}

	w := md5.New()
	io.WriteString(w, str)
	contentMd5 = fmt.Sprintf("%x", w.Sum(nil))
	return nil
}
