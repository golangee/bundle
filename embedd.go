package bundle

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
	"strings"
	"time"
)

type srcFile struct {
	blobs       []*blob
	PackageName string
	Resources   []*resource
	Version     string
}

func (s *srcFile) Blobs() []*blob {
	return s.blobs
}

func (s *srcFile) addFile(fname string, name string, opts Options) error {
	fmt.Println(fname)
	buf, err := ioutil.ReadFile(fname)
	if err != nil {
		return err
	}

	stat, err := os.Stat(fname)
	if err != nil {
		return err
	}

	hash := sha256.Sum256(buf)
	packed := mustEncodeAscii85(mustBrotliCompress(buf))

	blb := s.getBlob(hash)
	if blb == nil {
		blb = &blob{
			Hash: hash,
			Data: strconv.Quote(packed),
		}
		s.blobs = append(s.blobs, blb)
	}

	res := &resource{
		Name:          name,
		Size:          stat.Size(),
		Mode:          stat.Mode(),
		LastMod:       stat.ModTime(),
		Sha265:        hex.EncodeToString(hash[:]),
		CacheUnpacked: !opts.DisableCacheUnpacked,
		CacheBrotli:   !opts.DisableCacheBrotli,
		CacheGzip:     !opts.DisableCacheGzip,
		ConstName:     blb.ConstName(),
	}

	s.Resources = append(s.Resources, res)

	return nil
}

func (s *srcFile) getBlob(hash [32]byte) *blob {
	for _, blob := range s.blobs {
		if blob.Hash == hash {
			return blob
		}
	}

	return nil
}

type blob struct {
	Hash [32]byte
	Data string
}

func (b *blob) ConstName() string {
	return "blob_" + hex.EncodeToString(b.Hash[:])
}

// name string, size int64, mode os.FileMode, lastMod time.Time, sha256 string, cacheUnpacked, cacheBrotli, cacheGzip bool, data string
type resource struct {
	Name          string
	Size          int64
	Mode          os.FileMode
	LastMod       time.Time
	Sha265        string
	CacheUnpacked bool
	CacheBrotli   bool
	CacheGzip     bool
	ConstName     string
}

func (r *resource) FactoryMethod() string {
	sb := strings.Builder{}
	sb.WriteString("bundle.NewResource(")
	sb.WriteString(strconv.Quote(r.Name) + ",")
	sb.WriteString(strconv.Itoa(int(r.Size)) + ",")
	sb.WriteString(strconv.Itoa(int(r.Mode)) + ",")

	sb.WriteString("time.Unix(")
	sb.WriteString(strconv.FormatInt(r.LastMod.Unix(), 10) + ",")
	sb.WriteString(strconv.Itoa(r.LastMod.Nanosecond()))
	sb.WriteString("),")
	sb.WriteString(strconv.Quote(r.Sha265) + ",")
	sb.WriteString(strconv.FormatBool(r.CacheUnpacked) + ",")
	sb.WriteString(strconv.FormatBool(r.CacheBrotli) + ",")
	sb.WriteString(strconv.FormatBool(r.CacheGzip) + ",")
	sb.WriteString(r.ConstName)
	sb.WriteString(")")
	return sb.String()
}
