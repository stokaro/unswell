package generation

import (
	"crypto/sha256"
	"fmt"
	"regexp"
	"slices"
	"strconv"
	"strings"
)

// RequestsVersion identifies the request list a generator receives.
const RequestsVersion = "unswell-requests-v1"

// Operations of the pilot with their fixed opening lines from
// research/methods/prompts/README.md.
var openingLines = map[string]string{
	"generate": "Write the documentation described below.",
	"polish":   "Edit the text below for clarity and correctness. Keep its facts, structure, and length.",
}

// Operations lists the pilot operations in a stable order.
func Operations() []string { return []string{"generate", "polish"} }

// PromptCondition is one frozen prompt file with its shared body.
type PromptCondition struct {
	ID     string `json:"id"`
	File   string `json:"file"`
	SHA256 string `json:"sha256"`
	Body   string `json:"body"`
}

// Request is one generation call: the full text the generator receives and
// the material it embeds.
type Request struct {
	ID             string `json:"id"`
	TaskID         string `json:"task_id"`
	Operation      string `json:"operation"`
	Prompt         string `json:"prompt"`
	RequestedWords int    `json:"requested_words"`
	Text           string `json:"text"`
	Material       string `json:"material"`
	InputSHA256    string `json:"input_sha256"`
}

// HarnessInstruction is the one line an agent harness adds after the
// request, so the agent answers from the material alone instead of reading
// files or running tools. It is recorded as its own message.
const HarnessInstruction = "Do not read any files or run any tools; return the document text as your final answer."

// Requests is the request list of one run. Delivery names how the material
// reaches the generator: the protocol's separate message, or one message
// with the material after a separator when the harness allows one only.
type Requests struct {
	Version     string            `json:"version"`
	Protocol    string            `json:"protocol"`
	Run         string            `json:"run"`
	TasksSHA256 string            `json:"tasks_sha256"`
	Delivery    string            `json:"delivery"`
	Harness     string            `json:"harness_instruction"`
	Prompts     []PromptCondition `json:"prompts"`
	Operations  []string          `json:"operations"`
	Requests    []Request         `json:"requests"`
}

var fencedBody = regexp.MustCompile("(?s)```text\n(.*?)```")

// ParsePrompt reads the shared body of a frozen prompt file: the text inside
// its fenced block, with the placeholder left in place.
func ParsePrompt(id, file string, data []byte) (PromptCondition, error) {
	match := fencedBody.FindSubmatch(data)
	if match == nil || !strings.Contains(string(match[1]), "{words}") {
		return PromptCondition{}, fmt.Errorf("prompt %s has no fenced body with a {words} placeholder", file)
	}
	return PromptCondition{ID: id, File: file, SHA256: fmt.Sprintf("%x", sha256.Sum256(data)),
		Body: strings.TrimSpace(string(match[1]))}, nil
}

// BuildRequests pairs every task with every operation and prompt. The
// material of generate is the fact sheet; the material of polish is the
// original text. The requested length is the original's word count.
func BuildRequests(run string, tasks Tasks, tasksSHA256 string, prompts []PromptCondition, operations []string) (Requests, error) {
	if run == "" || len(prompts) == 0 || len(operations) == 0 || len(tasks.Tasks) == 0 {
		return Requests{}, fmt.Errorf("requests need a run, prompts, operations, and tasks")
	}
	for _, operation := range operations {
		if _, known := openingLines[operation]; !known {
			return Requests{}, fmt.Errorf("unknown operation %q", operation)
		}
	}
	result := Requests{Version: RequestsVersion, Protocol: tasks.Protocol, Run: run, TasksSHA256: tasksSHA256,
		Delivery: "single-message", Harness: HarnessInstruction, Prompts: prompts, Operations: slices.Clone(operations),
		Requests: []Request{}}
	for _, task := range tasks.Tasks {
		for _, operation := range operations {
			for _, prompt := range prompts {
				result.Requests = append(result.Requests, buildRequest(task, operation, prompt))
			}
		}
	}
	return result, nil
}

func buildRequest(task Task, operation string, prompt PromptCondition) Request {
	material := task.FactSheet.Text()
	if operation == "polish" {
		material = task.Text
	}
	body := strings.ReplaceAll(prompt.Body, "{words}", strconv.Itoa(task.Words))
	text := openingLines[operation] + "\n\n" + body + "\n\n---\n\n" + material
	id := sha256.Sum256([]byte("request\x00" + task.ID + "\x00" + operation + "\x00" + prompt.ID))
	return Request{ID: fmt.Sprintf("r%x", id[:8]), TaskID: task.ID, Operation: operation, Prompt: prompt.ID,
		RequestedWords: task.Words, Text: text, Material: material, InputSHA256: fmt.Sprintf("%x", sha256.Sum256([]byte(text)))}
}
