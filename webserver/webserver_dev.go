package webserver

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"strings"

	"github.com/brocaar/lorawan"
	socketio "github.com/zishang520/socket.io/socket"

	//	cnt "github.com/arslab/lwnsimulator/controllers"
	"github.com/arslab/lwnsimulator/codes"
	"github.com/arslab/lwnsimulator/simulator/components/device"
	dev "github.com/arslab/lwnsimulator/simulator/components/device"
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

func LinkedDevName(d *dev.Device) string {
	name := d.Info.Name
	if d.Info.Status.LinkedDev {
		name = fmt.Sprintf("*%s", d.Info.Name)
	}
	return name
}

type handle_cmdRespType interface{}

func handle_cmd(s socketio.Socket, cmd_msg socket.DevExecuteCmdInter) handle_cmdRespType {

	cmd := cmd_msg.GetCmd()
	devEUI := cmd_msg.GetDevEUI()
	ack := cmd_msg.GetAck()

	// acknowledge receiving the command and send back the data that was initially sent
	if ack {
		cmd_msg_tmp, err := json.Marshal(cmd_msg)
		var cmd_msg_str string
		if err != nil {
			cmd_msg_str = ""
		} else {
			cmd_msg_str = string(cmd_msg_tmp)
		}
		s.Emit(socket.DevEventAckCmd, socket.DevAckCmd{Cmd: cmd, Args: cmd_msg_str})
	}

	cmd_status := codes.DevCmdOK
	var cmd_resp handle_cmdRespType

	if !simulatorController.IsRunning() {
		cmd_status = codes.DevErrorSimulatorNotRunning
	}

	d, err := getDeviceWithDevEUI(devEUI)
	if err != nil {
		cmd_status = codes.DevErrorNoDeviceWithDevEUI
	}

	switch cmd {
	case socket.DevCmdLinkDev:
		if cmd_status == codes.DevCmdOK {
			cmd_status, _ = simulatorController.DevExecuteLinkDev(&s, d.Id)
		}
		cmd_resp = socket.DevResponseCmd{Cmd: socket.DevCmdLinkDev, Error: cmd_status}
	case socket.DevCmdUnlinkDev:
		if cmd_status == codes.DevCmdOK {
			cmd_status, _ = simulatorController.DevExecuteUnlinkDev(&s, d.Id)
		}
		cmd_resp = socket.DevResponseCmd{Cmd: socket.DevCmdUnlinkDev, Error: cmd_status}
	case socket.DevCmdJoinRequest:
		if cmd_status == codes.DevCmdOK {
			cmd_status, _ = simulatorController.DevExecuteJoinRequest(d.Id)
		}
		cmd_resp = socket.DevNoResponseCmd{}
	case socket.DevCmdSendUplink:
		if cmd_status == codes.DevCmdOK {
			dat := cmd_msg.(socket.DevExecuteSendUplink)
			cmd_status, _ = simulatorController.DevExecuteSendUplink(d.Id, dat.MType, dat.Payload)
		}
		cmd_resp = socket.DevNoResponseCmd{}
	case socket.DevCmdRecvDownlink:
		if cmd_status == codes.DevCmdOK {
			dat := cmd_msg.(socket.DevExecuteRecvDownlink)
			cmd_status_, mtype, payload, _ := simulatorController.DevExecuteRecvDownlink(d.Id, dat.BufferSize)
			cmd_status = cmd_status_
			cmd_resp = socket.DevResponseRecvDownlinkCmd{Cmd: socket.DevCmdRecvDownlink, MType: mtype, Payload: payload, Error: cmd_status}
		} else {
			cmd_resp = socket.DevResponseRecvDownlinkCmd{Cmd: socket.DevCmdRecvDownlink, MType: "", Payload: "", Error: cmd_status}
		}
	}

	log_cmd_error(cmd, devEUI, cmd_status)
	return cmd_resp
}

