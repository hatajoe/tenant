package main

import (
	"bytes"
	"io"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
	"sync"

	"github.com/gorilla/websocket"
)

var tenants = sync.Map{}

type Tenant struct {
	conn *websocket.Conn
	ch   chan []byte
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

func handler(w http.ResponseWriter, r *http.Request) {
	user := r.Header.Get("USER")
	if tenant, ok := tenants.Load(user); !ok {
		// pass through upstream
		log.Println("passing through upstream")

		body, err := io.ReadAll(r.Body)
		defer r.Body.Close()
		if err != nil {
			log.Fatal(err)
		}
		if req, err := http.NewRequest(r.Method, r.URL.String(), strings.NewReader(string(body))); err != nil {
			log.Fatal(err)
		} else {
			url, err := url.Parse("http://localhost:8081")
			if err != nil {
				log.Fatal(err)
			}
			proxy := httputil.NewSingleHostReverseProxy(url)
			proxy.ServeHTTP(w, req)
		}
	} else {
		t := tenant.(*Tenant)
		// write http request to websocket
		bin := &bytes.Buffer{}
		if err := r.Write(bin); err != nil {
			log.Fatal(err)
		}
		if err := t.conn.WriteMessage(websocket.BinaryMessage, bin.Bytes()); err != nil {
			log.Fatal(err)
		}
		// read http response from websocket
		b := <-t.ch
		if _, err := w.Write(b); err != nil {
			log.Fatal(err)
		}
	}
}

func ws(w http.ResponseWriter, r *http.Request) {
	upgrader.CheckOrigin = func(r *http.Request) bool { return true }

	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
	}
	user := r.Header.Get("USER")
	if user == "" {
		log.Println("USER header value is empty")
		return
	}
	ch := make(chan []byte)
	defer close(ch)

	tenants.Store(user, &Tenant{
		conn: ws,
		ch:   ch,
	})
	defer tenants.Delete(user)

	log.Println(user)

	for {
		messageType, p, err := ws.ReadMessage()
		if err != nil {
			log.Println(err)
			return
		}
		switch {
		case messageType == websocket.CloseMessage:
			log.Printf("%s connection closed", user)
			return
		case messageType == websocket.PingMessage:
			log.Printf("receive ping from %s", user)
			if err := ws.WriteMessage(websocket.PongMessage, []byte("PONG")); err != nil {
				log.Println(err)
			}
		default:
			ch <- p
		}
	}
}

func main() {
	http.HandleFunc("/", handler)
	http.HandleFunc("/ws", ws)

	log.Fatal(http.ListenAndServe(":8080", nil))
}
