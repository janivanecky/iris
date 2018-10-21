package app

import (
	"image"
	"image/jpeg"
	"io/ioutil"
	"os"
	"strconv"
	"strings"
)
var maxScreenshotNum int = 0
var SCREENSHOT_DIR string = "screenshots"
func init() {
	// Create screenshot dir
	os.Mkdir(SCREENSHOT_DIR, os.ModeDir)
	files, err := ioutil.ReadDir(SCREENSHOT_DIR)
	if err != nil {
		panic(err)
	}
	for _, file := range files {
		fileNumString := strings.Split(file.Name(), ".")[0]
		fileNum, _ := strconv.Atoi(fileNumString)

		if maxScreenshotNum < fileNum {
			maxScreenshotNum = fileNum
		}
	}
}

func SaveScreenshot(img image.Image) {
	maxScreenshotNum++
	f, err := os.Create(SCREENSHOT_DIR + "/" + strconv.Itoa(maxScreenshotNum) + ".jpg")
	if err != nil {
		panic(err)
	}

	defer f.Close()
	jpeg.Encode(f, img, nil)
}