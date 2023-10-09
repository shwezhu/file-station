package main

import (
	"fmt"
	"github.com/shwezhu/sessions"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

func (s *Server) handleUpload(w http.ResponseWriter, r *http.Request, _ *sessions.Session) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	// parse file from request body
	var maxUploadSize int64 = 1 << 25 // 32MB
	r.Body = http.MaxBytesReader(w, r.Body, maxUploadSize)
	if err := r.ParseMultipartForm(maxUploadSize); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		log.Print(err)
		return
	}
	// obtain file from form
	file, fileHeader, err := r.FormFile("uploaded_file")
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	defer file.Close()

	// Create a root folder if it doesn't exist
	err = os.MkdirAll(s.fileRoot, os.ModePerm)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Create a file to store the file parsed from the request.
	var dst *os.File
	// Check if the filename exists first.
	_, err = os.Stat(fmt.Sprintf("%v/%s", s.fileRoot, fileHeader.Filename))
	// File already exists, create a file with another name.
	if err == nil {
		filename := strings.Split(fileHeader.Filename, ".")[0] +
			strconv.FormatInt(time.Now().UnixNano(), 10) +
			filepath.Ext(fileHeader.Filename)
		dst, err = os.Create(fmt.Sprintf("%v/%s", s.fileRoot, filename))
	} else if os.IsNotExist(err) {
		// File doesn't exist, create directly
		dst, err = os.Create(fmt.Sprintf("%v/%s", s.fileRoot, fileHeader.Filename))
	}
	if err != nil {
		http.Error(w, fmt.Errorf("failed to create dst file: %v", err).Error(), http.StatusInternalServerError)
		return
	}
	defer dst.Close()

	// Copy the uploaded file to the filesystem
	// at the specified destination
	_, err = io.Copy(dst, file)
	if err != nil {
		http.Error(w, fmt.Errorf("failed to copy file into dst file: %v", err).Error(), http.StatusInternalServerError)
		return
	}
	// We need get error when close a writable file, because of buffering
	if err = dst.Close(); err != nil {
		http.Error(w, fmt.Errorf("failed to close dst file: %v", err).Error(), http.StatusInternalServerError)
		return
	}
	_, _ = fmt.Fprintf(w, "upload successful")
}

func (s *Server) handleDownload(w http.ResponseWriter, r *http.Request, _ *sessions.Session) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, fmt.Sprintf("failed to download file: %v", err), http.StatusBadRequest)
		return
	}
	filename := r.Form.Get("filename")
	if filename == "" {
		http.Error(w, "failed to download, no filename provided", http.StatusBadRequest)
		return
	}
	path := fmt.Sprintf("%v/%s", s.fileRoot, filename)
	// Check if specified exists.
	_, err := os.Stat(path)
	// File exists, send the file.
	if err == nil {
		w.Header().Set("Content-Disposition", "attachment; filename="+strconv.Quote(filename))
		w.Header().Set("Content-Type", "application/octet-stream")
		http.ServeFile(w, r, path)
	} else if os.IsNotExist(err) {
		http.Error(w, "no such file", http.StatusNotFound)
	} else {
		http.Error(w, fmt.Sprintf("failed to download file %v", err), http.StatusInternalServerError)
	}
}
