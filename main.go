package main

import (
	"fmt"
	"io/ioutil"
	"math/rand"
	"os"
	"strconv"
	"strings"
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

type guiState struct {
	zoom bool
	zoomTree *picture
	zoomPicture *sdl.Texture
}

type picture struct {
	r, g, b Node
}

func (p *picture) String() string {
	return "( picture\n" + p.r.String() + "\n" + p.g.String() + "\n" + p.b.String() + ")"
}

func saveTree(p *picture) {
	files, err := ioutil.ReadDir("./")
	if err != nil {
		panic(err)
	}

	maksNumber := 0
	for _, file := range files {
		fileName := file.Name()
		if strings.HasSuffix(fileName, ".apt") {
			newString := strings.TrimSuffix(fileName, ".apt")
			num, err := strconv.Atoi(newString)
			if err == nil {
				if maksNumber <= num {
					maksNumber = num+1
				}
			}
		}
	}

	savedFile := strconv.Itoa(maksNumber) + ".apt"
	file, err := os.Create(savedFile)
	if err != nil {
		panic(err)
	}
	defer file.Close()
	fmt.Fprintf(file, p.String())
}

func (p *picture) mutate() {
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

func (p *picture) pickRandomColor() Node {
	r := rand.Intn(3)
	switch r {
	case 0:
		return p.r
	case 1:
		return p.g
	case 2:
		return p.b
	default:
		panic("random out of the bounds")
	}
}

func cross(a, b *picture) *picture {
	aCopy := &picture{CopyTree(a.r, nil), CopyTree(a.g, nil), CopyTree(a.b,nil)}
	aColor := aCopy.pickRandomColor()
	bColor := b.pickRandomColor()

	aIndex := rand.Intn(aColor.CountNode())
	aNode, _ := GetNthChildren(aColor, aIndex, 0)

	bIndex := rand.Intn(bColor.CountNode())
	bNode, _ := GetNthChildren(bColor, bIndex, 0)
	bNodeCopy := CopyTree(bNode, bNode.GetParent())

	ReplaceNode(aNode, bNodeCopy)
	return aCopy
}

func evolve(survivor []*picture) []*picture {
	newPics := make([]*picture, numPics)
	i := 0
	for i < len(survivor) {
		a := survivor[i]
		b := survivor[rand.Intn(len(survivor))]
		newPics[i] = cross(a,b)
		i++
	}

	for i < len(newPics) {
		a := survivor[rand.Intn(len(survivor))]
		b := survivor[rand.Intn(len(survivor))]
		newPics[i] = cross(a,b)
		i++
	}

	return newPics
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

func aptToPixels(p *picture, w, h int) []byte {
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

func pixelsToTextureChan(renderer *sdl.Renderer, pixels []byte, w, h int, textureChan chan *sdl.Texture) {
	tex, err := renderer.CreateTexture(sdl.PIXELFORMAT_ABGR8888, sdl.TEXTUREACCESS_STREAMING, int32(w), int32(h))
	if err != nil {
		panic(err)
	}
	tex.Update(nil, pixels, w*4)
	textureChan <- tex
}

func aptToPixelsChan(p *picture, w, h int, pixelChan chan []byte) {
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
	pixelChan <- pixels
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

	window, err := sdl.CreateWindow("Evolving Pictures", 100, 100,
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
	prevKeyboardState := make([]uint8, len(keyboardState))

	for i, v := range keyboardState {
		prevKeyboardState[i] = v
	}

	rand.Seed(time.Now().UnixNano())

	picturesTree := make([]*picture, numPics)
	for i := range picturesTree {
		picturesTree[i] = newPicture()
	}

	picWidth := int(float32(winWidth/columns)*float32(.9))
	picHeight := int(float32(winHeight/rows)*float32(.8))

	buttons:= make([]*ImageButton, numPics)
	textureChan := make(chan pixelsTexture, numPics)

	evolveButtonTex := GetSinglePicText(renderer, sdl.Color{255,255,255,0})
	evolveButtonRect := sdl.Rect{int32(float32(winWidth/2)-float32(picWidth/2)), int32(float32(winHeight)-float32(winHeight)*.1),int32(picWidth),int32(float32(winHeight)*.08)}
	evolveButton := NewImageButton(renderer, evolveButtonTex, evolveButtonRect, sdl.Color{255,255,255,0})

	zoomState := guiState{false, nil, nil}

	args := os.Args
	if len(args) > 1 {
		fileBytes, err := ioutil.ReadFile(args[1])
		if err != nil {
			panic(err)
		}
		fileStr := string(fileBytes)
		pictureNode := BeginLexing(fileStr)
		p := &picture{pictureNode.GetChildren()[0], pictureNode.GetChildren()[1], pictureNode.GetChildren()[2]}
		pixels := aptToPixels(p, winWidth*2, winHeight*2)
		tex := pixelsToTexture(renderer, pixels, winWidth*2, winHeight*2)
		zoomState.zoom = true
		zoomState.zoomPicture = tex
		zoomState.zoomTree = p
	}


	for i := range buttons {
		go func(i int) {
			texturePix := aptToPixels(picturesTree[i],picWidth*2,picHeight*2)
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

		if !zoomState.zoom {
			select {
			case texAndIdx, ok := <- textureChan:
				if ok {
					tex := pixelsToTexture(renderer, texAndIdx.pixels, picWidth*2, picHeight*2)
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
			for i, button := range buttons {
				if button != nil {
					button.Update(currentMouseState)
					if button.WasLeftClicked {
						button.IsSelected = !button.IsSelected
					} else if button.WasRightClicked {
						timeStart := time.Now()
						texChan := make(chan *sdl.Texture, 10)
						pixelChan := make(chan []byte, 10)
						go aptToPixelsChan(picturesTree[i], winWidth*2, winHeight*2, pixelChan)
						go pixelsToTextureChan(renderer, <-pixelChan, winWidth*2, winHeight*2, texChan)
						zoomState.zoom = true
						zoomState.zoomTree = picturesTree[i]
						zoomState.zoomPicture = <-texChan
						fmt.Println(time.Since(timeStart).Seconds())
					}
					button.Draw(renderer)
				}
			}

			evolveButton.Update(currentMouseState)
			if evolveButton.WasLeftClicked {
				selectedPicture := make([]*picture,0)
				for i, button := range buttons {
					if button.IsSelected {
						selectedPicture = append(selectedPicture, picturesTree[i])
					}
				}
				if len(selectedPicture) != 0 {
					for i := range buttons {
						buttons[i] = nil
					}
					picturesTree = evolve(selectedPicture)
					for i := range picturesTree {
						go func(j int) {
							pixels := aptToPixels(picturesTree[j], picWidth*2, picHeight*2)
							textureChan <- pixelsTexture{pixels, j}
						}(i)
					}	
				}
			}
			evolveButton.Draw(renderer)
		} else {
			if !currentMouseState.RightButton && currentMouseState.PrevRightButton {
				zoomState.zoom = false
			}
			if keyboardState[sdl.SCANCODE_S] == 0 && prevKeyboardState[sdl.SCANCODE_S] != 0 {
				saveTree(zoomState.zoomTree)
			}
			renderer.Copy(zoomState.zoomPicture, nil,nil)
		}
		renderer.Present()
		for i, v := range keyboardState {
			prevKeyboardState[i] = v
		}
		elapsedTime = float32(time.Since(frameStart).Seconds() * 1000)
		//	fmt.Println("ms per frame:", elapsedTime)
		if elapsedTime < 5 {
			sdl.Delay(5 - uint32(elapsedTime))
			elapsedTime = float32(time.Since(frameStart).Seconds() * 1000)
		}
	}

}