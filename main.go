package main

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
	"time"

	"github.com/alecthomas/kingpin/v2"
)

var (
	app      = kingpin.New("bkup", "Rotating backup file utility. Only creates a backup on files that change since the last backup.")
	num      = app.Flag("num", "Number of rotating backup files.").Default("12").Short('n').Int()
	filespec = app.Arg("file", "Path and filename (with wildcards '*' or '?') to be backed up.").Required().String()
)

// go build -o bkup.exe -ldflags -H=windowsgui .
func main() {
	// Create a log file in the directory of the file(s) to be backed up
	fmt.Println("HELLO - bkup.exe")
	fmt.Printf("[%s]\n", *filespec)
	fmt.Printf("[%s]\n", filepath.Dir(*filespec))
	fmt.Printf("[%s]\n", filepath.Join(filepath.Dir(*filespec), "bkup.log"))
	logf, err := os.OpenFile(filepath.Join(filepath.Dir(*filespec), "bkup.log"), os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("error opening file: %v\r\n", err)
	}
	defer logf.Close()
	log.SetOutput(logf)

	// Usage: bkup.exe <file_to_be_backed_up> [-n <num_of_backups>]
	// Validate parameters
	kingpin.MustParse(app.Parse(os.Args[1:]))
	if *num <= 0 {
		log.Fatalf("num [%d] must be greater than 0\r\n", *num)
	}
	if *filespec == "" {
		log.Fatalf("file must be specified\r\n")
	}

	// Loop through the save game file names and back them up
	savedGameFileNames, err := getSavedGameFileNames(*filespec)
	if err != nil {
		log.Fatalf("error getting saved game file names: %v\r\n", err)
	}
	if len(savedGameFileNames) == 0 {
		log.Fatalf("no files found matching [%s]\r\n", *filespec)
	}
	for _, file := range savedGameFileNames {
		backups, err := getBackupFiles(file)
		if err != nil {
			log.Fatalf("error getting backup files: %v\r\n", err)
		}
		if len(backups) == 0 {
			dst, err := backupFile(file)
			if err != nil {
				log.Fatalf("error backing up %s: %v\r\n", file, err)
			}
			log.Printf("Copied [%s] to [%s]\r\n", file, dst)
			continue
		}
		slices.Sort(backups)
		same, err := CompareFiles(file, backups[len(backups)-1])
		if err != nil {
			log.Fatalf("error comparing files %v\r\n", err)
		}
		if same {
			continue
		}
		dst, err := backupFile(file)
		if err != nil {
			log.Fatalf("error backing up %s: %v\r\n", file, err)
		}
		log.Printf("Copied [%s] to [%s]\r\n", file, dst)

		// Delete the oldest backup file if we are at max files.
		if len(backups) >= *num {
			err := os.Remove(backups[0])
			if err != nil {
				log.Fatalf("error deleting %s: %v\r\n", backups[0], err)
			}
			log.Printf("Deleted [%s]\r\n", backups[0])
		}
	}
}

func getSavedGameFileNames(fullpath string) (filenames []string, err error) {
	// If the fullpath is a file (no "*" or "?"), return it
	if !strings.Contains(fullpath, "*") && !strings.Contains(fullpath, "?") {
		return []string{fullpath}, nil
	}
	filenames, err = filepath.Glob(fullpath)
	return
}

func getBackupFiles(file string) (backups []string, err error) {
	// Search for existing backup files
	files, err := filepath.Glob(file + ".*")
	if err != nil {
		return
	}

	// Backup files are OTF: <file>.YYYYMMDDHHMMSS (14 digits)
	restr := fmt.Sprintf("%s\\.[0-9]{14}$", strings.Replace(file, `\`, `\\`, -1))
	re, err := regexp.Compile(restr)
	if err != nil {
		return
	}

	for _, backup := range files {
		if re.MatchString(backup) {
			backups = append(backups, backup)
		}
	}
	return
}

func backupFile(file string) (dst string, err error) {
	// Generate backup file name.
	// <filename>.<datetime>
	t := time.Now().Local()
	ext := t.Format("20060102150405")
	dst = file + "." + ext

	// Ensure the backup file doesn't already exist
	if _, err := os.Stat(dst); os.IsExist(err) {
		return dst, os.ErrExist
	}

	in, err := os.Open(file)
	if err != nil {
		return dst, err
	}
	defer in.Close()

	out, err := os.Create(dst)
	if err != nil {
		return dst, err
	}
	defer out.Close()

	_, err = io.Copy(out, in)
	if err != nil {
		return dst, err
	}
	return dst, out.Close()
}

// CompareFiles does a deep compare of files; returns true if they are identical
func CompareFiles(file1, file2 string) (same bool, err error) {
	// BufSize of 1 MB Chunks
	const BufSize = 1024 * 1024

	f1, err := os.Open(file1)
	if err != nil {
		return
	}
	defer f1.Close()

	f2, err := os.Open(file2)
	if err != nil {
		return
	}
	defer f2.Close()

	for {
		b1 := make([]byte, BufSize)
		_, err1 := f1.Read(b1)

		b2 := make([]byte, BufSize)
		_, err2 := f2.Read(b2)

		if err1 != nil || err2 != nil {
			// We're out of file to compare
			if err1 == io.EOF && err2 == io.EOF {
				return true, nil
			} else if err1 == io.EOF || err2 == io.EOF {
				return false, nil
			} else {
				if err1 != nil {
					return false, err1
				}
				return false, err2
			}
		}

		if !bytes.Equal(b1, b2) {
			return false, err
		}
	}
}
