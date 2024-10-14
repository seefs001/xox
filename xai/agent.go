package xai

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"log/slog"

	"github.com/seefs001/xox/x"
	"github.com/seefs001/xox/xlog"
)

const (
	DefaultAgentModel      = "gpt-4o"
	DefaultMaxIterations   = 10
	DefaultTemperature     = 0.7
	EventChannelBufferSize = 100

	// Event types
	EventTypeStart             = "start"
	EventTypeIteration         = "iteration"
	EventTypeAssistantResponse = "assistant_response"
	EventTypeFinalAnswer       = "final_answer"
	EventTypeToolCall          = "tool_call"
	EventTypeToolResult        = "tool_result"
	EventTypeError             = "error"

	// Finish reasons
	FinishReasonStop      = "stop"
	FinishReasonToolCalls = "tool_calls"

	// Message roles
	RoleSystem = "system"
	RoleUser   = "user"
	RoleTool   = "tool"

	ExpertSystemPrompt = `You are an AI expert system with deep knowledge in various fields. Your goal is to provide detailed, authoritative answers to complex queries. Use your extensive knowledge base and the available tools to offer comprehensive explanations and solutions.

Available tools:
{{TOOL_DESCRIPTIONS}}

Structure your response as follows:
1. Brief Overview
2. Detailed Explanation
3. Key Points
4. Recommendations (if applicable)
5. References or Further Reading (if relevant)`

	TeacherSystemPrompt = `You are an AI teaching assistant designed to help students learn and understand complex topics. Your goal is to explain concepts clearly, provide examples, and guide students through problem-solving processes.

Available tools:
{{TOOL_DESCRIPTIONS}}

Structure your response as follows:
1. Concept Explanation
2. Example(s)
3. Step-by-Step Problem Solving (if applicable)
4. Practice Question
5. Summary`

	CreativeSystemPrompt = `You are an AI creative assistant, designed to help with various creative tasks such as brainstorming, storytelling, and content creation. Your goal is to provide imaginative and original ideas while considering the user's input and preferences.

Available tools:
{{TOOL_DESCRIPTIONS}}

Structure your response as follows:
1. Initial Ideas (2-3 concepts)
2. Detailed Exploration of Best Idea
3. Potential Variations or Alternatives
4. Next Steps or Implementation Suggestions`

	DebateSystemPrompt = `You are an AI debate assistant, capable of analyzing complex issues from multiple perspectives. Your goal is to present balanced arguments, consider counterpoints, and help users understand different sides of an issue.

Available tools:
{{TOOL_DESCRIPTIONS}}

Structure your response as follows:
1. Issue Overview
2. Argument For
3. Argument Against
4. Key Points of Contention
5. Balanced Conclusion`

	TroubleshootingSystemPrompt = `You are an AI troubleshooting assistant, designed to help users diagnose and resolve technical issues. Your goal is to guide users through a systematic problem-solving process, using available tools and information to identify and fix issues.

Available tools:
{{TOOL_DESCRIPTIONS}}

Structure your response as follows:
1. Problem Summary
2. Initial Diagnostic Questions
3. Potential Causes
4. Step-by-Step Troubleshooting Process
5. Resolution or Next Steps`
)

const DefaultSystemPrompt = `You are an AI assistant with access to various tools. Follow these steps for each user query:

1. Thought: Analyze the user's request and think about how to approach it.
2. Action: Decide if you need to use a tool. If so, specify which tool and its input.
3. Observation: If you used a tool, review its output.
4. Repeat steps 1-3 as necessary.
5. Final Answer: When you have enough information, provide the final answer to the user's query.

Always structure your response as follows:
Thought: [Your thought process]
Action: [Tool name] ([tool input]) or "None" if no tool is needed
Observation: [Tool output or "N/A" if no tool was used]
... (repeat as needed)
Final Answer: [Your final response to the user]

Available tools:
{{TOOL_DESCRIPTIONS}}

Remember to use tools when necessary and provide a final answer only when you have sufficient information.`

const SimpleSystemPrompt = `You are an AI assistant with access to various tools. Your goal is to provide accurate and helpful answers to user queries. Use the available tools when necessary to gather information or perform actions. Always aim to give a clear and concise final answer to the user's question.

Available tools:
{{TOOL_DESCRIPTIONS}}

Provide your final answer directly, without including your thought process or intermediate steps.`

