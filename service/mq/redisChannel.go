package mq

import (
	"context"
	"encoding/json"
	"kage-bunshin/collector"
	"kage-bunshin/controllers"
	"kage-bunshin/db"
	"kage-bunshin/util"

	"github.com/astaxie/beego/logs"
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

	var sendMessage = "最近七天的节日：\n"
	// 遍历接收到的消息
	for msg := range pubsub.Channel() {
		// 处理接收到的消息
		logs.Info("接收到[holiday_channel]消息：%s", msg.Payload)
		recvMessage := msg.Payload

		var holidays []collector.Holiday
		err := json.Unmarshal([]byte(recvMessage), &holidays)
		if err != nil {
			panic(err)
		}

		for i := 0; i < len(holidays); i++ {
			holiday := holidays[i]

			sendMessage += holiday.Name + " - " + holiday.Date

			timeUtil := util.TimeUtil{}
			week := timeUtil.GetWeekend(holiday.Date, "2006-01-02")
			if week != "" {
				sendMessage += "(" + week + ")"
			}
			sendMessage += "\n"
		}

		wechatService := controllers.WeChatController{}
		wechatService.SendBroadcastMessage(sendMessage)
	}
}
