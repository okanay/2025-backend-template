package GithubService

import (
	"context"

	"github.com/google/go-github/github"
	"golang.org/x/oauth2"
)

type Service struct {
	RepoOwner    string
	RepoName     string
	githubClient *github.Client
}

// NewService, yeni bir Repository örneği oluşturur ve GitHub istemcisini başlatır.
func NewService(repoOwner, repoName, token string) *Service {
	ctx := context.Background()
	ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: token})
	tc := oauth2.NewClient(ctx, ts)
	client := github.NewClient(tc)

	return &Service{
		RepoOwner:    repoOwner,
		RepoName:     repoName,
		githubClient: client,
	}
}
