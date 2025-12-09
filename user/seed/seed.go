package seed

// Пакет реализует раздачу файлов (seed)
// Протокол: 4 байта длины (big-endian) затем JSON

import (
	"encoding/binary"
	"encoding/json"
	"io"
	"net"
	"time"
	"user/logger"

	"user/core/filehandler"
	"user/netapi"
)
// переменная нашего сервера для возможности закрытия его из вне
var Listener net.Listener

// Параметр: размер куска
const PieceSize int64 = 128 * 1024 // 65536, изменить при необходимости

// Экземпляр PublicHouse (владеет mutex) для работы со списоком раздач
var ph *filehandler.PublicHouse

// Init позволяет передать готовый экземпляр PublicHouse из main
func Init(publicHouse *filehandler.PublicHouse) {
    ph = publicHouse
}

// SeedServer запускает TCP-сервер и принимает соединения.
func SeedServer() {

	if ph == nil {
        logger.Error("CRITICAL: SeedServer started without PublicHouse initialization!")
        return
    }


	logger.Info("start seed.SeedServer()")
	var err error
	Listener, err = net.Listen("tcp4", "0.0.0.0:3000")
	if err != nil {
		logger.Errorf("seed.SeedServer() have err = %v", err)
		return
	}
	defer Listener.Close()

	// создаём диспетчер и регистрируем обработчики
	dispatcher := netapi.NewDispatcher()
	dispatcher.Register("GET", GetPieceHandler)
	dispatcher.Register("HSH", HandshakeHandler)

	for {
		conn, err := Listener.Accept()
		if err != nil {
			opErr, ok := err.(*net.OpError)
            // Если это ошибка операции, и операция "accept", и она закрыта/прервана
            if ok && (opErr.Op == "accept" || opErr.Err.Error() == "use of closed network connection") {
                logger.Info("Seed server listener closed. Exiting server loop.")
                return 
			}
			// пропускаем соединения с ошибками
			logger.Errorf("seed.SeedServer() have err = %v", err)
			continue
		}
		go handleConn(conn, dispatcher)
	}
}

// handleConn читает один запрос, вызывает диспетчер и отправляет ответ
//1 запрос -> 1 ответ -> закрытие соединения
func handleConn(conn net.Conn, dispatcher *netapi.Dispatcher) {

	logger.Infof("start seed.handleConn(%s)", conn.RemoteAddr().String())

	defer conn.Close()

	// читаем длину (4 байта)
	conn.SetReadDeadline(time.Now().Add(5 * time.Second))
	lengthBuf := make([]byte, 4)
	_, err := io.ReadFull(conn, lengthBuf)
	if err != nil {
		logger.Errorf("seed.handleConn(%s) have err = %v", conn.RemoteAddr().String(), err)
		return
	}
	length := binary.BigEndian.Uint32(lengthBuf)

	// читаем JSON payload
	conn.SetReadDeadline(time.Now().Add(5 * time.Second))
	data := make([]byte, length)
	_, err = io.ReadFull(conn, data)
	if err != nil {
		logger.Errorf("seed.handleConn(%s) have err = %v", conn.RemoteAddr().String(), err)
		return
	}

	// вызываем диспетчер для обработки
	resp, err := dispatcher.Handle(data)
	if err != nil {
		logger.Errorf("seed.handleConn(%s) have err = %v", conn.RemoteAddr().String(), err)
		sendErrorResponse(conn, err.Error())
		return
	}

	// если хендлер вернул nil — не шлём ничего
	if resp == nil {
		return
	}

	// маршалим ответ в JSON и отправляем с префиксом длины
	sendResponse(conn, resp)
}

// sendResponse маршалит объект в JSON и шлёт его с 4-байтовой длиной (big-endian)
func sendResponse(conn net.Conn, resp interface{}) {

	logger.Infof("start seed.sendResponse(%s)", conn.RemoteAddr().String())

	payload, err := json.Marshal(resp)
	if err != nil {
		return
	}

	// выставляем write deadline и шлём
	conn.SetWriteDeadline(time.Now().Add(5 * time.Second))

	lenBuf := make([]byte, 4)
	binary.BigEndian.PutUint32(lenBuf, uint32(len(payload)))

	
	_, err = conn.Write(lenBuf)
	if err != nil {
		return
	}
	_, _ = conn.Write(payload)
}

// sendErrorResponse формирует сообщение об ошибке и отправляет его клиенту
func sendErrorResponse(conn net.Conn, errMsg string) {
	msg := netapi.CreateErrorMessage(errMsg)
	sendResponse(conn, msg)
}

// Хендлеры для Dispatcher

// GetPieceHandler — обрабатывает GET-запрос.
// Ожидает netapi.GetPieceStruct в jsonData.
// Возвращает netapi.PostPieceStruct (с полем Data) или netapi.ErrorRequestStruct.
func GetPieceHandler(jsonData []byte) (interface{}, error) {
	var req netapi.GetPieceStruct
	if err := json.Unmarshal(jsonData, &req); err != nil {
		return netapi.CreateErrorMessage(err.Error()), nil
	}

	// Проверим, есть ли у нас такой кусок
	has, err := ph.HasPiece(req.FileName, req.PieceIndex)
	if err != nil {
		// ошибка доступа к PublicHouse
		return netapi.CreateErrorMessage(err.Error()), nil
	}
	if !has {
		// у нас нет этого куска — отвечаем с ошибкой
		return netapi.CreateErrorMessage("piece not available"), nil
	}

	// Берём кусок из файловой системы
	fulName,err := ph.GetFullPathByName(req.FileName)
	if err != nil{
		return netapi.CreateErrorMessage(err.Error()), nil
	}
	data, err := filehandler.GetPiece(fulName, int64(req.PieceIndex), PieceSize)
	if err != nil {
		return netapi.CreateErrorMessage(err.Error()), nil
	}

	// формируем POST-ответ
	resp := netapi.CreatePostPieceMessage(req.FileName, data)
	return resp, nil
}

// HandshakeHandler — обрабатывает HSH-запрос.
// Ожидает netapi.HandShakeStruct (в котором интересует FileName).
// В ответ шлёт netapi.HandShakeStruct с bitmap (string).
func HandshakeHandler(jsonData []byte) (interface{}, error) {
	var req netapi.HandShakeStruct
	if err := json.Unmarshal(jsonData, &req); err != nil {
		return netapi.CreateErrorMessage(err.Error()), nil
	}

	bitmap, err := ph.GetBitmap(req.FileName)
	if err != nil {
		// если записи нет или ошибка — возвращаем сообщение об ошибке
		return netapi.CreateErrorMessage(err.Error()), nil
	}

	// Возвращаем Handshake с bitmap строкой
	resp := netapi.CreateHandShakeMessage(req.FileName, bitmap)
	return resp, nil
}
