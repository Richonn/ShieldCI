package generate

import (
	"strings"
	"testing"

	"github.com/Richonn/shieldci/internal/config"
	"github.com/Richonn/shieldci/internal/detect"
)

func stack(lang, buildTool string, docker, k8s, trivy, gitleaks, sast bool, sastTool string) *detect.StackConfig {
	return &detect.StackConfig{
		Language:       lang,
		BuildTool:      buildTool,
		HasDocker:      docker,
		HasK8s:         k8s,
		EnableTrivy:    trivy,
		EnableGitleaks: gitleaks,
		EnableSAST:     sast,
		SASTTool:       sastTool,
	}
}

func TestGenerateGoStack(t *testing.T) {
	files, err := Generate(stack("go", "go", false, false, true, true, true, "codeql"))
	if err != nil {
		t.Fatal(err)
	}
	if len(files) != 5 {
		t.Errorf("expected 5 files, got %d", len(files))
	}
}

func TestGenerateNodeWithDocker(t *testing.T) {
	files, err := Generate(stack("node", "npm", true, false, true, true, true, "codeql"))
	if err != nil {
		t.Fatal(err)
	}
	if len(files) != 7 {
		t.Errorf("expected 7 files, got %d", len(files))
	}
}

func TestGenerateWithK8s(t *testing.T) {
	files, err := Generate(stack("go", "go", true, true, true, true, true, "codeql"))
	if err != nil {
		t.Fatal(err)
	}
	if len(files) != 8 {
		t.Errorf("expected 8 files, got %d", len(files))
	}
}

func TestGenerateUnknownLang(t *testing.T) {
	files, err := Generate(stack("unknown", "", false, false, true, true, false, ""))
	if err != nil {
		t.Fatal(err)
	}
	if len(files) != 3 {
		t.Errorf("expected 3 files, got %d", len(files))
	}
}

func TestGitleaksInSecurity(t *testing.T) {
	files, err := Generate(stack("go", "go", false, false, false, true, false, ""))
	if err != nil {
		t.Fatal(err)
	}
	var securityFile string
	for _, f := range files {
		if f.Path == "security.yml" {
			securityFile = string(f.Content)
		}
	}
	if !strings.Contains(securityFile, "gitleaks") {
		t.Error("expected gitleaks job in security.yml")
	}
}

func TestTrivyInDocker(t *testing.T) {
	files, err := Generate(stack("go", "go", true, false, true, false, false, ""))
	if err != nil {
		t.Fatal(err)
	}
	var dockerFile string
	for _, f := range files {
		if f.Path == "docker.yml" {
			dockerFile = string(f.Content)
		}
	}
	if !strings.Contains(dockerFile, "trivy") {
		t.Error("expected trivy job in docker.yml")
	}
}

func TestPRBodyGo(t *testing.T) {
	s := stack("go", "go", false, false, true, true, true, "codeql")
	files := []GeneratedFile{
		{Path: "ci.yml"},
		{Path: "security.yml"},
		{Path: "lint.yml"},
		{Path: "test.yml"},
	}
	body := PRBody(s, files)
	if !strings.Contains(body, "go") {
		t.Error("expected language in PR body")
	}
	if !strings.Contains(body, "ci.yml") {
		t.Error("expected ci.yml in PR body")
	}
}

func TestPRBodyWithDocker(t *testing.T) {
	s := stack("node", "npm", true, false, true, true, true, "codeql")
	files := []GeneratedFile{
		{Path: "ci.yml"},
		{Path: "security.yml"},
		{Path: "lint.yml"},
		{Path: "test.yml"},
		{Path: "docker.yml"},
	}
	body := PRBody(s, files)
	if !strings.Contains(body, "Docker") {
		t.Error("expected Docker in PR body")
	}
}

func TestGenerateMonorepoNoDuplicateNames(t *testing.T) {
	components := []detect.Component{
		{Path: "/workspace/services/api", Language: "go", BuildTool: "go"},
		{Path: "/workspace/tools/api", Language: "node", BuildTool: "npm"},
		{Path: "/workspace/backend/services/api", Language: "go", BuildTool: "go"},
		{Path: "/workspace/frontend/services/api", Language: "node", BuildTool: "npm"},
	}
	cfg := &config.Config{
		WorkspaceDir:   "/workspace",
		EnableTrivy:    true,
		EnableGitleaks: true,
		EnableSAST:     true,
		SASTTool:       "codeql",
	}

	files, err := GenerateMonorepo(components, cfg)
	if err != nil {
		t.Fatal(err)
	}

	seen := make(map[string]bool)
	for _, f := range files {
		if seen[f.Path] {
			t.Errorf("duplicate file path: %s", f.Path)
		}
		seen[f.Path] = true
	}
}

func TestGenerateMonorepoFilePrefix(t *testing.T) {
	components := []detect.Component{
		{Path: "/workspace/services/api", Language: "go", BuildTool: "go"},
	}
	cfg := &config.Config{
		WorkspaceDir:   "/workspace",
		EnableTrivy:    true,
		EnableGitleaks: true,
		EnableSAST:     true,
		SASTTool:       "codeql",
	}

	files, err := GenerateMonorepo(components, cfg)
	if err != nil {
		t.Fatal(err)
	}

	for _, f := range files {
		if !strings.HasPrefix(f.Path, "services-api-") {
			t.Errorf("expected prefix 'services-api-', got: %s", f.Path)
		}
	}
}

func TestPRBodyWithK8s(t *testing.T) {
	s := stack("go", "go", true, true, true, true, true, "codeql")
	files := []GeneratedFile{
		{Path: "ci.yml"},
		{Path: "security.yml"},
		{Path: "lint.yml"},
		{Path: "test.yml"},
		{Path: "docker.yml"},
		{Path: "k8s-deploy.yml"},
	}
	body := PRBody(s, files)
	if !strings.Contains(body, "Kubernetes") {
		t.Error("expected Kubernetes in PR body")
	}
}
