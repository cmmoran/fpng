package main

import (
	"bytes"
	"fmt"
	"image"
	"image/png"
	"io/ioutil"
	"math"
	"os"
	"regexp"
)

var (
	helpflag = false
	infile   = ""
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
	helpflag = len(os.Args) == 1 || os.Args[1] == "-h"
	if !helpflag && len(os.Args) > 0 {
		infile = os.Args[1]
	}
}

func decode(fName string, oName string) {
	data, _ := ioutil.ReadFile(fName)

	file, err := os.Stat(fName)

	if err != nil {
		fmt.Println("y u no file?")
		return
	}

	fTime := file.ModTime()

	bufreader := bytes.NewReader(data)

	imageData, err := png.Decode(bufreader)

	if err != nil {
		fmt.Println("Cannot read image")
		Usage()
		return
	}

	nrgba := imageData.(*image.NRGBA)

	filelength := (int32(nrgba.Pix[0]) << 24) | (int32(nrgba.Pix[1]) << 16) | (int32(nrgba.Pix[2]) << 8) | int32(nrgba.Pix[3])

	ndata := make([]byte, filelength)

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

	err = os.Chtimes(oName, fTime, fTime)

	if err != nil {
		fmt.Println("y u no modtime?")
		return
	}
}

func Usage() {
	fmt.Printf("Can only send images through slack huh?\n\033[37mfpng\033[0m <\033[36minfile\033[0m>\n" +
		"     <infile>: png file -> decode to original\n" +
		"     <infile>: data file -> encode to <infile>.png\n")
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
			myImage.Pix[ii] = uint8(isize>>24) & 0xFF
			ii++
			myImage.Pix[ii] = uint8(isize>>16) & 0xFF
			ii++
			myImage.Pix[ii] = uint8(isize>>8) & 0xFF
			ii++
			myImage.Pix[ii] = uint8(isize) & 0xFF
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
