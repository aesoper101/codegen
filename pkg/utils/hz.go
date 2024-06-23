package utils

import (
	"path"

	"github.com/aesoper101/codegen/pkg/consts"
	"github.com/aesoper101/x/fileutil"
)

func IsHzNew(outputDir string) bool {
	return fileutil.IsExists(path.Join(outputDir, consts.HzFile))
}

func CreateHzFile(outputDir string) error {
	return fileutil.CreateFileFromString(path.Join(outputDir, consts.HzFile), false, "")
}
