package ui

import (
	gmath "../math"
	"../input"
    "fmt"
    "sort"
    "math"
    "../font"
)

var colorForeground gmath.Vec4 = gmath.Vec4{28.0 / 255.0,224.0 / 255.0,180.0 / 255.0,1}
var colorBackground gmath.Vec4 = gmath.Vec4{0.1,0.1,0.1,1}
var colorLabel gmath.Vec4 = gmath.Vec4{1,1,1,1}

var isInputResponsive bool = true

func SetInputResponsive(responsive bool) {
    isInputResponsive = responsive
}

type textRenderingData struct {
    Text string
    Position gmath.Vec2
    Origin gmath.Vec2
    Color gmath.Vec4
}

type rectRenderingData struct {
    Position gmath.Vec2
    Size gmath.Vec2
    Color gmath.Vec4
    Layer int
}

type rectRenderingDataList []rectRenderingData
func (rectList rectRenderingDataList)Less(i, j int) bool {
    return rectList[i].Layer < rectList[j].Layer
}

var textRenderingBuffer []textRenderingData
var rectRenderingBuffer []rectRenderingData

var screenHeight float64 = 0

var uiFont font.Font

func Init(windowWidth int, windowHeight int, font font.Font) {
    textRenderingBuffer = make([]textRenderingData, 100)
    rectRenderingBuffer = make([]rectRenderingData, 100)
    uiFont = font

    colorForeground = gmath.Vec4{
        float32(math.Pow(28.0 / 255.0, 2.2)), 
        float32(math.Pow(224.0 / 255.0, 2.2)),
        float32(math.Pow(180.0 / 255.0, 2.2)),
        1,
    }
	screenHeight = float64(windowHeight)
}


func (panel *Panel) AddToggle(label string, active bool) (newValue bool, changed bool) {
	changed = false
	newValue = active

    boxMiddleToTotal := 0.6
    height := float64(uiFont.RowHeight)
    
    itemPos := panel.position.Add(panel.itemPos)
	toggleID := hashString(label)

    bgBoxSize := gmath.Vec2{float32(height), float32(height)}
    bgBoxPos := itemPos

    // Check for mouse input
    if isInputResponsive {
        // Check if mouse over
		mouseX, mouseY := input.GetMousePosition()
		
		if isInRect(gmath.Vec2{float32(mouseX), float32(mouseY)}, bgBoxPos, bgBoxSize) {
			setHot(toggleID)
        } else {
            unsetHot(toggleID)
        }

        // If toggle is hot, check for mouse press
        if isHot(toggleID) && input.IsMouseLeftButtonPressed() {
            newValue = !newValue
            changed = true
            setActive(toggleID)
        } else if isActive(toggleID) && !input.IsMouseLeftButtonDown() {
            unsetActive(toggleID)
        }
    } else {
        unsetHot(toggleID);
    }
	// Toggle box background
	boxColor := colorForeground
	middleColor := colorBackground
    if isHot(toggleID) {
		for i := 0; i < 4; i++ {
			boxColor[i] *= 0.8
			middleColor[i] *= 0.8
		}
		middleColor[3] = 1.0
    }

    // Draw bg rectangle
    rectRenderingBuffer = append(rectRenderingBuffer, rectRenderingData {
        bgBoxPos, bgBoxSize, boxColor, 1,
    })

    // Active part of toggle box
    if !active {
        fgBoxSize := gmath.Vec2{float32(height * boxMiddleToTotal), float32(height *  boxMiddleToTotal)}
        fgBoxPos := bgBoxPos.Add(
            bgBoxSize.Sub(fgBoxSize).Mul(0.5),
        )
		
        rectRenderingBuffer = append(rectRenderingBuffer, rectRenderingData {
            fgBoxPos, fgBoxSize, middleColor, 1,
        })
    }

	// Draw toggle label
	innerPadding := float32(10.0)
	textPos := gmath.Vec2{bgBoxPos[0] + innerPadding + bgBoxSize[0], bgBoxPos[1]}
    textRenderingBuffer = append(textRenderingBuffer, textRenderingData {
        label, textPos, gmath.Vec2{}, colorLabel,
    })
    textWidth := uiFont.GetStringWidth(label)
	
    // Move current panel item position
    panel.itemPos[1] += float32(height) + innerPadding
    panel.maxWidth = math.Max(panel.maxWidth, float64(textPos[0]) + textWidth)
	return newValue, changed
}

