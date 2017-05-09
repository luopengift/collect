#!/bin/sh
#
# collect - this script starts and stops the collect daemon
#
# chkconfig:
# description:
#               
# processname: 
# config:      config.json
# config:      config.json
# pidfile:     /var/collect.pid

# Source env
#. /etc/profile

# Source function library.
#. /etc/rc.d/init.d/functions

# Source networking configuration.
#. /etc/sysconfig/network

NOW=$(date +%Y-%m-%d.%H:%M:%S)
echo $NOW
DIR=$(pwd) #当前目录
BASEDIR=$(dirname $(dirname $(dirname $DIR))) 
export GOPATH=$GOPATH:$BASEDIR
export GOBIN=$GOPATH/bin/
echo "GOPATH init Finished. GOPATH=$GOPATH"

#################################
APP=$(basename $DIR)
######创建程序运行临时文件#######

mkdir -p $DIR/var
PIDFile=$DIR/var/$APP.pid
LOGFile=$DIR/var/$APP.log

#编译$2,默认$2未指定时编译main.go
if [ -z $2 ];then #如果$2为空
    main="main"
else
    main=$2
fi 

################################
function check_PID() {
    if [ -f $PIDFile ];then
        PID=$(cat $PIDFile)
        if [ -n $PID ]; then
            running=$(ps -p $PID|grep -v "PID TTY" |wc -l)
            return $running
        fi
    fi
    return 0
}

function build() {
    gofmt -w .
    go build -o $APP $1.go
    if [ $? -ne 0 ]; then
        exit $?
    fi
    echo "build $APP success."
}

function start() {
    check_PID
    local running=$?
    if [ $running -gt 0 ];then
        echo -n "$APP now is running already, PID="
        cat $PIDFile
        return 1
    fi

    nohup  ./$APP  >$LOGFile 2>&1 &
    sleep 1
    running=`ps -p $! | grep -v "PID TTY" | wc -l`
    if [ $running -gt 0 ];then
        echo $! > $PIDFile
        echo "$APP started..., PID=$!"
    else
        echo "$APP failed to start!!!"
        return 1
    fi
}

function status() {
    ps -ef |grep $APP|grep -v grep
    check_PID    
    local running=$?
    if [ $running -gt 0 ];then
        echo -n "$APP now is running already, PID="
        cat $PIDFile
        return 1
    else
        echo "$APP is stopped..."
        return 1
    fi
    
}


function debug() {
    go run $1.go
}

function stop() {
    local PID=$(cat $PIDFile)
    kill $PID
    rm -f $PIDFile
    echo "$APP stoped..."
}

function restart() {
    stop
    sleep 1
    start   
}

function tailf() {
   tail -f $LOGFile
}

function help() {
    echo "$0 build|start|stop|kill|restart|reload|run|tail|docs|sslkey"
}

if [ "$1" == "" ]; then
    help
elif [ "$1" == "build" ];then
    build $main
elif [ "$1" == "start" ];then
    start
elif [ "$1" == "debug" ];then
    debug $main
elif [ "$1" == "stop" ];then
    stop
elif [ "$1" == "kill" ];then
    killall
elif [ "$1" == "restart" ];then
    restart
elif [ "$1" == "reload" ];then
    reload
elif [ "$1" == "status" ];then
    status
elif [ "$1" == "run" ];then
    run 
elif [ "$1" == "tail" ];then
    tailf
elif [ "$1" == "docs" ];then
    docs
elif [ "$1" == "sslkey" ];then
    sslkey
else
    help
fi
#bee api  cmdb-api -driver=mysql -conn="root:root888@tcp(127.0.0.1:3306)/cmdb"

