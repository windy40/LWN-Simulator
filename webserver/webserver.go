package webserver

import (
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"

	cnt "github.com/arslab/lwnsimulator/controllers"
	"github.com/arslab/lwnsimulator/models"
	dev "github.com/arslab/lwnsimulator/simulator/components/device"
	rp "github.com/arslab/lwnsimulator/simulator/components/device/regional_parameters"
	gw "github.com/arslab/lwnsimulator/simulator/components/gateway"
	"github.com/arslab/lwnsimulator/socket"
	_ "github.com/arslab/lwnsimulator/webserver/statik"
	"github.com/brocaar/lorawan"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/rakyll/statik/fs"
	socketio "github.com/zishang520/socket.io/socket"
)

// WebServer type
type WebServer struct {
	Address      string
	Port         int
	Router       *gin.Engine
	ServerSocket *socketio.Server
}

var (
	simulatorController cnt.SimulatorController
	configuration       *models.ServerConfig
)

func NewWebServer(config *models.ServerConfig, controller cnt.SimulatorController) *WebServer {

	var serverSocket *socketio.Server

	configuration = config
	simulatorController = controller

	serverSocket = newServerSocket()

	/* 	go func() {

		// v1		err := serverSocket.Serve()
		serverSocket = newServerSocket()

		// 		if err != nil {
		// 	log.Fatal(err)
		// }

	}() */

	gin.SetMode(gin.ReleaseMode)
	router := gin.New()

	configCors := cors.DefaultConfig()
	configCors.AllowAllOrigins = true
	configCors.AllowHeaders = []string{"Origin", "Access-Control-Allow-Origin",
		"Access-Control-Allow-Headers", "Content-type"}
	configCors.AllowMethods = []string{"GET", "POST", "DELETE", "OPTIONS"}
	configCors.AllowCredentials = true
	router.Use(cors.New(configCors))

	router.Use(gin.Recovery())

	ws := WebServer{
		Address:      configuration.Address,
		Port:         configuration.Port,
		Router:       router,
		ServerSocket: serverSocket,
	}

	staticFS, err := fs.New()
	if err != nil {
		log.Fatal(err)
	}

	staticGroup := router.Group("/dashboard")
	staticGroup.StaticFS("/", staticFS)
	//router.Use(static.Serve("/", staticFS))

	apiRoutes := router.Group("/api")
	{
		apiRoutes.GET("/start", startSimulator)
		apiRoutes.GET("/stop", stopSimulator)
		apiRoutes.GET("/bridge", getRemoteAddress)
		apiRoutes.GET("/gateways", getGateways)
		apiRoutes.GET("/devices", getDevices)
		apiRoutes.POST("/add-device", addDevice)
		apiRoutes.POST("/up-device", updateDevice)
		apiRoutes.POST("/del-device", deleteDevice)
		apiRoutes.POST("/del-gateway", deleteGateway)
		apiRoutes.POST("/add-gateway", addGateway)
		apiRoutes.POST("/up-gateway", updateGateway)
		apiRoutes.POST("/bridge/save", saveInfoBridge)

	}

	router.GET("/socket.io/*any", gin.WrapH(serverSocket.ServeHandler(nil)))
	router.POST("/socket.io/*any", gin.WrapH(serverSocket.ServeHandler(nil)))

	router.GET("/", func(context *gin.Context) { context.Redirect(http.StatusMovedPermanently, "/dashboard") })

	return &ws
}

func startSimulator(c *gin.Context) {
	c.JSON(http.StatusOK, simulatorController.Run())
}

func stopSimulator(c *gin.Context) {
	c.JSON(http.StatusOK, simulatorController.Stop())
}

func saveInfoBridge(c *gin.Context) {

	var ns models.AddressIP
	c.BindJSON(&ns)

	c.JSON(http.StatusOK, gin.H{"status": simulatorController.SaveBridgeAddress(ns)})
}

func getRemoteAddress(c *gin.Context) {
	c.JSON(http.StatusOK, simulatorController.GetBridgeAddress())
}

func getGateways(c *gin.Context) {

	gws := simulatorController.GetGateways()
	c.JSON(http.StatusOK, gws)
}

