package GithubService

import (
	"context"

	"github.com/google/go-github/github"
)

func (r *Service) GetFileContent(branch, path string) ([]byte, string, error) {
	opts := &github.RepositoryContentGetOptions{Ref: branch}
	fileContent, _, _, err := r.githubClient.Repositories.GetContents(context.Background(), r.RepoOwner, r.RepoName, path, opts)
	if err != nil {
		return nil, "", err
	}
	decodedContent, err := fileContent.GetContent()
	if err != nil {
		return nil, "", err
	}
	return []byte(decodedContent), *fileContent.SHA, nil
}
