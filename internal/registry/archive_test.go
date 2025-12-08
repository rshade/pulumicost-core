package registry

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func TestSanitizePath(t *testing.T) {
	tmpDir := t.TempDir()

	tests := []struct {
		name    string
		destDir string
		path    string
		wantErr bool
	}{
		{
			name:    "valid path",
			destDir: tmpDir,
			path:    "subdir/file.txt",
			wantErr: false,
		},
		{
			name:    "simple filename",
			destDir: tmpDir,
			path:    "file.txt",
			wantErr: false,
		},
		{
			name:    "zip-slip attempt",
			destDir: tmpDir,
			path:    "../../../etc/passwd",
			wantErr: true,
		},
		{
			name:    "absolute path is made relative",
			destDir: tmpDir,
			path:    "/etc/passwd",
			wantErr: false, // filepath.Join makes absolute paths relative
		},
		{
			name:    "hidden path traversal",
			destDir: tmpDir,
			path:    "foo/../../../etc/passwd",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := sanitizePath(tt.destDir, tt.path)
			if (err != nil) != tt.wantErr {
				t.Errorf("sanitizePath() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateBinary(t *testing.T) {
	tmpDir := t.TempDir()

	// Create executable file
	execPath := filepath.Join(tmpDir, "executable")
	if err := os.WriteFile(execPath, []byte("binary"), 0750); err != nil {
		t.Fatal(err)
	}

	// Create non-executable file
	nonExecPath := filepath.Join(tmpDir, "nonexec")
	if err := os.WriteFile(nonExecPath, []byte("data"), 0644); err != nil {
		t.Fatal(err)
	}

	// Create directory
	dirPath := filepath.Join(tmpDir, "directory")
	if err := os.MkdirAll(dirPath, 0750); err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name    string
		path    string
		wantErr bool
	}{
		{
			name:    "valid executable",
			path:    execPath,
			wantErr: runtime.GOOS == "windows", // On Windows, needs .exe
		},
		{
			name:    "directory",
			path:    dirPath,
			wantErr: true,
		},
		{
			name:    "non-existent",
			path:    filepath.Join(tmpDir, "nonexistent"),
			wantErr: true,
		},
	}

	// On Unix, non-executable should fail
	if runtime.GOOS != "windows" {
		tests = append(tests, struct {
			name    string
			path    string
			wantErr bool
		}{
			name:    "non-executable file",
			path:    nonExecPath,
			wantErr: true,
		})
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateBinary(tt.path)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateBinary() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestExtractArchive(t *testing.T) {
	// Create test tar.gz archive
	tmpDir := t.TempDir()
	tarPath := filepath.Join(tmpDir, "test.tar.gz")
	destDir := filepath.Join(tmpDir, "extracted")

	// Create a simple tar.gz file
	createTestTarGz(t, tarPath, map[string]string{
		"file1.txt": "content1",
		"file2.txt": "content2",
	})

	if err := os.MkdirAll(destDir, 0750); err != nil {
		t.Fatal(err)
	}

	err := ExtractArchive(tarPath, destDir)
	if err != nil {
		t.Errorf("ExtractArchive() error = %v", err)
	}

	// Verify files extracted
	if _, err := os.Stat(filepath.Join(destDir, "file1.txt")); os.IsNotExist(err) {
		t.Error("file1.txt was not extracted")
	}
}

func TestExtractArchiveZip(t *testing.T) {
	tmpDir := t.TempDir()
	zipPath := filepath.Join(tmpDir, "test.zip")
	destDir := filepath.Join(tmpDir, "extracted")

	// Create a simple zip file
	createTestZip(t, zipPath, map[string]string{
		"file1.txt": "content1",
		"file2.txt": "content2",
	})

	if err := os.MkdirAll(destDir, 0750); err != nil {
		t.Fatal(err)
	}

	err := ExtractArchive(zipPath, destDir)
	if err != nil {
		t.Errorf("ExtractArchive() error = %v", err)
	}

	// Verify files extracted
	if _, err := os.Stat(filepath.Join(destDir, "file1.txt")); os.IsNotExist(err) {
		t.Error("file1.txt was not extracted")
	}
}

func TestExtractArchiveUnsupported(t *testing.T) {
	tmpDir := t.TempDir()
	unsupportedPath := filepath.Join(tmpDir, "test.rar")
	destDir := filepath.Join(tmpDir, "extracted")

	if err := os.WriteFile(unsupportedPath, []byte("fake"), 0644); err != nil {
		t.Fatal(err)
	}

	err := ExtractArchive(unsupportedPath, destDir)
	if err == nil {
		t.Error("expected error for unsupported archive format")
	}
}

func TestExtractArchiveNonExistent(t *testing.T) {
	tmpDir := t.TempDir()
	err := ExtractArchive(filepath.Join(tmpDir, "nonexistent.tar.gz"), tmpDir)
	if err == nil {
		t.Error("expected error for non-existent archive")
	}
}

func TestMaxFileSizeBoundary(t *testing.T) {
	// Test that maxFileSize constant is set to 500MB
	expectedSize := int64(500 * 1024 * 1024)
	if maxFileSize != expectedSize {
		t.Errorf("maxFileSize = %d, want %d (500MB)", maxFileSize, expectedSize)
	}

	// Test that the boundary is reasonable (greater than 100MB, less than 1GB)
	if maxFileSize <= 100*1024*1024 {
		t.Error("maxFileSize should be greater than 100MB for plugin compatibility")
	}
	if maxFileSize >= 1024*1024*1024 {
		t.Error("maxFileSize should be less than 1GB to prevent excessive memory usage")
	}
}

// Helper to create test tar.gz archives.
func createTestTarGz(t *testing.T, path string, files map[string]string) {
	t.Helper()

	f, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	gw := gzip.NewWriter(f)
	defer gw.Close()

	tw := tar.NewWriter(gw)
	defer tw.Close()

	for name, content := range files {
		hdr := &tar.Header{
			Name: name,
			Mode: 0644,
			Size: int64(len(content)),
		}
		if err := tw.WriteHeader(hdr); err != nil {
			t.Fatal(err)
		}
		if _, err := tw.Write([]byte(content)); err != nil {
			t.Fatal(err)
		}
	}
}

// Helper to create test zip archives.
func createTestZip(t *testing.T, path string, files map[string]string) {
	t.Helper()

	f, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	zw := zip.NewWriter(f)
	defer zw.Close()

	for name, content := range files {
		w, err := zw.Create(name)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := w.Write([]byte(content)); err != nil {
			t.Fatal(err)
		}
	}
}
