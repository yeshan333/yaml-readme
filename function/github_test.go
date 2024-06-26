package function

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"testing"

	"github.com/h2non/gock"
	"github.com/stretchr/testify/assert"
)

func Test_printContributor(t *testing.T) {
	type args struct {
		owner string
		repo  string
	}
	tests := []struct {
		name       string
		args       args
		prepare    func()
		wantOutput func() string
	}{{
		name: "normal case",
		args: args{
			owner: "linuxsuren",
			repo:  "yaml-readme",
		},
		prepare: func() {
			gock.New("https://api.github.com").
				Get("/repos/linuxsuren/yaml-readme/contributors").
				Reply(http.StatusOK).
				File("data/yaml-readme.json")
		},
		wantOutput: func() string {
			data, _ := ioutil.ReadFile("data/yaml-readme-contributors.txt")
			return string(data)
		},
	}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer gock.Off()
			tt.prepare()
			assert.Equalf(t, tt.wantOutput(), PrintContributors(tt.args.owner, tt.args.repo), "printContributors(%v, %v)", tt.args.owner, tt.args.repo)
		})
	}
}

func TestGithubUserLink(t *testing.T) {
	type args struct {
		id  string
		bio bool
	}
	tests := []struct {
		name     string
		mockUser string
		args     args
		want     string
	}{{
		name:     "normal case without bio",
		mockUser: "linuxsuren",
		args: args{
			id:  "linuxsuren",
			bio: false,
		},
		want: `[Rick](https://github.com/LinuxSuRen)`,
	}, {
		name:     "normal case with bio",
		mockUser: "linuxsuren",
		args: args{
			id:  "linuxsuren",
			bio: true,
		},
		want: `[Rick](https://github.com/LinuxSuRen) (程序员，业余开源布道者)`,
	}, {
		name:     "with whitespace",
		mockUser: "linuxsuren",
		args: args{
			id:  "this is not id",
			bio: false,
		},
		want: "this is not id",
	}, {
		name:     "has Markdown style link",
		mockUser: "linuxsuren",
		args: args{
			id:  "[name](link)",
			bio: false,
		},
		want: "[name](link)",
	}, {
		name:     "has Markdown style link, want bio",
		mockUser: "linuxsuren",
		args: args{
			id:  "[Rick](https://github.com/linuxsuren)",
			bio: true,
		},
		want: `[Rick](https://github.com/LinuxSuRen) (程序员，业余开源布道者)`,
	}, {
		name:     "do not have bio",
		mockUser: "linuxsuren-bot",
		args: args{
			id:  "linuxsuren-bot",
			bio: true,
		},
		want: `[LinuxSuRen-bot](https://github.com/linuxsuren-bot)`,
	}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer gock.Off()
			mockGitHubUser(tt.mockUser)
			assert.Equalf(t, tt.want, GithubUserLink(tt.args.id, tt.args.bio), "GithubUserLink(%v, %v)", tt.args.id, tt.args.bio)
		})
	}
}

func mockGitHubUser(id string) {
	gock.New("https://api.github.com").
		Get(fmt.Sprintf("/users/%s", id)).Reply(http.StatusOK).File(fmt.Sprintf("data/%s.json", id))
}

func mockUserRepos(owner string) {
	gock.New("https://api.github.com").
		Get(fmt.Sprintf("/users/%s/repos", owner)).
		MatchParam("type", "owner").
		MatchParam("per_page", "100").
		MatchParam("sort", "updated").
		MatchParam("username", owner).
		Reply(http.StatusOK).File("data/repos.json")
}

func TestGitHubUsersLink(t *testing.T) {
	type args struct {
		ids string
		sep string
	}
	tests := []struct {
		name      string
		prepare   func()
		args      args
		wantLinks string
	}{{
		name: "two GitHub users",
		prepare: func() {
			mockGitHubUser("linuxsuren")
			mockGitHubUser("linuxsuren")
		},
		args: args{
			ids: "linuxsuren linuxsuren",
			sep: "",
		},
		wantLinks: "[Rick](https://github.com/LinuxSuRen) [Rick](https://github.com/LinuxSuRen)",
	}, {
		name: "two GitHub users with Chinese character as separate",
		prepare: func() {
			mockGitHubUser("linuxsuren")
			mockGitHubUser("linuxsuren")
		},
		args: args{
			ids: "linuxsuren、linuxsuren",
			sep: "、",
		},
		wantLinks: "[Rick](https://github.com/LinuxSuRen)、[Rick](https://github.com/LinuxSuRen)",
	}, {
		name: "two GitHub users with whitespace and comma as separate",
		prepare: func() {
			mockGitHubUser("linuxsuren")
			mockGitHubUser("linuxsuren")
		},
		args: args{
			ids: "linuxsuren, linuxsuren",
			sep: ",",
		},
		wantLinks: "[Rick](https://github.com/LinuxSuRen), [Rick](https://github.com/LinuxSuRen)",
	}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer gock.Off()
			tt.prepare()
			assert.Equalf(t, tt.wantLinks, GitHubUsersLink(tt.args.ids, tt.args.sep), "GitHubUsersLink(%v, %v)", tt.args.ids, tt.args.sep)
		})
	}
}

