package filehasher

// Пакет для хеширования данных
// хеширование происходит паралельно 

import (
	"crypto/sha1"
	"io"
	"math"
	"os"
	"runtime"
	"sync"
    "user/core/coretype"

    "user/logger"
)


// струтура для формирования канала задач 
type chunkJob struct{
    Index  int
    Offset int64
    Size   int
}


// структура у которой есть метод хеширования
// так как уровень логики независим от реализации 
// нам необходима струтура которая будет соответсвовать описываемой в meta интерфейсу
type FileHasher struct {

}


// функция реализцкт хеширование файла и возвращает []coretype.Piece
func (filehasher *FileHasher) HashFile(filePath string, pieceSize int) ([]coretype.Piece, error) {

    logger.Infof("start filehasher.HashFile(%s, %d)", filePath, pieceSize)

    file, err := os.Open(filePath)
    if err != nil{
        logger.Errorf("filehasher.HashFile(...) have err = %v", err)
        return nil, err

    }
    defer file.Close()


    stat,err := file.Stat()
    if err != nil{
        logger.Errorf("filehasher.HashFile(...) have err = %v", err)
        return nil, err
    }
    fileSize := stat.Size()


    countPieces := int(math.Ceil(float64(fileSize) / float64(pieceSize)))
    workers := runtime.NumCPU()

    jobs := make(chan chunkJob, workers*2)    // Буферизованный канал
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
                logger.Errorf("filehasher.HashFile(...) Jobs have err = %v", err)
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
    }

    // Проверяем были ли ошибки в воркерах
    select {
    case err := <- errors:
        logger.Errorf("filehasher.HashFile(...) worker have err = %v", err)
        return nil, err
    default:
        return pieces, nil
    }
}


// worker для обработки задач из канала задач
func worker(file *os.File, jobs <-chan chunkJob, results chan<- coretype.Piece, errors chan<- error, wg *sync.WaitGroup){

    logger.Infof("start filehasher.worker(...)")

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

