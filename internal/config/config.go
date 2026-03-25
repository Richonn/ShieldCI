package config

import (
	"fmt"
	"os"
	"strings"
)

type Config struct {
	GithubToken string
	Language string
	Docker string
	Kubernetes bool
	EnableTrivy bool
	EnableGitleaks bool
	EnableSAST bool
	SASTTool string
	BranchName string
	PRTitle string
	RepoOwner string
	RepoName string
	WorkspaceDir string
	OutputFile string
}

func Load() (*Config, error) {
	c := &Config{}

	c.GithubToken = os.Getenv("GITHUB_TOKEN")
	if c.GithubToken == "" {
		return nil, fmt.Errorf("GITHUB_TOKEN is required")
	}

	c.Language = getEnvDefault("INPUT_LANGUAGE", "auto")
	c.Docker = getEnvDefault("INPUT_DOCKER", "auto")
	c.Kubernetes = parseBool(getEnvDefault("INPUT_KUBERNETES", "false"))

	c.EnableTrivy = parseBool(getEnvDefault("INPUT_ENABLE_TRIVY", "true"))
	c.EnableGitleaks = parseBool(getEnvDefault("INPUT_ENABLE_GITLEAKS", "true"))
	c.EnableSAST = parseBool(getEnvDefault("INPUT_ENABLE_SAST", "true"))
	c.SASTTool = getEnvDefault("INPUT_SAST_TOOL", "codeql")

	c.BranchName = getEnvDefault("INPUT_BRANCH_NAME", "shieldci/generated-workflows")
	c.PRTitle = getEnvDefault("INPUT_PR_TITLE", "[ShieldCI] Add CI/CD DevSecOps pipeline")

	repo := os.Getenv("GITHUB_REPOSITORY")
	parts := strings.SplitN(repo, "/", 2)
	if len(parts) != 2 {
		return nil, fmt.Errorf("GITHUB_REPOSITORY %q has unexpected format", repo)
	}
	c.RepoOwner = parts[0]
	c.RepoName = parts[1]

	c.WorkspaceDir = getEnvDefault("GITHUB_WORKSPACE", "/github/workspace")
	c.OutputFile = os.Getenv("GITHUB_OUTPUT")

	return c, nil
}

func (c *Config) WriteOutput(key, value string) error {
	if c.OutputFile == "" {
		return nil
	}
	f, err := os.OpenFile(c.OutputFile, os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("open GITHUB_OUTPUT: %w", err)
	}
	defer func() { _ = f.Close() }()
	_, err = fmt.Fprintf(f, "%s=%s\n", key, value)
	return err
}

func getEnvDefault(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func parseBool(s string) bool {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "true", "1", "yes":
		return true
	default:
		return false
	}
}
