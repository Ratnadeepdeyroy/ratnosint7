// Copyright 2026 Ratnadeep.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Package storage provides output writers for scan results.
package storage

import (
	"bufio"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// Writer writes domains to output.
type Writer interface {
	Write(domain string) error
	Close() error
}

// TextWriter writes one domain per line.
type TextWriter struct {
	w    *bufio.Writer
	file *os.File
	path string
	tmp  string
	done chan struct{}
}

// NewTextWriter creates a TextWriter that writes to path.
// Writes to a .tmp file and renames to path on Close().
func NewTextWriter(path string) (*TextWriter, error) {
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return nil, fmt.Errorf("create output dir: %w", err)
	}
	tmp := path + ".tmp"
	f, err := os.Create(tmp)
	if err != nil {
		return nil, fmt.Errorf("create file: %w", err)
	}

	tw := &TextWriter{
		w:    bufio.NewWriter(f),
		file: f,
		path: path,
		tmp:  tmp,
		done: make(chan struct{}),
	}

	go tw.flushLoop()
	return tw, nil
}

func (tw *TextWriter) flushLoop() {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-tw.done:
			return
		case <-ticker.C:
			_ = tw.w.Flush()
		}
	}
}

func (tw *TextWriter) Write(domain string) error {
	_, err := tw.w.WriteString(domain + "\n")
	return err
}

func (tw *TextWriter) Close() error {
	close(tw.done)
	if err := tw.w.Flush(); err != nil {
		tw.file.Close()
		os.Remove(tw.tmp)
		return err
	}
	if err := tw.file.Close(); err != nil {
		os.Remove(tw.tmp)
		return err
	}
	return os.Rename(tw.tmp, tw.path)
}

// WriteStream consumes from a channel and writes to a file with periodic flush.
func WriteStream(in <-chan string, path string) (int, error) {
	tw, err := NewTextWriter(path)
	if err != nil {
		return 0, err
	}
	count := 0
	for d := range in {
		if err := tw.Write(d); err != nil {
			tw.Close()
			return count, err
		}
		count++
	}
	return count, tw.Close()
}

// OutputPath returns the path for scan output.
func OutputPath(domain string, overwrite bool, format string) string {
	dir := "output"
	ext := ".txt"
	switch format {
	case "json":
		ext = ".json"
	case "csv":
		ext = ".csv"
	}
	if overwrite {
		return filepath.Join(dir, domain+ext)
	}
	return filepath.Join(dir, fmt.Sprintf("%s-%d%s", domain, time.Now().Unix(), ext))
}

// JSONWriter writes JSON array of domains (for --format json).
type JSONWriter struct {
	w     *bufio.Writer
	file  *os.File
	path  string
	tmp   string
	first bool
}

// NewJSONWriter creates a JSONWriter.
func NewJSONWriter(path string) (*JSONWriter, error) {
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return nil, fmt.Errorf("create output dir: %w", err)
	}
	tmp := path + ".tmp"
	f, err := os.Create(tmp)
	if err != nil {
		return nil, fmt.Errorf("create file: %w", err)
	}
	jw := &JSONWriter{w: bufio.NewWriter(f), file: f, path: path, tmp: tmp, first: true}
	if _, err := jw.w.WriteString("[\n"); err != nil {
		f.Close()
		os.Remove(tmp)
		return nil, err
	}
	return jw, nil
}

func (jw *JSONWriter) Write(domain string) error {
	if !jw.first {
		if _, err := jw.w.WriteString(",\n"); err != nil {
			return err
		}
	}
	jw.first = false
	b, _ := json.Marshal(domain)
	_, err := jw.w.Write(append([]byte("  "), b...))
	return err
}

func (jw *JSONWriter) Close() error {
	if _, err := jw.w.WriteString("\n]\n"); err != nil {
		jw.file.Close()
		os.Remove(jw.tmp)
		return err
	}
	if err := jw.w.Flush(); err != nil {
		jw.file.Close()
		os.Remove(jw.tmp)
		return err
	}
	if err := jw.file.Close(); err != nil {
		os.Remove(jw.tmp)
		return err
	}
	return os.Rename(jw.tmp, jw.path)
}

// CSVWriter writes CSV with domain column.
type CSVWriter struct {
	cw   *csv.Writer
	file *os.File
	path string
	tmp  string
}

// NewCSVWriter creates a CSVWriter.
func NewCSVWriter(path string) (*CSVWriter, error) {
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return nil, fmt.Errorf("create output dir: %w", err)
	}
	tmp := path + ".tmp"
	f, err := os.Create(tmp)
	if err != nil {
		return nil, fmt.Errorf("create file: %w", err)
	}
	cw := &CSVWriter{cw: csv.NewWriter(f), file: f, path: path, tmp: tmp}
	if err := cw.cw.Write([]string{"domain"}); err != nil {
		f.Close()
		os.Remove(tmp)
		return nil, err
	}
	return cw, nil
}

func (cw *CSVWriter) Write(domain string) error {
	return cw.cw.Write([]string{domain})
}

func (cw *CSVWriter) Close() error {
	cw.cw.Flush()
	if err := cw.cw.Error(); err != nil {
		cw.file.Close()
		os.Remove(cw.tmp)
		return err
	}
	if err := cw.file.Close(); err != nil {
		os.Remove(cw.tmp)
		return err
	}
	return os.Rename(cw.tmp, cw.path)
}

// NewWriter creates a Writer based on format.
func NewWriter(path, format string) (Writer, error) {
	switch format {
	case "json":
		return NewJSONWriter(path)
	case "csv":
		return NewCSVWriter(path)
	default:
		return NewTextWriter(path)
	}
}

