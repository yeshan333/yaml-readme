package main

import (
	"bytes"
	"fmt"
	"html/template"
	"io"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/Masterminds/sprig"
	"github.com/linuxsuren/yaml-readme/function"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"
)

var logger *log.Logger

type option struct {
	pattern       string
	templateFile  string
	includeHeader bool
	sortBy        string
	groupBy       string

	printFunctions bool
	printVariables bool
}

func loadMetadata(pattern, groupBy string) (items []map[string]interface{},
	groupData map[string][]map[string]interface{}, err error) {
	groupData = make(map[string][]map[string]interface{})

	// find YAML files
	var files []string
	var data []byte
	if files, err = filepath.Glob(pattern); err == nil {
		for _, metaFile := range files {
			if data, err = ioutil.ReadFile(metaFile); err != nil {
				logger.Printf("failed to read file [%s], error: %v\n", metaFile, err)
				continue
			}

			metaMap := make(map[string]interface{})
			if err = yaml.Unmarshal(data, metaMap); err != nil {
				logger.Printf("failed to parse file [%s] as a YAML, error: %v\n", metaFile, err)
				continue
			}

			// skip this item if there is a 'ignore' key is true
			if val, ok := metaMap["ignore"]; ok {
				if ignore, ok := val.(bool); ok && ignore {
					continue
				}
			}

			filename := strings.TrimSuffix(filepath.Base(metaFile), filepath.Ext(metaFile))
			parentname := filepath.Base(filepath.Dir(metaFile))

			metaMap["filename"] = filename
			metaMap["parentname"] = parentname
			metaMap["fullpath"] = metaFile

			if val, ok := metaMap[groupBy]; ok && val != "" {
				var strVal string
				switch val.(type) {
				case string:
					strVal = val.(string)
				case int:
					strVal = strconv.Itoa(val.(int))
				}

				if _, ok := groupData[strVal]; ok {
					groupData[strVal] = append(groupData[strVal], metaMap)
				} else {
					groupData[strVal] = []map[string]interface{}{
						metaMap,
					}
				}
			}

			items = append(items, metaMap)
		}
	}
	return
}

func sortMetadata(items []map[string]interface{}, sortByField string) {
	descending := true
	if strings.HasPrefix(sortByField, "!") {
		sortByField = strings.TrimPrefix(sortByField, "!")
		descending = false
	}
	sortBy(items, sortByField, descending)
}

func loadTemplate(templateFile string, includeHeader bool) (readmeTpl string, err error) {
	// load readme template
	var data []byte
	if data, err = ioutil.ReadFile(templateFile); err != nil {
		fmt.Printf("failed to load README template, error: %v\n", err)
		err = nil
		readmeTpl = `|中文名称|英文名称|JD|
|---|---|---|
{{- range $val := .}}
|{{$val.zh}}|{{$val.en}}|{{$val.jd}}|
{{- end}}`
	}
	if includeHeader {
		readmeTpl = fmt.Sprintf("> This file was generated by [%s](%s) via [yaml-readme](https://github.com/LinuxSuRen/yaml-readme), please don't edit it directly!\n\n",
			filepath.Base(templateFile), filepath.Base(templateFile))
	}
	readmeTpl = readmeTpl + string(data)
	s, err := regexp.Compile("#!yaml-readme .*\n")
	readmeTpl = s.ReplaceAllString(readmeTpl, "")
	return
}

