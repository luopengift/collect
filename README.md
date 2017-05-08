### collect

####读取文件,将文本按行写入kafka中.写入速率70000行/秒

####配置文件示例(config.json)
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
```
字段说明
```
1. runtime.DEBUG: 是否开启cpu pprof调试模式,注意:只能在前台运行,使用ctrl+c退出后,写入cpu.pprof文件中.(详情:runtime/pprof)
2. runtime.MAXPROCS: 使用多个cpu核心
3. kafka.addrs: kafka地址
4. kafka.topic: kafka topic
5. kafka.maxthreads: 并发协程写kafka topic的数量,根据需要调整(本人测试可以开到100w,机型aws[t2.xlarge])
6. file: 收集文件列表,支持按时间格式匹配,注意:文件名中不能包含数字.(详情:https://github.com/luopengift/golibs/tree/master/file)
7. prefix/suffix: 字符串需要拼接的前后缀,默认为空("")时不做任何处理
8. http.addr: HTTP监控接口(http://http.addr/debug)
9. tags: 一些标记(TODO)
```

1. 下载:
```
git clone https://github.com/luopengift/collect.git
cd collect
```

2. 编译:
```
./control build
```

3. 运行:
```
./control start
```

4. 停止:
```
./control stop
```

5. 查看日志:
```
./control tail
```

