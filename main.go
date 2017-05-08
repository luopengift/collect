package main

import (
	"encoding/json"
	"github.com/luopengift/gohttp"
	"github.com/luopengift/golibs/file"
	"github.com/luopengift/golibs/kafka"
	"github.com/luopengift/golibs/logger"
	"os"
	"os/signal"
	"runtime"
	"runtime/pprof"
)

var (
	p *kafka.Producer
)

const (
	VERSION = "0.0.2"
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

	//init config
	config, err := NewConfig()
	if err != nil {
		logger.Error("<config error> %v", err)
	}
	runtime.GOMAXPROCS(config.Runtime.MAXPROCS)

	// http server
	gohttp.RouterRegister("^/debug$", &Debug{})
	go gohttp.HttpRun(&gohttp.Config{
		Addr: config.Http.Addr,
	})

	//set debug model
	if config.Runtime.DEBUG {

		//pprof file
		s := make(chan os.Signal, 1)
		signal.Notify(s, os.Interrupt, os.Kill)
		f, err := os.OpenFile("./cpu.prof", os.O_RDWR|os.O_CREATE, 0644)
		if err != nil {
			logger.Error("<file open error> %v", err)
		}
		defer f.Close()
		pprof.StartCPUProfile(f)
		defer pprof.StopCPUProfile()

		go func() {
			select {
			case sign := <-s:
				logger.Error("Got signal:", sign)
				pprof.StopCPUProfile()
				f.Close()
				os.Exit(-1)
			}
		}()
	}

	//init kafka producer
	p = kafka.NewProducer(config.Kafka.Addrs, config.Kafka.Topic, config.Kafka.MaxThreads)
	go p.WriteToTopic()

	//main loop
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
