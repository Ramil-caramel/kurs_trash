package uifunc

import (
	"path/filepath"
	"user/core/filehandler"
	"user/core/filehasher"
	"user/core/meta"
	"user/downloader"
	"user/logger"
	"user/seed"


	"fmt"

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


func StartSeederBackground(ph *filehandler.PublicHouse) {
	// Инициализируем сервер нашей базой
    seed.Init(ph)

    // Запускаем сервер
    go seed.SeedServer()
}

func StopSeeder() {
    if seed.Listener != nil { // Проверяем, что listener был создан
        seed.Listener.Close() // Закрытие listener разблокирует ln.Accept() в горутине
        fmt.Println("🛑 Сид-сервер остановлен.")
    }
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
