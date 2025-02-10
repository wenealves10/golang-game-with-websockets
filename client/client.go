package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"image"
	"image/color"
	"log"
	"math"
	"math/rand"
	"net/url"
	"os"
	"time"

	"github.com/gorilla/websocket"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/examples/resources/images"
)

const (
	screenWidth       = 800
	screenHeight      = 600
	groundY           = 500
	playerX           = 100
	playerWidth       = 20
	playerHeight      = 40
	gravity           = 800.0
	jumpImpulse       = -350.0
	periodSun         = 30.0
	playerBulletSpeed = 500.0
	shootCooldownTime = 0.5
)

const (
	frameOX     = 0
	frameOY     = 32
	frameWidth  = 32
	frameHeight = 32
	frameCount  = 8
)

var (
	runnerImage *ebiten.Image
)

type Player struct {
	ID string  `json:"id"`
	Y  float64 `json:"y"`
	Vy float64 `json:"vy"`
}

type Enemy struct {
	X          float64 `json:"x"`
	Y          float64 `json:"y"`
	Vx         float64 `json:"vx"`
	Vy         float64 `json:"vy"`
	ShootTimer float64 `json:"shootTimer"`
	Dead       bool    `json:"dead"`
	DeathTimer float64 `json:"deathTimer"`
	WalkPhase  float64 `json:"walkPhase"`
}

type Bullet struct {
	X    float64 `json:"x"`
	Y    float64 `json:"y"`
	Vx   float64 `json:"vx"`
	Vy   float64 `json:"vy"`
	From string  `json:"from"`
}

type Sun struct {
	X     float64    `json:"x"`
	Y     float64    `json:"y"`
	Color color.RGBA `json:"color"`
}

type GameState struct {
	Sun      Sun                `json:"sun"`
	Players  map[string]*Player `json:"players"`
	Enemies  []*Enemy           `json:"enemies"`
	Bullets  []*Bullet          `json:"bullets"`
	Points   int                `json:"points"`
	Level    int                `json:"level"`
	GameOver bool               `json:"gameOver"`
}

type Message struct {
	Type     string `json:"type"`
	PlayerID string `json:"playerId"`
	Command  string `json:"command"`
}

type Game struct {
	wsConn        *websocket.Conn
	state         GameState
	localPlayerID string

	lastSpace bool
	lastZ     bool

	shootCooldown float64
	time          float64
	count         int
}

func NewGame() *Game {
	rand.Seed(time.Now().UnixNano())
	return &Game{
		state: GameState{
			Players:  make(map[string]*Player),
			Enemies:  []*Enemy{},
			Bullets:  []*Bullet{},
			Points:   0,
			Level:    1,
			GameOver: false,
		},
		shootCooldown: shootCooldownTime,
		localPlayerID: "player1",
		time:          0,
	}
}

func (g *Game) connectWebSocket() {
	u := url.URL{
		Scheme:   "ws",
		Host:     "localhost:3000",
		Path:     "/ws",
		RawQuery: "id=" + g.localPlayerID,
	}
	conn, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		log.Fatal("Erro na conexão WebSocket:", err)
	}
	g.wsConn = conn
	go g.readMessages()
}

func (g *Game) readMessages() {
	for {
		_, msg, err := g.wsConn.ReadMessage()
		if err != nil {
			log.Println("Erro ao ler mensagem do servidor:", err)
			return
		}
		var newState GameState
		if err := json.Unmarshal(msg, &newState); err != nil {
			log.Println("Erro ao decodificar estado:", err)
			continue
		}

		g.state = newState
	}
}

func (g *Game) sendMessage(m Message) {
	if g.wsConn == nil {
		return
	}
	data, err := json.Marshal(m)
	if err != nil {
		log.Println("Erro ao codificar mensagem:", err)
		return
	}
	if err := g.wsConn.WriteMessage(websocket.TextMessage, data); err != nil {
		log.Println("Erro ao enviar mensagem:", err)
	}
}

func (g *Game) updateInput() {
	curSpace := ebiten.IsKeyPressed(ebiten.KeySpace)
	if curSpace && !g.lastSpace {
		m := Message{
			Type:     "command",
			PlayerID: g.localPlayerID,
			Command:  "jump",
		}
		g.sendMessage(m)
	}
	g.lastSpace = curSpace

	curZ := ebiten.IsKeyPressed(ebiten.KeyZ)
	if curZ && !g.lastZ && g.shootCooldown <= 0 {
		m := Message{
			Type:     "command",
			PlayerID: g.localPlayerID,
			Command:  "shoot",
		}
		g.sendMessage(m)
		g.shootCooldown = shootCooldownTime
	}
	g.lastZ = curZ
	if g.shootCooldown > 0 {
		g.shootCooldown -= 1.0 / 60.0
	}

	if g.state.GameOver && ebiten.IsKeyPressed(ebiten.KeyR) {
		m := Message{
			Type:     "command",
			PlayerID: g.localPlayerID,
			Command:  "reset",
		}
		g.sendMessage(m)
	}
}

