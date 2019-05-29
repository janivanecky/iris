package ui

import (
    "fmt"
    "math"
    "sort"
    "strings"
    "time"

    "github.com/go-gl/mathgl/mgl32"
    "github.com/go-gl/mathgl/mgl64"

	"../platform"
)

var colorForeground mgl32.Vec4 = mgl32.Vec4{0.0, 0.0, 0.0, .2,}
var colorBackground mgl32.Vec4 = mgl32.Vec4{0.8,0.8,0.8,0.8}
var colorItems mgl32.Vec4 = mgl32.Vec4{0.35,0.35,0.35,1.0}
var colorLabel mgl32.Vec4 = mgl32.Vec4{0.1,0.1,0.1,1.0}
var colorHover mgl32.Vec4 = mgl32.Vec4{0,0,0,1}
var verticalPadding float32 = 15.0
var horizontalPadding float32 = 15.0
var innerPadding float32 = 10.0

var isInputResponsive bool = true

func SetInputResponsive(responsive bool) {
    isInputResponsive = responsive
}

type textRenderingData struct {
    Text string
    Position mgl32.Vec2
    Origin mgl32.Vec2
    Color mgl32.Vec4
    Font *Font
}

type rectRenderingData struct {
    Position mgl32.Vec2
    Size mgl32.Vec2
    Color mgl32.Vec4
    Layer int
}

type rectRenderingDataList []rectRenderingData

type Font interface {
    GetStringHeight() float64
    GetStringWidth(string) float64
    GetStringFit(string, float64) int
    GetStringFitReverse(string, float64) int
}

var textRenderingBuffer []textRenderingData
var rectRenderingBuffer []rectRenderingData

var screenWidth float64
var screenHeight float64

var uiFont Font
var uiTitleFont Font

var cursorPosition int
var fieldStartPosition int
var fieldEndPosition int

var cursorTime time.Time

var keyTimers map[string] *time.Timer
var keyTickers map[string] *time.Ticker
var keyTimer *time.Timer
var keyTicker *time.Ticker
func startKeyTimer(name string) {
    keyTimers[name] = time.NewTimer(time.Millisecond * 500)
}

func getKeyReady(name string) bool {
    if keyTickers[name] != nil {
        select {
        case <- keyTickers[name].C:
            return true
        default:
            return false        
        }
    }
    select {
    case <- keyTimers[name].C:
        keyTickers[name] = time.NewTicker(time.Millisecond * 25)
        return true
    default:
        return false
    }
}

func stopKeyTimer(name string) {
    if keyTimers[name] != nil {
        keyTimers[name].Stop()
        keyTimers[name] = nil
    }

    if keyTickers[name] != nil {
        keyTickers[name].Stop()
        keyTickers[name] = nil
    }
}

func Init(windowWidth float64, windowHeight float64, font Font, titleFont Font) {
    textRenderingBuffer = make([]textRenderingData, 0, 100)
    rectRenderingBuffer = make([]rectRenderingData, 0, 100)

    keyTimers = make(map[string] *time.Timer)
    keyTickers = make(map[string] *time.Ticker)

    uiFont = font
    uiTitleFont = titleFont

    screenHeight = windowHeight
    screenWidth = windowWidth

}

func GetProjectionMatrix() mgl32.Mat4 {
	return mgl32.Ortho(0.0, float32(screenWidth), 0.0, float32(screenHeight), 10.0, -10.0)
}

