package seed

// Пакет реализует раздачу файлов
// Функция SeedServer является главной
// ожидвется что будут получены GET и HSHK запросы с указанием

import (
	"encoding/binary"
	"encoding/json"
	"io"
	"log"
	"net"
	"testfile/netapi"
	"time"
)

func SeedServer(pullTask []string) {
	listen, err := net.Listen("tcp4", "0.0.0.0::3000")

	if err != nil {
		log.Fatal(err)
		//log.Println(err)
		//TODO
	}
	defer listen.Close()

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

	conn.SetReadDeadline(time.Now().Add(5 * time.Second))

	lengthBuf := make([]byte, 4)
	_, err := io.ReadFull(conn, lengthBuf)

	if err != nil {
		if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
			log.Println("Timeout from", conn.RemoteAddr(), "- closing connection")
			return
		}
		if err == io.EOF {
			log.Println("Client closed:", conn.RemoteAddr())
		} else {
			log.Println("Read error:", err)
		}
		return
	}

	length := binary.BigEndian.Uint32(lengthBuf)
	data := make([]byte, length)
	_, err = io.ReadFull(conn, data)

	if err != nil {
		if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
			log.Println("Timeout from", conn.RemoteAddr(), "- closing connection")
			return
		}
		if err == io.EOF {
			log.Println("Client closed:", conn.RemoteAddr())
		} else {
			log.Println("Read error:", err)
		}
		return
	}

	cmd,err := netapi.ParsCommandRequest(data)
	if err != nil{
		//TODO
		log.Println(err)
		return 
	}

	switch cmd{
	case "GET":
		var getCmd netapi.GetPiece
		err = json.Unmarshal(data,&getCmd)
		if err != nil{
			//TODO
			log.Println(err)
			return 
		}	
		getCmd.PieceIndex
	case "HDSHK":
		handshakeCmd := json.Unmarshal()
	default:
		log.Println("Error: undefind request")
		return
	}

}
