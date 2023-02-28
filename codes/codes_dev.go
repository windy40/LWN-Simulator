package codes

const (
	DevCmdOK                   = iota
	DevCmdTimeout              // never occurs
	DevErrorNoDeviceWithDevEUI // simulator state
	DevErrorNIY
	DevErrorDeviceNotLinked     // inappropriate state error
	DevErrorDeviceLinked        // state error, never occurs
	DevErrorDeviceNotJoined     //state error
	DevErrorDeviceAlreadyJoined //state error
	DevErrorRecvBufferEmpty     //  lora simulation state
	DevErrorSimulatorNotRunning // simulator state
)
