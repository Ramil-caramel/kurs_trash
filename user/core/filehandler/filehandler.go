package filehandler

// Пакет для работы с файлами
// То есть создание, сборка и разборка
// А также для взаимодействия с кастомными структурами данных

import (
	"bufio"
	"crypto/sha1"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"user/logger"
)

var ErrPieceOutOfRange = errors.New("piece index out of file range")
var ErrUnexistPiece = errors.New("piece index out of range")
var ErrUnFoundRecord = errors.New("record not found")
//var ErrUnEqualPieceAndLen = errors.New("not equal piece len and data len") лишний


// Функция для аллоцирования пустого файла заданного размера в месте работы программы
func CreateFile(name string, size int64) error {

	logger.Infof("start filehandler.CreateFile(%v,...)", name)

	file, err := os.Create(name)
	if err != nil {
		logger.Errorf("filehandler.CreateFile(...) have err = %v", err)
		return err
	}

	defer file.Close()

	err = file.Truncate(size)
	if err != nil {
		logger.Errorf("ilehandler.CreateFile(...) have err = %v", err)
		return err
	}

	return nil
}


// Функция примает имя файла, индекс куска и его размер
// после чего возвращает срез байт данных этого куска 
// данные получаются чтением с отступом pieceIndex * pieceSize
// подсчет индекса начинается с 0
func GetPiece(name string, pieceIndex int64, pieceSize int64) ([]byte, error) {

	logger.Infof("start filehandler.GetPiece(%v,%d,%d)", name,pieceIndex, pieceSize)

	file, err := os.Open(name)
	if err != nil {
		logger.Errorf("filehandler.GetPiece(...) have err = %v", err)
		return nil, err
	}

	defer file.Close()

	stat,err := file.Stat()
	if err != nil {
		logger.Errorf("filehandler.GetPiece(...) have err = %v", err)
		return nil, err
	}
	fileSize := stat.Size()
	countPiece := int64(math.Ceil(float64(fileSize) / float64(pieceSize)))
	
	if pieceIndex >= countPiece{

		logger.Errorf("filehandler.GetPiece(...) have err = %v", ErrPieceOutOfRange)

		return nil, ErrPieceOutOfRange 

	}

	offset := pieceIndex * pieceSize

	remaining := fileSize - offset // так как последний кусок не равен pieceSize, поэтому буфер выделяем так
	if remaining < pieceSize {
		pieceSize = remaining
	}

	data := make([]byte, pieceSize)//создали буфер для данных
	n, err := file.ReadAt(data, offset) //прочли со сдвигом
	
	if err != nil && err != io.EOF {
		logger.Errorf("filehandler.GetPiece(...) have err = %v", err)
		return nil, err
	}

	data = data[:n] // обрезали до реально прочитанного
	return data, nil
}



// Функция осуществляет вставкку данных в файл
// вставка происходит с учетом отсупа pieceIndex * pieceSize
// подсчет индекса начинается с 0
func PutPiece(name string, data []byte, pieceIndex int64, pieceSize int64) (error){

	logger.Infof("start filehandler.PutPiece(%v,...,%d,..)", name,pieceIndex)

	file, err := os.OpenFile(name, os.O_RDWR,0644)
	if err != nil {
		logger.Errorf("filehandler.PutPiece(...) have err = %v", err)
		return err
	}
	defer file.Close()


	stat,err := file.Stat()
	if err != nil {
		logger.Errorf("filehandler.PutPiece(...) have err = %v", err)
		return err
	}
	fileSize := stat.Size()
	offset := pieceIndex * pieceSize
	countPiece := int64(math.Ceil(float64(fileSize) / float64(pieceSize)))
	
	if pieceIndex >= countPiece{
		logger.Errorf("filehandler.PutPiece(...) have err = %v", ErrPieceOutOfRange)
		return ErrPieceOutOfRange 

	}

	_, err = file.WriteAt(data, offset) // вставили со сдвигом
	
	if err != nil{
		logger.Errorf("filehandler.PutPiece(...) have err = %v", err)
		return err
	}

	return nil
}


const PublicHouseFile = "PublicHouse.txt"

// Структура расчитана на работу и управлением файла содержащего список раздач и скаченных файлов 
// файл содеражит наборы <full_path>|<file_name>|<bitmap>
type PublicHouse struct{
    mu sync.RWMutex
}

