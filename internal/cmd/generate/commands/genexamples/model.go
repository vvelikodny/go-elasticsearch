// Licensed to Elasticsearch B.V. under one or more agreements.
// Elasticsearch B.V. licenses this file to you under the Apache 2.0 License.
// See the LICENSE file in the project root for more information.

package genexamples

import (
	"bufio"
	"fmt"
	"regexp"
	"sort"
	"strings"
)

func init() {
	sort.Strings(EnabledFiles)
}

// EnabledFiles contains a list of files where documentation should be generated.
//
var EnabledFiles = []string{
	"docs/delete.asciidoc",
	"docs/get.asciidoc",
	"docs/index_.asciidoc",
	"getting-started.asciidoc",
	"query-dsl/query-string-query.asciidoc",
	"search/request-body.asciidoc",
	"setup/install/check-running.asciidoc",
}

var (
	reHTTPMethod = regexp.MustCompile(`^HEAD|GET|PUT|DELETE|POST`)
	reAnnotation = regexp.MustCompile(`^(?P<line>.*)\s*\<\d\>\s*$`)
)

// Example represents the code example in documentation.
//
// See: https://github.com/elastic/built-docs/blob/master/raw/en/elasticsearch/reference/master/alternatives_report.json
//
type Example struct {
	SourceLocation struct {
		File string
		Line int
	} `json:"source_location"`

	Digest string
	Source string
}

// IsEnabled returns true when the example should be processed.
//
func (e Example) IsEnabled() bool {
	// TODO(karmi): Use "filepatch.Match()" to support glob patterns

	index := sort.SearchStrings(EnabledFiles, e.SourceLocation.File)

	if index > len(EnabledFiles)-1 || EnabledFiles[index] != e.SourceLocation.File {
		return false
	}

	return true
}

// IsExecutable returns true when the example contains a request.
//
func (e Example) IsExecutable() bool {
	return reHTTPMethod.MatchString(e.Source)
}

// IsTranslated returns true when the example can be converted to Go source code.
//
func (e Example) IsTranslated() bool {
	return Translator{Example: e}.IsTranslated()
}

// ID returns example identifier.
//
func (e Example) ID() string {
	return fmt.Sprintf("%s:%d", e.SourceLocation.File, e.SourceLocation.Line)
}

// Chapter returns the example chapter.
//
func (e Example) Chapter() string {
	r := strings.NewReplacer("/", "_", "-", "_", ".asciidoc", "")
	return r.Replace(e.SourceLocation.File)
}

// GithubURL returns a link for the example source.
//
func (e Example) GithubURL() string {
	return fmt.Sprintf("https://github.com/elastic/elasticsearch/blob/master/docs/reference/%s#L%d", e.SourceLocation.File, e.SourceLocation.Line)
}

// Commands returns the list of commands from source.
//
func (e Example) Commands() ([]string, error) {
	var (
		buf  strings.Builder
		list []string
		scan = bufio.NewScanner(strings.NewReader(e.Source))
	)

	for scan.Scan() {
		line := scan.Text()

		line = reAnnotation.ReplaceAllString(line, "$1")

		if reHTTPMethod.MatchString(line) {
			if buf.Len() > 0 {
				list = append(list, buf.String())
			}
			buf.Reset()
		}

		buf.WriteString(line)
		buf.WriteString("\n")
	}

	if err := scan.Err(); err != nil {
		return list, err
	}

	if buf.Len() > 0 {
		list = append(list, buf.String())
	}

	return list, nil
}

// Translated returns the code translated from Console to Go.
//
func (e Example) Translated() (string, error) {
	return Translator{Example: e}.Translate()
}