func (panel *Panel) AddToggle(label string, active bool) (newValue bool, changed bool) {
	changed = false
	newValue = active

    boxMiddleToTotal := 0.6
    height := float64(uiFont.GetStringHeight())
    
    itemPos := panel.position.Add(panel.itemPos)
    stringID := panel.name + "/toggle/" + label
    toggleID := hashString(stringID)

    bgBoxSize := mgl32.Vec2{float32(height), float32(height)}
    bgBoxPos := itemPos

    // Check for mouse input
    if isInputResponsive {
        // Check if mouse over
		mouseX, mouseY := platform.GetMousePosition()
		
		if isInRect(mgl32.Vec2{float32(mouseX), float32(mouseY)}, bgBoxPos, bgBoxSize) {
			setHot(toggleID)
        } else {
            unsetHot(toggleID)
        }

        // If toggle is hot, check for mouse press
        if isHot(toggleID) && platform.IsMouseLeftButtonPressed() {
            newValue = !newValue
            changed = true
            setActive(toggleID)
        } else if isActive(toggleID) && !platform.IsMouseLeftButtonDown() {
            unsetActive(toggleID)
        }
    } else {
        unsetHot(toggleID);
    }
	// Toggle box background
	boxColor := colorItems
    middleColor := colorLabel
    textColor := colorLabel
    if isHot(toggleID) {
        textColor = colorHover
        middleColor = colorHover
        borderSize := float32(2.0)
        rectRenderingBuffer = append(rectRenderingBuffer, rectRenderingData {
            bgBoxPos.Sub(mgl32.Vec2{borderSize, borderSize}), bgBoxSize.Add(mgl32.Vec2{borderSize * 2, borderSize * 2}), colorHover, 2,
        })
    }

    // Draw bg rectangle
    rectRenderingBuffer = append(rectRenderingBuffer, rectRenderingData {
        bgBoxPos, bgBoxSize, boxColor, 3,
    })

    // Active part of toggle box
    if active {
        fgBoxSize := mgl32.Vec2{float32(height * boxMiddleToTotal), float32(height *  boxMiddleToTotal)}
        fgBoxPos := bgBoxPos.Add(
            bgBoxSize.Sub(fgBoxSize).Mul(0.5),
        )
		
        rectRenderingBuffer = append(rectRenderingBuffer, rectRenderingData {
            fgBoxPos, fgBoxSize, middleColor, 4,
        })
    }

	// Draw toggle label
	textPos := mgl32.Vec2{bgBoxPos[0] + innerPadding + bgBoxSize[0], bgBoxPos[1]}
    textRenderingBuffer = append(textRenderingBuffer, textRenderingData {
        label, textPos, mgl32.Vec2{}, textColor, &uiFont,
    })
    textWidth := uiFont.GetStringWidth(label)
	
    // Move current panel item position
    panel.itemPos[1] += float32(height) + innerPadding
    panel.maxWidth = math.Max(panel.maxWidth, float64(textPos[0]) + float64(textWidth))
	return newValue, changed
}

