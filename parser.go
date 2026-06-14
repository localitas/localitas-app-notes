package notes

import (
	"fmt"
	"regexp"
	"strings"
)

type CodeBlock struct {
	ID       string
	Language string
	VarName  string
	Content  string
	Index    int
}

var codeBlockRegex = regexp.MustCompile("```(\\w*)(?::(\\w+))?\\n([\\s\\S]*?)```")

func ParseCodeBlocks(markdown string) []*CodeBlock {
	matches := codeBlockRegex.FindAllStringSubmatch(markdown, -1)
	blocks := make([]*CodeBlock, 0, len(matches))

	for i, match := range matches {
		language := match[1]
		if language == "" {
			language = "js"
		}
		varName := match[2]
		content := strings.TrimSpace(match[3])

		if varName != "" && !isValidIdentifier(varName) {
			continue
		}

		block := &CodeBlock{
			ID:       fmt.Sprintf("block_%d", i),
			Language: language,
			VarName:  varName,
			Content:  content,
			Index:    i,
		}
		blocks = append(blocks, block)
	}
	return blocks
}

func ParseCSVBlock(content string) (filePath string, err error) {
	lines := strings.Split(content, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "file:") {
			filePath = strings.TrimSpace(strings.TrimPrefix(line, "file:"))
		}
	}
	if filePath == "" {
		return "", fmt.Errorf("CSV block missing 'file:' directive")
	}
	return filePath, nil
}

func isValidIdentifier(name string) bool {
	if name == "" {
		return false
	}
	firstChar := rune(name[0])
	if !((firstChar >= 'a' && firstChar <= 'z') ||
		(firstChar >= 'A' && firstChar <= 'Z') ||
		firstChar == '_') {
		return false
	}
	for _, char := range name[1:] {
		if !((char >= 'a' && char <= 'z') ||
			(char >= 'A' && char <= 'Z') ||
			(char >= '0' && char <= '9') ||
			char == '_') {
			return false
		}
	}
	return true
}

func GetBlockByID(blocks []*CodeBlock, id string) *CodeBlock {
	for _, block := range blocks {
		if block.ID == id {
			return block
		}
	}
	return nil
}

func GetBlocksByLanguage(blocks []*CodeBlock, language string) []*CodeBlock {
	filtered := make([]*CodeBlock, 0)
	for _, block := range blocks {
		if block.Language == language {
			filtered = append(filtered, block)
		}
	}
	return filtered
}
