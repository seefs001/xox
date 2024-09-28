package xlog_handlers

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/seefs001/xox/x"
	"github.com/seefs001/xox/xerror"
	"github.com/seefs001/xox/xhttpc"
	"github.com/seefs001/xox/xlog"
)

type AxiomHandler struct {
	client        *xhttpc.Client
	dataset       string
	debug         bool
	logOptions    xhttpc.LogOptions
	buffer        []map[string]interface{}
	bufferSize    int
	mu            sync.Mutex
	sending       bool
	shutdownCh    chan struct{}
	wg            sync.WaitGroup
	flushInterval time.Duration
	done          chan struct{} // Add this line
}

func NewAxiomHandler(apiToken, dataset string) *AxiomHandler {
	client := x.Must1(xhttpc.NewClient(
		xhttpc.WithBearerToken(apiToken),
		xhttpc.WithBaseURL("https://api.axiom.co"),
	))
	h := &AxiomHandler{
		client:        client,
		dataset:       dataset,
		buffer:        make([]map[string]interface{}, 0),
		bufferSize:    100, // Adjust this value as needed
		shutdownCh:    make(chan struct{}),
		flushInterval: 10 * time.Second,    // Default flush interval
		done:          make(chan struct{}), // Add this line
	}
	h.startFlusher() // Add this line
	return h
}

func (h *AxiomHandler) Enabled(_ context.Context, _ slog.Level) bool {
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
			if err := h.sendLogs(ctx); err != nil {
				xlog.Error("Error sending logs to Axiom", "error", err)
			}
		}()
	}

	return nil
}

func (h *AxiomHandler) sendLogs(ctx context.Context) error {
	h.mu.Lock()
	if h.sending {
		h.mu.Unlock()
		return nil
	}
	h.sending = true
	logs := h.buffer
	h.buffer = make([]map[string]interface{}, 0)
	h.mu.Unlock()

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-h.shutdownCh:
		return xerror.New("handler is shutting down")
	default:
		// Proceed with sending logs
	}

	if len(logs) == 0 {
		h.mu.Lock()
		h.sending = false
		h.mu.Unlock()
		return nil
	}

	url := fmt.Sprintf("/v1/datasets/%s/ingest", h.dataset)
	resp, err := h.client.Post(ctx, url, logs)
	if err != nil {
		h.mu.Lock()
		h.sending = false
		h.mu.Unlock()
		return xerror.Wrap(err, "failed to send log to Axiom")
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return xerror.Errorf("Axiom API error: %s", resp.Status)
	}

	h.mu.Lock()
	h.sending = false
	h.mu.Unlock()
	return nil
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

func (h *AxiomHandler) Shutdown() error {
	close(h.done) // Signal the flusher to stop
	close(h.shutdownCh)
	h.wg.Wait()

	// Flush any remaining logs
	h.flush()

	return nil
}

func (h *AxiomHandler) sendLogsImmediate(ctx context.Context, logs []map[string]interface{}) error {
	url := fmt.Sprintf("/v1/datasets/%s/ingest", h.dataset)
	h.client.SetForceContentType("application/json")
	resp, err := h.client.Post(ctx, url, logs)
	if err != nil {
		return xerror.Wrap(err, "failed to send log to Axiom immediately")
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return xerror.Errorf("Axiom API error: %s", resp.Status)
	}
	return nil
}

// SetFlushInterval sets the interval at which logs are flushed to Axiom
func (h *AxiomHandler) SetFlushInterval(interval time.Duration) {
	h.flushInterval = interval
}

func (h *AxiomHandler) startFlusher() {
	go func() {
		ticker := time.NewTicker(h.flushInterval)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				h.flush()
			case <-h.done:
				return
			}
		}
	}()
}

// Add this method
func (h *AxiomHandler) flush() {
	h.mu.Lock()
	if len(h.buffer) == 0 {
		h.mu.Unlock()
		return
	}
	logs := h.buffer
	h.buffer = make([]map[string]interface{}, 0)
	h.mu.Unlock()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := h.sendLogsImmediate(ctx, logs); err != nil {
		xlog.Error("Error flushing logs to Axiom", "error", err)
	}
}
