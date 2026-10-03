package update

import (
	"archive/zip"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

// BuildUpdateZip creates a temporary ZIP archive with the specified EDR layout.
// It maps the contents of parentDir/Common and parentDir/Packed into "x64/",
// creates a dummy file in "x86/", and adds versionFile to the root.
func BuildUpdateZip(parentDir, versionFile, outPath string) error {
	outFile, err := os.Create(outPath)
	if err != nil {
		return fmt.Errorf("failed to create zip file: %w", err)
	}
	defer outFile.Close()

	zw := zip.NewWriter(outFile)
	defer zw.Close()

	// Construct paths based on the parent directory
	commonDir := filepath.Join(parentDir, "Common")
	if err := appendDirToZip(zw, commonDir, "x64"); err != nil {
		return fmt.Errorf("failed adding Common dir (%s) to zip: %w", commonDir, err)
	}

	packedDir := filepath.Join(parentDir, "Packed")
	if err := appendDirToZip(zw, packedDir, "x64"); err != nil {
		return fmt.Errorf("failed adding Packed dir (%s) to zip: %w", packedDir, err)
	}

	if err := writeZipFile(zw, "x86/.keep", []byte("x86 placeholder")); err != nil {
		return fmt.Errorf("failed creating x86 dummy file: %w", err)
	}

	if err := addFileToZip(zw, versionFile, "version.json"); err != nil {
		return fmt.Errorf("failed adding version.json: %w", err)
	}

	return nil
}

func appendDirToZip(zw *zip.Writer, sourceDir, zipPrefix string) error {
	return filepath.Walk(sourceDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}

		relPath, err := filepath.Rel(sourceDir, path)
		if err != nil {
			return err
		}

		zipPath := filepath.ToSlash(filepath.Join(zipPrefix, relPath))
		return addFileToZip(zw, path, zipPath)
	})
}

func addFileToZip(zw *zip.Writer, physicalPath, zipPath string) error {
	file, err := os.Open(physicalPath)
	if err != nil {
		return err
	}
	defer file.Close()

	w, err := zw.Create(zipPath)
	if err != nil {
		return err
	}

	_, err = io.Copy(w, file)
	return err
}

func writeZipFile(zw *zip.Writer, zipPath string, data []byte) error {
	w, err := zw.Create(zipPath)
	if err != nil {
		return err
	}
	_, err = w.Write(data)
	return err
}