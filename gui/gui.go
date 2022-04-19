package gui

import (

	"github.com/veandco/go-sdl2/sdl"
)

type MouseState struct {
	LeftButton, RightButton bool
	PrevLeftButton, PrevRightButton bool
	PrevX, PrevY int
	X, Y        int
}

func GetMouseState() *MouseState {
	mouseX, mouseY, mouseButtonState := sdl.GetMouseState()
	leftButton := mouseButtonState & sdl.ButtonLMask()
	rightButton := mouseButtonState & sdl.ButtonRMask()
	var result MouseState
	result.X = int(mouseX)
	result.Y = int(mouseY)
	result.LeftButton = !(leftButton == 0)
	result.RightButton = !(rightButton == 0)
	return &result
}

func (mouseState *MouseState) Update() {
	mouseState.PrevLeftButton = mouseState.LeftButton
	mouseState.PrevRightButton = mouseState.RightButton
	mouseState.PrevX = mouseState.X
	mouseState.PrevY = mouseState.Y

	X, Y, mouseButtonState := sdl.GetMouseState()
	mouseState.X = int(X)
	mouseState.Y = int(Y)
	mouseState.LeftButton = mouseButtonState & sdl.ButtonLMask() != 0
	mouseState.RightButton = mouseButtonState & sdl.ButtonRMask() != 0
}

type ImageButton struct {
	Image, SelectedTex *sdl.Texture
	Rect sdl.Rect
	WasLeftClicked, WasRightClicked, IsSelected bool
}

func NewImageButton(renderer *sdl.Renderer, image *sdl.Texture, rect sdl.Rect, selectedColor sdl.Color) *ImageButton {
	tex, err := renderer.CreateTexture(sdl.PIXELFORMAT_ABGR8888, sdl.TEXTUREACCESS_STATIC, 1, 1)
	if err != nil {
		panic(err)
	}
	pixels := make([]byte, 4)
	pixels[0] = selectedColor.R
	pixels[1] = selectedColor.G
	pixels[2] = selectedColor.B
	pixels[3] = selectedColor.A
	tex.Update(nil, pixels, 4)
	return &ImageButton{image, tex, rect, false, false, false}
}

func (button *ImageButton) Update(mouseState *MouseState) {
	if button.Rect.HasIntersection(&sdl.Rect{int32(mouseState.X),int32(mouseState.Y), 1, 1}) {
		button.WasLeftClicked = mouseState.PrevLeftButton && !mouseState.LeftButton
		button.WasRightClicked = mouseState.PrevRightButton && !mouseState.RightButton
	} else {
		button.WasLeftClicked = false
		button.WasRightClicked = false
	}
}

func (button *ImageButton) Draw(renderer *sdl.Renderer) {
	if button.IsSelected {
		borderRect := button.Rect
		thickness := int32(float32(borderRect.W)*.01)
		borderRect.W = button.Rect.W+(thickness*2)
		borderRect.H = button.Rect.H+(thickness*2)
		borderRect.X -= thickness
		borderRect.Y -= thickness
		renderer.Copy(button.SelectedTex, nil, &borderRect)
	}
	renderer.Copy(button.Image, nil, &button.Rect)
}

func GetSinglePicText(renderer *sdl.Renderer, color sdl.Color) *sdl.Texture {
	tex, err := renderer.CreateTexture(sdl.PIXELFORMAT_ABGR8888, sdl.TEXTUREACCESS_STATIC, 1, 1)
	if err != nil {
		panic(err)
	}
	pixels := make([]byte, 4)
	pixels[0] = color.R
	pixels[1] = color.G
	pixels[2] = color.B
	pixels[3] = color.A
	tex.Update(nil, pixels, 4)
	return tex
}