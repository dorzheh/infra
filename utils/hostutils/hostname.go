package hostutils

import (
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"strings"
	"syscall"

	"github.com/dorzheh/infra/utils"
	"github.com/dorzheh/infra/utils/ioutils"
)

var template = map[uint8]map[string]interface{}{
	1: {
		"RELEASE_FILE":      "/etc/lsb_release",
		"HOST_FILE":         "/etc/hostname",
		"SET_HOSTNAME_FUNC": setHostnameDebian,
	},
	2: {
		"RELEASE_FILE":      "/etc/redhat-release",
		"HOST_FILE":         "/etc/sysconfig/network",
		"SET_HOSTNAME_FUNC": setHostnameRhel,
	},
}

// SetHostname - main wrapper for the host configuration
func SetHostname(hostname string) error {
	if err := utils.ValidateHostname(hostname); err != nil {
		return err
	}
	for _, leaf := range template {
		for key, val := range leaf {
			if key == "RELEASE_FILE" {
				if _, err := os.Stat(val.(string)); err == nil {
					if err := leaf["SET_HOSTNAME_FUNC"].(func(string, string) error)(hostname, leaf["HOST_FILE"].(string)); err != nil {
						return err
					}
				}
			}
		}
	}
	return syscall.Sethostname([]byte(hostname))
}

// SetHosts inyended for the hosts file configuration
func SetHosts(hostname, ipv4 string) error {
	fd, err := os.OpenFile("/etc/hosts", os.O_RDWR, 0644)
	if err != nil {
		return err
	}
	defer fd.Close()
	newString := fmt.Sprintf("\n%s\t%s\n", ipv4, hostname)
	patToCheck := ipv4 + `\s+` + hostname
	return AppendToFd(fd, newString, patToCheck)
}

// SetHostDefault sets hostname for appropriate interface
func SetHostsDefault(defaultIface string, force bool) error {
	hname, err := os.Hostname()
	if err != nil {
		return err
	}
	if _, err := net.LookupHost(hname); err == nil {
		if force == false {
			return nil
		}
	}
	iface, err := GetIfaceAddr(defaultIface)
	if err != nil {
		return err
	}
	return SetHosts(hname, strings.Split(iface.String(), "/")[0])
}

// setHostnameDebian responsible for setting hostname for a Debian
// based distribution(Debian,Ubuntu...)
func setHostnameDebian(hostname, file string) error {
	return ioutil.WriteFile(file, []byte(hostname), 0)
}

// setHostnameRhel responsible for setting hostname for
// a distro based on RHEL(RHEL,CentOS...)
func setHostnameRhel(hostname, file string) error {
	return FindAndReplace(file, file, `HOSTNAME\s*=\s*\S+`,
		fmt.Sprintf("HOSTNAME=%s", hostname), 0644)
}
