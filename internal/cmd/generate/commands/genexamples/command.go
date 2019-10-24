// Licensed to Elasticsearch B.V. under one or more agreements.
// Elasticsearch B.V. licenses this file to you under the Apache 2.0 License.
// See the LICENSE file in the project root for more information.

package genexamples

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
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
			Input:       *input,
			Output:      *output,
			DebugSource: *debug,
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
	Input       string
	Output      string
	DebugSource bool
}

// Execute runs the command.
//
func (cmd *Command) Execute() (err error) {
	var (
		processed int
		skipped   int
		start     = time.Now()
	)

	if cmd.Output != "-" {
		outputDir := filepath.Join(cmd.Output, "doc")
		if err := os.MkdirAll(outputDir, 0775); err != nil {
			return fmt.Errorf("error creating output directory %q: %s", outputDir, err)
		}
	}

	f, err := os.Open(cmd.Input)
	if err != nil {
		return fmt.Errorf("error reading input: %s", err)
	}
	defer f.Close()

	var examples []Example
	if err := json.NewDecoder(f).Decode(&examples); err != nil {
		return fmt.Errorf("error decoding input: %s", err)
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
	var out io.Reader

	fName := filepath.Join(cmd.Output, "doc", fmt.Sprintf("%s.asciidoc", e.Digest))
	out = e.Output()

	if cmd.DebugSource {
		var (
			err error
			buf bytes.Buffer
			tee = io.TeeReader(out, &buf)
		)

		if utils.IsTTY() {
			fmt.Fprint(os.Stderr, "\x1b[2m")
		}
		fmt.Fprintln(os.Stderr, strings.Repeat("━", utils.TerminalWidth()))
		if utils.IsTTY() {
			fmt.Fprint(os.Stderr, "\x1b[0m")
		}

		if _, err = io.Copy(os.Stderr, tee); err != nil {
			return fmt.Errorf("error copying output: %s", err)
		}

		fmt.Fprintf(os.Stderr, "\n\n")

		out = &buf
	}

	if cmd.Output == "-" {
		if _, err := io.Copy(os.Stdout, out); err != nil {
			return fmt.Errorf("error copying output: %s", err)
		}
	} else {
		f, err := os.OpenFile(fName, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
		if err != nil {
			return fmt.Errorf("error creating file: %s", err)
		}
		if _, err = io.Copy(f, out); err != nil {
			return fmt.Errorf("error copying output: %s", err)
		}
		if err := f.Close(); err != nil {
			return fmt.Errorf("error closing file: %s", err)
		}
	}

	return nil
}
