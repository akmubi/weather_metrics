package forecast

import (
	"encoding/json"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

var (
	// список возможных единиц измерения температуры
	Units = map[string]string{
		"standard": "K",
		"metric":   "C",
		"imperial": "F",
	}
	// список исключаемых данных о погоде
	ExcludeAll = []string{"current", "minutely", "hourly", "daily", "alerts"}
	OnlyDaily  = []string{"current", "minutely", "hourly", "alerts"}
)

// получение данных о погоде
// requestURL должен содержать адрес точки api с параметрами фильтрации
func RequestForecastInfo(requestURL string) ([]ForecastInfo, error) {
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

// Генерирует URL для последующего запроса данных погоды
// latitude - широта
// longitude - долгота
// apiKey - ключ API Open Weather
// units - единица измерения температуры.
// Возможные значения units: градусы Кельвина ("standard") ,
// градусы Цельсия ("metric") и градусы Фаренгейта ("imperial")
// language - язык, на котором будет выводиться
// результат (см. https://openweathermap.org/api/one-call-api#multi)
// exclude - данные о погоде, которые будут исключены из результата.
// Возможные значения: "current", "minutely", "hourly", "daily", "alerts"
func ConstructRequestURL(latitude, longitude float64, apiKey, units, language string, exclude []string) (requestURL string) {
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

// Проверяет unit на соответствие возможным значениям
func IsUnitValid(unit string) bool {
	_, exists := Units[unit]
	return exists
}

// Проверяет список исключаемых данных на соответствие
// возможным значениям
func IsExcludeValid(exclude []string) []bool {
	valid := make([]bool, len(exclude), cap(exclude))
	for i := range exclude {
		exists := false
		for j := range ExcludeAll {
			if exclude[i] == ExcludeAll[j] {
				exists = true
				break
			}
		}
		valid[i] = exists
	}
	return valid
}
