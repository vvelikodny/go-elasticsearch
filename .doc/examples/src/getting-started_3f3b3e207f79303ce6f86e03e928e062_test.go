// Licensed to Elasticsearch B.V under one or more agreements.
// Elasticsearch B.V. licenses this file to you under the Apache 2.0 License.
// See the LICENSE file in the project root for more information.
//
// Code generated, DO NOT EDIT

package elasticsearch_test

import (
	"fmt"
	"os"
	"testing"

	"github.com/elastic/go-elasticsearch/v8"
)

var (
	_ = fmt.Printf
	_ = os.Stdout
	_ = elasticsearch.NewDefaultClient
)

// <https://github.com/elastic/elasticsearch/blob/master/docs/reference/getting-started.asciidoc#L253>
//
// --------------------------------------------------------------------------------
// GET /customer/_doc/1
// --------------------------------------------------------------------------------

func Test_getting_started_3f3b3e207f79303ce6f86e03e928e062(t *testing.T) {
	es, _ := elasticsearch.NewDefaultClient()

	// tag:3f3b3e207f79303ce6f86e03e928e062[]
	res, err := es.Get("customer", "1", es.Get.WithPretty())
	// end:3f3b3e207f79303ce6f86e03e928e062[]
	if err != nil {
		fmt.Println("Error getting the response:", err)
		os.Exit(1)
	}
	defer res.Body.Close()
	fmt.Println(res)
}
