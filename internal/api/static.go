package api

import _ "embed"

//go:embed web/widget.js
var widgetJS []byte

//go:embed web/embed.html
var embedHTML []byte
