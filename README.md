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
- An ESP8266 Development Board with the Software from [this]("") repository flashed 
### Necessary Software
- Latest Raspbian installation
### mqtt Key
To establish connection with the adafruit.io mqtt server, a key is needed. As this key is private and grants
access to the entire adafruit.io mqtt services from my account, this is key is not shared in this repo 
### How to install
 Simply run setup.sh to install Go and its dependencies
### How to run
- cd into the root directory of the project (where this file is located)
- run ```./run.sh```

