package main

import (
	"bufio"
	"bytes"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"net/http/httputil"
	"net/url"
	"os"
	"os/signal"
	"strings"
	"time"

	"github.com/gorilla/websocket"
)

var TIMESWAIT = 0
var TIMESWAITMAX = 5

func main() {
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	URL := url.URL{Scheme: "ws", Host: "localhost:8080", Path: "ws"}
	c, _, err := websocket.DefaultDialer.Dial(URL.String(), http.Header{"USER": []string{"hatajoe"}})
	if err != nil {
		log.Println(err)
		return
	}
	defer c.Close()

	done := make(chan struct{})
	go func() {
		defer close(done)
		for {
			messageType, buf, err := c.ReadMessage()
			if err != nil {
				log.Println("ReadMessage() error:", err)
				return
			}
			switch {
			case messageType == websocket.PongMessage:
				log.Println("receive pong")
			default:
				re := bufio.NewReader(bytes.NewReader(buf))
				r, err := http.ReadRequest(re)
				if err != nil {
					log.Println(err)
					continue
				}
				// proxy req to local server
				log.Println(r)
				body, err := io.ReadAll(r.Body)
				if err != nil {
					log.Println(err)
				}
				if req, err := http.NewRequest(r.Method, r.URL.String(), strings.NewReader(string(body))); err != nil {
					log.Println(err)
				} else {
					url, _ := url.Parse("http://localhost:8082")
					proxy := httputil.NewSingleHostReverseProxy(url)
					w := httptest.NewRecorder()
					proxy.ServeHTTP(w, req)
					if err := c.WriteMessage(websocket.BinaryMessage, w.Body.Bytes()); err != nil {
						log.Println(err)
					}
				}
			}
		}
	}()

	for {
		select {
		case <-time.After(2 * time.Second):
			if err := c.WriteMessage(websocket.PingMessage, []byte("PING")); err != nil {
				log.Println(err)
			}
		case <-done:
			return
		case <-interrupt:
			log.Println("Caught interrupt signal - quitting!")
			err := c.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))

			if err != nil {
				log.Println("Write close error:", err)
				return
			}
			select {
			case <-done:
			case <-time.After(2 * time.Second):
			}
			return
		}
	}
}
