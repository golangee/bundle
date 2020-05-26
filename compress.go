package bundle

import (
	"bytes"
	"compress/gzip"
	"encoding/ascii85"
	"encoding/base64"
	"github.com/andybalholm/brotli"
	"io/ioutil"
)

func mustGzipCompress(in []byte) []byte {
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

func mustBrotliCompress(in []byte) []byte {
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

func mustBrotliDecompress(in []byte) []byte {
	buf := bytes.NewBuffer(in)
	reader := brotli.NewReader(buf)
	b, err := ioutil.ReadAll(reader)
	if err != nil {
		panic(err)
	}
	return b
}

// mustDecodeBase64 panics, if str cannot be decoded
func mustDecodeBase64(str string) []byte {
	b, err := base64.StdEncoding.DecodeString(str)
	if err != nil {
		panic(err)
	}
	return b
}

// mustDecodeAscii85 panics, if str cannot be decoded
func mustDecodeAscii85(str string) []byte {
	dec := ascii85.NewDecoder(bytes.NewReader([]byte(str)))
	b, err := ioutil.ReadAll(dec)
	if err != nil {
		panic(err)
	}
	return b
}

func mustEncodeAscii85(buf []byte) string {
	tmp := &bytes.Buffer{}
	enc := ascii85.NewEncoder(tmp)
	_, err := enc.Write(buf)
	if err != nil {
		panic(err)
	}
	return tmp.String()
}

// mustUnzip panics, if b cannot be gunzipped
func mustUnzip(b []byte) []byte {
	reader, err := gzip.NewReader(bytes.NewReader(b))
	if err != nil {
		panic(err)
	}
	res, err := ioutil.ReadAll(reader)
	if err != nil {
		panic(err)
	}
	return res
}
