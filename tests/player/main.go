package main

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	_ "image/png"
	"log"

	"github.com/hajimehoshi/ebiten/examples/resources/images"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
)

const (
	screenWidth  = 320
	screenHeight = 240

	frameOX     = 0
	frameOY     = 0
	frameWidth  = 32
	frameHeight = 32
	frameCount  = 5

	gravity     = 800.0
	jumpImpulse = -250.0
	tileSize    = 16
)

var (
	playerImage            *ebiten.Image
	tilesImage             *ebiten.Image
	mushroomEnemyImageIdle *ebiten.Image
	mushroomEnemyImageRun  *ebiten.Image
)

func init() {
	playerImage = loadImageBytes(images.Runner_png)
	tilesImage = loadImageBytes(images.Tiles_png)
	mushroomEnemyImageIdle = loadImage("assets/Enemies/Mushroom/Idle (32x32).png")
	mushroomEnemyImageRun = loadImage("assets/Enemies/Mushroom/Run (32x32).png")
}

type Sprite struct {
	Image *ebiten.Image
	X     float64
	Y     float64
}

type Player struct {
	*Sprite
	Vy    float64
	Flip  bool
	State string // "idle", "running", "jumping"
}

type Enemy struct {
	*Sprite
	State         string // "idle", "running", "jumping"
	Flip          bool
	FollowsPlayer bool
}

type Game struct {
	Player  *Player
	Enemys  []*Enemy
	counter int
	layers  [][]int
}

func (g *Game) Update() error {
	g.counter++
	g.controlPlayer()
	g.checkCollision()
	g.controlEnemy()
	return nil
}

func (g *Game) controlEnemy() {
	for _, e := range g.Enemys {
		if e.X < g.Player.X {
			e.X += 1
			e.Flip = true
			e.State = "running"
		} else if e.X > g.Player.X {
			e.X -= 1
			e.Flip = false
			e.State = "running"
		} else {
			e.State = "idle"
		}
	}
}

func (g *Game) controlPlayer() {

	movingHorizontal := false

	if ebiten.IsKeyPressed(ebiten.KeyArrowRight) {
		g.Player.X += 2
		g.Player.Flip = false
		movingHorizontal = true
	}

	if ebiten.IsKeyPressed(ebiten.KeyArrowLeft) {
		g.Player.X -= 2
		g.Player.Flip = true
		movingHorizontal = true
	}

	const dt = 1.0 / 60.0
	const groundY = screenHeight - (frameHeight / 2) - tileSize

	if ebiten.IsKeyPressed(ebiten.KeySpace) && g.Player.Y >= groundY {
		g.Player.Vy = jumpImpulse
		g.Player.State = "jumping"
	}

	g.Player.Vy += gravity * dt
	g.Player.Y += g.Player.Vy * dt

	if g.Player.Y > groundY {
		g.Player.Y = groundY
		g.Player.Vy = 0
		if movingHorizontal {
			g.Player.State = "running"
		} else {
			g.Player.State = "idle"
		}
	} else {
		g.Player.State = "jumping"
	}
}

func (g *Game) checkCollision() {
	if g.Player.X < frameWidth/4 {
		g.Player.X = frameWidth / 4
	}
	if g.Player.X > screenWidth-frameWidth/4 {
		g.Player.X = screenWidth - frameWidth/4
	}
}

func (g *Game) drawPlayer(screen *ebiten.Image, p *Player) {
	var sx, sy int
	switch p.State {
	case "idle":
		sx, sy = g.animateRestingPlayer()
	case "running":
		sx, sy = g.animateRunningPlayer()
	case "jumping":
		sx, sy = g.animateJumpingPlayer()
	default:
		sx, sy = g.animateRestingPlayer()
	}

	subImg := playerImage.SubImage(image.Rect(sx, sy, sx+frameWidth, sy+frameHeight)).(*ebiten.Image)

	op := &ebiten.DrawImageOptions{}

	cx := float64(frameWidth) / 2
	cy := float64(frameHeight) / 2

	op.GeoM.Translate(-cx, -cy)

	if p.Flip {
		op.GeoM.Scale(-1, 1)
		op.GeoM.Translate(0, 0)
	}

	op.GeoM.Translate(p.X, p.Y)

	screen.DrawImage(subImg, op)
}

