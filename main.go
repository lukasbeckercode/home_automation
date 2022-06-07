package main

import (
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"strings"
)

var feeds []*Feed
var client *Client

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
	//TODO: Set GPIO Pin to HIGH
	command.On = !command.On
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

func toggleRemotePart(context *gin.Context) {
	name := context.Param("part")
	idx, err := findMQTTChannel(name)
	if err != nil {
		context.IndentedJSON(http.StatusNotFound, gin.H{"message": "remote part does not exist"})
		fmt.Println(err)
	}

	client.SetFeed(feeds[idx])

	sampleRemoteBinPart.On = !sampleRemoteBinPart.On

	var message string
	if sampleRemoteBinPart.On {
		message = "TRUE"
	} else {
		message = "FALSE"
	}

	_, _, err = client.Data.Send(&Data{Value: message})
	if err != nil {
		return
	}

	context.IndentedJSON(http.StatusOK, sampleRemoteBinPart)
}

func getRemoteAnalogData(context *gin.Context) {
	name := context.Param("part")
	idx, err := findMQTTChannel(name)

	if err != nil {
		context.IndentedJSON(http.StatusNotFound, gin.H{"message": "remote part does not exist"})
		fmt.Println(err)
	}
	client.SetFeed(feeds[idx])
	data, _, _ := client.Data.Last()
	sampleRemoteAnalogPart.Value = data.Value
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

func getChannelKey(key *string) {
	file, err := ioutil.ReadFile("generated_key.txt")

	if err != nil {
		panic(err)
	}
	tmpKey := string(file)
	*key = strings.Trim(tmpKey, "\n")
}

func findMQTTChannel(part string) (int, error) {
	for index, feed := range feeds {
		if feed.Name == part {
			return index, nil
		}
	}
	return 9999, errors.New("cannot find feed")
}

func main() {
	var channelKey string
	getChannelKey(&channelKey)

	//TODO: add mqtt setup
	client = NewClient(channelKey)
	var err error
	feeds, _, err = client.Feed.All()
	if err != nil {
		panic(err)
	}

	//----------SETUP----------
	router := gin.Default()

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

	//----------RUN----------
	path := GetOutboundIP().String() + ":9090"
	//err := router.Run("localhost:9090")
	err = router.Run(path)
	if err != nil {
		return
	}
}
