package main

import (
	"crypto/md5"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"time"
)

func main() {
	if len(os.Args) != 2 {
		printUsage()
		os.Exit(1)
	}

	f, err := os.Open(os.Args[1])
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	h := md5.New()
	if _, err := io.Copy(h, f); err != nil {
		log.Fatal(err)
	}
	f.Close()
	fmt.Printf("%x\n", h.Sum(nil))

	path := filepath.Dir(os.Args[1])
	fmt.Printf("%s\n", path)

	t := time.Now().Local()
	fmt.Printf("%s\n", t.Format("20060102150405"))

	files, err := filepath.Glob(os.Args[1] + "*")
	if err != nil {
		log.Fatal(err)
	}

	for _, file := range files {
		fmt.Printf("%s\n", file)
	}

	err = backupFile(os.Args[1])
	if err != nil {
		log.Fatal(err)
	}
}

func printUsage() {
	fmt.Printf("Usage: %s <file_to_backup>\n", os.Args[0])
}

func backupFile(file string) error {
	// Generate backup file name.
	// <filename>.<datetime>
	t := time.Now().Local()
	ext := t.Format("20060102150405")
	dst := file + "." + ext

	// Ensure the backup file doesn't already exist
	if _, err := os.Stat(dst); os.IsExist(err) {
		return os.ErrExist
	}

	in, err := os.Open(file)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, in)
	if err != nil {
		return err
	}
	return out.Close()
}
