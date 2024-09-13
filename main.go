package main

import (
	"crypto/sha256"
	"database/sql"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

type FileInfo struct {
	Path    string
	Size    int64
	ModTime time.Time
	Hash    string
}

func main() {

	username, exists := os.LookupEnv("username")
	
	if !exists{
		fmt.Println("username is not set")
	}
	password, exists := os.LookupEnv("password")
	host := "127.0.0.1"
	port := 3306
	dbName := "proddb"
	if !exists{
		fmt.Println("password is not set")
	}
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s", username, password, host, port, dbName)
	db, err := sql.Open("mysql", dsn)

	if err != nil {
		fmt.Println("Error happened opening database:", err)
		return
	}
	defer db.Close()

	for {
	root := "C:\\"

	err = filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			fmt.Printf("Error accesing path %q: %v\n", path, err)
			return filepath.SkipDir
		}
		if !info.IsDir() {
			fileInfo := getFileInfo(path, info)
			insertFileInfo(db, fileInfo)
		}
		return nil
	})
	if err != nil {
		fmt.Printf("Error happened walking path %q: %v\n", root, err)
	}
	}
}


func getFileInfo(path string, info os.FileInfo) FileInfo {
	hash := calculateHash(path)
	return FileInfo{
		Path:    path,
		Size:    info.Size(),
		ModTime: info.ModTime(),
		Hash:    hash,
	}
}

func calculateHash(path string) string {
	f, err := os.Open(path)
	if err != nil {
		fmt.Printf("Error opening file %s: %v\n", path, err)
		return ""
	}
	defer f.Close()

	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		fmt.Printf("Error calculating hash for %s: %v\n", path, err)
		return ""
	}

	return fmt.Sprintf("%x", h.Sum(nil))
}

func insertFileInfo(db *sql.DB, info FileInfo) {
	query :="INSERT INTO file_info (PathToFile, Size, ModTime,HashValue) VALUES (?,?,?,?)"
	_ ,err := db.Exec(query, info.Path, info.Size, info.ModTime, info.Hash)

	if err != nil {
		fmt.Printf("Error inserting file info for %s: %v\n", info.Path, err)
	}
}
