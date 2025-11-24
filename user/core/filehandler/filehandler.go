package filehandler

// Пакет для работы с файлами
// То есть создание, сборка и разборка

import (
	"io"
	"math"
	"os"
	"errors"


	"user/logger"
)

var ErrPieceOutOfRange = errors.New("piece index out of file range")
//var ErrUnEqualPieceAndLen = errors.New("not equal piece len and data len") лишний


// Функция для аллоцирования пустого файла заданного размера
func CreateFile(name string, size int64) error {

	logger.Infof("start filehandler.CreateFile(%v,...)", name)

	file, err := os.Create(name)
	if err != nil {
		logger.Errorf("ilehandler.CreateFile(...) have err = %v", err)
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

