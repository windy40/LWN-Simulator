package device

import (
	"github.com/arslab/lwnsimulator/socket"
)

func (d *Device) ReturnLoraEvent(ev int) {

	if _, ok := d.Resources.LinkedDevSocket[d.Id]; ok {
		(d.Resources.LinkedDevSocket[d.Id]).Emit(socket.DevEventLoRa, socket.DevLoRaEvent{Event: ev})

	}
}
