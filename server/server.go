package main

import (
	"encoding/json"
	"image/color"
	"log"
	"math"
	"math/rand/v2"
	"sync"
	"time"

	"github.com/gofiber/fiber/v2"
	fws "github.com/gofiber/websocket/v2"
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
	playerBulletSpeed = 500.0
	enemyBulletSpeed  = -300.0
	periodSun         = 30.0
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
	X     float64     `json:"x"`
	Y     float64     `json:"y"`
	Color color.Color `json:"color"`
}

type GameState struct {
	Sun      Sun                `json:"sun"`
	Players  map[string]*Player `json:"players"`
	Enemies  []*Enemy           `json:"enemies"`
	Bullets  []*Bullet          `json:"bullets"`
	Points   int                `json:"points"`
	Level    int                `json:"level"`
	time     float64
	GameOver bool `json:"gameOver"`
}

type Message struct {
	Type     string `json:"type"`
	PlayerID string `json:"playerId"`
	Command  string `json:"command"`
}

var (
	gameState = GameState{
		Players: make(map[string]*Player),
		Enemies: []*Enemy{},
		Bullets: []*Bullet{},
		Points:  0,
		Level:   1,
	}
	stateMutex sync.Mutex

	clients      = make(map[string]*Client)
	clientsMutex sync.Mutex
)

type Client struct {
	ID   string
	Conn *fws.Conn
}

func updateSun() (sunX, sunY float64, sunColor color.Color) {
	centerX := float64(screenWidth) / 2
	centerY := float64(screenHeight)
	radius := float64(screenWidth) / 2
	progress := math.Mod(gameState.time, periodSun) / periodSun
	theta := math.Pi - progress*math.Pi
	sunX = centerX + radius*math.Cos(theta)
	sunY = centerY - radius*math.Sin(theta)
	startG := 255.0
	endG := 100.0
	G := uint8(lerp(startG, endG, progress))
	sunColor = color.RGBA{R: 255, G: G, B: 0, A: 255}
	return
}

func updateEnemies(dt float64) {
	if len(gameState.Enemies) == 0 {
		enemy := Enemy{
			X:          float64(screenWidth) + 50,
			Y:          float64(groundY) - 10,
			Vx:         -100 - float64(gameState.Level)*10,
			Vy:         0,
			ShootTimer: 2.0 + rand.Float64()*1.0,
			Dead:       false,
			DeathTimer: 0,
			WalkPhase:  0,
		}
		gameState.Enemies = append(gameState.Enemies, &enemy)
	}

	for i := range gameState.Enemies {
		if gameState.Enemies[i].Dead {
			gameState.Enemies[i].DeathTimer += dt
			gameState.Enemies[i].Vy += gravity * dt
			gameState.Enemies[i].Y += gameState.Enemies[i].Vy * dt
		} else {
			gameState.Enemies[i].X += gameState.Enemies[i].Vx * dt
			gameState.Enemies[i].WalkPhase += dt * 4
			gameState.Enemies[i].ShootTimer -= dt
			if gameState.Enemies[i].ShootTimer <= 0 {
				bullet := Bullet{
					X:    gameState.Enemies[i].X,
					Y:    gameState.Enemies[i].Y - float64(playerHeight)/2,
					Vx:   enemyBulletSpeed,
					Vy:   0,
					From: "enemy",
				}
				gameState.Bullets = append(gameState.Bullets, &bullet)
				gameState.Enemies[i].ShootTimer = 1.5 + rand.Float64()*1.0
			}
		}
	}

	newEnemies := gameState.Enemies[:0]
	for _, e := range gameState.Enemies {
		if e.X > -50 && e.Y < float64(groundY)+100 {
			newEnemies = append(newEnemies, e)
		}
	}
	gameState.Enemies = newEnemies
}

func lerp(a, b, t float64) float64 {
	return a + (b-a)*t
}

func rectsOverlap(a, b struct{ x, y, w, h float64 }) bool {
	return a.x < b.x+b.w &&
		a.x+a.w > b.x &&
		a.y < b.y+b.h &&
		a.y+a.h > b.y
}

func checkCollisions() {
	for _, enemy := range gameState.Enemies {
		if !enemy.Dead {
			for _, bullet := range gameState.Bullets {
				if bullet.From == "player" {
					dx := enemy.X - bullet.X
					dy := enemy.Y - bullet.Y
					if math.Sqrt(dx*dx+dy*dy) < 15 {
						enemy.Dead = true
						enemy.DeathTimer = 0
						enemy.Vy = 0
						gameState.Points += 100
						bullet.X = -1000
						break
					}
				}
			}
		}
	}

	newBullets := gameState.Bullets[:0]
	for _, bullet := range gameState.Bullets {
		if bullet.X > 0 {
			newBullets = append(newBullets, bullet)
		}
	}
	gameState.Bullets = newBullets

	for _, player := range gameState.Players {
		playerRect := struct{ x, y, w, h float64 }{
			x: float64(playerX) - float64(playerWidth)/2,
			y: player.Y - float64(playerHeight),
			w: float64(playerWidth),
			h: float64(playerHeight),
		}
		for _, bullet := range gameState.Bullets {
			if bullet.From == "enemy" {
				bulletRect := struct{ x, y, w, h float64 }{
					x: bullet.X - 2.5,
					y: bullet.Y - 2.5,
					w: 5,
					h: 5,
				}
				if rectsOverlap(playerRect, bulletRect) {
					// Game over logic here
					log.Println("Game Over")
				}
			}
		}
	}

	for _, enemy := range gameState.Enemies {
		if !enemy.Dead {
			for _, player := range gameState.Players {
				dx := enemy.X - float64(playerX)
				dy := enemy.Y - player.Y
				if math.Sqrt(dx*dx+dy*dy) < 20 {
					// Game over logic here
					log.Println("Game Over")
				}
			}
		}
	}
}