func EnsureFileExists() error {
    
    logger.Info("start filehandlerEnsureFileExists()")
    filePath := filepath.Join(".", PublicHouseFile) 

    _, err := os.Stat(filePath)
    if os.IsNotExist(err) {
        
        f, createErr := os.Create(filePath)
        if createErr != nil {
            logger.Errorf("filehandlerEnsureFileExists(...) have err = %v", err)
            return createErr
        }
        f.Close()
        logger.Info("create filehandlerEnsureFileExists()")
    } else if err != nil {
        logger.Errorf("filehandlerEnsureFileExists(...) have err = %v", err)
        return err
    }
    
    return nil
}
// функция в файл содеражащий всесь список наших раздач 
// добавляет новую раздачу то есть набор data.txt 111111
func (ph *PublicHouse) NewSeed(fullPath string, pieceSize int64) error {

    logger.Infof("start filehandler.PublicHouse.NewSeed(%v,...)", fullPath)

    ph.mu.Lock()
    defer ph.mu.Unlock()

    stat, err := os.Stat(fullPath)
    if err != nil {
        logger.Errorf("filehandler.PublicHouse.NewSeed(...) have err = %v", err)
        return err
    }

    fileName := filepath.Base(fullPath)
    fileSize := stat.Size()
    // считаем размер конкретного торрента
    countPiece := int(math.Ceil(float64(fileSize) / float64(pieceSize)))
    //fmt.Println(countPiece, fileSize, pieceSize,math.Ceil(float64(fileSize) / float64(pieceSize)))
    bitmap := strings.Repeat("1", countPiece)

    newLine := encodeRecord(fullPath, fileName, bitmap)

    // открываем единый файл
    file, err := os.OpenFile(PublicHouseFile, os.O_RDWR|os.O_CREATE, 0644)
    if err != nil {
        logger.Errorf("filehandler.PublicHouse.NewSeed(...) have err = %v", err)
        return err
    }
    defer file.Close()


    // читаем весь файл в память
    scanner := bufio.NewScanner(file)
    var lines []string
    found := false

    for scanner.Scan() {
        line := scanner.Text()
        p, n, _, err := decodeRecord(line)
        if err == nil && (p == fullPath || n == fileName) {
            lines = append(lines, newLine)
            found = true
        } else {
            lines = append(lines, line+"\n")
        }
    }

    if err := scanner.Err(); err != nil {
        logger.Errorf("filehandler.PublicHouse.NewSeed(...) have err = %v", err)
        return err
    }

    // если не нашли — добавляем как новую запись
    if !found {
        lines = append(lines, newLine)
    }

    // перезаписываем файл
    err = os.WriteFile(PublicHouseFile, []byte(strings.Join(lines, "")), 0644)
    if err != nil {
        logger.Errorf("filehandler.PublicHouse.NewSeed(...) have err = %v", err)
        return err
    }

    return nil
}

// Возвращает BitMap по любому из полей буть то абсолютный путь или название файла
func (ph *PublicHouse) GetBitmap(fileId string) (string, error) {

    logger.Infof("start filehandler.PublicHouse.GetBitmap(%v,...)", fileId)

    file, err := os.Open(PublicHouseFile)
    if err != nil {
        logger.Errorf("filehandler.PublicHouse.GetBitmap(...) have err = %v", err)
        return "", err
    }
    defer file.Close()

    scanner := bufio.NewScanner(file)

    for scanner.Scan() {
        line := scanner.Text()
        p, n, b, err := decodeRecord(line)
        if err != nil {
            continue
        }
        if fileId == p || fileId == n {
            return b, nil
        }
    }

    if err := scanner.Err(); err != nil {
        logger.Errorf("filehandler.PublicHouse.GetBitmap(...) have err = %v", err)
        return "", err
    }

    logger.Errorf("ilehandler.PublicHouse.GetBitmap(...) have err = %v", ErrUnFoundRecord)
    return "", ErrUnFoundRecord
}



// HasRecord — проверяет, существует ли запись в файле 
func (ph *PublicHouse) HasRecord(fileId string) (bool, error) {

    logger.Infof("start filehandler.PublicHouse.HasRecord(%v,...)", fileId)

    file, err := os.Open(PublicHouseFile)
    if err != nil {
        if os.IsNotExist(err) {
            logger.Errorf("filehandler.PublicHouse.HasRecord(...) have err = %v", err)
            return false, err
        }
        logger.Errorf("filehandler.PublicHouse.HasRecord(...) have err = %v", err)
        return false, err
    }
    defer file.Close()

    scanner := bufio.NewScanner(file)

    for scanner.Scan() {
        line := scanner.Text()
        p, n, _, err := decodeRecord(line)
        if err != nil {
            continue
        }
        if fileId == p || fileId == n {
            return true, nil
        }
    }

    if err := scanner.Err(); err != nil {
        logger.Errorf("ilehandler.PublicHouse.HasRecord(...) have err = %v", err)
        return false, err
    }

    logger.Error("ilehandler.PublicHouse.HasRecord(...) have err = not found Record")
    return false, nil
}



