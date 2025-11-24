package registry

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

const (
	// maxFileSize limits decompression to prevent zip bombs (100MB).
	maxFileSize = 100 * 1024 * 1024
	// dirPerm is the permission for directories.
	dirPerm = 0750
	// execPerm is the permission for executable files.
	execPerm = 0750
)

// ExtractArchive extracts an archive to the destination directory.
// Supports tar.gz and zip formats based on file extension.
func ExtractArchive(archivePath, destDir string) error {
	if strings.HasSuffix(archivePath, ".tar.gz") || strings.HasSuffix(archivePath, ".tgz") {
		return extractTarGz(archivePath, destDir)
	}
	if strings.HasSuffix(archivePath, ".zip") {
		return extractZip(archivePath, destDir)
	}
	return fmt.Errorf("unsupported archive format: %s", archivePath)
}

// extractTarGz extracts a tar.gz archive.
func extractTarGz(archivePath, destDir string) error {
	file, err := os.Open(archivePath)
	if err != nil {
		return fmt.Errorf("failed to open archive: %w", err)
	}
	defer func() { _ = file.Close() }()

	gzr, err := gzip.NewReader(file)
	if err != nil {
		return fmt.Errorf("failed to create gzip reader: %w", err)
	}
	defer func() { _ = gzr.Close() }()

	tr := tar.NewReader(gzr)

	for {
		header, headerErr := tr.Next()
		if headerErr == io.EOF {
			break
		}
		if headerErr != nil {
			return fmt.Errorf("failed to read tar entry: %w", headerErr)
		}

		if extractErr := extractTarEntry(tr, header, destDir); extractErr != nil {
			return extractErr
		}
	}
	return nil
}

// extractTarEntry extracts a single tar entry.
func extractTarEntry(tr *tar.Reader, header *tar.Header, destDir string) error {
	target, err := sanitizePath(destDir, header.Name)
	if err != nil {
		return err
	}

	switch header.Typeflag {
	case tar.TypeDir:
		return os.MkdirAll(target, dirPerm)
	case tar.TypeReg:
		return extractTarFile(tr, header, target)
	}
	return nil
}

// extractTarFile extracts a regular file from tar.
func extractTarFile(tr *tar.Reader, header *tar.Header, target string) error {
	if err := os.MkdirAll(filepath.Dir(target), dirPerm); err != nil {
		return fmt.Errorf("failed to create parent directory: %w", err)
	}

	outFile, err := os.Create(target)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}

	written, copyErr := io.CopyN(outFile, tr, maxFileSize+1)
	closeErr := outFile.Close()

	if copyErr != nil && !errors.Is(copyErr, io.EOF) {
		return fmt.Errorf("failed to extract file: %w", copyErr)
	}
	if closeErr != nil {
		return fmt.Errorf("failed to close file: %w", closeErr)
	}
	if written > maxFileSize {
		return fmt.Errorf("file exceeds maximum size: %s", target)
	}

	// Preserve executable bit
	if header.Mode&0111 != 0 {
		if chmodErr := os.Chmod(target, execPerm); chmodErr != nil {
			return fmt.Errorf("failed to set permissions: %w", chmodErr)
		}
	}
	return nil
}

// extractZip extracts a zip archive.
func extractZip(archivePath, destDir string) error {
	r, err := zip.OpenReader(archivePath)
	if err != nil {
		return fmt.Errorf("failed to open zip: %w", err)
	}
	defer func() { _ = r.Close() }()

	for _, f := range r.File {
		if extractErr := extractZipFile(f, destDir); extractErr != nil {
			return extractErr
		}
	}
	return nil
}

// extractZipFile extracts a single file from zip.
func extractZipFile(f *zip.File, destDir string) error {
	target, err := sanitizePath(destDir, f.Name)
	if err != nil {
		return err
	}

	if f.FileInfo().IsDir() {
		return os.MkdirAll(target, dirPerm)
	}

	if mkdirErr := os.MkdirAll(filepath.Dir(target), dirPerm); mkdirErr != nil {
		return fmt.Errorf("failed to create parent directory: %w", mkdirErr)
	}

	return writeZipFile(f, target)
}

// writeZipFile writes a zip file entry to disk.
func writeZipFile(f *zip.File, target string) error {
	outFile, err := os.Create(target)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}

	rc, err := f.Open()
	if err != nil {
		_ = outFile.Close()
		return fmt.Errorf("failed to open file in zip: %w", err)
	}

	written, copyErr := io.CopyN(outFile, rc, maxFileSize+1)
	_ = rc.Close()
	closeErr := outFile.Close()

	if copyErr != nil && !errors.Is(copyErr, io.EOF) {
		return fmt.Errorf("failed to extract file: %w", copyErr)
	}
	if closeErr != nil {
		return fmt.Errorf("failed to close file: %w", closeErr)
	}
	if written > maxFileSize {
		return fmt.Errorf("file exceeds maximum size: %s", target)
	}

	// Set executable permission on Unix
	if f.Mode()&0111 != 0 {
		if chmodErr := os.Chmod(target, execPerm); chmodErr != nil {
			return fmt.Errorf("failed to set permissions: %w", chmodErr)
		}
	}
	return nil
}

// sanitizePath validates and sanitizes archive paths to prevent zip-slip attacks.
func sanitizePath(destDir, name string) (string, error) {
	// Clean the destination directory first
	cleanDest := filepath.Clean(destDir)
	// Join with the name from the archive
	target := filepath.Join(cleanDest, name)
	// Ensure the target is within the destination
	if !strings.HasPrefix(target, cleanDest+string(os.PathSeparator)) && target != cleanDest {
		return "", fmt.Errorf("invalid file path in archive: %s", name)
	}
	return target, nil
}

// ValidateBinary checks if the installed binary is executable.
func ValidateBinary(binaryPath string) error {
	info, err := os.Stat(binaryPath)
	if err != nil {
		return fmt.Errorf("binary not found: %w", err)
	}

	if info.IsDir() {
		return fmt.Errorf("expected file but found directory: %s", binaryPath)
	}

	// On Windows, check for .exe extension
	if runtime.GOOS == osWindows {
		if !strings.HasSuffix(strings.ToLower(binaryPath), ".exe") {
			return fmt.Errorf("windows binary must have .exe extension: %s", binaryPath)
		}
		return nil
	}

	// On Unix, check executable permission
	if info.Mode()&0111 == 0 {
		return fmt.Errorf("binary is not executable: %s", binaryPath)
	}

	return nil
}
