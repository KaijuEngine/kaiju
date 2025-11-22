/******************************************************************************/
/* ollama.go                                                                  */
/******************************************************************************/
/*                            This file is part of                            */
/*                                KAIJU ENGINE                                */
/*                          https://kaijuengine.com/                          */
/******************************************************************************/
/* MIT License                                                                */
/*                                                                            */
/* Copyright (c) 2023-present Kaiju Engine authors (AUTHORS.md).              */
/* Copyright (c) 2015-present Brent Farris.                                   */
/*                                                                            */
/* May all those that this source may reach be blessed by the LORD and find   */
/* peace and joy in life.                                                     */
/* Everyone who drinks of this water will be thirsty again; but whoever       */
/* drinks of the water that I will give him shall never thirst; John 4:13-14  */
/*                                                                            */
/* Permission is hereby granted, free of charge, to any person obtaining a    */
/* copy of this software and associated documentation files (the "Software"), */
/* to deal in the Software without restriction, including without limitation  */
/* the rights to use, copy, modify, merge, publish, distribute, sublicense,   */
/* and/or sell copies of the Software, and to permit persons to whom the      */
/* Software is furnished to do so, subject to the following conditions:       */
/*                                                                            */
/* The above copyright, blessing, biblical verse, notice and                  */
/* this permission notice shall be included in all copies or                  */
/* substantial portions of the Software.                                      */
/*                                                                            */
/* THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS    */
/* OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF                 */
/* MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT.     */
/* IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY       */
/* CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT  */
/* OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE      */
/* OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.                              */
/******************************************************************************/

package ollama

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os/exec"
	"time"
)

type Base64EncodedImage = string

type APIRequest struct {
	Model      string               `json:"model"`
	Prompt     string               `json:"prompt"`
	Messages   []Message            `json:"messages"`
	Stream     bool                 `json:"stream"`
	Suffix     string               `json:"suffix,omitempty"`
	Images     []Base64EncodedImage `json:"images,omitempty"`
	Format     string               `json:"format,omitempty"`
	Template   string               `json:"template,omitempty"`
	KeepAlive  int                  `json:"keep_alive,omitempty"`
	System     string               `json:"system,omitempty"`
	Options    APIRequestOptions    `json:"options,omitempty"`
	Think      bool                 `json:"think,omitempty"`
	Raw        bool                 `json:"raw,omitempty"`
	Tools      []Tool               `json:"tools,omitempty"`
	RetryCount int
}

type APIRequestOptions struct {
	Mirostat      int     `json:"mirostat,omitempty"`
	MirostatEta   float64 `json:"mirostat_eta,omitempty"`
	MirostatTau   float64 `json:"mirostat_tau,omitempty"`
	NumCtx        int64   `json:"num_ctx,omitempty"`
	RepeatLastN   int     `json:"repeat_last_n,omitempty"`
	RepeatPenalty float64 `json:"repeat_penalty,omitempty"`
	Temperature   float64 `json:"temperature,omitempty"`
	Seed          int     `json:"seed,omitempty"`
	Stop          string  `json:"stop,omitempty"`
	NumPredict    int     `json:"num_predict,omitempty"`
	TopK          int     `json:"top_k,omitempty"`
	TopP          float64 `json:"top_p,omitempty"`
	MinP          float64 `json:"min_p,omitempty"`
}

type APIResponse struct {
	Model              string    `json:"model"`
	Created            time.Time `json:"created_at"`
	Thinking           string    `json:"thinking"`
	Response           string    `json:"response"`
	Message            Message   `json:"message"`
	Done               bool      `json:"done"`
	TotalDuration      int64     `json:"total_duration"`
	LoadDuration       int64     `json:"load_duration"`
	PromptEvalCount    int64     `json:"prompt_eval_count"`
	PromptEvalDuration int64     `json:"prompt_eval_duration"`
	EvalCount          int64     `json:"eval_count"`
	EvalDuration       int64     `json:"eval_duration"`
	Error              string    `json:"error"`
}

type Message struct {
	Role               string     `json:"role"`
	Content            string     `json:"content"`
	ToolCalls          []ToolCall `json:"tool_calls"`
	DoneReason         string     `json:"done_reason"`
	Done               bool       `json:"done"`
	TotalDuration      uint64     `json:"total_duration"`
	LoadDuration       uint64     `json:"load_duration"`
	PromptEvalCount    uint64     `json:"prompt_eval_count"`
	PromptEvalDuration uint64     `json:"prompt_eval_duration"`
	EvalCount          uint64     `json:"eval_count"`
	EvalDuration       uint64     `json:"eval_duration"`
}

