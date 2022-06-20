# Home Automation 
## Contributors 
Lukas Becker (MSD20)
## First Small Description
Use a RestAPI written in GoLang to turn LEDs on a RaspberryPi 4 on and off. Also, the API should be able to return values
measured from analog senors (such as temperature sensors)

## Setup instructions
### Necessary Hardware
- A raspberry pi (ideally version 4, not tested on other products)
- All the hardware needed to use the RasPi (SD Card, Power Supply, etc. )
- Jumper Wires (ideally Female to Male to work with a breadboard)
- Breadboard 
- LEDs + Resistors (around 330 Ohms)
- Some sort of analog Sensor (Potentiometer, Temperature Sensor, etc)
- An ESP8266 Development Board with the Software from [this]("https://github.com/lukasbeckercode/esp8266_mqtt") repository flashed 
### Necessary Software
- Latest Raspbian installation
- Working adafruit.io account
### mqtt Key
To establish connection with the adafruit.io mqtt server, a key is needed. As this key is private and grants
access to the entire adafruit.io mqtt services from my account, this is key is not shared in this repo 
### How to install
 Simply run setup.sh to install Go and its dependencies
### How to run
- cd into the root directory of the project (where this file is located)
- run ```./run.sh```

## Technologies used
- [Go]("https://go.dev") as the main programming language, as it is easy and fast to work with 
- [Gin Gonic Library]("https://gin-gonic.com") to easily create a REST API 
- [Adafruit IO Account]("https://io.adafruit.com/") to use as a [MQTT]("https://mqtt.org") Broker 
- [Arduino IDE]("https://www.arduino.cc") setup to use an [ESP8266 Development Board]("https://en.wikipedia.org/wiki/ESP8266")
- [Raspberry Pi 4]("https://www.raspberrypi.org") with [Raspbian]("https://www.raspberrypi.com/software/") installed on it 

### Go and Gin Gonic 
GoLang was used for this project as it is an up and coming, new programming language which makes it 
especially pleasant to create webservices. It is appealing to anyone, who likes the speed of 
interpreted languages but prefers compiled languages. Go resembles C-Like languages more than JS or 
Python, which makes it easy to learn. Even though Go is compiled, it is incredibly fast. Another advantage 
of Go is, that it is very Cross-platform friendly. This project, for instance, was developed on a Mac
and could, without any changes be transferred to the Raspberry Pi using Linux. 
Gin Gonic is a powerful framework for creating REST APIs. Using this framework made it a great experience 
to write the API. 

### REST vs MQTT
Initially, I wanted to compare Message Queueing to REST APIs in this project. However, it quickly
turned out, that those technologies aren't mutually exclusive, but can be used together. In this project,
the REST API is used to communicate with end user devices (such as a phone). Message Queueing, in the form 
of MQTT is used to establish communication between the Raspberry Pi and the ESP8266. Combining those 
technologies creates the opportunity to create powerful systems with very low cost. 

### Adafruit IO and MQTT 
In this project, adafruit io is used as the message broker for MQTT. During the development, 
I tried writing my own MQ Broker, but I quickly found myself in integration hell, where all 
the individual parts where working, but combining them was impossible. The easiest fix was
to create a free adafruit io account and use this. MQTT was used instead of "regular MQ" as it is 
the gold standard for working with IOT devices 

### Hardware: Raspberry Pi and ESP8266
A raspberry pi is a small and cheap single board computer. It can be used for a variety of projects. 
In this scenario, it is used as the "brain" of the operation. It exposes the REST API and 
sends commands intended for remote parts to the MQTT Broker. It also offers GPIO Pins that can be 
used to directly wire parts to them. On the other hand, we have the ESP8266. I used a development board
with standard pin spacing here. It is far less powerful than the Pi, but also cheaper. It offers 
GPIO Pins and WiFi connectivity out of the box and is therefore perfect to be used as a remote part. 
This development board only listens to the MQTT Broker and updated sensor values everytime they are changed. 
The only downside of this development board is, that is has to be programmed in C, luckily, the 
Arduino community provides a lot of libraries so very little C experience is required. 

## Known Issues and Limitations 
- Not all the CRUD commands are implemented for all the functionality. This is because they were not 
needed to showcase the technologies used here
- None of this was tested in an actual home automation environment. I am not an electrician, so I
refuse to work with main power voltage for safety reasons
- Analog sensor reading was not implemented on the raspberry from the hardware point of view. The reason for this being,
that this has the potential to destroy the Pi if done incorrectly which had caused significant delays in development 
- Setting everything up can be tedious. The ESP8266 has to be newly flashed every time the Wifi changes in some way
Also, finding the Pi in a network you aren't the administrator of can cause problems. 

