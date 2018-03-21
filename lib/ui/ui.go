package ui

import (
	"../graphics"
	gmath "../math"
	"../font"
	"../input"
	"github.com/go-gl/gl/v4.1-core/gl"
	"io/ioutil"
    "fmt"
    "sort"
    "math"
)

var quadVertices = [...]float32{
	-0.5, -0.5, 0.0, 1.0,
	0.0, 0.0,
	-0.5, 0.5, 0.0, 1.0,
	0.0, 1.0,
	0.5, 0.5, 0.0, 1.0,
	1.0, 1.0,
	0.5, -0.5, 0.0, 1.0,
	1.0, 0.0,
}

var quadIndices = [...]uint32{
	0, 1, 2,
	0, 2, 3,
}

var program_text graphics.Program
var textModelMatrixUniform graphics.Uniform
var projectionMatrixTextUniform graphics.Uniform
var sourceRectUniform graphics.Uniform
var textColorUniform graphics.Uniform

var program_rect graphics.Program
var projectionMatrixRectUniform graphics.Uniform
var rectColorUniform graphics.Uniform
var rectModelMatrixUniform graphics.Uniform

var quad graphics.Mesh
var uiFont font.Font

var colorForeground gmath.Vec4 = gmath.Vec4{0,1,0,1}
var colorBackground gmath.Vec4 = gmath.Vec4{0,0,1,1}
var colorLabel gmath.Vec4 = gmath.Vec4{1,0,0,1}

var isInputResponsive bool = true


type textRenderingData struct {
    text string
    position gmath.Vec2
    color gmath.Vec4
}

type rectRenderingData struct {
    position gmath.Vec2
    size gmath.Vec2
    color gmath.Vec4
    layer int
}

type rectRenderingDataList []rectRenderingData
func (rectList rectRenderingDataList)Less(i, j int) bool {
    return rectList[i].layer < rectList[j].layer
}

var textRenderingBuffer []textRenderingData
var rectRenderingBuffer []rectRenderingData

func Init() {
	truetypeBytes, err := ioutil.ReadFile("font.ttf")
	if err != nil {
		panic(err)
	}

	uiFont = font.GetFont(truetypeBytes, 20.0)
	
	// Text rendering shaders
	vertexShaderData, err := ioutil.ReadFile("shaders/text_vertex_shader.glsl")
	vertexShader, err := graphics.GetShader(string(vertexShaderData), graphics.VERTEX_SHADER)
	if err != nil {
		fmt.Println(err)
	}

	pixelShaderData, err := ioutil.ReadFile("shaders/text_pixel_shader.glsl")
	pixelShader, err := graphics.GetShader(string(pixelShaderData), graphics.PIXEL_SHADER)
	if err != nil {
		fmt.Println(err)
	}

	program_text, err = graphics.GetProgram(vertexShader, pixelShader)
	if err != nil {
		fmt.Println(err)
	}

	graphics.ReleaseShaders(vertexShader, pixelShader)
	graphics.SetProgram(program_text)


	projectionMatrix := gmath.GetOrthographicProjectionGLRH(-10.0, 10.0, -10.0, 10.0, -10.0, 10.0)
	projectionMatrixTextUniform = graphics.GetUniform(program_text, "projection_matrix")
	graphics.SetUniformMatrix(projectionMatrixTextUniform, projectionMatrix)

	sourceRectUniform = graphics.GetUniform(program_text, "source_rect")
	textModelMatrixUniform = graphics.GetUniform(program_text, "model_matrix")
	textColorUniform = graphics.GetUniform(program_text, "color")

	// Rect rendering shaders
	vertexShaderData, err = ioutil.ReadFile("shaders/rect_vertex_shader.glsl")
	vertexShader, err = graphics.GetShader(string(vertexShaderData), graphics.VERTEX_SHADER)
	if err != nil {
		fmt.Println(err)
	}

	pixelShaderData, err = ioutil.ReadFile("shaders/rect_pixel_shader.glsl")
	pixelShader, err = graphics.GetShader(string(pixelShaderData), graphics.PIXEL_SHADER)
	if err != nil {
		fmt.Println(err)
	}

	program_rect, err = graphics.GetProgram(vertexShader, pixelShader)
	if err != nil {
		fmt.Println(err)
	}

	graphics.ReleaseShaders(vertexShader, pixelShader)
	graphics.SetProgram(program_rect)

	projectionMatrixRectUniform = graphics.GetUniform(program_rect, "projection_matrix")
	rectColorUniform = graphics.GetUniform(program_rect, "color")
	rectModelMatrixUniform = graphics.GetUniform(program_rect, "model_matrix")
	
    quad = graphics.GetMesh(quadVertices[:], quadIndices[:], []int{4, 2})
    
    textRenderingBuffer = make([]textRenderingData, 100)
    rectRenderingBuffer = make([]rectRenderingData, 100)
}


