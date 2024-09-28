package main

import (
	"context"
	"time"

	"github.com/seefs001/xox/x"
	"github.com/seefs001/xox/xai"
	"github.com/seefs001/xox/xenv"
	"github.com/seefs001/xox/xlog"
)

func main() {
	xlog.Info("Starting AI example")

	xenv.Load()
	xlog.Info("Environment variables loaded")

	client := xai.NewOpenAIClient(xai.WithDebug(false))
	xlog.Info("OpenAI client created with debug mode enabled")

	req, err := client.GenerateImageWithMidjourney(context.Background(), "自分よりお酒が強い、見た目とのギャップが強い子 --ar 16:9 --q 2 --niji 6")
	if err != nil {
		xlog.Error("Error generating image", "error", err)
		return
	}
	// req, err := client.ActMidjourney(context.Background(), "MJ::JOB::upsample::3::f5ab64bd-0682-472f-be83-de7a99735069", "1727477334464428")
	// if err != nil {
	// 	xlog.Error("Error generating image", "error", err)
	// 	return
	// }
	xlog.Info("Image generation response:")
	xlog.Info(x.MustToJSON(req))

	for {
		status, err := client.GetMidjourneyStatus(context.Background(), req.Result)
		if err != nil {
			xlog.Error("Error getting midjourney status", "error", err)
			continue
		}
		if status.Status == "SUCCESS" {
			xlog.Info("Midjourney status:", "status", status)
			xlog.Info(x.MustToJSON(status))
			break
		}
		xlog.Info("Midjourney status in progress:", "status", status)
		time.Sleep(5 * time.Second)
	}
	return
	// Text generation (non-streaming)
	xlog.Info("Generating text using QuickGenerateText")
	response, err := client.QuickGenerateText(context.Background(), []string{"Hello, world!"}, xai.WithTextModel(xai.ModelGPT4o))
	if err != nil {
		xlog.Error("Error generating text", "error", err)
		return
	}
	xlog.Info("Text generation response:", "response", response)

	// Text generation (streaming)
	xlog.Info("Starting streaming text generation")
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
					xlog.Info("Stream finished")
					return
				}
				xlog.Info("Received text from stream", "text", text)
			case err, ok := <-errChan:
				if !ok {
					return
				}
				if err != nil {
					xlog.Error("Error generating text stream", "error", err)
					return
				}
			case <-ctx.Done():
				return
			}
		}
	}()

	<-streamFinished

	// Image generation
	xlog.Info("Generating image")
	imageURLs, err := client.GenerateImage(context.Background(), "A beautiful sunset over the ocean", xai.WithImageModel(xai.DefaultImageModel))
	if err != nil {
		xlog.Error("Error generating image", "error", err)
		return
	}
	for i, url := range imageURLs {
		xlog.Info("Generated image URL", "index", i+1, "url", url)
	}

	// Embedding generation
	xlog.Info("Creating embeddings")
	embeddings, err := client.CreateEmbeddings(context.Background(), []string{"Hello, world!"}, xai.DefaultEmbeddingModel)
	if err != nil {
		xlog.Error("Error creating embeddings", "error", err)
		return
	}
	xlog.Info("Embedding created", "embedding_sample", embeddings[0][:5]) // Print first 5 values of the embedding

	xlog.Info("AI example completed")
}