func (panel *Panel) AddTextField(label string, text string) (newValue string, changed bool) {
    changed = false
    newValue = text

    stringID := panel.name + "/text/" + label
    toggleID := hashString(stringID)
    itemPos := panel.position.Add(panel.itemPos)
    height := float32(uiFont.GetStringHeight())

    inputPos := mgl32.Vec2{itemPos[0], itemPos[1]}
    inputSize := mgl32.Vec2{200, height}
    inputColor := colorItems
    textFieldSize := float64(inputSize[0] - 2 * innerPadding)
    
    if isInputResponsive {
        mouseX, mouseY := platform.GetMousePosition()
		
		if isInRect(mgl32.Vec2{float32(mouseX), float32(mouseY)}, inputPos, inputSize) {
			setHot(toggleID)
        } else {
            unsetHot(toggleID)
        }

        // If toggle is hot, check for mouse press
        if isHot(toggleID) && platform.IsMouseLeftButtonPressed() {
            setActive(toggleID)
            cursorPosition = len(text)
            cursorTime = time.Now()
            fieldStartPosition = 0
            fieldEndPosition = uiFont.GetStringFit(newValue, textFieldSize)
        }
        if isActive(toggleID) && (!isHot(toggleID) && platform.IsMouseLeftButtonPressed()){
            unsetActive(toggleID)
            changed = true
        }
    }

    if isActive(toggleID) {
        preCursorString := text[:cursorPosition]
        postCursorString := text[cursorPosition :]
        enteredString := platform.GetCharsPressed()

        newValue = preCursorString + enteredString + postCursorString
        cursorPosition += len(enteredString)

        if platform.IsKeyPressed(platform.KeyEnter) {
            changed = true
        }

        backspace := false
        if platform.IsKeyPressed(platform.KeyBackspace) {
            startKeyTimer("backspace")
            backspace = true
        }
        if platform.IsKeyDown(platform.KeyBackspace) {
            if getKeyReady("backspace") {
                backspace = true
            }
        } else {
            stopKeyTimer("backspace")
        }

        delete := false
        if platform.IsKeyPressed(platform.KeyDelete) {
            startKeyTimer("delete")
            delete = true
        }
        if platform.IsKeyDown(platform.KeyDelete) {
            if getKeyReady("delete") {
                delete = true
            }
        } else {
            stopKeyTimer("delete")
        }


        moveLeft := false
        if platform.IsKeyPressed(platform.KeyLeft) {
            startKeyTimer("left")
            moveLeft = true
        }
        if platform.IsKeyDown(platform.KeyLeft) {
            if getKeyReady("left") {
                moveLeft = true
            }
        } else {
            stopKeyTimer("left")
        }

        moveRight := false
        if platform.IsKeyPressed(platform.KeyRight) {
            startKeyTimer("right")
            moveRight = true
        }
        if platform.IsKeyDown(platform.KeyRight) {
            if getKeyReady("right") {
                moveRight = true
            }
        } else {
            stopKeyTimer("right")
        }

        if moveLeft {
            if platform.IsKeyDown(platform.KeyLeftControl) {
                searchString := strings.TrimRight(preCursorString, " ")
                cutoffIndex := strings.LastIndex(searchString, " ")
                if cutoffIndex < 0 {
                    cutoffIndex = 0
                } 
                cursorPosition = cutoffIndex
            } else {
                cursorPosition--
                if cursorPosition < 0 {
                    cursorPosition = 0
                }
            }
        }

        if moveRight {
            if platform.IsKeyDown(platform.KeyLeftControl) {
                searchString := strings.TrimLeft(postCursorString, " ")
                leftSpaces := len(postCursorString) - len(searchString)
                cutoffIndex := strings.Index(searchString, " ")
                if cutoffIndex < 0 {
                    cutoffIndex = len(postCursorString)
                }else {
                    cutoffIndex += leftSpaces
                }
                cursorPosition += cutoffIndex
            } else {
                cursorPosition++
                if cursorPosition > len(newValue) {
                    cursorPosition = len(newValue)
                }
            }
        }

        if backspace {
            stringLength := len(newValue)

            if stringLength > 0 {
                preCursorString := newValue[:cursorPosition]
                postCursorString := newValue[cursorPosition :]

                if platform.IsKeyDown(platform.KeyLeftControl) {
                    searchString := strings.TrimRight(preCursorString, " ")
                    cutoffIndex := strings.LastIndex(searchString, " ")
                    if cutoffIndex < 0 {
                        cutoffIndex = 0
                    }
                    preCursorString = preCursorString[:cutoffIndex]
                    preCursorString = strings.TrimRight(preCursorString, " ")
                    newValue = preCursorString + postCursorString
                    cursorPosition = len(preCursorString)
                } else {
                    newValue = preCursorString[:len(preCursorString) - 1] + postCursorString
                    cursorPosition--
                }
            }
        }

        if delete {
            stringLength := len(newValue)

            if stringLength > 0 {
                preCursorString := newValue[:cursorPosition]
                postCursorString := newValue[cursorPosition:]
                if len(postCursorString) > 0 {
                    if platform.IsKeyDown(platform.KeyLeftControl) {
                        searchString := strings.TrimLeft(postCursorString, " ")
                        leftSpaces := len(postCursorString) - len(searchString)
                        cutoffIndex := strings.Index(searchString, " ")
                        if cutoffIndex < 0 {
                            cutoffIndex = len(postCursorString)
                        } else
                        {
                            cutoffIndex += leftSpaces
                        }
                        postCursorString = postCursorString[cutoffIndex:]
                        postCursorString = strings.TrimLeft(postCursorString, " ")
                        newValue = preCursorString + postCursorString
                    } else {
                        newValue = preCursorString + postCursorString[1:]
                    }
                }
            }
        }
        
        home := false
        if platform.IsKeyPressed(platform.KeyHome) {
            cursorPosition = 0
            home = true
        }

        end := false
        if platform.IsKeyPressed(platform.KeyEnd) {
            cursorPosition = len(newValue)
            end = true
        }

        if moveLeft || moveRight || backspace || delete || home || end {
            cursorTime = time.Now()
        }
    }

    textColor := colorLabel
    if isHot(toggleID) || isActive(toggleID) {
        textColor = colorHover
        borderSize := float32(2.0)
        rectRenderingBuffer = append(rectRenderingBuffer, rectRenderingData {
            inputPos.Sub(mgl32.Vec2{borderSize, borderSize}), inputSize.Add(mgl32.Vec2{borderSize * 2, borderSize * 2}), colorHover, 2,
        })
    }
    
    rectRenderingBuffer = append(rectRenderingBuffer, rectRenderingData {
        inputPos, inputSize, inputColor, 3,
    })

    textPos := itemPos.Add(mgl32.Vec2{innerPadding, 0.0})
    textOrigin := mgl32.Vec2{0,0}
    textStart := 0
    textEnd := uiFont.GetStringFit(newValue, textFieldSize)
    if isActive(toggleID) {
        fieldEndPosition = int(math.Min(float64(fieldEndPosition), float64(len(newValue))))
        
        fieldStartPosition = int(math.Min(float64(fieldStartPosition), float64(len(newValue) - 1)))
        fieldStartPosition = int(math.Max(float64(fieldStartPosition), 0))
        
        fieldEndPosition = int(math.Max(float64(fieldEndPosition), float64(uiFont.GetStringFit(newValue[fieldStartPosition:], textFieldSize))))

        if cursorPosition > fieldEndPosition {
            fieldEndPosition = cursorPosition
            fieldStartPosition = uiFont.GetStringFitReverse(newValue[:fieldEndPosition], textFieldSize)
        } else if cursorPosition < fieldStartPosition {
            fieldStartPosition = cursorPosition
            fieldEndPosition = uiFont.GetStringFit(newValue[fieldStartPosition:], textFieldSize) + fieldStartPosition
        }

        textStart = fieldStartPosition
        textEnd = fieldEndPosition
    }
    displayedText := newValue[textStart:textEnd]
    textRenderingBuffer = append(textRenderingBuffer, textRenderingData {
        displayedText, textPos, textOrigin, textColor, &uiFont,
    })

    if isActive(toggleID) {
        cursorPos := textPos.Add(mgl32.Vec2{float32(uiFont.GetStringWidth(displayedText[:cursorPosition - textStart])), 0.0})
        cursorSize := mgl32.Vec2{2.5, height}
        cursorColor := textColor
        
        cTime := time.Now().Sub(cursorTime).Seconds()
        if math.Sin(cTime * 6.0) < 0 {
            cursorColor[3] = 0.0
        }
        rectRenderingBuffer = append(rectRenderingBuffer, rectRenderingData {
            cursorPos, cursorSize, cursorColor, 4,
        })
    }
    
    
    // Slider label
    labelPos := mgl32.Vec2{inputPos[0] + inputSize[0] + innerPadding, itemPos[1]}
    textRenderingBuffer = append(textRenderingBuffer, textRenderingData {
        label, labelPos, mgl32.Vec2{}, textColor, &uiFont,
    })
    labelWidth := uiFont.GetStringWidth(label)

    panel.maxWidth = math.Max(panel.maxWidth, float64(labelPos[0]) + float64(labelWidth))
    panel.itemPos[1] += inputSize[1] + innerPadding

    return newValue, changed
}

