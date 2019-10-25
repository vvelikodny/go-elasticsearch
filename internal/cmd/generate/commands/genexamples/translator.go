// Licensed to Elasticsearch B.V. under one or more agreements.
// Elasticsearch B.V. licenses this file to you under the Apache 2.0 License.
// See the LICENSE file in the project root for more information.

package genexamples

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"strings"
)

const tail = "\t" + `if err != nil {
		fmt.Println("Error getting the response:", err)
		os.Exit(1)
	}
	defer res.Body.Close()
	fmt.Println(res)
`

var ConsoleToGo []TranslateRule

func init() {
	ConsoleToGo = []TranslateRule{

		{Pattern: "^GET /$",
			Func: func(e Example) string {
				return `res, err := es.Info()`
			}},

		{Pattern: `^GET /_cat/health\?v`,
			Func: func(e Example) string {
				return "\tres, err := es.Cat.Health(es.Cat.Health.WithV(true))"
			}},

		{Pattern: `^PUT /\w+/_doc/\w+`,
			Func: func(e Example) string {
				re := regexp.MustCompile(`(?ms)^PUT /(?P<index>\w+)/_doc/(?P<id>\w+)\s(?P<body>.*)`)
				matches := re.FindStringSubmatch(e.Source)
				if len(matches) < 4 {
					// TODO(karmi): Proper error handling
					fmt.Println(e.Source)
					panic("Cannot match example source")
				}

				var src strings.Builder
				src.WriteString("\tres, err := es.Index(\n")
				fmt.Fprintf(&src, "\t%q,\n", matches[1])
				var body bytes.Buffer
				json.Indent(&body, []byte(matches[3]), "\t\t", "  ")
				fmt.Fprintf(&src, "\tstrings.NewReader(`%s`),\n", body.String())
				fmt.Fprintf(&src, "\tes.Index.WithDocumentID(%q),\n", matches[2])
				src.WriteString("\tes.Index.WithPretty(),\n")
				src.WriteString("\t)\n")

				return src.String()
			}},

		{Pattern: `^GET /\w+/_doc/\w+$`,
			Func: func(e Example) string {
				re := regexp.MustCompile(`(?ms)^GET /(?P<index>\w+)/_doc/(?P<id>\w+)\s*$`)
				matches := re.FindStringSubmatch(e.Source)
				if len(matches) < 3 {
					// TODO(karmi): Proper error handling
					fmt.Println(e.Source)
					panic("Cannot match example source")
				}
				return fmt.Sprintf("\tres, err := es.Get(%q, %q, es.Get.WithPretty())", matches[1], matches[2])
			}},

		{Pattern: `^GET /\w+/_search`,
			Func: func(e Example) string {
				re := regexp.MustCompile(`(?ms)^GET /(?P<index>\w+)/_search\s(?P<body>.*)`)
				matches := re.FindStringSubmatch(e.Source)
				if len(matches) < 3 {
					// TODO(karmi): Proper error handling
					fmt.Println(e.Source)
					panic("Cannot match example source")
				}

				var src strings.Builder
				src.WriteString("\tres, err := es.Search(\n")
				fmt.Fprintf(&src, "\tes.Search.WithIndex(%q),\n", matches[1])
				var body bytes.Buffer
				json.Indent(&body, []byte(matches[2]), "\t\t", "  ")
				fmt.Fprintf(&src, "\tes.Search.WithBody(strings.NewReader(`%s`)),\n", body.String())
				src.WriteString("\tes.Search.WithPretty(),\n")
				src.WriteString("\t)\n")

				return src.String()
			}},
	}
}

// Translator represents converter from Console source code to Go source code.
//
type Translator struct {
	Example Example
}

// TranslateRule represents a rule for translating code from Console to Go.
//
type TranslateRule struct {
	Pattern string
	Func    func(Example) string
}

// IsTranslated returns true when a rule for translating the Console example to Go source code exists.
//
func (t Translator) IsTranslated() bool {
	for _, r := range ConsoleToGo {
		if r.Match(t.Example) {
			return true
		}
	}
	return false
}

// Translate returns the Console code translated to Go.
//
func (t Translator) Translate() (string, error) {
	for _, r := range ConsoleToGo {
		if r.Match(t.Example) {
			var out strings.Builder

			src := r.Func(t.Example)
			out.WriteRune('\n')
			fmt.Fprintf(&out, "\t// tag:%s[]\n", t.Example.Digest)
			out.WriteString(src)
			out.WriteRune('\n')
			fmt.Fprintf(&out, "\t// end:%s[]\n", t.Example.Digest)
			out.WriteString(tail)

			return out.String(), nil
		}
	}
	return "", errors.New("no rule to translate the example")
}

// Match returns true when the example source matches the rule pattern.
//
func (r TranslateRule) Match(e Example) bool {
	matched, _ := regexp.MatchString(r.Pattern, e.Source)
	return matched
}
