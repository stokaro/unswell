package cli

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"
)

// maxAnswerBytes bounds one model answer, over either transport.
const maxAnswerBytes = 8 << 20

// answerOnly is sent as the system message of a chat request.
const answerOnly = "You answer with the requested JSON object and nothing else."

// modelOptions names one model. Either an endpoint with a model name or a
// local command is set, never both.
type modelOptions struct {
	endpoint    string
	model       string
	apiKeyEnv   string
	maxTokens   int
	command     string
	commandArgs []string
}

func (o modelOptions) empty() bool {
	return o.endpoint == "" && o.model == "" && o.apiKeyEnv == "" && o.maxTokens == 0 && o.command == "" && len(o.commandArgs) == 0
}

// modelIdentity says who judged, without credentials or a full address.
type modelIdentity struct {
	Transport string `json:"transport"`
	Model     string `json:"model,omitempty"`
	Host      string `json:"host,omitempty"`
	Command   string `json:"command,omitempty"`
}

func (i modelIdentity) String() string {
	if i.Transport == "command" {
		return "command " + i.Command
	}
	return i.Model + " at " + i.Host
}

// modelClient asks one model one question and returns its text answer.
type modelClient interface {
	ask(ctx context.Context, prompt string) (string, error)
	identity() modelIdentity
}

// permanentError marks a failure that a repeated call cannot fix.
type permanentError struct{ err error }

func (e permanentError) Error() string { return e.err.Error() }
func (e permanentError) Unwrap() error { return e.err }

// newModelClient validates one model's options and builds its client. The
// role names the flags in error messages: "" for the reviewer, "confirm-" for
// the confirming model.
func newModelClient(environment Environment, options modelOptions, role string, insecure bool) (modelClient, error) {
	switch {
	case options.command != "" && (options.endpoint != "" || options.model != "" || options.apiKeyEnv != "" || options.maxTokens != 0):
		return nil, fmt.Errorf("--%scommand cannot be combined with --%sendpoint, --%smodel, --%sapi-key-env, or --%smax-tokens",
			role, role, role, role, role)
	case options.command != "":
		return newCommandClient(environment, options)
	case len(options.commandArgs) > 0:
		return nil, fmt.Errorf("--%scommand-arg requires --%scommand", role, role)
	case options.endpoint == "" || options.model == "":
		return nil, fmt.Errorf("name a model with --%sendpoint and --%smodel, or a local program with --%scommand", role, role, role)
	default:
		return newEndpointClient(environment, options, role, insecure)
	}
}

type endpointClient struct {
	address   string
	host      string
	model     string
	key       string
	maxTokens int
	client    *http.Client
}

func newEndpointClient(environment Environment, options modelOptions, role string, insecure bool) (modelClient, error) {
	address, err := endpointAddress(options.endpoint, role, insecure)
	if err != nil {
		return nil, err
	}
	if options.maxTokens < 0 {
		return nil, fmt.Errorf("--%smax-tokens cannot be negative", role)
	}
	key := ""
	if options.apiKeyEnv != "" {
		if key = environment.getenv(options.apiKeyEnv); key == "" {
			return nil, fmt.Errorf("environment variable %s named by --%sapi-key-env is empty", options.apiKeyEnv, role)
		}
	}
	// An API endpoint has no reason to redirect, and a redirect could carry
	// the key and the document to another address.
	client := &http.Client{CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }}
	return &endpointClient{address: address.String(), host: address.Host, model: options.model, key: key,
		maxTokens: options.maxTokens, client: client}, nil
}

