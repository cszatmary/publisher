package git

import (
	"fmt"
	"strings"
	"time"

	"github.com/cszatma/publisher/util"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/plumbing"
	"gopkg.in/src-d/go-git.v4/plumbing/object"
)

type Repository = git.Repository

func RootDir() (string, error) {
	path, err := util.ExecOutput("git", "rev-parse", "--show-toplevel")
	return strings.TrimSpace(path), errors.Wrapf(err, "Failed to get root dir of git repo")
}

func SHA(ref string) (string, error) {
	sha, err := util.ExecOutput("git", "rev-parse", ref)
	return strings.TrimSpace(sha), errors.Wrapf(err, "Failed to get SHA of ref %s", ref)
}

func Clone(name, branch, path string) (*git.Repository, error) {
	log.Debugf("Cloning repo %s to %s\n", name, path)
	repo, err := git.PlainClone(path, false, &git.CloneOptions{
		URL:           fmt.Sprintf("git@github.com:%s.git", name),
		ReferenceName: plumbing.NewBranchReferenceName(branch),
		SingleBranch:  true,
	})
	if err != nil {
		return nil, errors.Wrapf(err, "failed to clone %s to %s", name, path)
	}

	return repo, nil
}

func Open(name, branch, path string) (*git.Repository, error) {
	log.Debugf("Opening repo %s at path %s\n", name, path)
	repo, err := git.PlainOpen(path)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to open repo at path %s", path)
	}

	wt, err := repo.Worktree()
	if err != nil {
		return nil, errors.Wrapf(err, "failed to get worktree for repo %s", name)
	}

	log.Debugf("Cleaning %s\n", name)
	err = wt.Clean(&git.CleanOptions{
		Dir: true,
	})
	if err != nil {
		return nil, errors.Wrapf(err, "failed to clean repo %s", name)
	}

	log.Debugf("Checkout out branch %s in %s", branch, name)
	err = wt.Checkout(&git.CheckoutOptions{
		Branch: plumbing.NewBranchReferenceName(branch),
		Force:  true,
	})
	if err != nil {
		return nil, errors.Wrapf(err, "failed to checkout branch %s in repo %s", branch, name)
	}

	log.Debugf("Pulling changes from remote for %s", name)
	err = wt.Pull(&git.PullOptions{
		SingleBranch: true,
		Force:        true,
	})
	if err != nil && err != git.NoErrAlreadyUpToDate {
		return nil, errors.Wrapf(err, "failed to pull changes from remote for repo %s", name)
	}

	return repo, nil
}

func Add(name, path, addPath string) error {
	err := util.Exec("git", path, "add", addPath)
	return errors.Wrapf(err, "failed to add %s in repo %s", addPath, path)
}

func User() (name string, email string, err error) {
	args := []string{"config", "--get", "--global"}
	nameOutput, err := util.ExecOutput("git", append(args, "user.name")...)
	if err != nil {
		return "", "", errors.Wrap(err, "failed to get git user name")
	}

	emailOutput, err := util.ExecOutput("git", append(args, "user.email")...)
	if err != nil {
		return "", "", errors.Wrap(err, "failed to get git user email")
	}

	return strings.TrimSpace(nameOutput), strings.TrimSpace(emailOutput), nil
}

func Commit(name, msg string, repo *Repository) error {
	name, email, err := User()
	if err != nil {
		return errors.Wrap(err, "failed to get git user info")
	}

	wt, err := repo.Worktree()
	if err != nil {
		return errors.Wrapf(err, "failed to get worktree for repo %s", name)
	}

	_, err = wt.Commit(msg, &git.CommitOptions{
		Author: &object.Signature{
			Name:  name,
			Email: email,
			When:  time.Now(),
		},
	})

	return errors.Wrapf(err, "failed to commit files in repo %s", name)
}

func Push(name string, repo *Repository) error {
	err := repo.Push(&git.PushOptions{
		RemoteName: "origin",
	})

	return errors.Wrapf(err, "failed to push to remote in repo %s", name)
}
