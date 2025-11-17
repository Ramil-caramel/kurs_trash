package main

import (
	"bufio"
	"encoding/binary"
	"io"
	"log"
	"net"
	"os"
	"sync"
	"time"
)

var fileMutex sync.Mutex

func main() {
	listen, err := net.Listen("tcp4", "10.28.203.57:3000")
	if err != nil {
		log.Fatal(err)
		//log.Println(err)
		//TODO
	}
	defer listen.Close()

	log.Println("Listening on :3000")

	for {
		conn, err := listen.Accept()
		if err != nil {
			log.Println(err)
			//TODO
			continue
		}
		go handleConn(conn)
	}
}

func handleConn(conn net.Conn) {
	defer conn.Close()
	remoteAddr := conn.RemoteAddr().String()

	addr,_,err := net.SplitHostPort(remoteAddr)
	if err !=nil {
		log.Println("can`t split ip and port")
		return
	}

	log.Printf("New connection from %s\n", addr)

	for {
		conn.SetReadDeadline(time.Now().Add(5 * time.Second))

		header := make([]byte, 5)
		_, err := io.ReadFull(conn, header)

		if err != nil {
			// Проверяем, не таймаут ли это
			if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
				log.Println("Timeout from", addr, "- closing connection")
				return
			}
			if err == io.EOF {
				log.Println("Client closed:", addr)
			} else {
				log.Println("Read error:", err)
			}
			return
		}
		conn.SetReadDeadline(time.Time{})

		//length := int(header[0])<<24 | int(header[1])<<16 |
		//	int(header[2])<<8 | int(header[3])

		msgType := header[4]

		if msgType == 'G' {
			answer, err := handleFile(addr)
			if err != nil {
				log.Println("File error:", err)
				sendError(conn)
				continue
			}

			sendResponse(conn, answer)
		} else {
			sendError(conn)
		}
	}
}

func sendResponse(conn net.Conn, data []byte) {
	header := make([]byte, 5)
	binary.BigEndian.PutUint32(header, uint32(len(data)))
	header[4] = 'T'

	if _, err := conn.Write(header); err != nil {
		log.Println("Write header error:", err)
		return
	}

	if len(data) > 0 {
		if _, err := conn.Write(data); err != nil {
			log.Println("Write data error:", err)
			return
		}
	}
    log.Printf("Sending %d bytes to %s\n", len(data), conn.RemoteAddr())
    //TODO
}

func sendError(conn net.Conn) {
	header := make([]byte, 5)
    binary.BigEndian.PutUint32(header, 0)
	header[4] = 'E'
	if _, err := conn.Write(header); err != nil {
		log.Println("Write error:", err)
	}
}

func handleFile(addr string) ([]byte, error) {
	fileMutex.Lock() // ← Блокируем доступ к файлу
	defer fileMutex.Unlock()

	file, err := os.OpenFile("PeerData", os.O_CREATE|os.O_RDWR|os.O_APPEND, 0644)

	if err != nil {
		log.Println(err)
		//TODO
		return nil, err
	}
	defer file.Close()

	var result []byte

	scanner := bufio.NewScanner(file)
	flag := true
	for scanner.Scan() {
		line := scanner.Text()
		if line == addr {
			flag = false
			continue
		}
		result = append(result, line...)
		result = append(result, '\n')
	}
	if flag {
		_, err = file.WriteString(addr + "\n")
		if err != nil {
			log.Println(err)
			//TODO
			return nil, err
		}
	}

	return result, nil
}

/*

// Message — пример структуры данных, которую клиент шлёт в JSON
type Message struct {
    Type string `json:"type"`
    Text string `json:"text"`
}

// handleConn — обрабатывает одно клиентское соединение
func handleConn(conn net.Conn) {
    // defer — гарантирует, что соединение закроется при выходе из функции
    defer conn.Close()

    // Удобно знать, кто к нам подключился
    addr := conn.RemoteAddr().String()
    log.Printf("New connection from %s\n", addr)

    // Бесконечный цикл — обрабатываем поток сообщений от клиента
    for {
        // 1️⃣ Сначала читаем 4 байта длины
        lenBuf := make([]byte, 4)
        // io.ReadFull — читает ровно len(lenBuf) байт или возвращает ошибку
        _, err := io.ReadFull(conn, lenBuf)
        if err != nil {
            // io.EOF — это “клиент корректно закрыл соединение”
            if err == io.EOF {
                log.Printf("[%s] client closed connection\n", addr)
            } else {
                log.Printf("[%s] read length error: %v\n", addr, err)
            }
            return // прекращаем обработку этого клиента
        }

        // 2️⃣ Преобразуем 4 байта в число (uint32)
        msgLen := binary.BigEndian.Uint32(lenBuf)

        // 3️⃣ Теперь читаем само сообщение длиной msgLen байт
        msgBuf := make([]byte, msgLen)
        _, err = io.ReadFull(conn, msgBuf)
        if err != nil {
            log.Printf("[%s] read message error: %v\n", addr, err)
            return
        }

        // 4️⃣ Парсим JSON в структуру
        var msg Message
        if err := json.Unmarshal(msgBuf, &msg); err != nil {
            log.Printf("[%s] invalid JSON: %v\n", addr, err)
            // можно не разрывать соединение, просто пропустить это сообщение
            continue
        }

        // 5️⃣ Выводим полученные данные
        log.Printf("[%s] received message: %+v\n", addr, msg)

        // 6️⃣ Отправляем ответ (echo)
        reply := map[string]string{
            "status": "ok",
            "echo":   msg.Text,
        }
        replyData, _ := json.Marshal(reply)

        // Отправляем длину ответа
        lenBuf = make([]byte, 4)
        binary.BigEndian.PutUint32(lenBuf, uint32(len(replyData)))
        conn.Write(lenBuf)

        // И сам ответ
        conn.Write(replyData)
    }
}

// main — основной цикл сервера
func main() {
    // 1️⃣ Слушаем порт 3000 (TCP)
    ln, err := net.Listen("tcp", ":3000")
    if err != nil {
        log.Fatal("Listen error:", err)
    }
    defer ln.Close()
    log.Println("Server listening on :3000")

    // 2️⃣ Главный цикл: ждём клиентов
    for {
        // Блокирует выполнение, пока кто-то не подключится
        conn, err := ln.Accept()
        if err != nil {
            log.Println("Accept error:", err)
            continue
        }

        // 3️⃣ Для каждого клиента запускаем отдельную горутину
        // чтобы сервер мог обслуживать нескольких одновременно
        go handleConn(conn)
    }
}

*/
