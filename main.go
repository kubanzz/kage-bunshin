package main

import (
	_ "kage-bunshin/routers"

	"kage-bunshin/collector"
	"kage-bunshin/service/mq"

	beego "github.com/beego/beego/v2/server/web"
)

func main() {
	done := make(chan struct{})

	// 监听消息队列主题
	SubscribeChannel(done)

	// 启动定时任务
	// StartSchedule(done)

	beego.Run()
}

func SubscribeChannel(done chan struct{}) {
	// 订阅假期主题
	go mq.SubscribeChannel(done)
}

func StartSchedule(done chan struct{}) {
	// 启动采集器定时任务
	go collector.StartSchedule(done)
}
