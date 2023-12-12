package main

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/md5"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"golang.org/x/crypto/ssh/terminal"
	"image"
	"image/png"
	"io"
	"io/ioutil"
	"math"
	"os"
	"regexp"
)

var (
	helpflag = false
	crypt    = false
	bcrypt   = ""
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
	if helpflag {
		return
	}
	if !helpflag && len(os.Args) > 0 {
		infile = os.Args[1]
	}
	match, _ := regexp.MatchString("\\.png$", infile)
	if !helpflag && !match {
		readCrypt()
	}
}

func readCrypt() {
	fmt.Print("Enter passphrase: ")
	btext, err := terminal.ReadPassword(int(os.Stdin.Fd()))
	if err != nil {
		panic(err.Error())
	}
	text := string(btext)
	bcrypt = createHash(text)
	fmt.Println()
	crypt = len(text) > 0
}

func Usage() {
	fmt.Printf("Hate the message: \033[2m\"At the request of your administrator, only images can be uploaded to this workspace.\"\033[0m?\n" +
		"fpng is a codec for any arbitrary file to and from png format. Great for sharing arbitrary files where only images are allowed to be shared.\n\n" +
		"" +
		"\033[37mfpng\033[0m <\033[36minfile\033[0m>\n" +
		"     <\033[36minfile\033[0m>: png file -> decode to original\n" +
		"     <\033[36minfile\033[0m>: data file -> encode to <infile>.png\n" +
		"\033[1mNOTE\033[0m: [] = optional, <> = required, empty passphrase disables encrypted encoding.\n")

}

/*
Begin https://www.thepolyglotdeveloper.com/2018/02/encrypt-decrypt-data-golang-application-crypto-packages/
*/
func createHash(key string) string {
	hasher := md5.New()
	hasher.Write([]byte(key))
	return hex.EncodeToString(hasher.Sum(nil))
}

func encrypt(data []byte, passphrase string) []byte {
	block, _ := aes.NewCipher([]byte(createHash(passphrase)))
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		panic(err.Error())
	}
	nonce := make([]byte, gcm.NonceSize())
	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		panic(err.Error())
	}
	ciphertext := gcm.Seal(nonce, nonce, data, nil)
	return ciphertext
}

func decrypt(data []byte, passphrase string) []byte {
	key := []byte(createHash(passphrase))
	block, err := aes.NewCipher(key)
	if err != nil {
		panic(err.Error())
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		panic(err.Error())
	}
	nonceSize := gcm.NonceSize()
	nonce, ciphertext := data[:nonceSize], data[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		panic(err.Error())
	}
	return plaintext
}

/*
End https://www.thepolyglotdeveloper.com/2018/02/encrypt-decrypt-data-golang-application-crypto-packages/
*/

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

	dcrypt := nrgba.Pix[0] == uint8(1)

	filelength := (int32(nrgba.Pix[1]) << 24) | (int32(nrgba.Pix[2]) << 16) | (int32(nrgba.Pix[3]) << 8) | int32(nrgba.Pix[4])

	ndata := make([]byte, filelength)

	ndata = nrgba.Pix[5 : filelength+5]

	outputFile, err := os.Create(oName)
	if err != nil {
		fmt.Println("y u no output?")
		return
	}

	enc := ""
	if !crypt && dcrypt {
		readCrypt()
		crypt = true
	}

	if crypt {
		ndata = decrypt(ndata, bcrypt)
		enc = "(Encrypted) "
	}
	_, _ = outputFile.Write(ndata)
	_ = outputFile.Sync()
	_ = outputFile.Close()

	fmt.Printf("Decoded %s%s -> %s\n", enc, fName, oName)

	err = os.Chtimes(oName, fTime, fTime)

	if err != nil {
		fmt.Println("y u no modtime?")
		return
	}
}

func encode(fName string, oName string) {
	dat, err := ioutil.ReadFile(fName)

	if err != nil {
		fmt.Println("Input file cannot be empty")
		Usage()
		return
	}
	enc := ""
	icrypt := uint8(0)
	if crypt {
		dat = encrypt(dat, bcrypt)
		icrypt = uint8(1)
		enc = "(Encrypted) "
	}

	size := len(dat)
	isize := int32(size)

	addl := isize / 4
	if isize%4 > 0 {
		addl = isize/4 + 1
	}

	imageSize := int(math.Ceil(math.Sqrt(float64(addl))))

	img := image.NewNRGBA(image.Rect(0, 0, imageSize+1, imageSize+1))

	ii := 0
	encode8(&ii, img.Pix, icrypt)
	encode32(&ii, img.Pix, isize)
	img.Pix = append(img.Pix[:ii], append(dat, img.Pix[ii:]...)...)

	outputFile, err := os.Create(oName)

	if err != nil {
		fmt.Println("y u no output?")
		return
	}

	_ = png.Encode(outputFile, img)

	_ = outputFile.Close()

	fmt.Printf("Encoded %s%s -> %s\n", enc, fName, oName)
}

func encode8(index *int, data []uint8, value uint8) {
	data[*index] = value
	*index = *index + 1
}

func encode32(index *int, data []uint8, value int32) {
	data[*index] = uint8(value>>24) & 0xFF
	*index = *index + 1
	data[*index] = uint8(value>>16) & 0xFF
	*index = *index + 1
	data[*index] = uint8(value>>8) & 0xFF
	*index = *index + 1
	data[*index] = uint8(value) & 0xFF
	*index = *index + 1
}
