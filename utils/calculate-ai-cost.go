package utils

import "fmt"

func CalculateAICostWithOutput(inputTokens, outputTokens int) map[string]any {
	// Pricing: Input $0.05, Output $0.20 (per million tokens) 4.1 nano
	inputCost := float64(inputTokens) * 0.05 / 1000000.0
	outputCost := float64(outputTokens) * 0.20 / 1000000.0
	totalCost := inputCost + outputCost

	return map[string]any{
		"inputTokens":  inputTokens,
		"outputTokens": outputTokens,
		"inputCost":    fmt.Sprintf("$%.4f", inputCost),
		"outputCost":   fmt.Sprintf("$%.4f", outputCost),
		"totalCost":    fmt.Sprintf("$%.4f", totalCost),
	}
}
