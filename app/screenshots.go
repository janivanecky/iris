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
const screenshotDir = "screenshots"
const screenshotExtension = "jpg"

func init() {
	// Create screenshot dir if doesn't exist yet.
	os.Mkdir(screenshotDir, 0700)

	// Get the biggest number in screenshot names,
	// next screenshot will be that number plus one.
	files, err := ioutil.ReadDir(screenshotDir)
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

// SaveScreenshot saves image to screenshot folder.
func SaveScreenshot(img image.Image) {
	// Increment screenshot counter to be used as a name.
	maxScreenshotNum++

	// Create target file.
	screenshotPath := screenshotDir + "/" + strconv.Itoa(maxScreenshotNum) + "." + screenshotExtension
	f, err := os.Create(screenshotPath)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	// Save image into file.
	jpeg.Encode(f, img, nil)
}