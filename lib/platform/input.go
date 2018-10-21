package platform

import (
	"github.com/go-gl/glfw/v3.2/glfw"
)

var mouseX, mouseY float64 = -1.0, -1.0
var dmouseX, dmouseY float64 = 0.0, 0.0
var mouseButtonDown bool = false
var mouseButtonPressed bool = false
var mouseWheelDelta float64 = 0.0
var keyDown map[Key]bool
var keyPressed map[Key]bool

func scrollCallback(w *glfw.Window, xoff float64, yoff float64) {
	mouseWheelDelta = yoff
}

var inputString string
var currentInputString string
func charCallback(w *glfw.Window, r rune) {
	inputString += string(r)
}

func initInput(window *glfw.Window) {
	window.SetScrollCallback(scrollCallback)
	window.SetCharCallback(charCallback)

	keyDown = make(map[Key]bool)
	keyPressed = make(map[Key]bool)
}

func Update(window *glfw.Window) {
	glfw.PollEvents()
	
	mouseButtonPressed = false

	// Update input string.
	currentInputString = inputString
	inputString = ""

	// Update mouse position and get position delta.
	x, y := window.GetCursorPos()
	x, y = x / windowScale, y / windowScale
	if mouseX > 0.0 && mouseY > 0.0 {
		dmouseX, dmouseY = x - mouseX, y - mouseY
	}
	mouseX, mouseY = x, y
	lButtonState := window.GetMouseButton(glfw.MouseButtonLeft)
	if lButtonState == glfw.Press {
		if !mouseButtonDown {
			mouseButtonPressed = true
		}
		mouseButtonDown = true
	} else {
		mouseButtonDown = false
	}

	for key, _ := range keyPressed {
		keyPressed[key] = false
	}

	for _, key := range keys {
		state := window.GetKey(glfw.Key(key))
		if state == glfw.Press {
			if !keyDown[key] {
				keyPressed[key] = true
			}
			keyDown[key] = true
		} else {
			keyDown[key] = false
		}
	}
}

func IsKeyPressed(key Key) bool {
	return keyPressed[key]
}

func IsKeyDown(key Key) bool {
	return keyDown[key]
}

func GetCharsPressed() string {
	result := currentInputString
	return result
}

func GetMouseDeltaPosition() (float64, float64) {
	return dmouseX, dmouseY
}

func GetMousePosition() (float64, float64) {
	return mouseX, mouseY
}

func IsMouseLeftButtonPressed() bool {
	return mouseButtonPressed
}

func IsMouseLeftButtonDown() bool {
	return mouseButtonDown
}

func GetMouseWheelDelta() float64 {
	result := mouseWheelDelta
	mouseWheelDelta = 0.0
	return result
}

type Key int

