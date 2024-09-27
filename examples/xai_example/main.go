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
	xenv.Load()
	client := xai.NewOpenAIClient(xai.WithDebug(true))

	// "MJ::JOB::upsample::1::32839cdf-9db3-4836-a50c-43df2fed88d5"
	mj_response, err := client.ActMidjourney(context.Background(), xai.BuildMidjourneyActionContent(xai.MJJOBUpsample, "2", "32839cdf-9db3-4836-a50c-43df2fed88d5"), "1727458186093768")
	if err != nil {
		xlog.Error("Error generating image", "error", err)
		return
	}
	xlog.Info("Image generation response:")
	xlog.Info(x.MustToJSON(mj_response))
	return

	// mj_response, err := client.GenerateImageWithMidjourney(context.Background(), "cute cat girl --niji 5")
	// if err != nil {
	// 	xlog.Error("Error generating image", "error", err)
	// 	return
	// }
	// xlog.Info("Image generation response:")
	// xlog.Info(x.MustToJSON(mj_response))

	for {
		mj_status, err := client.GetMidjourneyStatus(context.Background(), mj_response.Result)
		if err != nil {
			xlog.Error("Error getting midjourney status", "error", err)
			return
		}
		xlog.Info("Midjourney status response:")
		xlog.Info(x.MustToJSON(mj_status))

		if mj_status.Status == "SUCCESS" || mj_status.FailReason != "" {
			break
		}

		time.Sleep(5 * time.Second)
	}

	// return

	// Text generation (non-streaming)
	response, err := client.QuickGenerateText(context.Background(), []string{"Hello, world!"}, xai.WithTextModel(xai.ModelGPT4o))
	if err != nil {
		xlog.Error("Error generating text", "error", err)
		return
	}
	xlog.Info("Text generation response:")
	xlog.Info(response)

	// Text generation (streaming)
	xlog.Info("Streaming text generation:")
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
				xlog.Info(text)
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
	xlog.Info("Image generation:")
	imageURLs, err := client.GenerateImage(context.Background(), "A beautiful sunset over the ocean", xai.WithImageModel(xai.DefaultImageModel))
	if err != nil {
		xlog.Error("Error generating image", "error", err)
		return
	}
	for i, url := range imageURLs {
		xlog.Infof("Image %d URL: %s", i+1, url)
	}

	// Embedding generation
	xlog.Info("Embedding generation:")
	embeddings, err := client.CreateEmbeddings(context.Background(), []string{"Hello, world!"}, xai.DefaultEmbeddingModel)
	if err != nil {
		xlog.Error("Error creating embeddings", "error", err)
		return
	}
	xlog.Infof("Embedding for 'Hello, world!': %v", embeddings[0][:5]) // Print first 5 values of the embedding
}
