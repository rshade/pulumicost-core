package migration

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// DetectLegacy checks if the legacy ~/.finfocus directory exists.
func DetectLegacy() (string, bool) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", false
	}
	legacyPath := filepath.Join(home, ".finfocus")
	info, err := os.Stat(legacyPath)
	if err != nil {
		return "", false
	}
	return legacyPath, info.IsDir()
}

// GetNewPath returns the path to the new ~/.finfocus directory.
func GetNewPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get user home directory: %w", err)
	}
	return filepath.Join(home, ".finfocus"), nil
}

// SafeCopy recursively copies the source directory to the destination.
// It preserves original data by performing a copy instead of a move.
func SafeCopy(src, dst string) error {
	return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Calculate relative path to source root
		rel, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}

		target := filepath.Join(dst, rel)

		if info.IsDir() {
			return os.MkdirAll(target, info.Mode())
		}

		return copyFile(path, target)
	})
}

// RunMigration handles the interactive migration from legacy to new configuration.
func RunMigration(out io.Writer, in io.Reader) error {
	legacyPath, exists := DetectLegacy()
	if !exists {
		return nil
	}

	newPath, err := GetNewPath()
	if err != nil {
		return err
	}

	// If new path already exists, don't prompt for migration
	if _, err := os.Stat(newPath); err == nil {
		return nil
	}

	fmt.Fprintf(out, "Detected legacy configuration at %s.\n", legacyPath)
	fmt.Fprintf(out, "Would you like to migrate to %s? [y/N] ", newPath)

	var response string
	fmt.Fscanln(in, &response)
	response = strings.ToLower(strings.TrimSpace(response))

	if response != "y" && response != "yes" {
		fmt.Fprintln(out, "Migration skipped. Please note that legacy configuration will be ignored unless FINFOCUS_COMPAT=1 is set.")
		return nil
	}

	fmt.Fprintln(out, "Migrating configuration...")
	if err := SafeCopy(legacyPath, newPath); err != nil {
		return fmt.Errorf("migration failed: %w", err)
	}

	fmt.Fprintf(out, "Migration complete. Your old config has been preserved at %s.\n", legacyPath)
	return nil
}

func copyFile(src, dst string) error {
	sourceFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer sourceFile.Close()

	// Ensure parent directory exists
	if err := os.MkdirAll(filepath.Dir(dst), 0700); err != nil {
		return err
	}

	destFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destFile.Close()

	_, err = io.Copy(destFile, sourceFile)
	if err != nil {
		return err
	}

	sourceInfo, err := os.Stat(src)
	if err != nil {
		return err
	}

	return os.Chmod(dst, sourceInfo.Mode())
}
