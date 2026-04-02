package detect

import (
	"testing"

	"github.com/Richonn/shieldci/internal/config"
)

func FuzzDetect(f *testing.F) {
	f.Add("auto", "auto", false)
	f.Add("go", "true", true)
	f.Add("node", "false", false)
	f.Add("python", "auto", false)
	f.Add("java", "true", false)
	f.Add("rust", "auto", true)
	f.Add("", "", false)
	f.Add("unknown-lang", "maybe", false)

	f.Fuzz(func(t *testing.T, language, docker string, k8s bool) {
		dir := t.TempDir()
		cfg := &config.Config{
			WorkspaceDir: dir,
			Language:     language,
			Docker:       docker,
			Kubernetes:   k8s,
		}
		_, _ = Detect(cfg)
	})
}

func FuzzDetectComponents(f *testing.F) {
	f.Add(0)
	f.Add(1)
	f.Add(3)
	f.Add(10)
	f.Add(-1)

	f.Fuzz(func(t *testing.T, maxDepth int) {
		dir := t.TempDir()
		cfg := &config.Config{
			WorkspaceDir: dir,
			Language:     "auto",
			Docker:       "auto",
		}
		_, _ = DetectComponents(cfg, maxDepth)
	})
}