// endpointAddress accepts https anywhere and http on this machine. Plain http
// to another host needs an explicit flag, because the document and the key
// would travel unencrypted.
func endpointAddress(endpoint, role string, insecure bool) (*url.URL, error) {
	address, err := url.Parse(endpoint)
	if err != nil || address.Host == "" || (address.Scheme != "https" && address.Scheme != "http") {
		return nil, fmt.Errorf("--%sendpoint must be an http or https URL of a chat completions endpoint", role)
	}
	if address.User != nil {
		return nil, fmt.Errorf("--%sendpoint must not carry credentials; name a variable with --%sapi-key-env", role, role)
	}
	if address.Scheme == "http" && !loopback(address.Hostname()) && !insecure {
		return nil, fmt.Errorf("--%sendpoint uses http outside this machine; the document and the key would travel "+
			"unencrypted. Use https or pass --insecure-endpoint", role)
	}
	return address, nil
}

func loopback(host string) bool {
	if host == "localhost" {
		return true
	}
	address := net.ParseIP(host)
	return address != nil && address.IsLoopback()
}

func (c *endpointClient) identity() modelIdentity {
	return modelIdentity{Transport: "endpoint", Model: c.model, Host: c.host}
}

type chatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type chatRequest struct {
	Model     string        `json:"model"`
	Messages  []chatMessage `json:"messages"`
	Stream    bool          `json:"stream"`
	MaxTokens int           `json:"max_tokens,omitempty"`
}

