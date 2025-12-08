package main

import (
	"encoding/binary"
	"encoding/json"
	"os"

	"fmt"
	"bufio"
	"strings"

	"io"
	"net"
	"user/netapi"

	//"path/filepath"

	"user/core/filehandler"
	"user/uifunc"
)


/* тест на получение пиров
	filePath := "/home/rama/Загрузки/avidreaders.ru__oblomov.txt"
	fileName := filepath.Base(filePath)
	a,_ := GetPeers(fileName, "192.168.1.2")
	fmt.Println(a)
*/

func main() {

	//filePath := "/home/rama/Загрузки/avidreaders.ru__oblomov.txt"
	//fileName := filepath.Base(filePath)

	// Инициализация PublicHouse (ph)
    ph := &filehandler.PublicHouse{}
	filehandler.EnsureFileExists()
    
/*
	filehandler.CreateFile(fileName, 1673619)
	for i := 0; i < 13; i++{
		post,_ := DownloadPiece(fileName, "127.0.0.1", i) 
		if post.Command =="ERR"{
			fmt.Println("err")
			continue
		}
		filehandler.PutPiece(fileName, post.Data,int64(i), 128*1024)
		
	}	
*/
    fmt.Println("--- Приложение BitTorrent Клиент/Сидер ---")

    reader := bufio.NewReader(os.Stdin)

    for {
        fmt.Println("\nВыберите действие:")
        fmt.Println("1. Создать мета-файл и начать сидировать (CreateMetaFile)")
        fmt.Println("2. Запустить сид-сервер (Seed)")
        fmt.Println("3. Скачать файл (Download)")
        fmt.Println("4. Выход")
        fmt.Print("Введите номер (1-4): ")

        input, _ := reader.ReadString('\n')
        input = strings.TrimSpace(input)

        switch input {
        case "1":
            handleCreateMetaFile(reader, ph)
        case "2":
            handleSeed(ph)
            // Seed обычно блокирует выполнение, если не вызывать его в отдельной горутине
            // Если Seed содержит os.Exit(0), как в вашем примере, это завершит программу.
            return 
        case "3":
            handleDownload(reader, ph)
        case "4":
            fmt.Println("👋 Выход из программы.")
            return
        default:
            fmt.Println("❌ Неверный ввод. Пожалуйста, введите число от 1 до 4.")
        }
    }

	//filePath := "/home/rama/Загрузки/avidreaders.ru__oblomov.txt"
	//fileName := filepath.Base(filePath)
	//pieceSize := 128 * 1024
	//trackerIP := "192.168.1.2"
	//filePath := "avidreaders.ru__oblomov.txt"
	//ph := &filehandler.PublicHouse{}
	//uifunc.CreateMetaFile(filePath, "192.168.1.2",ph)
	//uifunc.Seed(ph)
/*	
	metaGen := &meta.MetaGenerator{Hasher: &filehasher.FileHasher{}}
	err := metaGen.GenerateMyTorrent(filePath, pieceSize, trackerIP)
	if err != nil {
		return 
	}
		*//*
	pb := &filehandler.PublicHouse{}


    err := pb.NewSeed(filePath, 128*1024)
	if err != nil {
        fmt.Printf("Failed to add seed: %v", err)
		return
    }

*/

	/*
	data1, _ := os.ReadFile(filePath)
	data2, _ := os.ReadFile("123")
	

	if md5.Sum(data1) == md5.Sum(data2) {
		fmt.Println("✅ Файлы одинаковые")
	} else {
		fmt.Println("❌ Файлы разные")
	}
	//finalBitmap, actualPath, err := pb.VerifyTorrentFile(fileName + ".mytorrent")
	//fmt.Println(finalBitmap, actualPath ,err)
	*/

}	

// --- Обработчики Функций ---

func handleCreateMetaFile(reader *bufio.Reader, ph *filehandler.PublicHouse) {
    fmt.Print("Введите полный путь к файлу для сидирования (e.g., /path/to/file.txt): ")
    filePath, _ := reader.ReadString('\n')
    filePath = strings.TrimSpace(filePath)

    fmt.Print("Введите IP/адрес трекера (e.g., http://tracker.com:8080): ")
    trackerIP, _ := reader.ReadString('\n')
    trackerIP = strings.TrimSpace(trackerIP)

    if filePath != "" && trackerIP != "" {
        uifunc.CreateMetaFile(filePath, trackerIP, ph)
        fmt.Println("✅ Запрос на создание мета-файла и начало сидирования отправлен.")
    } else {
        fmt.Println("❌ Путь к файлу и IP трекера не могут быть пустыми.")
    }
}

func handleSeed(ph *filehandler.PublicHouse) {
    fmt.Println("🚀 Запуск сид-сервера...")
    // Примечание: Функция Seed в вашем примере содержит os.Exit(0)
    // внутри себя, что завершит всю программу после нажатия Enter.
    uifunc.Seed(ph) 
}

func handleDownload(reader *bufio.Reader, ph *filehandler.PublicHouse) {
    fmt.Print("Введите полный путь к мета-файлу (.meta) для скачивания (e.g., /path/to/file.meta): ")
    metaFilePath, _ := reader.ReadString('\n')
    metaFilePath = strings.TrimSpace(metaFilePath)

    if metaFilePath != "" {
        uifunc.Download(metaFilePath, ph)
    } else {
        fmt.Println("❌ Путь к мета-файлу не может быть пустым.")
    }
}





func DownloadPiece(fileName string, peer string, index int) (*netapi.PostPieceStruct, error) {


	conn, err := net.Dial("tcp4", peer+":3000")
	if err != nil {
		return  nil,err
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
		return nil,err
	}
	respLen := binary.BigEndian.Uint32(respLenBuf)

	respData := make([]byte, respLen)
	if _, err := io.ReadFull(conn, respData); err != nil {
		return nil,err
	}

	var postResp netapi.PostPieceStruct
	if err := json.Unmarshal(respData, &postResp); err != nil {
		return nil,err
	}
	return &postResp, nil
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
