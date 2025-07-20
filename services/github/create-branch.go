package GithubService

import (
	"context"

	"github.com/google/go-github/github"
)

func (r *Service) CreateBranch(baseBranch, newBranch string) error {
	// 1. Temel alınacak dalın (main) son commit'inin SHA'sını al.
	baseRef, _, err := r.githubClient.Git.GetRef(context.Background(), r.RepoOwner, r.RepoName, "refs/heads/"+baseBranch)
	if err != nil {
		return err
	}

	// 2. Yeni dal referansını oluştur.
	newRefStr := "refs/heads/" + newBranch
	newRef := &github.Reference{Ref: &newRefStr, Object: &github.GitObject{SHA: baseRef.Object.SHA}}
	_, _, err = r.githubClient.Git.CreateRef(context.Background(), r.RepoOwner, r.RepoName, newRef)
	return err
}
