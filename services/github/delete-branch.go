package GithubService

import "context"

func (r *Service) DeleteBranch(branch string) error {
	_, err := r.githubClient.Git.DeleteRef(context.Background(), r.RepoOwner, r.RepoName, "refs/heads/"+branch)
	return err
}
