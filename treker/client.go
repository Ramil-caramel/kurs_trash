package main

import (
	"encoding/binary"
	"fmt"
	"io"
	"log"
	"net"
	"time"
	"os"
)

func main() {
	conn, err := net.Dial("tcp4", "10.228.203.57:3000")
	time.Sleep(7000)
	if err != nil {
		log.Fatal("Dial error:", err)
	}
	defer conn.Close()

	// --- 1. Формируем и отправляем запрос ---
	// Длина полезной нагрузки (0 байт, просто запрос)
	header := make([]byte, 5)
	binary.BigEndian.PutUint32(header, 0) // длина = 0
	header[4] = 'G'                       // тип сообщения = 'G'

	_, err = conn.Write(header)
	if err != nil {
		log.Fatal("Write request error:", err)
	}

	// --- 2. Читаем заголовок ответа ---
	respHeader := make([]byte, 5)
	_, err = io.ReadFull(conn, respHeader)
	if err != nil {
		log.Fatal("Read header error:", err)
	}

	length := binary.BigEndian.Uint32(respHeader[:4])
	msgType := respHeader[4]

	fmt.Printf("Response header: len=%d, type=%c\n", length, msgType)

	switch msgType {
	case 'T':
		// --- 3. Читаем тело сообщения ---
		file, err := os.OpenFile("PeerDataClient", os.O_CREATE|os.O_RDWR|os.O_TRUNC,0644)
		if err != nil {
			log.Println(err)
			//TODO
			//return err
		}
		defer file.Close()

		data := make([]byte, length)
		_, err = io.ReadFull(conn, data)
		if err != nil {
			log.Fatal("Read body error:", err)
		}
		fmt.Println("Response body:")
		fmt.Println(string(data))
		file.WriteString(string(data) + "\n")

	case 'E':
		fmt.Println("Server returned an error")
	default:
		fmt.Println("Unknown response type:", msgType)
	}
}
