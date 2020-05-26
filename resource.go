package bundle

import (
	"bytes"
	"io"
	"os"
	"sync"
	"time"
)

// A Resource relates a bunch of bytes with a name and optionally cached variants of the same data.
type Resource struct {
	name              string
	encoded           string // brotli + asci85
	size              int64  // original size
	cacheUnpacked     []byte
	cacheBrotli       []byte
	cacheGzip         []byte
	mustCacheUnpacked bool
	mustCacheBrotli   bool
	mustCacheGzip     bool
	mutex             sync.Mutex
	mode              os.FileMode
	lastMod           time.Time
	sha256String      string
}

func NewResource(name string, size int64, mode os.FileMode, lastMod time.Time, sha256 string, cacheUnpacked, cacheBrotli, cacheGzip bool, data string) *Resource {
	return &Resource{
		name:              name,
		encoded:           data,
		size:              size,
		mustCacheUnpacked: cacheUnpacked,
		mustCacheBrotli:   cacheBrotli,
		mustCacheGzip:     cacheGzip,
		mode:              mode,
		lastMod:           lastMod,
		sha256String:      sha256,
	}
}

func (r *Resource) Mode() os.FileMode {
	return r.mode
}

func (r *Resource) ModTime() time.Time {
	return r.lastMod
}

func (r *Resource) IsDir() bool {
	return false
}

func (r *Resource) Sys() interface{} {
	return nil
}

// Name returns the resources unique name
func (r *Resource) Name() string {
	return r.name
}

func (r *Resource) Size() int64 {
	return r.size
}

func (r *Resource) unpack() []byte {
	if r.cacheUnpacked != nil {
		return r.cacheUnpacked
	}

	b := mustBrotliDecompress(mustDecodeAscii85(r.encoded))
	if r.mustCacheUnpacked {
		// kind of double check idiom
		r.mutex.Lock()
		defer r.mutex.Unlock()
		r.cacheUnpacked = b
	}
	return b
}

// Read opens the resource to read the unpacked data.
func (r *Resource) Read() io.Reader {
	return bytes.NewReader(r.unpack())
}

func (r *Resource) gzip() []byte {
	if r.cacheGzip != nil {
		return r.cacheGzip
	}

	buf := r.unpack()
	buf = mustGzipCompress(buf)
	if r.mustCacheGzip {
		// kind of double check idiom
		r.mutex.Lock()
		defer r.mutex.Unlock()
		r.cacheGzip = buf
	}
	return buf
}

// ReadGzip opens the resource to read the data as gzip stream
func (r *Resource) ReadGzip() io.Reader {
	return bytes.NewReader(r.gzip())
}

func (r *Resource) brotli() []byte {
	if r.cacheBrotli != nil {
		return r.cacheBrotli
	}

	buf := mustDecodeAscii85(r.encoded)
	if r.mustCacheBrotli {
		// kind of double check idiom
		r.mutex.Lock()
		defer r.mutex.Unlock()
		r.cacheBrotli = buf
	}
	return buf
}

// ReadGzip opens the resource to read the data as gzip stream
func (r *Resource) ReadBrotli() io.Reader {
	return bytes.NewReader(r.brotli())
}

// WriteBrotli writes the datastream as a brotli buffer into the writer
func (r *Resource) WriteBrotli(dst io.Writer) (int, error) {
	return dst.Write(r.brotli())
}

// WriteGzip writes the datastream as a gzip buffer into the writer
func (r *Resource) WriteGzip(dst io.Writer) (int, error) {
	return dst.Write(r.gzip())
}

// Write transfers the uncompressed data into the writer
func (r *Resource) Write(dst io.Writer) (int, error) {
	return dst.Write(r.unpack())
}
