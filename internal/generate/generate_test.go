package generate

import (
	"strings"
	"testing"

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
	if len(files) != 2 {
		t.Errorf("expected 2 files, got %d", len(files))
	}
}

func TestGenerateNodeWithDocker(t *testing.T) {
	files, err := Generate(stack("node", "npm", true, false, true, true, true, "codeql"))
	if err != nil {
		t.Fatal(err)
	}
	if len(files) != 3 {
		t.Errorf("expected 3 files, got %d", len(files))
	}
}

func TestGenerateWithK8s(t *testing.T) {
	files, err := Generate(stack("go", "go", true, true, true, true, true, "codeql"))
	if err != nil {
		t.Fatal(err)
	}
	if len(files) != 4 {
		t.Errorf("expected 4 files, got %d", len(files))
	}
}

func TestGenerateUnknownLang(t *testing.T) {
	files, err := Generate(stack("unknown", "", false, false, true, true, false, ""))
	if err != nil {
		t.Fatal(err)
	}
	if len(files) != 1 {
		t.Errorf("expected 1 file (security only), got %d", len(files))
	}
}

func TestGitleaksInSecurity(t *testing.T) {
	files, err := Generate(stack("go", "go", false, false, false, true, false, ""))
	if err != nil {
		t.Fatal(err)
	}
	content := string(files[0].Content)
	if !strings.Contains(content, "gitleaks") {
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