func (panel *Panel) AddSlider(label string, value float64, min float64, max float64) (newValue float64, changed bool) {
    changed = false
    newValue = value
    sliderID := hashString(label)
    itemPos := panel.position.Add(panel.itemPos)
    
    height := float32(uiFont.RowHeight)
    sliderWidth := float32(200.0)

    // Slider bar
    sliderStart := float32(0.0)

    sliderBarColor := gmath.Vec4{colorBackground[0] * 2.0, colorBackground[1] * 2.0, colorBackground[2] * 2.0, 1.0}
    //sliderBarColor := colorBackground.Mul(2.0)
    //sliderBarColor[3] = 1.0
    sliderBarPos := gmath.Vec2{itemPos[0] + sliderStart, itemPos[1]}
    sliderBarSize := gmath.Vec2{sliderWidth, height}
    rectRenderingBuffer = append(rectRenderingBuffer, rectRenderingData {
        sliderBarPos, sliderBarSize, sliderBarColor, 1,
    })
    
    sliderColor := colorForeground
    sliderSize := gmath.Vec2{height, height}
    sliderX := float32((value - min) / (max - min)) * (sliderWidth - sliderSize[0]) + sliderBarPos[0] + sliderSize[0] * 0.5
    sliderPos := gmath.Vec2{sliderX - sliderSize[0] * 0.5, itemPos[1]}

    // Number
    currentPos := gmath.Vec2{sliderBarPos[0] + sliderBarSize[0] / 2.0, itemPos[1]}
    numberString := fmt.Sprintf("%.02f", value)
    textRenderingBuffer = append(textRenderingBuffer, textRenderingData {
        numberString, currentPos, gmath.Vec2{0.5, 0}, colorLabel,
    })

    // Check for mouse input
    if isInputResponsive {
        mouseX, mouseY := input.GetMousePosition()
        mousePosition := gmath.Vec2{float32(mouseX), float32(mouseY)}
        if isInRect(mousePosition, sliderPos, sliderSize) {       
            setHot(sliderID)
        } else if !isActive(sliderID) {
            unsetHot(sliderID)
        }

        overallSliderSize := gmath.Vec2{sliderBarSize[0], sliderSize[1]}
        overallSliderPos := sliderBarPos
        if (isHot(sliderID) || isInRect(mousePosition, overallSliderPos, overallSliderSize)) && !isActive(sliderID) && input.IsMouseLeftButtonPressed() {
            setActive(sliderID)
        } else if isActive(sliderID) && !input.IsMouseLeftButtonDown() {
            unsetActive(sliderID)
        }
    } else {
        unsetHot(sliderID)
        unsetActive(sliderID)
    }

    if isHot(sliderID) {
        sliderColor = gmath.Vec4{
            sliderColor[0] * 0.8,
            sliderColor[1] * 0.8,
            sliderColor[2] * 0.8,
            1.0,
        }
    }

    if isActive(sliderID) {
        mouseX, _ := input.GetMousePosition()
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
        sliderPos, sliderSize, sliderColor, 1,
    })

    // Slider label
    textPos := gmath.Vec2{sliderBarSize[0] + sliderBarPos[0] + innerPadding, itemPos[1]}
    textRenderingBuffer = append(textRenderingBuffer, textRenderingData {
        label, textPos, gmath.Vec2{}, colorLabel,
    })
    textWidth := uiFont.GetStringWidth(label)

    panel.itemPos[1] += height + innerPadding
    
    panel.maxWidth = math.Max(panel.maxWidth, float64(textPos[0]) + textWidth)
    return newValue, changed
}


type Panel struct {
    position gmath.Vec2
    itemPos gmath.Vec2
    maxWidth float64
    name string
}

func StartPanel(name string, position gmath.Vec2) Panel {
    var panel Panel
    panel.position = position
    panel.name = name
    panel.maxWidth = float64(horizontalPadding)
    panel.itemPos[0] = horizontalPadding
    panel.itemPos[1] = float32(uiFont.RowHeight) + verticalPadding * 2.0
    return panel
}

func (panel *Panel) End() {
    titlePos := gmath.Vec2{panel.position[0] + horizontalPadding, innerPadding}
    textRenderingBuffer = append(textRenderingBuffer, textRenderingData {
        panel.name, titlePos, gmath.Vec2{}, colorBackground,
    })
    titleWidth := uiFont.GetStringWidth(panel.name)
    panel.maxWidth = math.Max(panel.maxWidth, float64(titlePos[1]) + titleWidth)

    panelHeight := panel.itemPos[1] + verticalPadding - innerPadding
    panelWidth := float32(panel.maxWidth) + horizontalPadding
    rectRenderingBuffer = append(rectRenderingBuffer, rectRenderingData {
        panel.position, gmath.Vec2{panelWidth, panelHeight}, colorBackground, 0,
    })

    titleBarHeight := float32(uiFont.RowHeight) + innerPadding * 2.0
    rectRenderingBuffer = append(rectRenderingBuffer, rectRenderingData {
        panel.position, gmath.Vec2{panelWidth, titleBarHeight}, colorForeground, 0,
    })
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

func isInRect(position gmath.Vec2, rectPosition gmath.Vec2, rectSize gmath.Vec2) bool {
    if position[0] >= rectPosition[0] && position[0] <= rectPosition[0] + rectSize[0] &&
       position[1] >= rectPosition[1] && position[1] <= rectPosition[1] + rectSize[1] {
		   return true
	   }
    return false
}

var verticalPadding float32 = 15.0
var horizontalPadding float32 = 15.0
var innerPadding float32 = 10.0
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
    if activeID == -1 {
        activeID = itemID
        IsRegisteringInput = true;
    }
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

