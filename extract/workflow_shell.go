package extract

import (
	"regexp"
	"strings"

	"go.yaml.in/yaml/v3"

	"github.com/stokaro/unswell/document"
)

var workflowRunner = regexp.MustCompile(`^(ubuntu-(latest|[0-9]{2}\.[0-9]{2})(-arm)?|` +
	`macos-(latest|[0-9]+)(-large|-xlarge|-intel)?|windows-(latest|[0-9]{4})(-arm)?)$`)
var workflowShellOption = regexp.MustCompile(`^[-A-Za-z0-9_./=:]+$`)

func declaredWorkflowShell(node *yaml.Node) workflowShell {
	unknown := workflowShell{reason: "actions-shell-unknown"}
	if node.Kind != yaml.ScalarNode || node.ShortTag() != "!!str" {
		return unknown
	}
	words := strings.Fields(node.Value)
	if len(words) == 0 || !workflowShellTemplate(words) {
		return unknown
	}
	formats := map[string]document.Format{
		"bash": document.Bash, "sh": document.Shell, "zsh": document.Zsh, "fish": document.Fish,
		"pwsh": document.PowerShell, "powershell": document.PowerShell,
	}
	if format, ok := formats[words[0]]; ok {
		return workflowShell{format: format}
	}
	return unknown
}

func workflowShellTemplate(words []string) bool {
	if len(words) == 1 {
		return true
	}
	placeholders := 0
	for _, word := range words[1:] {
		if word == "{0}" {
			placeholders++
		} else if !workflowShellOption.MatchString(word) {
			return false
		}
	}
	return placeholders == 1
}

func runnerWorkflowShell(node *yaml.Node) workflowShell {
	unknown := workflowShell{reason: "actions-shell-unknown"}
	if node == nil || node.Kind != yaml.ScalarNode || node.ShortTag() != "!!str" || !workflowRunner.MatchString(node.Value) {
		return unknown
	}
	if strings.HasPrefix(node.Value, "windows-") {
		return workflowShell{format: document.PowerShell}
	}
	return workflowShell{format: document.Bash}
}
