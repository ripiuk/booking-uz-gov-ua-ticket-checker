package booking_uz

import (
	"encoding/json"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

const (
	baseURL string = "https://booking.uz.gov.ua/uk/"
)

// Get a list of stations by station name
// name: e.g "Вінниця"
// return: e.g [map[region:<nil> title:Вінниця value:%!s(float64=2.2002e+06)]
// 				map[region:<nil> title:Вінниця-Вант. value:%!s(float64=2.200318e+06)]]
func GetStations(name string) []map[string]interface{} {
	// Generating URL
	name = url.QueryEscape(name)
	stationInfoURL := baseURL + "train_search/station/?term=" + name
	log.Printf("Sending request to the url %s", stationInfoURL)

	resp, err := http.Get(stationInfoURL)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	var result []map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		panic(err)
	}
	log.Printf("Got response: %s", result)
	return result
}

// Get trains list with amount of places in each one
// fromStation: e.g "2200200"
// toStation: e.g "2218200"
// date: e.g "2019-05-14"
// response: e.g map[
// 		data:map[list:[map[allowBooking:1 allowPrivilege:1 allowStudent:1 category:0 child:map[maxDate:2019-05-04
// 		minDate:2005-05-15] from:map[code:2200200 date:вівторок, 14.05.2019 sortTime:1.55786586e+09 srcDate:2019-05-14
// 		station:Вінниця stationTrain:Кременчук time:23:31] isCis:0 isEurope:0 isTransformer:0 num:150О
// 		to:map[code:2218200 date:середа, 15.05.2019 sortTime:1.55789976e+09 station:Івано-Франківськ
// 		stationTrain:Ворохта time:08:56] travelTime:9:25 types:[map[id:К letter:К places:9 title:Купе]]]]]]
func GetTrains(fromStation string, toStation string, date string) map[string]interface{} {
	apiUrl := baseURL
	resource := "/train_search/"
	data := url.Values{}
	data.Set("date", "2019-05-14")
	data.Set("from", "2200200")
	data.Set("to", "2218200")
	data.Set("time", "00:00")

	u, _ := url.ParseRequestURI(apiUrl)
	u.Path = resource
	urlStr := u.String()

	client := &http.Client{}
	r, err := http.NewRequest("POST", urlStr, strings.NewReader(data.Encode())) // URL-encoded payload
	if err != nil {
		panic(err)
	}
	r.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	r.Header.Add("Content-Length", strconv.Itoa(len(data.Encode())))

	resp, err := client.Do(r)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	log.Println(resp.Status)
	var trainsInfo map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&trainsInfo)
	if err != nil {
		panic(err)
	}
	log.Println("Response Body:", trainsInfo)
	return trainsInfo
}

//func GetTrainDetail (train string) string {
//
//}