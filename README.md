# Golang Game with WebSockets 🎮🚀  

Este é um projeto de jogo multiplayer desenvolvido com **Golang**, utilizando a biblioteca **Ebiten** para renderização gráfica e o **Fiber** para a comunicação via **WebSockets**.  

## 🚀 Tecnologias Utilizadas  
- **[Ebiten](https://ebiten.org/)** → Biblioteca para criação de jogos 2D em Go  
- **[Fiber](https://gofiber.io/)** → Framework web inspirado no Express.js, usado para WebSockets  
- **WebSockets** → Comunicação em tempo real entre os jogadores  

## 🎯 Objetivo do Projeto  
Criar um jogo multiplayer simples e eficiente, onde os jogadores possam interagir em tempo real através de WebSockets.  

## 🛠️ Como Rodar o Projeto  
1. Clone o repositório:  
   ```sh
   git clone https://github.com/wenealves10/golang-game-with-websockets.git
   cd golang-game-with-websockets
   ```
2. Instale as dependências:
   ```sh 
   go mod tidy
   ```
3. Inicie o Game Server:
   ```sh go run server/main.go```
4. Inicie o Game Client:
   ```sh go run client/main.go```