func (panel *Panel) AddSlider(label string, value float64, min float64, max float64) (newValue float64, changed bool) {
    changed = false
    newValue = mgl64.Clamp(value, min, max)
    stringID := panel.name + "/slider/" + label
    sliderID := hashString(stringID)
    itemPos := panel.position.Add(panel.itemPos)
    
    height := float32(uiFont.GetStringHeight())
    sliderWidth := float32(200.0)

    // Slider bar
    sliderStart := float32(0.0)

    sliderBarColor := colorItems
    sliderBarPos := mgl32.Vec2{itemPos[0] + sliderStart, itemPos[1]}
    sliderBarSize := mgl32.Vec2{sliderWidth, height}

    if isHot(sliderID) {
        borderSize := float32(2.0)
        rectRenderingBuffer = append(rectRenderingBuffer, rectRenderingData {
            sliderBarPos.Sub(mgl32.Vec2{borderSize, borderSize}), sliderBarSize.Add(mgl32.Vec2{borderSize * 2, borderSize * 2}), mgl32.Vec4{0,0,0,1}, 2,
        })
    }

    rectRenderingBuffer = append(rectRenderingBuffer, rectRenderingData {
        sliderBarPos, sliderBarSize, sliderBarColor, 3,
    })
    
    sliderColor := colorLabel.Mul(0.8)
    sliderSize := mgl32.Vec2{height, height}
    sliderX := float32((newValue - min) / (max - min)) * (sliderWidth - sliderSize[0]) + sliderBarPos[0] + sliderSize[0] * 0.5
    sliderPos := mgl32.Vec2{sliderX - sliderSize[0] * 0.5, itemPos[1]}

    // Check for mouse input
    if isInputResponsive {
        mouseX, mouseY := platform.GetMousePosition()
        mousePosition := mgl32.Vec2{float32(mouseX), float32(mouseY)}
        if isInRect(mousePosition, sliderBarPos, sliderBarSize) {       
            setHot(sliderID)
        } else if !isActive(sliderID) {
            unsetHot(sliderID)
        }

        overallSliderSize := mgl32.Vec2{sliderBarSize[0], sliderSize[1]}
        overallSliderPos := sliderBarPos
        if (isHot(sliderID) || isInRect(mousePosition, overallSliderPos, overallSliderSize)) && !isActive(sliderID) && platform.IsMouseLeftButtonPressed() {
            setActive(sliderID)
        } else if isActive(sliderID) && !platform.IsMouseLeftButtonDown() {
            unsetActive(sliderID)
        }
    } else {
        unsetHot(sliderID)
        unsetActive(sliderID)
    }

    textColor := colorLabel
    if isHot(sliderID) {
        textColor = colorHover
        sliderColor = colorHover.Mul(0.8)
    }

    if isActive(sliderID) {
        mouseX, _ := platform.GetMousePosition()
        mouseXRel := (float32(mouseX) - sliderBarPos[0] - sliderSize[0] * 0.5) / (sliderBarSize[0] - sliderSize[0]);
        
        // TODO: Clamp!!
        if mouseXRel < 0.0 {
            mouseXRel = 0.0
        }
        if mouseXRel > 1.0 {
            mouseXRel = 1.0
        }
        
        newValue = float64(mouseXRel) * (max - min) + min
        changed = true
    }

    // Slider
    rectRenderingBuffer = append(rectRenderingBuffer, rectRenderingData {
        sliderPos, sliderSize, sliderColor, 4,
    })

    // Number
    currentPos := mgl32.Vec2{sliderBarPos[0] + sliderBarSize[0] / 2.0, itemPos[1]}
    numberString := fmt.Sprintf("%.02f", newValue)
    textRenderingBuffer = append(textRenderingBuffer, textRenderingData {
        numberString, currentPos, mgl32.Vec2{0.5, 0}, sliderColor, &uiFont,
    })

    // Slider label
    textPos := mgl32.Vec2{sliderBarSize[0] + sliderBarPos[0] + innerPadding, itemPos[1]}
    textRenderingBuffer = append(textRenderingBuffer, textRenderingData {
        label, textPos, mgl32.Vec2{}, textColor, &uiFont,
    })
    textWidth := uiFont.GetStringWidth(label)

    panel.itemPos[1] += height + innerPadding
    panel.maxWidth = math.Max(panel.maxWidth, float64(textPos[0]) + float64(textWidth))
    return newValue, changed
}

