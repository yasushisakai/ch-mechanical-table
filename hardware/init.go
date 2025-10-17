package hardware

import (
	"errors"
	"fmt"
	"go.bug.st/serial/enumerator"
	"log"
)

type PortNames struct {
	Slider string
	Button string
}

func ListPortDetails() error {
	portsDetails, err := enumerator.GetDetailedPortsList()

	if err != nil {
		return err
	}

	if len(portsDetails) == 0 {
		log.Printf("no device found")

		return nil
	}

	count := 0

	for _, pd := range portsDetails {

		if pd.IsUSB {
			count += 1
			log.Printf("%#v", pd)
		}
	}

	if count == 0 {
		log.Printf("no usb port device found")
	}

	return nil
}

func PortLookup() (PortNames, error) {

	portsDetails, err := enumerator.GetDetailedPortsList()

	if err != nil {
		return PortNames{}, err
	}

	if len(portsDetails) == 0 {
		return PortNames{}, errors.New("no ports found.")
	}

	var portNames PortNames

	for _, pd := range portsDetails {
		if pd.IsUSB {
			portIdentifier := fmt.Sprintf("%s%s%s", pd.VID, pd.PID, pd.SerialNumber)
			if portIdentifier == fmt.Sprintf("%s%s%s", SLIDER_VID, SLIDER_PID, SLIDER_SERIAL_NUMBER) {
				log.Println("found slider")
				portNames.Slider = pd.Name
			} else if portIdentifier == fmt.Sprintf("%s%s", BUTTON_VID, BUTTON_PID) {
				log.Println("likely button found")
				portNames.Button = pd.Name
			}
		}
	}

	if portNames.Slider == "" {
		log.Printf("warning: slider is missing")
	}

	if portNames.Button == "" {
		log.Printf("warning: button is missing")
	}

	return portNames, nil

}
