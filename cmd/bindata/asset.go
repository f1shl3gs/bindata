package main

const header = `import (
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
)

type fileSystem struct {
	files map[string]file
}

func (fs *fileSystem) Open(name string) (http.File, error) {
	name = strings.Replace(name, "//", "/", -1)
	f, ok := fs.files[name]
	if ok {
		return newHTTPFile(f, false)
	}
	index := strings.Replace(name+"/index.html", "//", "/", -1)
	f, ok = fs.files[index]
	if !ok {
		return nil, os.ErrNotExist
	}
	
	return newHTTPFile(f, true)
}

type file struct {
	os.FileInfo

	// the compressed data
	data []byte

	// the decompressed and cached for later usage
	cache []byte
}

type fileInfo struct {
	name    string
	size    int64
	mode    os.FileMode
	modTime time.Time
	isDir   bool

	files []os.FileInfo
}

func (f *fileInfo) Name() string {
	return f.name
}

func (f *fileInfo) Size() int64 {
	return f.size
}

func (f *fileInfo) Mode() os.FileMode {
	return f.mode
}

func (f *fileInfo) ModTime() time.Time {
	return f.modTime
}

func (f *fileInfo) IsDir() bool {
	return f.isDir
}

func (f *fileInfo) Readdir(count int) ([]os.FileInfo, error) {
	return make([]os.FileInfo, 0), nil
}

func (f *fileInfo) Sys() interface{} {
	return nil
}

func bindataRead(data []byte) ([]byte, error) {
	gz, err := gzip.NewReader(bytes.NewBuffer(data))
	if err != nil {
		return nil, err
	}

	var buf bytes.Buffer
	_, err = io.Copy(&buf, gz)
	clErr := gz.Close()

	if err != nil {
		return nil, err
	}
	if clErr != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func newHTTPFile(file file, isDir bool) (*httpFile, error) {
	if file.cache == nil && file.data != nil {
		cache, err := bindataRead(file.data)
		if err != nil {
			return nil, fmt.Errorf("read %s failed", file.Name())
		}

		file.cache = cache
	}

	return &httpFile{
		file:   file,
		reader: bytes.NewReader(file.cache),
		isDir:  isDir,
	}, nil
}

type httpFile struct {
	file

	reader *bytes.Reader
	isDir  bool
}

func (f *httpFile) Read(p []byte) (n int, err error) {
	return f.reader.Read(p)
}

func (f *httpFile) Seek(offset int64, whence int) (ret int64, err error) {
	return f.reader.Seek(offset, whence)
}

func (f *httpFile) Stat() (os.FileInfo, error) {
	return f, nil
}

func (f *httpFile) IsDir() bool {
	return f.isDir
}

func (f *httpFile) Readdir(count int) ([]os.FileInfo, error) {
	return make([]os.FileInfo, 0), nil
}

func (f *httpFile) Close() error {
	return nil
}

// New returns an embedded http.FileSystem
func New() http.FileSystem {
	return &fileSystem{
		files: files,
	}
}

// Lookup returns the file at the specified path
func Lookup(path string) ([]byte, error) {
	f, ok := files[path]
	if !ok {
		return nil, os.ErrNotExist
	}
	return f.data, nil
}

// MustLookup returns the file at the specified path
// and panics if the file is not found.
func MustLookup(path string) []byte {
	d, err := Lookup(path)
	if err != nil {
		panic(err)
	}
	return d
}

`
