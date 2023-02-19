package codes

const (
	DevCmdOK = iota
	DevCmdTimeout
	DevErrorNoDeviceWithDevEUI
	DevErrorNIY
	DevErrorDeviceNotLinked
	DevErrorDeviceTurnedOFF
	DevErrorDeviceNotJoined
	DevErrorDeviceAlreadyJoined
	DevErrorRecvBufferEmpty
	DevErrorSimulatorNotRunning
)
