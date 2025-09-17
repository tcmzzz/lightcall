package appender

import (
	"log/slog"
	"os"
	"path"
	"sync"

	"github.com/pocketbase/pocketbase/core"
)

type Handler interface {
	File() string
	MustRegister(core.App, *Appender) error
}

// Appender handles asynchronous, queued writing to a log file.
type Appender struct {
	file    *os.File
	queue   chan string
	wg      sync.WaitGroup
	closeCh chan struct{}
	logger  *slog.Logger
}

// New creates and starts a new Appender.
// It opens the specified log file for appending, panicking if it fails.
// It also starts a goroutine to process the write queue.
func New(logFile string, logger *slog.Logger) *Appender {
	file, err := os.OpenFile(logFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		logger.Error("appender: Failed to open append log file", "file", logFile, "error", err)
		panic("appender: Failed to open append log file " + logFile + ": " + err.Error())
	}

	a := &Appender{
		file:    file,
		queue:   make(chan string, 100), // Buffered channel
		closeCh: make(chan struct{}),
		logger:  logger,
	}

	a.wg.Add(1)
	go a.run()

	return a
}

// run is the processing loop for the write queue.
func (a *Appender) run() {
	defer a.wg.Done()
	a.logger.Info("Appender: processing goroutine started.")
	for {
		select {
		case msg := <-a.queue:
			if _, err := a.file.WriteString(msg + "\n"); err != nil {
				a.logger.Error("Error writing to append log", "error", err)
			}
		case <-a.closeCh:
			a.logger.Info("Appender: Shutdown signaled. Draining queue...")
			// Drain the remaining messages from the queue.
			for {
				select {
				case msg := <-a.queue:
					if _, err := a.file.WriteString(msg + "\n"); err != nil {
						a.logger.Error("Error writing to append log during drain", "error", err)
					}
				default:
					// Queue is empty, we can exit.
					a.logger.Info("Appender: Queue drained. Exiting.")
					return
				}
			}
		}
	}
}

// Append adds a message to the write queue. It is non-blocking.
func (a *Appender) Append(msg string) {
	// Check if the appender is shutting down.
	select {
	case <-a.closeCh:
		a.logger.Warn("Appender is shutting down. Message dropped.")
		return
	default:
		// Not closing, proceed.
	}

	// Try to send the message, but don't block if the queue is full.
	select {
	case a.queue <- msg:
		a.logger.Debug("Append line to buffer success.", "line", msg)
	default:
		a.logger.Warn("Append queue is full. Dropping message.")
	}
}

// Close gracefully shuts down the appender.
func (a *Appender) Close() {
	a.logger.Info("Appender: Closing...")
	close(a.closeCh) // Signal the run goroutine to start shutdown
	a.wg.Wait()      // Wait for the run goroutine to finish draining and exit
	a.file.Close()   // Close the file descriptor
	a.logger.Info("Appender: Closed.")
}

// MustRegister initializes the appender and registers the handler
func MustRegister(app core.App, handler Handler) *Appender {
	logFile := handler.File()

	// 检查文件路径是否为空
	if logFile == "" {
		app.Logger().Error("日志文件路径不能为空")
		panic("appender: 必须指定日志文件路径")
	}

	// 检查文件目录是否存在
	dirPath := path.Dir(logFile)
	if _, err := os.Stat(dirPath); err != nil {
		app.Logger().Error("日志文件目录不存在", "path", dirPath)
		panic("appender: 文件目录不存在 - " + dirPath)
	}

	// 创建新的 appender 实例
	appender := New(logFile, app.Logger())

	// 注册handler
	if err := handler.MustRegister(app, appender); err != nil {
		app.Logger().Error("Handler注册失败", "error", err)
		panic("appender: Handler注册失败 - " + err.Error())
	}

	// 绑定应用终止事件，确保优雅关闭
	app.OnTerminate().BindFunc(func(e *core.TerminateEvent) error {
		appender.Close()
		return e.Next()
	})

	app.Logger().Info("Appender with Handler 已注册", "file", logFile)
	return appender
}
