package notes

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type ExecutionResult struct {
	BlockID  string `json:"block_id"`
	Success  bool   `json:"success"`
	Output   string `json:"output"`
	Error    string `json:"error,omitempty"`
	Duration int64  `json:"duration_ms"`
}

type ExecutionResults struct {
	ID           string                      `json:"id"`
	ExecutedAt   time.Time                   `json:"executed_at"`
	TotalBlocks  int                         `json:"total_blocks"`
	SuccessCount int                         `json:"success_count"`
	FailCount    int                         `json:"fail_count"`
	Results      map[string]*ExecutionResult `json:"results"`
}

func noteDataDir(noteID string) string {
	homeDir, _ := os.UserHomeDir()
	return filepath.Join(homeDir, ".localitas", "apps", "notes-standalone", "data", noteID)
}

func (s *Store) ExecuteAllBlocks(ctx context.Context, noteID string) (*ExecutionResults, error) {
	note, err := s.Get(ctx, noteID)
	if err != nil {
		return nil, fmt.Errorf("failed to load note: %w", err)
	}

	blocks := ParseCodeBlocks(note.Content)
	requirements := extractRequirements(blocks)

	dataDir := noteDataDir(noteID)
	os.MkdirAll(dataDir, 0755)

	registry, err := LoadVariableRegistry(dataDir)
	if err != nil {
		return nil, fmt.Errorf("failed to load variable registry: %w", err)
	}

	execResults := &ExecutionResults{
		ID:          noteID,
		ExecutedAt:  time.Now().UTC(),
		TotalBlocks: len(blocks),
		Results:     make(map[string]*ExecutionResult),
	}

	for _, block := range blocks {
		startTime := time.Now().UTC()

		select {
		case <-ctx.Done():
			return execResults, ctx.Err()
		default:
		}

		result := &ExecutionResult{BlockID: block.ID}

		var variable *Variable
		var output string
		var execErr error

		switch block.Language {
		case "csv":
			variable, execErr = LoadCSVVariable(dataDir, block.VarName, block.Content, block.ID)
			if execErr == nil && variable != nil {
				output = fmt.Sprintf("Loaded %s: %d rows x %d columns",
					block.VarName, variable.Schema.Rows, len(variable.Schema.Columns))
			}

		case "js":
			variable, output, execErr = ExecuteJSBlock(block, registry)

		case "python":
			if s.PythonRunner == nil {
				execErr = fmt.Errorf("Python runner not configured")
			} else {
				vars := make(map[string]interface{})
				for k, v := range registry.Variables {
					if v.Value != nil {
						vars[k] = v.Value
					}
				}
				output, execErr = s.PythonRunner.Execute(ctx, block.Content, vars, requirements)
			}

		case "requirements.txt":
			continue

		default:
			execErr = fmt.Errorf("unsupported language: %s", block.Language)
		}

		if execErr != nil {
			result.Success = false
			result.Error = execErr.Error()
			execResults.FailCount++
		} else {
			result.Success = true
			result.Output = output
			execResults.SuccessCount++

			if variable != nil {
				registry.Variables[variable.Name] = variable
			}
		}

		result.Duration = time.Since(startTime).Milliseconds()
		execResults.Results[block.ID] = result
	}

	SaveVariableRegistry(dataDir, registry)
	saveExecutionResults(dataDir, execResults)

	return execResults, nil
}

func (s *Store) ExecuteSingleBlock(ctx context.Context, noteID, blockID string) (*ExecutionResult, error) {
	note, err := s.Get(ctx, noteID)
	if err != nil {
		return nil, fmt.Errorf("failed to load note: %w", err)
	}

	blocks := ParseCodeBlocks(note.Content)
	var targetBlock *CodeBlock
	for _, block := range blocks {
		if block.ID == blockID {
			targetBlock = block
			break
		}
	}
	if targetBlock == nil {
		return nil, fmt.Errorf("block not found: %s", blockID)
	}

	dataDir := noteDataDir(noteID)
	os.MkdirAll(dataDir, 0755)

	registry, err := LoadVariableRegistry(dataDir)
	if err != nil {
		return nil, fmt.Errorf("failed to load variable registry: %w", err)
	}

	startTime := time.Now().UTC()
	result := &ExecutionResult{BlockID: blockID}

	var variable *Variable
	var output string
	var execErr error

	switch targetBlock.Language {
	case "csv":
		filePath, parseErr := ParseCSVBlock(targetBlock.Content)
		if parseErr != nil {
			execErr = parseErr
		} else {
			variable, execErr = LoadCSVVariable(dataDir, targetBlock.VarName, filePath, targetBlock.ID)
			if execErr == nil && variable != nil {
				output = fmt.Sprintf("Loaded %s: %d rows x %d columns",
					targetBlock.VarName, variable.Schema.Rows, len(variable.Schema.Columns))
			}
		}

	case "js":
		variable, output, execErr = ExecuteJSBlock(targetBlock, registry)

	case "python":
		if s.PythonRunner == nil {
			execErr = fmt.Errorf("Python runner not configured")
		} else {
			vars := make(map[string]interface{})
			for k, v := range registry.Variables {
				if v.Value != nil {
					vars[k] = v.Value
				}
			}
			reqs := extractRequirements(blocks)
			output, execErr = s.PythonRunner.Execute(ctx, targetBlock.Content, vars, reqs)
		}

	default:
		execErr = fmt.Errorf("unsupported language: %s", targetBlock.Language)
	}

	if execErr != nil {
		result.Success = false
		result.Error = execErr.Error()
	} else {
		result.Success = true
		result.Output = output
		if variable != nil {
			registry.Variables[variable.Name] = variable
			SaveVariableRegistry(dataDir, registry)
		}
	}

	result.Duration = time.Since(startTime).Milliseconds()
	return result, nil
}

func extractRequirements(blocks []*CodeBlock) []string {
	var requirements []string
	for _, block := range blocks {
		if block.Language == "requirements.txt" {
			lines := strings.Split(block.Content, "\n")
			for _, line := range lines {
				line = strings.TrimSpace(line)
				if line != "" && !strings.HasPrefix(line, "#") {
					requirements = append(requirements, line)
				}
			}
		}
	}
	return requirements
}

func saveExecutionResults(dataDir string, results *ExecutionResults) error {
	data, err := json.MarshalIndent(results, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(dataDir, "results.json"), data, 0644)
}
