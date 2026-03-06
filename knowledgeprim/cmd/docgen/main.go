// Command docgen extracts knowledgeprim command metadata and writes it as JSON to stdout.
// Used by scripts/docgen.sh to generate the commands table in docs/agent-reference.md.
// Local types mirror primkit/docgen.{Prim,Cmd,Flag}Meta — JSON field names must match.
package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/propifly/primkit/knowledgeprim/internal/cli"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

type primMeta struct {
	Name     string    `json:"name"`
	Commands []cmdMeta `json:"commands"`
}

type cmdMeta struct {
	Name     string     `json:"name"`
	Synopsis string     `json:"synopsis"`
	Short    string     `json:"short"`
	Flags    []flagMeta `json:"flags"`
}

type flagMeta struct {
	Name     string `json:"name"`
	Usage    string `json:"usage"`
	Default  string `json:"default"`
	Required bool   `json:"required"`
}

var skipCommands = map[string]bool{
	"serve":      true,
	"mcp":        true,
	"completion": true,
	"help":       true,
}

func main() {
	root := cli.NewRootCmd()
	meta := primMeta{
		Name:     "knowledgeprim",
		Commands: extractCommands(root.Commands(), ""),
	}
	if err := json.NewEncoder(os.Stdout).Encode(meta); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func extractCommands(cmds []*cobra.Command, prefix string) []cmdMeta {
	var result []cmdMeta
	for _, cmd := range cmds {
		if skipCommands[cmd.Name()] || cmd.Hidden {
			continue
		}
		fullName := cmd.Name()
		if prefix != "" {
			fullName = prefix + " " + cmd.Name()
		}
		synopsis := buildSynopsis(cmd.Use, fullName)
		if cmd.RunE != nil || cmd.Run != nil {
			m := cmdMeta{
				Name:     fullName,
				Synopsis: synopsis,
				Short:    cmd.Short,
			}
			cmd.Flags().VisitAll(func(f *pflag.Flag) {
				if f.Hidden {
					return
				}
				required := len(f.Annotations[cobra.BashCompOneRequiredFlag]) > 0
				m.Flags = append(m.Flags, flagMeta{
					Name:     f.Name,
					Usage:    f.Usage,
					Default:  f.DefValue,
					Required: required,
				})
			})
			result = append(result, m)
		}
		result = append(result, extractCommands(cmd.Commands(), fullName)...)
	}
	return result
}

func buildSynopsis(use, fullName string) string {
	parts := strings.SplitN(use, " ", 2)
	if len(parts) <= 1 || parts[1] == "" {
		return fullName
	}
	return fullName + " " + parts[1]
}
