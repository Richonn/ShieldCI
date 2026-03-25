package pr

import (
	"testing"

	"github.com/Richonn/shieldci/internal/detect"
	"github.com/Richonn/shieldci/internal/generate"
)

func TestBuildResult(t *testing.T) {
	stack := &detect.StackConfig{
		Language:  "go",
		HasDocker: true,
		HasK8s:    false,
	}
	files := []generate.GeneratedFile{
		{Path: "ci.yml", Content: []byte("content")},
		{Path: "docker.yml", Content: []byte("content")},
	}

	result := buildResult("https://github.com/owner/repo/pull/1", stack, files)

	if result.PRURL != "https://github.com/owner/repo/pull/1" {
		t.Errorf("unexpected PRURL: %s", result.PRURL)
	}
	if result.StackJSON != `{"language":"go","docker":true,"k8s":false}` {
		t.Errorf("unexpected StackJSON: %s", result.StackJSON)
	}
	if result.FilesList == "" {
		t.Error("expected non-empty FilesList")
	}
}

func TestBuildResultK8s(t *testing.T) {
	stack := &detect.StackConfig{
		Language:  "python",
		HasDocker: true,
		HasK8s:    true,
	}
	files := []generate.GeneratedFile{
		{Path: "security.yml"},
		{Path: "ci.yml"},
		{Path: "docker.yml"},
		{Path: "k8s-deploy.yml"},
	}

	result := buildResult("https://github.com/owner/repo/pull/2", stack, files)

	if result.StackJSON != `{"language":"python","docker":true,"k8s":true}` {
		t.Errorf("unexpected StackJSON: %s", result.StackJSON)
	}
}
