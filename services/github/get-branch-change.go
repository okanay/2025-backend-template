package GithubService

import (
	"context"
)

type Change struct {
	Path   string
	Status string
}

func (r *Service) GetBranchChanges(baseBranch, compareBranch string) ([]Change, error) {
	comparison, _, err := r.githubClient.Repositories.CompareCommits(
		context.Background(),
		r.RepoOwner,
		r.RepoName,
		baseBranch,
		compareBranch,
	)
	if err != nil {
		return nil, err
	}

	var changes []Change
	for _, file := range comparison.Files {
		status := "modified"
		switch *file.Status {
		case "added":
			status = "added"
		case "removed":
			status = "deleted"
		case "modified":
			status = "modified"
		}

		changes = append(changes, Change{
			Path:   *file.Filename,
			Status: status,
		})
	}

	return changes, nil
}
