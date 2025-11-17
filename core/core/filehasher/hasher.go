package filehasher

// Пакет для хеширования данных

import (
	"crypto/sha1"
	//"fmt"
	"io"
	"log"
	"math"
	"os"
	"runtime"
	"sync"
    "testfile/core/coretype"
)



type chunkJob struct{
    Index  int
    Offset int64
    Size   int
}



type FileHasher struct {

}



func (filehasher *FileHasher) HashFile(filePath string, pieceSize int) ([]coretype.Piece, error) {
    file, err := os.Open(filePath)
    if err != nil{
        log.Println("Can`t open file")
        //log.SetOutput(file)
        //file, err := os.OpenFile("app.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
        // TODO
        // НЕОБХОДИМО ПРОВЕРКИ НА ОШИБКИ И ЛОГИРОВАНИЕ ДОПИСАТЬ ПОЗЖЕ
        return nil, err

    }
    defer file.Close()


    stat,err := file.Stat()
    if err != nil{
        log.Println("Can`t take file.Stat()")
        //log.SetOutput(file)
        //file, err := os.OpenFile("app.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)

        // НЕОБХОДИМО ПРОВЕРКИ НА ОШИБКИ И ЛОГИРОВАНИЕ ДОПИСАТЬ ПОЗЖЕ
        return nil, err
    }
    fileSize := stat.Size()


    countPieces := int(math.Ceil(float64(fileSize) / float64(pieceSize)))
    workers := runtime.NumCPU()

    jobs := make(chan chunkJob, runtime.NumCPU()*2)    // Буферизованный канал
    results := make(chan coretype.Piece, countPieces)
    errors := make(chan error, 1) // канал буферезированный ровно на одну ошибку

    var wg sync.WaitGroup // Для синхронизации через закрытие канала 

    // Создаем пул воркеров
    for i := 0; i < workers; i++{
        wg.Add(1)
        go worker(file, jobs, results, errors, &wg)
    }

    // заполняем канал jobs
    go func ()  {
        defer close(jobs)
        for i := 0; i < countPieces; i++{
            offset := int64(i) * int64(pieceSize)
            size := pieceSize
            job := chunkJob{
                Index:  i,
                Offset: offset,
                Size:   size,
            }
            select{
            case jobs <- job:
            case err := <-errors:
                log.Println("error in gorutine with hash piece", err)
                //log.SetOutput(file)
                //file, err := os.OpenFile("app.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)

                // НЕОБХОДИМО ПРОВЕРКИ НА ОШИБКИ И ЛОГИРОВАНИЕ ДОПИСАТЬ ПОЗЖЕ
                return

            }
        }
    }()

    // Закрываем канал результатов
    go func(){
        wg.Wait()
        close(results)
    }()

    // Собираем ответ
    pieces := make([]coretype.Piece, countPieces)
    for piece := range results {
        pieces[piece.Index] = piece
        //log.Printf("Обработан чанк %d/%d, значение %v", piece.Index+1, countPieces, piece.Hash)
    }

    // Проверяем были ли ошибки в воркерах
    select {
    case err := <- errors:
        return nil, err
    default:
        return pieces, nil
    }
}



func worker(file *os.File, jobs <-chan chunkJob, results chan<- coretype.Piece, errors chan<- error, wg *sync.WaitGroup){
    defer wg.Done()

    for job := range jobs{
        data := make([]byte, job.Size)
        n,err := file.ReadAt(data, job.Offset)

        if err != nil && err != io.EOF {
            select {
            case errors <- err:
            default: 
            }
            return // если в канале уже есть ошибка то протсо останавливаем горутину
        }

        data = data[:n]
        hash := sha1.Sum(data)
        
        results <- coretype.Piece{
            Index: job.Index,
            Hash:  hash,
            Size:  n,
        }
    }
}

