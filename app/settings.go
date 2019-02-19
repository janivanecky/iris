package app

import (
	"encoding/json"
	"io/ioutil"
	"math"
	"os"
	"sort"
	"strconv"
	"strings"
	
	"github.com/go-gl/mathgl/mgl32"

) 

type CellSettings struct {
	PolarStd, PolarMean    float64
	RadiusMin, RadiusMax   float64
	HeightRatio            float64
	Count				   int
	Colors        		   []mgl32.Vec4
}

type CameraSettings struct {
	Radius, Azimuth, Polar, Height float64
}

type AppSettings struct {
	Cells         CellSettings
	Rendering     RenderingSettings
	Camera        CameraSettings
	id 			  int
}

type RenderingSettings struct {
	DirectLight  float64
	AmbientLight float64

	Roughness    float64
	Reflectivity float64

	SSAORadius   float64
	SSAORange    float64
	SSAOBoundary float64

	MinWhite float64
}

func copySettings(settings *AppSettings) AppSettings {
	newSettings := AppSettings{}
	newSettings.Cells = settings.Cells
	newSettings.Rendering = settings.Rendering
	newSettings.Camera = settings.Camera
	newSettings.Cells.Colors = make([]mgl32.Vec4, len(settings.Cells.Colors))
	copy(newSettings.Cells.Colors, settings.Cells.Colors)
	return newSettings
}

var defaultSettings = AppSettings{
	Cells: CellSettings{
		PolarStd: 0.00, PolarMean: math.Pi / 2.0,
		RadiusMin: 3.0, RadiusMax: 15.0,
		HeightRatio: 1.0,
		Count: 5000,
		Colors: []mgl32.Vec4{
			mgl32.Vec4{24 / 255.0, 193 / 255.0, 236 / 255.0, 1.0},
			mgl32.Vec4{0 / 255.0, 185 / 255.0, 121 / 255.0, 1.0},
			mgl32.Vec4{236 / 255.0, 24 / 255.0, 97 / 255.0, 1.0},
			mgl32.Vec4{33 / 255.0, 73 / 255.0, 83 / 255.0, 1.0},
			mgl32.Vec4{194 / 255.0, 55 / 255.0, 48 / 255.0, 1.0},
		},
	},

	Rendering: RenderingSettings{
		DirectLight:  0.5,
		AmbientLight: 0.75,

		Roughness:    1.0,
		Reflectivity: 0.05,

		SSAORadius:   0.5,
		SSAORange:    3.0,
		SSAOBoundary: 1.0,

		MinWhite: 8.0,
	},

	Camera: CameraSettings{100.0, 0.0, 0.0, 0.0},
}

func loadSingleSettings(path string) AppSettings {
	settings := copySettings(&defaultSettings)
	serializedSettings, err := ioutil.ReadFile(path)
	if err != nil {
		return settings
	}
	err = json.Unmarshal(serializedSettings, &settings)
	if err != nil {
		panic(err)
	}
	return settings
}

func saveSingleSettings(path string, settings AppSettings) {
	serializedSettings, err := json.Marshal(settings)
	if err != nil {
		panic(err)
	}
	err = ioutil.WriteFile(path, serializedSettings, 0644)
	if err != nil {
		panic(err)
	}
}


var SAVES_DIR = "saves"
var maxSaveNum int

var settingsList []AppSettings
func LoadSettings() (AppSettings, int) {
	os.Mkdir(SAVES_DIR, 0700)
	files, err := ioutil.ReadDir(SAVES_DIR)
	if err != nil {
		panic(err)
	}
	
	settingsList = make([]AppSettings, 0)
	
	settingsMap := make(map[int]AppSettings, 0)
	settingsNameList := make([]int, 0)
	for _, file := range files {
		fileNameParts := strings.Split(file.Name(), "_")
		fileNumString := "0"
		if len(fileNameParts) > 1 {
			fileNumString = fileNameParts[1]
		}
		fileNum, _ := strconv.Atoi(fileNumString)
		
		if maxSaveNum < fileNum {
			maxSaveNum = fileNum
		}
		loadedSettings := loadSingleSettings(SAVES_DIR + "/" + file.Name())
		loadedSettings.id = fileNum
		settingsMap[fileNum] = loadedSettings
		settingsNameList = append(settingsNameList, fileNum)
	}
	sort.Ints(settingsNameList)
	for _, settingName := range settingsNameList {
		settingsList = append(settingsList, settingsMap[settingName])
	}

	activeSettings := loadSingleSettings("settings")
	return activeSettings, len(settingsList)
}

func GetSettings(index int) AppSettings {
	settings := copySettings(&settingsList[index])
	return settings
}

func SaveSettings(settings AppSettings) int {
	newSettings := copySettings(&settings)
	settingsList = append(settingsList, newSettings)

	maxSaveNum += 1
	settingsName := "settings_" + strconv.Itoa(maxSaveNum)
	path := SAVES_DIR + "/" + settingsName
	saveSingleSettings(path, settings)
	return len(settingsList)
}

func SaveActiveSettings(settings AppSettings) {
	saveSingleSettings("settings", settings)
}

func DeleteSettings(index int) int {
	filePath := SAVES_DIR + "/settings_" + strconv.Itoa(settingsList[index].id)
	os.Remove(filePath)
	settingsList = append(settingsList[:index], settingsList[index + 1:]...)
	return len(settingsList)
}


