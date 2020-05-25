package bundle

import (
	"bytes"
	"compress/gzip"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"github.com/andybalholm/brotli"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

type srcFile struct {
	blobs []*blob
	PackageName string
}

func (s *srcFile) Blobs() []*blob {
	return s.blobs
}

func (s *srcFile) addFile(fname string, name string, includeSrc, includeGzip, includeBrotli bool) error {
	b, err := ioutil.ReadFile(fname)
	if err != nil {
		return err
	}

	res := resource{
		Name: name,
	}

	if includeSrc {
		myBlob := s.dedub(b)
		res.ConstName = myBlob.ConstName()
	}

	if includeBrotli {
		fork := &resource{
			Name:      "br",
			ConstName: s.dedub(brotliCompress(b)).ConstName(),
		}
		res.Forks = append(res.Forks, fork)
	}

	if includeGzip {
		fork := &resource{
			Name:      "gzip",
			ConstName: s.dedub(gzipCompress(b)).ConstName(),
		}
		res.Forks = append(res.Forks, fork)
	}

	return nil
}

func (s *srcFile) dedub(b []byte) *blob {
	hash := sha256.Sum256(b)
	hashHex := hex.EncodeToString(hash[:])

	var myBlob *blob
	for _, blb := range s.blobs {
		if blb.Hash == hashHex {
			myBlob = blb
			break
		}
	}

	if myBlob == nil {
		myBlob = &blob{
			Hash:   hashHex,
			Base64: base64.StdEncoding.EncodeToString(b),
		}
		s.blobs = append(s.blobs, myBlob)
	}
	return myBlob
}

func gzipCompress(in []byte) []byte {
	buf := &bytes.Buffer{}
	writer, err := gzip.NewWriterLevel(buf, gzip.BestCompression)
	if err != nil {
		panic(err)
	}
	_, err = writer.Write(in)
	if err != nil {
		panic(err)
	}
	err = writer.Close()
	if err != nil {
		panic(err)
	}
	return buf.Bytes()
}

func brotliCompress(in []byte) []byte {
	buf := &bytes.Buffer{}
	writer := brotli.NewWriterLevel(buf, brotli.BestCompression)
	_, err := writer.Write(in)
	if err != nil {
		panic(err)
	}
	err = writer.Close()
	if err != nil {
		panic(err)
	}
	return buf.Bytes()
}

type blob struct {
	Hash   string
	Base64 string
}

func (b *blob) ConstName() string {
	return "blob_" + b.Base64
}

type resource struct {
	Name      string
	ConstName string
	Forks     []*resource
}

func scan(dir string) ([]string, error) {
	var r []string
	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if info.IsDir() && strings.HasPrefix(info.Name(), ".") {
			return filepath.SkipDir
		}
		if info.Mode().IsRegular() {
			r = append(r, path)
		}

		return nil
	})
	return r, err
}
