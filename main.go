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

	"gopkg.in/alecthomas/kingpin.v2"
)

var (
	app  = kingpin.New("bkup", "Rotating backup file utility. Only creates a backup on files that change since the last backup.")
	num  = app.Flag("num", "Number of rotating backup files.").Default("12").Short('n').Int()
	file = app.Arg("file", "File to be backed up.").Required().String()
)

// go build -o bkup.exe -ldflags -H=windowsgui .
func main() {
	// Usage: bkup.exe <file_to_be_backed_up>
	kingpin.MustParse(app.Parse(os.Args[1:]))

	// Ensure the file to be backed up exists
	if _, err := os.Stat(*file); os.IsNotExist(err) {
		log.Fatalf("Can't find [%s]: %v\n", *file, err)
	}

	// Create a log file in the directory of the file to be backed up
	path, err := filepath.Abs(filepath.Dir(*file))
	if err != nil {
		log.Fatalf("error determining path: [%v]\n", err)
	}
	logf, err := os.OpenFile(filepath.Join(path, "bkup.log"), os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("error opening file: %v", err)
	}
	defer logf.Close()
	log.SetOutput(logf)

	if *num <= 0 {
		log.Fatalf("num [%d] must be greater than 0", *num)
	}

	// Search for existing backup files
	files, err := filepath.Glob(*file + ".*")
	if err != nil {
		log.Fatal(err)
	}

	// Backup files are OTF: <file>.YYYYMMDDHHMMSS (14 digits)
	restr := fmt.Sprintf("%s\\.[0-9]{14}$", strings.Replace(*file, `\`, `\\`, -1))
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
		dst, err := backupFile(*file)
		if err != nil {
			log.Fatalf("error backing up %s: %v\n", *file, err)
		}
		log.Printf("Copied [%s] to [%s]\n", *file, dst)
		return
	}

	// Compare the last file (most recent) to the argument.
	sort.Strings(backups)

	// Do nothing if the file and most recent backup are the same.
	same, err := fileMD5Compare(*file, backups[len(backups)-1])
	if err != nil {
		log.Fatalf("error comparing files %v", err)
	}
	if same {
		return
	}

	// Backup the file.
	dst, err := backupFile(*file)
	if err != nil {
		log.Fatalf("error backing up %s: %v", *file, err)
	}
	log.Printf("Copied [%s] to [%s]\n", *file, dst)

	// Delete the oldest backup file if we are at max files.
	if len(backups) >= *num {
		err := os.Remove(backups[0])
		if err != nil {
			log.Fatalf("error deleting %s: %v\n", backups[0], err)
		}
		log.Printf("Deleted [%s]\n", backups[0])
		return
	}
}

func printUsage() {
	usage := fmt.Sprintf("Usage: %s <file_to_backup>\n", os.Args[0])
	fmt.Fprintf(os.Stderr, usage)
	log.Printf(usage)
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

func fileMD5Compare(file1, file2 string) (same bool, err error) {
	f1, err := os.Open(file1)
	if err != nil {
		return false, err
	}
	defer f1.Close()
	h1 := md5.New()
	if _, err := io.Copy(h1, f1); err != nil {
		return false, err
	}
	f1.Close()

	f2, err := os.Open(file2)
	if err != nil {
		return false, err
	}
	defer f2.Close()
	h2 := md5.New()
	if _, err := io.Copy(h2, f2); err != nil {
		return false, err
	}
	f2.Close()

	same = reflect.DeepEqual(h1, h2)

	return
}
