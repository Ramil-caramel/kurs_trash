package meta

// Пакет для создания мета файла

import (
	"encoding/base64"
	"encoding/json"
	"os"
	"testfile/core/coretype"
)

// Структура для формирования JSON, то есть .mytorrent
type TorrentMeta struct {
    FileName   string    `json:"file_name"`
    FileSize   int64     `json:"file_size"`
    PieceSize  int       `json:"piece_size"`
    Pieces     []string  `json:"pieces"` // base64(SHA1)
    TrackerURL string    `json:"tracker_url,omitempty"`
}


// Интерфейс хешера, 
// то есть любая структура которая реализует хешер
// должна разделять фалйы на куски, получать хеш каждого куска и возвращать их в форме slice []coretype.Piece
type Hasher interface{
	HashFile(filePath string, pieceSize int) ([]coretype.Piece, error)
}

// Структура реализующая создание .mytorrent
type MetaGenerator struct {
    Hasher Hasher
}

func (metaGenerator *MetaGenerator) GenerateMyTorrent(filePath string, pieceSize int, trackerIP string) (error){
    pieces, err := metaGenerator.Hasher.HashFile(filePath, pieceSize)
    if err != nil {
        return err
    }
    fileInfo, _ := os.Stat(filePath)

    meta := TorrentMeta{
        FileName:  fileInfo.Name(),
        FileSize:  fileInfo.Size(),
        PieceSize: pieceSize,
        Pieces:    ConvertPiecesToBase64(pieces),
        TrackerURL: trackerIP,
    }

    data,err := json.MarshalIndent(meta, "", "  ")
    if err != nil {
        return err
    }
    err = os.WriteFile(fileInfo.Name() + ".mytorrent", data, 0644)
    
    if err != nil {
        return err
    }

	return nil
}

func ConvertPiecesToBase64(pieces []coretype.Piece) ([]string){

    if len(pieces) == 0 {
        return nil
    }

    strings := make([]string, len(pieces))

    for i ,val:= range pieces{
        strings[i] = base64.StdEncoding.EncodeToString(val.Hash[:])
    }
    return strings
}

