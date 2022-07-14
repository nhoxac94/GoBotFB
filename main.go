package main

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

func main() {
	r := mux.NewRouter()
	r.HandleFunc("/", goBotHandler)
	if err := http.ListenAndServe(":8080", r); err != nil {
		log.Fatal(err.Error())
	}

}

func goBotHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		verifyWebhook(w, r)
	case "POST":
		processWebhook(w, r)
	default:
		w.WriteHeader(http.StatusBadRequest)
		log.Printf("Method %v is not support!", r.Method)
	}
}

func verifyWebhook(w http.ResponseWriter, r *http.Request) {
	mode := r.URL.Query().Get("hub.mode")
	challenge := r.URL.Query().Get("hub.challenge")
	token := r.URL.Query().Get("hub.verify_token")
	if mode == "subscribe" && token == "GoBot" {
		w.WriteHeader(200)
		w.Write([]byte(challenge))
	} else {
		w.WriteHeader(404)
		w.Write([]byte("Error, wrong validation token!"))
	}
}

func processWebhook(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	var req Request
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		w.WriteHeader(404)
		w.Write([]byte("Message not support"))
		return
	}
	if req.Object == "page" {
		for _, entry := range req.Entry {
			for _, event := range entry.Messaging {
				if event.Message != nil {
					processMessage(&event)
				}
			}
		}
		w.WriteHeader(200)
		w.Write([]byte("Got your message!"))
	} else {
		w.WriteHeader(404)
		w.Write([]byte("Message not support"))
	}
}
