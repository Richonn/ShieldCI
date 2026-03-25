package detect

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/Richonn/shieldci/internal/config"
)

func cfg(dir, language, docker string, k8s bool) *config.Config {
	return &config.Config{
		WorkspaceDir: dir,
		Language: language,
		Docker: docker,
		Kubernetes: k8s,
	}
}

func touch(t *testing.T, path string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte{}, 0644); err != nil {
		t.Fatal(err)
	}
}

func TestDetectGo(t *testing.T) {
	dir := t.TempDir()
	touch(t, filepath.Join(dir, "go.mod"))
	stack, err := Detect(cfg(dir, "auto", "auto", false))
	if err != nil { t.Fatal(err) }
	if stack.Language != "go" { t.Errorf("got %q, want go", stack.Language) }
	if stack.BuildTool != "go" { t.Errorf("got %q, want go", stack.BuildTool) }
}

func TestDetectNode(t *testing.T) {
	dir := t.TempDir()
	touch(t, filepath.Join(dir, "package.json"))
	stack, err := Detect(cfg(dir, "auto", "auto", false))
	if err != nil { t.Fatal(err) }
	if stack.Language != "node" { t.Errorf("got %q, want node", stack.Language) }
	if stack.BuildTool != "npm" { t.Errorf("got %q, want npm", stack.BuildTool) }
}

func TestDetectNodeYarn(t *testing.T) {
	dir := t.TempDir()
	touch(t, filepath.Join(dir, "package.json"))
	touch(t, filepath.Join(dir, "yarn.lock"))
	stack, _ := Detect(cfg(dir, "auto", "auto", false))
	if stack.BuildTool != "yarn" { t.Errorf("got %q, want yarn", stack.BuildTool) }
}

func TestDetectPython(t *testing.T) {
	dir := t.TempDir()
	touch(t, filepath.Join(dir, "requirements.txt"))
	stack, _ := Detect(cfg(dir, "auto", "auto", false))
	if stack.Language != "python" { t.Errorf("got %q, want python", stack.Language) }
	if stack.BuildTool != "pip" { t.Errorf("got %q, want pip", stack.BuildTool) }
}

func TestDetectJava(t *testing.T) {
	dir := t.TempDir()
	touch(t, filepath.Join(dir, "pom.xml"))
	stack, _ := Detect(cfg(dir, "auto", "auto", false))
	if stack.Language != "java" { t.Errorf("got %q, want java", stack.Language) }
	if stack.BuildTool != "mvn" { t.Errorf("got %q, want mvn", stack.BuildTool) }
}

func TestDetectRust(t *testing.T) {
	dir := t.TempDir()
	touch(t, filepath.Join(dir, "Cargo.toml"))
	stack, _ := Detect(cfg(dir, "auto", "auto", false))
	if stack.Language != "rust" { t.Errorf("got %q, want rust", stack.Language) }
}

func TestLanguageOverride(t *testing.T) {
	dir := t.TempDir()
	touch(t, filepath.Join(dir, "package.json"))
	stack, _ := Detect(cfg(dir, "python", "auto", false))
	if stack.Language != "python" { t.Errorf("override failed, got %q", stack.Language) }
}

func TestGoWinsOverNode(t *testing.T) {
	dir := t.TempDir()
	touch(t, filepath.Join(dir, "go.mod"))
	touch(t, filepath.Join(dir, "package.json"))
	stack, _ := Detect(cfg(dir, "auto", "auto", false))
	if stack.Language != "go" { t.Errorf("got %q, want go", stack.Language) }
}

func TestDockerAuto(t *testing.T) {
	dir := t.TempDir()
	touch(t, filepath.Join(dir, "Dockerfile"))
	stack, _ := Detect(cfg(dir, "auto", "auto", false))
	if !stack.HasDocker { t.Error("expected HasDocker = true") }
}

func TestDockerForced(t *testing.T) {
	dir := t.TempDir()
	stack, _ := Detect(cfg(dir, "auto", "true", false))
	if !stack.HasDocker { t.Error("expected HasDocker = true") }
}

func TestDockerDisabled(t *testing.T) {
	dir := t.TempDir()
	touch(t, filepath.Join(dir, "Dockerfile"))
	stack, _ := Detect(cfg(dir, "auto", "false", false))
	if stack.HasDocker { t.Error("expected HasDocker = false") }
}

func TestK8sDir(t *testing.T) {
	dir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(dir, "k8s"), 0755); err != nil {
		t.Fatal(err)
	}
	stack, _ := Detect(cfg(dir, "auto", "auto", false))
	if !stack.HasK8s { t.Error("expected HasK8s = true") }
}

func TestK8sForced(t *testing.T) {
	dir := t.TempDir()
	stack, _ := Detect(cfg(dir, "auto", "auto", true))
	if !stack.HasK8s { t.Error("expected HasK8s = true") }
}
