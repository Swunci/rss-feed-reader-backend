package main

import (
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/exec"
	"runtime"
	"time"

	rssfeedbackend "github.com/Swunci/rss-feed-backend"
	"github.com/Swunci/rss-feed-backend/internal/app"
	"github.com/getlantern/systray"
)

const port = "7721"

func main() {
	app := app.NewApp(true)

	go func() {
		server_err := http.ListenAndServe(":"+port, app.Router)
		if server_err != nil {
			fmt.Println(server_err)
		}
	}()

	time.Sleep(1000 * time.Millisecond)

	openBrowser("http://localhost:" + port)
	systray.Run(onReady, onExit)
}

func onReady() {
	if runtime.GOOS == "darwin" {
		systray.SetIcon(rssfeedbackend.MacIcon)
	} else {
		systray.SetIcon(rssfeedbackend.Icon)
	}
	systray.SetTitle("RSS Reader")
	systray.SetTooltip("RSS Reader")
	mOpen := systray.AddMenuItem("Open", "")
	mQuit := systray.AddMenuItem("Quit", "")

	go func() {
		for {
			select {
			case <-mOpen.ClickedCh:
				openBrowser("http://localhost:" + port)
			case <-mQuit.ClickedCh:
				systray.Quit()
			}
		}
	}()
}

func onExit() {
	os.Exit(0)
}

func openBrowser(url string) {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "windows":
		cmd = exec.Command("rundll32", "url.dll,FileProtocolHandler", url)
	case "darwin":
		cmd = exec.Command("open", url)
	default:
		cmd = exec.Command("xdg-open", url)
	}
	if err := cmd.Start(); err != nil {
		slog.Error("Open browser: %v\n", "err", err)
	}
}
