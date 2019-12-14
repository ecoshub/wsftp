package rw

import (
	"strings"
	rw "github.com/eco9999/penman"
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

func nameControl(name string) bool{
	dirs := rw.Dir(maindir)
	for _,v := range dirs {
		if name == rw.SplitName(v){
			return true
		}
	}
	return false
}

func nameCreation(MAC, username, content string) string{
	return strings.ToLower(MAC + ":" + username + "-" + content + ".log")
}

func GetLog(MAC, username, content string, start, end int) string{
	name := nameCreation(MAC, username, content)
	if nameControl(name) {
		file := rw.SRead(maindir + rw.Sep() + name)
		tokens := strings.Split(file, rw.NewLine())
		lent := len(tokens)
		if lent < end {
			return strings.Join(tokens, rw.NewLine())
		}else{
			return strings.Join(tokens[start:end], rw.NewLine())
		}	
	}
	return ""
}

func SaveLog(MAC, username, content, input string){
	name := nameCreation(MAC, username, content)
	size := rw.GetFileSize(maindir + rw.Sep() + name)
	if size == 0 {
		rw.SWrite(maindir + rw.Sep() + name, input)
	}else{
		rw.SWrite(maindir + rw.Sep() + name, "," + rw.NewLine() + input)
	}
}