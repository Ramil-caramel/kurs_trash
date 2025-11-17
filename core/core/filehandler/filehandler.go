package filehandler

// Пакет для работы с файлами
// То есть создание, сборка и разборка

import (
	
	"io"
	"log"
	"math"
	"os"
	"errors"
)

var ErrPieceOutOfRange = errors.New("piece index out of file range")
//var ErrUnEqualPieceAndLen = errors.New("not equal piece len and data len") лишний


// Функция для аллоцирования пустого файла заданного размера
func CreateFile(name string, size int64) error {
	file, err := os.Create(name)
	if err != nil {
		log.Println("Error: Can`t create file")
		//TODO
		return err
	}

	defer file.Close()

	err = file.Truncate(size)
	if err != nil {
		log.Println("Error: Can`t trancute file")
		//TODO
		return err
	}

	return nil
}


// Функция примает имя файла, индекс куска и его размер
// после чего возвращает срез байт данных этого куска 
// данные получаются чтением с отступом pieceIndex * pieceSize
func GetPiece(name string, pieceIndex int64, pieceSize int64) ([]byte, error) {
	file, err := os.Open(name)
	if err != nil {
		log.Println("Error: Can`t read file")
		//TODO
		return nil, err
	}

	defer file.Close()

	stat,err := file.Stat()
	if err != nil {
		log.Println("Error: Can`t take file stat")
		//TODO
		return nil, err
	}
	fileSize := stat.Size()
	countPiece := int64(math.Ceil(float64(fileSize) / float64(pieceSize)))
	
	if pieceIndex >= countPiece{

		log.Println("Error: PieceIndex out of file range")
		//TODO
		return nil, io.EOF // написать кастомную ошибку а не io.EOF

	}

	offset := pieceIndex * pieceSize

	remaining := fileSize - offset // так как последний кусок не равен pieceSize, поэтому буфер выделяем так
	if remaining < pieceSize {
		pieceSize = remaining
	}

	data := make([]byte, pieceSize)//создали буфер для данных
	n, err := file.ReadAt(data, offset) //прочли со сдвигом
	
	if err != nil && err != io.EOF {
		
		return nil, err
	}

	data = data[:n] // обрезали до реально прочитанного
	return data, nil
}



// Функция осуществляет вставкку данных в файл
// вставка происходит с учетом отсупа pieceIndex * pieceSize
func PutPiece(name string, data []byte, pieceIndex int64, pieceSize int64) (error){

	file, err := os.OpenFile(name, os.O_RDWR,0644)
	if err != nil {
		log.Println("Error: Can`t read file")
		//TODO
		return err
	}
	defer file.Close()


	stat,err := file.Stat()
	if err != nil {
		log.Println("Error: Can`t take file stat")
		//TODO
		return err
	}
	fileSize := stat.Size()
	offset := pieceIndex * pieceSize
	countPiece := int64(math.Ceil(float64(fileSize) / float64(pieceSize)))
	
	if pieceIndex >= countPiece{
		log.Println("Error: PieceIndex out of file range")
		//TODO
		return ErrPieceOutOfRange // написать кастомную ошибку а не io.EOF

	}

	_, err = file.WriteAt(data, offset) // вставили со сдвигом
	
	if err != nil{
		
		return err
	}

	return nil
}