// HasPiece — проверяет, существует ли кусок index у файла (бит = '1')
func (ph *PublicHouse) HasPiece(fileId string, index int) (bool, error) {

    logger.Infof("start filehandler.PublicHouse.HasPiece(%v,%d)", fileId, index)

    bitmap, err := ph.GetBitmap(fileId)
    if err != nil {
        logger.Errorf("filehandler.PublicHouse.HasPiece(...) have err = %v", err)
        return false, err
    }

    if index < 0 || index >= len(bitmap) {
        logger.Errorf("filehandler.PublicHouse.HasPiece(...) have err = %v", err)

        return false, ErrUnexistPiece
    }

    return bitmap[index] == '1', nil
}

// устанавливает 1 на место скачанного куска 
func (ph *PublicHouse) SetPiece(fileId string, index int) error {
    ph.mu.Lock()

    logger.Infof("start filehandler.PublicHouse.SetPiece(%v,%d)", fileId, index)

    bitmap, err := ph.GetBitmap(fileId)
    if err != nil {
        logger.Errorf("filehandler.PublicHouse.SetPiece(...) have err = %v", err)
    }

    if index < 0 || index >= len(bitmap) {
        logger.Errorf("filehandler.PublicHouse.SetPiece(...) have err = %v", ErrUnexistPiece)
        return ErrUnexistPiece
    }

    // Если уже 1 — нечего менять
    if bitmap[index] == '1' {
        ph.mu.Unlock()
        return nil
    }

    // Изменяем требуемый бит
    newBitmap := []byte(bitmap)
    newBitmap[index] = '1'

    ph.mu.Unlock()

    return ph.UpdateRecord(fileId, string(newBitmap))
}


// меняет битовую карту на заданную, при условии что запись существует
func (ph *PublicHouse) UpdateRecord(fileId, newBitmap string) error {

    logger.Infof("start filehandler.PublicHouse.updateRecord(%v,%v)", fileId, newBitmap)

    ph.mu.Lock()
    defer ph.mu.Unlock()

    file, err := os.OpenFile(PublicHouseFile, os.O_CREATE|os.O_RDWR, 0644)
    if err != nil {
        logger.Errorf("filehandler.PublicHouse.updateRecord(...) have err = %v", err)
        return err
    }
    defer file.Close()

    scanner := bufio.NewScanner(file)
    var lines []string
    updated := false

    for scanner.Scan() {
        line := scanner.Text()
        p, n, _, err := decodeRecord(line)
        if err == nil && (fileId == p || fileId == n) {
            lines = append(lines, encodeRecord(p, n, newBitmap))
            updated = true
        } else {
            lines = append(lines, line+"\n")
        }
    }

    if err := scanner.Err(); err != nil {
        logger.Errorf("filehandler.PublicHouse.updateRecord(...) have err = %v", err)
        return err
    }

    if !updated {
        logger.Infof("filehandler.PublicHouse.updateRecord(...) have trouble = %v", ErrUnFoundRecord)
        lines = append(lines, encodeRecord(fileId, filepath.Base(fileId), newBitmap))
    }


    return os.WriteFile(PublicHouseFile, []byte(strings.Join(lines, "")), 0644)
}

func (ph *PublicHouse) GetFullPathByName(fileName string) (string, error) {

    logger.Infof("start filehandler.PublicHouse.GetFullPathByName(%s)", fileName)

    file, err := os.Open(PublicHouseFile)
    if err != nil {
        logger.Errorf("filehandler.PublicHouse.GetFullPathByName(...) have err = %v", err)
        return "", err
    }
    defer file.Close()

    scanner := bufio.NewScanner(file)

    for scanner.Scan() {
        line := scanner.Text()
        full, name, _, err := decodeRecord(line)
        if err != nil {
            continue // пропускаем битые строки
        }

        if name == fileName {
            return full, nil
        }
    }

    if err := scanner.Err(); err != nil {
        logger.Errorf("filehandler.PublicHouse.GetFullPathByName(...) have err = %v", err)
        return "", err
    }

    return "", ErrUnFoundRecord
}

