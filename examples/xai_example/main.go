package main

import (
	"context"
	"encoding/json"
	"log/slog"

	"github.com/seefs001/xox/x"
	"github.com/seefs001/xox/xai"
	"github.com/seefs001/xox/xenv"
	"github.com/seefs001/xox/xlog"
)

func main() {
	xlog.GreenLog(slog.LevelInfo, "Starting AI example")

	xenv.Load()
	xlog.GreenLog(slog.LevelInfo, "Environment variables loaded")

	client := xai.NewOpenAIClient(xai.WithDebug(true))
	xlog.GreenLog(slog.LevelInfo, "OpenAI client created with debug mode disabled")

	createChatCompletionResponse, err := client.CreateChatCompletion(context.Background(), xai.CreateChatCompletionRequest{
		Model: xai.ModelGPT4o,
		Messages: []xai.ChatCompletionMessage{
			{
				Role:    "user",
				Content: "How is the weather in Beijing?",
			},
		},
		Stream:     false,
		ToolChoice: "required",
		Tools: []xai.Tool{
			{
				Type: "function",
				Function: xai.Function{
					Name:        "get_weather",
					Parameters:  json.RawMessage(`{"type": "object", "properties": {"city": {"type": "string"}}}`),
					Description: "Get the weather for a given city",
				},
			},
		},
	})
	if err != nil {
		xlog.RedLog(slog.LevelError, "Error creating chat completion", "error", err)
		return
	}
	xlog.CyanLog(slog.LevelInfo, "Chat completion created", "response", x.MustToJSON(createChatCompletionResponse))

	// req, err := client.GenerateImageWithMidjourney(context.Background(), "自分よりお酒が強い、見た目とのギャップが強い子 --ar 16:9 --q 2 --niji 6")
	// if err != nil {
	// 	xlog.Error("Error generating image", "error", err)
	// 	return
	// }
	// // req, err := client.ActMidjourney(context.Background(), "MJ::JOB::upsample::3::f5ab64bd-0682-472f-be83-de7a99735069", "1727477334464428")
	// // if err != nil {
	// // 	xlog.Error("Error generating image", "error", err)
	// // 	return
	// // }
	// xlog.Info("Image generation response:")
	// xlog.Info(x.MustToJSON(req))

	// for {
	// 	status, err := client.GetMidjourneyStatus(context.Background(), req.Result)
	// 	if err != nil {
	// 		xlog.Error("Error getting midjourney status", "error", err)
	// 		continue
	// 	}
	// 	if status.Status == "SUCCESS" {
	// 		xlog.Info("Midjourney status:", "status", status)
	// 		xlog.Info(x.MustToJSON(status))
	// 		break
	// 	}
	// 	xlog.Info("Midjourney status in progress:", "status", status)
	// 	time.Sleep(5 * time.Second)
	// }
	// return
	// Text generation (non-streaming)
	xlog.YellowLog(slog.LevelInfo, "Generating text using QuickGenerateText")
	response, err := client.QuickGenerateText(context.Background(), []string{"Hello, world!"}, xai.WithTextModel(xai.ModelGPT4o))
	if err != nil {
		xlog.RedLog(slog.LevelError, "Error generating text", "error", err)
		return
	}
	xlog.CyanLog(slog.LevelInfo, "Text generation response:", "response", response)

	// Text generation (streaming)
	xlog.YellowLog(slog.LevelInfo, "Starting streaming text generation")
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	textChan, errChan := client.QuickGenerateTextStream(ctx, []string{"Hello, world!"}, xai.WithTextModel(xai.ModelClaude35Sonnet))

	streamFinished := make(chan struct{})
	go func() {
		defer close(streamFinished)
		for {
			select {
			case text, ok := <-textChan:
				if !ok {
					xlog.GreenLog(slog.LevelInfo, "Stream finished")
					return
				}
				xlog.BlueLog(slog.LevelInfo, "Received text from stream", "text", text)
			case err, ok := <-errChan:
				if !ok {
					return
				}
				if err != nil {
					xlog.RedLog(slog.LevelError, "Error generating text stream", "error", err)
					return
				}
			case <-ctx.Done():
				return
			}
		}
	}()

	<-streamFinished

	// Image generation
	xlog.YellowLog(slog.LevelInfo, "Generating image")
	imageURLs, err := client.GenerateImage(context.Background(), "A beautiful sunset over the ocean", xai.WithImageModel(xai.DefaultImageModel))
	if err != nil {
		xlog.RedLog(slog.LevelError, "Error generating image", "error", err)
		return
	}
	for i, url := range imageURLs {
		xlog.PurpleLog(slog.LevelInfo, "Generated image URL", "index", i+1, "url", url)
	}

	// Embedding generation
	xlog.YellowLog(slog.LevelInfo, "Creating embeddings")
	embeddings, err := client.CreateEmbeddings(context.Background(), []string{"Hello, world!"}, xai.DefaultEmbeddingModel)
	if err != nil {
		xlog.RedLog(slog.LevelError, "Error creating embeddings", "error", err)
		return
	}
	xlog.CyanLog(slog.LevelInfo, "Embedding created", "embedding_sample", embeddings[0][:5]) // Print first 5 values of the embedding

	xlog.GreenLog(slog.LevelInfo, "AI example completed")
}
