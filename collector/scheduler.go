package collector

import (
	"context"
	"encoding/json"
	db "kage-bunshin/db"

	"github.com/beego/beego/v2/core/logs"
	"github.com/robfig/cron"
)

var commonCron = "0 0 8 * * ?"

func StartSchedule(done chan struct{}) {

	// 开启节假日采集定时任务
	startHolidayCollectSchedule()

	// 开启天气预报采集定时任务
	startWeatherCollectSchedule()

	// 发送空结构体，保证后台运行
	done <- struct{}{}
}

func startHolidayCollectSchedule() {
	c := cron.New()

	err := c.AddFunc(commonCron, func() {
		logs.Info("开始定时任务 - [获取最近七天节假日]")

		days := 30
		holidayCollector := HolidayHB{}
		holidayList := holidayCollector.GetLastNHoliday(days)

		if len(holidayList) == 0 {
			logs.Info("最近%d天内无节假日", days)
			return
		}

		bytes, err := json.Marshal(holidayList)
		if err != nil {
			logs.Error("Error：", err)
			return
		}

		holidayStr := string(bytes)
		logs.Info("发送节假日采集数据[%s]", holidayStr)
		db.RedisClient.Publish(context.Background(), "holiday_channel", holidayStr)
	})

	if err != nil {
		logs.Error("执行定时任务异常 - ", err)
		return
	}

	c.Start()
}

func startWeatherCollectSchedule() {
	c := cron.New()

	err := c.AddFunc(commonCron, func() {
		logs.Info("开始定时任务 - [获取当天天气预报]")
		weatherCollector := Weather{}
		weatherPre := weatherCollector.GetTodaytWeather()

		bytes, err := json.Marshal(weatherPre)
		if err != nil {
			logs.Error("Error：", err)
			return
		}

		weatherStr := string(bytes)
		logs.Info("发送天气采集数据[%s]", weatherStr)
		db.RedisClient.Publish(context.Background(), "weather_channel", weatherStr)
	})

	if err != nil {
		logs.Error("执行定时任务异常 - ", err)
		return
	}

	c.Start()
}