// Формирует строку для запси формата <full_path>|<file_name>|<bitmap>
func encodeRecord(path, name, bitmap string) string {
    return fmt.Sprintf("%s|%s|%s\n", path, name, bitmap)
}

// Декодирует строку формата <full_path>|<file_name>|<bitmap> в набор переменных
func decodeRecord(line string) (path, name, bitmap string, err error) {
    parts := strings.Split(line, "|")
    if len(parts) != 3 {
        return "", "", "", errors.New("invalid record format")
    }
    return parts[0], parts[1], parts[2], nil
}


type TorrentMeta struct {
    FileName   string    `json:"file_name"`
    FileSize   int64     `json:"file_size"`
    PieceSize  int       `json:"piece_size"`
    Pieces     []string  `json:"pieces"` // base64(SHA1)
    TrackerURL string    `json:"tracker_url,omitempty"`
}


// VerifyTorrentFile принимает путь к .mytorrent файлу, проверяет базу и целостность данных.
// Возвращает актуальный BitMap (строку из 0 и 1), реальный путь к файлу и ошибку.
func (ph *PublicHouse) VerifyTorrentFile(metaPath string) (string, string, error) {
    logger.Infof("start verification for meta: %s", metaPath)

    // 1. Читаем и парсим мета-файл
    metaContent, err := os.ReadFile(metaPath)
    if err != nil {
        return "", "", fmt.Errorf("failed to read meta file: %w", err)
    }

    var meta TorrentMeta
    if err := json.Unmarshal(metaContent, &meta); err != nil {
        return "", "", fmt.Errorf("failed to parse meta json: %w", err)
    }

    // 2. Проверяем наличие в базе PublicHouse
    // Пытаемся найти полный путь по имени файла
    dbFullPath, err := ph.GetFullPathByName(meta.FileName)
    recordExists := err == nil

    // 3. Определяем фактическое расположение файла
    var actualPath string

    // Сценарий А: Файл есть по пути, указанному в базе (мы - создатель/сидер)
    if recordExists {
        if _, err := os.Stat(dbFullPath); err == nil {
            actualPath = dbFullPath
        }
    }

    // Сценарий Б: Файла нет по пути из базы (или записи нет), проверяем локальную папку (мы - личер)
    if actualPath == "" {
        localPath := meta.FileName // или filepath.Join(".", meta.FileName)
        if _, err := os.Stat(localPath); err == nil {
            actualPath = localPath
        }
    }

    // Если файл нигде не найден
    if actualPath == "" {
        // Если файла нет физически, мы не можем проверить хеши. 
        // Возвращаем пустую карту (все нули).
        emptyBitmap := strings.Repeat("0", len(meta.Pieces))
        return emptyBitmap, "", errors.New("data file not found physically")
    }

    // 4. Проверяем хеши кусков
    file, err := os.Open(actualPath)
    if err != nil {
        return "", "", err
    }
    defer file.Close()

    verifiedBitmap := make([]byte, len(meta.Pieces))
    buf := make([]byte, meta.PieceSize)

    for i, expectedHash := range meta.Pieces {
        offset := int64(i) * int64(meta.PieceSize)
        
        // Читаем кусок
        n, err := file.ReadAt(buf, offset)
        if err != nil && err != io.EOF {
            logger.Errorf("error reading piece %d: %v", i, err)
            verifiedBitmap[i] = '0'
            continue
        }
        
        // Если прочитали 0 байт (конец файла раньше времени), то куска нет
        if n == 0 {
            verifiedBitmap[i] = '0'
            continue
        }

        // Считаем SHA1
        hash := sha1.Sum(buf[:n])
        // Кодируем в Base64 (так как в структуре TorrentMeta указано base64)
        calculatedHash := base64.StdEncoding.EncodeToString(hash[:])

        if calculatedHash == expectedHash {
            fmt.Println(calculatedHash,expectedHash)
            verifiedBitmap[i] = '1'
        } else {
            verifiedBitmap[i] = '0'
        }
    }

    finalBitmap := string(verifiedBitmap)

    // Если запись была в базе, но битовая карта отличается — имеет смысл обновить базу
    if recordExists {
        oldBitmap, _ := ph.GetBitmap(meta.FileName)
        if oldBitmap != finalBitmap {
            logger.Infof("Bitmap mismatch for %s. Updating DB...", meta.FileName)
            _ = ph.UpdateRecord(meta.FileName, finalBitmap)
        }
    }

    return finalBitmap, actualPath, nil
}