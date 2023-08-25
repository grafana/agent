package main

import (
	"context"
	"fmt"
	"github.com/bradleyfalzon/ghinstallation"
	"github.com/google/go-github/github"
	"log"
	"net/http"
	"os"
	"strconv"
)

func main() {
	l := log.New(os.Stderr, "gen-github-token: ", 0)
	key := os.Getenv("GITHUB_APP_PRIVATE_KEY")
	appId := os.Getenv("GITHUB_APP_ID")

	if key == "" || appId == "" {
		l.Println("GITHUB_APP_PRIVATE_KEY and GITHUB_APP_ID must be set")
		os.Exit(1)
	}

	appIdInt, err := strconv.ParseInt(appId, 10, 64)
	if err != nil {
		l.Println("failed to parse GITHUB_APP_ID")
		l.Println(err)
		os.Exit(1)
	}

	itr, err := ghinstallation.NewAppsTransport(http.DefaultTransport, appIdInt, []byte(key))
	if err != nil {
		l.Println("failed to create transport for GH app")
		l.Println(err)
		os.Exit(1)
	}
	client := github.NewClient(&http.Client{Transport: itr})

	token, _, err := client.Apps.CreateInstallationToken(context.Background(), appIdInt)
	if err != nil {
		l.Println("failed to create installation token")
		l.Println(err)
		os.Exit(1)
	}
	fmt.Print(token.Token)
	os.Exit(0)
}
