package scanner

import (
	"encoding/json"
	"repo-scanner/internal"
	"repo-scanner/internal/utils/serror"
	"repo-scanner/internal/utils/utstring"
	"strings"

	log "github.com/sirupsen/logrus"

	"github.com/grab/secret-scanner/scanner"
	"github.com/grab/secret-scanner/scanner/gitprovider"
	"github.com/grab/secret-scanner/scanner/options"
	"github.com/grab/secret-scanner/scanner/session"
)

type grabScanner struct {
	github    *gitprovider.GithubProvider
	gitlab    *gitprovider.GitlabProvider
	bitbucket *gitprovider.BitbucketProvider
}

func NewGrabScanner(store internal.RepositoryStore) internal.IGrabScanner {
	grabScanner := grabScanner{}
	additionalParams := map[string]string{
		gitprovider.BitbucketParamClientID:     utstring.Env(gitprovider.BitbucketParamClientID),
		gitprovider.BitbucketParamClientSecret: utstring.Env(gitprovider.BitbucketParamClientSecret),
		gitprovider.BitbucketParamUsername:     utstring.Env(gitprovider.BitbucketParamUsername),
		gitprovider.BitbucketParamPassword:     utstring.Env(gitprovider.BitbucketParamPassword),
	}

	// Initialize Github provider
	github := &gitprovider.GithubProvider{}
	err := github.Initialize(utstring.Env(gitprovider.GithubParamBaseURL), utstring.Env(gitprovider.GithubParamToken), map[string]string{})
	if err != nil {
		log.Error("Unable to initialise Github provider")
	} else {
		grabScanner.github = github
	}

	// Initialize Gitlab provider
	gitlab := &gitprovider.GitlabProvider{}
	err = gitlab.Initialize(utstring.Env(gitprovider.GitlabParamBaseURL), utstring.Env(gitprovider.GitlabParamToken), map[string]string{})
	if err != nil {
		log.Error("Unable to initialise Gitlab provider")
	} else {
		grabScanner.gitlab = gitlab
	}

	// Initialize Bitbucket provider
	bitbucket := &gitprovider.BitbucketProvider{}
	err = bitbucket.Initialize(utstring.Env(gitprovider.BitbucketParamBaseURL), "", additionalParams)
	if err != nil {
		log.Error("Unable to initialise Bitbucket provider")
	} else {
		grabScanner.bitbucket = bitbucket
	}

	return grabScanner
}

func (g grabScanner) StartScanningSession(repo_url string) (res []byte, errx serror.SError) {
	pathParts := strings.Split(repo_url, "/")

	opt := options.Options{
		BaseURL:          new(string),
		CommitDepth:      new(int),
		Debug:            new(bool),
		EnvFilePath:      new(string),
		GitProvider:      new(string),
		Load:             new(string),
		LocalPath:        new(string),
		LogSecret:        new(bool),
		Report:           new(string),
		Repos:            new(string),
		ScanTarget:       new(string),
		Silent:           new(bool),
		SkipTestContexts: new(bool),
		State:            new(bool),
		Threads:          new(int),
		Token:            new(string),
		UI:               new(bool),
		UIHost:           new(string),
		UIPort:           new(string),
	}
	*opt.CommitDepth = 500
	*opt.Debug = false
	*opt.LogSecret = true
	*opt.Repos = strings.Join(pathParts[1:], "/")
	*opt.Silent = true
	*opt.SkipTestContexts = true
	*opt.State = false

	var gitProvider gitprovider.GitProvider
	switch pathParts[0] {
	case "github.com":
		if g.github == nil {
			res = []byte(`{"reason":"Github is not available for now"}`)
			errx = serror.New("Github is not available for now")
			errx.AddCommentf("[repository][StartScanningSession] Github is not available for now")
			return
		}
		*opt.GitProvider = "github"
		*opt.BaseURL = utstring.Env(gitprovider.GithubParamBaseURL)
		*opt.Token = utstring.Env(gitprovider.GithubParamToken)
		gitProvider = g.github
	case "gitlab.com":
		if g.gitlab == nil {
			res = []byte(`{"reason":"Gitlab is not available for now"}`)
			errx = serror.New("Gitlab is not available for now")
			errx.AddCommentf("[repository][StartScanningSession] Gitlab is not available for now")
			return
		}
		*opt.GitProvider = "gitlab"
		*opt.BaseURL = utstring.Env(gitprovider.GitlabParamBaseURL)
		*opt.Token = utstring.Env(gitprovider.GitlabParamToken)
		gitProvider = g.gitlab
	case "bitbucket.org":
		if g.gitlab == nil {
			res = []byte(`{"reason":"Bitbucket is not available for now"}`)
			errx = serror.New("Bitbucket is not available for now")
			errx.AddCommentf("[repository][StartScanningSession] Bitbucket is not available for now")
			return
		}
		*opt.GitProvider = "bitbucket"
		*opt.BaseURL = utstring.Env(gitprovider.BitbucketParamBaseURL)
		gitProvider = g.bitbucket
	}

	// Initialize new scan session
	sess := &session.Session{}
	sess.Initialize(opt)

	// Scan
	scanner.Scan(sess, gitProvider)

	// Return result
	if sess.Stats.Status == session.StatusFinished {
		sessionJSON, err := json.Marshal(sess.Findings)
		if err != nil {
			res = []byte(`{"reason":"Marshal struct failed"}`)
			errx = serror.NewFromError(err)
			errx.AddCommentf("[repository][StartScanningSession] while marshal struct")
			return
		}
		return sessionJSON, nil
	}
	return
}
