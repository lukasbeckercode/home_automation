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
	"strconv"
	"time"
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

var remoteBinParts = []remoteBinPart{
	sampleRemoteBinPart,
}

var sampleRemoteAnalogPart = remoteAnalogPart{
	part{5, "TEMP5", 9999}, "0",
}

var remoteAnalogParts = []remoteAnalogPart{
	sampleRemoteAnalogPart,
}

func wsReader() {
	err := ws.WriteJSON(sampleRemoteAnalogPart)
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

func getRemoteAnalogPartByName(name string) (*remoteAnalogPart, error) {
	for i, t := range remoteAnalogParts {
		if t.Part == name {
			return &remoteAnalogParts[i], nil
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

func getRemoteBinPartByName(name string) (*remoteBinPart, error) {
	for i, t := range remoteBinParts {
		if t.Part == name {
			return &remoteBinParts[i], nil
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

func removeBinPart(context *gin.Context) {
	part, _ := getBinPartByName(context.Param("name"))
	for i, parts := range binParts {
		if parts.Id == part.Id {
			binParts = append(binParts[:i], binParts[i+1:]...)
			context.IndentedJSON(http.StatusOK, binParts)
			return
		}
	}
	context.IndentedJSON(http.StatusNotFound, gin.H{"message": "part not found"})
}

func removeAnalogPart(context *gin.Context) {
	part, _ := getAnalogPartByName(context.Param("name"))
	for i, parts := range analogParts {
		if parts.Id == part.Id {
			analogParts = append(analogParts[:i], analogParts[i+1:]...)
			context.IndentedJSON(http.StatusOK, analogParts)
			return
		}
	}
	context.IndentedJSON(http.StatusNotFound, gin.H{"message": "part not found"})
}

func addRemoteBinPart(context *gin.Context) {
	var newPart remoteBinPart
	err := context.BindJSON(&newPart)
	if err != nil {
		//TODO: error handling
		return
	}
	remoteBinParts = append(remoteBinParts, newPart)
	context.IndentedJSON(http.StatusCreated, newPart)
}

func getRemoteBinPart(context *gin.Context) {
	name := context.Param("part")
	command, err := getRemoteBinPartByName(name)
	if err != nil {
		context.IndentedJSON(http.StatusNotFound, gin.H{"message": "part not found"})
		return
	}
	context.IndentedJSON(http.StatusOK, command)
}

func getRemoteBinParts(context *gin.Context) {
	context.IndentedJSON(http.StatusOK, remoteBinParts)

}

func removeRemoteBinPart(context *gin.Context) {
	part, _ := getRemoteBinPartByName(context.Param("name"))
	for i, parts := range remoteBinParts {
		if parts.Id == part.Id {
			remoteBinParts = append(remoteBinParts[:i], remoteBinParts[i+1:]...)
			context.IndentedJSON(http.StatusOK, remoteBinParts)
			return
		}
	}
	context.IndentedJSON(http.StatusNotFound, gin.H{"message": "part not found"})
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
	for !token.Wait() {

	}
	if token.Error() != nil {
		panic(token.Error())
	}
	log.Printf("MQTT TOKEN: Topic:%s Message:%s\n", topic, message)
	context.IndentedJSON(http.StatusOK, sampleRemoteBinPart)
}
func getRemoteAnalogParts(context *gin.Context) {
	context.IndentedJSON(http.StatusOK, remoteAnalogParts)
}

func addRemoteAnalogPart(context *gin.Context) {
	var newPart remoteAnalogPart
	err := context.BindJSON(&newPart)
	if err != nil {
		//TODO: error handling
		return
	}
	remoteAnalogParts = append(remoteAnalogParts, newPart)
	context.IndentedJSON(http.StatusCreated, newPart)
}

func getRemoteAnalogPart(context *gin.Context) {
	name := context.Param("part")
	command, err := getRemoteAnalogPartByName(name)
	if err != nil {
		context.IndentedJSON(http.StatusNotFound, gin.H{"message": "part not found"})
		return
	}
	context.IndentedJSON(http.StatusOK, command)
}

func removeRemoteAnalogPart(context *gin.Context) {
	part, _ := getRemoteAnalogPartByName(context.Param("name"))
	for i, parts := range remoteAnalogParts {
		if parts.Id == part.Id {
			remoteAnalogParts = append(remoteAnalogParts[:i], remoteAnalogParts[i+1:]...)
			context.IndentedJSON(http.StatusOK, remoteAnalogParts)
			return
		}
	}
	context.IndentedJSON(http.StatusNotFound, gin.H{"message": "part not found"})
}

func getRemoteAnalogData(context *gin.Context, c chan string) {

	var analogRemoteDataHandler = func(client mqtt.Client, msg mqtt.Message) {
		var receivedMsg = ""
		bytes := msg.Payload()
		for i := 0; i < len(bytes); i++ {
			num, _ := strconv.ParseInt(string(bytes[i]), 10, 0)
			receivedMsg = fmt.Sprintf("%s%d", receivedMsg, num)
		}

		log.Println(&receivedMsg)

		c <- receivedMsg

		time.Sleep(1 * time.Second)

		return
	}

	name := context.Param("part")
	topic := fmt.Sprintf("topic/%s", name)
	token := client.Subscribe(topic, 0, analogRemoteDataHandler)

	if token.Error() != nil {
		panic(token.Error())
	}
	token.Wait()
}

func retrieveRemoteAnalogData(c chan string, context *gin.Context) {
	for {
		receivedMsg, ok := <-c
		sampleRemoteAnalogPart.Value = receivedMsg
		log.Printf("MQTT TOKEN:  Message:%s\n", receivedMsg)

		if ok == false {
			break
		}
	}

	/*if err != nil {
		context.IndentedJSON(http.StatusNotFound, gin.H{"message": "remote part does not exist"})
		fmt.Println(err)
	}*/
	context.IndentedJSON(http.StatusOK, sampleRemoteAnalogPart)
}
func handleRemoteAnalogData(context *gin.Context) {
	c := make(chan string)
	go getRemoteAnalogData(context, c)
	go retrieveRemoteAnalogData(c, context)

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
	client.Connect()

	//----------SETUP----------
	router := gin.Default()

	err := rpio.Open()
	if err != nil {
		fmt.Println(err)
	}

	//----------BINARY PARTS----------
	router.GET("/binparts", getBinParts)            // gets available binary binParts
	router.GET("/binparts/:part", getBinPart)       // gets a specific binary part
	router.PATCH("/binparts/:part", toggleOn)       // changes the On status of a specific binary part
	router.POST("/addbinpart", addBinPart)          // adds a binary part
	router.DELETE("/binparts/:part", removeBinPart) //removes bin part

	//----------ANALOG PARTS----------
	router.GET("/analogparts", getAnalogParts)
	router.GET("/analogparts/:part", getAnalogPart)
	router.POST("/addanalogpart", addAnalogPart)
	router.DELETE("/analogparts/:part", removeAnalogPart)
	//UPDATE is implemented in the mqtt side of this project

	//----------REMOTE BIN PARTS----------
	router.GET("/binparts/remote/", getRemoteBinParts)
	router.GET("/binparts/remote/:part", getRemoteBinPart)
	router.POST("addbinpart/remote", addRemoteBinPart)
	router.PATCH("/binparts/remote/:part", toggleRemotePart) // changes the On status of a specific binary part
	router.DELETE("/binparts/remote/:part", removeRemoteBinPart)

	//----------REMOTE ANALOG PARTS----------
	router.GET("/analogparts/remote/", getRemoteAnalogParts)
	router.GET("/analogparts/remote/:part", handleRemoteAnalogData)
	router.POST("/analogparts/remote/addpart/", addRemoteAnalogPart)
	router.DELETE("/analogparts/remote/:part", removeRemoteAnalogPart)

	//----------WEBSOCKET----------
	router.GET("/values", wsAnalog)

	//----------RUN----------
	path := GetOutboundIP().String() + ":9090"
	//err = router.Run("localhost:9090")
	err = router.Run(path)
	if err != nil {
		return
	}

}
