package main

import (
	"github.com/luopengift/gohttp"
	"github.com/luopengift/golibs/file"
	"github.com/luopengift/golibs/kafka"
	"github.com/luopengift/golibs/logger"
	"io/ioutil"
	"os"
	"os/signal"
	"runtime"
	"runtime/pprof"
	"strconv"
	"time"
)

var (
	p *kafka.Producer
)

const (
	VERSION = "2017.07.06-0.0.3"
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

type FileConfig struct {
	Name   []string `json:name`
	Offset int64    `json:offset`
	Rule   string   `json:"rule"`
	Prefix string   `json:"prefix"`
	Suffix string   `json:"suffix"`
}

type HttpConfig struct {
	Addr string `json:"addr"`
}

type Config struct {
	Runtime RuntimeConfig
	Kafka   KafkaConfig
	File    FileConfig
	Http    HttpConfig
	Tags    string `json:"tags"`
	Version string `json:version`
}

func NewConfig(filename string) (*Config, error) {

	config := &Config{}
	conf := file.NewConfig("./config.json")
	err := conf.Parse(config)

	logger.Info("%+v", conf.String())
	return config, err
}

type Monitor struct {
	gohttp.HttpHandler
}

func (self *Monitor) GET() {
	self.Output(p.ChanInfo())
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
	config, err := NewConfig("./config.json")
	if err != nil {
		logger.Error("<config error> %+v", err)
		return
	}

	if VERSION != config.Version {
		logger.Error("The version of app and config is not match!exit...")
		return
	}

	if config.Runtime.MAXPROCS == 0 {
		if config.Runtime.DEBUG {
			config.Runtime.MAXPROCS = 1

		} else {
			config.Runtime.MAXPROCS = runtime.NumCPU()
		}
	}

	runtime.GOMAXPROCS(config.Runtime.MAXPROCS)

	// http server
	app := gohttp.Init()
	app.Route("^/monitor$", &Monitor{})
	go app.Run(config.Http.Addr)

	//set debug cpu model
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
	for _, v := range config.File.Name {
		logger.Warn("%v", v)

		go func(v string) {
			var f *file.Tail
			switch config.File.Rule {
			case "time":
				f = file.NewTail(v, file.TimeRule)
			default:
				f = file.NewTail(v, file.NullRule)
			}
			f.Seek(config.File.Offset)
			//开启groutine,定时刷新offset文件
			go func() {
				for {
					select {
					case <-time.After(60 * time.Second):
						str := strconv.FormatInt(f.Offset(), 10) + "\n"
						err := ioutil.WriteFile("var/"+f.BaseName()+".offset", []byte(str), 0666)
						if err != nil {
							logger.Error("<write offset file error:%+v>", err)
						}
					}
				}
			}()

			f.ReadLine()
			for value := range f.NextLine() {
				p.Write([]byte(config.File.Prefix + *value + config.File.Suffix))
				//	ret, err := p.Write([]byte(config.File.Prefix + *value + config.File.Suffix))
				//	logger.Debug("%v, %v, %v", *value, ret, err, f.Offset())
			}
			f.Stop()
		}(v)
	}
	logger.Info("APP=%s VERSION=%s, running success.", APP, VERSION)
	select {}
}