const(
	KeyUnknown      Key = Key(glfw.KeyUnknown)
	KeySpace        Key = Key(glfw.KeySpace)
	KeyApostrophe   Key = Key(glfw.KeyApostrophe)
	KeyComma        Key = Key(glfw.KeyComma)
	KeyMinus        Key = Key(glfw.KeyMinus)
	KeyPeriod       Key = Key(glfw.KeyPeriod)
	KeySlash        Key = Key(glfw.KeySlash)
	Key0            Key = Key(glfw.Key0)
	Key1            Key = Key(glfw.Key1)
	Key2            Key = Key(glfw.Key2)
	Key3            Key = Key(glfw.Key3)
	Key4            Key = Key(glfw.Key4)
	Key5            Key = Key(glfw.Key5)
	Key6            Key = Key(glfw.Key6)
	Key7            Key = Key(glfw.Key7)
	Key8            Key = Key(glfw.Key8)
	Key9            Key = Key(glfw.Key9)
	KeySemicolon    Key = Key(glfw.KeySemicolon)
	KeyEqual        Key = Key(glfw.KeyEqual)
	KeyA            Key = Key(glfw.KeyA)
	KeyB            Key = Key(glfw.KeyB)
	KeyC            Key = Key(glfw.KeyC)
	KeyD            Key = Key(glfw.KeyD)
	KeyE            Key = Key(glfw.KeyE)
	KeyF            Key = Key(glfw.KeyF)
	KeyG            Key = Key(glfw.KeyG)
	KeyH            Key = Key(glfw.KeyH)
	KeyI            Key = Key(glfw.KeyI)
	KeyJ            Key = Key(glfw.KeyJ)
	KeyK            Key = Key(glfw.KeyK)
	KeyL            Key = Key(glfw.KeyL)
	KeyM            Key = Key(glfw.KeyM)
	KeyN            Key = Key(glfw.KeyN)
	KeyO            Key = Key(glfw.KeyO)
	KeyP            Key = Key(glfw.KeyP)
	KeyQ            Key = Key(glfw.KeyQ)
	KeyR            Key = Key(glfw.KeyR)
	KeyS            Key = Key(glfw.KeyS)
	KeyT            Key = Key(glfw.KeyT)
	KeyU            Key = Key(glfw.KeyU)
	KeyV            Key = Key(glfw.KeyV)
	KeyW            Key = Key(glfw.KeyW)
	KeyX            Key = Key(glfw.KeyX)
	KeyY            Key = Key(glfw.KeyY)
	KeyZ            Key = Key(glfw.KeyZ)
	KeyLeftBracket  Key = Key(glfw.KeyLeftBracket)
	KeyBackslash    Key = Key(glfw.KeyBackslash)
	KeyRightBracket Key = Key(glfw.KeyRightBracket)
	KeyGraveAccent  Key = Key(glfw.KeyGraveAccent)
	KeyWorld1       Key = Key(glfw.KeyWorld1)
	KeyWorld2       Key = Key(glfw.KeyWorld2)
	KeyEscape       Key = Key(glfw.KeyEscape)
	KeyEnter        Key = Key(glfw.KeyEnter)
	KeyTab          Key = Key(glfw.KeyTab)
	KeyBackspace    Key = Key(glfw.KeyBackspace)
	KeyInsert       Key = Key(glfw.KeyInsert)
	KeyDelete       Key = Key(glfw.KeyDelete)
	KeyRight        Key = Key(glfw.KeyRight)
	KeyLeft         Key = Key(glfw.KeyLeft)
	KeyDown         Key = Key(glfw.KeyDown)
	KeyUp           Key = Key(glfw.KeyUp)
	KeyPageUp       Key = Key(glfw.KeyPageUp)
	KeyPageDown     Key = Key(glfw.KeyPageDown)
	KeyHome         Key = Key(glfw.KeyHome)
	KeyEnd          Key = Key(glfw.KeyEnd)
	KeyCapsLock     Key = Key(glfw.KeyCapsLock)
	KeyScrollLock   Key = Key(glfw.KeyScrollLock)
	KeyNumLock      Key = Key(glfw.KeyNumLock)
	KeyPrintScreen  Key = Key(glfw.KeyPrintScreen)
	KeyPause        Key = Key(glfw.KeyPause)
	KeyF1           Key = Key(glfw.KeyF1)
	KeyF2           Key = Key(glfw.KeyF2)
	KeyF3           Key = Key(glfw.KeyF3)
	KeyF4           Key = Key(glfw.KeyF4)
	KeyF5           Key = Key(glfw.KeyF5)
	KeyF6           Key = Key(glfw.KeyF6)
	KeyF7           Key = Key(glfw.KeyF7)
	KeyF8           Key = Key(glfw.KeyF8)
	KeyF9           Key = Key(glfw.KeyF9)
	KeyF10          Key = Key(glfw.KeyF10)
	KeyF11          Key = Key(glfw.KeyF11)
	KeyF12          Key = Key(glfw.KeyF12)
	KeyF13          Key = Key(glfw.KeyF13)
	KeyF14          Key = Key(glfw.KeyF14)
	KeyF15          Key = Key(glfw.KeyF15)
	KeyF16          Key = Key(glfw.KeyF16)
	KeyF17          Key = Key(glfw.KeyF17)
	KeyF18          Key = Key(glfw.KeyF18)
	KeyF19          Key = Key(glfw.KeyF19)
	KeyF20          Key = Key(glfw.KeyF20)
	KeyF21          Key = Key(glfw.KeyF21)
	KeyF22          Key = Key(glfw.KeyF22)
	KeyF23          Key = Key(glfw.KeyF23)
	KeyF24          Key = Key(glfw.KeyF24)
	KeyF25          Key = Key(glfw.KeyF25)
	KeyKP0          Key = Key(glfw.KeyKP0)
	KeyKP1          Key = Key(glfw.KeyKP1)
	KeyKP2          Key = Key(glfw.KeyKP2)
	KeyKP3          Key = Key(glfw.KeyKP3)
	KeyKP4          Key = Key(glfw.KeyKP4)
	KeyKP5          Key = Key(glfw.KeyKP5)
	KeyKP6          Key = Key(glfw.KeyKP6)
	KeyKP7          Key = Key(glfw.KeyKP7)
	KeyKP8          Key = Key(glfw.KeyKP8)
	KeyKP9          Key = Key(glfw.KeyKP9)
	KeyKPDecimal    Key = Key(glfw.KeyKPDecimal)
	KeyKPDivide     Key = Key(glfw.KeyKPDivide)
	KeyKPMultiply   Key = Key(glfw.KeyKPMultiply)
	KeyKPSubtract   Key = Key(glfw.KeyKPSubtract)
	KeyKPAdd        Key = Key(glfw.KeyKPAdd)
	KeyKPEnter      Key = Key(glfw.KeyKPEnter)
	KeyKPEqual      Key = Key(glfw.KeyKPEqual)
	KeyLeftShift    Key = Key(glfw.KeyLeftShift)
	KeyLeftControl  Key = Key(glfw.KeyLeftControl)
	KeyLeftAlt      Key = Key(glfw.KeyLeftAlt)
	KeyLeftSuper    Key = Key(glfw.KeyLeftSuper)
	KeyRightShift   Key = Key(glfw.KeyRightShift)
	KeyRightControl Key = Key(glfw.KeyRightControl)
	KeyRightAlt     Key = Key(glfw.KeyRightAlt)
	KeyRightSuper   Key = Key(glfw.KeyRightSuper)
	KeyMenu         Key = Key(glfw.KeyMenu)
	KeyLast         Key = Key(glfw.KeyLast)
)

