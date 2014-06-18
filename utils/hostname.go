package utils

import (
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"regexp"
	"strings"
	"syscall"
)

//hostname max length is 255 chars
const hostnameLength = 256

// hostname: names separated by '.' every name must be max 63 chars in length
// according to RFC 952 <name>  ::= <let>[*[<let-or-digit-or-hyphen>]<let-or-digit>]
// according to RFC 1123 - trailing and starting digits are allowed
var hostnameExpr = regexp.MustCompile(`^([a-zA-Z]|[0-9]|_)(([a-zA-Z]|[0-9]|-|_)*([a-zA-Z]|[0-9]|_))?(\.([a-zA-Z]|[0-9]|_)(([a-zA-Z]|[0-9]|-|_)*([a-zA-Z]|[0-9]))?)*$`)
var hostnameLenExpr = regexp.MustCompile(`^[a-zA-Z0-9-_]{1,64}(\.([a-zA-Z0-9-_]{1,64}))*$`)
var ipAddrExpr = regexp.MustCompile(`^[0-9]+\.[0-9]+\.[0-9]+\.[0-9]+$`)

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
