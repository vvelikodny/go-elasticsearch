// Licensed to Elasticsearch B.V. under one or more agreements.
// Elasticsearch B.V. licenses this file to you under the Apache 2.0 License.
// See the LICENSE file in the project root for more information.

package genexamples

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/elastic/go-elasticsearch/v8/internal/cmd/generate/commands"
	"github.com/elastic/go-elasticsearch/v8/internal/cmd/generate/utils"
)

var (
	input  *string
	output *string
	color  *bool
	debug  *bool
)

func init() {
	input = genexamplesCmd.Flags().StringP("input", "i", "", "Path to a file with specification for examples")
	output = genexamplesCmd.Flags().StringP("output", "o", "", "Path to a folder for generated output")
	color = genexamplesCmd.Flags().BoolP("color", "c", true, "Syntax highlight the debug output")
	debug = genexamplesCmd.Flags().BoolP("debug", "d", false, "Print the generated source to terminal")

	genexamplesCmd.MarkFlagRequired("input")
	genexamplesCmd.MarkFlagRequired("output")
	genexamplesCmd.Flags().SortFlags = false

	commands.RegisterCmd(genexamplesCmd)
}

var genexamplesCmd = &cobra.Command{
	Use:   "examples",
	Short: "Generate the Go examples for documentation",
	Run: func(cmd *cobra.Command, args []string) {
		command := &Command{
			Input:          *input,
			Output:         *output,
			DebugSource:    *debug,
			ColorizeSource: *color,
		}
		err := command.Execute()
		if err != nil {
			utils.PrintErr(err)
			os.Exit(1)
		}
	},
}

// Command represents the "genexamples" command.
//
type Command struct {
	Input          string
	Output         string
	DebugSource    bool
	ColorizeSource bool
}

// Execute runs the command.
//
func (cmd *Command) Execute() (err error) {
	var (
		processed int
		skipped   int
		start     = time.Now()
	)

	f, err := os.Open(cmd.Input)
	if err != nil {
		return err
	}
	defer f.Close()

	var examples []Example
	if err := json.NewDecoder(f).Decode(&examples); err != nil {
		return err
	}

	for _, e := range examples {
		if e.Enabled() {
			if utils.IsTTY() {
				fmt.Fprint(os.Stderr, "\x1b[2m")
			}
			fmt.Fprintln(os.Stderr, strings.Repeat("━", utils.TerminalWidth()))
			fmt.Fprintf(os.Stderr, "Processing example %q @ %s\n", e.ID(), e.Digest)
			if utils.IsTTY() {
				fmt.Fprint(os.Stderr, "\x1b[0m")
			}
			if err := cmd.processExample(e); err != nil {
				return fmt.Errorf("error processing example %s: %v", e.ID(), err)
			}
			processed++
		} else {
			skipped++
		}
	}

	if utils.IsTTY() {
		fmt.Fprint(os.Stderr, "\x1b[2m")
	}
	fmt.Fprintln(os.Stderr, strings.Repeat("━", utils.TerminalWidth()))
	fmt.Fprintf(os.Stderr, "Processed %d examples, skipped %d examples in %s\n", processed, skipped, time.Since(start).Truncate(time.Millisecond))
	if utils.IsTTY() {
		fmt.Fprint(os.Stderr, "\x1b[0m")
	}

	return nil
}

func (cmd *Command) processExample(e Example) error {
	fmt.Printf("%+v\n\n", e.Source)
	return nil
}
