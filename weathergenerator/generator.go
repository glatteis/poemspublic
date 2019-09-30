package weathergenerator

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path"
	"text/template"

	"github.com/glatteis/poemspublic/imggenerator"
)

var weatherTemplate *template.Template

func init() {
	var err error
	weatherTemplate, err = template.New(path.Base("weather_template.html")).ParseFiles("weather_template.html")
	if err != nil {
		log.Fatal(err)
	}
}

// GenerateWeather renders the weather template from weather data
func GenerateWeather(locationID string, units string, language string) ([]byte, error) {
	values, err := getWeatherInfo(locationID, units, language)
	if err != nil {
		return nil, err
	}

	cwd, err := os.Getwd()
	if err != nil {
		return nil, err
	}

	values.CurrentWorkingDir = cwd

	var buffer bytes.Buffer

	err = weatherTemplate.Execute(&buffer, *values)
	if err != nil {
		log.Fatal(err)
	}

	tempFile, err := ioutil.TempFile("/tmp", "weather*.html")
	if err != nil {
		log.Fatal(err)
	}

	tempFile.Write(buffer.Bytes())

	tempImage, err := ioutil.TempFile("/tmp", "weather*.png")
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(tempImage.Name())

	err = imggenerator.GenerateImageFromHTML(tempFile, tempImage)

	bytes := imggenerator.PNGToBinary(tempImage)
	return bytes, nil
}
