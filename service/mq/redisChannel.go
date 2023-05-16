package mq

import (
	"context"
	"kage-bunshin/db"

	"github.com/beego/beego/v2/core/logs"
)

func SubscribeChannel(done chan struct{}) {

	// 订阅节假日采集
	subscribeHolidayChannel()

	// 发送信号通知主Goroutine程序已关闭
	done <- struct{}{}
}

func subscribeHolidayChannel() {
	logs.Info("订阅消息队列主题：holiday_channel")
	pubsub := db.RedisClient.Subscribe(context.Background(), "holiday_channel")
	// 在函数结束前执行（延迟执行）
	defer pubsub.Close()

	// 遍历接收到的消息
	for msg := range pubsub.Channel() {
		// 处理接收到的消息
		logs.Info("接收到 [holiday_channel]消息：%s", msg.Payload)
	}
}