func posFromHS(h, s float32, pickerSize float32) mgl32.Vec2 {
    x := h * pickerSize
    y := s * s * pickerSize
    return mgl32.Vec2{x,y}
}

func posFromV(v float32, pickerSize float32) float32 {
    x := v * pickerSize
    return x
}

func (panel *Panel) AddColorPalette(label string, color mgl32.Vec4, selected bool) (newValue bool, changed bool) {
    changed = false

    colorPosition := panel.position.Add(panel.itemPos)
    colorHeight := float32(uiFont.GetStringHeight())
    colorWidth := float32(200)
    colorSize := mgl32.Vec2{colorWidth, colorHeight}
    rectRenderingBuffer = append(rectRenderingBuffer, rectRenderingData {
        colorPosition, colorSize, color, 3,
    })

    stringID := panel.name + "/color/" + label
    colorID := hashString(stringID)
    if isInputResponsive {
        mouseX, mouseY := platform.GetMousePosition()
        mousePosition := mgl32.Vec2{float32(mouseX), float32(mouseY)}
        if isInRect(mousePosition, colorPosition, colorSize) {
            setHot(colorID)
        } else {
            unsetHot(colorID)
        }
    }

    if isHot(colorID) {
        if platform.IsMouseLeftButtonPressed() {
            selected = !selected
        }
    }

    textColor := colorLabel
    if isHot(colorID) || selected {
        textColor = colorHover
        borderSize := float32(2.0)
        rectRenderingBuffer = append(rectRenderingBuffer, rectRenderingData {
            colorPosition.Sub(mgl32.Vec2{borderSize, borderSize}), colorSize.Add(mgl32.Vec2{borderSize * 2, borderSize * 2}), colorHover, 2,
        })
    }

    panel.itemPos[1] += colorHeight + innerPadding
    
    // Slider label
    textPos := mgl32.Vec2{colorWidth + colorPosition[0] + innerPadding, colorPosition[1]}
    textRenderingBuffer = append(textRenderingBuffer, textRenderingData {
        label, textPos, mgl32.Vec2{}, textColor, &uiFont,
    })
    textWidth := uiFont.GetStringWidth(label)
    panel.maxWidth = math.Max(panel.maxWidth, float64(textPos[0]) + float64(textWidth))

    newValue = selected
    return newValue, changed
}

