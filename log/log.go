package log

import (
	"fmt"
	"strings"
	rw "github.com/eco9999/penman"
	parse "github.com/eco9999/jparse"
)

var maindir string = FolderCreation("tunnel-logs")

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
	return strings.ToLower(username + ":" + content + ":" + MAC + ".log")
}

func GetLog(MAC, username, content string, start, end int) (string, int){
	name := nameCreation(MAC, username, content)
	lent := 0
	if nameControl(name) {
		file := rw.SRead(maindir + rw.Sep() + name)
		tokens := strings.Split(file, rw.NewLine())
		lent = len(tokens)
		if lent < end {
			return strings.Join(tokens, rw.NewLine()), lent
		}else{
			return strings.Join(tokens[start:end], rw.NewLine()), lent
		}	
	}
	return "", 0
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

func DelBase(MAC, username, content string) bool {
	name := nameCreation(MAC, username, content)
	res := rw.DelFile(maindir + rw.Sep() + name)
	return res
}

func DelLine(MAC, username, content, key , value string) bool{
	name := nameCreation(MAC, username, content)
	file := rw.SRead(maindir + rw.Sep() + name)
	tokens := strings.Split(file, rw.NewLine())
	lent := len(tokens)
	if lent < 1 {
		return false
	}
	count := -1
	for  i := 0 ; i < lent ; i ++ {
		line := tokens[i]
		lene := len(line)
		if lene > 0 {
			if line[lene - 1] == ',' {
				line = line[:lene - 1]
			}
			json, done := parse.JSONParser(line)
			if !done {fmt.Println("json parse error -log-");return false}
			input, done := json.Get(key)
			if done {
				if input == value {
					count = i
					break
				}
			}
		}
	}
	if count == -1 {
		fmt.Println("not found")
		return false
	}
	if count == lent - 1{
		tokens[count] = ""
		newFile := strings.Join(tokens, rw.Sep())
		lenf := len(newFile)
		rw.SWrite(maindir + rw.Sep() + name, newFile[:lenf - 1])
	}else{
		tokens = delFromList(tokens, count)
		newFile := strings.Join(tokens, rw.NewLine())
		rw.SOWrite(maindir + rw.Sep() + name, newFile)
	}
	return true
}

func delFromList(list []string,index int) []string{
	newList := make([]string, 0, len(list) - 1)
	for i := 0 ; i < len(list) ; i ++{
		if i != index {
			newList = append(newList, list[i])
		}
	}
	return newList
}