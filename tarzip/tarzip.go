package main

import (
	"archive/tar"
	"compress/gzip"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

func main() {
	// Define command-line flags for source directory and output file
	srcDir := flag.String("src", "", "Source directory to archive")
	outFile := flag.String("out", "-", "Output file (use '-' for stdout)")
	flag.Parse()

	// Check if source directory is provided
	if *srcDir == "" {
		fmt.Println("Source directory is required")
		os.Exit(1)
	}

	// Initialize output writer
	var out io.Writer
	if *outFile == "-" {
		// If output is set to "-", use standard output
		out = os.Stdout
	} else {
		// Otherwise, create a file with the specified output file name
		f, err := os.Create(*outFile)
		if err != nil {
			fmt.Println("Error creating file:", err)
			os.Exit(1)
		}
		defer f.Close()
		out = f
	}

	// Create gzip writer for compression
	gz := gzip.NewWriter(out)

	// Create tar writer for archiving
	tw := tar.NewWriter(gz)

	// Walk through each file in the source directory
	err := filepath.Walk(*srcDir, func(file string, fi os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Create tar header based on file info
		hdr, err := tar.FileInfoHeader(fi, file)
		if err != nil {
			return err
		}

		// Set the relative path in the tar header
		hdr.Name, _ = filepath.Rel(*srcDir, file)

		// Write the header to the tar file
		if err := tw.WriteHeader(hdr); err != nil {
			return err
		}

		// If it's a regular file, write the file content to the tar
		if fi.Mode().IsRegular() {
			srcFile, err := os.Open(file)
			if err != nil {
				return err
			}
			defer srcFile.Close()

			// Copy file data to tar writer
			_, err = io.Copy(tw, srcFile)
			if err != nil {
				return err
			}
		}
		return nil
	})

	// Check for any errors during the walk process
	if err != nil {
		fmt.Println("Error walking the path:", err)
		os.Exit(1)
	}

	// Close tar writer explicitly to finalize tar archive
	if err := tw.Close(); err != nil {
		fmt.Println("Error closing tar.Writer:", err)
		os.Exit(1)
	}

	// Close gzip writer explicitly to finalize compression
	if err := gz.Close(); err != nil {
		fmt.Println("Error closing gzip.Writer:", err)
		os.Exit(1)
	}

	fmt.Fprintln(os.Stderr, "Archiving completed successfully.")
}