func addGateway(c *gin.Context) {

	var g gw.Gateway
	c.BindJSON(&g)

	code, id, err := simulatorController.AddGateway(&g)
	errString := fmt.Sprintf("%v", err)

	c.JSON(http.StatusOK, gin.H{"status": errString, "code": code, "id": id})

}

func updateGateway(c *gin.Context) {

	var g gw.Gateway
	c.BindJSON(&g)

	code, err := simulatorController.UpdateGateway(&g)
	errString := fmt.Sprintf("%v", err)

	c.JSON(http.StatusOK, gin.H{"status": errString, "code": code})

}

func deleteGateway(c *gin.Context) {

	Identifier := struct {
		Id int `json:"id"`
	}{}

	c.BindJSON(&Identifier)

	c.JSON(http.StatusOK, gin.H{"status": simulatorController.DeleteGateway(Identifier.Id)})

}

func getDevices(c *gin.Context) {
	c.JSON(http.StatusOK, simulatorController.GetDevices())
}

func addDevice(c *gin.Context) {

	var device dev.Device
	c.BindJSON(&device)

	code, id, err := simulatorController.AddDevice(&device)
	errString := fmt.Sprintf("%v", err)

	c.JSON(http.StatusOK, gin.H{"status": errString, "code": code, "id": id})

}

func updateDevice(c *gin.Context) {

	var device dev.Device
	c.BindJSON(&device)

	code, err := simulatorController.UpdateDevice(&device)
	errString := fmt.Sprintf("%v", err)

	c.JSON(http.StatusOK, gin.H{"status": errString, "code": code})

}

func deleteDevice(c *gin.Context) {

	Identifier := struct {
		Id int `json:"id"`
	}{}

	c.BindJSON(&Identifier)

	c.JSON(http.StatusOK, gin.H{"status": simulatorController.DeleteDevice(Identifier.Id)})
}

/*
func newServerSocket_v3() *socketio.Server {

	serverSocket := socketio.NewServer(nil, nil)

	serverSocket.OnConnect("/", func(s socketio.Socket) error {
		// windy40 dev socket : distinguish connection from web interface and from device
		remote_hdr := s.RemoteHeader()
		v, ok := remote_hdr["User-Agent"]

		if ok && strings.Contains(v[0], "Mozilla") {

			log.Println("[WS]: Socket connected\n")

			s.SetContext("")
			simulatorController.AddWebSocket(&s)
		}

		return nil

	})

	serverSocket.OnDisconnect("/", func(s socketio.Socket, reason string) {
		s.Close()
	})

	serverSocket.OnEvent("/", socket.EventToggleStateDevice, func(s socketio.Socket, Id int) {
		simulatorController.ToggleStateDevice(Id)
	})

	serverSocket.OnEvent("/", socket.EventToggleStateGateway, func(s socketio.Socket, Id int) {
		simulatorController.ToggleStateGateway(Id)
	})

	serverSocket.OnEvent("/", socket.EventMacCommand, func(s socketio.Socket, data socket.MacCommand) {

		switch data.CID {
		case "DeviceTimeReq":
			simulatorController.SendMACCommand(lorawan.DeviceTimeReq, data)
		case "LinkCheckReq":
			simulatorController.SendMACCommand(lorawan.LinkCheckReq, data)
		case "PingSlotInfoReq":
			simulatorController.SendMACCommand(lorawan.PingSlotInfoReq, data)

		}

	})

	serverSocket.OnEvent("/", socket.EventChangePayload, func(s socketio.Socket, data socket.NewPayload) (string, bool) {
		return simulatorController.ChangePayload(data)
	})

	serverSocket.OnEvent("/", socket.EventSendUplink, func(s socketio.Socket, data socket.NewPayload) {
		simulatorController.SendUplink(data)
	})

	serverSocket.OnEvent("/", socket.EventGetParameters, func(s socketio.Socket, code int) mrp.Informations {
		return rp.GetInfo(code)
	})

	serverSocket.OnEvent("/", socket.EventChangeLocation, func(s socketio.Socket, info socket.NewLocation) bool {
		return simulatorController.ChangeLocation(info)
	})

	setupDevEventHandler(serverSocket)

	return serverSocket
}
*/

