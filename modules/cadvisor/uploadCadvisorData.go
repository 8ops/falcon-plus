package main

import (
	"time"
	"github.com/golang/glog"
	"gitlab.wuxingdev.cn/go/util/logs"
	"flag"
)

var (
	//CadvisorPort = "18080"
	CadvisorPort = "4194"
	Interval time.Duration
)

func main() {
	flag.Parse()
	logs.InitLogs()
	defer logs.FlushLogs()
	glog.Info("sys start")

	Interval = 60 * time.Second
	t := time.NewTicker(Interval)
	for {
		<-t.C
		pushData()
	}
}

