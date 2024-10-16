package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"

	"github.com/seefs001/xox/xai"
	"github.com/seefs001/xox/xenv"
	"github.com/seefs001/xox/xlog"
)

func main() {
	xlog.SetDefaultLogLevel(slog.LevelDebug)
	xlog.GreenLog(slog.LevelInfo, "Starting AI Agent example")

	xenv.Load()
	xlog.GreenLog(slog.LevelInfo, "Environment variables loaded")

	// Initialize OpenAI client
	client := xai.NewOpenAIClient(xai.WithDebug(true), xai.WithHttpClientDebug(true))

	// Define tools
	tools := []xai.Tool{
		{
			Type: "function",
			Function: xai.Function{
				Name:        "get_weather",
				Description: "Get the current weather for a given location",
				Parameters: json.RawMessage(`{
					"type": "object",
					"properties": {
						"location": {
							"type": "string",
							"description": "The city and state, e.g. San Francisco, CA"
						}
					},
					"required": ["location"]
				}`),
			},
		},
		{
			Type: "function",
			Function: xai.Function{
				Name:        "search_web",
				Description: "Search the web for a given query",
				Parameters: json.RawMessage(`{
					"type": "object",
					"properties": {
						"query": {
							"type": "string",
							"description": "The search query"
						}
					},
					"required": ["query"]
				}`),
			},
		},
	}

	// Define tool executor
	toolExecutor := func(toolName string, args map[string]interface{}) (string, error) {
		switch toolName {
		case "get_weather":
			location, _ := args["location"].(string)
			return fmt.Sprintf("The weather in %s is sunny and 25°C", location), nil
		case "search_web":
			query, _ := args["query"].(string)
			return fmt.Sprintf("Top search result for '%s': AI and machine learning are transforming various industries.", query), nil
		default:
			return "", fmt.Errorf("unknown tool: %s", toolName)
		}
	}

	// Create agent
	agent := xai.NewAgent(client, toolExecutor,
		xai.WithAgentModel("gpt-4o"),
		xai.WithAgentTemperature(0.7),
		xai.WithAgentTools(tools),
		xai.WithAgentDebug(true),
		xai.WithAgentSystemPrompt(xai.DebateSystemPrompt),
	)

	// Run agent
	ctx := context.Background()
	userInput := "What's the weather like in New York? Also, tell me about the latest developments in AI."

	// Using Run
	result, err := agent.Run(ctx, userInput)
	if err != nil {
		xlog.RedLog(slog.LevelError, "Agent run failed", "error", err)
		return
	}

	xlog.GreenLog(slog.LevelInfo, "Agent result", "result", result)

	// Using RunWithEvents
	eventChan, result, err := agent.RunWithEvents(ctx, userInput)
	if err != nil {
		xlog.RedLog(slog.LevelError, "Agent run failed", "error", err)
		return
	}

	for event := range eventChan {
		// Process events as they come in
		xlog.Info("Event", "type", event.Type, "data", event.Data)
	}

	fmt.Printf("Final result: %s\n", result)
}