func (panel *Panel) AddColorPicker(label string, color mgl32.Vec4, showLabel bool) (newValue mgl32.Vec4, changed bool) {
    stringID := panel.name + "/colorPick/" + label
    pickerID := hashString(stringID)
    
    changed = false
    newValue = color

    pickerSize := float32(200.0)
    resolution := 100
    step := pickerSize / float32(resolution)
    pickerPos := panel.position.Add(panel.itemPos)

    hsvColor := RGBtoHSV(color)

    mouseX, mouseY := platform.GetMousePosition()
    mousePosition := mgl32.Vec2{float32(mouseX), float32(mouseY)}
    if isActive(pickerID) {
        if mousePosition[0] < pickerPos[0] {
            mousePosition[0] = pickerPos[0]
        }
        if mousePosition[1] < pickerPos[1] {
            mousePosition[1] = pickerPos[1]
        }
        if mousePosition[0] > pickerPos[0] + pickerSize {
            mousePosition[0] = pickerPos[0] + pickerSize
        }
        if mousePosition[1] > pickerPos[1] + pickerSize {
            mousePosition[1] = pickerPos[1] + pickerSize
        }
    }

    for hi := 0; hi < resolution; hi++ {
        for si := 0; si < resolution; si++ {
            h := float32(hi) / float32(resolution)
            s := float32(si + 1) / float32(resolution)
            s = float32(math.Sqrt(float64(s)))
            v := hsvColor[2]

            color := mgl32.Vec4{h, s, v, 1.0}
            colorRGB := HSVtoRGB(color)
            
            pos := pickerPos
            pos[0] += float32(hi) * step
            pos[1] += float32(si) * step
            size := mgl32.Vec2{float32(step), float32(step)}
            rectRenderingBuffer = append(rectRenderingBuffer, rectRenderingData {
                pos, size, colorRGB, 2,
            })

            if isInputResponsive {
                if isInRect(mousePosition, pos, size) {
                    if isActive(pickerID) {
                        newValue = colorRGB;
                        changed = true
                    } else if platform.IsMouseLeftButtonPressed(){ 
                        setActive(pickerID)
                        newValue = colorRGB;
                        changed = true
                    } 
                }
            }
        }
    }

    if isActive(pickerID) && !platform.IsMouseLeftButtonDown() {
        unsetActive(pickerID)
    }

    newValueHSV := RGBtoHSV(newValue)
    markerPosCenter := pickerPos.Add(posFromHS(newValueHSV[0], newValueHSV[1], pickerSize))
    markerSize := mgl32.Vec2{10,10}
    markerPos := markerPosCenter.Sub(markerSize.Mul(0.5))
    rectRenderingBuffer = append(rectRenderingBuffer, rectRenderingData {
        markerPos, markerSize, colorHover, 3,
    })

    panel.itemPos[1] += pickerSize + innerPadding

    valuePickerHeight := float32(uiFont.GetStringHeight())
    valuePickerPos := pickerPos.Add(mgl32.Vec2{0, pickerSize + innerPadding})
    
    valueStringID := stringID + "/val"
    valuePickerID := hashString(valueStringID)

    if isActive(valuePickerID) {
        if mousePosition[0] < valuePickerPos[0] {
            mousePosition[0] = valuePickerPos[0]
        }
        if mousePosition[1] < valuePickerPos[1] {
            mousePosition[1] = valuePickerPos[1]
        }
        if mousePosition[0] > valuePickerPos[0] + pickerSize {
            mousePosition[0] = valuePickerPos[0] + pickerSize
        }
        if mousePosition[1] > valuePickerPos[1] + valuePickerHeight {
            mousePosition[1] = valuePickerPos[1] + valuePickerHeight
        }
    }

    for vi := 0; vi < resolution; vi++ {
        h := newValueHSV[0]
        s := newValueHSV[1]
        v := float32(vi + 1) / float32(resolution)

        color := mgl32.Vec4{h, s, v, 1.0}
        colorRGB := HSVtoRGB(color)
        pos := valuePickerPos
        pos[0] += float32(vi) * step
        size := mgl32.Vec2{float32(step), valuePickerHeight}
        rectRenderingBuffer = append(rectRenderingBuffer, rectRenderingData {
            pos, size, colorRGB, 2,
        })
        if isInputResponsive {
            if isInRect(mousePosition, pos, size) {
                if isActive(valuePickerID) {
                    newValue = colorRGB;
                    changed = true
                } else if platform.IsMouseLeftButtonPressed(){ 
                    setActive(valuePickerID)
                    newValue = colorRGB;
                    changed = true
                } 
            }
        }
    }
    
    if isActive(valuePickerID) && !platform.IsMouseLeftButtonDown() {
        unsetActive(valuePickerID)
    }

    newValueHSV = RGBtoHSV(newValue)
    markerPos = valuePickerPos.Add(mgl32.Vec2{posFromV(newValueHSV[2], pickerSize), 0})
    markerSize = mgl32.Vec2{10, valuePickerHeight}
    markerPos[0] -= markerSize[0] * 0.5
    rectRenderingBuffer = append(rectRenderingBuffer, rectRenderingData {
        markerPos, markerSize, mgl32.Vec4{0,0,0,1}, 3,
    })

    panel.itemPos[1] += valuePickerHeight + innerPadding
    
    textPos := mgl32.Vec2{pickerSize + pickerPos[0] + innerPadding, pickerPos[1]}
    // Slider label
    if showLabel {
        textRenderingBuffer = append(textRenderingBuffer, textRenderingData {
            label, textPos, mgl32.Vec2{}, colorLabel, &uiFont,
        })
    }
    textWidth := uiFont.GetStringWidth(label)
    panel.maxWidth = math.Max(panel.maxWidth, float64(textPos[0]) + float64(textWidth))
    

    return newValue, changed
}

