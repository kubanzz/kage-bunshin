package controllers

import (
	"context"
	"encoding/json"
	"kage-bunshin/collector"
	db "kage-bunshin/db"

	"github.com/beego/beego/v2/core/logs"
	beego "github.com/beego/beego/v2/server/web"
	"go.mongodb.org/mongo-driver/bson"
)

type MainController struct {
	beego.Controller
}

func (c *MainController) Get() {
	c.Data["Website"] = "beego.vip"
	c.Data["Email"] = "astaxie@gmail.com"
	c.TplName = "index.tpl"
}

type CollectorController struct {
	beego.Controller
}

func (u *CollectorController) HelloWorld() {
	key := "841e2ad14aac42259cc6eb630965f848"
	// 香洲区
	location := 101280704

	weatherCollector := collector.Weather{}
	weatherInfo := weatherCollector.Collect(location, key)

	logs.Info(weatherInfo)
	u.Data["json"] = weatherInfo

	u.ServeJSON()
}

func (u *CollectorController) MongoConn() {
	// mongodb, _ := beego.AppConfig.String("Mongodb")
	collection := db.MongoDB.Collection("game_notify")

	var result bson.M
	filter := bson.M{"appkey": "mecha"}
	collection.FindOne(context.TODO(), filter).Decode(&result)
	logs.Info("================= %o", result)
}

func (u *CollectorController) HolidayCollect_TM() {
	holidayCollector := collector.HolidayTM{}
	res := holidayCollector.Collect("2023")

	u.Ctx.WriteString(res)
}

func (u *CollectorController) HolidayCollect_HB() {
	holidayCollector := collector.HolidayHB{}
	holidayList := holidayCollector.GetLast7Holiday()
	bytes, err := json.Marshal(holidayList)
	if err != nil {
		logs.Error("Error：", err)
		return
	}

	holidayStr := string(bytes)
	db.RedisClient.Publish(context.Background(), "holiday_channel", holidayStr)

	u.Ctx.WriteString(holidayStr)
}
