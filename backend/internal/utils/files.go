package utils

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

// ReadFile reads a file and returns its content as a string
func ReadFile(path string) (string, error) {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("failed to read file: %w", err)
	}
	return string(data), nil
}

// WriteFile writes content to a file
func WriteFile(path, content string) error {
	dir := filepath.Dir(path)
	if dir != "." {
		err := os.MkdirAll(dir, 0755)
		if err != nil {
			return fmt.Errorf("failed to create directory: %w", err)
		}
	}

	err := ioutil.WriteFile(path, []byte(content), 0644)
	if err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}
	return nil
}

// FileExists checks if a file exists
func FileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// IsDirectory checks if a path is a directory
func IsDirectory(path string) bool {
	info, err := os.Stat(path)
	return err == nil && info.IsDir()
}

// ListFiles lists all files in a directory
func ListFiles(dir string) ([]string, error) {
	var files []string
	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			files = append(files, path)
		}
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list files: %w", err)
	}
	return files, nil
}

// ListFilesWithExtension lists all files with a specific extension in a directory
func ListFilesWithExtension(dir, ext string) ([]string, error) {
	var files []string
	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && filepath.Ext(path) == ext {
			files = append(files, path)
		}
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list files: %w", err)
	}
	return files, nil
}

// GetFileExtension returns the extension of a file
func GetFileExtension(path string) string {
	return filepath.Ext(path)
}

// GetFileName returns the name of a file without the directory
func GetFileName(path string) string {
	return filepath.Base(path)
}

// GetFileNameWithoutExtension returns the name of a file without the extension
func GetFileNameWithoutExtension(path string) string {
	return strings.TrimSuffix(filepath.Base(path), filepath.Ext(path))
}

// GetFilesWithExtensions returns all files with any of the given extensions
func GetFilesWithExtensions(dir string, exts []string) ([]string, error) {
	var files []string
	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			fileExt := filepath.Ext(path)
			for _, ext := range exts {
				if fileExt == ext {
					files = append(files, path)
					break
				}
			}
		}
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list files: %w", err)
	}
	return files, nil
}

// DeleteFile deletes a file
func DeleteFile(path string) error {
	err := os.Remove(path)
	if err != nil {
		return fmt.Errorf("failed to delete file: %w", err)
	}
	return nil
}

// CreateTempFile creates a temporary file with the given content
func CreateTempFile(prefix, content string) (string, error) {
	tmpFile, err := ioutil.TempFile("", prefix)
	if err != nil {
		return "", fmt.Errorf("failed to create temp file: %w", err)
	}
	
	_, err = tmpFile.WriteString(content)
	if err != nil {
		return "", fmt.Errorf("failed to write to temp file: %w", err)
	}
	
	err = tmpFile.Close()
	if err != nil {
		return "", fmt.Errorf("failed to close temp file: %w", err)
	}
	
	return tmpFile.Name(), nil
}