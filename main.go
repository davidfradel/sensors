package main

import (
	"encoding/json"
	"log"
	"os"
	"os/signal"
	"time"

	"github.com/gorilla/websocket"
)

var done chan interface{}
var interrupt chan os.Signal

func receiveHandler(connection *websocket.Conn) {
	defer close(done)
	for {
		_, msg, err := connection.ReadMessage()
		if err != nil {
			log.Println("Error in receive:", err)
			return
		}
		log.Printf("Received: %s\n", msg)
	}
}

type myStruct struct {
	Sensor_id    string    `json:"sensor_id"`
	Room_id      string    `json:"room_id"`
	Floor_id     string    `json:"floor_id"`
	Building_id  string    `json:"building_id"`
	Current_time time.Time `json:"current_time"`
}

func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if len(value) == 0 {
		return defaultValue
	}
	return value
}

func main() {
	done = make(chan interface{})    // Channel to indicate that the receiverHandler is done
	interrupt = make(chan os.Signal) // Channel to listen for interrupt signal to terminate gracefully

	signal.Notify(interrupt, os.Interrupt) // Notify the interrupt channel for SIGINT

	host := getEnv("SERVER_HOST", "localhost")
	port := getEnv("SERVER_PORT", ":8080")
	socketUrl := "ws://" + host + port + "/socket"
	conn, _, err := websocket.DefaultDialer.Dial(socketUrl, nil)
	if err != nil {
		log.Fatal("Error connecting to Websocket Server:", err)
	}
	defer conn.Close()
	go receiveHandler(conn)

	// Our main loop for the client
	// We send our relevant packets here
	for {
		select {
		case <-time.After(time.Duration(600) * time.Millisecond * 1000):
			// Send an echo packet every 10 minutes (600s)

			data := myStruct{Sensor_id: "S_4320593002", Room_id: "R_2394023243", Floor_id: "F_4549042302", Building_id: "B_8950240359"}
			test, err := json.Marshal(data)
			//valStr := fmt.Sprintf("%+v\n", data)

			if err != nil {
				log.Println("okoko:", err)
				return
			}

			err2 := conn.WriteMessage(websocket.TextMessage, []byte(test))
			if err != nil {
				log.Println("Error during writing to websocket:", err2)
				return
			}

		case <-interrupt:
			// We received a SIGINT (Ctrl + C). Terminate gracefully...
			log.Println("Received SIGINT interrupt signal. Closing all pending connections")

			// Close our websocket connection
			err := conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
			if err != nil {
				log.Println("Error during closing websocket:", err)
				return
			}

			select {
			case <-done:
				log.Println("Receiver Channel Closed! Exiting....")
			case <-time.After(time.Duration(1) * time.Second):
				log.Println("Timeout in closing receiving channel. Exiting....")
			}
			return
		}
	}
}
