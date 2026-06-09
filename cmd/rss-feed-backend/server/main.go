package main

import (
	"fmt"
	"net/http"
	"os"

	"github.com/Swunci/rss-feed-backend/internal/app"
	"github.com/joho/godotenv"
)

func main() {
	if os.Getenv("APP_ENV") != "production" {
		err := godotenv.Load(".env")
		if err != nil {
			panic(err)
		}
	}

	app := app.NewApp(false)
	var port = os.Getenv("PORT")
	fmt.Println("Server running on :" + port)
	server_err := http.ListenAndServe(":"+port, app.Router)
	if server_err != nil {
		fmt.Println(server_err)
	}

}
