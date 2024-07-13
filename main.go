package main

import (
	"context"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"strings"

	"github.com/google/go-github/v63/github"
)

var client *github.Client

func main() {
	client = github.NewClient(nil)
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

		switch strings.ToLower(r.Header.Get("x-github-event")) {
		case "pull_request":
			if payload["action"].(string) == "closed" && payload["pull_request"].(map[string]any)["merged"].(bool) {
				startDeployment(payload)
				_, _ = w.Write([]byte("A pull request was merged! A deployment should start now..."))
				return
			}
		}

		_, _ = w.Write([]byte("OK"))
	}
	_, _ = w.Write([]byte("Hello World!"))
}

func startDeployment(payload map[string]any) {
	user := payload["pull_request"].(map[string]any)["user"].(map[string]any)["login"].(string)
	repo := payload["repository"].(map[string]any)["name"].(string)
	ref := payload["pull_request"].(map[string]any)["head"].(map[string]any)["sha"].(string)
	env := "production"
	desc := "Just a deployment"

	deployment, res, err := client.Repositories.CreateDeployment(
		context.Background(),
		user,
		repo,
		&github.DeploymentRequest{Ref: &ref, Environment: &env, Description: &desc},
	)

	if err != nil {
		log.Println(err)
		log.Println(res.StatusCode)
		return
	}

	log.Println(deployment)
}
