package app

import (
	"os"
	"image"
	"github.com/go-gl/mathgl/mgl32"

	"../lib/font"
	"../lib/graphics"
	"../lib/platform"
)

type SettingsBar struct {
	PosX      		   FloatParameter
	ContentY 		   FloatParameter
	Color    		   ColorParameter
	SaveButtonColor	   ColorParameter
	DeleteButtonColors []ColorParameter
	SettingsColors     []ColorParameter
	SettingsTextures   []graphics.Texture

	height		       float32 // TODO: Should be f32 or f64?
	font			   font.Font
	deleteIcon         graphics.Texture
}

type Action int
const (
	NONE Action = iota
	SAVE   = iota
	DELETE = iota
	SELECT = iota
)

const settingsBarWidth 		 = 500.0
const settingsBarWidthHidden = 50.0
const settingsBarHiddenX	 = settingsBarWidthHidden - settingsBarWidth
const settingPadding		 = 10.0

const scrollSpeed			 = 100.0

var settingsBarColor 		   		  = mgl32.Vec4{0.0, 0.0, 0.0, 0.4}
var settingsBarColorHover 	   		  = mgl32.Vec4{0.0, 0.0, 0.0, 0.8}
var settingsBarColorInactive	      = mgl32.Vec4{0.0, 0.0, 0.0, 0.01}
var settingsColor 		    		  = mgl32.Vec4{1, 1, 1, 0.8}
var settingsColorHover   			  = mgl32.Vec4{1, 1, 1, 1.0}
var settingsDeleteButtonColor 		  = mgl32.Vec4{1, 0, 0, 0.5}
var settingsDeleteButtonColorHover    = mgl32.Vec4{1, 0, 0, 0.9}
var settingsDeleteButtonColorInactive = mgl32.Vec4{0, 0, 0, 0}
const settingsDeleteButtonWidth		  = 70.0

var deleteIconSize = mgl32.Vec2{50, 50}

var   saveButtonColor 	     = mgl32.Vec4{0, 1, 0.5, 0.6}
var   saveButtonColorHover   = mgl32.Vec4{0, 1, 0.5, 1.0}
var   saveTextColor  		 = mgl32.Vec4{0, 0, 0, 0.6}
const saveButtonHeight 		 = 50.0

func GetSettingsBar(font font.Font, deleteIcon graphics.Texture, height float32) SettingsBar {
	var settingsBar SettingsBar
	settingsBar.PosX    		   = FloatParameter{settingsBarHiddenX, settingsBarHiddenX}
	settingsBar.ContentY 		   = FloatParameter{0, 0}
	settingsBar.Color   		   = ColorParameter{settingsBarColor, settingsBarColor}
	settingsBar.SaveButtonColor    = ColorParameter{saveButtonColor, saveButtonColor}
	settingsBar.DeleteButtonColors = make([]ColorParameter, 0)
	settingsBar.SettingsColors     = make([]ColorParameter, 0)
	
	settingsBar.height 	   = height
	settingsBar.font 	   = font
	
	trashIconFile, err := os.Open("trash.png")
	if err != nil {
		panic(err)
	}
	trashIconImg, _, err := image.Decode(trashIconFile)
	if err != nil {
		panic(err)
	}
	trashIconFile.Close()
	trashIconImgData := trashIconImg.(*image.NRGBA)
	iconWidth, iconHeight := trashIconImgData.Bounds().Max.X, trashIconImgData.Bounds().Max.Y
	deleteIcon = graphics.GetTextureUint8(iconWidth, iconHeight, 4, trashIconImgData.Pix, true)
	settingsBar.deleteIcon = deleteIcon
	
	return settingsBar
}

func (settingsBar *SettingsBar) AddSettings(texture graphics.Texture) {
	settingsBar.DeleteButtonColors = append(settingsBar.DeleteButtonColors, ColorParameter{settingsDeleteButtonColorInactive, settingsDeleteButtonColorInactive})
	settingsBar.SettingsColors     = append(settingsBar.SettingsColors, ColorParameter{settingsColor, settingsColor})
	settingsBar.SettingsTextures   = append(settingsBar.SettingsTextures, texture)
}

