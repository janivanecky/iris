package app

import (
	"bytes"
	"net/http"
	"encoding/json"
	"io/ioutil"

	"github.com/go-gl/mathgl/mgl32"
)

var responseValue = make(map[string]([5][]int))
var colors [5]mgl32.Vec4
var data = []byte (`{"model": "default"}`)

func GetRandomColorPalette() [] mgl32.Vec4{
	res, err := http.Post("http://colormind.io/api/", "text/json", bytes.NewBuffer(data))
	if err != nil {
		return nil
	}
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil
	}

	err = json.Unmarshal(body, &responseValue)
	if err != nil {
		return nil
	}

	colorInt := responseValue["result"]
	for i, color := range colorInt {
		colors[i] = mgl32.Vec4{
			float32(color[0]) / 255.0,
			float32(color[1]) / 255.0,
			float32(color[2]) / 255.0,
			float32(1.0),
		}
	}

	return colors[:]
}