var keys = []Key {
	KeySpace,
	KeyApostrophe,
	KeyComma,
	KeyMinus,
	KeyPeriod,
	KeySlash,
	Key0,
	Key1,
	Key2,
	Key3,
	Key4,
	Key5,
	Key6,
	Key7,
	Key8,
	Key9,
	KeySemicolon,
	KeyEqual,
	KeyA,
	KeyB,
	KeyC,
	KeyD,
	KeyE,
	KeyF,
	KeyG,
	KeyH,
	KeyI,
	KeyJ,
	KeyK,
	KeyL,
	KeyM,
	KeyN,
	KeyO,
	KeyP,
	KeyQ,
	KeyR,
	KeyS,
	KeyT,
	KeyU,
	KeyV,
	KeyW,
	KeyX,
	KeyY,
	KeyZ,
	KeyLeftBracket,
	KeyBackslash,
	KeyRightBracket,
	KeyGraveAccent,
	KeyWorld1,
	KeyWorld2,
	KeyEscape,
	KeyEnter,
	KeyTab,
	KeyBackspace,
	KeyInsert,
	KeyDelete,
	KeyRight,
	KeyLeft,
	KeyDown,
	KeyUp,
	KeyPageUp,
	KeyPageDown,
	KeyHome,
	KeyEnd,
	KeyCapsLock,
	KeyScrollLock,
	KeyNumLock,
	KeyPrintScreen,
	KeyPause,
	KeyF1,
	KeyF2,
	KeyF3,
	KeyF4,
	KeyF5,
	KeyF6,
	KeyF7,
	KeyF8,
	KeyF9,
	KeyF10,
	KeyF11,
	KeyF12,
	KeyF13,
	KeyF14,
	KeyF15,
	KeyF16,
	KeyF17,
	KeyF18,
	KeyF19,
	KeyF20,
	KeyF21,
	KeyF22,
	KeyF23,
	KeyF24,
	KeyF25,
	KeyKP0,
	KeyKP1,
	KeyKP2,
	KeyKP3,
	KeyKP4,
	KeyKP5,
	KeyKP6,
	KeyKP7,
	KeyKP8,
	KeyKP9,
	KeyKPDecimal,
	KeyKPDivide,
	KeyKPMultiply,
	KeyKPSubtract,
	KeyKPAdd,
	KeyKPEnter,
	KeyKPEqual,
	KeyLeftShift,
	KeyLeftControl,
	KeyLeftAlt,
	KeyLeftSuper,
	KeyRightShift,
	KeyRightControl,
	KeyRightAlt,
	KeyRightSuper,
	KeyMenu,
	KeyLast,
}