// Event represents various events that occur during agent execution
type Event struct {
	Type      string      `json:"type"`
	Timestamp time.Time   `json:"timestamp"`
	Data      interface{} `json:"data"`
}

// Add this near the top of the file with other type definitions
type AgentCallback func(eventType, message string)

type Agent struct {
	client         *OpenAIClient
	model          string
	systemPrompt   string
	maxIterations  int
	temperature    float32
	tools          []Tool
	messageHistory []ChatCompletionMessage
	toolExecutor   ToolExecutor
	debug          bool
	eventChan      chan Event
	name           string // Add a name field to identify agents
	callback       AgentCallback
}

type AgentOption func(*Agent)

type ToolExecutor func(toolName string, args map[string]interface{}) (string, error)

// WithAgentDebug enables or disables debug mode for the Agent
func WithAgentDebug(debug bool) AgentOption {
	return func(a *Agent) {
		a.debug = debug
	}
}

// WithAgentHttpClientDebug enables or disables debug mode for the Agent's HTTP client
func WithAgentHttpClientDebug(debug bool) AgentOption {
	return func(a *Agent) {
		a.client.httpClient.SetDebug(debug)
	}
}

// Add this new option function
func WithAgentCallback(callback AgentCallback) AgentOption {
	return func(a *Agent) {
		a.callback = callback
	}
}

func NewAgent(client *OpenAIClient, toolExecutor ToolExecutor, options ...AgentOption) *Agent {
	agent := &Agent{
		client:        client,
		model:         DefaultAgentModel,
		systemPrompt:  DefaultSystemPrompt,
		maxIterations: DefaultMaxIterations,
		temperature:   DefaultTemperature,
		toolExecutor:  toolExecutor,
		debug:         false,
		eventChan:     make(chan Event, EventChannelBufferSize),
		name:          "default",
	}

	for _, option := range options {
		option(agent)
	}

	agent.messageHistory = []ChatCompletionMessage{
		{Role: RoleSystem, Content: agent.systemPrompt},
	}

	return agent
}

func WithAgentModel(model string) AgentOption {
	return func(a *Agent) {
		a.model = model
	}
}

func WithAgentSystemPrompt(prompt string) AgentOption {
	return func(a *Agent) {
		a.systemPrompt = prompt
		if len(a.messageHistory) > 0 && a.messageHistory[0].Role == RoleSystem {
			a.messageHistory[0].Content = prompt
		} else {
			a.messageHistory = append([]ChatCompletionMessage{{Role: RoleSystem, Content: prompt}}, a.messageHistory...)
		}
	}
}

func WithAgentMaxIterations(max int) AgentOption {
	return func(a *Agent) {
		a.maxIterations = max
	}
}

func WithAgentTemperature(temp float32) AgentOption {
	return func(a *Agent) {
		a.temperature = temp
	}
}

func WithAgentTools(tools []Tool) AgentOption {
	return func(a *Agent) {
		a.tools = tools
	}
}

func WithAgentName(name string) AgentOption {
	return func(a *Agent) {
		a.name = name
	}
}

func (a *Agent) Run(ctx context.Context, userInput string) (string, error) {
	a.messageHistory = append(a.messageHistory, ChatCompletionMessage{Role: RoleUser, Content: userInput})

	for i := 0; i < a.maxIterations; i++ {
		response, err := a.getCompletion(ctx)
		if err != nil {
			return "", err
		}

		content := response.Choices[0].Message.Content
		a.messageHistory = append(a.messageHistory, response.Choices[0].Message)

		if len(a.tools) == 0 || response.Choices[0].FinishReason != FinishReasonToolCalls {
			return content, nil
		}

		for _, toolCall := range response.Choices[0].Message.ToolCalls {
			toolResult, err := a.executeTool(toolCall)
			if err != nil {
				return "", err
			}

			a.messageHistory = append(a.messageHistory, ChatCompletionMessage{
				Role:       RoleTool,
				Content:    toolResult,
				ToolCallID: toolCall.ID,
			})
		}
	}

	return "", fmt.Errorf("max iterations reached without resolution")
}

// Helper function to truncate long strings
func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}

