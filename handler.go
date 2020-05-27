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
	".json": "application/json",
	".woff": "font/woff2",
	".ttf":  "font/ttf",
}

// Handle is currently a trivial implementation for delivering resources, however
// it will use the resources to support gzip and brotli compression and more importantly etags.
// It simply matches the url path against the name of a resource.
func Handle(prefix string, resources ...*Resource) func(http.ResponseWriter, *http.Request) {
	files := make(map[string]*Resource)
	for _, r := range resources {
		files[r.name] = r
	}

	return func(writer http.ResponseWriter, request *http.Request) {
		path := request.URL.Path
		if strings.HasPrefix(path, prefix) {
			path = path[len(prefix):]
			if !strings.HasPrefix(path, "/") {
				path = "/" + path
			}
		}

		resource := files[path]
		if resource == nil {
			if path == "/" {
				resource = files["/index.html"]
				if resource == nil {
					resource = files["/index.htm"]
				}
			}

			if resource == nil {
				http.NotFound(writer, request)
				return
			}
		}

		contentType := mimeTypes[strings.ToLower(filepath.Ext(resource.Name()))]
		if contentType == "" {
			contentType = "application/octet"
		}

		writer.Header().Set("content-type", contentType)
		writer.Header().Set("cache-control", "no-cache")
		writer.Header().Set("etag", resource.sha256String)

		etag := request.Header.Get("If-None-Match")
		if etag == resource.sha256String {
			writer.WriteHeader(http.StatusNotModified)
			return
		}

		if strings.Contains(request.Header.Get("Accept-Encoding"), "br") {
			writer.Header().Set("Content-Encoding", "br")
			writer.Write(resource.brotli())
			return
		} else {
			if strings.Contains(request.Header.Get("Accept-Encoding"), "gzip") {
				writer.Header().Set("Content-Encoding", "gzip")
				writer.Write(resource.gzip())
			} else {
				writer.Write(resource.unpack())
			}
		}
	}
}
