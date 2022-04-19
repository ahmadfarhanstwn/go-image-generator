package main

import (
	"fmt"
	"math/rand"
	"time"

	. "github.com/ahmadfarhanstwn/evolving-pictures/apt"
	. "github.com/ahmadfarhanstwn/evolving-pictures/gui"
	"github.com/veandco/go-sdl2/sdl"
)

var winWidth, winHeight, rows, columns, numPics int = 800, 600, 3, 3, rows*columns

type audioState struct {
	explosionBytes []byte
	deviceID       sdl.AudioDeviceID
	audioSpec      *sdl.AudioSpec
}

type pixelsTexture struct {
	pixels []byte
	num int
}

type rgba struct {
	r, g, b byte
}

type picture struct {
	r, g, b Node
}

func (p *picture) String() string {
	return "r:" + p.r.String() + ", g:" + p.g.String() + ", b:" + p.b.String()
}

func (p *picture) Mutate() {
	r := rand.Intn(3)
	var mutateNode Node
	switch r {
	case 0:
		mutateNode = p.r
	case 1:
		mutateNode = p.g
	case 2:
		mutateNode = p.b
	}

	count := mutateNode.CountNode()
	r = rand.Intn(count)
	mutateNode, count = GetNthChildren(mutateNode, r, 0)
	mutation := Mutate(mutateNode)
	if mutateNode == p.r {
		p.r = mutation
	} else if mutateNode == p.g {
		p.g = mutation
	} else if mutateNode == p.b {
		p.b = mutation
	}
}

func newPicture() *picture {
	p := &picture{}

	p.r = GetRandomNodeOpt()
	p.g = GetRandomNodeOpt()
	p.b = GetRandomNodeOpt()

	//operation type
	r := rand.Intn(20) + 10
	for i := 0; i < r; i++ {
		p.r.AddRandom(GetRandomNodeOpt())
	}

	r = rand.Intn(20) + 10
	for i := 0; i < r; i++ {
		p.g.AddRandom(GetRandomNodeOpt())
	}

	r = rand.Intn(20) + 10
	for i := 0; i < r; i++ {
		p.b.AddRandom(GetRandomNodeOpt())
	}

	//leaf node
	for p.r.AddLeaf(GetRandomLeafNode()){}

	for p.g.AddLeaf(GetRandomLeafNode()){}

	for p.b.AddLeaf(GetRandomLeafNode()){}

	return p
}

func clear(pixels []byte) {
	for i := range pixels {
		pixels[i] = 0
	}
}

func setPixel(x, y int, c rgba, pixels []byte) {
	index := (y*winWidth + x) * 4
	if index < len(pixels)-4 && index >= 0 {
		pixels[index] = c.r
		pixels[index+1] = c.g
		pixels[index+2] = c.b
	}

}

func pixelsToTexture(renderer *sdl.Renderer, pixels []byte, w, h int) *sdl.Texture {
	tex, err := renderer.CreateTexture(sdl.PIXELFORMAT_ABGR8888, sdl.TEXTUREACCESS_STREAMING, int32(w), int32(h))
	if err != nil {
		panic(err)
	}
	tex.Update(nil, pixels, w*4)
	return tex
}

func AptToPixels (p *picture, w, h int, renderer *sdl.Renderer) []byte {
	scale := float32(255/2)
	offset := float32(-1.0*scale)
	pixels := make([]byte, w*h*4)
	pixelIndex := 0
	for yi := 0; yi < h; yi++ {
		y := float32(yi)/float32(h)*2-1
		for xi := 0; xi < w; xi++ {
			x := float32(xi)/float32(w)*2-1
			c := p.r.Eval(x,y)
			c2 := p.g.Eval(x,y)
			c3 := p.b.Eval(x,y)
			pixels[pixelIndex] = byte(c*scale-offset)
			pixelIndex++
			pixels[pixelIndex] = byte(c2*scale-offset)
			pixelIndex++
			pixels[pixelIndex] = byte(c3*scale-offset)
			pixelIndex++
			pixelIndex++
		}
	}
	return pixels
}

func lerp(b1 byte, b2 byte, pct float32) byte {
	return byte(float32(b1) + pct*(float32(b2)-float32(b1)))
}

func colorLerp(c1, c2 rgba, pct float32) rgba {
	return rgba{lerp(c1.r, c2.r, pct), lerp(c1.g, c2.g, pct), lerp(c1.b, c2.b, pct)}
}

func getGradient(c1, c2 rgba) []rgba {
	result := make([]rgba, 256)
	for i := range result {
		pct := float32(i) / float32(255)
		result[i] = colorLerp(c1, c2, pct)
	}
	return result
}

