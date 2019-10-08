package main

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"text/template"

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
)

func NewGenerateCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "generate",
		Short:   "generate go source code",
		Aliases: []string{"gen"},
		RunE:    gen,
	}

	cmd.Flags().StringSliceVarP(&inputs, "input", "i", nil, "specify source you want to bind")
	cmd.Flags().StringSliceVarP(&transfer, "transfer", "t", nil, "specify transfer (eg: [org]:[new])")
	cmd.Flags().StringSliceVarP(&trims, "trim", "", nil, "trim empty line")
	cmd.Flags().StringVarP(&pkg, "package", "p", "", "package name")
	cmd.Flags().StringVarP(&output, "output", "o", "", "output name(default: [package]_gen.go)")
	cmd.Flags().StringSliceVarP(&ignores, "ignore", "", nil, "skip some files")
	cmd.Flags().BoolVarP(&verbose, "verbose", "", false, "print details")

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
		"package": pkg,
		"misc":    header,
		"entries": paths,
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
	gz := gzip.NewWriter(bw)

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
	gz := gzip.NewWriter(sw)

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
