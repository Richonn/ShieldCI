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

	branchRef := "refs/heads" + cfg.BranchName
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

	var treeEntries []*github.TreeEntry
	for _, f := range files {
		blob, _, err := client.Git.CreateBlob(ctx, owner, repo, &github.Blob{
			Content:  github.String(string(f.Content)),
			Encoding: github.String("utf-8"),
		})
		if err != nil {
			return nil, fmt.Errorf("create blob for %s: %w", f.Path, err)
		}
		treeEntries = append(treeEntries, &github.TreeEntry{
			Path: github.String(".github/workflows/" + f.Path),
			Mode: github.String("100644"),
			Type: github.String("blob"),
			SHA:  blob.SHA,
		})
	}

	tree, _, err := client.Git.CreateTree(ctx, owner, repo, baseSHA, treeEntries)
	if err != nil {
		return nil, fmt.Errorf("create tree: %w", err)
	}

	commit, _, err := client.Git.CreateCommit(ctx, owner, repo, &github.Commit{
		Message: github.String("chore: add ShieldCI generated workflows"),
		Tree:    tree,
		Parents: []*github.Commit{{SHA: github.String(baseSHA)}},
	}, nil)
	if err != nil {
		return nil, fmt.Errorf("create commit: %w", err)
	}

	_, _, err = client.Git.UpdateRef(ctx, owner, repo, &github.Reference{
		Ref:    github.String(branchRef),
		Object: &github.GitObject{SHA: commit.SHA},
	}, true)
	if err != nil {
		return nil, fmt.Errorf("update branch ref: %w", err)
	}

	prURL, err := createOrGetPR(ctx, client, owner, repo, cfg, prBody)
	if err != nil {
		return nil, err
	}

	ensureLabels(ctx, client, owner, repo)
	addLabels(ctx, client, owner, repo, prURL)

	return buildResult(prURL, stack, files), nil
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

func addLabels(ctx context.Context, client *github.Client, owner, repo, prURL string) {
	// plus tard
}

func buildResult(prURL string, stack *detect.StackConfig, files []generate.GeneratedFile) *PRResult {
	filesList := ""
	for _, f := range files {
		filesList += ".github/workflows/" + f.Path + "\n"
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