func (settingsBar *SettingsBar) RemoveSettings(index int) {
	settingsBar.SettingsTextures   = append(settingsBar.SettingsTextures[:index], settingsBar.SettingsTextures[index + 1:]...)
	settingsBar.SettingsColors     = append(settingsBar.SettingsColors[:index], settingsBar.SettingsColors[index + 1:]...)
	settingsBar.DeleteButtonColors = append(settingsBar.DeleteButtonColors[:index], settingsBar.DeleteButtonColors[index + 1:]...)
}

// TODO: hidden refers both to hidden positionaly and UI as a whole faded out - hidden
// TODO: settingsBar.height refers to "container" not content, ambigious especially in the Update function.
func (settingsBar *SettingsBar) Update(dt float64, mouseX, mouseY float32, hidden bool) (Action, int) {
	returnAction, returnIndex := NONE, -1

	barPos := mgl32.Vec2{float32(settingsBar.PosX.Val), 0}
	barSize := mgl32.Vec2{float32(settingsBarWidth), settingsBar.height}
	
	aspectRatio := float32(settingsBar.SettingsTextures[0].Width) / float32(settingsBar.SettingsTextures[1].Height)
	settingSizeX := barSize[0] - settingPadding * 2.0
	settingSizeY := settingSizeX / aspectRatio
	settingSize := mgl32.Vec2{settingSizeX, settingSizeY}
	hiddenPortion := barPos[0] / (settingsBarWidth - settingsBarWidthHidden)
	hiddenOffset := hiddenPortion * settingsBarWidthHidden
	saveSize := mgl32.Vec2{settingSizeX, saveButtonHeight}
	savePos := mgl32.Vec2{
		settingPadding + barPos[0] + hiddenOffset,
		settingPadding + float32(settingsBar.ContentY.Val),
	}
	mousePos := mgl32.Vec2{mouseX, mouseY}
	settingsCount := len(settingsBar.SettingsColors)

	// Input updates and drawing for the whole bar container.
	{
		if isInRect(mousePos, barPos, barSize) {
			settingsBar.PosX.Target = 0.0
			settingsBar.Color.Target = settingsBarColorHover
			
			scrollDelta := platform.GetMouseWheelDelta()
			settingsBar.ContentY.Target += scrollDelta * scrollSpeed
	
			// This is to make sure that there's not going to be a gap between top of the screen and top of the bar's content.
			if settingsBar.ContentY.Target > 0 {
				settingsBar.ContentY.Target = 0.0
			}
			
			// This is to make sure that there's not going to be a gap between bottom of the screen and bottom of the bar's content.
			saveButtonPartHeight := settingPadding + float64(saveSize[1]) + settingPadding
			settingsPartHeight := float64(settingsCount) * (float64(settingSizeY) + settingPadding)
			barContentHeight := saveButtonPartHeight + settingsPartHeight 
			barContentBottom := settingsBar.ContentY.Target + barContentHeight
			if barContentBottom < float64(settingsBar.height) {
				settingsBar.ContentY.Target = float64(settingsBar.height) - barContentHeight
			}
		} else {
			settingsBar.PosX.Target = settingsBarHiddenX
			if hidden {
				settingsBar.Color.Target = settingsBarColorInactive
			} else {
				settingsBar.Color.Target = settingsBarColor
			}
		}
		DrawUIRect(barPos, barSize, settingsBar.Color.Val, 0)
	}

	// Input updates and drawing for save button.
	{
		if isInRect(mousePos, savePos, saveSize) {
			settingsBar.SaveButtonColor.Target = saveButtonColorHover
			if platform.IsMouseLeftButtonPressed() {
				returnAction = SAVE
			}
		} else {
			settingsBar.SaveButtonColor.Target = saveButtonColor
		}
		DrawUIRect(savePos, saveSize, settingsBar.SaveButtonColor.Val, 0)
		
		saveTextPos    := mgl32.Vec2{savePos[0] + saveSize[0]*0.5, savePos[1] + saveSize[1]*0.5}
		saveTextOrigin := mgl32.Vec2{0.5, 0.5}
		DrawUIText("SAVE", &settingsBar.font, saveTextPos, saveTextColor, saveTextOrigin, 1)
	}

	// Input updates and drawing for saved settings.
	{
		for i := 0; i < settingsCount; i++ {
			settingX := settingPadding + barPos[0] + hiddenOffset
			settingY := float32(settingsBar.ContentY.Val) + (2 * settingPadding + saveSize[1]) +  float32(settingsCount - 1 - i) * (settingSize[1] + settingPadding)
			settingsPos := mgl32.Vec2{settingX, settingY}
	
			// Skip if out of screen.
			if settingY + settingSize[0] < 0.0 {
				break
			}
	
			deleteButtonSize := mgl32.Vec2{settingsDeleteButtonWidth, settingSize[1]}
			deleteButtonPos := mgl32.Vec2{settingsPos[0] + settingSize[0] - deleteButtonSize[0], settingsPos[1]}
			deleteButtonColor := settingsBar.DeleteButtonColors[i].Val
	
			if isInRect(mousePos, settingsPos, settingSize) {
				settingsBar.SettingsColors[i].Target = settingsColorHover
				
				if isInRect(mousePos, deleteButtonPos, deleteButtonSize) {
					settingsBar.DeleteButtonColors[i].Target = settingsDeleteButtonColorHover
					if platform.IsMouseLeftButtonPressed() {
						returnAction = DELETE
						returnIndex = i
					}
				} else {
					settingsBar.DeleteButtonColors[i].Target = settingsDeleteButtonColor
					if platform.IsMouseLeftButtonPressed() {
						returnAction = SELECT
						returnIndex = i
					}
				}
			} else {
				settingsBar.DeleteButtonColors[i].Target = settingsDeleteButtonColorInactive
				settingsBar.SettingsColors[i].Target = settingsColor
			}
			DrawUIRect(deleteButtonPos, deleteButtonSize, deleteButtonColor, 1)
	
			deleteIconOffsetX := (deleteButtonSize[0] - deleteIconSize[0]) * 0.5
			deleteIconOffsetY := (deleteButtonSize[1] - deleteIconSize[1]) * 0.5
			deleteIconPos := mgl32.Vec2{deleteButtonPos[0] + deleteIconOffsetX, deleteButtonPos[1] + deleteIconOffsetY}
			DrawUIRectTextured(deleteIconPos, deleteIconSize, settingsBar.deleteIcon, mgl32.Vec4{1, 1, 1, deleteButtonColor[3]}, 1)
	
			texture := settingsBar.SettingsTextures[i]
			DrawUIRectTextured(settingsPos, settingSize, texture, settingsBar.SettingsColors[i].Val, 0)
		}
	}

	// Update parameters.
	// TODO: constants
	for i := range settingsBar.DeleteButtonColors {
		settingsBar.DeleteButtonColors[i].Update(dt, 8.0)
		settingsBar.SettingsColors[i].Update(dt, 8.0)
	}
	settingsBar.PosX.Update(dt, 10.0)
	settingsBar.ContentY.Update(dt, 10.0)
	settingsBar.Color.Update(dt, 4.0)
	settingsBar.SaveButtonColor.Update(dt, 4.0)

	return returnAction, returnIndex
}

func isInRect(position mgl32.Vec2, rectPosition mgl32.Vec2, rectSize mgl32.Vec2) bool {
	if position[0] >= rectPosition[0] && position[0] <= rectPosition[0]+rectSize[0] &&
		position[1] >= rectPosition[1] && position[1] <= rectPosition[1]+rectSize[1] {
		return true
	}
	return false
}
