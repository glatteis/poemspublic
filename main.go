package main

import (
	"encoding/base64"
	"flag"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"strconv"
	"time"

	"github.com/glatteis/poemspublic/newsgenerator"
	"github.com/glatteis/poemspublic/poemgenerator"
	"github.com/glatteis/poemspublic/weathergenerator"
	"github.com/patrickmn/go-cache"

	_ "image/png"
)

var poemCache *cache.Cache

func init() {
	poemCache = cache.New(5*time.Minute, 10*time.Minute)
}

func getBase64(name string, fromByte int, lengthByte int,
	generator func() ([]byte, error)) ([]byte, error) {
	var err error
	var bytes []byte
	if cacheBytes, ok := poemCache.Get(name); ok {
		bytes = cacheBytes.([]byte)
	} else {
		bytes, err = generator()
	}

	if err != nil {
		return nil, err
	}

	poemCache.Set(name, bytes, cache.DefaultExpiration)

	if len(bytes) < fromByte {
		return nil, err
	}

	if len(bytes) < fromByte+lengthByte || lengthByte == 0 {
		lengthByte = len(bytes) - fromByte - 1
	}

	dataString := strconv.Itoa(len(bytes)) + "\n" + base64.StdEncoding.EncodeToString(bytes[fromByte:fromByte+lengthByte+1])
	return []byte(dataString), nil
}

func getParams(r *http.Request) (fromByte int, lengthByte int, err error) {
	from := r.FormValue("from")
	length := r.FormValue("length")

	if from != "" {
		fromByte, err = strconv.Atoi(from)
		if err != nil {
			return
		}
	}

	if length != "" {
		lengthByte, err = strconv.Atoi(length)
		if err != nil {
			return
		}
	}

	return
}

func getPoem(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	name := r.FormValue("name")

	if name == "" {
		w.WriteHeader(404)
		w.Write([]byte("name empty"))
		return
	}

	fromByte, lengthByte, err := getParams(r)
	if err != nil {
		w.WriteHeader(400)
		w.Write([]byte(err.Error()))
		return
	}

	data, err := getBase64(name, fromByte, lengthByte, func() ([]byte, error) {
		return poemgenerator.GeneratePoem(name)
	})
	if err != nil {
		w.WriteHeader(400)
		w.Write([]byte(err.Error()))
		return
	}

	w.Write(data)
}

func getNews(w http.ResponseWriter, r *http.Request) {
	fromByte, lengthByte, err := getParams(r)
	if err != nil {
		w.WriteHeader(400)
		w.Write([]byte(err.Error()))
		return
	}

	r.ParseForm()
	time := r.FormValue("time")

	if time == "" {
		w.WriteHeader(404)
		w.Write([]byte("time empty"))
		return
	}

	name := time

	data, err := getBase64(name, fromByte, lengthByte, func() ([]byte, error) {
		return newsgenerator.GenerateNews()
	})
	if err != nil {
		w.WriteHeader(400)
		w.Write([]byte(err.Error()))
		return
	}

	w.Write(data)
}

func getName(w http.ResponseWriter, r *http.Request) {
	poemList, err := ioutil.ReadDir("data/")
	if err != nil {
		log.Fatal(err)
	}

	chosenPoemFile := poemList[rand.Intn(len(poemList))]
	w.Write([]byte(chosenPoemFile.Name()))
}

func getWeather(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	locationID := r.FormValue("id")
	units := r.FormValue("units")
	language := r.FormValue("lang")

	fromByte, lengthByte, err := getParams(r)
	if err != nil {
		w.WriteHeader(400)
		w.Write([]byte(err.Error()))
		return
	}

	name := "weather_" + locationID + "_" + units + "_" + language

	data, err := getBase64(name, fromByte, lengthByte, func() ([]byte, error) {
		return weathergenerator.GenerateWeather(locationID, units, language)
	})
	if err != nil {
		w.WriteHeader(400)
		w.Write([]byte(err.Error()))
		return
	}

	w.Write(data)
}

func main() {
	portPtr := flag.Int("port", 80, "the port")
	flag.Parse()

	rand.Seed(time.Now().UTC().UnixNano())
	http.HandleFunc("/name", getName)
	http.HandleFunc("/poem", getPoem)
	http.HandleFunc("/weather", getWeather)
	http.HandleFunc("/news", getNews)
	http.ListenAndServe(":"+strconv.Itoa(*portPtr), nil)
}
