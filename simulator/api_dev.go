package simulator

import (
	"errors"

	"github.com/arslab/lwnsimulator/codes"
	"github.com/brocaar/lorawan"

	// dev "github.com/arslab/lwnsimulator/simulator/components/device"
	"github.com/arslab/lwnsimulator/simulator/util"
	socketio "github.com/zishang520/socket.io/socket"
)

/* func (s *Simulator) getDeviceWithDevEUI(devEUIstr string) (dev *dev.Device, err error) {
	devices := s.Devices

	var devEUI lorawan.EUI64
	devEUI.UnmarshalText([]byte(devEUIstr))

	// get device Id for devEUI device if it exists
	for _, d := range devices {

		if d.Info.DevEUI == devEUI {
			return d, nil
		}
	}
	err = errors.New("No device with given devEUI")
	return nil, err
} */

func (s *Simulator) DevExecuteLinkDev(DevSocket *socketio.Socket, Id int) (int, error) {
	d := s.Devices[Id]

	s.Resources.DevAddSocket(DevSocket, Id)
	d.Info.Status.LinkedDev = true
	d.Print("Linked to external MCU", nil, util.PrintBoth)

	return codes.DevCmdOK, nil
}

func (s *Simulator) DevExecuteUnlinkDev(DevSocket *socketio.Socket, Id int) (int, error) {
	d := s.Devices[Id]

	if !d.Info.Status.LinkedDev {
		return codes.DevErrorDeviceNotLinked, errors.New("device not linked")
	}
	s.Resources.DevDeleteSocket(string(DevSocket.Id()))
	// if unlink, keep join status as is
	/* 	if d.Info.Status.Joined {
		d.UnJoined()
		log.Println(fmt.Sprintf("DEV[%s] unjoined", LinkedDevName(d)))
	} */

	d.Info.Status.LinkedDev = false
	d.Print("Linked from external MCU", nil, util.PrintBoth)

	return codes.DevCmdOK, nil
}

func (s *Simulator) DevExecuteJoinRequest(Id int) (int, error) {
	d := s.Devices[Id]

	// simulation is supposed to be running therefor dev is turned ON
	/*
		// if device si turned OFF return error
		if !d.IsOn() {
			log.Println(fmt.Sprintf("DEV[%s][CMD %s] device turned off", LinkedDevName(d), socket.DevCmdJoinRequest))
			s.Resources.LinkedDevSocket[d.Id].Emit(socket.DevEventResponseCmd, socket.DevResponseCmd{Cmd: socket.DevCmdJoinRequest, Error: codes.DevErrorDeviceTurnedOFF})
			return
		} */

	// at present if dev joined don't join again ...
	// but then rejoining not possible unless dev gets previously unjoined
	if d.Info.Status.Joined {
		return codes.DevErrorDeviceAlreadyJoined, errors.New("device already joined (rejoining not implemented)")
	}

	//	go d.DevJoinAndProcessUplink()
	d.OtaaActivate <- struct{}{}

	return codes.DevCmdOK, nil
}

func (s *Simulator) DevExecuteSendUplink(Id int, mt string, pl string) (int, error) {
	d := s.Devices[Id]
	mtype := mt
	payload := pl

	// simulation is supposed to be running therefor dev is turned ON
	/* 	if !d.IsOn() {
		log.Println(fmt.Sprintf("DEV[%s][CMD %s] device turned off", LinkedDevName(d), data.Cmd))
		s.Resources.LinkedDevSocket[d.Id].Emit(socket.DevEventResponseCmd, socket.DevResponseCmd{Cmd: data.Cmd, Error: codes.DevErrorDeviceTurnedOFF})
		return
	} */

	if !d.Info.Status.Joined {
		return codes.DevErrorDeviceNotJoined, errors.New("device not joined")
	}

	MType := lorawan.UnconfirmedDataUp
	if mtype == "ConfirmedDataUp" {
		MType = lorawan.ConfirmedDataUp
	}

	d.NewUplink(MType, payload)
	d.UplinkWaiting <- struct{}{}

	return codes.DevCmdOK, nil
}

func (s *Simulator) DevExecuteRecvDownlink(Id int, buff_size int) (int, string, string, error) {

	d := s.Devices[Id]
	BufferSize := buff_size

	if !d.Info.Status.Joined {
		return codes.DevErrorDeviceNotJoined, "", "", errors.New("device not joined")
	}

	if len(d.Info.Status.BufferDataDownlinks) > 0 {
		payload := d.Info.Status.BufferDataDownlinks[0].DataPayload

		size := len(payload)
		if size > BufferSize {
			size = BufferSize
		}
		payload = payload[:size]

		mtype := "UnconfirmedDataDown"
		if d.Info.Status.BufferDataDownlinks[0].MType == lorawan.ConfirmedDataDown {
			mtype = "ConfirmedDataDown"
		}

		switch len(d.Info.Status.BufferDataDownlinks) {
		case 1:
			d.Info.Status.BufferDataDownlinks = d.Info.Status.BufferDataDownlinks[:0]

		default:
			d.Info.Status.BufferDataDownlinks = d.Info.Status.BufferDataDownlinks[1:]
		}

		return codes.DevCmdOK, mtype, string(payload), nil
	}

	return codes.DevErrorRecvBufferEmpty, "", "", errors.New("downlink receive buffer empty")

}

// called when DevSocket is disconnected because of transport error
func (s *Simulator) DevDeleteSocket(SId string) {
	s.Resources.DevDeleteSocket(SId)
}

func (s *Simulator) DevUnjoin(SId string) {
	Id, err := s.Resources.DevGetIdFromSocketID(SId)
	if err == nil {
		d := s.Devices[Id]
		d.UnJoined()
		d.Print("device unjoined", nil, util.PrintBoth)
	}
}
