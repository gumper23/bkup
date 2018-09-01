package main

import (
	"crypto/md5"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"time"
)

func main() {
	maxBackupFiles := 10

	logf, err := os.OpenFile("c:\\users\\gumper\\documents\\bkup.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("error opening file: %v", err)
	}
	defer logf.Close()
	log.SetOutput(logf)

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

	files, err := filepath.Glob(os.Args[1] + "*")
	if err != nil {
		log.Fatal(err)
	}

	// Create a backup file if one doesn't exist.
	if len(files) == 1 {
		backupFile(os.Args[1])
		os.Exit(0)
	}

	// Assume backup files exist OTF: <file>.<date>
	// Compare the last file (most recent) to the argument.
	sort.Strings(files)
	lf, err := os.Open(files[len(files)-1])
	if err != nil {
		log.Fatal(err)
	}
	defer lf.Close()

	lh := md5.New()
	if _, err := io.Copy(lh, lf); err != nil {
		log.Fatal(err)
	}
	lf.Close()

	// Do nothing if the file to be backed up and
	// the most recent backup file contain the same data.
	if reflect.DeepEqual(h, lh) {
		os.Exit(0)
	}

	// Backup the file.
	backupFile(os.Args[1])

	// Delete the oldest backup file if we are at max files.
	if len(files) == maxBackupFiles+1 {
		os.Remove(files[1])
	}
}

func printUsage() {
	usage := fmt.Sprintf("Usage: %s <file_to_backup>\n", os.Args[0])
	fmt.Fprintf(os.Stderr, usage)
	log.Printf(usage)
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
