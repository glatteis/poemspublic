package weathergenerator

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/goodsign/monday"
	"github.com/wcharczuk/go-chart"
	"github.com/wcharczuk/go-chart/drawing"
)

type weatherConfig struct {
	OpenWeatherAppID   string
	StandardLocationID string
	StandardUnitType   string
	StandardLanguage   string
}

type forecast struct {
	DateTime         int64 `json:"dt"`
	DateTimeReadable string
	DateTimeGolang   time.Time
	TempValues       struct {
		RoundedTemperature int
		Temperature        float32 `json:"temp"`
		MinTemperature     float32 `json:"temp_min"`
		MaxTemperature     float32 `json:"temp_max"`
		Humidity           float32 `json:"humidity"`
	} `json:"main"`
	WeatherTypes []struct {
		Description string `json:"description"`
		IconID      string `json:"icon"`
	} `json:"weather"`
	Clouds struct {
		OvercastPercentage float32 `json:"all"`
	} `json:"clouds"`
	Wind struct {
		WindSpeed float32 `json:"speed"`
		WindAngle float32 `json:"deg"`
	} `json:"wind"`
	Rain struct {
		RainVolume float32 `json:"3h"`
	} `json:"rain"`
	Snow struct {
		SnowVolume float32 `json:"3h"`
	} `json:"snow"`
}

type weatherAPIValues struct {
	Forecasts         []forecast `json:"list"`
	SelectedForecasts []forecast
	Charts            []string
	CurrentWorkingDir string // For the template
}

var config weatherConfig

const url = "https://api.openweathermap.org/data/2.5/forecast?mode=json&id=%s&appid=%s&units=%s&lang=%s"

func init() {
	// load config file
	file, err := os.Open("weather_config.toml")
	if err != nil {
		log.Fatal(errors.New("File does not exist"))
	}
	defer file.Close()

	b, err := ioutil.ReadAll(file)
	if _, err := toml.Decode(string(b), &config); err != nil {
		log.Fatal(err)
	}
}

