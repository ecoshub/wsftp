package utils

import (
    "fmt"
    "os"
    "io/ioutil"
    "unsafe"
    "strings"
    "os/user"
    "net"
    "time"
)

var Sep = string(os.PathSeparator)

func GetFileSize(dir string) int64{
    info, err := os.Stat(dir)
    if err != nil {
        return int64(0)
    }
    return info.Size()
}

func GetFileName(dir string) string{
    tokens := strings.Split(dir, Sep)
    name := tokens[len(tokens) - 1]
    return name
}

func GetFileExt(dir string) string{
    tokens := strings.Split(dir, ".")
    ext := strings.Join(tokens[1:], ".")
    return ext
}

func GetPackNumber(totalsize, speed int64) int{
    totalFrag := (totalsize / speed)
    if float64(totalFrag) < (float64(totalsize) / float64(speed)) {
        totalFrag++
    }
    return int(totalFrag)
}

func GetUsername() string{
    user, err := user.Current()
    if err != nil {
        return "unknown"
    }else{
        return user.Username
    }
}

func GetInterfaceIP() net.IP{
    ins, _ := net.Interfaces()
    inslen := len(ins)
    myAddr := ""
    for i := 0 ; i < inslen ; i++ {
        if ins[i].Flags &  net.FlagLoopback != net.FlagLoopback && ins[i].Flags & net.FlagUp == net.FlagUp{
            addr, _ := ins[i].Addrs()
            if addr != nil {
                for _,ad := range addr{
                    if strings.Contains(ad.String(), "."){
                        myAddr = ad.String()
                        break
                    }
                }
                ip, _, _ := net.ParseCIDR(myAddr)
                return ip
            }
        }
    }
    fmt.Println("Interface IP resolve error in func GetInterfaceIP()")
    return net.ParseIP("0.0.0.0")
}

func IntToByteArray(num int64, size int) []byte {
    // size := int(unsafe.Sizeof(num))
    arr := make([]byte, size)
    for i := 0 ; i < size ; i++ {
        byt := *(*uint8)(unsafe.Pointer(uintptr(unsafe.Pointer(&num)) + uintptr(i)))
        arr[i] = byt
    }
    return arr
}

func ByteArrayToInt(arr []byte) int64{
    val := int64(0)
    size := len(arr)
    for i := 0 ; i < size ; i++ {
        *(*uint8)(unsafe.Pointer(uintptr(unsafe.Pointer(&val)) + uintptr(i))) = arr[i]
    }
    return val
}

func FReadAt(dir string, offset int64, length int64) []byte {
    f, err := os.Open(dir)
    if err != nil {
        fmt.Println("File Open Error:", err)
    }else{
        defer f.Close()
    }
    data := make([]byte, length)
    _, err = f.Seek(offset, 0)
    if err != nil {
        fmt.Println("Seeker Error:", err)
    }
    _, err = f.Read(data)
    if err != nil {
        fmt.Println("Read Error:", err)
    }
    return data
}

// Write
// if file exist append end
// curr prefix not lower-upper key senstive
// dir: curr\new_folder\new_text.txt is current directory
// desk prefix not lower-upper key senstive
// dir: desk\new_folder\new_text.txt is desktop directory
func FWrite(dir string, buff []byte){
    newdir, newfile := SplitDir(dir)
    err := os.MkdirAll(newdir, os.ModePerm)
    if err != nil {
        fmt.Println("Make Directory Error:", err)
    }else{
        if IsFileExist(newfile) {
            appendFile(newfile, buff)
        }else{
            writeFile(newfile, buff)
        }
    }
}

// main write function
func writeFile(filedir string, buffer []byte) {
    err := ioutil.WriteFile(filedir, buffer, os.ModePerm)   
    if err != nil {
        fmt.Printf("File Write Error:%v\n", err)
    }
}

// main append function
func appendFile(filedir string, buff []byte) {
    f, err := os.OpenFile(filedir, os.O_APPEND|os.O_WRONLY, os.ModeAppend)
    if err != nil {
        fmt.Printf("File Open Error:%v\n", err)
    }
    defer f.Close()
    if _, err = f.Write(buff); err != nil {
        fmt.Printf("File Write Error:%v\n", err)
    }
}

func SplitDir(dir string) (string, string){
    sp := Sep
    tokens := strings.Split(dir, sp)
    dirPart := strings.Join(tokens[:len(tokens) - 1], sp)
    return dirPart, dir
}

func IsFileExist(file string) bool {
    if _, err := os.Stat(file); os.IsNotExist(err){
        return false
    }
    return true
}


func UniqName(dest, fileName string, filesize int64) string{
    if IsFileExist(dest + "/" + fileName){
        if GetFileSize(dest + "/" + fileName) < filesize {
            return fileName
        }
    }else{
        return fileName
    }
    tokens := strings.Split(fileName, ".")
    name := tokens[0]
    ext := strings.Join(tokens[1:], ".")
    count := 1
    for {
        newName := fmt.Sprintf("%v(%v).%v", name, count, ext)
        if IsFileExist(dest + "/" + newName){
            if GetFileSize(dest + "/" + newName) < filesize {
                return newName
            }
        }else{
            return newName
        }
        count++
    }
}

func TimeStamp() string {
    current := time.Now()
    times := current.Format("02-01-2006-15-04-05")
    return times
}