func drawCircle(screen *ebiten.Image, cx, cy, r float64, clr color.Color) {
	const segments = 40
	theta := 2 * math.Pi / segments
	var prevX, prevY float64
	for i := 0; i <= segments; i++ {
		angle := float64(i) * theta
		x := cx + r*math.Cos(angle)
		y := cy + r*math.Sin(angle)
		if i > 0 {
			ebitenutil.DrawLine(screen, prevX, prevY, x, y, clr)
		}
		prevX, prevY = x, y
	}
}

func drawFilledCircle(screen *ebiten.Image, cx, cy, r float64, clr color.Color) {
	R := int(math.Ceil(r))
	for y := -R; y <= R; y++ {
		dx := math.Sqrt(r*r - float64(y*y))
		x0 := cx - dx
		x1 := cx + dx
		ebitenutil.DrawLine(screen, x0, cy+float64(y), x1, cy+float64(y), clr)
	}
}

func (g *Game) drawPlayer(screen *ebiten.Image, p *Player) {
	op := &ebiten.DrawImageOptions{}

	scale := 4.0
	op.GeoM.Scale(scale, scale)
	op.GeoM.Translate(-float64(frameWidth)*scale/2, -float64(frameHeight)*scale/2)
	op.GeoM.Translate(playerX, p.Y)

	i := (g.count / 5) % frameCount
	sx, sy := frameOX+i*frameWidth, frameOY

	screen.DrawImage(runnerImage.SubImage(image.Rect(sx, sy, sx+frameWidth, sy+frameHeight)).(*ebiten.Image), op)
}

func drawEnemy(screen *ebiten.Image, e *Enemy) {
	clr := color.RGBA{R: 200, G: 0, B: 0, A: 255}
	headRadius := 15.0
	headX := e.X
	headY := e.Y - float64(playerHeight) - headRadius
	drawCircle(screen, headX, headY, headRadius, clr)
	if e.Dead {
		offset := headRadius * 0.7
		ebitenutil.DrawLine(screen, headX-offset, headY-offset, headX+offset, headY+offset, color.White)
		ebitenutil.DrawLine(screen, headX+offset, headY-offset, headX-offset, headY+offset, color.White)
	}
	ebitenutil.DrawLine(screen, headX, headY+headRadius, headX, e.Y, clr)
	shoulderY := headY + headRadius*2
	armLength := 20.0
	ebitenutil.DrawLine(screen, headX, shoulderY, headX-armLength, shoulderY, clr)
	ebitenutil.DrawLine(screen, headX, shoulderY, headX+armLength, shoulderY, clr)
	legLength := 20.0
	if !e.Dead {
		legOffset := 5.0 * math.Sin(e.WalkPhase)
		ebitenutil.DrawLine(screen, headX, e.Y, headX-10+legOffset, e.Y+legLength, clr)
		ebitenutil.DrawLine(screen, headX, e.Y, headX+10-legOffset, e.Y+legLength, clr)
	} else {
		ebitenutil.DrawLine(screen, headX, e.Y, headX-15, e.Y+legLength, clr)
		ebitenutil.DrawLine(screen, headX, e.Y, headX+15, e.Y+legLength, clr)
	}
}

func (g *Game) Update() error {
	g.count++
	dt := 1.0 / 60.0
	g.time += dt
	g.updateInput()
	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
	skyColor := color.RGBA{R: 30, G: 30, B: 80, A: 255}
	screen.Fill(skyColor)

	drawFilledCircle(screen, g.state.Sun.X, g.state.Sun.Y, 40, g.state.Sun.Color)

	ebitenutil.DrawRect(screen, 0, float64(groundY), float64(screenWidth), float64(screenHeight)-float64(groundY), color.RGBA{R: 80, G: 50, B: 20, A: 255})

	for id, p := range g.state.Players {
		if id == g.localPlayerID {
			g.drawPlayer(screen, p)
		} else {
			g.drawPlayer(screen, p)
			ebitenutil.DebugPrintAt(screen, id, 10, 20)
		}
	}

	for _, e := range g.state.Enemies {
		drawEnemy(screen, e)
	}

	for _, b := range g.state.Bullets {
		clr := color.White
		if b.From == "enemy" {
			clr = color.Gray16{0x8888}
		}
		ebitenutil.DrawCircle(screen, b.X, b.Y, 2, clr)
	}

	scoreStr := fmt.Sprintf("Pontos: %d  Nível: %d", g.state.Points, g.state.Level)
	ebitenutil.DebugPrintAt(screen, scoreStr, screenWidth/2-100, 0)

	if g.state.GameOver {
		gameOverStr := "Você Perdeu! Pressione R para Recomeçar"
		ebitenutil.DebugPrintAt(screen, gameOverStr, screenWidth/2-100, screenHeight/2)
	}
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return screenWidth, screenHeight
}

func main() {
	playerID := "player1"
	if len(os.Args) > 1 {
		playerID = os.Args[1]
	}

	img, _, err := image.Decode(bytes.NewReader(images.Runner_png))
	if err != nil {
		log.Fatal(err)
	}
	runnerImage = ebiten.NewImageFromImage(img)

	game := NewGame()
	game.localPlayerID = playerID
	game.connectWebSocket()
	ebiten.SetWindowSize(screenWidth, screenHeight)
	ebiten.SetWindowTitle("Jogo Multiplayer com WebSocket e Ebiten")
	if err := ebiten.RunGame(game); err != nil {
		log.Fatal(err)
	}
}
