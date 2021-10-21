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

type EdgeBehaviour int

type GameWorld struct {
	name              string
	screenWidth       int
	screenHeight      int
	maxAngle          int
	minSprites        int
	maxSprites        int
	angleOfRotation   int
	bigSpriteGrowth   int
	smallSpriteGrowth int
	spriteCount       int
	showInfo          bool
	speedFactor       int
	rebound           EdgeBehaviour
}

var world GameWorld

const (
	Rebound EdgeBehaviour = iota
	PassThrough
)

var spriteImage *ebiten.Image

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

func spriteSpeed(factor int, fixed bool) (int, int) {

	if fixed {
		return factor, factor
	}

	x := 2*rand.Intn(2) - 1
	y := 2*rand.Intn(2) - 1

	xSpeedVariation := 2*rand.Intn(factor) + 1
	ySpeedVariation := 2*rand.Intn(factor) + 1

	x *= xSpeedVariation
	y *= ySpeedVariation

	return x, y
}

func init() {
	img, _, err := ebitenutil.NewImageFromFile("amber-orb.png")

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
		} else if mx := world.screenWidth - s.spriteWidth; mx <= s.x {
			s.x = world.screenWidth - s.spriteWidth
			s.vx = -s.vx
		}
		if s.y < 0 {
			s.y = -s.y
			s.vy = -s.vy
		} else if my := world.screenHeight - s.spriteHeight; my <= s.y {
			s.y = world.screenHeight - s.spriteHeight
			s.vy = -s.vy
		}
	} else if behave == PassThrough {
		if s.x < -s.spriteWidth {
			s.x = world.screenWidth
		} else if s.x > world.screenWidth+s.spriteWidth {
			s.x = 0
		}
		if s.y < -s.spriteHeight {
			s.y = world.screenHeight
		} else if s.y > world.screenHeight+s.spriteHeight {
			s.y = 0
		}
	}
}

func (s *Sprite) Update() {

	// Update the sprites position and angle.
	s.x += s.vx
	s.y += s.vy

	s.DealWithScreenEdges(world.rebound)

	s.angle += world.angleOfRotation * s.rotation

	if s.angle == world.maxAngle {
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

	game.sprites.sprites = make([]*Sprite, world.maxSprites)
	game.sprites.num = world.spriteCount
	for i := range game.sprites.sprites {
		w, h := spriteImage.Size()
		x, y := rand.Intn(world.screenWidth-w), rand.Intn(world.screenHeight-h)

		r := utils.PlusOrMinus()
		vx, vy := spriteSpeed(world.speedFactor, false)
		a := rand.Intn(world.maxAngle)

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
		world.speedFactor += 1
	}

	if ebiten.IsKeyPressed(ebiten.KeyA) {
		world.speedFactor -= 1
	}

	if ebiten.IsKeyPressed(ebiten.KeyArrowUp) {
		game.sprites.num += world.smallSpriteGrowth
		if world.maxSprites < game.sprites.num {
			game.sprites.num = world.maxSprites
		}
	}

	if ebiten.IsKeyPressed(ebiten.KeyArrowRight) {
		game.sprites.num += world.bigSpriteGrowth
		if world.maxSprites < game.sprites.num {
			game.sprites.num = world.maxSprites
		}
	}

	if ebiten.IsKeyPressed(ebiten.KeyArrowLeft) {
		game.sprites.num -= world.bigSpriteGrowth
		if game.sprites.num < world.minSprites {
			game.sprites.num = world.minSprites
		}
	}

	if ebiten.IsKeyPressed(ebiten.KeyArrowDown) {
		game.sprites.num -= world.smallSpriteGrowth
		if game.sprites.num < world.minSprites {
			game.sprites.num = world.minSprites
		}
	}

	game.sprites.Update()
	return nil
}

func (game *Game) DisplayInformation(showInfo bool, screen *ebiten.Image) {
	if showInfo {
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
		//game.op.GeoM.Rotate(2 * math.Pi * float64(s.angle) / maxAngle)
		//game.op.GeoM.Translate(float64(w)/2, float64(h)/2)
		game.op.GeoM.Translate(float64(s.x), float64(s.y))
		screen.DrawImage(spriteImage, &game.op)
	}

	game.DisplayInformation(world.showInfo, screen)
}

func (game *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return world.screenWidth, world.screenHeight
}

func main() {

	println("Starting the app")

	world = GameWorld{
		name:              "world",
		screenWidth:       1600,
		screenHeight:      1000,
		maxAngle:          360,
		minSprites:        1,
		maxSprites:        50000,
		angleOfRotation:   16,
		bigSpriteGrowth:   50,
		smallSpriteGrowth: 1,
		spriteCount:       3,
		showInfo:          true,
		speedFactor:       10,
		rebound:           PassThrough,
	}

	ebiten.SetWindowSize(world.screenWidth, world.screenHeight)
	ebiten.SetWindowTitle("Sprite handling")
	ebiten.SetWindowResizable(true)
	if err := ebiten.RunGame(&Game{}); err != nil {
		log.Fatal(err)
	}
}
