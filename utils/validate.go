// Verification functions

package utils

import (
	"errors"
	"fmt"
	"io/ioutil"
	"net"
	"os/user"
	"regexp"
	"runtime"
)

func ValidateUser(userName string) error {
	userInfo, err := user.Current()
	if err != nil {
		return err
	}
	if userInfo.Name != userName {
		return errors.New(fmt.Sprintf("You are \"%s\",only \"%s\" is permitted to run it.",
			userInfo.Username, userName))
	}
	return nil
}

// ValidateDistro validates distribution
func ValidateDistro(regExp, file string) error {
	buf, err := ioutil.ReadFile(file)
	if err == nil {
		ok, err := regexp.Match(regExp, buf)
		if ok && err == nil {
			return nil
		}
	}
	return errors.New("unsupported Linux distribution")
}

// ValidateNics gets a reference to a string slice containing
// NICs names and verifies if appropriate NIC is available
func ValidateNics(nicList []string) error {
	for _, iface := range nicList {
		if _, err := net.InterfaceByName(iface); err != nil {
			return err
		}
	}
	return nil
}

// ValidateAmountOfCpus gets amount of CPUs required
// It verifies that amount of installed CPUs is equal
// to amount of required
func ValidateAmountOfCpus(required int) error {
	amountOfInstalledCpus := runtime.NumCPU()
	if amountOfInstalledCpus != required {
		return errors.New(fmt.Sprintf("CPU amount.Required %d.Installed %d",
			required, amountOfInstalledCpus))
	}
	return nil
}

// ValidateAmountOfRam gets amount of RAM needed for proceeding
// The function verifies if eligable amount RAM is installed
func ValidateAmountOfRam(minRequiredInMb int) error {
	installedRamInMb, err := SysinfoRam()
	if err != nil {
		return err
	}
	if minRequiredInMb > installedRamInMb {
		return errors.New(fmt.Sprintf("RAM amount.Required %d.Installed %d",
			minRequiredInMb, installedRamInMb))
	}
	return nil
}
