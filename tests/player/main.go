package main

import (
	"bytes"
	"image"
	_ "image/png"
	"log"

	"github.com/hajimehoshi/ebiten/examples/resources/images"
	"github.com/hajimehoshi/ebiten/v2"
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
)

var (
	runnerImage *ebiten.Image
)

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
	const groundY = screenHeight - frameHeight/2

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

	subImg := runnerImage.SubImage(image.Rect(sx, sy, sx+frameWidth, sy+frameHeight)).(*ebiten.Image)

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
	g.drawPlayer(screen, &g.Player)
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return screenWidth, screenHeight
}

func main() {

	runnerImage = loadImage(images.Runner_png)

	game := &Game{
		Player: Player{
			X:     screenWidth / 2,
			Y:     screenHeight - frameHeight/2,
			Flip:  false,
			State: "idle",
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
