package utils

import (
	"github.com/aesoper101/x/fileutil"
	"io/fs"
	"path/filepath"
)

func GetThriftFiles(paths []string) []string {
	return getFiles(paths, isThriftFile)
}

func GetProtoFiles(paths []string) []string {
	return getFiles(paths, isProtoFile)
}

func GetYamlFiles(paths []string) []string {
	return getFiles(paths, isYamlFile)
}

func getFiles(paths []string, filter func(path string) bool) []string {
	var files []string
	for _, path := range paths {
		if fileutil.IsDir(path) {
			_ = filepath.Walk(path, func(p string, info fs.FileInfo, err error) error {
				if !info.IsDir() && filter(p) {
					files = append(files, p)
				}
				return nil
			})
		} else if filter(path) {
			files = append(files, path)
		}
	}

	return files
}

func isThriftFile(path string) bool {
	return filepath.Ext(path) == ".thrift"
}

func isProtoFile(path string) bool {
	return filepath.Ext(path) == ".proto"
}

func isYamlFile(path string) bool {
	base := filepath.Base(path)
	if base == "idl-ref.yml" || base == "idl-ref.yaml" {
		return false
	}
	return filepath.Ext(path) == ".yaml" || filepath.Ext(path) == ".yml"
}
