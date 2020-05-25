package bundle

import (
	"bytes"
	"fmt"
	"go/format"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

type Options struct {
	TargetDir         string
	PackageName       string
	Include           []string
	StoreUncompressed bool
	StoreGzip         bool
	StoreBrotli       bool
	StripPrefixes     []string // removes this prefix from all Include paths, if they begin with it
	Prefix            string   // attach this prefix to all included files
}

// Embed includes the given files or folders and creates a new go src file
func Embed(opts Options) error {
	var files []string
	for _, inc := range opts.Include {
		stat, err := os.Stat(inc)
		if err != nil {
			return err
		}

		if stat.IsDir() {
			fnames, err := scan(inc)
			if err != nil {
				return err
			}
			files = append(files, fnames...)
		}
	}

	src := &srcFile{
		PackageName: opts.PackageName,
	}
	for _, file := range files {
		name := file
		for _, strip := range opts.StripPrefixes {
			if strings.HasPrefix(file, strip) {
				name = file[len(strip):]
				break
			}
		}
		err := src.addFile(file, name, opts.StoreUncompressed, opts.StoreGzip, opts.StoreBrotli)
		if err != nil {
			return err
		}
	}

	tmp := &bytes.Buffer{}
	err := goTpl.Execute(tmp, src)
	if err != nil {
		return err
	}
	formatted, err := format.Source(tmp.Bytes())
	if err != nil {
		fmt.Println(string(tmp.Bytes()))
		return err
	}

	return ioutil.WriteFile(filepath.Join(opts.TargetDir, "bundle.gen.go"), formatted, os.ModePerm)
}
