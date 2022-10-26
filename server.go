package main

import (
	"flag"
	"log"
	"net/http"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/olahol/melody"
)

func main() {
	flag.Parse()

	f := flag.Arg(0)

	if f == "" {
		log.Fatalln("no rom file")
	}

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

	r.GET("/rom", func(c *gin.Context) {
		http.ServeFile(c.Writer, c.Request, f)
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
