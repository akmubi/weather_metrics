package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"math"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"
)

const (
	baseURL string = "https://api.openweathermap.org/data/2.5/onecall"
	// долгота и широта Уфы
	defaultLatitude  float64 = 54.733334
	defaultLongitude float64 = 56.0
	// единица измерения - градусы Цельсия
	defaultUnit string = "metric"
)

var (
	// список возможных единиц измерения температуры
	possibleUnits = map[string]string{
		"standard": "K",
		"metric":   "C",
		"imperial": "F",
	}
	// список исключаемых данных о погоде
	onlyDaily = []string{"current", "minutely", "hourly", "alerts"}
)

func main() {
	// если ключ API присутствует в переменной окружения, берём его оттуда
	apiKey, apiKeyEnvExists := os.LookupEnv("OPENWEATHER_API_KEY")
	flagAPIKeyflag := flag.String("api_key", "",
		"Ключ API от сервиса Open Weather. Если этот флаг указан, значение ключа указанное в переменной среде игнорируется")

	latitude := flag.Float64("latitude", defaultLatitude, "Широта (географическая координата)")
	longitude := flag.Float64("longitude", defaultLongitude, "Долгота (географическая координата")
	userUnit := flag.String("units", defaultUnit, "Единица измерения температуры. Возможные значения standard (градусы Кельвина), metric (градусы Цельсия), imperial (градусы Фаренгейта)")
	flag.Parse()

	// если ключ API указан через флаг,
	// то ключ в переменной среды игнорируется
	if len(*flagAPIKeyflag) > 0 {
		apiKey = *flagAPIKeyflag
	} else if !apiKeyEnvExists {
		log.Fatal("Вы должны указать ключ API от сервиса Open Weater. Это можно сделать либо с помощью флага -api_key или с помощью переменной среды OPENWEATHER_API_KEY")
	}

	// проверяем на соответствие возможным единицам измерения
	if _, exists := possibleUnits[*userUnit]; !exists {
		log.Fatal("Единица измерения температуры не соответствует ни однму из возможных значений (standard, metric, imperial)")
	}

	// конструируем адрес запроса с указанием параметров расположения, ключа API, единицы измерения
	// и также указываем список данных, которые нам не нужны
	requestURL := constructRequestURL(*latitude, *longitude, apiKey, *userUnit, "", onlyDaily)
	// делаем запрос и получаем массив данных по каждому дню
	forecastInfo, err := requestForecastInfo(requestURL)
	if err != nil {
		log.Fatal(err)
	}

	// за минимальное берем разницу между температурами первого дня
	minDiff := float64(10000)
	// индекс записи (дня), в котором наблюдалась минимальная разница между температурами
	minDiffDayIndex := -1

	// проходимся по каждому дню и находим разницу между "ощущаемой" и фактической температурой
	for i, info := range forecastInfo {
		diff := math.Abs(info.FeelsLike.Value - info.Temp.Value)
		if diff < minDiff {
			minDiff = diff
			minDiffDayIndex = i
		}
	}
	fmt.Println(`День с минимальной разницей между фактической и "ощущаемой" температурой (`+
		strconv.FormatFloat(minDiff, 'f', 3, 64)+" "+possibleUnits[*userUnit]+
		`):`,
		forecastInfo[minDiffDayIndex].DateTime.Format("02/01/2006"))

	// максимальная продолжительность дня
	var maxDuration time.Duration
	// индекс дня с максимальной продолжительностью
	maxDurationDayIndex := -1

	// выбираем первые 5 дней (включая текущий)
	for i, info := range forecastInfo[:5] {
		// вычисляем разницу между временем заката и временем рассвета
		duration := info.SunsetTime.Sub(info.SunriseTime.Time)
		if duration > maxDuration {
			maxDuration = duration
			maxDurationDayIndex = i
		}
	}
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

func constructRequestURL(latitude, longitude float64, apiKey, units, language string, exclude []string) (requestURL string) {
	// создаем карты с обязательными и необязательными параметрами
	requiredParams := map[string]string{
		"lat":   strconv.FormatFloat(latitude, 'f', 3, 32),
		"lon":   strconv.FormatFloat(longitude, 'f', 3, 32),
		"appid": apiKey,
	}
	optionalParams := map[string]string{
		"units":   units,
		"lang":    language,
		"exclude": strings.Join(exclude, ","),
	}

	// добавляем все параметры с их значениями в url.Values
	v := make(url.Values)
	for param, value := range requiredParams {
		v.Set(param, value)
	}
	for param, value := range optionalParams {
		// если значение пустое, то параметр не включается
		if len(value) > 0 {
			v.Set(param, value)
		}
	}
	// кодируем значения и склеиваем
	return baseURL + "?" + v.Encode()
}

// среди значений температуры нужна только температура ночью
type Temperature struct {
	Value float64 `json:"night"`
}

// из данных о погоде на день нужны:
// дата и время, дата и время рассвета и заката,
// и фактическая и "ощущаемая" температуры
type ForecastInfo struct {
	DateTime    UnixTime    `json:"dt"`
	SunriseTime UnixTime    `json:"sunrise"`
	SunsetTime  UnixTime    `json:"sunset"`
	Temp        Temperature `json:"temp"`
	FeelsLike   Temperature `json:"feels_like"`
}

// вспомогательная структура для преобразования
// целочисленного значения в дату и время
type UnixTime struct {
	time.Time
}

// реализация интерфейса десериализации JSON данных
// в дату и время формата Unix UTC
func (u *UnixTime) UnmarshalJSON(data []byte) error {
	timestamp := int64(0)
	if err := json.Unmarshal(data, &timestamp); err != nil {
		return err
	}
	u.Time = time.Unix(timestamp, 0)
	return nil
}

// реализация интерфейса для вывода данных структуры в удобном виде
func (info ForecastInfo) String() string {
	var builder strings.Builder
	builder.WriteString("Дата и время: ")
	builder.WriteString(info.DateTime.String())
	builder.WriteString("\nВремя рассвета: ")
	builder.WriteString(info.SunriseTime.String())
	builder.WriteString("\nВремя заката: ")
	builder.WriteString(info.SunsetTime.String())
	builder.WriteString("\nФактическая температура: ")
	builder.WriteString(strconv.FormatFloat(info.Temp.Value, 'f', 3, 64))
	builder.WriteString("\nОщущаемая температура: ")
	builder.WriteString(strconv.FormatFloat(info.FeelsLike.Value, 'f', 3, 64))
	builder.WriteRune('\n')
	return builder.String()
}

// получение данных о погоде
// requestURL должен содержать адрес точки api с параметрами фильтрации
func requestForecastInfo(requestURL string) ([]ForecastInfo, error) {
	response, err := http.Get(requestURL)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	// вспомогательная структура для излечения информации всех дней
	info := new(struct {
		Daily []ForecastInfo `json:"daily"`
	})

	if err := json.NewDecoder(response.Body).Decode(info); err != nil {
		return nil, err
	}
	return info.Daily, nil
}
