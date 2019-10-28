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

const tail = "\t" + `if err != nil {
		fmt.Println("Error getting the response:", err)
		os.Exit(1)
	}
	defer res.Body.Close()
	fmt.Println(res)
`

// ConsoleToGo contains translation rules.
//
var ConsoleToGo = []TranslateRule{

	{ // ----- Info() -----------------------------------------------------------
		Pattern: "^GET /$",
		Func: func(e Example) (string, error) {
			return "res, err := es.Info()", nil
		}},

	{ // ----- Cat.Health() -----------------------------------------------------
		Pattern: `^GET /_cat/health\?v`,
		Func: func(e Example) (string, error) {
			return "\tres, err := es.Cat.Health(es.Cat.Health.WithV(true))", nil
		}},

	{ // ----- Index() ---------------------------------------------------------
		Pattern: `^PUT /?\w+/_doc/\w+`,
		Func: func(e Example) (string, error) {
			var src strings.Builder

			re := regexp.MustCompile(`(?ms)^PUT /?(?P<index>\w+)/_doc/(?P<id>\w+)(?P<params>\??[\S]+)?\s(?P<body>.*)`)
			matches := re.FindStringSubmatch(e.Source)
			if matches == nil {
				return "", errors.New("cannot match example source to pattern")
			}

			src.WriteString("\tres, err := es.Index(\n")

			fmt.Fprintf(&src, "\t%q,\n", matches[1])

			body, _ := bodyStringToReader(matches[4])
			fmt.Fprintf(&src, "\t%s,\n", body)

			fmt.Fprintf(&src, "\tes.Index.WithDocumentID(%q),\n", matches[2])

			if matches[3] != "" {
				params, err := queryToParams(matches[3])
				if err != nil {
					return "", fmt.Errorf("error parsing URL params: %s", err)
				}
				args, err := paramsToArguments("Index", params)
				if err != nil {
					return "", fmt.Errorf("error converting params to arguments: %s", err)
				}
				fmt.Fprintf(&src, args)
			}

			src.WriteString("\tes.Index.WithPretty(),\n")
			src.WriteString("\t)")

			return src.String(), nil
		}},

	{ // ----- Indices.Create() -------------------------------------------------
		Pattern: `^PUT /?[\S]+\s?(?P<body>.+)?`,
		Func: func(e Example) (string, error) {
			var src strings.Builder

			re := regexp.MustCompile(`(?ms)^PUT /?(?P<index>[\S]+)(?P<params>\??[\S/]+)?\s?(?P<body>.+)?`)
			matches := re.FindStringSubmatch(e.Source)
			if matches == nil {
				return "", errors.New("cannot match example source to pattern")
			}

			src.WriteString("\tres, err := es.Indices.Create(")
			if matches[2] != "" || matches[3] != "" {
				fmt.Fprintf(&src, "\n\t%q,\n", matches[1])

				if matches[3] != "" {
					body, _ := bodyStringToReader(matches[3])
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
		Func: func(e Example) (string, error) {
			var src strings.Builder

			re := regexp.MustCompile(`(?ms)^GET /?(?P<index>\w+)/(?P<api>_doc|_source)/(?P<id>\w+)(?P<params>\??\S+)?\s*$`)
			matches := re.FindStringSubmatch(e.Source)
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
		Func: func(e Example) (string, error) {
			var src strings.Builder

			re := regexp.MustCompile(`(?ms)^HEAD /?(?P<index>\w+)/(?P<api>_doc|_source)/(?P<id>\w+)(?P<params>\??[\S]+)?\s*$`)
			matches := re.FindStringSubmatch(e.Source)
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
		Func: func(e Example) (string, error) {
			var src strings.Builder

			re := regexp.MustCompile(`(?ms)^DELETE /?(?P<index>\w+)/_doc/(?P<id>\w+)(?P<params>\??\S+)?\s*$`)
			matches := re.FindStringSubmatch(e.Source)
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
		Pattern: `^GET /\w+/_search`,
		Func: func(e Example) (string, error) {
			var src strings.Builder

			re := regexp.MustCompile(`(?ms)^GET /(?P<index>\w+)/_search(?P<params>\??[\S/]+)?\s?(?P<body>.+)?`)
			matches := re.FindStringSubmatch(e.Source)
			if matches == nil {
				return "", errors.New("cannot match example source to pattern")
			}

			src.WriteString("\tres, err := es.Search(\n")
			fmt.Fprintf(&src, "\tes.Search.WithIndex(%q),\n", matches[1])

			body, _ := bodyStringToReader(matches[3])
			fmt.Fprintf(&src, "\tes.Search.WithBody(%s),\n", body)

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
	Func    func(Example) (string, error)
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

			src, err := r.Func(t.Example)
			if err != nil {
				return "", fmt.Errorf("error translating the example: %s", err)
			}

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
		val := strings.Join(params[k], ",")
		if k == "timeout" {
			dur, err := time.ParseDuration(params[k][0])
			if err != nil {
				return "", fmt.Errorf("error parsing duration: %s", err)
			}
			val = fmt.Sprintf("time.Duration(%d)", time.Duration(dur))
		} else {
			val = strconv.Quote(val)
		}
		fmt.Fprintf(&b, "\tes.%s.With%s(%s),\n", api, utils.NameToGo(k), val)
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
	return fmt.Sprintf("strings.NewReader(`%s`)", body.String()), nil
}
