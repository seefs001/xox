package xlog_handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/seefs001/xox/x"
	"github.com/seefs001/xox/xhttpc"
)

type AxiomHandler struct {
	client     *xhttpc.Client
	dataset    string
	debug      bool
	logOptions xhttpc.LogOptions
	buffer     []map[string]interface{}
	bufferSize int
	mu         sync.Mutex
	sending    bool
	shutdownCh chan struct{}
	wg         sync.WaitGroup
}

func NewAxiomHandler(apiToken, dataset string) *AxiomHandler {
	client := xhttpc.NewClient(
		xhttpc.WithBearerToken(apiToken),
		xhttpc.WithBaseURL("https://api.axiom.co"),
	)
	return &AxiomHandler{
		client:     client,
		dataset:    dataset,
		buffer:     make([]map[string]interface{}, 0),
		bufferSize: 100, // Adjust this value as needed
		shutdownCh: make(chan struct{}),
	}
}

func (h *AxiomHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return true
}

func (h *AxiomHandler) Handle(ctx context.Context, r slog.Record) error {
	data := make(map[string]interface{})
	data["level"] = r.Level.String()
	data["message"] = r.Message
	data["time"] = r.Time.Format(time.RFC3339)

	r.Attrs(func(a slog.Attr) bool {
		data[a.Key] = a.Value.Any()
		return true
	})

	h.mu.Lock()
	h.buffer = append(h.buffer, data)
	shouldSend := len(h.buffer) >= h.bufferSize && !h.sending
	h.mu.Unlock()

	if shouldSend {
		h.wg.Add(1)
		go func() {
			defer h.wg.Done()
			h.sendLogs(ctx)
		}()
	}

	return nil
}

func (h *AxiomHandler) sendLogs(ctx context.Context) {
	h.mu.Lock()
	if h.sending {
		h.mu.Unlock()
		return
	}
	h.sending = true
	logs := h.buffer
	h.buffer = make([]map[string]interface{}, 0)
	h.mu.Unlock()

	select {
	case <-ctx.Done():
		return
	case <-h.shutdownCh:
		return
	default:
		// Proceed with sending logs
	}

	if len(logs) == 0 {
		h.mu.Lock()
		h.sending = false
		h.mu.Unlock()
		return
	}

	fmt.Println("lllll" + x.MustToJSON(logs))
	jsonData, err := json.Marshal(logs)
	if err != nil {
		fmt.Printf("failed to marshal log data: %v\n", err)
		h.mu.Lock()
		h.sending = false
		h.mu.Unlock()
		return
	}

	url := fmt.Sprintf("/v1/datasets/%s/ingest", h.dataset)
	resp, err := h.client.Post(ctx, url, bytes.NewReader(jsonData))
	if err != nil {
		fmt.Printf("failed to send log to Axiom: %v\n", err)
	} else {
		defer resp.Body.Close()
		if resp.StatusCode >= 400 {
			fmt.Printf("Axiom API error: %s\n", resp.Status)
		}
	}

	h.mu.Lock()
	h.sending = false
	h.mu.Unlock()
}

func (h *AxiomHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return h
}

func (h *AxiomHandler) WithGroup(name string) slog.Handler {
	return h
}

func (h *AxiomHandler) SetDebug(debug bool) {
	h.debug = debug
	h.client.SetDebug(debug)
}

func (h *AxiomHandler) SetLogOptions(options xhttpc.LogOptions) {
	h.logOptions = options
	h.client.SetLogOptions(options)
}

func (h *AxiomHandler) Shutdown() {
	close(h.shutdownCh)
	h.wg.Wait()

	// Flush any remaining logs
	h.mu.Lock()
	remainingLogs := h.buffer
	h.buffer = nil
	h.mu.Unlock()

	if len(remainingLogs) > 0 {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		h.sendLogsImmediate(ctx, remainingLogs)
	}
}

func (h *AxiomHandler) sendLogsImmediate(ctx context.Context, logs []map[string]interface{}) {
	// Similar to sendLogs, but sends immediately without buffering
	// jsonData, err := json.Marshal(logs)
	// if err != nil {
	// 	fmt.Printf("failed to marshal log data: %v\n", err)
	// 	return
	// }

	url := fmt.Sprintf("/v1/datasets/%s/ingest", h.dataset)
	resp, err := h.client.PostJSON(ctx, url, logs)
	if err != nil {
		fmt.Printf("failed to send log to Axiom immediately: %v\n", err)
	} else {
		defer resp.Body.Close()
		if resp.StatusCode >= 400 {
			fmt.Printf("Axiom API error: %s\n", resp.Status)
		}
	}
}
