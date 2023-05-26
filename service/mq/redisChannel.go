package mq

import (
	"context"
	"encoding/json"
	"fmt"
	"kage-bunshin/collector"
	"kage-bunshin/controllers"
	"kage-bunshin/db"
	"kage-bunshin/util"

	"github.com/astaxie/beego/logs"
	"go.mongodb.org/mongo-driver/bson"
)

func SubscribeChannel(done chan struct{}) {

	// 订阅节假日采集
	go subscribeHolidayChannel()

	// 订阅天气预报采集
	go subscribeWeatherChannel()

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
		logs.Info("接收到消息：%s", msg.Payload)
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
		go wechatService.SendBroadcastMessage(sendMessage)

		timeUtil := util.TimeUtil{}

		bson := bson.D{
			{Key: "type", Value: "holiday"},
			{Key: "data", Value: recvMessage},
			{Key: "createDate", Value: timeUtil.GetNowDate()},
		}

		mongoClient := db.MongoDB
		mongoClient.Collection("notification_comm").InsertOne(context.TODO(), bson)
	}
}

func subscribeWeatherChannel() {
	logs.Info("订阅消息队列主题：weather_channel")
	pubsub := db.RedisClient.Subscribe(context.Background(), "weather_channel")
	// 在函数结束前执行（延迟执行）
	defer pubsub.Close()

	var sendMessage = ""
	// 遍历接收到的消息
	for msg := range pubsub.Channel() {
		// 处理接收到的消息
		logs.Info("接收到消息：%s", msg.Payload)
		recvMessage := msg.Payload

		var weatherPre collector.WeatherPrediction
		err := json.Unmarshal([]byte(recvMessage), &weatherPre)
		if err != nil {
			panic(err)
		}

		sendMessage += fmt.Sprintf("气温区间：(%s - %s)\n", weatherPre.TempMin, weatherPre.TempMax)
		sendMessage += fmt.Sprintf("相对湿度：%s\n", weatherPre.Humidity)
		sendMessage += fmt.Sprintf("能见度：%s\n", weatherPre.Vis)
		sendMessage += fmt.Sprintf("白天天气：%s\n", weatherPre.TextDay)
		sendMessage += fmt.Sprintf("白天风力：%s\n", weatherPre.WindScaleDay)
		sendMessage += fmt.Sprintf("白天风向：%s\n", weatherPre.WindDirDay)
		sendMessage += fmt.Sprintf("晚上天气：%s\n", weatherPre.TextNight)
		sendMessage += fmt.Sprintf("晚上风力：%s\n", weatherPre.WindScaleNight)
		sendMessage += fmt.Sprintf("晚上风向：%s\n", weatherPre.WindDirNight)

		wechatService := controllers.WeChatController{}
		go wechatService.SendBroadcastMessage(sendMessage)

		timeUtil := util.TimeUtil{}

		bson := bson.D{
			{Key: "type", Value: "weather"},
			{Key: "data", Value: recvMessage},
			{Key: "createDate", Value: timeUtil.GetNowDate()},
		}

		mongoClient := db.MongoDB
		mongoClient.Collection("notification_comm").InsertOne(context.TODO(), bson)
	}
}
