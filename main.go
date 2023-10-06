package main

import (
	"fmt"
	"html/template"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

func handleUploadFile(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var maxUploadSize int64 = 1 << 25 // 32MB
	r.Body = http.MaxBytesReader(w, r.Body, maxUploadSize)
	if err := r.ParseMultipartForm(maxUploadSize); err != nil {
		http.Error(w, "Uploaded file is too big. Choose an file less than 32MB", http.StatusBadRequest)
		return
	}

	file, fileHeader, err := r.FormFile("uploaded_file")
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	defer file.Close()
	folderName := "files"
	// Create the uploads folder if it doesn't already exist
	err = os.MkdirAll(fmt.Sprintf("./%v", folderName), os.ModePerm)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	// Create a new file in the uploads directory
	dst, err := os.Create(fmt.Sprintf("./%v/%d%s", folderName, time.Now().UnixNano(), filepath.Ext(fileHeader.Filename)))
	if err != nil {
		http.Error(w, fmt.Errorf("failed to create dst file: %v", err).Error(), http.StatusInternalServerError)
		return
	}
	// make sure file will always be closed
	defer dst.Close()
	// Copy the uploaded file to the filesystem
	// at the specified destination
	_, err = io.Copy(dst, file)
	if err != nil {
		http.Error(w, fmt.Errorf("failed to copy file into dst file: %v", err).Error(), http.StatusInternalServerError)
		return
	}
	if err = dst.Close(); err != nil {
		http.Error(w, fmt.Errorf("failed to close dst file: %v", err).Error(), http.StatusInternalServerError)
		return
	}
	_, _ = fmt.Fprintf(w, "upload successful")
}

type file struct {
	Path string
	Size int
}

func handleDownloadFile(w http.ResponseWriter, r *http.Request) {
	files := []file{
		{
			Path: "hello",
			Size: 1,
		},
		{
			Path: "hihi",
			Size: 2,
		},
	}
	tmpl := template.Must(template.ParseFiles("template/index.template"))
	_ = tmpl.Execute(w, files)
}

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/login", handleUploadFile)
	mux.HandleFunc("/hi", handleDownloadFile)
	_ = http.ListenAndServe(":8080", mux)
}
