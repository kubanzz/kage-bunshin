package collector

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/beego/beego/v2/core/logs"
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

func (w *Weather) Collect(location int, key string) Weather {
	url := fmt.Sprintf("https://devapi.qweather.com/v7/weather/now?location=%d&key=%s", location, key)
	// params := map[string]string{
	// 	"key":      key,
	// 	"location": strconv.FormatInt(int64(location), 10),
	// }

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

	var weatherInfoDto Weather
	jsonObject := make(map[string]interface{})
	err = json.Unmarshal(unGzipResult, &jsonObject)
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
