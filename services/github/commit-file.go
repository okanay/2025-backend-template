package GithubService

import (
	"context"

	"github.com/google/go-github/github"
)

func (r *Service) CommitFile(branch, path, content, sha, message string) error {
	opts := &github.RepositoryContentFileOptions{
		Message: &message,
		Content: []byte(content),
		Branch:  &branch,
		SHA:     &sha, // Eğer dosya güncelleniyorsa, eski SHA değeri buraya gelir.
	}

	// SHA boşsa, bu yeni bir dosyadır (CreateFile). Doluysa, mevcut bir dosyadır (UpdateFile).
	if sha == "" {
		opts.SHA = nil // CreateFile için SHA boş olmalı.
		_, _, err := r.githubClient.Repositories.CreateFile(context.Background(), r.RepoOwner, r.RepoName, path, opts)
		return err
	}

	_, _, err := r.githubClient.Repositories.UpdateFile(context.Background(), r.RepoOwner, r.RepoName, path, opts)
	return err
}
