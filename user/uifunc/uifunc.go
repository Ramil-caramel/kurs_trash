package uifunc

import (
	"path/filepath"
	"user/core/filehandler"
	"user/core/filehasher"
	"user/core/meta"
	"user/downloader"
	"user/logger"
	"user/seed"

	"bufio"
	"fmt"
	"os"
)

func CreateMetaFile(filePath string, trackerIP string, ph *filehandler.PublicHouse){

	pieceSize := 128*1024

	metaGen := &meta.MetaGenerator{Hasher: &filehasher.FileHasher{}}
	err := metaGen.GenerateMyTorrent(filePath, pieceSize, trackerIP)
	if err != nil {
		return 
	}

	ph.NewSeed(filePath, int64(pieceSize))

	fileName := filepath.Base(filePath)

	downloader.GetPeers(fileName, trackerIP)

}


func Seed(ph *filehandler.PublicHouse) {
	// Инициализируем сервер нашей базой
    seed.Init(ph)

    // Запускаем сервер
    go seed.SeedServer()

	fmt.Println("Server is running. Press ENTHER key to stop...")
    
    // Считываем один байт (любую клавишу)
    reader := bufio.NewReader(os.Stdin)
    reader.ReadByte()
    
    // Выходим из программы
    os.Exit(0)
}

func Download(metaFilePath string, ph *filehandler.PublicHouse){

	logger.Infof("start uifunc.downloader.NewDownloader(%s)", metaFilePath)

	d, err := downloader.NewDownloader(metaFilePath, ph)
	if err != nil {
		logger.Error("Can`t create downloader.NewDownloader")
		return
	}

	//fmt.Printf("Downloader initialized for file: %s (Size: %d bytes)\n", d.meta.FileName, d.meta.FileSize)
	//fmt.Printf("Number of pieces to download: %d\n", len(d.meta.Pieces))

	// Запуск процесса загрузки
	fmt.Println("Starting download...")
	err = d.DownloadAll()
	
	// Проверка результата
	if err != nil {
		logger.Errorf("downloader.NewDownloader have err = %v", err)
	}
	
	fmt.Println("🎉 Download completed successfully!")
}
/*
func GetPeers(fileName string, trackerURL string) ([]string, error) {
	conn, err := net.DialTimeout("tcp4", trackerURL+":4000", 5*time.Second)
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
	*/