func DrawText(text string, font *font.Font, position gmath.Vec2, color gmath.Vec4, origin gmath.Vec2) {

	graphics.SetProgram(program_text)
	projectionMatrix := gmath.GetOrthographicProjectionGLRH(0.0, 1920.0, 0.0, 1080, 10.0, -10.0)
	graphics.SetUniformMatrix(projectionMatrixTextUniform, projectionMatrix)
	gl.Disable(gl.DEPTH_TEST)
	gl.Enable(gl.BLEND)
	gl.BlendFunc(gl.SRC_ALPHA, gl.ONE_MINUS_SRC_ALPHA)
	
	width, height := font.GetStringWidth(text), font.RowHeight
	x := float64(position[0]) - width * float64(origin[0])
	y := float64(position[1]) + float64(font.TopPad) - float64(height) * float64(origin[1])
	for _, char := range text {
		glyphA := font.Glyphs[char]
		relX := float32(glyphA.X) / 256.0
		relY := 1.0 - float32(glyphA.Y + glyphA.BitmapHeight) / 256.0
		relWidth := float32(glyphA.BitmapWidth) / 256.0
		relHeight := float32(glyphA.BitmapHeight) / 256.0
		sourceRect := gmath.Vec4{relX,relY,relWidth,relHeight}
		graphics.SetUniformVec4(sourceRectUniform, sourceRect)
		graphics.SetUniformVec4(textColorUniform, color)

		currentX := x + float64(glyphA.XOffset)
		currentY := y + float64(glyphA.YOffset)
		modelMatrix := gmath.Matmul(
			gmath.GetTranslation(currentX, 1080 - currentY, 0),
			gmath.Matmul(
				gmath.GetScale(float64(glyphA.BitmapWidth), float64(glyphA.BitmapHeight), 1.0),
				gmath.GetTranslation(0.5, -0.5, 0.0),
			),
		)
		graphics.SetUniformMatrix(textModelMatrixUniform, modelMatrix)			
		
		graphics.DrawMesh(quad)
		
		x += float64(glyphA.Advance)
	}
}

func DrawRect(pos gmath.Vec2, size gmath.Vec2, color gmath.Vec4) {
	graphics.SetProgram(program_rect)
	gl.Disable(gl.DEPTH_TEST)
	gl.Enable(gl.BLEND)
	gl.BlendFunc(gl.SRC_ALPHA, gl.ONE_MINUS_SRC_ALPHA)
	projectionMatrix := gmath.GetOrthographicProjectionGLRH(0.0, 1920.0, 0.0, 1080, 10.0, -10.0)
	graphics.SetUniformMatrix(projectionMatrixRectUniform, projectionMatrix)
	gl.Enable(gl.BLEND)
	gl.BlendFunc(gl.SRC_ALPHA, gl.ONE_MINUS_SRC_ALPHA)
	
	graphics.SetUniformVec4(rectColorUniform, color)
	
	x, y := pos[0], 1080 - pos[1]
	modelMatrix := gmath.Matmul(
		gmath.GetTranslation(float64(x), float64(y), 0),
		gmath.Matmul(
			gmath.GetScale(float64(size[0]), float64(size[1]), 1.0),
			gmath.GetTranslation(0.5, -0.5, 0.0),
		),
	)
	graphics.SetUniformMatrix(rectModelMatrixUniform, modelMatrix)
		
	graphics.DrawMesh(quad)
}

