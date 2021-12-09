package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"./forecast"
)

const (
	// долгота и широта Уфы
	defaultLatitude  float64 = 54.733334
	defaultLongitude float64 = 56.0
	// единица измерения - градусы Цельсия
	defaultUnit string = "metric"
)

func main() {
	// если ключ API присутствует в переменной окружения, берём его оттуда
	apiKey, apiKeyEnvExists := os.LookupEnv("OPENWEATHER_API_KEY")
	flagAPIKey := flag.String("api_key", "",
		"Ключ API от сервиса Open Weather. Если этот флаг указан, значение ключа указанное в переменной среде игнорируется")

	latitude := flag.Float64("latitude", defaultLatitude, "Широта (географическая координата)")
	longitude := flag.Float64("longitude", defaultLongitude, "Долгота (географическая координата")
	userUnit := flag.String("units", defaultUnit, "Единица измерения температуры. Возможные значения standard (градусы Кельвина), metric (градусы Цельсия), imperial (градусы Фаренгейта)")
	flag.Parse()

	// если ключ API указан через флаг,
	// то ключ в переменной среды игнорируется
	if len(*flagAPIKey) > 0 {
		apiKey = *flagAPIKey
	} else if !apiKeyEnvExists {
		log.Fatal("Вы должны указать ключ API от сервиса Open Weater. Это можно сделать либо с помощью флага -api_key или с помощью переменной среды OPENWEATHER_API_KEY")
	}

	// проверяем на соответствие возможным единицам измерения
	if forecast.IsUnitValid(*userUnit) {
		log.Fatal("Единица измерения температуры не соответствует ни однму из возможных значений (standard, metric, imperial)")
	}

	// конструируем адрес запроса с указанием параметров расположения, ключа API, единицы измерения
	// и также указываем список данных, которые нам не нужны
	requestURL := forecast.ConstructRequestURL(
		*latitude,
		*longitude,
		apiKey,
		*userUnit,
		"",
		forecast.OnlyDaily)

	// делаем запрос и получаем массив данных по каждому дню
	forecastInfo, err := forecast.RequestForecastInfo(requestURL)
	if err != nil {
		log.Fatal(err)
	}

	minDiff, minDiffDayIndex := forecast.MinTemperatureDifference(forecastInfo)
	fmt.Println(`День с минимальной разницей между фактической и "ощущаемой" температурой (`+
		strconv.FormatFloat(minDiff, 'f', 3, 64)+" "+forecast.Units[*userUnit]+
		`):`,
		forecastInfo[minDiffDayIndex].DateTime.Format("02/01/2006"))

	// находим максимальную длительность дня среди первых 5 дней
	maxDuration, maxDurationDayIndex := forecast.MaxDayDuration(forecastInfo[:5])

	fmt.Println("День с максимальной продолжительностью дня ("+
		formatDuration(maxDuration)+
		"): ",
		forecastInfo[maxDurationDayIndex].DateTime.Format("02/01/2006"))
}

// для форматирования продолжительности времени
func formatDuration(duration time.Duration) string {
	duration = duration.Round(time.Second)
	hours := duration / time.Hour
	duration -= hours * time.Hour
	minutes := duration / time.Minute
	duration -= minutes * time.Minute
	seconds := duration / time.Second
	return fmt.Sprintf("%02d:%02d:%02d", hours, minutes, seconds)
}
