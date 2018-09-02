## wsnet
Simple websocket framework in golang


### Install
`go get -u github.com/rezam90/wsnet`

### using [gin](https://github.com/gin-gonic/gin)
```go
package main

import(
    "github.com/gin-gonic/gin"
    "github.com/rezam90/wsnet"
)

func main(){
    r := gin.Default()
    ws := wsnet.New()

    r.GET("/ws", func(c *gin.Context){
        ws.HandleRequest(c.Writer, c.Request)
    })

    ws.HandleConnect(func(conn *wsnet.Connection){
        fmt.Printf("client %d connected\n", conn.ID())
    })

    ws.HandleMessage(func(conn *wsnet.Connection, data []byte) error {
        fmt.Printf("client %d sent %s\n", conn.ID(), string(data))
        return nil
    })

    ws.HandleClose(func(conn *wsnet.Connection){
        fmt.Printf("client %d closed\n", conn.ID())
    })
}
```

> this package is highly inspired by melody framework