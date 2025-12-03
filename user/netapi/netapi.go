package netapi

import (
	"encoding/json"
	"errors"

	"user/logger"
)

var ErrUnExistHandler = errors.New("Handlr not exist")
var ErrUnExistCmd = errors.New("Command not not exist")

type BaseRequestStruct struct {
	Command string `json:"command"`
}

type GetPieceStruct struct {
	Command string `json:"command"`
	FileName   string `json:"file_name"`
	PieceIndex int    `json:"piece_index"`
}

type PostPieceStruct struct {
	Command string `json:"command"`
	FileName string `json:"file_name"`
	Data []byte `json:"data"`
}

type ErrorRequestStruct struct {
	Command string `json:"command"`
	Error   string `json:"error"`
}

type HandShakeStruct struct {
	Command string `json:"command"`
	FileName string `json:"file_name"`
	BitMap   string `json:"bitmap"`
}

type GetPeersStruct struct {
	Command  string `json:"command"`
	FileName string `json:"file_name"`
}


// И парсер для ответа от трекера
type PeersResponseStruct struct {
	Command string   `json:"command"`
	Peers   []string `json:"peers"`
}

type ErrorResponseStruct struct {
	Command string `json:"command"`
	Error   string `json:"error"`
}

func CreateGetPeersMessage(fileName string) *GetPeersStruct {
	return &GetPeersStruct{
		Command:  "PEERS",
		FileName: fileName,
	}
}

func CreateGetMessage(fileName string, pieceIndex int) (*GetPieceStruct){
	cmd := "GET"
	return &GetPieceStruct{
		Command: cmd,
		FileName: fileName,
		PieceIndex: pieceIndex,
	}
}

func CreateErrorMessage(eRror string) (*ErrorRequestStruct){
	cmd := "ERR"
	return &ErrorRequestStruct{
		Command: cmd,
		Error: eRror,
	}
}

func CreateHandShakeMessage(fileName string, bitMap string ) (*HandShakeStruct){	
	cmd := "HSH"
	return &HandShakeStruct{
		Command: cmd,
		FileName: fileName,

	}
}


func CreatePostPieceMessage(fileName string, data []byte) (*PostPieceStruct){
	cmd := "POST"
	return &PostPieceStruct{
		Command: cmd,
		FileName: fileName,
		Data: data,
	}
}


func ParsCommandRequest(jsonData []byte) (string, error) {
	var base BaseRequestStruct
	err := json.Unmarshal(jsonData, &base)

	if err != nil {
		logger.Errorf("netapi.ParsCommandRequest(...) have err = %v", err)
		return "", err
	}

	return base.Command, nil

}

// абстарктный обработчик запроса
type Handler func(jsonData []byte) (interface{}, error)

// диспетчер обработчиков
type Dispatcher struct {
	handlers map[string]Handler
}

// Создание диспетчера обаботчика
func NewDispatcher() *Dispatcher {

	logger.Info("start NewDispatcher()")

	return &Dispatcher{
		handlers: make(map[string]Handler),
	}
}

// функция для решистрации обработчиков
func (d *Dispatcher) Register(command string, handler Handler) {

	d.handlers[command] = handler
}

// функция для обработки полученной последовательнсти байт и вызова хендлеров относительно запроса
func (d *Dispatcher) Handle(jsonData []byte) (interface{}, error) {

	logger.Info("start netapi.Dispatcher.Handle(...)")

	cmd,err := ParsCommandRequest(jsonData)
	if err != nil{
		logger.Errorf("netapi.Dispatcher.Handle(...) have err = %v", err)
		return "", err
	}

	handler, exists := d.handlers[cmd]

	if !exists {
		logger.Errorf("netapi.Dispatcher.Handle(...) have err = %v", ErrUnExistHandler)
		return "", ErrUnExistHandler
	}

	return handler(jsonData)
}