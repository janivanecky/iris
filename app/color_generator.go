package app

import (
	"bytes"
	"net/http"
	"encoding/json"
	"io/ioutil"

	"github.com/go-gl/mathgl/mgl32"
)


// GetRandomColorPalette returns a random color palette of 5 colors.
func GetRandomColorPalette() [] mgl32.Vec4{
	// Send a request to colormind.io.
	var requestData = []byte (`{"model": "default"}`)
	res, err := http.Post("http://colormind.io/api/", "text/json", bytes.NewBuffer(requestData))
	if err != nil {
		return nil
	}
	defer res.Body.Close()
	
	// Read the response and get its body as a `map[string]([5][]int)`.
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil
	}
	var responseValue = make(map[string]([5][]int))
	err = json.Unmarshal(body, &responseValue)
	if err != nil {
		return nil
	}
	responseColors := responseValue["result"]
	
	// Convert colors in response to Vec4s.
	var colors [5]mgl32.Vec4
	for i, color := range responseColors {
		colors[i] = mgl32.Vec4{
			float32(color[0]) / 255.0,
			float32(color[1]) / 255.0,
			float32(color[2]) / 255.0,
			float32(1.0),
		}
	}

	return colors[:]
}