func (a *Agent) getCompletion(ctx context.Context) (*CreateChatCompletionResponse, error) {
	req := CreateChatCompletionRequest{
		Model:       a.model,
		Messages:    a.messageHistory,
		Temperature: a.temperature,
	}

	if len(a.tools) > 0 {
		req.Tools = a.tools
		req.ToolChoice = "auto"
	}

	if a.debug {
		xlog.YellowLog(slog.LevelDebug, "Sending request to OpenAI", x.MustToJSON(req))
	}

	resp, err := a.client.CreateChatCompletion(ctx, req)
	if err != nil {
		return nil, err
	}

	if a.debug {
		xlog.YellowLog(slog.LevelDebug, "Received response from OpenAI", x.MustToJSON(resp))
	}

	return resp, nil
}

func (a *Agent) executeTool(toolCall ToolCall) (string, error) {
	var args map[string]interface{}
	err := json.Unmarshal([]byte(toolCall.Function.Arguments), &args)
	if err != nil {
		return "", fmt.Errorf("failed to parse tool arguments: %w", err)
	}

	result, err := a.toolExecutor(toolCall.Function.Name, args)
	if err != nil {
		errorMsg := fmt.Sprintf("Tool %s execution failed. Error: %s", toolCall.Function.Name, err.Error())
		a.messageHistory = append(a.messageHistory, ChatCompletionMessage{
			Role:       RoleTool,
			Content:    errorMsg,
			ToolCallID: toolCall.ID,
		})
		return errorMsg, nil // Return error message without stopping execution
	}

	successMsg := fmt.Sprintf("Tool %s executed successfully. Result: %s", toolCall.Function.Name, result)
	return successMsg, nil
}

func extractFinalAnswer(content string) string {
	lines := strings.Split(content, "\n")
	for _, line := range lines {
		if strings.HasPrefix(line, "Final Answer:") {
			return strings.TrimPrefix(line, "Final Answer:")
		}
	}
	return content
}

func (a *Agent) RunWithEvents(ctx context.Context, userInput string) (<-chan Event, string, error) {
	a.messageHistory = append(a.messageHistory, ChatCompletionMessage{Role: RoleUser, Content: userInput})
	a.eventChan = make(chan Event, EventChannelBufferSize) // Reinitialize channel for each run

	a.sendEvent(EventTypeStart, map[string]interface{}{"input": userInput})

	resultChan := make(chan string, 1)
	errChan := make(chan error, 1)

	go func() {
		defer close(a.eventChan)
		defer close(resultChan)
		defer close(errChan)

		for i := 0; i < a.maxIterations; i++ {
			a.sendEvent(EventTypeIteration, map[string]interface{}{"current": i + 1, "max": a.maxIterations})

			response, err := a.getCompletion(ctx)
			if err != nil {
				a.sendEvent(EventTypeError, map[string]interface{}{"message": err.Error()})
				errChan <- err
				return
			}

			content := response.Choices[0].Message.Content
			a.messageHistory = append(a.messageHistory, response.Choices[0].Message)

			a.sendEvent(EventTypeAssistantResponse, map[string]interface{}{"content": content})

			if response.Choices[0].FinishReason == FinishReasonStop {
				finalAnswer := extractFinalAnswer(content)
				a.sendEvent(EventTypeFinalAnswer, map[string]interface{}{"answer": finalAnswer})
				resultChan <- finalAnswer
				return
			}

			if response.Choices[0].FinishReason == FinishReasonToolCalls {
				for _, toolCall := range response.Choices[0].Message.ToolCalls {
					a.sendEvent(EventTypeToolCall, map[string]interface{}{
						"tool":      toolCall.Function.Name,
						"arguments": toolCall.Function.Arguments,
					})

					toolResult, err := a.executeTool(toolCall)
					if err != nil {
						a.sendEvent(EventTypeError, map[string]interface{}{"message": err.Error()})
						errChan <- err
						return
					}

					a.sendEvent(EventTypeToolResult, map[string]interface{}{"result": toolResult})

					a.messageHistory = append(a.messageHistory, ChatCompletionMessage{
						Role:       RoleTool,
						Content:    toolResult,
						ToolCallID: toolCall.ID,
					})
				}
			}
		}

		a.sendEvent(EventTypeError, map[string]interface{}{"message": "max iterations reached without resolution"})
		errChan <- fmt.Errorf("max iterations reached without resolution")
	}()

	var result string
	var err error

	select {
	case result = <-resultChan:
	case err = <-errChan:
	}

	return a.eventChan, result, err
}