func gameLoop() {
	ticker := time.NewTicker(16 * time.Millisecond)
	dt := 1.0 / 60.0
	for range ticker.C {
		stateMutex.Lock()
		gameState.time += dt
		gameState.Sun.X, gameState.Sun.Y, gameState.Sun.Color = updateSun()

		for _, p := range gameState.Players {
			p.Vy += gravity * dt
			p.Y += p.Vy * dt
			if p.Y > float64(groundY) {
				p.Y = float64(groundY)
				p.Vy = 0
			}
		}

		newPlayerBullets := gameState.Bullets[:0]
		for _, b := range gameState.Bullets {
			b.X += b.Vx * dt
			b.Y += b.Vy * dt
			if b.X > 0 && b.X < float64(screenWidth) && b.Y > 0 && b.Y < float64(screenHeight) {
				newPlayerBullets = append(newPlayerBullets, b)
			}
		}
		gameState.Bullets = newPlayerBullets

		updateEnemies(dt)

		checkCollisions()

		stateMutex.Unlock()

		broadcastGameState()
	}
}

func broadcastGameState() {
	stateMutex.Lock()
	data, err := json.Marshal(gameState)
	stateMutex.Unlock()
	if err != nil {
		log.Println("Erro ao serializar o estado:", err)
		return
	}
	clientsMutex.Lock()
	defer clientsMutex.Unlock()
	for _, c := range clients {
		if err := c.Conn.WriteMessage(1, data); err != nil {
			log.Println("Erro ao enviar para", c.ID, err)
		}
	}
}

func wsHandler(c *fws.Conn) {
	playerID := c.Query("id")
	if playerID == "" {
		playerID = c.RemoteAddr().String()
	}
	client := &Client{ID: playerID, Conn: c}
	clientsMutex.Lock()
	clients[playerID] = client
	clientsMutex.Unlock()
	log.Println("Cliente conectado:", playerID)

	stateMutex.Lock()
	gameState.Players[playerID] = &Player{ID: playerID, Y: float64(groundY), Vy: 0}
	stateMutex.Unlock()

	for {
		_, msg, err := c.ReadMessage()
		if err != nil {
			log.Println("Erro ao ler mensagem de", playerID, ":", err)
			break
		}
		var m Message
		if err := json.Unmarshal(msg, &m); err != nil {
			log.Println("Erro ao decodificar mensagem de", playerID, ":", err)
			continue
		}
		handleMessage(m)
	}

	clientsMutex.Lock()
	delete(clients, playerID)
	clientsMutex.Unlock()
	stateMutex.Lock()
	delete(gameState.Players, playerID)
	stateMutex.Unlock()
	log.Println("Cliente desconectado:", playerID)
}

func handleMessage(m Message) {
	stateMutex.Lock()
	defer stateMutex.Unlock()
	p, ok := gameState.Players[m.PlayerID]
	if !ok {
		return
	}
	switch m.Command {
	case "reset":
		p.Y = -float64(playerHeight)
		p.Vy = 0
		gameState.Points = 0
		gameState.Level = 1
		gameState.GameOver = false
		gameState.Enemies = []*Enemy{}
		gameState.Bullets = []*Bullet{}
		gameState.time = 0
		go func() {
			for p.Y < float64(groundY) {
				time.Sleep(16 * time.Millisecond)
				stateMutex.Lock()
				p.Vy += gravity * (1.0 / 60.0)
				p.Y += p.Vy * (1.0 / 60.0)
				if p.Y > float64(groundY) {
					p.Y = float64(groundY)
					p.Vy = 0
				}
				stateMutex.Unlock()
			}
		}()
	case "jump":
		if p.Y >= float64(groundY) {
			p.Vy = jumpImpulse
		}
	case "shoot":
		bullet := &Bullet{
			X:    float64(playerX),
			Y:    p.Y - float64(playerHeight)/2,
			Vx:   playerBulletSpeed,
			Vy:   0,
			From: "player",
		}
		gameState.Bullets = append(gameState.Bullets, bullet)
	}
}

func main() {
	go gameLoop()

	app := fiber.New()

	app.Get("/ws", fws.New(wsHandler))

	log.Println("Servidor iniciado na porta 3000")
	log.Fatal(app.Listen(":3000"))
}
