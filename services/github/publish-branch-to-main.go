package GithubService

import (
	"context"
	"fmt"

	"github.com/google/go-github/github"
)

func (r *Service) PublishBranchToMain(branch string) (report string, err error) {
	// 1. Pull Request oluştur.
	prTitle := fmt.Sprintf("feat: Publish changes from %s", branch)
	prBody := "This pull request was automatically generated to publish changes."
	newPR := &github.NewPullRequest{
		Title: &prTitle,
		Head:  &branch,
		Base:  github.String("main"),
		Body:  &prBody,
	}
	pr, _, err := r.githubClient.PullRequests.Create(context.Background(), r.RepoOwner, r.RepoName, newPR)
	if err != nil {
		return "Pull request oluşturulamadı", err
	}

	// 2. Oluşturulan Pull Request'i birleştir (merge).
	mergeResult, _, err := r.githubClient.PullRequests.Merge(context.Background(), r.RepoOwner, r.RepoName, *pr.Number, "Automatic merge", nil)
	if err != nil {
		// Çakışma varsa, PR'ı kapatıp hata dönebiliriz.
		// r.githubClient.PullRequests.Edit(...) ile PR'ı kapat
		return "Pull request birleştirilemedi, muhtemelen çakışma var.", err
	}

	return *mergeResult.Message, nil
}
