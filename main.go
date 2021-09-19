package main

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/cszatmary/publisher/internal/file"
	"github.com/cszatmary/publisher/internal/git"
	"github.com/cszatmary/publisher/internal/log"
	"github.com/sc-lang/go-sc"
	flag "github.com/spf13/pflag"
)

type config struct {
	CommitMessage string                      `sc:"message"`
	ExcludedFiles []string                    `sc:"exclude"`
	Files         []string                    `sc:"files"`
	PreRunScript  string                      `sc:"preRun"`
	Targets       map[string]deploymentTarget `sc:"targets"`
}

type deploymentTarget struct {
	Branch     string `sc:"branch"`
	GithubRepo string `sc:"repo"`
	CustomURL  string `sc:"url"`
}

func main() {
	if err := execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func execute() error {
	flags := flag.NewFlagSet("publisher", flag.ExitOnError)
	configPath := flags.String("path", "publisher.sc", "The path to the publisher.sc config file.")
	skipPreRun := flags.Bool("skip-prerun", false, "Skip preRun step.")
	tag := flags.String("tag", "", "The git tag to create. Omit if you do not want to create a tag.")
	verbose := flags.BoolP("verbose", "v", false, "Enables verbose logging.")
	flags.Usage = func() {
		var sb strings.Builder
		sb.WriteString(`publisher is a small CLI for publishing static sites to GitHub Pages.

Usage:
  publisher [target]

Flags:
`)
		sb.WriteString(flags.FlagUsages())
		fmt.Fprint(os.Stderr, sb.String())
	}

	// Ignore error because it is set to ExitOnError
	_ = flags.Parse(os.Args[1:])
	if flags.NArg() == 0 {
		fmt.Fprint(os.Stderr, "Error: No target specified\n\n")
		flags.Usage()
		os.Exit(1)
	}

	targetName := flags.Arg(0)
	logger := log.New(os.Stderr)
	logger.SetDebug(*verbose)

	cacheDir, err := os.UserCacheDir()
	if err != nil {
		return fmt.Errorf("unable to get user's cache directory: %w", err)
	}
	reposDir := filepath.Join(cacheDir, "publisher", "repos")
	if err := os.MkdirAll(reposDir, 0o755); err != nil {
		return fmt.Errorf("failed to create directory %q: %w", reposDir, err)
	}

	srcRootPath, err := git.RootDir()
	if err != nil {
		return fmt.Errorf("failed to get root directory of git repo: %w", err)
	}
	sha, err := git.SHA("HEAD")
	if err != nil {
		return fmt.Errorf("failed to get SHA of HEAD for project: %w", err)
	}

	logger.Debugf("Reading %s config", *configPath)
	data, err := os.ReadFile(*configPath)
	if errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("config file not found at %q", *configPath)
	}
	if err != nil {
		return fmt.Errorf("failed to read %q: %w", *configPath, err)
	}
	vars := sc.MustVariables(map[string]string{
		"SHA":  sha,
		"TAG":  *tag,
		"DATE": time.Now().Local().Format("01-02-2006"),
	})
	var conf config
	err = sc.Unmarshal(data, &conf, sc.WithVariables(vars))
	if err != nil {
		return fmt.Errorf("failed to parse config file: %w", err)
	}

	target, ok := conf.Targets[targetName]
	if !ok {
		return fmt.Errorf("%s is not a valid deployment target", targetName)
	}

	// Setup target repo
	targetRepoPath := filepath.Join(reposDir, target.GithubRepo)
	repo, err := git.Prepare(target.GithubRepo, targetRepoPath, target.Branch, logger)
	if err != nil {
		return fmt.Errorf("failed to prepare target git repo: %w", err)
	}

	// Empty target repo
	logger.Debugf("Emptying directory %s", targetRepoPath)
	contents, err := os.ReadDir(targetRepoPath)
	if err != nil {
		return fmt.Errorf("failed to read contents of directory %q: %w", targetRepoPath, err)
	}
	for _, item := range contents {
		// Don't remove .git dir
		if item.Name() == ".git" && item.IsDir() {
			continue
		}
		p := filepath.Join(targetRepoPath, item.Name())
		if err := os.RemoveAll(p); err != nil {
			return fmt.Errorf("failed to remove %q: %w", p, err)
		}
	}

	if !*skipPreRun && conf.PreRunScript != "" {
		logger.Printf("Executing preRun script...")
		args := strings.Split(conf.PreRunScript, " ")
		cmd := exec.Command(args[0], args[1:]...)
		cmd.Dir = srcRootPath
		var stderr bytes.Buffer
		cmd.Stderr = &stderr
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("failed to execute preRun script, stderr: %s, error: %w", stderr.String(), err)
		}
	}

	// Copy files
	logger.Printf("Copying files...")
	var files []string
	for _, f := range conf.Files {
		matches, err := filepath.Glob(f)
		if err != nil {
			return fmt.Errorf("failed to parse files listed in config: %w", err)
		}
		files = append(files, matches...)
	}
	for _, f := range files {
		logger.Debugf("Copying %s...", f)
		srcPath := filepath.Join(srcRootPath, f)
		var dstFile string
		components := strings.Split(f, "/")
		if len(components) == 1 {
			dstFile = f
		} else {
			dstFile = filepath.Join(components[1:]...)
		}

		dstPath := filepath.Join(targetRepoPath, dstFile)
		if err := file.Copy(srcPath, dstPath); err != nil {
			return fmt.Errorf("failed to copy %q to %q: %w", srcPath, dstFile, err)
		}
	}

	if target.CustomURL != "" {
		cnamePath := filepath.Join(targetRepoPath, "CNAME")
		err := os.WriteFile(cnamePath, []byte(target.CustomURL), 0o644)
		if err != nil {
			return fmt.Errorf("failed to write CNAME file to target repo %s: %w", targetRepoPath, err)
		}
	}

	logger.Debugf("Committing files...")
	if err := repo.CommitChanges(conf.CommitMessage); err != nil {
		return fmt.Errorf("failed to commit files in target repo %q: %w", targetRepoPath, err)
	}
	logger.Printf("Pushing to branch %s in repo %s", target.Branch, target.GithubRepo)
	if err := repo.Push(); err != nil {
		return fmt.Errorf("failed to push changes to GitHub for target repo %q: %w", targetRepoPath, err)
	}
	logger.Printf("Successfully published to GitHub Pages! Enjoy!")
	return nil
}
