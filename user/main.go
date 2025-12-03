package main

import (
	//"time"
	//"os"
	"encoding/binary"
	"encoding/json"
	"fmt"

	"io"
	"net"
	"user/netapi"

	//"crypto/sha1"

	//"user/core/filehandler"
	"user/core/filehasher"
	"user/core/meta"
	//"user/seed"
	//"user/downloader"
)

func main() {

	filePath := "/home/rama/Загрузки/avidreaders.ru__oblomov.txt"
	//fileName := "avidreaders.ru__oblomov.txt.mytorrent"
	pieceSize := 128 * 1024
	trackerIP := "10.249.85.57"

	//ph := &filehandler.PublicHouse{}
	
	metaGen := &meta.MetaGenerator{Hasher: &filehasher.FileHasher{}}
	err := metaGen.GenerateMyTorrent(filePath, pieceSize, trackerIP)
	if err != nil {
		return 
	}
	//a,_ := GetPeers(fileName, "0.0.0.0")

	//pb :=&filehandler.PublicHouse{} 
	//pb.NewData(filePath, 128*1024)
	//seed.SeedServer() // у нас есть путь файл
	_,post := DownloadPiece(filePath, "127.0.0.1", 0) // мы знаем только имя файла но не его путь
	a :=string(post.Data)
	fmt.Println(a)

}	

func DownloadPiece(fileName string, peer string, index int) (error,*netapi.PostPieceStruct) {


	conn, err := net.Dial("tcp4", peer+":3000")
	if err != nil {
		return err, nil
	}
	defer conn.Close()

	req := netapi.CreateGetMessage(fileName, index)
	data, _ := json.Marshal(req)

	lenBuf := make([]byte, 4)
	binary.BigEndian.PutUint32(lenBuf, uint32(len(data)))
	conn.Write(lenBuf)
	conn.Write(data)

	// читаем POST ответ
	respLenBuf := make([]byte, 4)
	if _, err := io.ReadFull(conn, respLenBuf); err != nil {
		return err,nil
	}
	respLen := binary.BigEndian.Uint32(respLenBuf)

	respData := make([]byte, respLen)
	if _, err := io.ReadFull(conn, respData); err != nil {
		return err,nil
	}

	var postResp netapi.PostPieceStruct
	if err := json.Unmarshal(respData, &postResp); err != nil {
		return err,nil
	}
	return nil,&postResp
/*
	// проверка хэша
	h := sha1.Sum(postResp.Data)
	expected := d.piecesHash[index]
	if !equalBytes(h[:], expected) {
		return fmt.Errorf("hash mismatch on piece %d", index)
	}

	// сохраняем кусок
	if err := filehandler.PutPiece(d.meta.FileName, postResp.Data, int64(index), d.meta.PieceSize); err != nil {
		return err
	}

	// отмечаем в памяти
	d.mu.Lock()
	d.piecesHave[index] = true
	d.mu.Unlock()

	// обновляем bitmap в PublicHouse
	d.UpdateBitmap()

	return nil
*/	
}

func equalBytes(a, b []byte) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}


func GetPeers(fileName string, trackerURL string) ([]string, error) {
	conn, err := net.Dial("tcp4", trackerURL+":4000")
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	req := netapi.CreateGetPeersMessage(fileName)
	data, _ := json.Marshal(req)

	lenBuf := make([]byte, 4)
	binary.BigEndian.PutUint32(lenBuf, uint32(len(data)))
	conn.Write(lenBuf)
	conn.Write(data)

	// читаем ответ
	respLenBuf := make([]byte, 4)
	if _, err := io.ReadFull(conn, respLenBuf); err != nil {
		return nil, err
	}
	respLen := binary.BigEndian.Uint32(respLenBuf)

	respData := make([]byte, respLen)
	if _, err := io.ReadFull(conn, respData); err != nil {
		return nil, err
	}

	var resp netapi.PeersResponseStruct
	if err := json.Unmarshal(respData, &resp); err != nil {
		return nil, err
	}

	return resp.Peers, nil
}

/*
	// --- P2PClient структура ---
type P2PClient struct {
	ph        *filehandler.PublicHouse
	metaPath  string
	trackerIP string
}

// Генерация метафайла и уведомление трекера
func (c *P2PClient) GenerateMeta(filePath string, pieceSize int) error {
	metaGen := &meta.MetaGenerator{&filehasher.FileHasher{}}
	err := metaGen.GenerateMyTorrent(filePath, pieceSize, c.trackerIP)
	if err != nil {
		return err
	}

	// уведомляем трекер о наличии файла
	dl, err := downloader.NewDownloader(c.metaPath, c.ph)
	if err != nil {
		return err
	}

	_, err = dl.GetPeers() // фактически делает один запрос к трекеру
	return err
}

// Запуск скачивания
func (c *P2PClient) Download() error {
	dl, err := downloader.NewDownloader(c.metaPath, c.ph)
	if err != nil {
		return err
	}
	return dl.DownloadAll()
}

// Запуск раздачи
func (c *P2PClient) Seed() error {
	metaData, err := os.ReadFile(c.metaPath)
	if err != nil {
		return err
	}

	var meta downloader.MetaFile
	if err := json.Unmarshal(metaData, &meta); err != nil {
		return err
	}

	// запускаем SeedServer в отдельной горутине
	go seed.SeedServer([]string{meta.FileName})

	// можно держать программу живой, пока раздаём
	for {
		time.Sleep(10 * time.Second)
	}

	return nil
}
	*/
