package bundle

import (
	"net/http"
	"path/filepath"
	"strings"
)

var mimeTypes = map[string]string{
	".css":  "text/css; charset=utf-8",
	".gif":  "image/gif",
	".htm":  "text/html; charset=utf-8",
	".html": "text/html; charset=utf-8",
	".jpeg": "image/jpeg",
	".jpg":  "image/jpeg",
	".js":   "text/javascript; charset=utf-8",
	".mjs":  "text/javascript; charset=utf-8",
	".pdf":  "application/pdf",
	".png":  "image/png",
	".svg":  "image/svg+xml",
	".wasm": "application/wasm",
	".webp": "image/webp",
	".xml":  "text/xml; charset=utf-8",
	".map":  "application/json",
	".woff": "font/woff2",
	".ttf":  "font/ttf",
}

// Handle is currently a trivial implementation for delivering resources, however
// it will use the resources to support gzip and brotli compression and more importantly etags.
// It simply matches the url path against the name of a resource.
func Handle(resources ...*Resource) func(http.ResponseWriter, *http.Request) {
	files := make(map[string]*Resource)
	for _, r := range resources {
		files[r.name] = r
	}

	return func(writer http.ResponseWriter, request *http.Request) {
		path := request.URL.Path
		resource := files[path]
		if resource == nil {
			http.NotFound(writer, request)
			return
		}

		etag := request.Header.Get("If-None-Match")
		if etag == resource.sha256String {
			writer.WriteHeader(http.StatusNotModified)
			return
		}

		contentType := mimeTypes[strings.ToLower(filepath.Ext(path))]
		if contentType == "" {
			contentType = "application/octet"
		}

		writer.Header().Set("content-type", contentType)
		writer.Header().Set("cache-control", "no-cache")
		writer.Header().Set("etag", resource.sha256String)

		if strings.Contains(request.Header.Get("Accept-Encoding"), "br") {
			writer.Header().Set("Content-Encoding", "br")
			writer.Write(files[path].brotli())
			return
		} else {
			if strings.Contains(request.Header.Get("Accept-Encoding"), "gzip") {
				writer.Header().Set("Content-Encoding", "gzip")
				writer.Write(files[path].gzip())
			} else {
				writer.Write(files[path].unpack())
			}
		}
	}
}
