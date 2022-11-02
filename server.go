package main

import (
	"embed"
	"flag"
	"log"
	"net/http"
	"sync"

	"github.com/olahol/melody"
)

var (
	//go:embed jsnes/source/*.js
	jsnesSourceDir embed.FS

	//go:embed jsnes/lib/*.js
	jsnesLibDir embed.FS

	//go:embed index.html
	indexHTML []byte
)

func main() {
	flag.Parse()

	f := flag.Arg(0)

	if f == "" {
		log.Fatalln("no rom file")
	}

	http.Handle("/jsnes/source/", http.FileServer(http.FS(jsnesSourceDir)))
	http.Handle("/jsnes/lib/", http.FileServer(http.FS(jsnesLibDir)))

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

	m := melody.New()
	m.Upgrader.ReadBufferSize = 65536
	m.Upgrader.WriteBufferSize = 65536
	m.Config.MaxMessageSize = 65536
	m.Config.MessageBufferSize = 2048

	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		m.HandleRequest(w, r)
	})

	var mutex sync.Mutex
	pairs := make(map[*melody.Session]*melody.Session)

	m.HandleConnect(func(s *melody.Session) {
		log.Println("connect")
		mutex.Lock()
		var partner *melody.Session
		for player1, player2 := range pairs {
			if player2 == nil {
				partner = player1
				pairs[partner] = s
				log.Println("start")
				partner.Write([]byte("join 1"))
				s.Write([]byte("join 2"))
				break
			}
		}
		pairs[s] = partner
		mutex.Unlock()
	})

	m.HandleMessageBinary(func(s *melody.Session, msg []byte) {
		partner := pairs[s]
		if partner != nil {
			partner.WriteBinary(msg)
		}
	})

	m.HandleDisconnect(func(s *melody.Session) {
		log.Println("disconnect")
		mutex.Lock()
		partner := pairs[s]
		if partner != nil {
			pairs[partner] = nil
			log.Println("stop")
			partner.Write([]byte("part"))
		}
		delete(pairs, s)
		mutex.Unlock()
	})

	log.Println("listening on http://localhost:5000")

	http.ListenAndServe(":5000", nil)
}
