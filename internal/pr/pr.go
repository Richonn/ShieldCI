package pr

import (
	"context"
	"fmt"

	"github.com/Richonn/shieldci/internal/config"
	"github.com/Richonn/shieldci/internal/detect"
	"github.com/Richonn/shieldci/internal/generate"
	"github.com/google/go-github/v60/github"
	"golang.org/x/oauth2"
)

type PRResult struct {
	PRURL     string
	StackJSON string
	FilesList string
}

func CreateOrUpdatePR(ctx context.Context, cfg *config.Config,
	stack *detect.StackConfig, files []generate.GeneratedFile,
	prBody string) (*PRResult, error) {
	client := newClient(ctx, cfg.GithubToken)
	owner, repo := cfg.RepoOwner, cfg.RepoName

	ref, _, err := client.Git.GetRef(ctx, owner, repo, "refs/heads/main")
	if err != nil {
		return nil, fmt.Errorf("get main ref: %w", err)
	}
	baseSHA := ref.Object.GetSHA()

	branchRef := "refs/heads/" + cfg.BranchName
	_, _, err = client.Git.GetRef(ctx, owner, repo, branchRef)
	if err != nil {
		_, _, err = client.Git.CreateRef(ctx, owner, repo, &github.Reference{
			Ref:    github.String(branchRef),
			Object: &github.GitObject{SHA: github.String(baseSHA)},
		})
		if err != nil {
			return nil, fmt.Errorf("create branch: %w", err)
		}
	}

	for _, f := range files {
		filePath := ".github/workflows/" + f.Path
		if err := upsertFile(ctx, client, owner, repo, filePath, cfg.BranchName, f.Content); err != nil {
			return nil, fmt.Errorf("upsert %s: %w", filePath, err)
		}
	}

	prURL, err := createOrGetPR(ctx, client, owner, repo, cfg, prBody)
	if err != nil {
		return nil, err
	}

	if prNumber := extractPRNumber(prURL); prNumber > 0 {
		ensureLabels(ctx, client, owner, repo)
		_, _, _ = client.Issues.AddLabelsToIssue(ctx, owner, repo, prNumber, []string{"automated", "ci-cd"})
	}

	return buildResult(prURL, stack, files), nil
}

func extractPRNumber(prURL string) int {
	var number int
	for i := len(prURL) - 1; i >= 0; i-- {
		if prURL[i] == '/' {
			fmt.Sscanf(prURL[i+1:], "%d", &number)
			return number
		}
	}
	return 0
}

func upsertFile(ctx context.Context, client *github.Client, owner, repo, path, branch string, content []byte) error {
	opts := &github.RepositoryContentFileOptions{
		Message: github.String("chore: add ShieldCI generated workflows [skip ci]"),
		Content: content,
		Branch:  github.String(branch),
	}

	existing, _, _, _ := client.Repositories.GetContents(ctx, owner, repo, path,
		&github.RepositoryContentGetOptions{Ref: branch})
	if existing != nil {
		opts.SHA = existing.SHA
		_, _, err := client.Repositories.UpdateFile(ctx, owner, repo, path, opts)
		return err
	}

	_, _, err := client.Repositories.CreateFile(ctx, owner, repo, path, opts)
	return err
}

func createOrGetPR(ctx context.Context, client *github.Client, owner,
	repo string, cfg *config.Config, body string) (string, error) {
	pr, _, err := client.PullRequests.Create(ctx, owner, repo, &github.NewPullRequest{
		Title: github.String(cfg.PRTitle),
		Head:  github.String(cfg.BranchName),
		Base:  github.String("main"),
		Body:  github.String(body),
	})
	if err == nil {
		return pr.GetHTMLURL(), nil
	}

	prs, _, listErr := client.PullRequests.List(ctx, owner, repo, &github.PullRequestListOptions{
		Head:  owner + ":" + cfg.BranchName,
		State: "open",
	})
	if listErr != nil || len(prs) == 0 {
		return "", fmt.Errorf("create PR: %w", err)
	}
	return prs[0].GetHTMLURL(), nil
}

func ensureLabels(ctx context.Context, client *github.Client, owner, repo string) {
	labels := []struct{ name, color string }{
		{"automated", "0075ca"},
		{"ci-cd", "e4e669"},
	}
	for _, l := range labels {
		_, _, _ = client.Issues.CreateLabel(ctx, owner, repo, &github.Label{
			Name:  github.String(l.name),
			Color: github.String(l.color),
		})
	}
}

func buildResult(prURL string, stack *detect.StackConfig, files []generate.GeneratedFile) *PRResult {
	filesList := ""
	for i, f := range files {
		if i > 0 {
			filesList += ","
		}
		filesList += ".github/workflows/" + f.Path
	}
	stackJSON := fmt.Sprintf(`{"language":"%s","docker":%t,"k8s":%t}`,
		stack.Language, stack.HasDocker, stack.HasK8s)

	return &PRResult{
		PRURL:     prURL,
		StackJSON: stackJSON,
		FilesList: filesList,
	}
}

func newClient(ctx context.Context, token string) *github.Client {
	ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: token})
	tc := oauth2.NewClient(ctx, ts)
	return github.NewClient(tc)
}
