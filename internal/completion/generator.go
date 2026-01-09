package completion

import (
	"bytes"
	"fmt"
	"text/template"
)

// Data contains the data needed to generate shell completion scripts.
type Data struct {
	ProgramName string
	Flags       []FlagInfo
}

// Generate produces a shell completion script for the specified shell type.
// Supported shells: bash, zsh, fish.
// Returns an error if the shell type is unsupported.
func Generate(shell string) (string, error) {
	data := Data{
		ProgramName: "lazyworktree",
		Flags:       GetFlags(),
	}

	switch shell {
	case "bash":
		return generateBash(data)
	case "zsh":
		return generateZsh(data)
	case "fish":
		return generateFish(data)
	default:
		return "", fmt.Errorf("unsupported shell: %s (supported: bash, zsh, fish)", shell)
	}
}

// generateBash generates a Bash completion script.
func generateBash(data Data) (string, error) {
	tmpl, err := template.New("bash").Parse(bashTemplate)
	if err != nil {
		return "", fmt.Errorf("failed to parse bash template: %w", err)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("failed to execute bash template: %w", err)
	}

	return buf.String(), nil
}

// generateZsh generates a Zsh completion script.
func generateZsh(data Data) (string, error) {
	tmpl, err := template.New("zsh").Parse(zshTemplate)
	if err != nil {
		return "", fmt.Errorf("failed to parse zsh template: %w", err)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("failed to execute zsh template: %w", err)
	}

	return buf.String(), nil
}

// generateFish generates a Fish completion script.
func generateFish(data Data) (string, error) {
	tmpl, err := template.New("fish").Parse(fishTemplate)
	if err != nil {
		return "", fmt.Errorf("failed to parse fish template: %w", err)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("failed to execute fish template: %w", err)
	}

	return buf.String(), nil
}