func Test_hasLink(t *testing.T) {
	type args struct {
		text string
	}
	tests := []struct {
		name   string
		args   args
		wantOk bool
	}{{
		name: "normal text",
		args: args{
			text: "This is a normal text",
		},
		wantOk: false,
	}, {
		name: "has Markdown style link",
		args: args{
			text: "[here](link)",
		},
		wantOk: true,
	}, {
		name: "more complex Markdown style link",
		args: args{
			text: "Hi there, this is [my card](link).",
		},
		wantOk: true,
	}, {
		name: "multiple Markdown style link",
		args: args{
			text: "I have two links, [one](link) and [two](link).",
		},
		wantOk: true,
	}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equalf(t, tt.wantOk, hasLink(tt.args.text), "hasLink(%v)", tt.args.text)
		})
	}
}

func Test_ghRequest(t *testing.T) {
	type args struct {
		api string
	}
	tests := []struct {
		name     string
		args     args
		wantData []byte
		wantErr  assert.ErrorAssertionFunc
	}{{
		name: "with token",
		args: args{
			api: "https://fake.com",
		},
		wantData: []byte("body"),
		wantErr: func(t assert.TestingT, err error, i ...interface{}) bool {
			assert.Nil(t, err)
			return true
		},
	}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			oldToken := os.Getenv("GITHUB_TOKEN")
			_ = os.Setenv("GITHUB_TOKEN", "fake")
			defer func() {
				_ = os.Setenv("GITHUB_TOKEN", oldToken)
			}()

			gock.New(tt.args.api).Get("/").Reply(http.StatusOK).Body(bytes.NewBufferString("body"))

			gotData, err := ghRequest(tt.args.api)
			if !tt.wantErr(t, err, fmt.Sprintf("ghRequest(%v)", tt.args.api)) {
				return
			}
			assert.Equalf(t, tt.wantData, gotData, "ghRequest(%v)", tt.args.api)
		})
	}
}

func TestGitHubEmojiLink(t *testing.T) {
	type args struct {
		user string
	}
	tests := []struct {
		name       string
		args       args
		wantOutput string
	}{{
		name:       "user is empty",
		wantOutput: "",
	}, {
		name: "user is not empty",
		args: args{
			user: "linuxsuren",
		},
		wantOutput: "[:octocat:](https://github.com/linuxsuren)",
	}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equalf(t, tt.wantOutput, GitHubEmojiLink(tt.args.user), "GitHubEmojiLink(%v)", tt.args.user)
		})
	}
}

func Test_getIDFromGHLink(t *testing.T) {
	type args struct {
		link string
	}
	tests := []struct {
		name string
		args args
		want string
	}{{
		name: "normal",
		args: args{
			link: "[Rick](https://github.com/LinuxSuRen)",
		},
		want: "LinuxSuRen",
	}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equalf(t, tt.want, GetIDFromGHLink(tt.args.link), "GetIDFromGHLink(%v)", tt.args.link)
		})
	}
}

func TestPrintUserAsTable(t *testing.T) {
	type args struct {
		id string
	}
	tests := []struct {
		name       string
		args       args
		wantResult string
	}{{
		name: "normal",
		args: args{
			id: "linuxsuren",
		},
		wantResult: `|||
|---|---|
| Name | Rick |
| Location | China |
| Bio | 程序员，业余开源布道者 |
| Blog | https://linuxsuren.github.io/open-source-best-practice/ |
| Twitter | [linuxsuren](https://twitter.com/linuxsuren) |
| Organization | @opensource-f2f @kubesphere |
`,
	}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer gock.Off()
			mockGitHubUser(tt.args.id)
			assert.Equalf(t, tt.wantResult, PrintUserAsTable(tt.args.id), "PrintUserAsTable(%v)", tt.args.id)
		})
	}
}

func TestPrintPages(t *testing.T) {
	type args struct {
		owner string
	}
	tests := []struct {
		name       string
		args       args
		wantOutput string
	}{{
		name: "normal",
		args: args{owner: "linuxsuren"},
		wantOutput: `||||
|---|---|---|
|Hello-World|![GitHub Repo stars](https://img.shields.io/github/stars/linuxsuren/Hello-World?style=social)|[view](https://linuxsuren.github.io/Hello-World/)|`,
	}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer gock.Off()
			mockUserRepos(tt.args.owner)
			assert.Equalf(t, tt.wantOutput, PrintPages(tt.args.owner), "PrintPages(%v)", tt.args.owner)
		})
	}
}
