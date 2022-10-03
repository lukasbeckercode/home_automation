package main

import (
	"errors"
	"fmt"
	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/stianeikeland/go-rpio"
	"log"
	"net"
	"net/http"
)

var client mqtt.Client
var messagePubHandler mqtt.MessageHandler = func(client mqtt.Client, msg mqtt.Message) {
	log.Printf("Received message: %s from topic: %s\n", msg.Payload(), msg.Topic())
}

var connectHandler mqtt.OnConnectHandler = func(client mqtt.Client) {
	log.Println("Connected")
}

var connectLostHandler mqtt.ConnectionLostHandler = func(client mqtt.Client, err error) {
	log.Printf("Connect lost: %v\n", err)
}
var ws *websocket.Conn

var wsConnected = false

type part struct {
	Id   int    `json:"id"`
	Part string `json:"part"`
	Pin  int    `json:"pin"`
}
type binPart struct { // represents a part with a binary state (e.g. a LED)
	part
	On bool `json:"on"`
}

type analogPart struct { //represents an analog part
	part
	Value int `json:"value"`
}

type remoteBinPart struct {
	part
	On bool `json:"on"`
}

type remoteAnalogPart struct {
	part
	Value string `json:"value"`
}

var binParts = []binPart{ //Array of binary part s
	//TODO get right pin numbers
	{part{0, "LED1", 0}, false},
	{part{1, "LED2", 1}, false},
}

var analogParts = []analogPart{ // Array of analog parts
	{part{2, "TEMP1", 2}, 0},
	{part{3, "TEMP2", 3}, 0},
}

var sampleRemoteBinPart = remoteBinPart{
	part{4, "LED5", 9999}, false,
}

var sampleRemoteAnalogPart = remoteAnalogPart{
	part{5, "TEMP5", 9999}, "0",
}

func wsReader() {
	err := ws.WriteJSON(binParts)
	if err != nil {
		panic(err)
	}

}
func toggleInternalLed(pinNum int) {
	pin := rpio.Pin(pinNum)
	pin.Output()
	pin.Toggle()
}

func getAnalogParts(context *gin.Context) {
	context.IndentedJSON(http.StatusOK, analogParts)
}

func getAnalogPartByName(name string) (*analogPart, error) {
	for i, t := range analogParts {
		if t.Part == name {
			return &analogParts[i], nil
		}
	}
	return nil, errors.New("part not found")
}

func getAnalogPart(context *gin.Context) {
	name := context.Param("part")
	part, err := getAnalogPartByName(name)
	if err != nil {
		context.IndentedJSON(http.StatusNotFound, gin.H{"message": "part not found"})
		return
	}
	context.IndentedJSON(http.StatusOK, part)
}

func addAnalogPart(context *gin.Context) {
	var newPart analogPart
	err := context.BindJSON(&newPart)
	if err != nil {
		return
	}
	analogParts = append(analogParts, newPart)
	context.IndentedJSON(http.StatusOK, newPart)
}

func getBinParts(context *gin.Context) {
	context.IndentedJSON(http.StatusOK, binParts)

}

func getBinPartByName(name string) (*binPart, error) {
	for i, t := range binParts {
		if t.Part == name {
			return &binParts[i], nil
		}
	}
	return nil, errors.New("part not found")
}

func getBinPart(context *gin.Context) {
	name := context.Param("part")
	command, err := getBinPartByName(name)
	if err != nil {
		context.IndentedJSON(http.StatusNotFound, gin.H{"message": "part not found"})
		return
	}
	context.IndentedJSON(http.StatusOK, command)
}

func toggleOn(context *gin.Context) {
	name := context.Param("part")
	command, err := getBinPartByName(name)
	if err != nil {
		context.IndentedJSON(http.StatusNotFound, gin.H{"message": "part not found"})
		return
	}
	command.On = !command.On
	toggleInternalLed(command.Pin)
	context.IndentedJSON(http.StatusOK, command)

	if wsConnected {
		wsReader()
	}
}

func addBinPart(context *gin.Context) {
	var newPart binPart
	err := context.BindJSON(&newPart)
	if err != nil {
		//TODO: error handling
		return
	}
	binParts = append(binParts, newPart)
	context.IndentedJSON(http.StatusCreated, newPart)

}

