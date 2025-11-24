package main

import (
	/*
		"crypto/sha1"
		"encoding/base64"
		"fmt"
	*/
	"math"
	"os"

	"user/core/filehandler"
	"user/core/filehasher"
	"user/core/meta"
)

func main() {

	filepath := "/home/rama/Загрузки/avidreaders.ru__oblomov.txt"
	/*
		f := file.FileHasher{}
		pieces1,_ := f.HashFile(filepath, 256*1024)
		pieces2 := meta.ConvertPiecesToBase64(pieces1)
		for i:= range pieces1{
			fmt.Println(pieces1[i].Hash)
			fmt.Println(pieces2[i])
		}
	*/

	jsonGenerator := meta.MetaGenerator{
		Hasher: &filehasher.FileHasher{},
	}

	jsonGenerator.GenerateMyTorrent(filepath, 128*1024, "0.0")

	/*
	data1,_ := filehandler.GetPiece(filepath, 12, 128*1024)
	q:=sha1.Sum(data1)
	f := base64.StdEncoding.EncodeToString(q[:])
	fmt.Println(f)
	*/

	filehandler.CreateFile("avidreaders.ru__oblomov.txt", 1673619)

	file, err := os.OpenFile(filepath, os.O_RDWR, 0644)
	if err != nil {
		return
	}
	defer file.Close()
	fileStat, _ := file.Stat()

	countPiece := int64(math.Ceil(float64(fileStat.Size()) / float64(128*1024)))
	for i := 0; i < int(countPiece); i++ {
		data, _ := filehandler.GetPiece(filepath, int64(i), 128*1024)

		filehandler.PutPiece("avidreaders.ru__oblomov.txt", data, int64(i), 128*1024)
	}
}
