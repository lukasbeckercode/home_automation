echo "Looking for GoLang installation..."
if [[ $(go version) == *"go version"* ]]; then
  echo "GoLang already installed!"
  go version
else
  #download go
  echo "Getting GoLang from apt"
  sudo apt-get install golang
fi

echo "install dependency for REST API"
go get github.com/gin-gonic/gin
go get github.com/stianeikeland/go-rpio