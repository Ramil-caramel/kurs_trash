package downloader
/*
import (
	"encoding/base64"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"os"
	"sync"
	"crypto/sha1"
	
	"user/core/filehandler"
	"user/netapi"
)

// --- структура мета-файла ---
type MetaFile struct {
	FileName   string   `json:"file_name"`
	FileSize   int64    `json:"file_size"`
	PieceSize  int64    `json:"piece_size"`
	Pieces     []string `json:"pieces"` // base64 SHA1
	TrackerURL string   `json:"tracker_url"`
}

// основной объект Downloader
type Downloader struct {
	meta       MetaFile
	ph         *filehandler.PublicHouse
	piecesHash [][]byte
	piecesHave []bool
	mu         sync.Mutex
}

// создание Downloader
func NewDownloader(metaFilePath string, ph *filehandler.PublicHouse) (*Downloader, error) {
	// --- загрузка мета-файла ---
	data, err := os.ReadFile(metaFilePath)
	if err != nil {
		return nil, err
	}
	var meta MetaFile
	if err := json.Unmarshal(data, &meta); err != nil {
		return nil, err
	}

	// --- декодирование SHA1 ---
	piecesHash := make([][]byte, len(meta.Pieces))
	for i, s := range meta.Pieces {
		h, err := base64.StdEncoding.DecodeString(s)
		if err != nil {
			return nil, err
		}
		piecesHash[i] = h
	}

	// --- пробуем прочитать битмап в PublicHouse ---
	bitmap, exists := ph.GetRecord(meta.FileName)

	// ожидаемый битмап
	emptyBitmap := make([]byte, len(meta.Pieces))
	for i := range emptyBitmap {
		emptyBitmap[i] = '0'
	}

	// Если записи нет — создаём нулевую
	if !exists {
		bitmap = string(emptyBitmap)
		if err := ph.UpdateRecord(meta.FileName, bitmap); err != nil {
			return nil, err
		}
	}

	// --- проверка длины битмапа ---
	if len(bitmap) != len(meta.Pieces) {
		// битмап повреждён, пересоздаём
		bitmap = string(emptyBitmap)
		ph.UpdateRecord(meta.FileName, bitmap)
	}

	// --- восстанавливаем piecesHave из битмапа ---
	piecesHave := make([]bool, len(meta.Pieces))
	for i := range bitmap {
		if bitmap[i] == '1' {
			piecesHave[i] = true
		}
	}

	// --- Проверяем файл ---
	fileExists := false
	if st, err := os.Stat(meta.FileName); err == nil {
		fileExists = true

		// проверяем размер файла
		if st.Size() != meta.FileSize {
			// файл повреждён → удаляем и создаём заново
			os.Remove(meta.FileName)
			fileExists = false
		}
	}

	// --- создаём файл, если он отсутствует ---
	if !fileExists {
		if err := filehandler.CreateFile(meta.FileName, meta.FileSize); err != nil {
			return nil, err
		}
	}

	// --- возвращаем объект ---
	return &Downloader{
		meta:       meta,
		ph:         ph,
		piecesHash: piecesHash,
		piecesHave: piecesHave,
	}, nil
}


// --- запрос списка пиров у трекера ---
func GetPeers(fileName string, trackerURL string) ([]string, error) {
	conn, err := net.Dial("tcp4", trackerURL+":3000")
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

// --- выбрать следующий кусок для загрузки ---
func (d *Downloader) NextPiece() int {
	d.mu.Lock()
	defer d.mu.Unlock()
	for i, have := range d.piecesHave {
		if !have {
			return i
		}
	}
	return -1
}

// --- скачивание одного куска с пира ---
func (d *Downloader) DownloadPiece(peer string, index int) error {
	conn, err := net.Dial("tcp4", peer+":3000")
	if err != nil {
		return err
	}
	defer conn.Close()

	req := netapi.CreateGetMessage(d.meta.FileName, index)
	data, _ := json.Marshal(req)

	lenBuf := make([]byte, 4)
	binary.BigEndian.PutUint32(lenBuf, uint32(len(data)))
	conn.Write(lenBuf)
	conn.Write(data)

	// читаем POST ответ
	respLenBuf := make([]byte, 4)
	if _, err := io.ReadFull(conn, respLenBuf); err != nil {
		return err
	}
	respLen := binary.BigEndian.Uint32(respLenBuf)

	respData := make([]byte, respLen)
	if _, err := io.ReadFull(conn, respData); err != nil {
		return err
	}

	var postResp netapi.PostPieceStruct
	if err := json.Unmarshal(respData, &postResp); err != nil {
		return err
	}

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
}

// обновление bitmap в PublicHouse
func (d *Downloader) UpdateBitmap() {
	d.mu.Lock()
	defer d.mu.Unlock()
	bitmap := ""
	for _, have := range d.piecesHave {
		if have {
			bitmap += "1"
		} else {
			bitmap += "0"
		}
	}
	d.ph.UpdateRecord(d.meta.FileName, bitmap)
}

// главная функция загрузки
func (d *Downloader) DownloadAll() error {
	peers, err := d.GetPeers()
	if err != nil {
		return err
	}
	if len(peers) == 0 {
		return fmt.Errorf("no peers available")
	}

	for {
		index := d.NextPiece()
		if index == -1 {
			break
		}
		success := false
		for _, peer := range peers {
			if err := d.DownloadPiece(peer, index); err == nil {
				success = true
				break
			}
		}
		if !success {
			return fmt.Errorf("failed to download piece %d from all peers", index)
		}
	}

	return nil
}

// --- вспомогательная функция для сравнения байтов ---
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

*/