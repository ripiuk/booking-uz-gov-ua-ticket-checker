package booking_uz

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

const (
	baseURL string = "https://booking.uz.gov.ua/uk/"
)

type TrainsInfo struct {
	Data struct{
		List []struct{
			TrainId string `json:"num"`
			TravelTime string
			From struct{
				Station string
				Date string
				Time string
			}
			To struct{
				Station string
				Date string
				Time string
			}
			Types []struct{
				Title string
				Places int16
			}
		}
		Warning string `json:",omitempty"`
	}
}

type TrainsInfoError struct {
	Data string
	Error int
}

// Get a list of stations by station name
// name: e.g "Вінниця"
// return: e.g [map[region:<nil> title:Вінниця value:%!s(float64=2.2002e+06)]
// 				map[region:<nil> title:Вінниця-Вант. value:%!s(float64=2.200318e+06)]]
func Stations(name string) ([]map[string]interface{}, error) {
	// Generating URL
	name = url.QueryEscape(name)
	stationInfoURL := baseURL + "train_search/station/?term=" + name
	log.Printf("Sending request to the url %s", stationInfoURL)

	resp, err := http.Get(stationInfoURL)
	if err != nil {
		log.Println(err)
		return []map[string]interface{}{}, errors.New("неможливо виконати GET запит. ")
	}
	defer resp.Body.Close()

	var result []map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		log.Println(err)
		return []map[string]interface{}{}, errors.New("неможливо перетворити отримані дані у JSON формат. ")
	}
	log.Printf("Got response: %s", result)
	if len(result) == 0 {
		log.Println("No stations found")
		return []map[string]interface{}{}, errors.New("відповідних станцій не знайдено. ")
	}
	return result, nil
}

func FirstStationId(stationsInfo []map[string]interface{}) (string, error) {
	if stationsInfo[0]["value"] == nil {
		return "", errors.New("помилка парсингу даних. ")
	}
	stationId := fmt.Sprintf("%.0f", stationsInfo[0]["value"])
	log.Println(stationId)
	return stationId, nil
}

func PotentialStations(stationsInfo []map[string]interface{}) ([]string, error) {
	var titles []string
	if len(stationsInfo) == 0 {
		log.Println("No stations found")
		return []string{}, errors.New("відповідних станцій не знайдено. ")
	}
	for i := range stationsInfo {
		if stationsInfo[i]["title"] == nil {
			return []string{}, errors.New("помилка парсингу даних. ")
		}
		titles = append(titles, stationsInfo[i]["title"].(string))
	}
	return titles, nil
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
func Trains(fromStation string, toStation string, date string) (TrainsInfo, error) {
	apiUrl := baseURL
	resource := "/train_search/"
	data := url.Values{}
	data.Set("date", date)
	data.Set("from", fromStation)
	data.Set("to", toStation)
	data.Set("time", "00:00")

	u, _ := url.ParseRequestURI(apiUrl)
	u.Path = resource
	urlStr := u.String()

	client := &http.Client{}
	r, err := http.NewRequest("POST", urlStr, strings.NewReader(data.Encode())) // URL-encoded payload
	if err != nil {
		return TrainsInfo{}, errors.New("неможливо згенерувати POST запит. ")
	}
	r.Header.Set("Connection", "keep-alive")
	r.Header.Set("Content-Length", strconv.Itoa(len(data.Encode())))
	r.Header.Set("Accept", "*/*")
	r.Header.Set("Origin", "https://booking.uz.gov.ua")
	r.Header.Set("X-Requested-With", "XMLHttpRequest")
	r.Header.Set("cache-version", "753")
	r.Header.Set("User-Agent", "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 " +
		"(KHTML, like Gecko) Chrome/74.0.3729.169 Safari/537.36")
	r.Header.Set("Content-Type", "application/x-www-form-urlencoded; charset=UTF-8")
	r.Header.Set("Accept-Encoding", "gzip, deflate, br")
	r.Header.Set("Accept-Language", "uk-UA,uk;q=0.9,ru;q=0.8,en-US;q=0.7,en;q=0.6")

	resp, err := client.Do(r)
	if err != nil {
		return TrainsInfo{}, errors.New("неможливо виконати POST запит. ")
	}
	defer resp.Body.Close()

	log.Println(resp.Status)
	var info TrainsInfo
	bodyBytes, _ := ioutil.ReadAll(resp.Body)
	resp.Body = ioutil.NopCloser(bytes.NewBuffer(bodyBytes))
	decErr := json.NewDecoder(resp.Body).Decode(&info)
	if decErr != nil {
		var infoWithError TrainsInfoError
		resp.Body = ioutil.NopCloser(bytes.NewBuffer(bodyBytes))
		newErr := json.NewDecoder(resp.Body).Decode(&infoWithError)
		if newErr != nil {
			return TrainsInfo{}, errors.New("неможливо перетворити отримані дані у JSON формат. ")
		}
		return TrainsInfo{}, errors.New(infoWithError.Data)
	}
	log.Println("Response Body:", info)
	return info, nil
}

//func GetTrainDetail (train string) string {
//
//}