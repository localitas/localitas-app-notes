package notes

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

type PythonRunner struct {
	baseURL string
	token   string
}

func NewPythonRunner(baseURL, token string) *PythonRunner {
	return &PythonRunner{baseURL: baseURL, token: token}
}

type pythonExecRequest struct {
	Code         string                 `json:"code"`
	Variables    map[string]interface{} `json:"variables,omitempty"`
	Requirements []string               `json:"requirements,omitempty"`
}

type pythonExecResponse struct {
	Output string `json:"output"`
	Error  string `json:"error,omitempty"`
}

func (p *PythonRunner) Execute(ctx context.Context, code string, variables map[string]interface{}, requirements []string) (string, error) {
	body, err := json.Marshal(pythonExecRequest{
		Code:         code,
		Variables:    variables,
		Requirements: requirements,
	})
	if err != nil {
		return "", err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", p.baseURL+"/execute", bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	if p.token != "" {
		req.Header.Set("Authorization", "Bearer "+p.token)
	}

	client := &http.Client{Timeout: 120 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("python runner unavailable: %w", err)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusOK {
		var errResp pythonExecResponse
		json.Unmarshal(respBody, &errResp)
		if errResp.Error != "" {
			return "", fmt.Errorf("%s", errResp.Error)
		}
		return "", fmt.Errorf("python execution failed: %s", string(respBody))
	}

	var result pythonExecResponse
	if err := json.Unmarshal(respBody, &result); err != nil {
		return "", err
	}
	return result.Output, nil
}