func log_cmd_error(cmd string, devEUI string, status int) {
	var error_str string

	if status == codes.DevCmdOK {
		return
	}
	switch status {
	case codes.DevCmdTimeout:
		error_str = ""
	case codes.DevErrorNoDeviceWithDevEUI:
		error_str = "no device with devEUI"
	case codes.DevErrorNIY:
		error_str = "not implemented yet"
	case codes.DevErrorDeviceNotLinked:
		error_str = "device not linked"
	case codes.DevErrorDeviceTurnedOFF:
		error_str = "device turned off"
	case codes.DevErrorDeviceNotJoined:
		error_str = "device not joined"
	case codes.DevErrorDeviceAlreadyJoined:
		error_str = "device already joined (rejoining not implemented)"
	case codes.DevErrorRecvBufferEmpty:
		error_str = "downlink receive buffer empty"
	case codes.DevErrorSimulatorNotRunning:
		error_str = "simulator not running"
	}

	log.Println(fmt.Sprintf("[DEV %s][CMD %s][ERROR]  %s", devEUI, cmd, error_str))
}

func setupDevEventHandler(serverSocket *socketio.Server) {

	serverSocket.Of(
		"/dev",
		nil,
	).On("connection", func(clients ...interface{}) {
		s := clients[0].(*socketio.Socket)

		log.Println(fmt.Sprintf("[SocketIO][ns=/dev][id=%s]: connected", s.Id()))
		log.Println(fmt.Sprintf("      Remote_header %s", (*s).Client().Request().Headers()))

		s.On("disconnect", func(clients ...interface{}) {
			// func(s socketio.Socket, reason string)
			reason := strings.Trim(clients[0].(string), " ")
			log.Println(fmt.Sprintf("[SocketIO][ns=/dev][id=%s] disconnected : %s", s.Id(), reason))
			switch reason {
			case "client namespace disconnect":
				log.Println("\thandlers activated : DevUnjoin() and DevDeleteSocket()")
				if simulatorController.IsRunning() {
					simulatorController.DevUnjoin(string(s.Id()))
					simulatorController.DevDeleteSocket(string(s.Id()))
				}
			case "transport error":
				log.Println("\thandlers activated : DevDeleteSocket()")
				simulatorController.DevDeleteSocket(string(s.Id()))
			}
		})

		s.On(socket.DevEventLinkDev, func(clients ...interface{}) {
			//func(s socketio.Socket, cmd_msg socket.DevExecuteCmd) socket.DevResponseCmd {
			// resp := handle_cmd(s, cmd_msg)
			//	return resp.(socket.DevResponseCmd)		}
			//			cmd_msg := clients[0].(socket.DevExecuteCmd)
			tmp := (clients[0].(map[string]interface{}))
			cmd_msg := socket.DevExecuteCmd{
				Cmd:    tmp["Cmd"].(string),
				Ack:    tmp["Ack"].(bool),
				DevEUI: tmp["DevEUI"].(string),
			}
			resp := handle_cmd(*s, cmd_msg)
			/*			resp_tmp, err := json.Marshal(resp.(socket.DevResponseCmd))
						 			var resp_str string
									if err != nil {
										resp_str = ""
									} else {
										resp_str = string(resp_tmp)
									}
									clients[1].(func(...any))(resp_str)*/
			var resp_map map[string]interface{} = make(map[string]interface{})
			resp_map["cmd"] = resp.(socket.DevResponseCmd).Cmd
			resp_map["error"] = resp.(socket.DevResponseCmd).Error

			clients[1].(func(...any))(resp_map)
		})

		s.On(socket.DevEventUnlinkDev, func(clients ...interface{}) {
			//func(s socketio.Socket, cmd_msg socket.DevExecuteCmd) socket.DevResponseCmd {
			//			cmd_msg := clients[0].(socket.DevExecuteCmd)
			tmp := (clients[0].(map[string]interface{}))
			cmd_msg := socket.DevExecuteCmd{
				Cmd:    tmp["Cmd"].(string),
				Ack:    tmp["Ack"].(bool),
				DevEUI: tmp["DevEUI"].(string),
			}
			resp := handle_cmd(*s, cmd_msg)
			/* 			resp_tmp, err := json.Marshal(resp.(socket.DevResponseCmd))
			   			var resp_str string
			   			if err != nil {
			   				resp_str = ""
			   			} else {
			   				resp_str = string(resp_tmp)
			   			} */
			clients[1].(func(...any))(resp)

		})

		s.On(socket.DevEventJoinRequest, func(clients ...interface{}) {
			//func(s socketio.Socket, cmd_msg socket.DevExecuteCmd) {
			// cmd_msg := clients[0].(socket.DevExecuteCmd)
			tmp := (clients[0].(map[string]interface{}))
			cmd_msg := socket.DevExecuteCmd{
				Cmd:    tmp["Cmd"].(string),
				Ack:    tmp["Ack"].(bool),
				DevEUI: tmp["DevEUI"].(string),
			}

			handle_cmd(*s, cmd_msg)
			//	return

		})

		s.On(socket.DevEventSendUplink, func(clients ...interface{}) {
			// func(s socketio.Socket, cmd_msg socket.DevExecuteSendUplink) {
			//			cmd_msg := clients[0].(socket.DevExecuteSendUplink)
			tmp := (clients[0].(map[string]interface{}))
			cmd_msg := socket.DevExecuteSendUplink{
				Cmd:     tmp["Cmd"].(string),
				Ack:     tmp["Ack"].(bool),
				DevEUI:  tmp["DevEUI"].(string),
				MType:   tmp["MType"].(string),
				Payload: tmp["Payload"].(string),
			}
			handle_cmd(*s, cmd_msg)
			//	return
		})

		s.On(socket.DevEventRecvDownlink, func(clients ...interface{}) {
			//func(s socketio.Socket, cmd_msg socket.DevExecuteRecvDownlink) socket.DevResponseRecvDownlink {
			//return resp.(socket.DevResponseRecvDownlink)
			//			cmd_msg := clients[0].(socket.DevExecuteRecvDownlink)
			tmp := (clients[0].(map[string]interface{}))
			cmd_msg := socket.DevExecuteRecvDownlink{
				Cmd:        tmp["Cmd"].(string),
				Ack:        tmp["Ack"].(bool),
				DevEUI:     tmp["DevEUI"].(string),
				BufferSize: int(tmp["BufferSize"].(float64)),
			}
			resp := handle_cmd(*s, cmd_msg)
			/*
				resp_tmp, err := json.Marshal(resp.(socket.DevResponseRecvDownlink))
				var resp_str string
				if err != nil {
					resp_str = ""
				} else {
					resp_str = string(resp_tmp)
				} */
			clients[1].(func(...any))(resp)

		})
	})
}

