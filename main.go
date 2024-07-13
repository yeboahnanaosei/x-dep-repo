package main

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
)

func main() {
	http.HandleFunc("/", eventHandler)
	log.Fatalln(http.ListenAndServe(":9922", nil))
}

func eventHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		requestPayload, _ := io.ReadAll(r.Body)
		payload := map[string]any{}
		if err := json.Unmarshal(requestPayload, &payload); err != nil {
			_, _ = w.Write([]byte(err.Error()))
			return
		}

		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent(" ", "   ")
		if err := enc.Encode(payload); err != nil {
			_, _ = w.Write([]byte(err.Error()))
			return
		}

		switch strings.ToLower(r.Header.Get("x-github-event")) {
		case "pull_request":
			_, _ = w.Write([]byte("A pull request was merged! A deployment should start now..."))
			return
		}

		_, _ = w.Write([]byte("OK"))
	}
	_, _ = w.Write([]byte("Hello, World!"))
}
