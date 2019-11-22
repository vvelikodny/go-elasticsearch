// Licensed to Elasticsearch B.V. under one or more agreements.
// Elasticsearch B.V. licenses this file to you under the Apache 2.0 License.
// See the LICENSE file in the project root for more information.

package genexamples

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/elastic/go-elasticsearch/v8/internal/cmd/generate/utils"
)

const testCheck = "\t" + `if err != nil {            // SKIP
		t.Fatalf("Error getting the response: %s", err)	 // SKIP
	}                                                  // SKIP
	defer res.Body.Close()                             // SKIP
`

// ConsoleToGo contains translation rules.
//
var ConsoleToGo = []TranslateRule{

	{ // ----- Info() -----------------------------------------------------------
		Pattern: "^GET /$",
		Func: func(in string) (string, error) {
			return "res, err := es.Info()", nil
		}},

	{ // ----- Cat.Health() -----------------------------------------------------
		Pattern: `^GET /_cat/health\?v`,
		Func: func(in string) (string, error) {
			return "\tres, err := es.Cat.Health(es.Cat.Health.WithV(true))", nil
		}},

	{ // ----- Cluster.PutSettings() --------------------------------------------
		Pattern: `^PUT /?_cluster/settings`,
		Func: func(in string) (string, error) {
			var src strings.Builder

			re := regexp.MustCompile(`(?ms)^(?P<method>PUT) /?_cluster/settings/?(?P<params>\??[\S]+)?\s?(?P<body>.*)`)
			matches := re.FindStringSubmatch(in)
			if matches == nil {
				return "", errors.New("cannot match example source to pattern")
			}

			fmt.Fprintf(&src, "\tres, err := es.Cluster.PutSettings(\n")
			body, err := bodyStringToReader(matches[3])
			if err != nil {
				return "", fmt.Errorf("error converting body: %s", err)
			}
			fmt.Fprintf(&src, "\t%s", body)
			src.WriteString(")")

			return src.String(), nil

		}},

	{ // ----- Index() or Create() ----------------------------------------------
		Pattern: `^(PUT|POST) /?\w+/(_doc|_create)/?.*`,
		Func: func(in string) (string, error) {
			var (
				src     strings.Builder
				apiName string
			)

			re := regexp.MustCompile(`(?ms)^(?P<method>PUT|POST) /?(?P<index>\w+)/(?P<api>_doc|_create)/?(?P<id>\w+)?(?P<params>\??[\S]+)?\s?(?P<body>.*)`)
			matches := re.FindStringSubmatch(in)
			if matches == nil {
				return "", errors.New("cannot match example source to pattern")
			}

			if matches[3] == "_create" {
				apiName = "Create"
			} else {
				apiName = "Index"
			}

			src.WriteString("\tres, err := es." + apiName + "(\n")

			fmt.Fprintf(&src, "\t%q,\n", matches[2])

			if apiName == "Create" {
				fmt.Fprintf(&src, "\t%q,\n", matches[4])
			}

			body, err := bodyStringToReader(matches[6])
			if err != nil {
				return "", fmt.Errorf("error converting body: %s", err)
			}
			fmt.Fprintf(&src, "\t%s,\n", body)

			if apiName == "Index" {
				if matches[4] != "" {
					fmt.Fprintf(&src, "\tes."+apiName+".WithDocumentID(%q),\n", matches[4])
				}
			}

			if matches[5] != "" {
				params, err := queryToParams(matches[5])
				if err != nil {
					return "", fmt.Errorf("error parsing URL params: %s", err)
				}
				args, err := paramsToArguments(apiName, params)
				if err != nil {
					return "", fmt.Errorf("error converting params to arguments: %s", err)
				}
				fmt.Fprintf(&src, args)
			}

			src.WriteString("\tes." + apiName + ".WithPretty(),\n")
			src.WriteString("\t)")

			return src.String(), nil
		}},

	{ // ----- Indices.Create() -------------------------------------------------
		Pattern: `^PUT /?[^/\s]+\s?(?P<body>.+)?`,
		Func: func(in string) (string, error) {
			var src strings.Builder

			re := regexp.MustCompile(`(?ms)^PUT /?(?P<index>[^/\s]+)(?P<params>\??[\S/]+)?\s?(?P<body>.+)?`)
			matches := re.FindStringSubmatch(in)
			if matches == nil {
				return "", errors.New("cannot match example source to pattern")
			}

			src.WriteString("\tres, err := es.Indices.Create(")
			if matches[2] != "" || matches[3] != "" {
				fmt.Fprintf(&src, "\n\t%q,\n", matches[1])

				if matches[3] != "" {
					body, err := bodyStringToReader(matches[3])
					if err != nil {
						return "", fmt.Errorf("error converting body: %s", err)
					}
					fmt.Fprintf(&src, "\tes.Indices.Create.WithBody(%s),\n", body)
				}

				if matches[2] != "" {
					params, err := queryToParams(matches[2])
					if err != nil {
						return "", fmt.Errorf("error parsing URL params: %s", err)
					}
					args, err := paramsToArguments("Indices.Create", params)
					if err != nil {
						return "", fmt.Errorf("error converting params to arguments: %s", err)
					}
					fmt.Fprintf(&src, args)
				}
			} else {
				fmt.Fprintf(&src, "%q", matches[1])
			}

			src.WriteString(")")

			return src.String(), nil
		}},

	{ // ----- Get() or GetSource() ---------------------------------------------
		Pattern: `^GET /?\w+/(_doc|_source)/\w+`,
		Func: func(in string) (string, error) {
			var src strings.Builder

			re := regexp.MustCompile(`(?ms)^GET /?(?P<index>\w+)/(?P<api>_doc|_source)/(?P<id>\w+)(?P<params>\??\S+)?\s*$`)
			matches := re.FindStringSubmatch(in)
			if matches == nil {
				return "", errors.New("cannot match example source to pattern")
			}

			var apiName string
			switch matches[2] {
			case "_doc":
				apiName = "Get"
			case "_source":
				apiName = "GetSource"
			default:
				return "", fmt.Errorf("unknown API variant %q", matches[2])
			}

			if matches[4] == "" {
				fmt.Fprintf(&src, "\tres, err := es."+apiName+"(%q, %q, es."+apiName+".WithPretty()", matches[1], matches[3])
			} else {
				fmt.Fprintf(&src, "\tres, err := es."+apiName+"(\n\t%q,\n\t%q,\n\t", matches[1], matches[3])

				params, err := url.ParseQuery(strings.TrimPrefix(strings.TrimPrefix(matches[4], "/"), "?"))
				if err != nil {
					return "", fmt.Errorf("error parsing URL params: %s", err)
				}
				args, err := paramsToArguments(apiName, params)
				if err != nil {
					return "", fmt.Errorf("error converting params to arguments: %s", err)
				}
				fmt.Fprintf(&src, args)

				src.WriteString("\tes." + apiName + ".WithPretty(),\n")
			}

			src.WriteString(")")

			return src.String(), nil
		}},

	{ // ----- Exists() or ExistsSource() ---------------------------------------
		Pattern: `^HEAD /?\w+/(_doc|_source)/\w+`,
		Func: func(in string) (string, error) {
			var src strings.Builder

			re := regexp.MustCompile(`(?ms)^HEAD /?(?P<index>\w+)/(?P<api>_doc|_source)/(?P<id>\w+)(?P<params>\??[\S]+)?\s*$`)
			matches := re.FindStringSubmatch(in)
			if matches == nil {
				return "", errors.New("cannot match example source to pattern")
			}

			var apiName string
			switch matches[2] {
			case "_doc":
				apiName = "Exists"
			case "_source":
				apiName = "ExistsSource"
			default:
				return "", fmt.Errorf("unknown API variant %q", matches[2])
			}

			if matches[4] == "" {
				fmt.Fprintf(&src, "\tres, err := es."+apiName+"(%q, %q, es."+apiName+".WithPretty()", matches[1], matches[2])
			} else {
				fmt.Fprintf(&src, "\tres, err := es."+apiName+"(\n\t%q,\n\t%q,\n\t", matches[1], matches[2])
				params, err := url.ParseQuery(strings.TrimPrefix(strings.TrimPrefix(matches[4], "/"), "?"))
				if err != nil {
					return "", fmt.Errorf("error parsing URL params: %s", err)
				}
				args, err := paramsToArguments(apiName, params)
				if err != nil {
					return "", fmt.Errorf("error converting params to arguments: %s", err)
				}
				fmt.Fprintf(&src, args)

				src.WriteString("\tes." + apiName + ".WithPretty(),\n")
			}

			src.WriteString(")")

			return src.String(), nil
		}},

	{ // ----- Delete() ---------------------------------------------------------
		Pattern: `^DELETE /?\w+/_doc/\w+`,
		Func: func(in string) (string, error) {
			var src strings.Builder

			re := regexp.MustCompile(`(?ms)^DELETE /?(?P<index>\w+)/_doc/(?P<id>\w+)(?P<params>\??\S+)?\s*$`)
			matches := re.FindStringSubmatch(in)
			if matches == nil {
				return "", errors.New("cannot match example source to pattern")
			}

			fmt.Fprintf(&src, "\tres, err := es.Delete(")

			if matches[3] != "" {
				fmt.Fprintf(&src, "\t%q,\n\t%q,\n", matches[1], matches[2])
				params, err := url.ParseQuery(strings.TrimPrefix(strings.TrimPrefix(matches[3], "/"), "?"))
				if err != nil {
					return "", fmt.Errorf("error parsing URL params: %s", err)
				}
				args, err := paramsToArguments("Delete", params)
				if err != nil {
					return "", fmt.Errorf("error converting params to arguments: %s", err)
				}
				fmt.Fprintf(&src, args)

				src.WriteString("\tes.Delete.WithPretty(),\n")
			} else {
				fmt.Fprintf(&src, "\t%q, %q, es.Delete.WithPretty()", matches[1], matches[2])
			}
			src.WriteString(")")

			return src.String(), nil
		}},

	{ // ----- Search() ---------------------------------------------------------
		Pattern: `^GET /?(\w+)?/_search`,
		Func: func(in string) (string, error) {
			var src strings.Builder

			re := regexp.MustCompile(`(?ms)^GET /?(?P<index>\w+)?/_search(?P<params>\??[\S/]+)?\s?(?P<body>.+)?`)
			matches := re.FindStringSubmatch(in)
			if matches == nil {
				return "", errors.New("cannot match example source to pattern")
			}

			src.WriteString("\tres, err := es.Search(\n")
			if matches[1] != "" {
				fmt.Fprintf(&src, "\tes.Search.WithIndex(%q),\n", matches[1])
			}

			if matches[3] != "" {
				body, err := bodyStringToReader(matches[3])
				if err != nil {
					return "", fmt.Errorf("error converting body: %s", err)
				}
				fmt.Fprintf(&src, "\tes.Search.WithBody(%s),\n", body)
			}

			if matches[2] != "" {
				params, err := url.ParseQuery(strings.TrimPrefix(strings.TrimPrefix(matches[2], "/"), "?"))
				if err != nil {
					return "", fmt.Errorf("error parsing URL params: %s", err)
				}
				args, err := paramsToArguments("Search", params)
				if err != nil {
					return "", fmt.Errorf("error converting params to arguments: %s", err)
				}
				fmt.Fprintf(&src, args)
			}

			src.WriteString("\tes.Search.WithPretty(),\n")

			src.WriteString("\t)")

			return src.String(), nil
		}},
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
	Func    func(string) (string, error)
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

			cmds, err := t.Example.Commands()
			if err != nil {
				return "", fmt.Errorf("error getting example commands: %s", err)
			}

			out.WriteRune('\n')
			fmt.Fprintf(&out, "\t// tag:%s[]\n", t.Example.Digest)
			for i, c := range cmds {
				src, err := r.Func(c)
				if err != nil {
					return "", fmt.Errorf("error translating the example: %s", err)
				}

				if len(cmds) > 1 {
					out.WriteString("\t{\n")
				}
				out.WriteString(src)
				out.WriteRune('\n')
				out.WriteString("\tfmt.Println(res, err)\n")
				out.WriteString(testCheck)
				if len(cmds) > 1 {
					out.WriteString("\t}\n")
					if i != len(cmds)-1 {
						out.WriteString("\n")
					}
				}
			}
			fmt.Fprintf(&out, "\t// end:%s[]\n", t.Example.Digest)

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

// queryToParams extracts the URL params.
//
func queryToParams(input string) (url.Values, error) {
	input = strings.TrimPrefix(input, "/")
	input = strings.TrimPrefix(input, "?")
	return url.ParseQuery(input)
}

// paramsToArguments converts params to API arguments.
//
func paramsToArguments(api string, params url.Values) (string, error) {
	var b strings.Builder

	keys := make([]string, 0, len(params))
	for k := range params {
		keys = append(keys, k)
	}

	sort.Strings(keys)
	for _, k := range keys {
		var (
			name  string
			value string
		)

		value = strings.Join(params[k], ",")

		switch k {
		case "q":
			name = "Query"
		default:
			name = utils.NameToGo(k)
		}

		switch k {
		case "timeout": // duration
			dur, err := time.ParseDuration(params[k][0])
			if err != nil {
				return "", fmt.Errorf("error parsing duration: %s", err)
			}
			value = fmt.Sprintf("time.Duration(%d)", time.Duration(dur))
		case "from", "size", "terminate_after", "version": // numeric
			value = fmt.Sprintf("%s", value)
		default:
			value = strconv.Quote(value)
		}
		fmt.Fprintf(&b, "\tes.%s.With%s(%s),\n", api, name, value)
	}

	return b.String(), nil
}

// bodyStringToReader reformats input JSON string and returns it wrapped in strings.NewReader.
//
func bodyStringToReader(input string) (string, error) {
	var body bytes.Buffer
	err := json.Indent(&body, []byte(input), "\t\t", "  ")
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("strings.NewReader(`%s`)", strings.TrimRight(body.String(), "\n")), nil
}
