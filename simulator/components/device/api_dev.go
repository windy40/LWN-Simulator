package device

import (
	"github.com/arslab/lwnsimulator/simulator/components/device/classes"
	"github.com/arslab/lwnsimulator/simulator/util"
	"github.com/arslab/lwnsimulator/socket"
)

// THIS FUNCTION IS OBSOLETE
func (d *Device) DevJoinAndProcessUplink() {

	//	defer d.Resources.ExitGroup.Done()

	d.Print("trying to join ...", nil, util.PrintBoth)
	d.OtaaActivation()

	if d.Info.Status.Joined {
		if d.Info.Status.LinkedDev {
			d.ReturnLoraEvent(socket.JOIN_ACCEPT_EVENT)
			go d.Run()
		} else {
			d.Print("Could not join", nil, util.PrintBoth)
			return
		}
	}

	/* 	for {

		select {

		case <-d.UplinkWaiting:
			break

		case <-d.Exit:
			d.Print("Turn OFF", nil, util.PrintBoth)
			return
		}

		if d.CanExecute() {

			if d.Info.Status.Joined {

				if d.Info.Configuration.SupportedClassC {
					d.SwitchClass(classes.ClassC)
				} else if d.Info.Configuration.SupportedClassB {
					d.SwitchClass(classes.ClassB)
				}

				d.Execute()

			} else {
				//				d.OtaaActivation()
			}

		}
	} */
}

func (d *Device) RunLinkableDev() {
	defer d.Resources.ExitGroup.Done()

	for {

		select {
		case <-d.OtaaActivate:
			if d.Info.Status.Joined {
				d.UnJoined()
			}
			d.Print("trying to join ...", nil, util.PrintBoth)
			d.OtaaActivation()
			if d.Info.Status.Joined {
				d.ReturnLoraEvent(socket.JOIN_ACCEPT_EVENT)
			} else {
				d.Print("Could not join", nil, util.PrintBoth)
				//				return
			}
		case <-d.UplinkWaiting:
			if d.CanExecute() {

				if d.Info.Status.Joined {

					if d.Info.Configuration.SupportedClassC {
						d.SwitchClass(classes.ClassC)
					} else if d.Info.Configuration.SupportedClassB {
						d.SwitchClass(classes.ClassB)
					}

					d.Execute()

				} else {
					d.Print("WARNING LinkedDev sending uplink while unjoined", nil, util.PrintBoth)
					/* 				d.OtaaActivation()
					   if d.Info.Status.Joined {
						   d.ReturnLoraEvent(socket.JOIN_ACCEPT_EVENT)
					   } */
					//				return
				}

			}

		case <-d.Exit:
			d.Print("Turn OFF", nil, util.PrintBoth)
			return
		}

	}

}
