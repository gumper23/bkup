package main

import (
	"crypto/md5"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"sort"
	"strings"
	"time"
)

// go build -o bkup.exe -ldflags -H=windowsgui .
func main() {
	// Intended to be ran on a schedule of once every five minutes
	// One hour's worth of backups
	maxBackupFiles := 12

	// Usage: bkup.exe <file_to_be_backed_up>
	if len(os.Args) != 2 {
		printUsage()
		os.Exit(1)
	}

	// Ensure the file to be backed up exists
	if _, err := os.Stat(os.Args[1]); os.IsNotExist(err) {
		log.Fatalf("Can't find [%s]: %v\n", os.Args[1], err)
	}

	// Create a log file in the directory of the file to be backed up
	path, err := filepath.Abs(filepath.Dir(os.Args[1]))
	if err != nil {
		log.Fatalf("error determining path: [%v]\n", err)
	}
	logf, err := os.OpenFile(filepath.Join(path, "bkup.log"), os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("error opening file: %v", err)
	}
	defer logf.Close()
	log.SetOutput(logf)

	// Calculate the MD5 hash of the file
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

	// Search for existing backup files
	files, err := filepath.Glob(os.Args[1] + ".*")
	if err != nil {
		log.Fatal(err)
	}

	// Backup files are OTF: <file>.YYYYMMDDHHMMSS (14 digits)
	restr := fmt.Sprintf("%s\\.[0-9]{14}$", strings.Replace(os.Args[1], `\`, `\\`, -1))
	re, err := regexp.Compile(restr)
	if err != nil {
		log.Fatal(err)
	}

	var backups []string
	for _, backup := range files {
		if re.MatchString(backup) {
			backups = append(backups, backup)
		}
	}

	// Create a backup file if one doesn't exist.
	if len(backups) == 0 {
		backupFile(os.Args[1])
		os.Exit(0)
	}

	// Compare the last file (most recent) to the argument.
	sort.Strings(backups)
	lf, err := os.Open(backups[len(backups)-1])
	if err != nil {
		log.Fatal(err)
	}
	defer lf.Close()

	lh := md5.New()
	if _, err := io.Copy(lh, lf); err != nil {
		log.Fatal(err)
	}
	lf.Close()

	// Do nothing if the hashes of the argument and
	// the most recent backup match
	if reflect.DeepEqual(h, lh) {
		os.Exit(0)
	}

	// Backup the file.
	backupFile(os.Args[1])

	// Delete the oldest backup file if we are at max files.
	if len(backups) == maxBackupFiles {
		os.Remove(backups[0])
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
