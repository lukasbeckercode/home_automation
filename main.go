package main

import (
	"errors"
	"github.com/gin-gonic/gin"
	"net/http"
)

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

var binParts = []binPart{ //Array of binary parts
	//TODO get right pin numbers
	{part{0, "LED1", 0}, false},
	{part{1, "LED2", 1}, false},
}

var analogParts = []analogPart{ // Array of analog parts
	{part{2, "TEMP1", 2}, 0},
	{part{3, "TEMP2", 3}, 0},
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

func main() {
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

	//----------RUN----------
	err := router.Run("localhost:9090")
	if err != nil {
		return
	}
}