type Tool struct {
	Type     string   `json:"type"`
	Function Function `json:"function"`
}

type Function struct {
	Name        string             `json:"name"`
	Description string             `json:"description"`
	Parameters  FunctionParameters `json:"parameters"`
}

type FunctionParameters struct {
	Type       string                               `json:"type"`
	Properties map[string]FunctionParameterProperty `json:"properties"`
	Required   []string                             `json:"required,omitempty"`
}

type FunctionParameterProperty struct {
	Type        string   `json:"type"`
	Description string   `json:"description"`
	Enum        []string `json:"enum,omitempty"`
}

type ToolCall struct {
	Function ToolCallFunction `json:"function"`
}

type ToolCallFunction struct {
	Name      string         `json:"name"`
	Arguments map[string]any `json:"arguments"`
}

func Chat(hostAddr string, req APIRequest) (APIResponse, error) {
	var res APIResponse
	var err error
	retries := max(1, req.RetryCount)
	for _, v := range tools {
		req.Tools = append(req.Tools, v.tool)
	}
	for retries > 0 {
		res, err = callInternal(hostAddr, req)
		if err == nil {
			for len(res.Message.ToolCalls) > 0 {
				toolRetries := max(1, req.RetryCount)
				req.Messages = append(req.Messages, res.Message)
				for i := range res.Message.ToolCalls {
					str, tmpErr := callToolFunc(res.Message.ToolCalls[i])
					if tmpErr != nil {
						str = tmpErr.Error()
					}
					req.Messages = append(req.Messages, Message{
						Role:    "tool",
						Content: str,
					})
				}
				for toolRetries >= 0 {
					res, err = callInternal(hostAddr, req)
					toolRetries--
					if err == nil {
						toolRetries = -1
					}
				}
				if err != nil {
					err = errors.New("internal tool call errror within ollama")
				}
			}
			return res, err
		}
		retries--
	}
	return res, err
}

func Stream(hostAddr string, request APIRequest, reader func(res APIResponse) error) error {
	request.Stream = true
	requestData, err := json.Marshal(request)
	if err != nil {
		return err
	}
	endpoint := "generate"
	if len(request.Messages) > 0 {
		endpoint = "chat"
	}
	req, err := http.NewRequest("POST", hostAddr+"/api/"+endpoint, bytes.NewBuffer(requestData))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	client := &http.Client{}
	client.Timeout = time.Second * time.Duration(request.KeepAlive)
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	scanner := bufio.NewScanner(resp.Body)
	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			continue
		}
		var result APIResponse
		if err := json.Unmarshal([]byte(line), &result); err != nil {
			log.Println("Failed to parse Ollama response:", err)
			continue
		}
		if err = reader(result); err != nil || result.Done {
			break
		}
	}
	return err
}

func Stdin(hostAddr string, req APIRequest, reader func(think, msg string) error) error {
	cmd := exec.Command("ollama", "run", req.Model, req.Prompt)
	cmd.Stdin = bytes.NewBufferString(req.Prompt)
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("error getting stdout pipe: %v", err)
	}
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("error starting command: %v", err)
	}
	scanner := bufio.NewScanner(stdout)
	thought := false
	for scanner.Scan() {
		line := scanner.Text()
		if !thought {
			switch line {
			case "Thinking...":
			case "...done thinking.":
				thought = true
			default:
				reader(line, "")
			}
		} else {
			reader("", line)
		}
	}
	if err := scanner.Err(); err != nil {
		return err
	}
	if err := cmd.Wait(); err != nil {
		return err
	}
	return nil
}

func UnloadModel(hostAddr, modelName string) error {
	request := APIRequest{
		Model:     modelName,
		KeepAlive: 0,
	}
	_, err := callInternal(hostAddr, request)
	return err
}

func callInternal(hostAddr string, request APIRequest) (APIResponse, error) {
	requestData, err := json.Marshal(request)
	if err != nil {
		return APIResponse{}, err
	}
	endpoint := "generate"
	if len(request.Messages) > 0 {
		endpoint = "chat"
	}
	req, err := http.NewRequest("POST", hostAddr+"/api/"+endpoint, bytes.NewBuffer(requestData))
	if err != nil {
		return APIResponse{}, err
	}
	req.Header.Set("Content-Type", "application/json")
	client := &http.Client{}
	client.Timeout = time.Second * time.Duration(request.KeepAlive)
	resp, err := client.Do(req)
	if err != nil {
		return APIResponse{}, err
	}
	defer resp.Body.Close()
	var result APIResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return APIResponse{}, err
	}
	if result.Error != "" {
		return result, errors.New(result.Error)
	}
	return result, nil
}
