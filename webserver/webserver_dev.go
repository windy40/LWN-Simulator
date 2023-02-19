package webserver

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"

	"github.com/brocaar/lorawan"
	socketio "github.com/googollee/go-socket.io"

	//	cnt "github.com/arslab/lwnsimulator/controllers"
	"github.com/arslab/lwnsimulator/codes"
	"github.com/arslab/lwnsimulator/simulator/components/device"
	"github.com/arslab/lwnsimulator/socket"
)

func getDeviceWithDevEUI(devEUIstr string) (dev *device.Device, err error) {
	devices := simulatorController.GetDevices()

	var devEUI lorawan.EUI64
	devEUI.UnmarshalText([]byte(devEUIstr))

	// get device Id for devEUI device if it exists
	for _, d := range devices {

		if d.Info.DevEUI == devEUI {
			return &d, nil
		}
	}
	err = errors.New("No device with given devEUI")
	return nil, err
}

func setupDevEventHandler(serverSocket *socketio.Server) {

	serverSocket.OnConnect("/dev", func(s socketio.Conn) error {

		log.Println(fmt.Sprintf("[DevWS]: DevSocket %s connected", s.ID()))
		log.Println(fmt.Sprintf("[DevWS]: Remote_header %s", s.RemoteHeader()))

		return nil

	})

	serverSocket.OnDisconnect("/dev", func(s socketio.Conn, reason string) {

		log.Println(fmt.Sprintf("[DevWS] DevSocket %s disconnected : %s", s.ID(), reason))
		simulatorController.DeleteDevSocket(s.ID())

		return

	})

	serverSocket.OnEvent("/dev", socket.DevEventLinkDev, func(s socketio.Conn, data socket.DevExecuteCmd) (error, string) {

		handle_cmd(s, data)
		return nil, "linked"

	})

	serverSocket.OnEvent("/dev", socket.DevEventUnlinkDev, func(s socketio.Conn, data socket.DevExecuteCmd) error {

		handle_cmd(s, data)
		return nil

	})

	serverSocket.OnEvent("/dev", socket.DevEventJoinRequest, func(s socketio.Conn, data socket.DevExecuteCmd) error {

		handle_cmd(s, data)
		return nil

	})

	serverSocket.OnEvent("/dev", socket.DevEventSendUplink, func(s socketio.Conn, data socket.DevExecuteSendUplink) error {

		handle_cmd(s, data)
		return nil
	})

	serverSocket.OnEvent("/dev", socket.DevEventRecvDownlink, func(s socketio.Conn, data socket.DevExecuteRecvDownlink) error {

		handle_cmd(s, data)
		return nil
	})
}

func handle_cmd(s socketio.Conn, data socket.DevExecuteCmdInter) {

	cmd := data.GetCmd()
	devEUI := data.GetDevEUI()
	ack := data.GetAck()

	if ack {
		data_tmp, err := json.Marshal(data)
		var data_str string
		if err != nil {
			data_str = ""
		} else {
			data_str = string(data_tmp)
		}
		s.Emit(socket.DevEventAckCmd, socket.DevAckCmd{Cmd: cmd, Args: data_str})
	}

	d, err := getDeviceWithDevEUI(devEUI)

	if err != nil {
		log.Println(fmt.Sprintf("DEV[???][CMD %s]Error could not find dev with devEUI %s", cmd, devEUI))
		s.Emit(socket.DevEventResponseCmd, socket.DevResponseCmd{Cmd: cmd, Error: codes.DevErrorNoDeviceWithDevEUI})
	}

	switch cmd {
	case socket.DevCmdLinkDev:
		simulatorController.DevExecuteLinkDev(&s, d.Id)
	case socket.DevCmdUnlinkDev:
		simulatorController.DevExecuteUnlinkDev(&s, d.Id)
	case socket.DevCmdJoinRequest:
		simulatorController.DevExecuteJoinRequest(d.Id)
	case socket.DevCmdSendUplink:
		dat := data.(socket.DevExecuteSendUplink)
		simulatorController.DevExecuteSendUplink(d.Id, dat)
	case socket.DevCmdRecvDownlink:
		dat := data.(socket.DevExecuteRecvDownlink)
		simulatorController.DevExecuteRecvDownlink(d.Id, dat)

	}

	return

}
