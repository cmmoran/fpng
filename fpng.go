package main

import (
	"bytes"
	"fmt"
	"github.com/pborman/getopt"
	"image"
	"image/png"
	"io/ioutil"
	"math"
	"os"
	"regexp"
)

var (
	infile = ""
	helpflag = false
)

func main() {
	if helpflag {
		Usage()
		os.Exit(0)
	}
	bfName := []byte(infile)
	match, _ := regexp.MatchString("\\.png$", infile)
	if match {
		r, _ := regexp.Compile("\\.png$")
		bfName = r.ReplaceAll([]byte(infile), []byte(""))
		oFile := string(bfName)
		decode(infile, oFile)
	} else {
		oFile := infile + ".png"
		encode(infile, oFile)
	}
}

func init() {
	getopt.BoolVar(&helpflag, 'h', "halp")
	getopt.StringVar(&infile, 'i', "the file to [en|de]code")

	getopt.Parse()
	err := getopt.Getopt(nil)
	if err != nil {
		// code to handle error
		_, _ = fmt.Fprintln(os.Stderr, err)
		return
	}
}

func decode(fName string, oName string) {
	data, _ := ioutil.ReadFile(fName)

	bufreader := bytes.NewReader(data)

	imageData, err := png.Decode(bufreader)

	if err != nil {
		fmt.Println("Cannot read image")
		Usage()
		return
	}

	_r, _g, _b, _a := imageData.At(0, 0).RGBA()
	var a = _a & (_a >> 8)

	var r = ((_r << 8) / (_a & (_a >> 8))) >> 8
	var g = ((_g << 8) / (_a & (_a >> 8))) >> 8
	var b = ((_b << 8) / (_a & (_a >> 8))) >> 8

	var filelength = ((r & 0xFF) << 24) | ((g & 0xFF) << 16) | ((b & 0xFF) << 8) | a

	ndata := make([]byte, filelength)

	nrgba := imageData.(*image.NRGBA)
	ndata = nrgba.Pix[4 : filelength+4]

	outputFile, err := os.Create(oName)
	if err != nil {
		fmt.Println("y u no output?")
		return
	}

	_, _ = outputFile.Write(ndata)

	_ = outputFile.Sync()

	_ = outputFile.Close()
	fmt.Printf("Decoded %s -> %s\n", fName, oName)
}

func Usage() {
	fmt.Printf("Can only send images through slack huh?\n\033[37mfpng\033[0m -i <\033[36minfile\033[0m>\n" +
		"\t<infile>: png file -> decode to original\n" +
		"\t<infile>: data file -> encode to <infile>.png\n")
}

func encode(fName string, oName string) {
	dat, err := ioutil.ReadFile(fName)

	if err != nil {
		fmt.Println("Input file cannot be empty")
		Usage()
		return
	}

	size := len(dat)
	isize := int32(size)

	addl := isize / 4
	if isize%4 > 0 {
		addl = isize/4 + 1
	}

	imageSize := int(math.Ceil(math.Sqrt(float64(addl))))

	myImage := image.NewNRGBA(image.Rect(0, 0, imageSize+1, imageSize+1))

	ii := 0

	for i := 0; i < size; i++ {
		if i == 0 {
			myImage.Pix[ii] = uint8(isize >> 24)
			ii++
			myImage.Pix[ii] = uint8(isize >> 16)
			ii++
			myImage.Pix[ii] = uint8(isize >> 8)
			ii++
			myImage.Pix[ii] = uint8(isize)
			ii++
		}
		myImage.Pix[ii] = dat[i]
		ii++
	}
	outputFile, err := os.Create(oName)

	if err != nil {
		fmt.Println("y u no output?")
		return
	}

	_ = png.Encode(outputFile, myImage)

	_ = outputFile.Close()
	fmt.Printf("Decoded %s -> %s\n", fName, oName)
}
