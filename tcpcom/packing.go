package tcpcom

import (
	"fmt"
	utils "wsftp/utils"
)


func PackHeader(dir , dest string) []byte{
	// 1 byte for filename size
	// n byte filename
	// 1 byte for filedest size
	// m byte filedest
	// 8 byte for file size
	// 8 byte for checksum
	fileSize := utils.GetFileSize(dir)

	username := utils.GetUsername()
	userlen := len(username)
	
	name := utils.GetFileName(dir)
	namelen := len(name)

	destlen := len(dest)

	userSizeB:= byte(userlen)
	userB := []byte(username)
	nameSizeB := byte(namelen)
	nameB := []byte(name)
	destSizeB := byte(destlen)
	destB := []byte(dest)
	sizeB := utils.IntToByteArray(fileSize, 8)

	totalSize := 1 + userlen + 1 + namelen + 1 + destlen + 8 + 8

	checkSum := utils.IntToByteArray(int64(totalSize), 8)

	header := make([]byte,0, totalSize)

	header = append(header, userSizeB)
	header = append(header, userB...)

	header = append(header, nameSizeB)
	header = append(header, nameB...)

	header = append(header, destSizeB)
	header = append(header, destB...)
	header = append(header, sizeB...)
	header = append(header, checkSum...)

	return header
}

func UnpackHeader(header []byte)(username, dest, fileName string, fileSize int64){

    userSize := int(header[0:1][0])
    username = string(header[1:1 + userSize])

    nameSize := int(header[1 + userSize:1 + userSize + 1][0])
    fileName = string(header[1 + userSize + 1:1 + userSize + 1 + nameSize])

    destlen := int(header[1 + userSize + 1 + nameSize:1 + userSize + 1 + nameSize + 1][0])
    dest = string(header[1 + userSize + 1 + nameSize + 1:1 + userSize + 1 + nameSize + 1 + destlen])

    fileSize = utils.ByteArrayToInt(header[1 + userSize + 1 + nameSize + 1 + destlen:1 + userSize + 1 + nameSize + 1 + destlen + 8])
    checkSum := utils.ByteArrayToInt(header[1 + userSize + 1 + nameSize + 1 + destlen + 8:1 + userSize + 1 + nameSize + 1 + destlen + 8 + 8])

    if checkSum != int64(len(header)) {
        fmt.Println("broken fragment")
    }
    return
}