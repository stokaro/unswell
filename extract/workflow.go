package extract

import (
	"context"
	"fmt"
	"path"
	"strings"

	"go.yaml.in/yaml/v3"

	"github.com/stokaro/unswell/document"
)

type workflowShell struct {
	format document.Format
	reason string
}

func actionsWorkflow(name string) bool {
	name = path.Clean(strings.ReplaceAll(name, "\\", "/"))
	dir := path.Dir(name)
	return (dir == ".github/workflows" || strings.HasSuffix(dir, "/.github/workflows")) &&
		(path.Ext(name) == ".yml" || path.Ext(name) == ".yaml")
}

// Only the workflow's structural run fields select a program. Ordinary YAML
// values named run, including action inputs, retain their scalar semantics.
func markWorkflowRuns(ctx context.Context, root *yaml.Node, values map[yamlPosition]yamlValue) error {
	if root.Kind != yaml.DocumentNode || len(root.Content) != 1 {
		return nil
	}
	reader := workflowReader{}
	root = root.Content[0]
	jobs := reader.field(root, "jobs")
	if jobs == nil || jobs.Kind != yaml.MappingNode {
		return reader.err
	}
	for i := 1; i < len(jobs.Content); i += 2 {
		if err := ctx.Err(); err != nil {
			return err
		}
		reader.job(ctx, root, jobs.Content[i], values)
	}
	return reader.err
}

type workflowReader struct{ err error }

func (r *workflowReader) field(node *yaml.Node, keys ...string) *yaml.Node {
	for _, key := range keys {
		node = r.mappingField(node, key)
	}
	return node
}

func (r *workflowReader) mappingField(node *yaml.Node, key string) *yaml.Node {
	if node == nil || node.Kind != yaml.MappingNode {
		return nil
	}
	var found *yaml.Node
	for i := 0; i+1 < len(node.Content); i += 2 {
		if node.Content[i].Value != key || node.Content[i].Kind != yaml.ScalarNode {
			continue
		}
		if found != nil {
			r.err = fmt.Errorf("duplicate GitHub Actions field %q at line %d", key, node.Line)
			return nil
		}
		found = node.Content[i+1]
	}
	return found
}

func (r *workflowReader) job(ctx context.Context, root, job *yaml.Node, values map[yamlPosition]yamlValue) {
	steps := r.field(job, "steps")
	if steps == nil || steps.Kind != yaml.SequenceNode {
		return
	}
	for _, step := range steps.Content {
		if err := ctx.Err(); err != nil {
			r.err = err
			return
		}
		run := r.field(step, "run")
		if run == nil || run.Kind != yaml.ScalarNode || run.ShortTag() != "!!str" {
			continue
		}
		selection := r.shell(root, job, step)
		if strings.Contains(run.Value, "${{") {
			selection.reason = "actions-expression"
		}
		position := yamlPosition{run.Line, run.Column}
		value := values[position]
		value.program = &selection
		values[position] = value
	}
}

func (r *workflowReader) shell(root, job, step *yaml.Node) workflowShell {
	for _, node := range []*yaml.Node{r.field(step, "shell"), r.field(job, "defaults", "run", "shell"),
		r.field(root, "defaults", "run", "shell")} {
		if node != nil {
			return declaredWorkflowShell(node)
		}
	}
	if container := r.field(job, "container"); container != nil && container.ShortTag() != "!!null" {
		if container.Kind == yaml.ScalarNode && (container.ShortTag() != "!!str" ||
			container.Value == "" || strings.Contains(container.Value, "${{")) {
			return workflowShell{reason: "actions-shell-unknown"}
		}
		return workflowShell{format: document.Shell}
	}
	return runnerWorkflowShell(r.field(job, "runs-on"))
}
