package config

import (
	"os"
	"path/filepath"

	"github.com/cszatma/publisher/util"
	"github.com/pkg/errors"
)

type DeploymentTarget struct {
	Branch     string `yaml:"branch"`
	GithubRepo string `yaml:"repo"`
	CustomURL  string `yaml:"url"`
}

type PublisherConfig struct {
	CommitMessage string                      `yaml:"message"`
	ExcludedFiles []string                    `yaml:"exclude"`
	Files         []string                    `yaml:"files"`
	Targets       map[string]DeploymentTarget `yaml:"targets"`
}

var (
	config       PublisherConfig
	publisherDir string
)

func Config() *PublisherConfig {
	return &config
}

func PublisherDir() string {
	return publisherDir
}

func ReposDir() string {
	return filepath.Join(publisherDir, "repos")
}

func Init(configPath string, vars map[string]string) (*PublisherConfig, error) {
	publisherDir = filepath.Join(os.Getenv("HOME"), ".publisher")

	if !util.FileOrDirExists(publisherDir) {
		err := os.Mkdir(publisherDir, 0755)
		if err != nil {
			return nil, errors.Wrapf(err, "Failed to create publisher directory at %s", publisherDir)
		}
	}

	if !util.FileOrDirExists(configPath) {
		return nil, errors.Errorf("No such file %s", configPath)
	}

	err := util.ReadYaml(configPath, &config)
	if err != nil {
		return nil, errors.Wrapf(err, "Failed to read yaml file at path %s", configPath)
	}

	config.CommitMessage = util.ExpandVars(config.CommitMessage, vars)
	return &config, nil
}
