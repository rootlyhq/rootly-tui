package debug

import (
	"bytes"
	"encoding/json"
	"io"
	"os"
	"sync"

	"github.com/charmbracelet/log"
)

// ringBuffer is a thread-safe circular buffer for log entries
type ringBuffer struct {
	mu      sync.RWMutex
	entries []string
	maxSize int
	pos     int
	full    bool
}

func newRingBuffer(size int) *ringBuffer {
	return &ringBuffer{
		entries: make([]string, size),
		maxSize: size,
	}
}

func (r *ringBuffer) Write(p []byte) (n int, err error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	entry := string(p)
	r.entries[r.pos] = entry
	r.pos = (r.pos + 1) % r.maxSize
	if r.pos == 0 {
		r.full = true
	}
	return len(p), nil
}

func (r *ringBuffer) GetEntries() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if !r.full {
		return r.entries[:r.pos]
	}

	// Return entries in order (oldest first)
	result := make([]string, r.maxSize)
	copy(result, r.entries[r.pos:])
	copy(result[r.maxSize-r.pos:], r.entries[:r.pos])
	return result
}

func (r *ringBuffer) Clear() {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.entries = make([]string, r.maxSize)
	r.pos = 0
	r.full = false
}

var (
	// LogBuffer stores recent log entries in memory
	LogBuffer = newRingBuffer(1000)

	// Logger is the global debug logger
	Logger *log.Logger

	// Enabled indicates if debug mode is active
	Enabled bool

	// fileOutput holds file writer if logging to file
	fileOutput io.Writer
)

func init() {
	// Create logger that writes to buffer (always) and optionally to stderr/file
	Logger = log.NewWithOptions(LogBuffer, log.Options{
		ReportTimestamp: true,
		Prefix:          "rootly-tui",
		Level:           log.DebugLevel, // Always capture to buffer
	})
}

// PrettyJSON formats JSON bytes for readable logging
func PrettyJSON(data []byte) string {
	var prettyJSON bytes.Buffer
	if err := json.Indent(&prettyJSON, data, "", "  "); err != nil {
		return string(data) // Return raw if indent fails
	}
	return prettyJSON.String()
}

// Enable turns on debug logging to stderr
func Enable() {
	Enabled = true
	if fileOutput != nil {
		Logger.SetOutput(io.MultiWriter(LogBuffer, fileOutput))
	} else {
		Logger.SetOutput(io.MultiWriter(LogBuffer, os.Stderr))
	}
	Logger.Debug("Debug mode enabled")
}

// Disable turns off debug logging to stderr (still logs to buffer)
func Disable() {
	Enabled = false
	Logger.SetOutput(LogBuffer)
}

// SetLogFile writes logs to a file in addition to buffer
func SetLogFile(path string) error {
	f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return err
	}
	fileOutput = f
	Logger.SetOutput(io.MultiWriter(LogBuffer, f))
	return nil
}

// GetLogs returns all log entries from the buffer
func GetLogs() []string {
	return LogBuffer.GetEntries()
}

// ClearLogs clears the log buffer
func ClearLogs() {
	LogBuffer.Clear()
}
