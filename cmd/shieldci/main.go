package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/Richonn/shieldci/internal/config"
	"github.com/Richonn/shieldci/internal/detect"
	"github.com/Richonn/shieldci/internal/generate"
	"github.com/Richonn/shieldci/internal/pr"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "shieldci: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	ctx := context.Background()

	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}
	log.Printf("repo: %s/%s | language: %s", cfg.RepoOwner, cfg.RepoName, cfg.Language)

	stack, err := detect.Detect(cfg)
	if err != nil {
		return fmt.Errorf("detect stack: %w", err)
	}
	log.Printf("detected: language=%s docker=%v k8s=%v buildTool=%s",
		stack.Language, stack.HasDocker, stack.HasK8s, stack.BuildTool)

	files, err := generate.Generate(stack)
	if err != nil {
		return fmt.Errorf("generate workflows: %w", err)
	}
	log.Printf("generated %d workflows file(s)", len(files))

	if cfg.DryRun {
		if err := dryRun(cfg, files); err != nil {
			return fmt.Errorf("dry-run detection: %w", err)
		}
		return nil
	}

	body := generate.PRBody(stack, files)

	result, err := pr.CreateOrUpdatePR(ctx, cfg, stack, files, body)
	if err != nil {
		return fmt.Errorf("create PR: %w", err)
	}
	log.Printf("PR: %s", result.PRURL)

	if err := cfg.WriteOutput("pr-url", result.PRURL); err != nil {
		return fmt.Errorf("write output pr-url: %w", err)
	}
	if err := cfg.WriteOutput("detected-stack", result.StackJSON); err != nil {
		return fmt.Errorf("write output detected-stack: %w", err)
	}
	if err := cfg.WriteOutput("generated-files", result.FilesList); err != nil {
		return fmt.Errorf("write output generated-files: %w", err)
	}

	return nil
}

func dryRun(cfg *config.Config, files []generate.GeneratedFile) error {
	var summary string
	for _, f := range files {
		summary += "## "
		summary += f.Path
		summary += "\n```yaml\n"
		summary += string(f.Content)
		summary += "```\n\n"
	}
	if err := cfg.WriteSummary(summary); err != nil {
		return fmt.Errorf("write summary: %w", err)
	}
	return nil
}
