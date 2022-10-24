package main

import (
	"fmt"
	"image"
	"image/color"
	"image/png"
	"log"
	"os"

	"github.com/llgcode/draw2d/draw2dsvg"
)

type drawconfig struct {
	width         int
	height        int
	linesPerPixel int
	pixelSize     float64
	subpixelSize  float64
}

func main() {
	img := Load(os.Args[1])
	dest := draw2dsvg.NewSvg()
	gc := draw2dsvg.NewGraphicContext(dest)

	w := 1000
	imgW := float64(img.Bounds().Dx())
	imgH := float64(img.Bounds().Dy())
	pixelSize := float64(w) / imgW
	linesPerPixel := 5

	config := drawconfig{width: w,
		height:        int(pixelSize * imgH),
		linesPerPixel: linesPerPixel,
		pixelSize:     pixelSize,
		subpixelSize:  pixelSize / float64(linesPerPixel)}

	palatte := GetPalatte(img)
	for i := range palatte {
		c := palatte[i]
		// black bitmask := GenerateBitmask(img, color.NRGBA{0x18, 0x18, 0x18, 0xff})
		bitmask := GenerateBitmask(img, c)

		remainingBitmask := GenerateSubpixelBitmask(bitmask, config)

		// DrawDebug(gc, config, bitmask, remainingBitmask)

		Draw(gc, config, remainingBitmask, c)
		//Outline(gc, config, bitmask)
	}

	draw2dsvg.SaveToSvgFile("main.svg", dest)

	fmt.Println("done.")
}

func GetPalatte(img image.Image) []color.Color {
	w := img.Bounds().Dy()
	h := img.Bounds().Dx()

	set := make(map[color.Color]bool)
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			set[img.At(x, y)] = true
		}
	}

	i := 0
	palatte := make([]color.Color, len(set))

	for k := range set {
		palatte[i] = k
		i++
	}
	return palatte
}

func GenerateBitmask(img image.Image, c color.Color) [][]bool {
	w := img.Bounds().Dx()
	h := img.Bounds().Dy()

	bitmask := make([][]bool, h)

	for y := range bitmask {
		bitmask[y] = make([]bool, w)
		for x := range bitmask[y] {
			bitmask[y][x] = img.At(x, y) == c
		}
	}
	fmt.Println("w/h", w, h)
	fmt.Println("bitmask size x/y:", len(bitmask[0]), len(bitmask))

	return bitmask
}

func GenerateSubpixelBitmask(bitmask [][]bool, c drawconfig) [][]bool {
	remainingBitmask := make([][]bool, c.linesPerPixel*len(bitmask))

	for y := range remainingBitmask {
		remainingBitmask[y] = make([]bool, c.linesPerPixel*len(bitmask[0]))
		for x := range remainingBitmask[y] {
			remainingBitmask[y][x] = bitmask[y/c.linesPerPixel][x/c.linesPerPixel]
		}
	}

	return remainingBitmask
}

func Load(file string) image.Image {
	existingImageFile, err := os.Open(file)
	if err != nil {
		log.Fatal(err)
	}
	defer existingImageFile.Close()

	loadedImage, err := png.Decode(existingImageFile)
	if err != nil {
		log.Fatal(err)
	}

	return loadedImage
}

func CloneBitMask(bitMask [][]bool) [][]bool {
	dup := make([][]bool, len(bitMask))
	for i := range bitMask {
		dup[i] = make([]bool, len(bitMask[0]))
		copy(dup[i], bitMask[i])
	}
	return dup
}

func GetStartPositions(x, y int, bitmask [][]bool, progress [][]bool) [][]int {
	// look for every connected pixel that has 5 false neighbours

	startPositions := [][]int{}
	unfilledNeighbourCount := 0

	if !progress[y][x] {
		return startPositions
	}
	progress[y][x] = false

	for i := -1; i <= 1; i++ {
		for j := -1; j <= 1; j++ {
			pX := x + j
			pY := y + i

			if pX < 0 || pX >= len(bitmask[0]) || pY < 0 || pY >= len(bitmask) || !bitmask[pY][pX] {
				unfilledNeighbourCount += 1
				continue
			}

			startPositions = append(startPositions, GetStartPositions(pX, pY, bitmask, progress)...)
		}
	}

	// todo: do I need this? Why not just try every connected pixel?
	if unfilledNeighbourCount >= 5 {
		startPositions = append(startPositions, []int{x, y})
	}

	return startPositions
}

func Draw(gc *draw2dsvg.GraphicContext, c drawconfig, remainingBitmask [][]bool, strokeColor color.Color) {
	gc.SetStrokeColor(strokeColor)
	subpixelOffset := c.subpixelSize / 2
	pathCount := 0

	fmt.Println("draw bitmask size x/y:", len(remainingBitmask[0]), len(remainingBitmask))
	for y := range remainingBitmask {
		for x := range remainingBitmask[y] {
			if remainingBitmask[y][x] {
				path := FindLongestPath(gc, c, x, y, remainingBitmask)

				gc.BeginPath()
				for i, pos := range path {
					x := (float64(pos[0]) * c.subpixelSize) + subpixelOffset
					y := (float64(pos[1]) * c.subpixelSize) + subpixelOffset

					remainingBitmask[pos[1]][pos[0]] = false

					if i == 0 {
						gc.MoveTo(x, y)
					} else {
						gc.SetStrokeColor(strokeColor)
						gc.LineTo(x, y)
					}
				}
				gc.Stroke()
				pathCount += 1
			}
		}
	}

	fmt.Println("path count:", pathCount)
}

