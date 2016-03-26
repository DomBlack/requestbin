package main

import (
	"archive/zip"
	"io"
	"os"
	"path/filepath"
	"text/template"
	"time"
)

func writeHeader(archive *zip.Writer, filename string) (io.Writer, error) {
	header := &zip.FileHeader{
		Name:         filename,
		Method:       zip.Store,
		ModifiedTime: uint16(time.Now().UnixNano()),
		ModifiedDate: uint16(time.Now().UnixNano()),
	}
	writer, err := archive.CreateHeader(header)
	if err != nil {
		return nil, err
	}
	return writer, nil
}

func addFileToArchive(archive *zip.Writer, source string, filename string) error {
	writer, err := writeHeader(archive, filename)
	if err != nil {
		return err
	}

	file, err := os.Open(filepath.Join(source, filename))
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = io.Copy(writer, file)
	if err != nil {
		return err
	}
	return nil
}

func addTemplatedFileToArchive(archive *zip.Writer, source string, filename string, url string) error {
	writer, err := writeHeader(archive, filename)
	if err != nil {
		return err
	}

	t := template.Must(template.New("template").ParseFiles(filepath.Join(source, filename)))
	params := struct{ Url string }{Url: url}
	err = t.ExecuteTemplate(writer, "document", params)
	return err
}

func generateODT(source string, writer io.Writer, url string) error {
	archive := zip.NewWriter(writer)
	defer archive.Close()

	if err := addFileToArchive(archive, source, "mimetype"); err != nil {
		return err
	}
	if err := addFileToArchive(archive, source, "META-INF/manifest.xml"); err != nil {
		return err
	}
	if err := addTemplatedFileToArchive(archive, source, "content.xml", url); err != nil {
		return err
	}

	return nil
}
