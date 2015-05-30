package main

import (
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/olahol/melody"
	"path/filepath"
	"net/http"
	"sync"
)

func main() {
	r := gin.New()
	m := melody.New()

	size := 65536
	m.Upgrader = &websocket.Upgrader{
		ReadBufferSize:  size,
		WriteBufferSize: size,
	}
	m.Config.MaxMessageSize = int64(size)
	m.Config.MessageBufferSize = 2048

	r.Static("/jsnes", "./jsnes")

	r.GET("/", func(c *gin.Context) {
		http.ServeFile(c.Writer, c.Request, "index.html")
	})

	r.GET("/gamelist", func(c *gin.Context) {
		files, _ := filepath.Glob("*.nes")
		c.JSON(200, gin.H{"games":files})
	})
	
	r.GET("/games?name=:name", func(c *gin.Context) {
		name := c.Params.ByName("name")
		http.ServeFile(c.Writer, c.Request, name)
	})

	r.GET("/ws", func(c *gin.Context) {
		m.HandleRequest(c.Writer, c.Request)
	})

	var mutex sync.Mutex
	pairs := make(map[*melody.Session]*melody.Session)

	m.HandleConnect(func(s *melody.Session) {
		mutex.Lock()
		var partner *melody.Session
		for player1, player2 := range pairs {
			if player2 == nil {
				partner = player1
				pairs[partner] = s
				partner.Write([]byte("join 1"))
				break
			}
		}
		pairs[s] = partner
		if partner != nil {
			s.Write([]byte("join 2"))
		}
		mutex.Unlock()
	})

	m.HandleMessage(func(s *melody.Session, msg []byte) {
		partner := pairs[s]
		if partner != nil {
			partner.Write(msg)
		}
	})

	m.HandleDisconnect(func(s *melody.Session) {
		mutex.Lock()
		partner := pairs[s]
		if partner != nil {
			pairs[partner] = nil
			partner.Write([]byte("part"))
		}
		delete(pairs, s)
		mutex.Unlock()
	})

	r.Run(":5000")
}