type Panel struct {
    position mgl32.Vec2
    itemPos mgl32.Vec2
    maxWidth float64
    width float64
    name string
}

func StartPanel(name string, position mgl32.Vec2, width float64) Panel {
    var panel Panel
    panel.position = position
    panel.name = name
    panel.maxWidth = float64(horizontalPadding)
    panel.width = width
    panel.itemPos[0] = horizontalPadding
    panel.itemPos[1] = float32(uiTitleFont.GetStringHeight()) + verticalPadding * 2
    return panel
}

func (panel *Panel) GetBottom() float32{
    return panel.position[1] + panel.itemPos[1] + verticalPadding - innerPadding;
}

func (panel *Panel) GetWidth() float32 {
    width := float32(panel.width)
    if width <= 0 {
        width = float32(panel.maxWidth) + horizontalPadding
    }
    return width
}

func (panel *Panel) End() {
    titlePos := mgl32.Vec2{panel.position[0] + horizontalPadding, panel.position[1] + verticalPadding}
    textRenderingBuffer = append(textRenderingBuffer, textRenderingData {
        panel.name, titlePos, mgl32.Vec2{}, colorLabel, &uiTitleFont, //colorBackground,
    })
    titleWidth := uiTitleFont.GetStringWidth(panel.name)
    panel.maxWidth = math.Max(panel.maxWidth, float64(titlePos[0]) + float64(titleWidth))

    panelHeight := panel.itemPos[1] + verticalPadding - innerPadding
    panelWidth := panel.GetWidth()
    rectRenderingBuffer = append(rectRenderingBuffer, rectRenderingData {
        panel.position, mgl32.Vec2{panelWidth, panelHeight}, colorBackground, 0,
    })

    //titleBarHeight := float32(uiFont.GetStringHeight())
    //rectRenderingBuffer = append(rectRenderingBuffer, rectRenderingData {
    //    panel.position, mgl32.Vec2{panelWidth, titleBarHeight}, colorBackground, 2,//colorForeground, 2,
    //})
}

