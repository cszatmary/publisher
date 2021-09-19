package git

import (
	"bytes"
	"errors"
	"fmt"
	"os/exec"
	"strings"
	"time"

	"github.com/cszatmary/publisher/internal/file"
	"github.com/cszatmary/publisher/internal/log"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
)

func RootDir() (string, error) {
	return execGit("rev-parse", "--show-toplevel")
}

func SHA(ref string) (string, error) {
	return execGit("rev-parse", ref)
}

type Repository struct {
	name string
	r    *git.Repository
	w    *git.Worktree
}

func Prepare(name, path, branch string, logger *log.Logger) (*Repository, error) {
	repo := &Repository{name: name}
	skipCleanup := false
	var err error
	if !file.Exists(path) {
		skipCleanup = true
		logger.Debugf("Cloning repo %s to %s", name, path)
		repo.r, err = git.PlainClone(path, false, &git.CloneOptions{
			URL:           fmt.Sprintf("git@github.com:%s.git", name),
			ReferenceName: plumbing.NewBranchReferenceName(branch),
			SingleBranch:  true,
		})
		if err != nil {
			return nil, fmt.Errorf("failed to clone %s to %s: %w", name, path, err)
		}
	} else {
		logger.Debugf("Opening repo %s at path %s", name, path)
		repo.r, err = git.PlainOpen(path)
		if err != nil {
			return nil, fmt.Errorf("failed to open repo at path %s", path)
		}
	}

	repo.w, err = repo.r.Worktree()
	if err != nil {
		return nil, fmt.Errorf("failed to get worktree for repo %s: %w", name, err)
	}
	if skipCleanup {
		return repo, nil
	}

	logger.Debugf("Cleaning %s", name)
	err = repo.w.Clean(&git.CleanOptions{Dir: true})
	if err != nil {
		return nil, fmt.Errorf("failed to clean repo %s: %w", name, err)
	}

	logger.Debugf("Checkout out branch %s in %s", branch, name)
	err = repo.w.Checkout(&git.CheckoutOptions{
		Branch: plumbing.NewBranchReferenceName(branch),
		Force:  true,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to checkout branch %s in repo %s: %w", branch, name, err)
	}

	logger.Debugf("Pulling changes from remote for %s", name)
	err = repo.w.Pull(&git.PullOptions{
		SingleBranch: true,
		Force:        true,
	})
	if err != nil && !errors.Is(err, git.NoErrAlreadyUpToDate) {
		return nil, fmt.Errorf("failed to pull changes from remote for repo %s: %w", name, err)
	}
	return repo, nil
}

func (repo *Repository) CommitChanges(msg string) error {
	// This doesn't detect deleted files: https://github.com/go-git/go-git/issues/113
	if err := repo.w.AddGlob("."); err != nil {
		return fmt.Errorf("failed to stage files: %w", err)
	}
	username, email, err := user()
	if err != nil {
		return err
	}
	_, err = repo.w.Commit(msg, &git.CommitOptions{
		// Will add deleted files
		All: true,
		Author: &object.Signature{
			Name:  username,
			Email: email,
			When:  time.Now(),
		},
	})
	if err != nil {
		return fmt.Errorf("failed to commit changes in repo %s: %w", repo.name, err)
	}
	return nil
}

func (repo *Repository) Push() error {
	err := repo.r.Push(&git.PushOptions{RemoteName: "origin"})
	if err != nil {
		return fmt.Errorf("failed to push to remote in repo %s: %w", repo.name, err)
	}
	return nil
}

func user() (username string, email string, err error) {
	username, err = execGit("config", "--get", "--global", "user.name")
	if err != nil {
		err = fmt.Errorf("failed to get git user name: %w", err)
		return
	}
	email, err = execGit("config", "--get", "--global", "user.email")
	if err != nil {
		err = fmt.Errorf("failed to get git user email: %w", err)
		return
	}
	return
}

func execGit(args ...string) (string, error) {
	var stdout, stderr bytes.Buffer
	cmd := exec.Command("git", args...)
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		argsStr := strings.Join(args, " ")
		return "", fmt.Errorf("failed to run 'git %s', stderr: %s, error: %w", argsStr, stderr.String(), err)
	}
	if stdout.Len() == 0 {
		return "", nil
	}
	return strings.TrimSpace(stdout.String()), nil
}
