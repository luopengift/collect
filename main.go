/*
    collect

###
配置文件示例(./config.json)
{
    "runtime":{
        "DEBUG":true,
        "MAXPROCS":2
    },
    "kafka": {
        "addrs":[
            "10.10.20.14:9092",
            "10.10.20.15:9092",
            "10.10.20.16:9092"
        ],
        "topic":"falcon_monitor_us",
        "maxthreads":100000
    },
    "file":[
        "/data/inveno/HoneyBee/INV/honeybee/INV.honeybee_monitor_%Y%M%D.log",
        "/data/inveno/HoneyBee_report/INV/honeybee/INV.honeybee_monitor_%Y%M%D.log"
    ],
    "prefix":"",
    "suffix":"",
    "http": {
        "addr":":9143"
    },
    "tags":"collect"
}
*/

package main

import (
	"easyWork/framework/kafka"
	"encoding/json"
	"github.com/luopengift/gohttp"
	"github.com/luopengift/golibs/file"
	"github.com/luopengift/golibs/logger"
	"os"
	"runtime"
)

var (
	p *kafka.Producer
)

const (
	VERSION = "0.0.1"
	APP     = "collect"
)

type RuntimeConfig struct {
	DEBUG    bool `json:"DEBUG"`
	MAXPROCS int  `json:"MAXPROCS"`
}

type KafkaConfig struct {
	Addrs      []string `json:"addrs"`
	Topic      string   `json:"topic"`
	MaxThreads int64    `json:"maxthreads"`
}

type HttpConfig struct {
	Addr string `json:"addr"`
}

type Config struct {
	Runtime RuntimeConfig
	Kafka   KafkaConfig
	File    []string `json:"file"`
	Prefix  string   `json:"prefix"`
	Suffix  string   `json:"suffix"`
	Http    HttpConfig
	Tags    string `json:"tags"`
}

func NewConfig() (*Config, error) {
	conf, err := file.NewFile("./config.json", os.O_RDONLY).ReadAll()
	if err != nil {
		return nil, err
	}
	logger.Info("%v", string(conf))
	config := &Config{}

	err = json.Unmarshal(conf, config)
	if err != nil {
		return nil, err
	}

	if config.Runtime.MAXPROCS == 0 {
		if config.Runtime.DEBUG {
			config.Runtime.MAXPROCS = 1

		} else {
			config.Runtime.MAXPROCS = runtime.NumCPU()
		}
	}

	logger.Info("%v", config)
	return config, err
}

type Debug struct {
	gohttp.HttpHandler
}

func (self *Debug) GET() {
	self.Output(p.QueueInfo())
}

func main() {
	logger.Info("%v %v", APP, VERSION)
	//获取全局panic
	defer func() {
		if err := recover(); err != nil {
			logger.Error("Panic error:%v", err)
		}
	}()

	config, err := NewConfig()
	if err != nil {
		logger.Error("<config error> %v", err)
	}
	runtime.GOMAXPROCS(config.Runtime.MAXPROCS)

	p = kafka.NewProducer(config.Kafka.Addrs, config.Kafka.Topic, config.Kafka.MaxThreads)
	go p.WriteToTopic()

	gohttp.RouterRegister("^/debug$", &Debug{})
	go gohttp.HttpRun(&gohttp.Config{
		Addr: config.Http.Addr,
	})

	for _, v := range config.File {
		logger.Warn("%v", v)

		go func(v string) {
			f := file.NewTail(v)
			f.ReadLine()

			for value := range f.NextLine() {
				p.Write([]byte(config.Prefix + *value + config.Suffix))
				//ret, err := p.Write([]byte(config.Prefix + *value + config.Suffix))
				//logger.Info("%v,%v", ret, err)
			}
			f.Stop()
		}(v)
	}
	select {}
}