// GetWeather gets a weather bitmap
func getWeatherInfo(locationID string, units string, language string) (*weatherAPIValues, error) {
	if locationID == "" {
		locationID = config.StandardLocationID
	}
	if units == "" {
		units = config.StandardUnitType
	}
	if language == "" {
		language = config.StandardLanguage
	}
	requestURL := fmt.Sprintf(url, locationID, config.OpenWeatherAppID, units, language)

	fmt.Println(requestURL)

	response, err := http.Get(requestURL)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}
	if response.StatusCode != 200 {
		return nil, errors.New("Invalid return code: " + strconv.Itoa(response.StatusCode) + ", with response: " + string(body))
	}

	var values weatherAPIValues

	json.Unmarshal(body, &values)

	numForecasts := len(values.Forecasts)

	dateTimes := make([]time.Time, numForecasts)
	minTemps := make([]float64, numForecasts)
	maxTemps := make([]float64, numForecasts)
	humidities := make([]float64, numForecasts)
	rains := make([]float64, numForecasts)

	for i, f := range values.Forecasts {
		dateTimes[i] = time.Unix(f.DateTime, 0)
		minTemps[i] = float64(f.TempValues.MinTemperature)
		maxTemps[i] = float64(f.TempValues.MaxTemperature)
		humidities[i] = float64(f.TempValues.Humidity)
		rains[i] = float64(f.Rain.RainVolume)
	}

	var tempUnit string
	switch units {
	case "metric":
		tempUnit = "°C"
	case "standard":
		tempUnit = "K"
	case "imperial":
		tempUnit = "°F"
	}

	lineStyle := chart.Style{
		StrokeWidth: 3,
		DotColor:    drawing.ColorBlack,
		StrokeColor: drawing.ColorBlack,
		Show:        true,
	}

	var graphFiles [4]*os.File
	for i := 0; i < 4; i++ {
		tempFile, err := ioutil.TempFile("/tmp", "chart*.png")
		if err != nil {
			return nil, err
		}

		graphFiles[i] = tempFile
		values.Charts = append(values.Charts, tempFile.Name())

		defer tempFile.Close()
	}

	chartWidth := 375
	dateFormatter := func(x interface{}) string {
		return monday.Format(time.Unix(0, int64(x.(float64))), "Mon 02.01 15:04", monday.LocaleDeDE)
	}
	timeFormatter := func(x interface{}) string {
		return monday.Format(time.Unix(0, int64(x.(float64))), "Mon 15:04", monday.LocaleDeDE)
	}

	temperatureGraph := chart.Chart{
		XAxis: chart.XAxis{
			Style:          chart.StyleShow(),
			ValueFormatter: dateFormatter,
		},
		YAxis: chart.YAxis{
			Name:      tempUnit,
			NameStyle: chart.StyleShow(),
			Style:     chart.StyleShow(),
		},
		Series: []chart.Series{
			chart.TimeSeries{
				XValues: dateTimes,
				YValues: minTemps,
				Style:   lineStyle,
			},
			chart.TimeSeries{
				XValues: dateTimes,
				YValues: maxTemps,
				Style:   lineStyle,
			},
		},
		Width:  chartWidth,
		Height: 200,
	}

	buffer := bytes.NewBuffer([]byte{})
	err = temperatureGraph.Render(chart.PNG, buffer)
	if err != nil {
		return nil, err
	}
	graphFiles[0].Write(buffer.Bytes())

	temperatureGraph24h := chart.Chart{
		XAxis: chart.XAxis{
			Style:          chart.StyleShow(),
			ValueFormatter: timeFormatter,
		},
		YAxis: chart.YAxis{
			Style:     chart.StyleShow(),
			Name:      tempUnit,
			NameStyle: chart.StyleShow(),
		},
		Series: []chart.Series{
			chart.TimeSeries{
				XValues: dateTimes[:9],
				YValues: minTemps[:9],
				Style:   lineStyle,
			},
			chart.TimeSeries{
				XValues: dateTimes[:9],
				YValues: maxTemps[:9],
				Style:   lineStyle,
			},
		},
		Width:  chartWidth,
		Height: 200,
	}
	buffer = bytes.NewBuffer([]byte{})
	err = temperatureGraph24h.Render(chart.PNG, buffer)
	if err != nil {
		return nil, err
	}
	graphFiles[1].Write(buffer.Bytes())

	rainGraph := chart.Chart{
		XAxis: chart.XAxis{
			Style:          chart.StyleShow(),
			ValueFormatter: dateFormatter,
		},
		YAxis: chart.YAxis{
			Name:      "mm",
			NameStyle: chart.StyleShow(),
			Style:     chart.StyleShow(),
		},
		Series: []chart.Series{
			chart.TimeSeries{
				XValues: dateTimes,
				YValues: rains,
				Style:   lineStyle,
			},
		},
		Width:  chartWidth,
		Height: 200,
	}
	buffer = bytes.NewBuffer([]byte{})
	err = rainGraph.Render(chart.PNG, buffer)
	if err != nil {
		return nil, err
	}
	graphFiles[2].Write(buffer.Bytes())

	humidityGraph := chart.Chart{
		XAxis: chart.XAxis{
			Style:          chart.StyleShow(),
			ValueFormatter: dateFormatter,
		},
		YAxis: chart.YAxis{
			Name:      "%",
			NameStyle: chart.StyleShow(),
			Style:     chart.StyleShow(),
		},
		Series: []chart.Series{
			chart.TimeSeries{
				XValues: dateTimes,
				YValues: humidities,
				Style:   lineStyle,
			},
		},
		Width:  chartWidth,
		Height: 200,
	}
	buffer = bytes.NewBuffer([]byte{})
	err = humidityGraph.Render(chart.PNG, buffer)
	if err != nil {
		return nil, err
	}
	graphFiles[3].Write(buffer.Bytes())

	for i := range values.Forecasts {
		f := &values.Forecasts[i]
		f.DateTimeGolang = time.Unix(f.DateTime, 0)
		f.DateTimeReadable =
			monday.Format(f.DateTimeGolang, "Monday, 15:04", monday.LocaleDeDE)
		f.TempValues.RoundedTemperature = int(f.TempValues.Temperature)
	}

	forecastLength := 4
	forecastsAdded := 0
	values.SelectedForecasts = make([]forecast, forecastLength)
	for _, f := range values.Forecasts {
		if forecastLength == forecastsAdded {
			break
		}
		if f.DateTimeGolang.Hour() == 8 || f.DateTimeGolang.Hour() == 14 || f.DateTimeGolang.Hour() == 20 {
			values.SelectedForecasts[forecastsAdded] = f
			forecastsAdded++
		}
	}

	return &values, nil
}
