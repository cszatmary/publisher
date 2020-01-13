package main

import (
	"fmt"
	"time"

	"github.com/cszatma/publisher/config"
	"github.com/cszatma/publisher/fatal"
	"github.com/cszatma/publisher/git"
	"github.com/cszatma/publisher/util"
	flag "github.com/spf13/pflag"
)

var (
	configPath string
	tag        string
	targetName string
	verbose    bool
)

func parseFlags() {
	flag.StringVarP(&configPath, "path", "p", "publisher.yml", "The path to the publisher.yml config file.")
	flag.StringVarP(&tag, "tag", "t", "", "The git tag to create. Omit if you do not want to create a tag.")
	flag.StringVarP(&targetName, "target", "T", "", "The target to deploy.")
	flag.BoolVarP(&verbose, "verbose", "v", false, "Enables verbose logging.")

	flag.Parse()

	if targetName == "" {
		fatal.Exit("Must provider a target to deploy using the --target flag")
	}
}

func main() {
	parseFlags()

	util.SetVerboseMode(verbose)
	fatal.ShowStackTraces(verbose)

	sourceRootPath, err := git.RootDir()
	if err != nil {
		fatal.ExitErr(err, "Project is not a git repo")
	}

	sha, err := git.SHA("HEAD")
	if err != nil {
		fatal.ExitErr(err, "Failed to get SHA of HEAD for project")
	}

	vars := map[string]string{
		"SHA":  sha,
		"TAG":  tag,
		"DATE": time.Now().Local().Format("01-02-2006"),
	}

	util.VerbosePrintln("Reading publisher.yml config")
	conf, err := config.Init(configPath, vars)
	if err != nil {
		fatal.ExitErr(err, "Failed to read config file")
	}

	target, ok := conf.Targets[targetName]
	if !ok {
		fatal.Exitf("%s is not a valid deployment target", targetName)
	}

	targetRepoPath := fmt.Sprintf("%s/%s", config.ReposDir(), target.GithubRepo)
	var repo *git.Repository
	if !util.FileOrDirExists(targetRepoPath) {
		util.VerbosePrintf("Target repo %s does not exist, cloning...")

		repo, err = git.Clone(target.GithubRepo, target.Branch, targetRepoPath)
		if err != nil {
			fatal.ExitErrf(err, "Failed to clone repo %s to %s", target.GithubRepo, targetRepoPath)
		}

		util.VerbosePrintf("Successfully cloned repo %s", target.GithubRepo)
	} else {
		util.VerbosePrintf("Target repo %s exists, opening and setting up")

		repo, err = git.Open(target.GithubRepo, target.Branch, targetRepoPath)
		if err != nil {
			fatal.ExitErrf(err, "Failed to open repo %s at path %s", target.GithubRepo, targetRepoPath)
		}

		util.VerbosePrintf("Successfully opened repo %s", target.GithubRepo)
	}
}