func toggleRemotePart(context *gin.Context) {
	name := context.Param("part")
	/*if err != nil {
		context.IndentedJSON(http.StatusNotFound, gin.H{"message": "remote part does not exist"})
		fmt.Println(err)
	}*/

	sampleRemoteBinPart.On = !sampleRemoteBinPart.On

	var message string
	if sampleRemoteBinPart.On {
		message = "TRUE"
	} else {
		message = "FALSE"
	}

	topic := fmt.Sprintf("topic/%s", name)
	token := client.Publish(topic, 0, false, message)
	token.Wait()
	log.Printf("MQTT TOKEN: Topic:%s Message:%s\n", topic, message)
	context.IndentedJSON(http.StatusOK, sampleRemoteBinPart)
}

func getRemoteAnalogData(context *gin.Context) {
	var receivedMsg string
	var analogRemoteDataHandler = func(client mqtt.Client, msg mqtt.Message) {
		receivedMsg = fmt.Sprintf("%s", msg.Payload())
	}
	name := context.Param("part")
	topic := fmt.Sprintf("topic/%s", name)
	token := client.Subscribe(topic, 0, analogRemoteDataHandler)
	token.Wait()
	log.Printf("MQTT TOKEN: Topic:%s Message:%s\n", topic, receivedMsg)
	/*if err != nil {
		context.IndentedJSON(http.StatusNotFound, gin.H{"message": "remote part does not exist"})
		fmt.Println(err)
	}*/
	sampleRemoteAnalogPart.Value = receivedMsg
	context.IndentedJSON(http.StatusOK, sampleRemoteAnalogPart)
}

// GetOutboundIP Get preferred outbound ip of this machine
func GetOutboundIP() net.IP {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		log.Fatal(err)
	}
	defer func(conn net.Conn) {
		err := conn.Close()
		if err != nil {

		}
	}(conn)

	localAddr := conn.LocalAddr().(*net.UDPAddr)

	return localAddr.IP
}

func wsAnalog(context *gin.Context) {
	var upgrader = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
	}
	upgrader.CheckOrigin = func(r *http.Request) bool {
		return true
	}
	var err error
	ws, err = upgrader.Upgrade(context.Writer, context.Request, nil)
	if err != nil {
		panic(err)
	}
	wsReader()
	wsConnected = true

}

func main() {
	options := mqtt.NewClientOptions()
	broker := GetOutboundIP().String()
	port := 1883
	options.AddBroker(fmt.Sprintf("%s:%d", broker, port))
	log.Printf("Connected to: tcp://%s:%d\n", broker, port)
	options.SetClientID("raspberry_pi")
	options.SetUsername("pi")
	options.SetPassword("raspberry")
	options.SetDefaultPublishHandler(messagePubHandler)
	options.OnConnect = connectHandler
	options.OnConnectionLost = connectLostHandler
	client = mqtt.NewClient(options)

	//----------SETUP----------
	router := gin.Default()

	err := rpio.Open()
	if err != nil {
		fmt.Println(err)
	}

	//----------BINARY PARTS----------
	router.GET("/binparts", getBinParts)      // gets available binary binParts
	router.GET("/binparts/:part", getBinPart) // gets a specific binary part
	router.PATCH("/binparts/:part", toggleOn) // changes the On status of a specific binary part
	router.POST("/addbinpart", addBinPart)    // adds a binary part

	//----------ANALOG PARTS----------
	router.GET("/analogparts", getAnalogParts)
	router.GET("/analogparts/:part", getAnalogPart)
	router.POST("/addanalogpart", addAnalogPart)

	//----------REMOTE PARTS----------
	router.PATCH("/binparts/remote/:part", toggleRemotePart) // changes the On status of a specific binary part
	router.GET("/analogparts/remote/:part", getRemoteAnalogData)
	//Websocket
	router.GET("/values", wsAnalog)

	//----------RUN----------
	path := GetOutboundIP().String() + ":9090"
	//err = router.Run("localhost:9090")
	err = router.Run(path)
	if err != nil {
		return
	}

}
