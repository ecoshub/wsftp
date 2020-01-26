package utils

import (
    "fmt"
    "os"
    "strings"
    "os/user"
    "net"
    rw "github.com/ecoshub/penman"
    "github.com/ecoshub/jint"
)

var Sep = string(os.PathSeparator)
// ./data/settings.json
var settingDir string = rw.GetHome() + Sep + "Documents" + Sep + "wsftp-settings.json"

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

func GetCustomUsername() string{
    if rw.IsFileExist(settingDir) {
        if !rw.IsFileEmpty(settingDir) {
            file := rw.Read(settingDir)
            username, done := jint.Get(file, "username")
            if done == nil{
                return string(username)
            }
        }
    }
    user, err := user.Current()
    if err != nil {
        return "unknown"
    }else{
        return user.Username
    }
}

func GetUsername() string{
    user, err := user.Current()
    if err != nil {
        return "unknown"
    }
    return user.Username
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

func GetEthMac() string {
    ins, _ := net.Interfaces()
    for _,in := range ins {
        if len(in.Name) > 0{
            if in.Name[0] == 'e' {
                return in.HardwareAddr.String()
            }
        }
    }
    for _,in := range ins {
        if in.HardwareAddr.String() != ""{
            return in.HardwareAddr.String()
        }
    }
    return "null"
}
func GetBroadcastIP() net.IP{
    IP := GetInterfaceIP()
    IP[len(IP) - 1] = 255
    return IP
}

func IsFileExist(file string) bool {
    if _, err := os.Stat(file); os.IsNotExist(err){
        return false
    }
    return true
}


func UniqName(dest, fileName string, filesize int64) string{
    if !IsFileExist(dest + Sep + fileName){
        return fileName
    }
    tokens := strings.Split(fileName, ".")
    name := tokens[0]
    ext := strings.Join(tokens[1:], ".")
    count := 1
    for {
        newName := fmt.Sprintf("%v(%v).%v", name, count, ext)
        if IsFileExist(dest + Sep + newName){
            if GetFileSize(dest + Sep + newName) < filesize {
                return newName
            }
        }else{
            return newName
        }
        count++
    }
}
