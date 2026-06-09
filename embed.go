package rssfeedbackend

import "embed"

//go:embed all:frontend/dist
var FrontendAssets embed.FS

//go:embed all:cmd/rss-feed-backend/desktop/tempicon.ico
var Icon []byte
