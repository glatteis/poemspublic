// Package poemgenerator is the poem generator domain
package poemgenerator

import (
	"bytes"
	"encoding/json"
	"errors"
	"html/template"
	"io/ioutil"
	"log"
	"os"
	"path"
	"strings"

	"github.com/glatteis/poemspublic/imggenerator"
)

type savedPoem struct {
	Poem          string `json:"poem"`
	PoemHTML      template.HTML
	Title         string   `json:"title"`
	Author        string   `json:"author"`
	YearWritten   string   `json:"year_written"`
	YearPublished string   `json:"year_published"`
	Origin        string   `json:"origin"`
	References    []string `json:"references"`
}

var poemTemplate *template.Template

func init() {
	var err error
	poemTemplate, err = template.New(path.Base("poem_template.html")).ParseFiles("poem_template.html")
	if err != nil {
		log.Fatal(err)
	}
}

// GeneratePoem generates a poem bitmap and returns it
func GeneratePoem(name string) ([]byte, error) {
	file, err := os.Open(path.Join("data/", name))
	if err != nil {
		return nil, errors.New("File does not exist")
	}
	defer file.Close()

	b, err := ioutil.ReadAll(file)

	// Render the template
	var poem savedPoem
	json.Unmarshal(b, &poem)

	poem.Poem = strings.Replace(poem.Poem, "\n", "<br>\n", -1)
	poem.PoemHTML = template.HTML(poem.Poem)

	var buffer bytes.Buffer

	err = poemTemplate.Execute(&buffer, poem)
	if err != nil {
		log.Fatal(err)
	}

	tempFile, err := ioutil.TempFile("/tmp", "poem*.html")
	if err != nil {
		log.Fatal(err)
	}

	tempFile.Write(buffer.Bytes())

	tempImage, err := ioutil.TempFile("/tmp", "poem*.png")
	if err != nil {
		log.Fatal(err)
	}

	log.Println(tempFile.Name())
	log.Println(tempImage.Name())

	err = imggenerator.GenerateImageFromHTML(tempFile, tempImage)

	bytes := imggenerator.PNGToBinary(tempImage)
	return bytes, nil
}
