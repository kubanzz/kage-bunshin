package controllers

import (
	"context"
	"encoding/json"
	"fmt"
	"kage-bunshin/collector"
	db "kage-bunshin/db"

	"github.com/beego/beego/v2/core/logs"
	beego "github.com/beego/beego/v2/server/web"
	"github.com/silenceper/wechat/v2"
	"github.com/silenceper/wechat/v2/cache"
	"github.com/silenceper/wechat/v2/officialaccount"
	"go.mongodb.org/mongo-driver/bson"

	offConfig "github.com/silenceper/wechat/v2/officialaccount/config"
	"github.com/silenceper/wechat/v2/officialaccount/message"
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

// =========================== 微信公众号 =================================

type WeChatController struct {
	beego.Controller
}

var officialAccount *officialaccount.OfficialAccount

func init() {
	wc := wechat.NewWechat()
	memory := cache.NewMemory()

	appId, _ := beego.AppConfig.String("wechat_appid")
	wechatAppsecret, _ := beego.AppConfig.String("wechat_appid")
	wechatToken, _ := beego.AppConfig.String("wechat_token")

	cfg := &offConfig.Config{
		AppID:     appId,
		AppSecret: wechatAppsecret,
		Token:     wechatToken,
		// EncodingAESKey: "xxxx",
		Cache: memory,
	}

	// 微信公众号API
	officialAccount = wc.GetOfficialAccount(cfg)

	logs.Info("连接微信公众号成功")
}

func (w *WeChatController) ServerWechat() {

	writer := w.Ctx.ResponseWriter
	req := w.Ctx.Request

	// 传入request和responseWriter
	server := officialAccount.GetServer(req, writer)

	// 跳过接口验证
	server.SkipValidate(true)

	// 设置接收消息的处理方法
	server.SetMessageHandler(func(msg *message.MixMessage) *message.Reply {

		// 回复消息：演示回复用户发送的消息
		// text := message.NewText(msg.Content)
		holidayDb := collector.HolidayHB{}
		holidayList := holidayDb.GetLast7Holiday()

		bytes, err := json.Marshal(holidayList)
		if err != nil {
			logs.Error("Error：", err)
		}

		holidayData := string(bytes)
		text := message.NewText(holidayData)
		return &message.Reply{MsgType: message.MsgTypeText, MsgData: text}
	})

	// 处理消息接收以及回复
	err := server.Serve()
	if err != nil {
		fmt.Println(err)
		return
	}
	// 发送回复的消息
	server.Send()
}
