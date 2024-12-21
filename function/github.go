package function

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"net/http"
	"os"
	"regexp"
	"strings"

	"github.com/gofri/go-github-ratelimit/github_ratelimit"
	"github.com/google/go-github/github"
	"golang.org/x/oauth2"
)

// PrintContributors from a GitHub repository
func PrintContributors(owner, repo string) (output string) {
	api := fmt.Sprintf("https://api.github.com/repos/%s/%s/contributors", owner, repo)

	var (
		contributors []map[string]interface{}
		err          error
	)

	if contributors, err = ghRequestAsSlice(api); err == nil {
		var text string
		group := 6
		for i := 0; i < len(contributors); {
			next := i + group
			if next > len(contributors) {
				next = len(contributors)
			}
			text = text + "<tr>" + generateContributor(contributors[i:next]) + "</tr>"
			i = next
		}

		output = fmt.Sprintf(`<table>%s</table>
`, text)
	}
	return
}

// GetRepoStars awesome ops func
func GetStarLicense(owner, repo string) (rst string) {
	pro, err := GetProject(owner, repo)
	if err != nil {
		panic(err)
	} else {
		var spdxID string
		if license := pro.GetLicense(); license != nil {
			spdxID = license.GetSPDXID()
		} else {
			spdxID = "N/A"
		}
		fmt.Printf("%v|%v|%v|%v", spdxID, pro.GetStargazersCount(), pro.GetCreatedAt().Format("2006-01-02"), pro.GetPushedAt().Format("2006-01-02"))
	}
	return
}

// GetRepoStars Get the stars of a GitHub repository
func GetRepoStars(owner, repo string) (star int) {
	pro, err := GetProject(owner, repo)
	if err != nil {
		panic(err)
	} else {
		star = pro.GetStargazersCount()
	}
	return
}

// GetRepoForks Get the forks of a GitHub repository
func GetRepoForks(owner, repo string) (fork int) {
	pro, err := GetProject(owner, repo)
	if err != nil {
		panic(err)
	} else {
		fork = pro.GetForksCount()
	}
	return
}

// GetRepoWatchers Get the watchers of a GitHub repository
func GetRepoLicenses(owner, repo string) (spdxID string) {
	pro, err := GetProject(owner, repo)
	if err != nil {
		panic(err)
	} else {
		if license := pro.GetLicense(); license != nil {
			spdxID = license.GetSPDXID()
		} else {
			spdxID = "N/A"
		}
	}
	return
}

// GetRepoCreateAt Get the watchers of a GitHub repository
func GetRepoCreateAt(owner, repo string) (create string) {
	pro, err := GetProject(owner, repo)
	if err != nil {
		panic(err)
	} else {
		create = pro.GetCreatedAt().Format("2006-01-02")
	}
	return
}

// GetRepoWatchers Get the watchers of a GitHub repository
func GetRepoPushAt(owner, repo string) (lastupdate interface{}) {
	pro, err := GetProject(owner, repo)
	if err != nil {
		panic(err)
	} else {
		lastupdate = pro.GetPushedAt().Format("2006-01-02")
	}
	return
}

// PrintPages prints the repositories which enabled pages
func PrintPages(owner string) (output string) {
	api := fmt.Sprintf("https://api.github.com/users/%s/repos?type=owner&per_page=100&sort=updated&username=%s", owner, owner)

	var (
		repos []map[string]interface{}
		err   error
	)

	if repos, err = ghRequestAsSlice(api); err == nil {
		var text string
		for i := 0; i < len(repos); i++ {
			repo := strings.TrimSpace(generateRepo(repos[i]))
			if repo != "" {
				text = text + repo + "\n"
			}
		}

		output = fmt.Sprintf(`||||
|---|---|---|
%s`, strings.TrimSpace(text))
	}
	return
}

func ghRequest(api string) (data []byte, err error) {
	var (
		resp *http.Response
		req  *http.Request
	)

	if req, err = http.NewRequest(http.MethodGet, api, nil); err == nil {
		token := os.Getenv("GITHUB_TOKEN")
		if token == "" {
			token = os.Getenv("GH_TOKEN")
		}
		if token != "" {
			req.Header.Set("Authorization", fmt.Sprintf("token %s", token))
		}

		if resp, err = http.DefaultClient.Do(req); err == nil && resp.StatusCode == http.StatusOK {
			data, err = io.ReadAll(resp.Body)
		}
	}
	return
}

func ghRequestAsSlice(api string) (data []map[string]interface{}, err error) {
	var byteData []byte
	if byteData, err = ghRequest(api); err == nil {
		err = json.Unmarshal(byteData, &data)
	}
	return
}

func ghRequestAsMap(api string) (data map[string]interface{}, err error) {
	var byteData []byte
	if byteData, err = ghRequest(api); err == nil {
		err = json.Unmarshal(byteData, &data)
	}
	return
}

var pageRepoTemplate = `
{{if eq .has_pages true}}
|{{.name}}|![GitHub Repo stars](https://img.shields.io/github/stars/{{.owner.login}}/{{.name}}?style=social)|[view](https://{{.owner.login}}.github.io/{{.name}}/)|
{{end}}
`

