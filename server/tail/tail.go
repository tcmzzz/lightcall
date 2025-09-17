package tail

import (
	"os"
	"path"

	"github.com/nxadm/tail"
	"github.com/pocketbase/pocketbase/core"
)

type Handler interface {
	File() string
	Deal(app core.App, line string) error
}

func MustRegister(app core.App, h Handler) {
	filePath := h.File()

	// 检查文件路径是否为空
	if filePath == "" {
		app.Logger().Error("文件路径不能为空")
		panic("tail: 必须指定文件路径")
	}

	// 检查文件目录是否存在
	dirPath := path.Dir(filePath)
	if _, err := os.Stat(dirPath); err != nil {
		app.Logger().Error("文件目录不存在", "path", dirPath)
		panic("tail: 文件目录不存在 - " + dirPath)
	}

	// 创建信号通道用于协调退出
	sig := make(chan struct{})

	// TODO: 1. 文件tail状态, 恢复上次读取位置; 2. 记录文件状态, 测试新文件是否会错误从该位置读取
	// 启动文件tail
	t, err := tail.TailFile(filePath, tail.Config{
		Follow: true,
		ReOpen: true,
		Logger: tail.DiscardingLogger,
	})

	if err != nil {
		app.Logger().Error("启动文件监听失败", "file", filePath, "error", err)
		panic("tail: 无法启动监听 - " + filePath)
	}

	start := false
	// 绑定应用生命周期事件
	app.OnServe().BindFunc(func(e *core.ServeEvent) error {
		go func() {
			app.Logger().Info("开始监听文件", "file", filePath)
			start = true
			for line := range t.Lines {
				if err := h.Deal(app, line.Text); err != nil {
					app.Logger().Error("处理消息失败",
						"file", filePath,
						"error", err,
						"line", line.Text)
				}
			}
			close(sig)
		}()
		return e.Next()
	})

	app.OnTerminate().BindFunc(func(e *core.TerminateEvent) error {
		if err := t.Stop(); err != nil {
			app.Logger().Error("停止文件监听失败", "file", filePath, "error", err)
		}
		if start {
			<-sig
		}
		app.Logger().Info("文件监听已停止", "file", filePath)
		return e.Next()
	})
}
