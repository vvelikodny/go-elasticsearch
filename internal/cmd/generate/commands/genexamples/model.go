// Licensed to Elasticsearch B.V. under one or more agreements.
// Elasticsearch B.V. licenses this file to you under the Apache 2.0 License.
// See the LICENSE file in the project root for more information.

package genexamples

import (
	"fmt"
	"regexp"
	"sort"
)

func init() {
	sort.Strings(EnabledFiles)
}

// EnabledFiles contains a list of files where documentation should be generated.
//
var EnabledFiles = []string{"getting-started.asciidoc"}

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

func (e Example) Enabled() bool {
	index := sort.SearchStrings(EnabledFiles, e.SourceLocation.File)

	if index > len(EnabledFiles)-1 || EnabledFiles[index] != e.SourceLocation.File {
		return false
	}

	return true
}

func (e Example) Executable() bool {
	matched, _ := regexp.MatchString(`^HEAD|GET|PUT|DELETE|POST`, e.Source)
	return matched
}

func (e Example) ID() string {
	return fmt.Sprintf("%s:%d", e.SourceLocation.File, e.SourceLocation.Line)
}

func (e Example) GithubURL() string {
	return fmt.Sprintf("https://github.com/elastic/elasticsearch/blob/master/docs/reference/%s#L%d", e.SourceLocation.File, e.SourceLocation.Line)
}