func (panel *Panel) AddToggle(label string, active bool) (newValue bool, changed bool) {
	changed = false
	newValue = active

    boxMiddleToTotal := 0.6
    height := float64(uiFont.RowHeight)
    
	itemPos := gmath.Vec2{panel.position[0] + panel.itemPos[0], panel.position[1] + panel.itemPos[1]}
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
        fgBoxPos := gmath.Vec2{
			bgBoxPos[0] + (bgBoxSize[0] - fgBoxSize[0]) / 2.0,
			bgBoxPos[1] + (bgBoxSize[1] - fgBoxSize[1]) / 2.0,
		}
		
        rectRenderingBuffer = append(rectRenderingBuffer, rectRenderingData {
            fgBoxPos, fgBoxSize, middleColor, 1,
        })
    }

	// Draw toggle label
	innerPadding := float32(10.0)
	textPos := gmath.Vec2{bgBoxPos[0] + innerPadding + bgBoxSize[0], bgBoxPos[1]}
    textRenderingBuffer = append(textRenderingBuffer, textRenderingData {
        label, textPos, colorLabel,
    })
    textWidth := uiFont.GetStringWidth(label)
	
    // Move current panel item position
    panel.itemPos[1] += float32(height) + innerPadding
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
        panel.name, titlePos, colorBackground,
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

