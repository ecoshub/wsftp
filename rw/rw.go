package rw

import (
	"strings"
	rw "github.com/eco9999/rw"
)

var maindir string= FolderCreation("tunnel-logs")

func FolderCreation(mainFolderName string) string{
	downDir := rw.GetHome()
	downDir += rw.Sep()
	downDir += mainFolderName
	downDir += rw.Sep()
	if !rw.IsFileExist(downDir){
		rw.Mkdir(downDir)
	}
	return downDir
}

func NameControl(name string) bool{
	dirs := rw.Dir(maindir)
	for _,v := range dirs {
		if name == rw.SplitName(v){
			return true
		}
	}
	return false
}

func NameCreation(MAC, username string) string{
	return strings.ToLower(MAC + ":" + username + ".log")
}

func GetLog(MAC, username string) string{
	name := NameCreation(MAC, username)
	if NameControl(name) {
		file := rw.SRead(maindir + rw.Sep() + name)
		return file
	}
	return ""
}

func SaveLog(MAC, username, input string){
	name := NameCreation(MAC, username)
	rw.SWrite(maindir + rw.Sep() + name, input + rw.NewLine())
}