type chatResponse struct {
	Choices []struct {
		FinishReason string `json:"finish_reason"`
		Message      struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
}

func (c *endpointClient) ask(ctx context.Context, prompt string) (string, error) {
	body, err := json.Marshal(chatRequest{Model: c.model, MaxTokens: c.maxTokens,
		Messages: []chatMessage{{Role: "system", Content: answerOnly}, {Role: "user", Content: prompt}}})
	if err != nil {
		return "", permanentError{err}
	}
	request, err := http.NewRequestWithContext(ctx, http.MethodPost, c.address, bytes.NewReader(body))
	if err != nil {
		return "", permanentError{err}
	}
	request.Header.Set("Content-Type", "application/json")
	if c.key != "" {
		request.Header.Set("Authorization", "Bearer "+c.key)
	}
	response, err := c.client.Do(request)
	if err != nil {
		// The transport's error repeats the whole address. A query string can
		// hold a credential, so only the host and the cause are reported.
		if transport, ok := errors.AsType[*url.Error](err); ok {
			err = transport.Err
		}
		return "", fmt.Errorf("request to %s failed: %w", c.host, err)
	}
	data, readErr := io.ReadAll(io.LimitReader(response.Body, maxAnswerBytes+1))
	if closeErr := response.Body.Close(); readErr == nil {
		readErr = closeErr
	}
	if readErr != nil {
		return "", fmt.Errorf("reading the answer of %s failed: %w", c.host, readErr)
	}
	if len(data) > maxAnswerBytes {
		return "", permanentError{fmt.Errorf("the answer of %s exceeds %d bytes", c.host, maxAnswerBytes)}
	}
	if response.StatusCode != http.StatusOK {
		return "", statusError(c.host, response.StatusCode, c.withoutKey(string(data)))
	}
	return chatAnswer(c.host, data)
}

// withoutKey removes the key from text that goes into an error message, in
// case a server repeats it.
func (c *endpointClient) withoutKey(text string) string {
	if c.key == "" {
		return text
	}
	return strings.ReplaceAll(text, c.key, "[key]")
}

// statusError keeps the server's own explanation, shortened. A client error
// other than a timeout or a rate limit will not change on a second call.
func statusError(host string, status int, body string) error {
	err := fmt.Errorf("%s answered %d: %s", host, status, excerpt(body, 300))
	retry := status == http.StatusRequestTimeout || status == http.StatusTooManyRequests || status >= http.StatusInternalServerError
	if retry {
		return err
	}
	return permanentError{err}
}

func chatAnswer(host string, data []byte) (string, error) {
	var parsed chatResponse
	if err := json.Unmarshal(data, &parsed); err != nil {
		return "", fmt.Errorf("%s did not answer in the chat completions format: %w", host, err)
	}
	if len(parsed.Choices) == 0 {
		return "", fmt.Errorf("%s answered without choices", host)
	}
	choice := parsed.Choices[0]
	if strings.TrimSpace(choice.Message.Content) == "" {
		return "", fmt.Errorf("%s returned no text (finish reason %q); a model that reasons before it answers "+
			"may need a larger --max-tokens", host, choice.FinishReason)
	}
	return choice.Message.Content, nil
}

type commandClient struct {
	program string
	name    string
	args    []string
}

func newCommandClient(environment Environment, options modelOptions) (modelClient, error) {
	program := options.command
	if strings.ContainsRune(program, filepath.Separator) || strings.Contains(program, "/") {
		if !filepath.IsAbs(program) {
			program = filepath.Join(environment.Dir, program)
		}
	}
	resolved, err := exec.LookPath(program)
	if err != nil {
		return nil, fmt.Errorf("review command %q cannot be started: %w", options.command, err)
	}
	return &commandClient{program: resolved, name: filepath.Base(options.command), args: options.commandArgs}, nil
}

func (c *commandClient) identity() modelIdentity {
	return modelIdentity{Transport: "command", Command: c.name}
}

// ask runs the program in an empty directory with the prompt on standard
// input. The program sees no project file, so the answer depends on the
// prompt alone.
func (c *commandClient) ask(ctx context.Context, prompt string) (answer string, err error) {
	dir, err := os.MkdirTemp("", "unswell-review-*")
	if err != nil {
		return "", permanentError{err}
	}
	defer func() {
		if removeErr := os.RemoveAll(dir); err == nil && removeErr != nil {
			err = removeErr
		}
	}()
	// #nosec G204 -- The user names the review program and its arguments explicitly; it runs by argv, without a shell.
	command := exec.CommandContext(ctx, c.program, c.args...)
	command.Dir = dir
	command.Stdin = strings.NewReader(prompt)
	stdout, stderr := &cappedBuffer{limit: maxAnswerBytes}, &cappedBuffer{limit: 64 << 10}
	command.Stdout, command.Stderr = stdout, stderr
	command.WaitDelay = 5 * time.Second
	if runErr := command.Run(); runErr != nil {
		if ctxErr := ctx.Err(); ctxErr != nil {
			return "", fmt.Errorf("%s: %w", c.name, ctxErr)
		}
		if _, missing := errors.AsType[*exec.Error](runErr); missing {
			return "", permanentError{fmt.Errorf("%s: %w", c.name, runErr)}
		}
		return "", fmt.Errorf("%s: %w: %s", c.name, runErr, excerpt(stderr.String(), 300))
	}
	if stdout.overflow {
		return "", permanentError{fmt.Errorf("the answer of %s exceeds %d bytes", c.name, maxAnswerBytes)}
	}
	if strings.TrimSpace(stdout.String()) == "" {
		return "", fmt.Errorf("%s printed no answer: %s", c.name, excerpt(stderr.String(), 300))
	}
	return stdout.String(), nil
}

// cappedBuffer keeps the first limit bytes and remembers that more arrived.
type cappedBuffer struct {
	bytes.Buffer
	limit    int
	overflow bool
}

func (b *cappedBuffer) Write(data []byte) (int, error) {
	room := b.limit - b.Len()
	if len(data) > room {
		b.overflow = true
		b.Buffer.Write(data[:max(room, 0)])
		return len(data), nil
	}
	return b.Buffer.Write(data)
}

// excerpt makes text from a model or a server safe to print on one line: it
// replaces control characters, joins white space, and cuts at the limit.
func excerpt(text string, limit int) string {
	text = strings.Join(strings.Fields(strings.Map(printable, text)), " ")
	if len(text) <= limit {
		return text
	}
	cut := limit
	for cut > 0 && !utf8.RuneStart(text[cut]) {
		cut--
	}
	return text[:cut] + "…"
}

// printable turns a control character into a space, so that text a model
// wrote cannot move the cursor or change colors in a terminal.
func printable(char rune) rune {
	if unicode.IsControl(char) {
		return ' '
	}
	return char
}
