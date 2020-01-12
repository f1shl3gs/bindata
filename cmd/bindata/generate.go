package main

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"text/template"
	"time"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

var (
	inputs   []string
	output   string
	transfer []string
	pkg      string
	trims    []string
	ignores  []string
	verbose  bool
	tags     []string

	// gzip
	gzipBestCompress bool
	gzipBestSpeed    bool
	gzipLevel        int
)

func NewGenerateCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "generate",
		Short:   "generate go source code",
		Aliases: []string{"gen"},
		PreRunE: func(cmd *cobra.Command, args []string) error {
			if gzipBestSpeed {
				gzipLevel = gzip.BestSpeed
			}

			if gzipBestCompress {
				gzipLevel = gzip.BestCompression
			}

			if gzipLevel < gzip.HuffmanOnly || gzipLevel > gzip.BestCompression {
				return errors.Errorf("invalid gzip compress level %d", gzipLevel)
			}

			return nil
		},
		RunE: gen,
	}

	cmd.Flags().StringSliceVarP(&inputs, "input", "i", nil, "specify source you want to bind")
	cmd.Flags().StringSliceVarP(&transfer, "transfer", "t", nil, "specify transfer (eg: [org]:[new])")
	cmd.Flags().StringSliceVarP(&trims, "trim", "", nil, "trim empty line")
	cmd.Flags().StringVarP(&pkg, "package", "p", "", "package name")
	cmd.Flags().StringVarP(&output, "output", "o", "", "output name(default: [package]_gen.go)")
	cmd.Flags().StringSliceVarP(&ignores, "ignore", "", nil, "skip some files")
	cmd.Flags().BoolVarP(&verbose, "verbose", "", false, "print details")
	cmd.Flags().StringSliceVarP(&tags, "tag", "", nil, "add tags")
	cmd.Flags().BoolVarP(&gzipBestCompress, "gzip-best-compress", "", false, "compress with level 9, decompress only happened when the asset first accessed")
	cmd.Flags().BoolVarP(&gzipBestSpeed, "gzip-best-speed", "", false, "compress with level 1")
	cmd.Flags().IntVarP(&gzipLevel, "gzip-level", "", -1, "specify compress level")

	_ = cmd.MarkFlagRequired("input")
	_ = cmd.MarkFlagRequired("package")

	return cmd
}

func gen(cmd *cobra.Command, args []string) error {
	paths := make(map[string]string)
	transFunc := transferFunc(transfer)

	if verbose {
		fmt.Println("transfers", transfer)
	}

	if output == "" {
		output = pkg + "_gen.go"
	}

	if output == "" {
		return errors.New("output can't be empty")
	}

	for _, input := range inputs {
		err := filepath.Walk(input, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}

			if info.IsDir() {
				return nil
			}

			newPath := transFunc(path)
			if verbose && path != newPath {
				fmt.Println("read path", path, "transferred", newPath)
			}

			_, exist := paths[newPath]
			if exist {
				return errors.Errorf("duplicate entry %q", newPath)
			}

			paths[newPath] = path
			return nil
		})

		if err != nil {
			return err
		}
	}

	funcs := map[string]interface{}{
		"bytes":  bytesWriter,
		"string": stringWriter,
		"stat":   stat,
	}

	tmpl, err := template.New("bindata").Funcs(funcs).Parse(outputTemplate)
	if err != nil {
		panic(err)
	}

	data := map[string]interface{}{
		"tags": func() string {
			if len(tags) == 0 {
				return ""
			}

			return "// +build " + strings.Join(tags, " ")
		}(),
		"package":   pkg,
		"misc":      header,
		"entries":   paths,
		"timestamp": time.Now(),
		"gzip":      gzipLevel,
	}

	f, err := os.OpenFile(output, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		return errors.Wrap(err, "open output file failed")
	}
	defer f.Close()

	return tmpl.Execute(f, data)
}

func bytesWriter(file string) string {
	buf := bytes.NewBuffer(nil)

	f, err := os.Open(file)
	if err != nil {
		panic(err)
	}

	defer f.Close()

	bw := &ByteWriter{Writer: buf}
	gz, _ := gzip.NewWriterLevel(bw, gzipLevel)
	_, err = io.Copy(gz, f)
	if err != nil {
		panic(err)
	}

	gz.Close()

	return buf.String()
}

func stringWriter(file string) string {
	buf := bytes.NewBuffer(nil)

	f, err := os.Open(file)
	if err != nil {
		panic(err)
	}

	defer f.Close()

	sw := &StringWriter{Writer: buf}
	gz, _ := gzip.NewWriterLevel(sw, gzipLevel)

	_, err = io.Copy(gz, f)
	if err != nil {
		panic(err)
	}

	gz.Close()

	return buf.String()
}

func stat(file string) os.FileInfo {
	info, err := os.Stat(file)
	if err != nil {
		panic(err)
	}

	return info
}
