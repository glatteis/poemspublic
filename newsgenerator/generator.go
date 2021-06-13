// Package newsgenerator is the news generator domain
package newsgenerator

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"path"
	"strings"
	"text/template"

	"github.com/glatteis/poemspublic/imggenerator"
	"github.com/mmcdole/gofeed"
)

const feed = "https://www.tagesschau.de/xml/rss2/"

var newsTemplate *template.Template

func init() {
	var err error
	newsTemplate, err = template.New(path.Base("news_template.html")).ParseFiles("news_template.html")
	if err != nil {
		log.Fatal(err)
	}
}

// GenerateNews generates a news bitmap and returns it
func GenerateNews() ([]byte, error) {
	fp := gofeed.NewParser()
	feed, err := fp.ParseURL(feed)
	if err != nil {
		return nil, err
	}

	feed.Items = feed.Items[:7]
	for _, item := range feed.Items {
		split := strings.Split(item.Content, "\n")
		item.Content = strings.Join(split[:len(split)-2], "\n")
	}

	var buffer bytes.Buffer

	err = newsTemplate.Execute(&buffer, feed)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(feed)
	fmt.Println(buffer.String())

	tempFile, err := ioutil.TempFile("/tmp", "news*.html")
	if err != nil {
		log.Fatal(err)
	}

	tempFile.Write(buffer.Bytes())

	tempImage, err := ioutil.TempFile("/tmp", "news*.png")
	if err != nil {
		log.Fatal(err)
	}

	log.Println(tempFile.Name())
	log.Println(tempImage.Name())

	err = imggenerator.GenerateImageFromHTML(tempFile, tempImage)

	if err != nil {
		return nil, err
	}

	bytes := imggenerator.PNGToBinary(tempImage)
	return bytes, nil
}
