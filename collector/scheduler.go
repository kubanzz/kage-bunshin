package collector

import (
	"context"
	"encoding/json"
	db "kage-bunshin/db"

	"github.com/beego/beego/v2/core/logs"
	"github.com/robfig/cron"
)

func StartSchedule(done chan struct{}) {

	// 开启节假日采集定时任务
	startHolidaySchedule()

	// 发送空结构体，保证后台运行
	done <- struct{}{}
}

func startHolidaySchedule() {
	c := cron.New()

	err := c.AddFunc("*/5 * * * * ?", func() {
		logs.Info("开始定时任务 - [获取最近七天节假日]")
		holidayCollector := HolidayHB{}
		holidayList := holidayCollector.GetLast7Holiday()
		bytes, err := json.Marshal(holidayList)
		if err != nil {
			logs.Error("Error：", err)
			return
		}

		holidayStr := string(bytes)
		db.RedisClient.Publish(context.Background(), "holiday_channel", holidayStr)
	})

	if err != nil {
		logs.Error("执行定时任务异常 - ", err)
		return
	}

	c.Start()
}
