package utils

import (
	"path"

	"github.com/aesoper101/codegen/pkg/consts"
	"github.com/aesoper101/x/filex"
)

func IsHzNew(outputDir string) bool {
	return filex.IsExists(path.Join(outputDir, consts.HzFile))
}

func CreateHzFile(outputDir string) error {
	return filex.CreateFileFromString(path.Join(outputDir, consts.HzFile), false, "")
}
