// Copyright 2018 Drone.IO Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package github

import (
	"crypto/tls"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"

	// "github.com/sirupsen/logrus"
	"github.com/coderun-top/coderun/src/model"
	"github.com/coderun-top/coderun/src/remote"
	"github.com/coderun-top/coderun/src/shared/httputil"

	"github.com/google/go-github/github"
	"golang.org/x/net/context"
	"golang.org/x/oauth2"
)

const (
	defaultURL = "https://github.com"     // Default GitHub URL
	defaultAPI = "https://api.github.com/" // Default GitHub API URL
)

var (
	defaultScopes = []string{"user", "repo", "admin:org", "admin:org_hook", "admin:repo_hook"}
)

// Opts defines configuration options.
type Opts struct {
	URL         string // GitHub server url.
	Context     string // Context to display in status check
	Client      string // GitHub oauth client id.
	Secret      string // GitHub oauth client secret.
	Username    string // Optional machine account username.
	Password    string // Optional machine account password.
	PrivateMode bool   // GitHub is running in private mode.
	SkipVerify  bool   // Skip ssl verification.
	MergeRef    bool   // Clone pull requests using the merge ref.
}

// New returns a Remote implementation that integrates with a GitHub Cloud or
// GitHub Enterprise version control hosting provider.
func New(opts Opts) (remote.Remote, error) {
	url, err := url.Parse(opts.URL)
	if err != nil {
		return nil, err
	}
	host, _, err := net.SplitHostPort(url.Host)
	if err == nil {
		url.Host = host
	}
	remote := &client{
		API:         defaultAPI,
		URL:         defaultURL,
		Context:     opts.Context,
		Client:      opts.Client,
		Secret:      opts.Secret,
		PrivateMode: opts.PrivateMode,
		SkipVerify:  opts.SkipVerify,
		MergeRef:    opts.MergeRef,
		Machine:     url.Host,
		Username:    opts.Username,
		Password:    opts.Password,
	}
	if opts.URL != defaultURL {
		remote.URL = strings.TrimSuffix(opts.URL, "/")
		remote.API = remote.URL + "/api/v3/"
	}

	// Hack to enable oauth2 access in older GHE
	oauth2.RegisterBrokenAuthHeaderProvider(remote.URL)
	return remote, nil
}

type client struct {
	URL         string
	Context     string
	API         string
	Client      string
	Secret      string
	Machine     string
	Username    string
	Password    string
	PrivateMode bool
	SkipVerify  bool
	MergeRef    bool
}

// Login authenticates the session and returns the remote user details.
func (c *client) Login(res http.ResponseWriter, req *http.Request) (*model.User, error) {
	config := c.newConfig(httputil.GetURL(req))

	// get the OAuth errors
	if err := req.FormValue("error"); err != "" {
		return nil, &remote.AuthError{
			Err:         err,
			Description: req.FormValue("error_description"),
			URI:         req.FormValue("error_uri"),
		}
	}

	// get the OAuth code
	code := req.FormValue("code")
	if len(code) == 0 {
		// TODO(bradrydzewski) we really should be using a random value here and
		// storing in a cookie for verification in the next stage of the workflow.

		http.Redirect(res, req, config.AuthCodeURL("drone"), http.StatusSeeOther)
		return nil, nil
	}

	token, err := config.Exchange(c.newContext(), code)
	if err != nil {
		return nil, err
	}

	client, ctx := c.newClientToken(token.AccessToken)
	user, _, err := client.Users.Get(ctx, "")
	if err != nil {
		return nil, err
	}

	emails, _, err := client.Users.ListEmails(ctx, nil)
	if err != nil {
		return nil, err
	}
	email := matchingEmail(emails, c.API)
	if email == nil {
		return nil, fmt.Errorf("No verified Email address for GitHub account")
	}

	return &model.User{
		Login:  *user.Login,
		Email:  *email.Email,
		Token:  token.AccessToken,
		Avatar: *user.AvatarURL,
	}, nil
}