func generateRepo(repo interface{}) (output string) {
	var tpl *template.Template
	var err error
	if tpl, err = template.New("repo").Parse(pageRepoTemplate); err == nil {
		buf := bytes.NewBuffer([]byte{})
		if err = tpl.Execute(buf, repo); err == nil {
			output = buf.String()
		}
	}
	return
}

func generateContributor(contributors []map[string]interface{}) (output string) {
	var tpl *template.Template
	var err error
	if tpl, err = template.New("contributors").Parse(contributorsTpl); err == nil {
		buf := bytes.NewBuffer([]byte{})
		if err = tpl.Execute(buf, contributors); err == nil {
			output = buf.String()
		}
	}
	return
}

var contributorsTpl = `{{- range $i, $val := .}}
	<td align="center">
		<a href="{{$val.html_url}}">
			<img src="{{$val.avatar_url}}" width="100;" alt="{{$val.login}}"/>
			<br />
			<sub><b>{{$val.login}}</b></sub>
		</a>
	</td>
{{- end}}
`

// GitHubUsersLink parses a text and try to make the potential GitHub IDs be links
func GitHubUsersLink(ids, sep string) (links string) {
	if sep == "" {
		sep = " "
	}

	splits := strings.Split(ids, sep)
	var items []string
	for _, item := range splits {
		items = append(items, GithubUserLink(strings.TrimSpace(item), false))
	}

	// having additional whitespace it's an ASCII character
	if sep == "," {
		sep = sep + " "
	}
	links = strings.Join(items, sep)
	return
}

// GithubUserLink makes a GitHub user link
func GithubUserLink(id string, bio bool) (link string) {
	link = id
	if strings.Contains(id, " ") { // only handle the valid GitHub ID
		return
	}

	// return the original text if there are Markdown style link exist
	if hasLink(id) {
		if bio {
			return GithubUserLink(GetIDFromGHLink(id), bio)
		}
		return
	}

	api := fmt.Sprintf("https://api.github.com/users/%s", id)

	var (
		err  error
		data map[string]interface{}
	)
	if data, err = ghRequestAsMap(api); err == nil {
		link = fmt.Sprintf("[%s](%s)", data["name"], data["html_url"])
		if bioText, ok := data["bio"]; ok && bio && bioText != nil {
			link = fmt.Sprintf("%s (%s)", link, bioText)
		}
	}
	return
}

// GitHubEmojiLink returns a Markdown style link or empty
func GitHubEmojiLink(user string) (output string) {
	if user != "" {
		output = Link(":octocat:", fmt.Sprintf("https://github.com/%s", user))
	}
	return
}

// GetIDFromGHLink return the GitHub ID from a link
func GetIDFromGHLink(link string) string {
	reg, _ := regexp.Compile("\\[.*\\]\\(.*/|\\)")
	return reg.ReplaceAllString(link, "")
}

// PrintUserAsTable generates a table for a GitHub user
func PrintUserAsTable(id string) (result string) {
	api := fmt.Sprintf("https://api.github.com/users/%s", id)

	result = `|||
|---|---|
`

	var (
		err  error
		data map[string]interface{}
	)
	if data, err = ghRequestAsMap(api); err == nil {
		result = result + addWithEmpty("Name", "name", data) +
			addWithEmpty("Location", "location", data) +
			addWithEmpty("Bio", "bio", data) +
			addWithEmpty("Blog", "blog", data) +
			addWithEmpty("Twitter", "twitter_username", data) +
			addWithEmpty("Organization", "company", data)
	}
	return
}

func addWithEmpty(title, key string, data map[string]interface{}) (result string) {
	if val, ok := data[key]; ok && val != "" {
		desc := val
		switch key {
		case "twitter_username":
			desc = fmt.Sprintf("[%s](https://twitter.com/%s)", val, val)
		}
		result = fmt.Sprintf(`| %s | %s |
`, title, desc)
	}
	return
}

// hasLink determines if there are Markdown style links
func hasLink(text string) (ok bool) {
	reg, _ := regexp.Compile(".*\\[.*\\]\\(.*\\)")
	ok = reg.MatchString(text)
	return
}

var (
	client *github.Client
)

func init() {
	token := os.Getenv("GITHUB_TOKEN")
	if token == "" {
		token = os.Getenv("GH_TOKEN")
	}
	ctx := context.Background()
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)
	tc := oauth2.NewClient(ctx, ts)
	rateLimiter, err := github_ratelimit.NewRateLimitWaiterClient(tc.Transport)
	if err != nil {
		panic(err)
	}
	client = github.NewClient(rateLimiter)
}

// GetProject 获取项目信息
func GetProject(owner, repoName string) (*github.Repository, error) {
	ctx := context.Background()
	repo, _, err := client.Repositories.Get(ctx, owner, repoName)
	if err != nil {
		return nil, err
	}
	return repo, nil
}

// UpdateRepoDescription 更新项目描述
func UpdateRepoDescription(owner, repoName, description string) error {
	ctx := context.Background()
	_, _, err := client.Repositories.Edit(ctx, owner, repoName, &github.Repository{
		Description: &description,
	})
	if err != nil {
		return err
	}

	return nil
}
