package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/google/go-github/v63/github"
	"golang.org/x/oauth2"
)

var client *github.Client

func main() {
	ctx := context.Background()
	ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: os.Getenv("GH_TOKEN")})
	tc := oauth2.NewClient(ctx, ts)
	client = github.NewClient(tc)
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

		case "deployment":
			processDeployment(payload)
		case "deployment_status":
			json.NewEncoder(os.Stdout).Encode(payload)
		}

		_, _ = w.Write([]byte("OK"))
	}
	_, _ = w.Write([]byte("Hello World!"))
}

func startDeployment(payload map[string]any) {
	user := payload["pull_request"].(map[string]any)["user"].(map[string]any)["login"].(string)
	repo := payload["repository"].(map[string]any)["name"].(string)
	ref := github.String(payload["pull_request"].(map[string]any)["head"].(map[string]any)["sha"].(string))
	env := github.String("production")
	desc := github.String("Just a deployment")

	deployment, res, err := client.Repositories.CreateDeployment(
		context.Background(),
		user,
		repo,
		&github.DeploymentRequest{
			Ref:         ref,
			Environment: env,
			Description: desc,
			AutoMerge:   github.Bool(false),
		},
	)

	if err != nil {
		log.Println(err)
		log.Println(res.StatusCode)
		return
	}

	log.Println("Deployment created: ")
	log.Println(deployment)
	log.Println()
	log.Println()
}

func processDeployment(payload map[string]any) {
	desc := payload["deployment"].(map[string]any)["description"].(string)
	deployUser := payload["deployment"].(map[string]any)["creator"].(map[string]any)["login"].(string)
	env := payload["deployment"].(map[string]any)["environment"].(string)
	repo := payload["repository"].(map[string]any)["name"].(string)
	deploymentID := payload["deployment"].(map[string]any)["id"].(int64)

	fmt.Printf("Deployment [%s] created by [%s] for environment [%s]\n", desc, deployUser, env)
	time.Sleep(time.Second * 10)
	_, res, err := client.Repositories.CreateDeploymentStatus(
		context.Background(),
		deployUser,
		repo,
		deploymentID,
		&github.DeploymentStatusRequest{
			State:       github.String("pending"),
			Description: github.String("Deployment pending"),
		},
	)

	if err != nil {
		log.Println(err)
		log.Println(res.StatusCode)
		return
	}

	time.Sleep(time.Second * 10)

	_, res, err = client.Repositories.CreateDeploymentStatus(
		context.Background(),
		deployUser,
		repo,
		deploymentID,
		&github.DeploymentStatusRequest{
			State:       github.String("success"),
			Description: github.String("Deployment success"),
		},
	)
}
