build-desktop:
	cd ../rss-feed-frontend && npm run build:desktop
	cp -r ../rss-feed-frontend/dist ./frontend/dist
	GOOS=windows GOARCH=amd64 go build -ldflags="-H windowsgui" -o rss-reader.exe ./cmd/rss-feed-backend/desktop
	@echo "Done! rss-reader.exe is ready."

build-macos:
	cd ../rss-feed-frontend && npm run build:desktop
	cp -r ../rss-feed-frontend/dist ./frontend/dist
	GOOS=darwin GOARCH=amd64 go build -o rss-reader-macos ./cmd/rss-feed-backend/desktop
	./macos_package.sh 1.0 rss-reader-macos .

build-macos-arm:
	cd ../rss-feed-frontend && npm run build:desktop
	cp -r ../rss-feed-frontend/dist ./frontend/dist
	GOOS=darwin GOARCH=arm64 go build -o rss-reader-macos-arm ./cmd/rss-feed-backend/desktop
	./macos_package.sh 1.0 rss-reader-macos-arm .

build-all:
	cd ../rss-feed-frontend && npm run build:desktop
	cp -r ../rss-feed-frontend/dist ./frontend/dist
	GOOS=windows GOARCH=amd64 go build -ldflags="-H windowsgui" -o rss-reader.exe ./cmd/rss-feed-backend/desktop
	GOOS=darwin GOARCH=amd64 go build -o rss-reader-macos ./cmd/rss-feed-backend/desktop
	./macos_package.sh 1.0 rss-reader-macos .
	GOOS=darwin GOARCH=arm64 go build -o rss-reader-macos-arm ./cmd/rss-feed-backend/desktop
	./macos_package.sh 1.0 rss-reader-macos-arm .
	@echo "Done! All binaries are ready."

clean:
	rm -rf ./frontend/dist
	rm -rf "RSS Reader.app"
	rm -f rss-reader.exe rss-reader-macos rss-reader-macos-arm

.PHONY: build-desktop build-macos build-macos-arm build-all clean