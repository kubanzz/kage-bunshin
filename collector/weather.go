package collector

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	db "kage-bunshin/db"
	"kage-bunshin/util"
	"net/http"

	"github.com/beego/beego/v2/core/logs"
	"go.mongodb.org/mongo-driver/bson"
)

type Weather struct {
	ObsTime    string `json:"obsTime"`
	Temp       string `json:"temp"`
	FeelsLike  string `json:"feelsLike"`
	Icon       string `json:"icon"`
	Text       string `json:"text"`
	Wind360    string `json:"wind360"`
	WindDir    string `json:"windDir"`
	WindScale  string `json:"windScale"`
	WindSpeed  string `json:"windSpeed"`
	Humidity   string `json:"humidity"`
	Precip     string `json:"precip"`
	Pressure   string `json:"pressure"`
	Visibility string `json:"vis"`
	Cloud      string `json:"cloud"`
	Dew        string `json:"dew"`
}

type WeatherPrediction struct {
	FxDate         string `json:"fxDate"`
	Sunrise        string `json:"sunrise"`
	Sunset         string `json:"sunset"`
	Moonrise       string `json:"moonrise"`
	Moonset        string `json:"moonset"`
	MoonPhase      string `json:"moonPhase"`
	MoonPhaseIcon  string `json:"moonPhaseIcon"`
	TempMax        string `json:"tempMax"`
	TempMin        string `json:"tempMin"`
	IconDay        string `json:"iconDay"`
	TextDay        string `json:"textDay"`
	IconNight      string `json:"iconNight"`
	TextNight      string `json:"textNight"`
	Wind360Day     string `json:"wind360Day"`
	WindDirDay     string `json:"windDirDay"`
	WindScaleDay   string `json:"windScaleDay"`
	WindSpeedDay   string `json:"windSpeedDay"`
	Wind360Night   string `json:"wind360Night"`
	WindDirNight   string `json:"windDirNight"`
	WindScaleNight string `json:"windScaleNight"`
	WindSpeedNight string `json:"windSpeedNight"`
	Humidity       string `json:"humidity"`
	Precip         string `json:"precip"`
	Pressure       string `json:"pressure"`
	Vis            string `json:"vis"`
	Cloud          string `json:"cloud"`
	UvIndex        string `json:"uvIndex"`
}

func (w *Weather) Collect(location int, key string) Weather {
	url := fmt.Sprintf("https://devapi.qweather.com/v7/weather/now?location=%d&key=%s", location, key)
	// params := map[string]string{
	// 	"key":      key,
	// 	"location": strconv.FormatInt(int64(location), 10),
	// }

	unGzipResult := requestHttp(url)

	var weatherInfoDto Weather
	jsonObject := make(map[string]interface{})
	err := json.Unmarshal(unGzipResult, &jsonObject)
	if err != nil {
		panic(err)
	}

	nowObj := jsonObject["now"]
	nowJSON, err := json.Marshal(nowObj)
	if err != nil {
		panic(err)
	}

	err = json.Unmarshal(nowJSON, &weatherInfoDto)
	if err != nil {
		panic(err)
	}

	logs.Info("获取到的天气数据：%o", weatherInfoDto)

	return weatherInfoDto
}

func (w *Weather) GetTodaytWeather() WeatherPrediction {
	mongoClient := db.MongoDB

	timeUtil := util.TimeUtil{}
	nowDate := timeUtil.GetNowDate()

	key := "841e2ad14aac42259cc6eb630965f848"
	// 香洲区
	location := 101280704

	// 先查数据库，不存在则http请求API获取
	var result bson.M
	err := mongoClient.Collection("weather_pre").FindOne(context.TODO(), bson.M{"date": nowDate}).Decode(&result)
	if err != nil {
		w.CollectWeatherPrediction(location, key)
	}

	mongoClient.Collection("weather_pre").FindOne(context.TODO(), bson.M{"date": nowDate}).Decode(&result)
	weatherPreJsonStr := result["data"].(string)
	var weatherPre WeatherPrediction
	json.Unmarshal([]byte(weatherPreJsonStr), &weatherPre)

	return weatherPre

}

// 获取当天的天气预报
func (w *Weather) CollectWeatherPrediction(location int, key string) []WeatherPrediction {
	mongoClient := db.MongoDB
	url := fmt.Sprintf("https://devapi.qweather.com/v7/weather/3d?location=%d&key=%s", location, key)

	unGzipResult := requestHttp(url)

	var weatherPreList []WeatherPrediction
	jsonObject := make(map[string]interface{})
	err := json.Unmarshal(unGzipResult, &jsonObject)
	if err != nil {
		panic(err)
	}

	nowObj := jsonObject["daily"]
	nowJSON, err := json.Marshal(nowObj)
	if err != nil {
		panic(err)
	}

	err = json.Unmarshal(nowJSON, &weatherPreList)
	if err != nil {
		panic(err)
	}

	for i := 0; i < len(weatherPreList); i++ {
		weatherPre := weatherPreList[i]
		weatherPreByte, err := json.Marshal(weatherPre)
		if err != nil {
			logs.Error("格式化[%o]天气数据失败", weatherPre.FxDate, err)
			continue
		}
		bson := bson.D{
			{Key: "date", Value: weatherPre.FxDate},
			{Key: "location", Value: location},
			{Key: "source", Value: "HF"},
			{Key: "data", Value: string(weatherPreByte)},
		}

		_, err = mongoClient.Collection("weather_pre").InsertOne(context.TODO(), bson)
		if err != nil {
			logs.Error(err)
		}
	}

	logs.Info("获取到的天气数据：%o", weatherPreList)

	if weatherPreList == nil {
		return nil
	}

	return weatherPreList
}

func requestHttp(url string) []byte {
	client := &http.Client{}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		panic(err)
	}

	req.Header.Set("Accept-Encoding", "gzip, deflate, br")
	req.Header.Set("Accept-Charset", "utf-8")
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}

	var unGzipResult []byte
	if resp.Header.Get("Content-Encoding") == "gzip" {
		gzipReader, err := gzip.NewReader(bytes.NewReader(body))
		if err != nil {
			panic(err)
		}
		defer gzipReader.Close()

		unGzipResult, err = ioutil.ReadAll(gzipReader)
		if err != nil {
			panic(err)
		}
	} else {
		unGzipResult = body
	}

	return unGzipResult
}
