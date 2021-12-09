package forecast

import (
	"encoding/json"
	"math"
	"strconv"
	"strings"
	"time"
)

const (
	baseURL string = "https://api.openweathermap.org/data/2.5/onecall"
)

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

// Нахождение минимальной разницы между "ощущаемой" и фактической температурой
// Возвращает минимальную разницу и индекс соответствующего дня
func MinTemperatureDifference(info []ForecastInfo) (minDiff float64, dayIndex int) {
	// поскольку это разница температур, то за начальное минимальное значение
	// можно взять 10000
	minDiff = float64(10000)
	// индекс записи (дня), в котором наблюдалась минимальная разница между температурами
	minDiffDayIndex := -1

	// проходимся по каждому дню и находим разницу между "ощущаемой" и фактической температурой
	for i, info := range info {
		diff := math.Abs(info.FeelsLike.Value - info.Temp.Value)
		if diff < minDiff {
			minDiff = diff
			minDiffDayIndex = i
		}
	}
	return minDiff, minDiffDayIndex
}

// Возращает максимальную продолжительность дня (разницу между временем заката и рассвета)
// и индекс соответствующего дня
func MaxDayDuration(info []ForecastInfo) (maxDuration time.Duration, dayIndex int) {
	// индекс дня с максимальной продолжительностью
	maxDurationDayIndex := -1

	for i, info := range info {
		// вычисляем разницу между временем заката и временем рассвета
		duration := info.SunsetTime.Sub(info.SunriseTime.Time)
		if duration > maxDuration {
			maxDuration = duration
			maxDurationDayIndex = i
		}
	}
	return maxDuration, maxDurationDayIndex
}
