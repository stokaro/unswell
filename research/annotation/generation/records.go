package generation

import (
	"crypto/sha256"
	"fmt"
	"math"
	"regexp"
	"slices"
	"strings"
)

// ResponsesVersion identifies the raw responses a run brought back.
const ResponsesVersion = "unswell-responses-v1"

// GenerationVersion identifies the saved generation records.
const GenerationVersion = "unswell-generation-v1"

// Statuses a response may carry.
var statuses = []string{"complete", "refused", "truncated", "error"}

// Response is one raw response as the run tooling saved it.
type Response struct {
	RequestID      string `json:"request_id"`
	Text           string `json:"text"`
	Status         string `json:"status"`
	Note           string `json:"note,omitempty"`
	DurationMillis int    `json:"duration_ms,omitempty"`
	Tokens         int    `json:"tokens,omitempty"`
	// RemoteRequestID is the identifier the generator's endpoint returned for
	// this response, when the harness had one; the record says unavailable
	// otherwise.
	RemoteRequestID string `json:"remote_request_id,omitempty"`
}

// Responses is the run's raw response set with the identity the harness
// could supply. A field the harness cannot supply says "unavailable".
type Responses struct {
	Version     string     `json:"version"`
	Run         string     `json:"run"`
	Family      string     `json:"family"`
	Model       string     `json:"model"`
	ModelBasis  string     `json:"model_basis"`
	Harness     string     `json:"harness"`
	AgentType   string     `json:"agent_type"`
	GeneratedOn string     `json:"generated_on"`
	Parameters  string     `json:"parameters"`
	Responses   []Response `json:"responses"`
}

// Message is one visible input message.
type Message struct {
	Role string `json:"role"`
	Text string `json:"text"`
}

// Comparability reports which facts of the sheet the response carries.
type Comparability struct {
	IdentifiersCovered int  `json:"identifiers_covered"`
	IdentifiersTotal   int  `json:"identifiers_total"`
	NumbersCovered     int  `json:"numbers_covered"`
	NumbersTotal       int  `json:"numbers_total"`
	SignaturePresent   bool `json:"signature_present"`
}

// Record is one saved generation with everything the protocol asks for. A
// field the harness cannot supply says "unavailable" rather than a guess.
type Record struct {
	ResponseID      string        `json:"response_id"`
	TaskID          string        `json:"task_id"`
	Operation       string        `json:"operation"`
	Prompt          string        `json:"prompt"`
	PromptSHA256    string        `json:"prompt_sha256"`
	Family          string        `json:"family"`
	Model           string        `json:"model"`
	ModelBasis      string        `json:"model_basis"`
	GeneratedOn     string        `json:"generated_on"`
	Parameters      string        `json:"parameters"`
	RemoteRequestID string        `json:"remote_request_id"`
	Cost            string        `json:"cost"`
	Input           []Message     `json:"input"`
	InputSHA256     string        `json:"input_sha256"`
	RawText         string        `json:"raw_text"`
	Transformations []string      `json:"transformations"`
	Text            string        `json:"text"`
	OutputSHA256    string        `json:"output_sha256"`
	Status          string        `json:"status"`
	Note            string        `json:"note,omitempty"`
	RequestedWords  int           `json:"requested_words"`
	RealizedWords   int           `json:"realized_words"`
	LengthRatio     float64       `json:"length_ratio"`
	OffLength       bool          `json:"off_length"`
	Overlap         float64       `json:"overlap"`
	OverlapHigh     bool          `json:"overlap_high"`
	Comparability   Comparability `json:"comparability"`
	Contamination   string        `json:"contamination"`
	Path            string        `json:"path,omitempty"`
}

// Coverage counts the run's responses by outcome.
type Coverage struct {
	Requests    int `json:"requests"`
	Complete    int `json:"complete"`
	Refused     int `json:"refused"`
	Truncated   int `json:"truncated"`
	Error       int `json:"error"`
	Missing     int `json:"missing"`
	OffLength   int `json:"off_length"`
	OverlapHigh int `json:"overlap_high"`
}

// Generation is the saved record set of one run.
type Generation struct {
	Version        string   `json:"version"`
	Run            string   `json:"run"`
	Protocol       string   `json:"protocol"`
	TasksSHA256    string   `json:"tasks_sha256"`
	RequestsSHA256 string   `json:"requests_sha256"`
	Family         string   `json:"family"`
	Model          string   `json:"model"`
	ModelBasis     string   `json:"model_basis"`
	Harness        string   `json:"harness"`
	AgentType      string   `json:"agent_type"`
	Delivery       string   `json:"delivery"`
	WordCount      string   `json:"word_count"`
	Coverage       Coverage `json:"coverage"`
	Records        []Record `json:"records"`
}

// Length tolerance and overlap threshold fixed by the protocol.
const (
	lengthTolerance  = 0.30
	overlapThreshold = 0.5
	overlapGram      = 8
)

var wordPattern = regexp.MustCompile(`[A-Za-z0-9_]+`)

