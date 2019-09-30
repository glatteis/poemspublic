package imggenerator

import (
	"bytes"
	"fmt"
	"image"
	"log"
	"os"
	"os/exec"
)

// GenerateImageFromHTML generates an image from the specified HTML file at the specified with and saves it in imageFile
func GenerateImageFromHTML(htmlFile *os.File, imageFile *os.File) error {
	var res *exec.Cmd
	_, err := exec.LookPath("xvfb-run")

	if err == nil {
		res = exec.Command("sh", "-c", "xvfb-run -a -s \"-screen 0 640x480x16\" wkhtmltoimage --width 384 --disable-smart-width "+htmlFile.Name()+" "+imageFile.Name())
	} else {
		res = exec.Command("wkhtmltoimage", "--width", "384", "--disable-smart-width", htmlFile.Name(), imageFile.Name())

	}

	out, err := res.CombinedOutput()
	log.Println(string(out))
	if err != nil {
		return err
	}
	return nil
}

// PNGToBinary converts a png file to a binary that the printer understands
func PNGToBinary(file *os.File) []byte {
	f, err := os.Open(file.Name())
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	defer f.Close()
	img, _, err := image.Decode(f)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	var txt bytes.Buffer
	bounds := img.Bounds()
	var bufferByte byte
	var byteIndex byte
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			r, g, b, _ := img.At(x, y).RGBA()
			var bin byte = 1
			if float64((r+g+b))/3 >= 65535/2 {
				bin = 0
			}
			bufferByte |= bin << (7 - byteIndex)
			byteIndex++
			if byteIndex == 8 {
				txt.Write([]byte{bufferByte})
				byteIndex = 0
				bufferByte = 0
			}
		}
	}
	return txt.Bytes()
}