func FindLongestPath(gc *draw2dsvg.GraphicContext, c drawconfig, xSubpixel, ySubpixel int, remainingBitmask [][]bool) [][]int {
	movementOrders := [][]string{
		{"left", "up", "down", "right"}, // top left vertical
		{"right", "up", "down", "left"}, // top right
		{"left", "down", "up", "right"}, // bottom left
		{"right", "down", "up", "left"}, // bottom right

		{"up", "left", "right", "down"}, // top left horrizontal
		{"up", "right", "left", "down"}, // top right
		{"down", "left", "right", "up"}, // bottom left
		{"down", "right", "left", "up"}, // bottom right
	}
	startPositions := GetStartPositions(xSubpixel, ySubpixel, remainingBitmask, CloneBitMask(remainingBitmask))

	paths := make([][][]int, 1)

	// TODO: try using different movements (maybe just hor / vert) at start of each pixel (recursive?)
	// try looping all the movement patterns for each interesting start positions
	// also select best option by least turns / diags
	// need to build a move tree and test all movement patterns at each pixel change and then take the longest set of nodes
	for _, startPosition := range startPositions {
		x := startPosition[0]
		y := startPosition[1]

		for _, movementOrder := range movementOrders {
			prospectiveBitmask := CloneBitMask(remainingBitmask)
			path := [][]int{{x, y}}
			prospectiveBitmask[y][x] = false

			for {
				nextPosition := NextLinePosition(prospectiveBitmask, movementOrder, x, y)
				if nextPosition == nil {
					break
				}

				path = append(path, nextPosition)
				x = nextPosition[0]
				y = nextPosition[1]
				prospectiveBitmask[y][x] = false
			}

			paths = append(paths, path)
		}
	}

	return LongestPath(paths)
}

func LongestPath(paths [][][]int) [][]int {
	longestLen := 0
	longestIndex := 0

	for i, path := range paths {
		if longestLen < len(path) {
			longestLen = len(path)
			longestIndex = i
		}
	}

	return paths[longestIndex]
}

func NextLinePosition(bitmask [][]bool, movementOrder []string, x, y int) []int {
	for _, movement := range movementOrder {
		switch movement {
		case "left":
			if IsValidPosition(bitmask, x-1, y) {
				return []int{x - 1, y}
			}
		case "up":
			if IsValidPosition(bitmask, x, y-1) {
				return []int{x, y - 1}
			}
		case "down":
			if IsValidPosition(bitmask, x, y+1) {
				return []int{x, y + 1}
			}
		case "right":
			if IsValidPosition(bitmask, x+1, y) {
				return []int{x + 1, y}
			}
		}
	}

	// diagonal order doesn't matter / change since there is at most one option
	// TODO: prevent disgional moves into white space
	if IsValidPosition(bitmask, x+1, y+1) {
		return []int{x + 1, y + 1}
	}
	if IsValidPosition(bitmask, x-1, y+1) {
		return []int{x - 1, y + 1}
	}
	if IsValidPosition(bitmask, x-1, y-1) {
		return []int{x - 1, y - 1}
	}
	if IsValidPosition(bitmask, x+1, y-1) {
		return []int{x + 1, y - 1}
	}

	return nil
}

func IsValidPosition(bitmask [][]bool, x, y int) bool {
	if y < 0 || y >= len(bitmask) {
		return false
	}
	if x < 0 || x >= len(bitmask[y]) {
		return false
	}

	return bitmask[y][x]
}

func DrawDebug(gc *draw2dsvg.GraphicContext, c drawconfig, bitmask, remainingBitmask [][]bool) {
	gc.SetFillColor(color.RGBA{0xff, 0xD0, 0xD0, 0x90})
	for y := range bitmask {
		for x := range bitmask[y] {
			if !bitmask[y][x] {
				continue
			}
			gc.BeginPath()
			gc.MoveTo(float64(x)*c.pixelSize, float64(y)*c.pixelSize)
			gc.LineTo(float64(x+1)*c.pixelSize, float64(y)*c.pixelSize)
			gc.LineTo(float64(x+1)*c.pixelSize, float64(y+1)*c.pixelSize)
			gc.LineTo(float64(x)*c.pixelSize, float64(y+1)*c.pixelSize)
			gc.LineTo(float64(x)*c.pixelSize, float64(y)*c.pixelSize)
			gc.Fill()
		}
	}

	//gc.SetStrokeColor(color.RGBA{0x00, 0x00, 0x00, 0x55})
	//for y := range remainingBitmask {
	//for x := range remainingBitmask[y] {
	//if !remainingBitmask[y][x] { continue }
	//gc.BeginPath()
	//gc.MoveTo(float64(x) * c.subpixelSize, float64(y) * c.subpixelSize)
	//gc.LineTo(float64(x+1) * c.subpixelSize, float64(y) * c.subpixelSize)
	//gc.LineTo(float64(x+1) * c.subpixelSize, float64(y+1) * c.subpixelSize)
	//gc.LineTo(float64(x) * c.subpixelSize, float64(y+1) * c.subpixelSize)
	//gc.LineTo(float64(x) * c.subpixelSize, float64(y) * c.subpixelSize)
	//gc.Stroke()
	//gc.Fill()
	//}
	//}

	//gc.SetStrokeColor(color.RGBA{0xff, 0xD0, 0xD0, 0xff})
	//for x := 0.0  ; x <= c.width / c.pixelSize ; x++ {
	//gc.BeginPath()
	//gc.MoveTo(float64(x) * c.pixelSize, 0)
	//gc.LineTo(float64(x) * c.pixelSize, float64(c.height))
	//gc.Stroke()
	//}

	//for y := 0.0 ; y <= c.height / c.pixelSize ; y++ {
	//gc.BeginPath()
	//gc.MoveTo(0, float64(y) * c.pixelSize)
	//gc.LineTo(float64(c.width), float64(y) * c.pixelSize)
	//gc.Stroke()
	//}
}