func newServerSocket() *socketio.Server {

	serverSocket := socketio.NewServer(nil, nil)

	serverSocket.Of(
		"/",
		nil,
	).On("connection", func(clients ...interface{}) {
		s := clients[0].(*socketio.Socket)

		// windy40 dev socket : differentiate connection from web interface and from device
		// if simulatorController.
		remote_hdr := (*s).Client().Request().Headers() // RemoteHeader()
		v, ok := remote_hdr.Gets("User-Agent")
		web_ui_sio := ok && strings.Contains(v[0], "Mozilla")
		if web_ui_sio { // from Web UI
			log.Println(fmt.Sprintf("[SocketIo][ns=/][web UI][id=%s]: connected", s.Id()))
			log.Println(fmt.Sprintf("        Remote_header %s", (*s).Client().Request().Headers()))

			// v1 s.SetContext("")
			simulatorController.AddWebSocket(s)
		} else {
			log.Println(fmt.Sprintf("[SocketIO][ns=/][Dev][id=%s]: connected", s.Id()))
			log.Println(fmt.Sprintf("        Remote_header %s", (*s).Client().Request().Headers()))
		}

		s.On("disconnect", func(...interface{}) {
			remote_hdr := (*s).Client().Request().Headers() // RemoteHeader()
			v, ok := remote_hdr.Gets("User-Agent")
			web_ui_sio := ok && strings.Contains(v[0], "Mozilla")
			//			reason := clients[0].(string)
			if web_ui_sio {
				log.Println(fmt.Sprintf("[SocketIo][ns=/][Web UI][id=%s] disconnected", s.Id()))
			} else {
				log.Println(fmt.Sprintf("[SocketIo][ns=/ ][Dev][id=%s] disconnected ", s.Id()))

			}

		})
		s.On(socket.EventToggleStateDevice, func(clients ...interface{}) {
			Id := clients[0].(int)
			simulatorController.ToggleStateDevice(Id)
		})

		s.On(socket.EventToggleStateGateway, func(clients ...interface{}) {
			// func(s socketio.Socket, Id int)
			Id := int(clients[0].(float64)) // why float64 ?
			simulatorController.ToggleStateGateway(Id)
		})

		s.On(socket.EventMacCommand, func(clients ...interface{}) {
			//func(s socketio.Socket, data socket.MacCommand)
			data := clients[0].(socket.MacCommand)
			switch data.CID {
			case "DeviceTimeReq":
				simulatorController.SendMACCommand(lorawan.DeviceTimeReq, data)
			case "LinkCheckReq":
				simulatorController.SendMACCommand(lorawan.LinkCheckReq, data)
			case "PingSlotInfoReq":
				simulatorController.SendMACCommand(lorawan.PingSlotInfoReq, data)

			}

		})

		s.On(socket.EventChangePayload, func(clients ...interface{}) {
			//func(s socketio.Socket, data socket.NewPayload) (string, bool)
			data := clients[0].(socket.NewPayload)
			clients[1].(func(...any))(simulatorController.ChangePayload(data))
		})

		s.On(socket.EventSendUplink, func(clients ...interface{}) {
			//func(s socketio.Socket, data socket.NewPayload)
			data := clients[0].(socket.NewPayload)
			simulatorController.SendUplink(data)
		})

		s.On(socket.EventGetParameters, func(clients ...interface{}) {
			// func(s socketio.Socket, code int) mrp.Informations
			// return rp.GetInfo(code)
			code := int(clients[0].(float64))
			info := rp.GetInfo(code)
			clients[1].(func(...any))(info)

		})

		s.On(socket.EventChangeLocation, func(clients ...interface{}) {
			// func(s socketio.Socket, info socket.NewLocation) bool
			info := clients[0].(socket.NewLocation)
			status := simulatorController.ChangeLocation(info)
			clients[1].(func(...any))(status)
		})

	})
	setupDevEventHandler(serverSocket)
	return serverSocket
}

func (ws *WebServer) Run() {

	log.Println("[WS]: Listen [", ws.Address+":"+strconv.Itoa(ws.Port), "]")

	err := ws.Router.Run(ws.Address + ":" + strconv.Itoa(ws.Port))
	if err != nil {
		log.Println("[WS] [ERROR]:", err.Error())
	}

}