func (o *option) runE(cmd *cobra.Command, args []string) (err error) {
	logger = log.New(cmd.ErrOrStderr(), "", log.LstdFlags)
	if o.printFunctions {
		printFunctions(cmd.OutOrStdout())
		return
	}

	if o.printVariables {
		printVariables(cmd.OutOrStdout())
		return
	}

	// load metadata from YAML files
	var items []map[string]interface{}
	var groupData map[string][]map[string]interface{}
	if items, groupData, err = loadMetadata(o.pattern, o.groupBy); err != nil {
		err = fmt.Errorf("failed to load metadat from %q", o.pattern)
		return
	}
	groupNum := len(groupData)
	itemNum := len(items)
	if o.sortBy != "" {
		sortMetadata(items, o.sortBy)
	}

	// load readme template
	var readmeTpl string
	if readmeTpl, err = loadTemplate(o.templateFile, o.includeHeader); err != nil {
		err = fmt.Errorf("failed to load template file from %q", o.templateFile)
		return
	}

	// render it with grouped data
	if o.groupBy != "" {
		err = renderTemplate(readmeTpl, groupData, uint(groupNum), uint(itemNum), cmd.OutOrStdout())
	} else {
		err = renderTemplate(readmeTpl, items, uint(groupNum), uint(itemNum), cmd.OutOrStdout())
	}
	return
}

func renderTemplateToString(tplContent string, object interface{}) (output string, err error) {
	buf := bytes.NewBuffer([]byte{})
	if err = renderTemplate(tplContent, object, 0, 0, buf); err == nil {
		output = buf.String()
	}
	return
}

func renderTemplate(tplContent string, object interface{}, groupNum, itemNum uint, writer io.Writer) (err error) {
	var tpl *template.Template
	if tpl, err = template.New("readme").
		Funcs(getFuncMap(tplContent, groupNum, itemNum)).
		Funcs(sprig.FuncMap()).Parse(tplContent); err == nil {
		err = tpl.Execute(writer, object)
	}
	return
}

func printVariables(stdout io.Writer) {
	_, _ = stdout.Write([]byte(`filename
parentname
fullpath`))
}

func printFunctions(stdout io.Writer) {
	funcMap := getFuncMap("", 0, 0)
	var funcs []string
	for k := range funcMap {
		funcs = append(funcs, k)
	}
	sort.SliceStable(funcs, func(i, j int) bool {
		return strings.Compare(funcs[i], funcs[j]) < 0
	})
	_, _ = stdout.Write([]byte(strings.Join(funcs, "\n")))
}

func getFuncMap(readmeTpl string, groupNum, itemNum uint) template.FuncMap {
	return template.FuncMap{
		"printHelp": func(cmd string) (output string) {
			var err error
			var data []byte
			if data, err = exec.Command(cmd, "--help").Output(); err != nil {
				_, _ = fmt.Fprintln(os.Stderr, "failed to run command", cmd)
			} else {
				output = fmt.Sprintf(`%s
%s
%s`, "```shell", string(data), "```")
			}
			return
		},
		"lenItemNum": func() uint {
			return itemNum
		},
		"lenGroupNum": func() uint {
			return groupNum
		},
		"updateDesc": func(owner, repo string) string {
			desc := fmt.Sprintf("🧰 记录每一个与运维相关的优秀项目，⚗️ 项目内表格通过 GitHub Action 自动生成，📥 当前收录项目 %d 个。", itemNum)
			err := function.UpdateRepoDescription(owner, repo, desc)
			if err != nil {
				fmt.Printf("failed to update repo description, error: %v\n", err)
				os.Exit(1)
			}
			return ""
		},
		"printToc": func() string {
			return generateTOC(readmeTpl)
		},
		"printContributors": func(owner, repo string) template.HTML {
			return template.HTML(function.PrintContributors(owner, repo))
		},
		"printStarHistory": func(owner, repo string) string {
			return printStarHistory(owner, repo)
		},
		"printVisitorCount": func(id string) string {
			return fmt.Sprintf(`![Visitor Count](https://profile-counter.glitch.me/%s/count.svg)`, id)
		},
		"printPages": func(owner string) string {
			return function.PrintPages(owner)
		},
		"getLatestFeedPost": func(feedLink string, defaultContent string) string {
			return function.GetLatestFeedPost(feedLink, defaultContent)
		},
		"render":       dataRender,
		"gh":           function.GithubUserLink,
		"ghs":          function.GitHubUsersLink,
		"ghEmoji":      function.GitHubEmojiLink,
		"link":         function.Link,
		"linkOrEmpty":  function.LinkOrEmpty,
		"twitterLink":  function.TwitterLink,
		"youTubeLink":  function.YouTubeLink,
		"gstatic":      function.GStatic,
		"ghID":         function.GetIDFromGHLink,
		"ghStar":       function.GetRepoStars,
		"ghFork":       function.GetRepoForks,
		"ghCreate":     function.GetRepoCreateAt,
		"ghUpdate":     function.GetRepoPushAt,
		"ghLicense":    function.GetRepoLicenses,
		"ghCustom":     function.GetStarLicense,
		"printGHTable": function.PrintUserAsTable,
	}
}

