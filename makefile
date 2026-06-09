build-desktop:
	cd ../rss-feed-frontend && npm run build:desktop
	cp -r ../rss-feed-frontend/dist ./frontend/dist
	GOOS=windows GOARCH=amd64 go build -ldflags="-H windowsgui" -o rss-reader.exe ./cmd/rss-feed-backend/desktop
	@echo "Done! rss-reader.exe is ready."

clean:
	rm -rf ./frontend/dist
	rm -f rss-reader.exe

.PHONY: build-desktop clean