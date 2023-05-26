package collector

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"time"

	db "kage-bunshin/db"

	"github.com/beego/beego/v2/core/logs"
	"go.mongodb.org/mongo-driver/bson"
)

/** ------------------------ common -------------------------------**/

func GetHttpURL(url string, header http.Header) []byte {
	client := &http.Client{}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		panic(err)
	}

	req.Header = header

	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}

	body, _ := ioutil.ReadAll(resp.Body)
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

/** ------------------------ 提莫节假日 -------------------------------**/

type HolidayTM struct {
	Holiday bool   `json:"holiday"`
	Name    string `json:"name"`
	Wage    int    `json:"wage"`
	Date    string `json:"date"`
	Rest    int    `json:"rest"`
	After   bool   `json:"after,omitempty"`
	Target  string `json:"target,omitempty"`
}

type Response struct {
	Code    int                  `json:"code"`
	Holiday map[string]HolidayTM `json:"holiday"`
}

type Result struct {
	data string `bson:"data"`
}

func (h *HolidayTM) Collect(year string) string {
	if year == "" {
		return ""
	}

	logs.Info(" 获取[%s]年的假期", year)
	mongoClient := db.MongoDB

	// 先查数据库，不存在则http请求API获取
	var result bson.M
	err := mongoClient.Collection("holiday").FindOne(context.TODO(), bson.M{"year": year, "source": "TM"}).Decode(&result)
	if err == nil {
		return result["data"].(string)
	}

	url := fmt.Sprintf("http://timor.tech/api/holiday/year/%s", year)
	header := http.Header{}
	header.Set("Accept-Encoding", "gzip, deflate")
	header.Set("Accept-Charset", "utf-8")
	header.Set("Content-Type", "application/json")
	header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/112.0.0.0 Safari/537.36")
	unGzipResult := GetHttpURL(url, header)

	responseJson := make(map[string]interface{})
	err = json.Unmarshal(unGzipResult, &responseJson)
	if err != nil {
		panic(err)
	}

	holidayData := responseJson["holiday"]
	holidayJson, _ := json.Marshal(holidayData)
	data := string(holidayJson)

	bson := bson.D{
		{Key: "year", Value: year},
		{Key: "source", Value: "TM"},
		{Key: "data", Value: data},
	}
	_, err = mongoClient.Collection("holiday").InsertOne(context.TODO(), bson)
	if err != nil {
		logs.Error(err)
	}

	return data
}

func (h *HolidayTM) GetLast7Holiday() map[string]HolidayTM {
	// 获取当前时间
	currentTime := time.Now()
	year := currentTime.Year()

	holidaysJson := h.Collect(strconv.Itoa(year))
	// logs.Info("holidaysJson ========== " + holidaysJson)

	var jsonMap map[string]HolidayTM
	json.Unmarshal([]byte(holidaysJson), &jsonMap)

	return jsonMap
}

/** ------------------------ HB节假日 -------------------------------**/

type Holiday struct {
	Date      string
	IsHoliday bool //是否为假期节假日（即有假期的）
	Name      string
	IsLegal   bool //是否为法定节假日（三倍工资）
}

type HolidayDataHB struct {
	List []HolidayHB `json:"list"`
}

type HolidayHB struct {
	Year              int    `json:"year"`
	Month             int    `json:"month"`
	Date              int    `json:"date"`
	YearWeek          int    `json:"yearweek"`
	YearDay           int    `json:"yearday"`
	LunarYear         int    `json:"lunar_year"`
	LunarMonth        int    `json:"lunar_month"`
	LunarDate         int    `json:"lunar_date"`
	LunarYearDay      int    `json:"lunar_yearday"`
	Week              int    `json:"week"`
	Weekend           int    `json:"weekend"`
	Workday           int    `json:"workday"`
	Holiday           int    `json:"holiday"`
	HolidayOr         int    `json:"holiday_or"`
	HolidayOvertime   int    `json:"holiday_overtime"`
	HolidayToday      int    `json:"holiday_today"`
	HolidayLegal      int    `json:"holiday_legal"`
	HolidayRecess     int    `json:"holiday_recess"`
	YearCN            string `json:"year_cn"`
	MonthCN           string `json:"month_cn"`
	DateCN            string `json:"date_cn"`
	YearWeekCN        string `json:"yearweek_cn"`
	YearDayCN         string `json:"yearday_cn"`
	LunarYearCN       string `json:"lunar_year_cn"`
	LunarMonthCN      string `json:"lunar_month_cn"`
	LunarDateCN       string `json:"lunar_date_cn"`
	LunarYearDayCN    string `json:"lunar_yearday_cn"`
	WeekCN            string `json:"week_cn"`
	WeekendCN         string `json:"weekend_cn"`
	WorkdayCN         string `json:"workday_cn"`
	HolidayCN         string `json:"holiday_cn"`
	HolidayOrCN       string `json:"holiday_or_cn"`
	HolidayOvertimeCN string `json:"holiday_overtime_cn"`
	HolidayTodayCN    string `json:"holiday_today_cn"`
	HolidayLegalCN    string `json:"holiday_legal_cn"`
	HolidayRecessCN   string `json:"holiday_recess_cn"`
}