func sortBy(items []map[string]interface{}, sortBy string, descending bool) {
	sort.SliceStable(items, func(i, j int) (compare bool) {
		left, ok := items[i][sortBy].(string)
		if !ok {
			return false
		}
		right, ok := items[j][sortBy].(string)
		if !ok {
			return false
		}

		compare = strings.Compare(left, right) < 0
		if !descending {
			compare = !compare
		}
		return
	})
}

func generateTOC(txt string) (toc string) {
	items := strings.Split(txt, "\n")
	for i := range items {
		item := items[i]

		var prefix string
		var tag string
		if strings.HasPrefix(item, "## ") {
			tag = strings.TrimPrefix(item, "## ")
			prefix = "- "
		} else if strings.HasPrefix(item, "### ") {
			tag = strings.TrimPrefix(item, "### ")
			prefix = " - "
		} else {
			continue
		}

		// not support those titles which have whitespaces
		tag = strings.TrimSpace(tag)
		if len(strings.Split(tag, " ")) > 1 {
			continue
		}

		toc = toc + fmt.Sprintf("%s[%s](#%s)\n", prefix, tag, strings.ToLower(tag))
	}
	return
}

func printStarHistory(owner, repo string) string {
	return fmt.Sprintf(`[![Star History Chart](https://api.star-history.com/svg?repos=%[1]s/%[2]s&type=Date)](https://star-history.com/#%[1]s/%[2]s&Date)`,
		owner, repo)
}

func dataRender(data interface{}) string {
	switch val := data.(type) {
	case bool:
		if val {
			return ":white_check_mark:"
		} else {
			return ":x:"
		}
	case string:
		return val
	}
	return ""
}

func newRootCommand() (cmd *cobra.Command) {
	opt := &option{}
	cmd = &cobra.Command{
		Use:   "yaml-readme",
		Short: "A helper to generate a README file from Golang-based template",
		Long: `A helper to generate a README file from Golang-based template
Some functions rely on the GitHub API, in order to avoid X-RateLimit-Limit errors you can set an environment variable: 'GITHUB_TOKEN'`,
		RunE: opt.runE,
	}
	cmd.SetOut(os.Stdout)
	flags := cmd.Flags()
	flags.StringVarP(&opt.pattern, "pattern", "p", "items/*.yaml",
		"The glob pattern with Golang spec to find files")
	flags.StringVarP(&opt.templateFile, "template", "t", "README.tpl",
		"The template file which should follow Golang template spec")
	flags.BoolVarP(&opt.includeHeader, "include-header", "", true,
		"Indicate if include a notice header on the top of the README file")
	flags.StringVarP(&opt.sortBy, "sort-by", "", "",
		"Sort the array data descending by which field, or sort it ascending with the prefix '!'. For example: --sort-by !year")
	flags.StringVarP(&opt.groupBy, "group-by", "", "",
		"Group the array data by which field")
	flags.BoolVarP(&opt.printFunctions, "print-functions", "", false,
		"Print all the functions and exit")
	flags.BoolVarP(&opt.printVariables, "print-variables", "", false,
		"Print all the variables and exit")
	return
}

func main() {
	if err := newRootCommand().Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