// Obsolete
/*
func setupDevEventHandler_v3(serverSocket *socketio.Server) {

	serverSocket.OnConnect("/dev", func(s socketio.Socket) error {

		log.Println(fmt.Sprintf("[DevWS]: DevSocket %s connected", s.ID()))
		log.Println(fmt.Sprintf("[DevWS]: Remote_header %s", s.RemoteHeader()))

		return nil

	})

	serverSocket.OnDisconnect("/dev", func(s socketio.Socket, reason string) {

		log.Println(fmt.Sprintf("[DevWS] DevSocket %s disconnected : %s", s.ID(), reason))
		simulatorController.DevDeleteSocket(s.ID())

		return

	})

	serverSocket.OnEvent("/dev", socket.DevEventLinkDev, func(s socketio.Socket, cmd_msg socket.DevExecuteCmd) socket.DevResponseCmd {

		resp := handle_cmd(s, cmd_msg)
		return resp.(socket.DevResponseCmd)

	})

	serverSocket.OnEvent("/dev", socket.DevEventUnlinkDev, func(s socketio.Socket, cmd_msg socket.DevExecuteCmd) socket.DevResponseCmd {

		resp := handle_cmd(s, cmd_msg)
		return resp.(socket.DevResponseCmd)

	})

	serverSocket.OnEvent("/dev", socket.DevEventJoinRequest, func(s socketio.Socket, cmd_msg socket.DevExecuteCmd) {

		handle_cmd(s, cmd_msg)
		return

	})

	serverSocket.OnEvent("/dev", socket.DevEventSendUplink, func(s socketio.Socket, cmd_msg socket.DevExecuteSendUplink) {

		handle_cmd(s, cmd_msg)
		return
	})

	serverSocket.OnEvent("/dev", socket.DevEventRecvDownlink, func(s socketio.Socket, cmd_msg socket.DevExecuteRecvDownlink) socket.DevResponseRecvDownlink {

		resp := handle_cmd(s, cmd_msg)
		return resp.(socket.DevResponseRecvDownlink)
	})
}
*/