func (c *client) User(token string, t int) (*model.User, error) {
	client, ctx := c.newClientToken(token)
	user, _, err := client.Users.Get(ctx, "")
	if err != nil {
		return nil, err
	}

	emails, _, err := client.Users.ListEmails(ctx, nil)
	if err != nil {
		return nil, err
	}
	email := matchingEmail(emails, c.API)
	if email == nil {
		return nil, fmt.Errorf("No verified Email address for GitHub account")
	}

	return &model.User{
		Login:  *user.Login,
		Email:  *email.Email,
		Token:  token,
		Avatar: *user.AvatarURL,
	}, nil
}

func (c *client) AccessToken(token string) (string, error) {
	client, ctx := c.newClientToken(token)
	authReq := github.AuthorizationRequest{
		Note:   github.String("coderun"),
		Scopes: []github.Scope{"user", "repo", "admin:org", "admin:org_hook", "admin:repo_hook"},
	}
	auth, _, err := client.Authorizations.Create(ctx, &authReq)
	if err != nil {
		return "", err
	}

	return *auth.Token, nil
}

// Auth returns the GitHub user login for the given access token.
func (c *client) Auth(token, secret string) (string, error) {
	client, ctx := c.newClientToken(token)
	user, _, err := client.Users.Get(ctx, "")
	if err != nil {
		return "", err
	}
	return *user.Login, nil
}

// Teams returns a list of all team membership for the GitHub account.
func (c *client) Teams(u *model.User) ([]*model.Team, error) {
	client, ctx := c.newClientToken(u.Token)

	opts := new(github.ListOptions)
	opts.Page = 1

	var teams []*model.Team
	for opts.Page > 0 {
		list, resp, err := client.Organizations.List(ctx, "", opts)
		if err != nil {
			return nil, err
		}
		teams = append(teams, convertTeamList(list)...)
		opts.Page = resp.NextPage
	}
	return teams, nil
}

// Repo returns the named GitHub repository.
func (c *client) Repo(u *model.User, owner, name string) (*model.Repo, error) {
	client, ctx := c.newClientToken(u.Token)
	nn := strings.Split(name, "/")
	repo, _, err := client.Repositories.Get(ctx, nn[0], nn[1])
	if err != nil {
		return nil, err
	}
	return convertRepo(repo, c.PrivateMode), nil
}

// Repos returns a list of all repositories for GitHub account, including
// organization repositories.
func (c *client) Repos(u *model.User) ([]*model.Repo, error) {
	client, ctx := c.newClientToken(u.Token)

	opts := new(github.RepositoryListOptions)
	opts.PerPage = 100
	opts.Page = 1

	var repos []*model.Repo

	//if search == "" {
	for opts.Page > 0 {
		list, resp, err := client.Repositories.List(ctx, "", opts)
		if err != nil {
			return nil, err
		}
		repos = append(repos, convertRepoList(list, c.PrivateMode)...)
		opts.Page = resp.NextPage
	}
	//} else {
	//    for opts.Page > 0 {
	//	list, resp, err := client.Search.Repositories("user:"+u.Login+"+"+search, opts)
	//	if err != nil {
	//		return nil, err
	//	}
	//	repos = append(repos, convertRepoList(list, c.PrivateMode)...)
	//	opts.Page = resp.NextPage
	//    }
	//}
	return repos, nil
}

// Repos returns a list of all repositories for GitHub account, including
// organization repositories.
func (c *client) Branches(u *model.User, owner, name string) ([]*model.Branch, error) {
	client, ctx := c.newClientToken(u.Token)

	opts := new(github.ListOptions)
	opts.PerPage = 100
	opts.Page = 1
	nn := strings.Split(name, "/")

	var branches []*model.Branch
	for opts.Page > 0 {
		list, branche, err := client.Repositories.ListBranches(ctx, nn[0], nn[1], opts)
		if err != nil {
			return nil, err
		}
		branches = append(branches, convertBranchList(list)...)
		opts.Page = branche.NextPage
	}
	return branches, nil
}

