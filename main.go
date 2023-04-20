package main

import (
	"bufio"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/websocket"
)

// CheckOrigin := func(r *http.Request) bool{
// 	return true
// }

var ClientUpgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

var logFilePath string = "/Users/adityaagrawal/Downloads/testServer.txt"

// We can have a channel and the server will be streaming updates in that channel

func WriteToClient(clientConn *websocket.Conn, broadCastlogs chan string) {

	//this will continuously pull out from channel and write to client.
	for {
		select {
		case logs := <-broadCastlogs:
			err := clientConn.WriteMessage(websocket.TextMessage, []byte(logs))
			if err != nil {
				log.Printf("Error in sending message to client: %v", err)
				return
			}
		}
	}

}

func pushUpdatestoChannel(filePath string, startPosition int64, broadCastlogs chan string) int64 {

	file, err := os.Open(filePath)
	if err != nil {
		log.Printf("Error in opening file %v", err)
		panic(err)
	}

	defer file.Close()

	// if statrtposition is 0 we can print last 10 lines else only recent updates

	if startPosition == 0 {
		lineCount := 0
		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			lineCount++
		}

		// seek to last n lines
		numberOfLines := 10
		//move to beginning
		_, err := file.Seek(0, 0)
		if err != nil {
			fmt.Printf("Error in moving back to starting of file %v", err)
			return 0
		}
		scanner = bufio.NewScanner(file)
		startLine := lineCount - numberOfLines
		//moving to the start of the last 10 lines
		for i := 0; i < startLine; i++ {
			scanner.Scan()
		}

		for scanner.Scan() {
			log := scanner.Text()
			broadCastlogs <- log
		}
		startPosition = int64(lineCount)
	} else {
		scanner := bufio.NewScanner(file)
		//
		var ct int64
		//just check whether we are before startPosition or not?
		for scanner.Scan() {
			log := scanner.Text()
			ct++
			if ct-startPosition > 0 {
				broadCastlogs <- log
			}
		}
		startPosition = ct
	}

	return startPosition

}

func logWatchhandler(w http.ResponseWriter, req *http.Request) {
	clientConn, err := ClientUpgrader.Upgrade(w, req, nil)
	if err != nil {
		log.Println("Failed in creating websocket connection ", err)
		return
	}
	var broadcastLogs = make(chan string)
	var startPosition int64 = 0

	go func() {
		for {
			// this startposition will be updated each time.
			startPosition = pushUpdatestoChannel(logFilePath, startPosition, broadcastLogs)
			time.Sleep(time.Second)
		}
	}()

	//here server will write to the client.
	go WriteToClient(clientConn, broadcastLogs)

}

func main() {

	http.HandleFunc("/log", logWatchhandler)

	log.Println("Server starting at port: 8080")

	log.Fatal(http.ListenAndServe(":8080", nil))
}
