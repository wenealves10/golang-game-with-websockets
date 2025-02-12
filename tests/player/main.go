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
	playerImage *ebiten.Image
	tilesImage  *ebiten.Image
)

func init() {
	playerImage = loadImage(images.Runner_png)
	tilesImage = loadImage(images.Tiles_png)
}

type Player struct {
	X     float64
	Y     float64
	Vy    float64
	Flip  bool
	State string // "idle", "running", "jumping"
}

type Game struct {
	Player  Player
	counter int
	layers  [][]int
}

func (g *Game) Update() error {
	g.counter++
	g.controlPlayer()
	g.checkCollision()
	return nil
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
	g.drawPlayer(screen, &g.Player)
	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("TPS: %0.2f", ebiten.CurrentTPS()), 0, 0)
	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("FPS: %0.2f", ebiten.CurrentFPS()), 0, 20)
	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("X: %0.2f", g.Player.X), 0, 40)
	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("Y: %0.2f", g.Player.Y), 0, 60)
	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("Vy: %0.2f", g.Player.Vy), 0, 80)
	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("State: %s", g.Player.State), 0, 100)
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return screenWidth, screenHeight
}

func main() {

	game := &Game{
		Player: Player{
			X:     screenWidth / 2,
			Y:     screenHeight - (frameHeight / 2) - tileSize,
			Flip:  false,
			State: "idle",
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

func loadImage(imageBytes []byte) *ebiten.Image {
	img, _, err := image.Decode(bytes.NewReader(imageBytes))
	if err != nil {
		log.Fatal(err)
	}
	return ebiten.NewImageFromImage(img)
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
