package generate

import (
	"testing"

	"github.com/Richonn/shieldci/internal/detect"
)

func FuzzGenerate(f *testing.F) {
	f.Add("go", "go", false, false, true, true, true, "codeql")
	f.Add("node", "npm", true, false, true, true, true, "semgrep")
	f.Add("python", "pip", false, false, false, false, false, "")
	f.Add("java", "mvn", true, true, true, true, true, "codeql")
	f.Add("rust", "cargo", false, false, true, false, false, "")
	f.Add("unknown", "", false, false, false, false, false, "")
	f.Add("", "", false, false, false, false, false, "")
	f.Add("go", "go", true, true, false, false, false, "unknown-tool")

	f.Fuzz(func(t *testing.T, lang, buildTool string, docker, k8s, trivy, gitleaks, sast bool, sastTool string) {
		s := &detect.StackConfig{
			Language:       lang,
			BuildTool:      buildTool,
			HasDocker:      docker,
			HasK8s:         k8s,
			EnableTrivy:    trivy,
			EnableGitleaks: gitleaks,
			EnableSAST:     sast,
			SASTTool:       sastTool,
		}
		_, _ = Generate(s)
	})
}
