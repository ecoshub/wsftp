package rw

import (
	"fmt"
	"os"
	"strings"	
	"runtime"
	"os/user"
	"path/filepath"
)

func DelFile(dir string) bool{
	_, filedir := SplitDir(dir)
	err := os.RemoveAll(filedir)
	if err != nil {
		fmt.Printf("File Delete Error %v\n:", err)
		return false
	}
	return true
}

func Rename(olddir string, newdir string){
	_, oldfiledir := SplitDir(olddir)
	_, newfiledir := SplitDir(newdir)
	err := os.Rename(oldfiledir, newfiledir)
	if err != nil {
		fmt.Printf("File Rename Error: %v", err)
	}
}

func CopyFile(olddir, newdir string) {
	temp := Read(olddir)
	OWrite(newdir, temp)
}

func MoveFile(olddir, newdir string) {
	temp := Read(olddir)
	OWrite(newdir, temp)
	DelFile(olddir)
}

func CopyDir(olddir, newdir string) {
	goon := true
	_, oldDIR := SplitDir(olddir)
	_, newDIR := SplitDir(newdir)
	tokens := strings.Split(newdir, Sep())
	last := tokens[len(tokens) - 1]
	if strings.Contains(last, ".") {
	    fmt.Println("New Directory is Not a Directory: ", newDIR)
	    goon = false
	}
	if !IsDir(oldDIR){
	    fmt.Println("Old Directory is Not a Directory: ", oldDIR)
	    goon = false
	}
	if goon {
		err := filepath.Walk(oldDIR + ".",
		    func(path string, info os.FileInfo, err error) error {
		    if err != nil {
		        return err
		    }
		    if !IsDir(path){
		        correctDir := strings.Replace(path, oldDIR, newDIR, -1)
		        CopyFile(path, correctDir)
		    }
		    return nil
		})
		if err != nil {
		    fmt.Println("File Copy Error:", err)
		}
	}
}

func Dir(dir string) []string{
	files := make([]string, 0 ,100)
	dir = PreProcess(dir + Sep())
	err := filepath.Walk(dir,
    func(path string, info os.FileInfo, err error) error {
    if err != nil {
        return err
    }
    files = append(files, strings.TrimPrefix(path, dir))
    return nil})
	if err != nil {
		fmt.Println("Error walking true path", err)
	}
    return files[1:]
}

func Mkdir(dir string){
	err := os.MkdirAll(PreProcess(dir), os.ModePerm)
	if err != nil {
		fmt.Println("Make Directory Error:", err)
	}
}

func MoveDir(olddir, newdir string) {
	goon := true
	_, oldDIR := SplitDir(olddir)
	_, newDIR := SplitDir(newdir)
	tokens := strings.Split(newdir, Sep())
	last := tokens[len(tokens) - 1]
	if strings.Contains(last, ".") {
	    fmt.Println("New Directory is Not a Directory: ", newDIR)
	    goon = false
	}
	if !IsDir(oldDIR){
	    fmt.Println("Old Directory is Not a Directory: ", oldDIR)
	    goon = false
	}
	if goon {
		err := filepath.Walk(oldDIR + ".",
		    func(path string, info os.FileInfo, err error) error {
		    if err != nil {
		        return err
		    }
		    if !IsDir(path){
		    	correctDir := strings.Replace(path, oldDIR, newDIR, -1)
		    	MoveFile(path, correctDir)
		    }
		    return nil
		})
		if err != nil {
		    fmt.Println("File Copy Error:", err)
		}
	}
	DelFile(olddir)
}

func Cd(dir string) string{
	tokens := strings.Split(dir , Sep())
	tokens = tokens[:len(tokens) - 1]
	newdir := strings.Join(tokens, Sep())
	return newdir
}

func GetCurrentDir() string {
	wd, _ := os.Getwd()
	return wd
}

func GetDesktop() string{
	myself, _ := user.Current()
	var deskdir string = myself.HomeDir
	deskdir = deskdir + Sep() + "Desktop"
	return deskdir
}

func GetDownloads() string{
	myself, _ := user.Current()
	var homedir string = myself.HomeDir
	homedir = homedir + Sep() + "Downloads"
	return homedir
}

func GetHome() string{
	myself, _ := user.Current()
	return myself.HomeDir
}

func Sep() string{
	return string(os.PathSeparator)
}

func NewLine() string {
	goos := runtime.GOOS
	if goos == "linux" || goos == "darwin"{
		return "\n"
	}
	return "\r\n"
}

func PreProcess(dir string) string {
	if dir == "" {
		return dir
	}
	tokens := strings.Split(dir, Sep())
	switch tokens[0] {
	case "desk":
		tokens[0] =  GetDesktop()
	case "curr":
		tokens[0] =  GetCurrentDir()
	case "down":
		tokens[0] =  GetDownloads()
	}
	dir = strings.Join(tokens, Sep())
	return dir
}

func SplitDir(dir string) (string, string){
	dir = PreProcess(dir)
	tokens := strings.Split(dir, Sep())
	dirPart := strings.Join(tokens[:len(tokens) - 1], Sep())
	return dirPart, dir
}

func SplitName(dir string) string{
	sp := Sep()
	tokens := strings.Split(dir, sp)
	namePart := tokens[len(tokens) - 1]
	return namePart
}

func IsLinux() bool{
	return runtime.GOOS == "linux"
}

func IsFileEmpty(dir string) bool{
	dir = PreProcess(dir)
	file, err := os.Stat(dir);
	if err != nil {
		fmt.Println("File stat error: ", err)
	}
	return file.Size() == 0
}

func IsDir(dir string) bool{
    fi, err := os.Stat(dir)
    if err != nil {
    	return false
    }
    if fi.Mode().IsDir() {
    	return true
    }
    return false
}

func IsFileExist(file string) bool {
	if _, err := os.Stat(file); os.IsNotExist(err){
		return false
	}
	return true
}

func GetFileSize(dir string) int64{
    info, err := os.Stat(dir)
    if err != nil {
        return int64(0)
    }
    return info.Size()
}

func GetFileName(dir string) string{
    tokens := strings.Split(dir, Sep())
    name := tokens[len(tokens) - 1]
    return name
}

func GetFileExt(dir string) string{
    tokens := strings.Split(dir, ".")
    ext := strings.Join(tokens[1:], ".")
    return ext
}