func (panel *Panel) GetBoundingRect() mgl32.Vec4 {
    return mgl32.Vec4{panel.position[0], panel.position[1], panel.GetWidth(), panel.GetBottom() - panel.position[1]}
}

func GetDrawData() ([]rectRenderingData, []textRenderingData){
    sort.Slice(rectRenderingBuffer, func(i, j int) bool {
        return rectRenderingBuffer[i].Layer < rectRenderingBuffer[j].Layer
      })

    return rectRenderingBuffer, textRenderingBuffer
}

func Clear() {
    rectRenderingBuffer = rectRenderingBuffer[:0]
    textRenderingBuffer = textRenderingBuffer[:0]
}

func isInRect(position mgl32.Vec2, rectPosition mgl32.Vec2, rectSize mgl32.Vec2) bool {
    if position[0] >= rectPosition[0] && position[0] <= rectPosition[0] + rectSize[0] &&
       position[1] >= rectPosition[1] && position[1] <= rectPosition[1] + rectSize[1] {
		   return true
	   }
    return false
}

var hotID int = -1
var activeID int = -1

func setHot(itemID int) {
	if hotID == -1 {
		hotID = itemID;
	}
}

func unsetHot(itemID int) {
    if hotID == itemID {
        hotID = -1
    }
}

func isHot(itemID int) bool {
	return itemID == hotID
}

var IsRegisteringInput bool = false

func setActive(itemID int) {
    activeID = itemID
    IsRegisteringInput = true;
}

func unsetActive(itemID int) {
    if activeID == itemID {
        activeID = -1
        IsRegisteringInput = false
    }
}

func isActive(itemID int) bool {
    return itemID == activeID
}

func hashString(text string) int {
    hashValue := 5381
	for _, c := range text {
        hashValue = ((hashValue << 5) + hashValue) + int(c);
    }

    return hashValue + 1;
}

func HSVtoRGB(color mgl32.Vec4) mgl32.Vec4 {
    h, s, v, a := color[0], color[1], color[2], color[3]
    var min float32
    var r, g, b float32
	chroma := s * v
	hDash := h * 6.0
	x := chroma * float32((1.0 - math.Abs(math.Mod(float64(hDash), 2.0) - 1.0)))
	if hDash < 1.0 {
		r = chroma
		g = x
	} else if hDash < 2.0 {
		r = x
		g = chroma
	} else if hDash < 3.0 {
		g = chroma
		b = x
	} else if hDash < 4.0 {
		g = x
		b = chroma
	} else if hDash < 5.0 {
		r = x
		b = chroma
	} else if hDash <= 6.0 {
		r = chroma
		b = x
	}

	min = v - chroma

	r += min
	g += min
	b += min

	return mgl32.Vec4{r, g, b, a}
}

func RGBtoHSV(color mgl32.Vec4) mgl32.Vec4 {
	cMax := float32(math.Max(float64(color[0]), math.Max(float64(color[1]), float64(color[2]))))
	cMin := float32(math.Min(float64(color[0]), math.Min(float64(color[1]), float64(color[2]))))
	delta := cMax - cMin
	h := float32(0)
	coef := float32(1.0 / 6.0)
    if delta == 0.0 {
        h = 0
    } else if cMax == color[0] {
		h = coef * float32(math.Mod(float64((color[1] - color[2]) / delta), 6.0))
	} else if cMax == color[1] {
		h = coef * ((color[2] - color[0]) / delta + 2.0)
	} else if cMax == color[2] {
		h = coef * ((color[0] - color[1]) / delta + 4.0)
	}

	s := float32(0)
	if cMax == 0 {
		s = 0
	} else {
		s = delta / cMax;
    } 
    if h < 0 {
		h += 1.0
	}

	v := cMax
	return mgl32.Vec4{h, s, v, color[3]}
}
