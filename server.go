package main

import (
	"embed"
	"flag"
	"log"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
	"github.com/olahol/melody"
)

var (
	//go:embed jsnes
	jsnesDir embed.FS

	//go:embed index.html
	indexHTML []byte
)

func main() {
	flag.Parse()

	f := flag.Arg(0)

	if f == "" {
		log.Fatalln("no rom file")
	}

	m := melody.New()

	size := 65536
	m.Upgrader = &websocket.Upgrader{
		ReadBufferSize:  size,
		WriteBufferSize: size,
	}
	m.Config.MaxMessageSize = int64(size)
	m.Config.MessageBufferSize = 2048

	http.Handle("/jsnes/", http.FileServer(http.FS(jsnesDir)))

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}

		w.Write(indexHTML)
	})

	http.HandleFunc("/rom", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, f)
	})

	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		m.HandleRequest(w, r)
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

	m.HandleMessageBinary(func(s *melody.Session, msg []byte) {
		partner := pairs[s]
		if partner != nil {
			partner.WriteBinary(msg)
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

	http.ListenAndServe(":5000", nil)
}