func (g *Game) drawEnemy(screen *ebiten.Image, e *Enemy) {
	var sx, sy int
	switch e.State {
	case "idle":
		e.Sprite.Image = mushroomEnemyImageIdle
		sx, sy = g.animateMushroomEnemyIdle()
	case "running":
		e.Sprite.Image = mushroomEnemyImageRun
		sx, sy = g.animateMushroomEnemyRun()
	default:
		e.Sprite.Image = mushroomEnemyImageIdle
		sx, sy = g.animateMushroomEnemyIdle()
	}

	subImg := e.Image.SubImage(image.Rect(sx, sy, sx+frameWidth, sy+frameHeight)).(*ebiten.Image)

	op := &ebiten.DrawImageOptions{}

	cx := float64(frameWidth) / 2
	cy := float64(frameHeight) / 2

	op.GeoM.Translate(-cx, -cy)

	if e.Flip {
		op.GeoM.Scale(-1, 1)
		op.GeoM.Translate(0, 0)
	}

	op.GeoM.Translate(e.X, e.Y)

	screen.DrawImage(subImg, op)
}

func (g *Game) animateMushroomEnemyIdle() (int, int) {
	i := (g.counter / 3) % 14
	return frameOX + i*32, 0
}

func (g *Game) animateMushroomEnemyRun() (int, int) {
	i := (g.counter / 5) % 16
	return frameOX + i*32, 0
}

func (g *Game) renderGround(screen *ebiten.Image) {
	w := tilesImage.Bounds().Dx()
	tileXCount := w / tileSize
	const xCount = screenWidth / tileSize
	for _, l := range g.layers {
		for i, t := range l {
			op := &ebiten.DrawImageOptions{}
			op.GeoM.Translate(float64((i%xCount)*tileSize), float64((i/xCount)*tileSize))
			sx := (t % tileXCount) * tileSize
			sy := (t / tileXCount) * tileSize
			screen.DrawImage(tilesImage.SubImage(image.Rect(sx, sy, sx+tileSize, sy+tileSize)).(*ebiten.Image), op)
		}
	}
}

func (g *Game) animateRestingPlayer() (int, int) {
	i := (g.counter / 5) % 5
	return frameOX + i*frameWidth, 0
}

func (g *Game) animateRunningPlayer() (int, int) {
	i := (g.counter / 5) % 8
	return frameOX + i*frameWidth, frameHeight
}

func (g *Game) animateJumpingPlayer() (int, int) {
	i := (g.counter / 5) % 4
	return frameOX + i*frameWidth, frameHeight * 2
}

func (g *Game) Draw(screen *ebiten.Image) {
	skyColor := color.RGBA{R: 30, G: 30, B: 80, A: 255}
	screen.Fill(skyColor)
	g.renderGround(screen)
	g.drawPlayer(screen, g.Player)
	for _, e := range g.Enemys {
		g.drawEnemy(screen, e)
	}
	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("TPS: %0.2f", ebiten.CurrentTPS()), 10, 0)
	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("FPS: %0.2f", ebiten.CurrentFPS()), screenWidth-70, 0)
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return screenWidth, screenHeight
}

func main() {

	game := &Game{
		Player: &Player{
			Sprite: &Sprite{
				Image: playerImage,
				X:     screenWidth / 2,
				Y:     screenHeight - (frameHeight / 2) - tileSize,
			},
			Flip:  false,
			Vy:    0,
			State: "idle",
		},
		Enemys: []*Enemy{
			{
				Sprite: &Sprite{
					Image: mushroomEnemyImageIdle,
					X:     screenWidth / 2,
					Y:     screenHeight - (frameHeight / 2) - tileSize,
				},
				State: "idle",
			},
		},
		layers: [][]int{
			generateGroundLayer(),
		},
	}

	ebiten.SetWindowSize(screenWidth*2, screenHeight*2)
	ebiten.SetWindowTitle("Player")
	if err := ebiten.RunGame(game); err != nil {
		log.Fatal(err)
	}
}

func loadImageBytes(imageBytes []byte) *ebiten.Image {
	img, _, err := image.Decode(bytes.NewReader(imageBytes))
	if err != nil {
		log.Fatal(err)
	}
	return ebiten.NewImageFromImage(img)
}

func loadImage(path string) *ebiten.Image {
	img, _, err := ebitenutil.NewImageFromFile(path)
	if err != nil {
		log.Fatal(err)
	}
	return img
}

func generateGroundLayer() []int {
	tileIndex := 247
	xCount := screenWidth / tileSize
	yCount := screenHeight / tileSize
	layer := make([]int, xCount*yCount)
	for i := (yCount - 2) * xCount; i < len(layer); i++ {
		layer[i] = tileIndex
	}
	return layer
}
