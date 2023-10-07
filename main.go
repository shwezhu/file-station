package main

import (
	"errors"
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

var root = "./files"

func handleUploadFile(w http.ResponseWriter, r *http.Request) {
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
	err = os.MkdirAll(root, os.ModePerm)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Create a file to store the file parsed from the request.
	var dst *os.File
	// Check if the filename exists first.
	_, err = os.Stat(fmt.Sprintf("%v/%s", root, fileHeader.Filename))
	// File already exists, create a file with another name.
	if err == nil {
		filename := strings.Split(fileHeader.Filename, ".")[0] +
			strconv.FormatInt(time.Now().UnixNano(), 10) +
			filepath.Ext(fileHeader.Filename)
		dst, err = os.Create(fmt.Sprintf("%v/%s", root, filename))
	} else if os.IsNotExist(err) {
		// File doesn't exist, create directly
		dst, err = os.Create(fmt.Sprintf("%v/%s", root, fileHeader.Filename))
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

func handleDownloadFile(w http.ResponseWriter, r *http.Request) {
	//w.Header().Set()
}

func handleHomePage(w http.ResponseWriter, r *http.Request) {
	if _, err := os.Stat(root); errors.Is(err, os.ErrNotExist) {
		err := os.Mkdir(root, os.ModePerm)
		if err != nil {
			log.Println(err)
		}
	}
	type File struct {
		Name string
		Path string
		Size string
	}
	// get files in the specified root folder
	files, err := os.ReadDir(root)
	if err != nil {
		http.Error(w, fmt.Errorf("failed to open root folder: %v", err).Error(), http.StatusInternalServerError)
		return
	}
	data := make([]File, 0)
	for _, f := range files {
		// ignore folder
		if f.IsDir() {
			continue
		}
		info, err := f.Info()
		if err != nil {
			log.Printf("failed to get info of %v: %v", f.Name(), err)
			continue
		}
		// size less than 100KB, displayed as KB
		var size string
		if info.Size() <= 102400 {
			size = fmt.Sprintf("%.2f KB", float32(info.Size())/(1<<10))
		} else {
			size = fmt.Sprintf("%.2f MB", float32(info.Size())/(1<<20))
		}
		data = append(data, File{
			Name: f.Name(),
			Path: fmt.Sprintf("%v/%v", root, f.Name()),
			Size: size,
		})
	}
	//pass files' info to the HTML template
	tmpl, err := template.ParseFiles("./template/home.html")
	err = tmpl.Execute(w, data)
	if err != nil {
		log.Printf("failed to execute template:%v", err)
	}
}

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/home", handleHomePage)
	mux.HandleFunc("/upload", handleUploadFile)
	mux.HandleFunc("/hi", handleDownloadFile)
	_ = http.ListenAndServe(":8080", mux)
}