func getDualGradient(c1, c2, c3, c4 rgba) []rgba {
	result := make([]rgba, 256)
	for i := range result {
		pct := float32(i) / float32(255)
		if pct < 0.5 {
			result[i] = colorLerp(c1, c2, pct*float32(2))
		} else {
			result[i] = colorLerp(c3, c4, pct*float32(1.5)-float32(0.5))
		}
	}
	return result
}

func clamp(min, max, v int) int {
	if v < min {
		v = min
	} else if v > max {
		v = max
	}
	return v
}

func rescaleAndDraw(noise []float32, min, max float32, gradient []rgba, w, h int) []byte {
	result := make([]byte, w*h*4)
	scale := 255.0 / (max - min)
	offset := min * scale
	for i := range noise {
		noise[i] = noise[i]*scale - offset
		c := gradient[clamp(0, 255, int(noise[i]))]
		p := i * 4
		result[p] = c.r
		result[p+1] = c.g
		result[p+2] = c.b
	}
	return result
}

func main() {
	sdl.LogSetAllPriority(sdl.LOG_PRIORITY_VERBOSE)
	err := sdl.Init(sdl.INIT_EVERYTHING)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer sdl.Quit()

	window, err := sdl.CreateWindow("Evolving Pictures", 200, 200,
		int32(winWidth), int32(winHeight), sdl.WINDOW_SHOWN)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer window.Destroy()

	renderer, err := sdl.CreateRenderer(window, -1, sdl.RENDERER_ACCELERATED)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer renderer.Destroy()

	// explosionBytes, audioSpec := sdl.LoadWAV("explode.wav")
	// audioID, err := sdl.OpenAudioDevice("", false, audioSpec, nil, 0)
	// if err != nil {
	// 	panic(err)
	// }
	// defer sdl.FreeWAV(explosionBytes)

	// audioState := audioState{explosionBytes, audioID, audioSpec}

	sdl.SetHint(sdl.HINT_RENDER_SCALE_QUALITY, "1")

	var elapsedTime float32
	currentMouseState := GetMouseState()
	
	keyboardState := sdl.GetKeyboardState()

	rand.Seed(time.Now().UnixNano())

	picturesTree := make([]*picture, numPics)
	for i := range picturesTree {
		picturesTree[i] = newPicture()
	}

	picWidth := int(float32(winWidth/columns)*float32(.9))
	picHeight := int(float32(winHeight/rows)*float32(.9))

	buttons:= make([]*ImageButton, numPics)
	textureChan := make(chan pixelsTexture, numPics)

	for i := range buttons {
		go func(i int) {
			texturePix := AptToPixels(picturesTree[i],picWidth,picHeight,renderer)
			textureChan <- pixelsTexture{texturePix, i}
		}(i)
	}

	// p := newPicture()
	// tex := AptToTexture(p, winWidth, winHeight, renderer)

	for {
		frameStart := time.Now()

		for event := sdl.PollEvent(); event != nil; event = sdl.PollEvent() {
			switch e := event.(type) {
			case *sdl.QuitEvent:
				return
			case *sdl.TouchFingerEvent:
				if e.Type == sdl.FINGERDOWN {
					touchX := int(e.X * float32(winWidth))
					touchY := int(e.Y * float32(winHeight))
					currentMouseState.X = touchX
					currentMouseState.Y = touchY
					currentMouseState.LeftButton = true
				}
			}
		}

		currentMouseState.Update()

		if keyboardState[sdl.SCANCODE_ESCAPE] != 0 {
			return
		}

		select {
		case texAndIdx, ok := <- textureChan:
			if ok {
				tex := pixelsToTexture(renderer, texAndIdx.pixels, picWidth, picHeight)
				xi := texAndIdx.num % columns
				yi := (texAndIdx.num-xi)/rows
				x := int32(xi*picWidth)
				y := int32(yi*picHeight)
				xPad := int32(float32(winWidth)*.1/float32(columns+1))
				yPad := int32(float32(winHeight)*.1/float32(rows+1))
				x += xPad*(int32(xi)+1)
				y += yPad*(int32(yi)+1)
				rect := sdl.Rect{x,y,int32(picWidth),int32(picHeight)}
				button := NewImageButton(renderer, tex, rect, sdl.Color{255,255,255,0})
				buttons[texAndIdx.num] = button
			}
		default:

		}

		renderer.Clear()
		for _, button := range buttons {
			if button != nil {
				button.Update(currentMouseState)
				if button.WasLeftClicked {
					button.IsSelected = !button.IsSelected
				}
				button.Draw(renderer)
			}
		}
		renderer.Present()
		elapsedTime = float32(time.Since(frameStart).Seconds() * 1000)
		//	fmt.Println("ms per frame:", elapsedTime)
		if elapsedTime < 5 {
			sdl.Delay(5 - uint32(elapsedTime))
			elapsedTime = float32(time.Since(frameStart).Seconds() * 1000)
		}
	}

}