### collect

####读取文件,将文本按行写入kafka中.写入速率70000行/秒

####配置文件为JSON格式,可以使用"#"单行注释.示例(config.json)
```
{
    "runtime":{
        "DEBUG":false,
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
    "file": {
    "name":[
        "/data/inveno/HoneyBee/INV/honeybee/INV.honeybee_monitor_%Y%M%D.log",
        "/data/inveno/HoneyBee_report/INV/honeybee/INV.honeybee_monitor_%Y%M%D.log"
    ],
        #offset:0从头开始读取数据,num(int64)从指定字节开始读取数据
        "offset":0,
        "prefix":"",
        "suffix":""
    },
    "http": {
        "addr":":9143"
    },
    "tags":"collect",
    "version":"0.0.3"
}
```

配置说明
```
01. runtime.DEBUG: 是否开启cpu pprof调试模式,注意:只能在前台运行,使用ctrl+c退出后,写入cpu.pprof文件中.(详情:runtime/pprof)
02. runtime.MAXPROCS: 使用多个cpu核心
03. kafka.addrs: kafka地址
04. kafka.topic: kafka topic
05. kafka.maxthreads: 并发协程写kafka topic的数量,根据需要调整(本人测试可以开到100w,机型aws[t2.xlarge])
06. file.name: 收集文件列表,支持按时间格式匹配.
07. file.offset: 文件起始读取位置(bytes),默认每分钟将offset值记录在"var/{{file.mame}}.offset"中
08. file.prefix/suffix: 字符串需要拼接的前后缀,默认为空("")时不做任何处理
09. http.addr: HTTP监控接口(http://http.addr/monitor)
10. tags: 一些标记(TODO)
11. version: 程序版本号,必须与程序内部的版本号对应
```

1. 下载:
```
git clone https://github.com/luopengift/collect.git
cd collect
```

2. 编译:
```
./init.sh build
```

3. 运行:
```
./init.sh start
```

4. 停止:
```
./init.sh stop
```

5. 查看日志:
```
./init.sh tail
```