func (c *client) Branch(u *model.User, owner, name, branchName string) (*model.Branch, error) {
	client, ctx := c.newClientToken(u.Token)
	nn := strings.Split(name, "/")
	branch, _, err := client.Repositories.GetBranch(ctx, nn[0], nn[1], branchName)
	if err != nil {
		return nil, err
	}
	commit, _, err := client.Repositories.GetCommit(ctx, nn[0], nn[1], *branch.Commit.SHA)
	if err != nil {
		return nil, err
	}
	dougoCommit := convertCommit(commit)
	dougoBranch := convertBranch(branch)
	dougoBranch.Commit = dougoCommit

	return dougoBranch, nil
}

// Perm returns the user permissions for the named GitHub repository.
func (c *client) Perm(u *model.User, owner, name string) (*model.Perm, error) {
	client, ctx := c.newClientToken(u.Token)
	repo, _, err := client.Repositories.Get(ctx, owner, name)
	if err != nil {
		return nil, err
	}
	return convertPerm(repo), nil
}

// File fetches the file from the GitHub repository and returns its contents.
func (c *client) File(u *model.User, r *model.Repo, b *model.Build, f string) ([]byte, error) {
	return c.FileRef(u, r, b.Commit, f)
}

// FileRef fetches the file from the GitHub repository and returns its contents.
func (c *client) FileRef(u *model.User, r *model.Repo, ref, f string) ([]byte, error) {
	client, ctx := c.newClientToken(u.Token)
	nn := strings.Split(r.FullName, "/")

	opts := new(github.RepositoryContentGetOptions)
	opts.Ref = ref
	data, _, _, err := client.Repositories.GetContents(ctx, nn[0], nn[1], f, opts)
	if err != nil {
		return nil, err
	}
	content, err := data.GetContent()
	if err != nil {
		return nil, err
	}
	return []byte(content), nil
}

// Netrc returns a netrc file capable of authenticating GitHub requests and
// cloning GitHub repositories. The netrc will use the global machine account
// when configured.
func (c *client) Netrc(u *model.User, r *model.Repo) (*model.Netrc, error) {
	if c.Password != "" {
		return &model.Netrc{
			Login:    c.Username,
			Password: c.Password,
			Machine:  c.Machine,
		}, nil
	}
	return &model.Netrc{
		Login:    u.Token,
		Password: "x-oauth-basic",
		Machine:  c.Machine,
	}, nil
}

// Deactivate deactives the repository be removing registered push hooks from
// the GitHub repository.
func (c *client) Deactivate(u *model.User, r *model.Repo, link string) error {
	client, ctx := c.newClientToken(u.Token)
	nn := strings.Split(r.FullName, "/")
	hooks, _, err := client.Repositories.ListHooks(ctx, nn[0], nn[1], nil)
	if err != nil {
		return err
	}
	match := matchingHooks(hooks, link)
	if match == nil {
		return nil
	}
	_, err = client.Repositories.DeleteHook(ctx, nn[0], nn[1], *match.ID)
	return err
}

// helper function to return the GitHub oauth2 context using an HTTPClient that
// disables TLS verification if disabled in the remote settings.
func (c *client) newContext() context.Context {
	if !c.SkipVerify {
		return oauth2.NoContext
	}
	return context.WithValue(nil, oauth2.HTTPClient, &http.Client{
		Transport: &http.Transport{
			Proxy: http.ProxyFromEnvironment,
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
		},
	})
}

// helper function to return the GitHub oauth2 config
func (c *client) newConfig(redirect string) *oauth2.Config {
	return &oauth2.Config{
		ClientID:     c.Client,
		ClientSecret: c.Secret,
		Scopes:       defaultScopes,
		Endpoint: oauth2.Endpoint{
			AuthURL:  fmt.Sprintf("%s/login/oauth/authorize", c.URL),
			TokenURL: fmt.Sprintf("%s/login/oauth/access_token", c.URL),
		},
		RedirectURL: fmt.Sprintf("%s/authorize", redirect),
	}
}

// helper function to return the GitHub oauth2 client
func (c *client) newClientToken(token string) (*github.Client, context.Context) {
	ctx := context.Background()
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)
	tc := oauth2.NewClient(oauth2.NoContext, ts)
	if c.SkipVerify {
		tc.Transport.(*oauth2.Transport).Base = &http.Transport{
			Proxy: http.ProxyFromEnvironment,
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
		}
	}
	github := github.NewClient(tc)
	github.BaseURL, _ = url.Parse(c.API)
	return github, ctx
}

