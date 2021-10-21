package main

import (
	"fmt"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"go1/utils"
	_ "image/png"
	"log"
	"math/rand"
)

const (
	ScreenWidth       = 1600
	ScreenHeight      = 1000
	MaxAngle          = 360
	MinSprites        = 1
	MaxSprites        = 50000
	AngleOfRotation   = 16
	BigSpriteGrowth   = 50
	SmallSpriteGrowth = 1
	SpriteCount       = 1
	ShowInfo          = true
)

type EdgeBehaviour int

const (
	Rebound EdgeBehaviour = iota
	PassThrough
)

var spriteImage *ebiten.Image
var SpeedFactor = 1

type Sprite struct {
	spriteWidth  int
	spriteHeight int
	x            int
	y            int
	vx           int
	vy           int
	angle        int
	rotation     int
	xRand        int
	yRand        int
}

type Sprites struct {
	sprites []*Sprite
	num     int
}

type Game struct {
	touchIDs    []ebiten.TouchID
	sprites     Sprites
	op          ebiten.DrawImageOptions
	initialised bool
}

func spriteSpeed(factor int) (int, int){
	x := 2*rand.Intn(2)-1
	y := 2*rand.Intn(2)-1

	xSpeedVariation := 2 * rand.Intn(factor) + 1
	ySpeedVariation := 2 * rand.Intn(factor) + 1

	x *= xSpeedVariation
	y *= ySpeedVariation

	return x, y
}

func init() {
	img, _, err := ebitenutil.NewImageFromFile("cell.png")

	if err != nil {
		log.Fatal(err)
	}

	w, h := img.Size()
	spriteImage = ebiten.NewImage(w, h)

	drawOptions := &ebiten.DrawImageOptions{}
	spriteImage.DrawImage(img, drawOptions)
}

func (s *Sprite) DealWithScreenEdges(behave EdgeBehaviour) {

	if behave == Rebound {
		if s.x < 0 {
			s.x = -s.x
			s.vx = -s.vx
		} else if mx := ScreenWidth - s.spriteWidth; mx <= s.x {
			s.x = ScreenWidth - s.spriteWidth
			s.vx = -s.vx
		}
		if s.y < 0 {
			s.y = -s.y
			s.vy = -s.vy
		} else if my := ScreenHeight - s.spriteHeight; my <= s.y {
			s.y = ScreenHeight - s.spriteHeight
			s.vy = -s.vy
		}
	} else if behave == PassThrough {
		if s.x < -s.spriteWidth {
			s.x = ScreenWidth
		} else if s.x > ScreenWidth + s.spriteWidth {
			s.x = 0
		}
		if s.y < -s.spriteHeight {
			s.y = ScreenHeight
		} else if s.y > ScreenHeight + s.spriteHeight {
			s.y = 0
		}
	}
}

func (s *Sprite) Update() {

	// Update the sprites position and angle.
	s.x += s.vx
	s.y += s.vy

	s.DealWithScreenEdges(Rebound)

	s.angle += AngleOfRotation * s.rotation

	if s.angle == MaxAngle {
		s.angle = 0
	}
}

func (sprite *Sprites) Update() {
	for i := 0; i < sprite.num; i++ {
		sprite.sprites[i].Update()
	}
}

func (game *Game) init() {
	defer func() {
		game.initialised = true
	}()

	game.sprites.sprites = make([]*Sprite, MaxSprites)
	game.sprites.num = SpriteCount
	for i := range game.sprites.sprites {
		w, h := spriteImage.Size()
		x, y := rand.Intn(ScreenWidth-w), rand.Intn(ScreenHeight-h)

		r := utils.PlusOrMinus()
		vx, vy := spriteSpeed(SpeedFactor)
		a := rand.Intn(MaxAngle)

		game.sprites.sprites[i] = &Sprite{
			spriteWidth:  w,
			spriteHeight: h,
			x:            x,
			y:            y,
			vx:           vx,
			vy:           vy,
			angle:        a,
			rotation:     r,
		}
	}
}

func (game *Game) Update() error {
	if !game.initialised {
		game.init()
	}

	if ebiten.IsKeyPressed(ebiten.KeyQ) {
		SpeedFactor += 1
	}

	if ebiten.IsKeyPressed(ebiten.KeyA) {
		SpeedFactor -= 1
	}

	if ebiten.IsKeyPressed(ebiten.KeyArrowUp) {
		game.sprites.num += SmallSpriteGrowth
		if MaxSprites < game.sprites.num {
			game.sprites.num = MaxSprites
		}
	}

	if ebiten.IsKeyPressed(ebiten.KeyArrowRight) {
		game.sprites.num += BigSpriteGrowth
		if MaxSprites < game.sprites.num {
			game.sprites.num = MaxSprites
		}
	}

	if ebiten.IsKeyPressed(ebiten.KeyArrowLeft) {
		game.sprites.num -= BigSpriteGrowth
		if game.sprites.num < MinSprites {
			game.sprites.num = MinSprites
		}
	}

	if ebiten.IsKeyPressed(ebiten.KeyArrowDown) {
		game.sprites.num -= SmallSpriteGrowth
		if game.sprites.num < MinSprites {
			game.sprites.num = MinSprites
		}
	}

	game.sprites.Update()
	return nil
}

func (game *Game) DisplayInformation(showInfo bool, screen *ebiten.Image) {
	if ShowInfo {
		msg := fmt.Sprintf("TPS: %0.2f\nFPS: %0.1f\nNum of sprites: %d",
			ebiten.CurrentTPS(), ebiten.CurrentFPS(), game.sprites.num)
		ebitenutil.DebugPrint(screen, msg)
	}
}

func (game *Game) Draw(screen *ebiten.Image) {
	// Draw each sprite.
	// DrawImage can be called many times, but in the implementation,
	// the actual draw call to GPU is very few since these calls satisfy
	// some conditions e.game. all the rendering sources and targets are same.
	// For more detail, see:
	// https://pkg.go.dev/github.com/hajimehoshi/ebiten/v2#Image.DrawImage

	//w, h := spriteImage.Size()
	for i := 0; i < game.sprites.num; i++ {
		s := game.sprites.sprites[i]
		game.op.GeoM.Reset()
		//game.op.GeoM.Translate(-float64(w)/2, -float64(h)/2)
		//game.op.GeoM.Rotate(2 * math.Pi * float64(s.angle) / MaxAngle)
		//game.op.GeoM.Translate(float64(w)/2, float64(h)/2)
		game.op.GeoM.Translate(float64(s.x), float64(s.y))
		screen.DrawImage(spriteImage, &game.op)
	}

	game.DisplayInformation(ShowInfo, screen)
}

func (game *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return ScreenWidth, ScreenHeight
}

func main() {
	ebiten.SetWindowSize(ScreenWidth, ScreenHeight)
	ebiten.SetWindowTitle("Sprite handling")
	ebiten.SetWindowResizable(true)
	if err := ebiten.RunGame(&Game{}); err != nil {
		log.Fatal(err)
	}
}