// BuildRecords joins the run's responses with its requests and tasks. Every
// request must have exactly one response or is counted as missing; a
// response naming no request is an error.
func BuildRecords(requests Requests, requestsSHA256 string, tasks Tasks, responses Responses) (Generation, error) {
	if err := checkResponses(requests, responses); err != nil {
		return Generation{}, err
	}
	byTask := map[string]Task{}
	for _, task := range tasks.Tasks {
		byTask[task.ID] = task
	}
	byRequest := map[string]Request{}
	for _, request := range requests.Requests {
		byRequest[request.ID] = request
	}
	prompts := map[string]PromptCondition{}
	for _, prompt := range requests.Prompts {
		prompts[prompt.ID] = prompt
	}
	seen := map[string]bool{}
	result := Generation{Version: GenerationVersion, Run: requests.Run, Protocol: requests.Protocol,
		TasksSHA256: requests.TasksSHA256, RequestsSHA256: requestsSHA256, Family: responses.Family, Model: responses.Model,
		ModelBasis: responses.ModelBasis, Harness: responses.Harness, AgentType: responses.AgentType,
		Delivery: requests.Delivery, WordCount: "whitespace-fields-v1", Records: []Record{}}
	result.Coverage.Requests = len(requests.Requests)
	for _, response := range responses.Responses {
		request, found := byRequest[response.RequestID]
		if !found || seen[response.RequestID] {
			return Generation{}, fmt.Errorf("response %q names no request of the run or repeats one", response.RequestID)
		}
		if !slices.Contains(statuses, response.Status) {
			return Generation{}, fmt.Errorf("response %q has unknown status %q", response.RequestID, response.Status)
		}
		seen[response.RequestID] = true
		record := buildRecord(request, byTask[request.TaskID], prompts[request.Prompt], responses, response)
		result.Coverage.count(record)
		result.Records = append(result.Records, record)
	}
	result.Coverage.Missing = len(requests.Requests) - len(seen)
	slices.SortFunc(result.Records, func(a, b Record) int { return strings.Compare(a.ResponseID, b.ResponseID) })
	return result, nil
}

func checkResponses(requests Requests, responses Responses) error {
	if responses.Version != ResponsesVersion || responses.Run != requests.Run {
		return fmt.Errorf("responses must carry version %s for run %s", ResponsesVersion, requests.Run)
	}
	for _, field := range []string{responses.Family, responses.Model, responses.ModelBasis, responses.Harness,
		responses.AgentType, responses.GeneratedOn, responses.Parameters} {
		if strings.TrimSpace(field) == "" {
			return fmt.Errorf("responses must name the family, model, model basis, harness, agent type, date, and parameters")
		}
	}
	return nil
}

func (c *Coverage) count(record Record) {
	switch record.Status {
	case "complete":
		c.Complete++
	case "refused":
		c.Refused++
	case "truncated":
		c.Truncated++
	default:
		c.Error++
	}
	if record.Status == "complete" && record.OffLength {
		c.OffLength++
	}
	if record.Status == "complete" && record.OverlapHigh {
		c.OverlapHigh++
	}
}

// remoteRequestID returns the endpoint's identifier for a response, or the
// protocol's word for a harness that has none.
func remoteRequestID(response Response) string {
	if response.RemoteRequestID == "" {
		return "unavailable"
	}
	return response.RemoteRequestID
}

func buildRecord(request Request, task Task, prompt PromptCondition, run Responses, response Response) Record {
	text := strings.TrimSpace(response.Text)
	record := Record{ResponseID: request.ID, TaskID: task.ID, Operation: request.Operation, Prompt: request.Prompt,
		PromptSHA256: prompt.SHA256, Family: run.Family, Model: run.Model, ModelBasis: run.ModelBasis,
		GeneratedOn: run.GeneratedOn, Parameters: run.Parameters, RemoteRequestID: remoteRequestID(response), Cost: "unavailable",
		Input:       []Message{{Role: "user", Text: request.Text}, {Role: "harness", Text: HarnessInstruction}},
		InputSHA256: request.InputSHA256, RawText: response.Text,
		Transformations: []string{"trim-whitespace-v1"}, Text: text,
		OutputSHA256: fmt.Sprintf("%x", sha256.Sum256([]byte(text))), Status: response.Status, Note: response.Note,
		RequestedWords: request.RequestedWords, RealizedWords: len(strings.Fields(text)), Contamination: "unknown"}
	if request.RequestedWords > 0 {
		record.LengthRatio = math.Round(float64(record.RealizedWords)/float64(request.RequestedWords)*1000) / 1000
		record.OffLength = math.Abs(record.LengthRatio-1) > lengthTolerance
	}
	record.Overlap = overlap(text, task.Text)
	record.OverlapHigh = record.Overlap > overlapThreshold
	record.Comparability = comparability(text, task.FactSheet)
	return record
}

// overlap is the share of response tokens inside word 8-gram matches with
// the source text, on lowercased alphanumeric tokens.
func overlap(response, source string) float64 {
	sourceTokens := tokens(source)
	responseTokens := tokens(response)
	if len(responseTokens) == 0 || len(sourceTokens) < overlapGram {
		return 0
	}
	grams := map[string]bool{}
	for i := 0; i+overlapGram <= len(sourceTokens); i++ {
		grams[strings.Join(sourceTokens[i:i+overlapGram], " ")] = true
	}
	covered := make([]bool, len(responseTokens))
	for i := 0; i+overlapGram <= len(responseTokens); i++ {
		if grams[strings.Join(responseTokens[i:i+overlapGram], " ")] {
			for j := i; j < i+overlapGram; j++ {
				covered[j] = true
			}
		}
	}
	count := 0
	for _, hit := range covered {
		if hit {
			count++
		}
	}
	return math.Round(float64(count)/float64(len(responseTokens))*1000) / 1000
}

func tokens(text string) []string {
	return wordPattern.FindAllString(strings.ToLower(text), -1)
}

func comparability(text string, sheet FactSheet) Comparability {
	result := Comparability{IdentifiersTotal: len(sheet.Identifiers), NumbersTotal: len(sheet.Numbers),
		SignaturePresent: sheet.Signature != "" && strings.Contains(text, strings.TrimSpace(sheet.Signature))}
	for _, name := range sheet.Identifiers {
		if strings.Contains(text, name) {
			result.IdentifiersCovered++
		}
	}
	for _, number := range sheet.Numbers {
		if strings.Contains(text, number) {
			result.NumbersCovered++
		}
	}
	return result
}