// helper function to return matching user email.
func matchingEmail(emails []*github.UserEmail, rawurl string) *github.UserEmail {
	for _, email := range emails {
		if email.Email == nil || email.Primary == nil || email.Verified == nil {
			continue
		}
		if *email.Primary && *email.Verified {
			return email
		}
	}
	// github enterprise does not support verified email addresses so instead
	// we'll return the first email address in the list.
	if len(emails) != 0 && rawurl != defaultAPI {
		return emails[0]
	}
	return nil
}

// helper function to return matching hook.
func matchingHooks(hooks []*github.Hook, rawurl string) *github.Hook {
	link, err := url.Parse(rawurl)
	if err != nil {
		return nil
	}
	for _, hook := range hooks {
		if hook.ID == nil {
			continue
		}
		v, ok := hook.Config["url"]
		if !ok {
			continue
		}
		s, ok := v.(string)
		if !ok {
			continue
		}
		hookurl, err := url.Parse(s)
		if err == nil && hookurl.Host == link.Host {
			return hook
		}
	}
	return nil
}

//
// TODO(bradrydzewski) refactor below functions
//

// Status sends the commit status to the remote system.
// An example would be the GitHub pull request status.
func (c *client) Status(u *model.User, r *model.Repo, b *model.Build, link string) error {
	client, ctx := c.newClientToken(u.Token)
	switch b.Event {
	case "deployment":
		return deploymentStatus(client, ctx, r, b, link)
	default:
		return repoStatus(client, ctx, r, b, link, c.Context)
	}
}

func repoStatus(client *github.Client, ctx context.Context, r *model.Repo, b *model.Build, link, context string) error {
	nn := strings.Split(r.FullName, "/")
	switch b.Event {
	case model.EventPull:
		context += "/pr"
	default:
		if len(b.Event) > 0 {
			context += "/" + b.Event
		}
	}

	data := github.RepoStatus{
		Context:     github.String(context),
		State:       github.String(convertStatus(b.Status)),
		Description: github.String(convertDesc(b.Status)),
		TargetURL:   github.String(link),
	}
	_, _, err := client.Repositories.CreateStatus(ctx, nn[0], nn[1], b.Commit, &data)
	return err
}

var reDeploy = regexp.MustCompile(".+/deployments/(\\d+)")

func deploymentStatus(client *github.Client, ctx context.Context, r *model.Repo, b *model.Build, link string) error {
	nn := strings.Split(r.FullName, "/")
	matches := reDeploy.FindStringSubmatch(b.Link)
	if len(matches) != 2 {
		return nil
	}
	// id, _ := strconv.Atoi(matches[1])
	id, _ := strconv.ParseInt(matches[1], 10, 64)

	data := github.DeploymentStatusRequest{
		State:       github.String(convertStatus(b.Status)),
		Description: github.String(convertDesc(b.Status)),
		LogURL:      github.String(link),
	}
	_, _, err := client.Repositories.CreateDeploymentStatus(ctx, nn[0], nn[1], id, &data)
	return err
}

// Activate activates a repository by creating the post-commit hook and
// adding the SSH deploy key, if applicable.
func (c *client) Activate(u *model.User, r *model.Repo, link string) error {
	nn := strings.Split(r.FullName, "/")
	if err := c.Deactivate(u, r, link); err != nil {
		return err
	}
	client, ctx := c.newClientToken(u.Token)
	hook := &github.Hook{
		// Name: github.String("web"),
		Events: []string{
			"push",
			"pull_request",
			"deployment",
		},
		Config: map[string]interface{}{
			"url":          link,
			"content_type": "form",
		},
		Active: github.Bool(true),
	}
	_, _, err := client.Repositories.CreateHook(ctx, nn[0], nn[1], hook)
	return err
}

// Hook parses the post-commit hook from the Request body
// and returns the required data in a standard format.
func (c *client) Hook(r *http.Request) (*model.Repo, *model.Build, error) {
	return parseHook(r, c.MergeRef)
}