func Present() {
    sort.Slice(rectRenderingBuffer, func(i, j int) bool {
        return rectRenderingBuffer[i].layer < rectRenderingBuffer[j].layer
      })

    for _, rectData := range rectRenderingBuffer {
        DrawRect(rectData.position, rectData.size, rectData.color)
    }

    rectRenderingBuffer = rectRenderingBuffer[:0]

    for _, textData := range textRenderingBuffer {
        DrawText(textData.text, &uiFont, textData.position, textData.color, gmath.Vec2{0,0})
    }

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

var isRegisteringInput bool = false
func setActive(itemID int) {
    if activeID == -1 {
        activeID = itemID
        isRegisteringInput = true;
    }
}

func unsetActive(itemID int) {
    if activeID == itemID {
        activeID = -1
        isRegisteringInput = false
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

/*const float vertical_padding = 15.0f;
const float horizontal_padding = 15.0f;
const float inner_padding = 10.0f;
static int32_t active_id = -1;
static int32_t hot_id = -1;


#define COLOR(r,g,b,a) Vector4((r) / 255.0f, (g) / 255.0f, (b) / 255.0f, (a) / 255.0f)
// This means that the color has been NOT gamma corrected - it was seen displayed incorrectly.
#define COLOR_LINEAR(r,g,b,a) Vector4(math::pow((r) / 255.0f, 2.2f), math::pow((g) / 255.0f, 2.2f), math::pow((b) / 255.0f, 2.2f), math::pow((a) / 255.0f, 2.2f))
static Vector4 color_background = Vector4(0.1f, 0.1f, 0.1f, 1.0f);
//static Vector4 color_foreground = Vector4(1.0f, 1.0f, 1.0f, 0.6f);
static Vector4 color_foreground = COLOR_LINEAR(28, 224, 180, 255);//Vector4(1.0f, 1.0f, 1.0f, 0.6f);
static Vector4 color_title = COLOR_LINEAR(28, 224, 180, 255);//Vector4(1.0f, 1.0f, 1.0f, 0.6f);
static Vector4 color_label = Vector4(1.0f, 1.0f, 1.0f, 0.6f);

Panel ui::start_panel(char *name, Vector2 pos, float width)
{
    Panel panel = {};

    panel.pos = pos;
    panel.width = width;
    panel.item_pos.x = horizontal_padding;
    panel.item_pos.y = font::get_row_height(&font_ui) + vertical_padding * 2.0f;
    panel.name = name;
    
    return panel;
}

Panel ui::start_panel(char *name, float x, float y, float width)
{
    return ui::start_panel(name, Vector2(x,y), width);
}

void ui::end_panel(Panel *panel)
{
    float panel_height = panel->item_pos.y + vertical_padding - inner_padding;
    RectItem panel_bg = {color_background, panel->pos, Vector2(panel->width, panel_height)};
    array::add(&rect_items_bg, panel_bg);

    float title_bar_height = font::get_row_height(&font_ui) + inner_padding * 2;
    RectItem title_bar = {color_foreground, panel->pos, Vector2(panel->width, title_bar_height)};
    array::add(&rect_items_bg, title_bar);

    Vector2 title_pos = panel->pos + Vector2(horizontal_padding, inner_padding);
    TextItem title = {};
    title.color = color_background;
    title.pos = title_pos;
    memcpy(title.text, panel->name, strlen(panel->name) + 1);
    array::add(&text_items, title);
}

bool ui::add_slider(Panel *panel, char *label, float *pos, float min, float max)
{
    bool changed = false;
    int32_t slider_id = hash_string(label);
    Vector2 item_pos = panel->pos + panel->item_pos;
    float height = font::get_row_height(&font_ui);
    float slider_width = 200.0f;

    // Slider bar
    float slider_start = 0.0f;

    Vector4 slider_bar_color = color_background * 2.0f;
    Vector2 slider_bar_pos = item_pos + Vector2(slider_start, 0.0f);
    Vector2 slider_bar_size = Vector2(slider_width, height);
    RectItem slider_bar = { slider_bar_color, slider_bar_pos, slider_bar_size };
    array::add(&rect_items, slider_bar);

    Vector4 slider_color = color_foreground;
    Vector2 slider_size = Vector2(height, height);
    float slider_x = (*pos - min) / (max - min) * (slider_width - slider_size.x) + slider_bar_pos.x + slider_size.x * 0.5f;
    Vector2 slider_pos = Vector2(slider_x - slider_size.x * 0.5f, item_pos.y);

    // Max number
    Vector2 current_pos = Vector2(slider_bar_pos.x + slider_bar_size.x / 2.0f, item_pos.y);
    TextItem current_label = {};
    current_label.color = color_label;
    current_label.pos = current_pos; 
    current_label.origin = Vector2(0.5f, 0.0f);
    sprintf_s(current_label.text, ARRAYSIZE(current_label.text), "%.2f", *pos);
    array::add(&text_items, current_label);

    // Check for mouse input
    if(ui::is_input_responsive())
    {
        Vector2 mouse_position = input::mouse_position();
        if(is_in_rect(mouse_position, slider_pos, slider_size))
        {       
            set_hot(slider_id);
        }
        else if(!is_active(slider_id))
        {
            unset_hot(slider_id);
        }

        Vector2 overall_slider_size = Vector2(slider_bar_size.x, slider_size.y);
        Vector2 overall_slider_pos = Vector2(slider_bar_pos.x, slider_pos.y);

        if((is_hot(slider_id) || is_in_rect(mouse_position, overall_slider_pos, overall_slider_size)) && !is_active(slider_id) && input::mouse_left_button_down())
        {
            set_active(slider_id);
        }
        else if (is_active(slider_id) && !input::mouse_left_button_down())
        {
            unset_active(slider_id);
        }
    }
    else 
    {
        unset_hot(slider_id);
        unset_active(slider_id);
    }

    if (is_hot(slider_id))
    {
        slider_color *= 0.8f;
        slider_color.w = 1.0f;
    }

    if(is_active(slider_id))
    {
        float mouse_x = input::mouse_position().x;
        float mouse_x_rel = (mouse_x - slider_bar_pos.x - slider_size.x * 0.5f) / (slider_bar_size.x - slider_size.x);
        mouse_x_rel = math::clamp(mouse_x_rel, 0.0f, 1.0f);
        
        *pos = mouse_x_rel * (max - min) + min;

        changed = true;
    }

    // Slider
    RectItem slider = { slider_color, slider_pos, slider_size };
    array::add(&rect_items, slider);

    // Slider label
    Vector2 text_pos = Vector2(slider_bar_size.x + slider_bar_pos.x + inner_padding, item_pos.y);
    TextItem slider_label = {};
    slider_label.color = color_label;
    slider_label.pos = text_pos;
    memcpy(slider_label.text, label, strlen(label) + 1);
    array::add(&text_items, slider_label);

    panel->item_pos.y += height + inner_padding;
    return changed;
}

void ui::set_input_responsive(bool is_responsive)
{
    is_input_rensposive_ = is_responsive;
}

bool ui::is_input_responsive()
{
    return is_input_rensposive_;
}

bool ui::is_registering_input()
{
    return is_registering_input_;
}

float ui::get_screen_width()
{
    return screen_width;
}

Font *ui::get_font()
{
    return &font_ui;
}


*/