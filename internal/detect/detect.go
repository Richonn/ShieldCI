package detect

import (
	"os"
	"path/filepath"

	"github.com/Richonn/shieldci/internal/config"
)

type StackConfig struct {
	Language       string
	HasDocker      bool
	HasK8s         bool
	HasSemgrep     bool
	BuildTool      string
	EnableTrivy    bool
	EnableGitleaks bool
	EnableSAST     bool
	SASTTool       string
	RepoOwner      string
	RepoName       string
}

type Component struct {
	Path      string
	Language  string
	HasDocker bool
	HasK8s    bool
	BuildTool string
}

func Detect(cfg *config.Config) (*StackConfig, error) {
	dir := cfg.WorkspaceDir

	stack := &StackConfig{
		EnableTrivy:    cfg.EnableTrivy,
		EnableGitleaks: cfg.EnableGitleaks,
		EnableSAST:     cfg.EnableSAST,
		SASTTool:       cfg.SASTTool,
		RepoOwner:      cfg.RepoOwner,
		RepoName:       cfg.RepoName,
	}

	if cfg.Language != "auto" {
		stack.Language = cfg.Language
	} else {
		stack.Language = detectLanguage(dir)
	}

	stack.BuildTool = detectBuildTool(dir, stack.Language)

	switch cfg.Docker {
	case "true":
		stack.HasDocker = true
	case "false":
		stack.HasDocker = false
	default:
		stack.HasDocker = fileExists(filepath.Join(dir, "Dockerfile")) ||
			fileExists(filepath.Join(dir, "docker", "Dockerfile"))
	}

	stack.HasK8s = cfg.Kubernetes ||
		dirExists(filepath.Join(dir, "k8s")) ||
		dirExists(filepath.Join(dir, "manifests")) ||
		dirExists(filepath.Join(dir, "helm")) ||
		fileExists(filepath.Join(dir, "Chart.yaml"))

	stack.HasSemgrep = dirExists(filepath.Join(dir, ".semgrep"))

	return stack, nil
}

func DetectComponents(cfg *config.Config, maxDepth int) ([]Component, error) {
	return scanDir(cfg.WorkspaceDir, cfg.WorkspaceDir, maxDepth, 0)
}

func detectLanguage(dir string) string {
	switch {
	case fileExists(filepath.Join(dir, "go.mod")):
		return "go"
	case fileExists(filepath.Join(dir, "package.json")):
		return "node"
	case fileExists(filepath.Join(dir, "requirements.txt")),
		fileExists(filepath.Join(dir, "pyproject.toml")),
		fileExists(filepath.Join(dir, "setup.py")):
		return "python"
	case fileExists(filepath.Join(dir, "pom.xml")),
		fileExists(filepath.Join(dir, "build.gradle")),
		fileExists(filepath.Join(dir, "build.gradle.kts")):
		return "java"
	case fileExists(filepath.Join(dir, "Cargo.toml")):
		return "rust"
	default:
		return "unknown"
	}
}

func detectBuildTool(dir, language string) string {
	switch language {
	case "node":
		switch {
		case fileExists(filepath.Join(dir, "yarn.lock")):
			return "yarn"
		case fileExists(filepath.Join(dir, "pnpm-lock.yaml")):
			return "pnpm"
		default:
			return "npm"
		}
	case "java":
		if fileExists(filepath.Join(dir, "pom.xml")) {
			return "mvn"
		}
		return "gradle"
	case "python":
		if fileExists(filepath.Join(dir, "pyproject.toml")) {
			return "poetry"
		}
		return "pip"
	case "go":
		return "go"
	case "rust":
		return "cargo"
	default:
		return ""
	}
}

func fileExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && !info.IsDir()
}

func dirExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && info.IsDir()
}

var excludeDirs = map[string]bool{
	".git":            true,
	"node_modules":    true,
	"vendor":          true,
	"bin":             true,
	"dist":            true,
	"build":           true,
	"target":          true,
	"testdata":        true,
	".idea":           true,
	".vscode":         true,
	"migrations":      true,
	"monitoring":      true,
	"docs":            true,
	"infrastructure":  true,
	"shared-protos":   true,
	"scripts":         true,
}

func scanDir(root, dir string, maxDepth int, currentDepth int) ([]Component, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	var components []Component
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		if excludeDirs[entry.Name()] {
			continue
		}

		if currentDepth >= maxDepth {
			continue
		}

		subDir := filepath.Join(dir, entry.Name())

		language := detectLanguage(subDir)
		if language != "unknown" {
			components = append(components, Component{
				Path:      subDir,
				Language:  language,
				BuildTool: detectBuildTool(subDir, language),
				HasDocker: fileExists(filepath.Join(subDir, "Dockerfile")),
				HasK8s:    dirExists(filepath.Join(subDir, "k8s")) || dirExists(filepath.Join(subDir, "manifests")),
			})
		}

		if currentDepth < maxDepth {
			subComponents, err := scanDir(root, subDir, maxDepth, currentDepth+1)
			if err != nil {
				return nil, err
			}
			components = append(components, subComponents...)
		}

	}
	return components, nil
}
