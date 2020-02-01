package tools

import (
	"fmt"
	"github.com/ecoshub/penman"
	"os"
	"strings"
)

func GetFileSize(dir string) int64 {
	info, err := os.Stat(dir)
	if err != nil {
		return int64(0)
	}
	return info.Size()
}

func GetFileName(dir string) string {
	tokens := strings.Split(dir, SEPARATOR)
	name := tokens[len(tokens)-1]
	return name
}

func GetFileExt(dir string) string {
	tokens := strings.Split(dir, ".")
	ext := strings.Join(tokens[1:], ".")
	return ext
}

func UniqName(dest, fileName string, filesize int64) string {
	if !penman.IsFileExist(dest + SEPARATOR + fileName) {
		return fileName
	}
	tokens := strings.Split(fileName, ".")
	name := tokens[0]
	ext := strings.Join(tokens[1:], ".")
	count := 1
	for {
		newName := fmt.Sprintf("%v(%v).%v", name, count, ext)
		if penman.IsFileExist(dest + SEPARATOR + newName) {
			if GetFileSize(dest+SEPARATOR+newName) < filesize {
				return newName
			}
		} else {
			return newName
		}
		count++
	}
}

func GetPackNumber(totalsize, speed int64) int {
	totalFrag := (totalsize / speed)
	if float64(totalFrag) < (float64(totalsize) / float64(speed)) {
		totalFrag++
	}
	return int(totalFrag)
}