type ResponseHB struct {
	Code int           `json:"code"`
	Msg  string        `json:"msg"`
	Data HolidayDataHB `json:"data"`
}

func (h *HolidayHB) Collect(year string) string {
	logs.Info(" 获取[%s]年的假期", year)
	mongoClient := db.MongoDB

	// 先查数据库，不存在则http请求API获取
	var result bson.M
	err := mongoClient.Collection("holiday").FindOne(context.TODO(), bson.M{"year": year, "source": "HB"}).Decode(&result)
	if err == nil {
		return result["data"].(string)
	}

	url := fmt.Sprintf("https://api.apihubs.cn/holiday/get?cn=1&&year=%s&&size=3660&&holiday_today=1", year)
	header := http.Header{}
	header.Set("Accept-Encoding", "gzip, deflate")
	header.Set("Accept-Charset", "utf-8")
	header.Set("Content-Type", "application/json")
	header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/112.0.0.0 Safari/537.36")
	unGzipResult := GetHttpURL(url, header)

	response := ResponseHB{}
	json.Unmarshal(unGzipResult, &response)
	logs.Info("获取到的API的数据：%o", response)

	holidayMap := make(map[string]interface{})
	for _, h := range response.Data.List {

		// 将日期字符串解析为时间对象
		date, err := time.Parse("20060102", strconv.Itoa(h.Date))
		if err != nil {
			logs.Error("日期解析错误:", err)
			return "日期解析错误"
		}

		// 将时间对象格式化为指定的日期格式
		formattedDate := date.Format("2006-01-02")

		holiday := Holiday{
			Date:      formattedDate,
			IsHoliday: h.HolidayRecess == 1,
			Name:      h.HolidayCN,
			IsLegal:   h.HolidayLegal == 1,
		}
		holidayMap[formattedDate] = holiday
	}

	holidayJson, _ := json.Marshal(holidayMap)
	data := string(holidayJson)

	bson := bson.D{
		{Key: "year", Value: year},
		{Key: "source", Value: "HB"},
		{Key: "data", Value: data},
	}

	_, err = mongoClient.Collection("holiday").InsertOne(context.TODO(), bson)
	if err != nil {
		logs.Error(err)
	}

	return data
}

func (h *HolidayHB) GetLast7Holiday() []Holiday {
	// 获取当前时间
	currentTime := time.Now()
	year := currentTime.Year()

	holidaysJson := h.Collect(strconv.Itoa(year))
	// logs.Info("holidaysJson ========== " + holidaysJson)

	jsonMap := make(map[string]Holiday)
	json.Unmarshal([]byte(holidaysJson), &jsonMap)

	holidayList := []Holiday{}
	for i := 0; i < 7; i++ {
		date := currentTime.AddDate(0, 0, i)
		formateDate := date.Format("2006-01-02")
		if targetHoliday, ok := jsonMap[formateDate]; ok {
			holidayList = append(holidayList, targetHoliday)
		}
	}

	return holidayList
}

func (h *HolidayHB) GetLastNHoliday(days int) []Holiday {
	// 获取当前时间
	currentTime := time.Now()
	year := currentTime.Year()

	holidaysJson := h.Collect(strconv.Itoa(year))
	// logs.Info("holidaysJson ========== " + holidaysJson)

	jsonMap := make(map[string]Holiday)
	json.Unmarshal([]byte(holidaysJson), &jsonMap)

	holidayList := []Holiday{}
	for i := 0; i < days; i++ {
		date := currentTime.AddDate(0, 0, i)
		formateDate := date.Format("2006-01-02")
		if targetHoliday, ok := jsonMap[formateDate]; ok {
			holidayList = append(holidayList, targetHoliday)
		}
	}

	return holidayList
}