// Modify the sendEvent method to use the callback if it's set
func (a *Agent) sendEvent(eventType string, data interface{}) {
	event := Event{
		Type:      eventType,
		Timestamp: time.Now(),
		Data:      data,
	}
	a.eventChan <- event

	// If a callback is set, call it with the event information
	if a.callback != nil {
		message, _ := json.Marshal(data)
		a.callback(eventType, string(message))
	}
}

// InteractWithAgent allows interaction with the Agent using channels
func (a *Agent) InteractWithAgent(ctx context.Context, inputChan <-chan string, outputChan chan<- Event) {
	defer close(outputChan)

	for {
		select {
		case <-ctx.Done():
			return
		case input, ok := <-inputChan:
			if !ok {
				return
			}

			eventChan, result, err := a.RunWithEvents(ctx, input)
			if err != nil {
				outputChan <- Event{
					Type:      EventTypeError,
					Timestamp: time.Now(),
					Data:      map[string]interface{}{"message": err.Error()},
				}
				continue
			}

			for event := range eventChan {
				outputChan <- event
			}

			outputChan <- Event{
				Type:      EventTypeFinalAnswer,
				Timestamp: time.Now(),
				Data:      map[string]interface{}{"answer": result},
			}
		}
	}
}

// CollaborateWithAgents allows an agent to collaborate with other agents
func (a *Agent) CollaborateWithAgents(ctx context.Context, query string, agents map[string]*Agent) (string, error) {
	a.sendEvent(EventTypeStart, map[string]interface{}{"input": query})

	// Add the initial query to the message history
	a.messageHistory = append(a.messageHistory, ChatCompletionMessage{Role: RoleUser, Content: query})

	for i := 0; i < a.maxIterations; i++ {
		a.sendEvent(EventTypeIteration, map[string]interface{}{"current": i + 1, "max": a.maxIterations})

		response, err := a.getCompletion(ctx)
		if err != nil {
			a.sendEvent(EventTypeError, map[string]interface{}{"message": err.Error()})
			return "", err
		}

		content := response.Choices[0].Message.Content
		a.messageHistory = append(a.messageHistory, response.Choices[0].Message)

		a.sendEvent(EventTypeAssistantResponse, map[string]interface{}{"content": content})

		if response.Choices[0].FinishReason == FinishReasonStop {
			finalAnswer := extractFinalAnswer(content)
			a.sendEvent(EventTypeFinalAnswer, map[string]interface{}{"answer": finalAnswer})
			return finalAnswer, nil
		}

		if response.Choices[0].FinishReason == FinishReasonToolCalls {
			for _, toolCall := range response.Choices[0].Message.ToolCalls {
				if toolCall.Function.Name == "ask_agent" {
					var args struct {
						AgentName string `json:"agent_name"`
						Question  string `json:"question"`
					}
					err := json.Unmarshal([]byte(toolCall.Function.Arguments), &args)
					if err != nil {
						return "", fmt.Errorf("failed to parse ask_agent arguments: %w", err)
					}

					targetAgent, exists := agents[args.AgentName]
					if !exists {
						return "", fmt.Errorf("agent not found: %s", args.AgentName)
					}

					result, err := targetAgent.Run(ctx, args.Question)
					if err != nil {
						return "", fmt.Errorf("error asking agent %s: %w", args.AgentName, err)
					}

					a.messageHistory = append(a.messageHistory, ChatCompletionMessage{
						Role:       RoleTool,
						Content:    fmt.Sprintf("Response from %s: %s", args.AgentName, result),
						ToolCallID: toolCall.ID,
					})

					a.sendEvent(EventTypeToolResult, map[string]interface{}{
						"agent":  args.AgentName,
						"result": result,
					})
				} else {
					toolResult, err := a.executeTool(toolCall)
					if err != nil {
						return "", err
					}

					a.messageHistory = append(a.messageHistory, ChatCompletionMessage{
						Role:       RoleTool,
						Content:    toolResult,
						ToolCallID: toolCall.ID,
					})

					a.sendEvent(EventTypeToolResult, map[string]interface{}{
						"tool":   toolCall.Function.Name,
						"result": toolResult,
					})
				}
			}
		}
	}

	return "", fmt.Errorf("max iterations reached without resolution")
}
