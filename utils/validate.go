package utils

import (
	"errors"
	"fmt"
	"io/ioutil"
	"net"
	"os/user"
	"regexp"
	"runtime"
	"syscall"

	"github.com/dorzheh/infra/utils/sysutils"
)

func ValidateUserName(userName string) error {
	userInfo, err := user.Current()
	if err != nil {
		return err
	}
	if userInfo.Name != userName {
		return errors.New(fmt.Sprintf("You are not \"%s\"", userName))
	}
	return nil
}

func ValidateUserID(id int) error {
	if syscall.Getuid() != id {
		return errors.New(fmt.Sprintf("Your ID is not \"%d\"", id))
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
	installedRamInMb, err := sysutils.SysinfoRam()
	if err != nil {
		return err
	}
	if minRequiredInMb > installedRamInMb {
		return errors.New(fmt.Sprintf("RAM amount.Required %d.Installed %d",
			minRequiredInMb, installedRamInMb))
	}
	return nil
}

//hostname max length is 255 chars
const hostnameLength = 256

// hostname: names separated by '.' every name must be max 63 chars in length
// according to RFC 952 <name>  ::= <let>[*[<let-or-digit-or-hyphen>]<let-or-digit>]
// according to RFC 1123 - trailing and starting digits are allowed
var hostnameExpr = regexp.MustCompile(`^([a-zA-Z]|[0-9]|_)(([a-zA-Z]|[0-9]|-|_)*([a-zA-Z]|[0-9]|_))?(\.([a-zA-Z]|[0-9]|_)(([a-zA-Z]|[0-9]|-|_)*([a-zA-Z]|[0-9]))?)*$`)
var hostnameLenExpr = regexp.MustCompile(`^[a-zA-Z0-9-_]{1,64}(\.([a-zA-Z0-9-_]{1,64}))*$`)
var ipAddrExpr = regexp.MustCompile(`^[0-9]+\.[0-9]+\.[0-9]+\.[0-9]+$`)

func ValidateHostname(hostname string) error {
	switch {
	case len(hostname) > hostnameLength:
		return fmt.Errorf("hostname length is greater than %d", hostnameLength)
	case !hostnameExpr.MatchString(hostname):
		return fmt.Errorf("hostname doesn't match RFC")
	case !hostnameLenExpr.MatchString(hostname):
		return fmt.Errorf("hostname length")
	case ipAddrExpr.MatchString(hostname):
		return fmt.Errorf("hostname cannot be like an ip")
	}
	return nil
}
