package netapi

import (
	"encoding/json"
	"errors"

	"user/logger"
)

var UnExistHandler = errors.New("Handlr not exist")
var UnExistCmd = errors.New("Command not not exist")

type BaseRequestStruct struct {
	Command string `json:"command"`
}

type GetPiece struct {
	Command string `json:"command"`

	FileName   string `json:"file_name"`
	PieceIndex int    `json:"piece_index"`
}

type PostPiece struct {
	Command string `json:"command"`

	FileName string `json:"file_name"`
	//PieceIndex int       `json:"piece_index"`
	Data []byte `json:"data"`
}

type Error struct {
	Command string `json:"command"`
	Error   string `json:"error"`
}

type HndShake struct {
	Command string `json:"command"`

	FileName string `json:"file_name"`
	//mu       sync.RWMutex // для потокобезопасности
	//PieceIndex int       `json:"piece_index"`
	BitMap int `json:"data"`
}

func ParsCommandRequest(jsonData []byte) (string, error) {
	var base BaseRequestStruct
	err := json.Unmarshal(jsonData, &base)

	if err != nil {
		// TODO
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

	cmd,err := ParsCommandRequest(jsonData)
	if err != nil{
		//TODO
		log.Println(err)
		return "", err
	}

	handler, exists := d.handlers[cmd]

	if !exists {
		// TODO
		return "", UnExistHandler
	}

	return handler(jsonData)
}
