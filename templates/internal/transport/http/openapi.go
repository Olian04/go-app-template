// [[ when (modeIs "http") ]]
package http

import (
	"bytes"
	stdhttp "net/http"
	"strings"
	"time"

	"[[.ModulePath]]/api"
)

// Swagger UI is loaded from a pinned CDN build with subresource-integrity
// hashes, so the browser refuses a tampered or swapped asset. Nothing is
// vendored: embedding swagger-ui-dist would add ~11 MB to the binary, most of it
// source maps.
//
// If the UI must work without internet access, drop swagger-ui-bundle.js and
// swagger-ui.css next to this package, embed them, and point the tags below at
// the local paths. The spec endpoints below never depend on the network.
const (
	swaggerUIVersion = "5.17.14"
	swaggerUIBaseURL = "https://cdn.jsdelivr.net/npm/swagger-ui-dist@" + swaggerUIVersion
	swaggerUICSSHash = "sha384-wxLW6kwyHktdDGr6Pv1zgm/VGJh99lfUbzSn6HNHBENZlCN7W602k9VkGdxuFvPn"
	swaggerUIJSHash  = "sha384-wmyclcVGX/WhUkdkATwhaK1X1JtiNrr2EoYJ+diV3vj4v6OC5yCeSu+yW13SYJep"
)

// SpecPathYAML and friends are the documentation surface. They are separate from
// the API itself, so operators can turn them off without touching routing.
const (
	SpecPathYAML = "/openapi.yaml"
	SpecPathJSON = "/openapi.json"
	DocsPath     = "/docs"
)

const docsPage = `<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="utf-8">
<meta name="viewport" content="width=device-width, initial-scale=1">
<title>[[ .ModuleBasename ]] API</title>
<link rel="stylesheet" href="` + swaggerUIBaseURL + `/swagger-ui.css"
      integrity="` + swaggerUICSSHash + `" crossorigin="anonymous">
</head>
<body>
<div id="swagger-ui"></div>
<script src="` + swaggerUIBaseURL + `/swagger-ui-bundle.js"
        integrity="` + swaggerUIJSHash + `" crossorigin="anonymous"></script>
<script>
  window.ui = SwaggerUIBundle({
    url: "` + SpecPathYAML + `",
    dom_id: "#swagger-ui",
    deepLinking: true,
    tryItOutEnabled: true
  });
</script>
</body>
</html>
`

// docs serves the embedded contract and the Swagger UI shell.
type docs struct {
	specYAML []byte
	specJSON []byte
	modTime  time.Time
}

// newDocs validates and pre-renders everything up front, so a malformed spec
// fails at startup instead of on the first request.
func newDocs() (docs, error) {
	specJSON, err := api.SpecJSON()
	if err != nil {
		return docs{}, err
	}
	return docs{
		specYAML: api.SpecYAML,
		specJSON: specJSON,
		// The spec ships inside the binary, so the process start time is the
		// most honest "last modified" available for caching.
		modTime: time.Now(),
	}, nil
}

// register wires the documentation routes onto mux.
func (d docs) register(mux *stdhttp.ServeMux) {
	mux.Handle("GET "+SpecPathYAML, d.serveSpec("application/yaml", d.specYAML))
	mux.Handle("GET "+SpecPathJSON, d.serveSpec("application/json", d.specJSON))
	mux.Handle("GET "+DocsPath, stdhttp.HandlerFunc(func(w stdhttp.ResponseWriter, r *stdhttp.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		stdhttp.ServeContent(w, r, "index.html", d.modTime, strings.NewReader(docsPage))
	}))
	// Browsers and humans both try the trailing slash.
	mux.Handle("GET "+DocsPath+"/", stdhttp.RedirectHandler(DocsPath, stdhttp.StatusMovedPermanently))
}

func (d docs) serveSpec(contentType string, body []byte) stdhttp.Handler {
	return stdhttp.HandlerFunc(func(w stdhttp.ResponseWriter, r *stdhttp.Request) {
		w.Header().Set("Content-Type", contentType)
		// ServeContent handles Range, If-Modified-Since, and HEAD for free.
		stdhttp.ServeContent(w, r, "openapi", d.modTime, bytes.NewReader(body))
	})
}
