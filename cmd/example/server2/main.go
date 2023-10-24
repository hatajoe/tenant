package main

import (
	"io"
	"log"
	"net/http"
)

func handler(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	defer r.Body.Close()
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("server2 %s\n", string(body))
	w.WriteHeader(http.StatusOK)
	if _, err := w.Write([]byte("server2")); err != nil {
		log.Println("server2 error: ", err)
	}
}

func main() {
	http.HandleFunc("/", handler)

	log.Fatal(http.ListenAndServe(":8082", nil))
}
