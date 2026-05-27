package main

import (
	"fmt"
	"net/http"

	"github.com/Swunci/rss-feed-backend/internal/app"
)

func main() {
	app := app.NewApp()

	fmt.Println("Server running on :8081")
	server_err := http.ListenAndServe("127.0.0.1:8081", app.Router)
	if server_err != nil {
		fmt.Println(server_err)
	}

}
