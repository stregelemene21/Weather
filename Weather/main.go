package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"math"
	"net/http"
	"os"
	"sort"
	"time"
)

//структуры для получения необходимых данных

type Weather struct {
	Daily []Day `json:"daily"`
}
type Temp struct {
	NightTemp float64 `json:"night"`
}
type FeelTemp struct {
	NightFeel float64 `json:"night"`
}
type Day struct {
	Datetime int      `json:"dt"`
	Sunrise  int      `json:"sunrise"`
	Sunset   int      `json:"sunset"`
	Temp     Temp     `json:"temp"`
	FeelTemp FeelTemp `json:"feels_like"`
}

//функция для загрузки JSON файла из api openweathermap.org/
func downloadFile(URL, fileName string) error {
	//Get the response bytes from the url
	response, err := http.Get(URL)
	if err != nil {
		log.Fatal(err)
	}
	defer response.Body.Close()

	if response.StatusCode != 200 {
		return errors.New("Received non 200 response code")
	}
	//Create a empty file
	file, err := os.Create(fileName)
	if err != nil {
		return err
	}
	defer file.Close()

	//Write the bytes to the fiel
	_, err = io.Copy(file, response.Body)
	if err != nil {
		return err
	}
	return nil
}

func main() {

	//количество дней для анализа (включая текущий)
	countDays := 5

	fileName := "data.json"

	//ключ api
	apiKey := "2692f264b4e224a1ba36dafc7f4748d1"

	//получаем данные о погоде
	URL := fmt.Sprintf("https://api.openweathermap.org/data/2.5/onecall?lat=55.45&lon=37.37&exclude=current,minutely,hourly,alerts&appid=%v&units=metric&lang=ru", apiKey)
	err := downloadFile(URL, fileName)
	if err != nil {
		fmt.Println("Error opening file:", err)
	}

	jsonFile, err := os.Open(fileName)
	if err != nil {
		fmt.Println("Error opening file:", err)
	}
	defer jsonFile.Close()

	JSONdata, _ := ioutil.ReadAll(jsonFile)

	//итоговая структура с данными
	var WeatherData Weather
	json.Unmarshal(JSONdata, &WeatherData)

	//вычисляем самую большую продолжительность дня за 5-ти дневный период

	DayLength := make(map[int]int)

	for i := 0; i < countDays; i++ {
		DayLength[WeatherData.Daily[i].Sunset-WeatherData.Daily[i].Sunrise] = WeatherData.Daily[i].Datetime
	}

	keys := make([]int, 0, len(DayLength))
	for v := range DayLength {
		keys = append(keys, v)
	}
	sort.Ints(keys)

	mostLongDay := DayLength[keys[len(keys)-1]]
	m := time.Unix(int64(mostLongDay), 0)
	mostLongDay1 := m.Format("02-01-2006")

	mostLongTime := keys[len(keys)-1]
	n := time.Unix(int64(mostLongTime), 0).UTC()
	mostLongTime1 := n.Format("15:04")

	//вычисляем минимальную разницу "ощущаемой" и фактической температуры ночью за 5-ти дневный период

	NightTemp := make(map[float64]int)

	for i := 0; i < countDays; i++ {
		day := WeatherData.Daily[i].Datetime
		nightReal := WeatherData.Daily[i].Temp.NightTemp
		nightFeel := WeatherData.Daily[i].FeelTemp.NightFeel
		var diff float64
		switch {
		case nightFeel >= 0.0, nightReal >= 0.0:
			diff = math.Abs(nightFeel) - math.Abs(nightReal)
		case nightFeel <= 0.0, nightReal <= 0.0:
			diff = math.Abs(nightFeel - nightReal)
		case nightFeel <= 0.0, nightReal >= 0:
			diff = math.Abs(nightFeel) + nightReal
		case nightFeel >= 0, nightReal <= 0:
			diff = nightFeel + math.Abs(nightReal)
		}
		NightTemp[diff] = day
	}
	keys1 := make([]float64, 0, len(NightTemp))
	for v := range NightTemp {
		keys1 = append(keys1, v)
	}
	sort.Float64s(keys1)

	MinDiffTemp := keys1[0]

	DateMinDiffTemp1 := NightTemp[keys1[0]]
	d := time.Unix(int64(DateMinDiffTemp1), 0)
	DateMinDiffTemp := d.Format("02-01-2006")

	Today := time.Now()
	StartDate := Today.Format("02-01-2006")

	f := Today.Add(96 * time.Hour)
	FinishDate := f.Format("02-01-2006")

	fmt.Printf("\nДанные за период с %v по %v включительно\n", StartDate, FinishDate)
	fmt.Printf("\nДень, с минимальной разницей ощущаемой и фактической температуры ночью: %v,\nРазница температур: %.2f°C.\n\n", DateMinDiffTemp, MinDiffTemp)
	fmt.Printf("Максимальная продолжительность светового дня:\nДата: %v, Длительность дня: %v.\n", mostLongDay1, mostLongTime1)

	os.Remove("data.json")
}
