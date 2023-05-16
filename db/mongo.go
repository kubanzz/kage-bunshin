package db

import (
	"context"
	"log"

	"github.com/astaxie/beego/logs"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	beego "github.com/beego/beego/v2/server/web"
)

var MongoDB *mongo.Database

func init() {
	mongodbURI, _ := beego.AppConfig.String("mongodb_uri")
	mongodbDataBase, _ := beego.AppConfig.String("mongodb_db")

	clientOptions := options.Client().ApplyURI(mongodbURI)
	client, err := mongo.Connect(context.Background(), clientOptions)
	if err != nil {
		log.Fatal(err)
	}

	// 检查连接
	err = client.Ping(context.Background(), nil)
	if err != nil {
		logs.Error(err)
	}

	logs.Info("数据库连接成功")
	// 获取数据库
	MongoDB = client.Database(mongodbDataBase)
}
