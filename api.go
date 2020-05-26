package bundle

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"go/format"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

const constPrefixHash = "const BundleVersion = "

type Options struct {
	TargetDir            string
	PackageName          string
	Include              []string
	StripPrefixes        []string // removes this prefix from all Include paths, if they begin with it
	Prefix               string   // attach this prefix to all included files
	IgnoreRegex          string   // e.g. '.*\.map|^\..*' will ignore all map and hidden files from inclusion
	DisableCacheUnpacked bool
	DisableCacheGzip     bool
	DisableCacheBrotli   bool
}

type Bundle struct {
	resources []*Resource
}

func Make(resources ...*Resource) *Bundle {
	return &Bundle{
		resources: resources,
	}
}

// Embed includes the given files or folders and creates a new go src file. It expects a working dir somewhere
// within a go module and picks the root itself.
func Embed(opts Options) error {
	cwd, err := modRoot()
	if err != nil {
		return err
	}

	ignoreRegex, err := regexp.Compile(opts.IgnoreRegex)
	if err != nil {
		return err
	}

	fmt.Println("working dir", cwd)
	var files []string
	for _, inc := range opts.Include {
		fname := filepath.Clean(inc)
		if !filepath.IsAbs(fname) {
			fname = filepath.Join(cwd, fname)
		}
		stat, err := os.Stat(fname)
		if err != nil {
			return err
		}

		if stat.IsDir() {
			fnames, err := scan(fname, ignoreRegex)
			if err != nil {
				return err
			}
			files = append(files, fnames...)
		}
	}

	totalSize, err := stat(files)
	if err != nil {
		return err
	}

	fmt.Printf("found %d files, total %d bytes (%fMB)\n", len(files), totalSize, float32(totalSize)/1024/1024)

	requiredHash, err := fileHash(files, opts)
	if err != nil {
		return err
	}

	targetGenFile := filepath.Join(opts.TargetDir, "bundle.gen.go")
	foundHash, err := extractFileHash(targetGenFile)
	if err != nil {
		return err
	}

	if requiredHash == foundHash {
		fmt.Println("bundle is already up to date, nothing to do")
		return nil
	}

	if opts.StripPrefixes == nil {
		opts.StripPrefixes = []string{cwd}
	}

	src := &srcFile{
		PackageName: opts.PackageName,
		Version:     requiredHash,
	}

	for _, file := range files {
		name := file
		for _, strip := range opts.StripPrefixes {
			if strings.HasPrefix(file, strip) {
				name = file[len(strip):]
				break
			}
		}

		err := src.addFile(file, name, opts)
		if err != nil {
			return err
		}
	}

	tmp := &bytes.Buffer{}
	err = goTpl.Execute(tmp, src)
	if err != nil {
		return err
	}

	formatted, err := format.Source(tmp.Bytes())
	if err != nil {
		fmt.Println(string(tmp.Bytes()))
		return err
	}

	return ioutil.WriteFile(targetGenFile, formatted, os.ModePerm)
}

func modRoot() (string, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return "", err
	}
	root := cwd
	for {
		stat, err := os.Stat(filepath.Join(root, "go.mod"))
		if err == nil && stat.Mode().IsRegular() {
			return root, nil
		}
		root = filepath.Dir(root)
		if root == "/" || root == "." {
			return "", fmt.Errorf("%s is not withing a go module", cwd)
		}
	}
}

func fileHash(files []string, opts interface{}) (string, error) {
	sort.Strings(files)
	hash := sha256.New()
	for _, file := range files {
		buf, err := ioutil.ReadFile(file)
		if err != nil {
			return "", err
		}
		_, err = hash.Write(buf)
		if err != nil {
			return "", err
		}
	}
	if opts != nil {
		tmp, err := json.Marshal(opts)
		if err != nil {
			return "", err
		}
		_, err = hash.Write(tmp)
		if err != nil {
			return "", err
		}
	}
	return hex.EncodeToString(hash.Sum(nil)), nil
}

func stat(files []string) (int64, error) {
	sum := int64(0)
	for _, file := range files {
		stat, err := os.Stat(file)
		if err != nil {
			return 0, err
		}
		sum += stat.Size()
	}
	return sum, nil
}

func extractFileHash(fname string) (string, error) {
	if _, err := os.Stat(fname); err != nil {
		return "", nil
	}

	file, err := os.Open(fname)
	if err != nil {
		return "", err
	}
	defer file.Close()

	buf := make([]byte, 1024)
	_, err = file.Read(buf)
	if err != nil {
		return "", err
	}

	for _, line := range strings.Split(string(buf), "\n") {
		line = strings.TrimSpace(line)

		if strings.HasPrefix(line, constPrefixHash) {
			hash := line[len(constPrefixHash):]
			hash = strings.ReplaceAll(hash, "\"", "")
			return strings.TrimSpace(hash), nil
		}
	}
	return "", fmt.Errorf(constPrefixHash + " not found")
}

func scan(dir string, expIgnore *regexp.Regexp) ([]string, error) {
	var r []string
	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if info.IsDir() && strings.HasPrefix(info.Name(), ".") {
			return filepath.SkipDir
		}

		if info.Mode().IsRegular() && !expIgnore.MatchString(info.Name()) {
			r = append(r, path)
		}

		return nil
	})
	return r, err
}
