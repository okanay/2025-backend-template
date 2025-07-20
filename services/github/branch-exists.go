package GithubService

import (
	"context"
	"net/http"

	"github.com/google/go-github/github"
)

func (r *Service) BranchExists(branch string) (bool, error) {
	_, _, err := r.githubClient.Repositories.GetBranch(context.Background(), r.RepoOwner, r.RepoName, branch)
	if err != nil {
		if ghErr, ok := err.(*github.ErrorResponse); ok && ghErr.Response.StatusCode == http.StatusNotFound {
			return false, nil
		}
		return false, err
	}
	return true, nil
}
