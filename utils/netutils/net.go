// Generic functions

package netutils

import (
	"errors"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strings"
)

var ErrorNoIpv4 = errors.New("no IPv4 address")

///// Functions providing verification services /////
// HexToDec converts hexadecimal number (two bytes) to decimal
func HexToDec(hexMask string, base int) (int, bool) {
	var number int = 0
	var i int
	for i = base; i < len(hexMask); i++ {
		if '0' <= hexMask[i] && hexMask[i] <= '9' {
			number *= 16
			number += int(hexMask[i] - '0')
		} else if 'a' <= hexMask[i] && hexMask[i] <= 'f' {
			number *= 16
			number += int(hexMask[i]-'a') + 10
		} else if 'A' <= hexMask[i] && hexMask[i] <= 'F' {
			number *= 16
			number += int(hexMask[i]-'A') + 10
		} else {
			break
		}
		if number >= 0xFFFFFF {
			return 0, false
		}
	}
	if i == base {
		return 0, false
	}
	return number, true
}

// ConvertMaskHexToDotString converts a subnet mask from hexadecimal to a decimal form
func ConvertMaskHexToDotString(hexMask net.IPMask) string {
	var strMask string // dec mask (255.255.255.0 for example)
	var octet int      // octet (255 for example)
	var ok bool
	maxOctet := 4 // IPv4 consists of 4 octets

	for i := 0; i < maxOctet; i++ {
		if octet, ok = HexToDec(fmt.Sprintf("%s", (hexMask[i:i+1])), 0); !ok {
			return ""
			//log.Fatalf("wrong mask is provided %s ", hexMask[:])
		}
		// in case it's the last octet add do not add '.'
		if i == (maxOctet - 1) {
			strMask += fmt.Sprintf("%d", octet)
		} else {
			strMask += fmt.Sprintf("%d.", octet)
		}
	}
	// return mask represented as dec
	return strMask
}

// CalculateBcast
// Gets strings (ip and netmask)
// Returns a string representing broadcast address
func CalculateBcast(ip, netmask string) string {
	if ip == "" && netmask == "" {
		return ""
	}
	var ipAddr [4]int
	fmt.Sscanf(ip, "%d.%d.%d.%d", &ipAddr[0], &ipAddr[1], &ipAddr[2], &ipAddr[3])

	var netMask [4]int
	fmt.Sscanf(netmask, "%d.%d.%d.%d", &netMask[0], &netMask[1], &netMask[2], &netMask[3])

	var brdCast [4]int
	for oct := 0; oct < 4; oct++ {
		brdCast[oct] = ((ipAddr[oct] * 1) & (netMask[oct] * 1)) + (255 ^ (netMask[oct] * 1))
	}
	return fmt.Sprintf("%d.%d.%d.%d", brdCast[0], brdCast[1], brdCast[2], brdCast[3])
}

// ConvertMaskDotStringToBits converts decimal mask (255.255.0.0 for example) to
// net.IPMask and returns once and bits of the mask
func ConvertMaskDotStringToBits(netmask string) (int, int) {
	return net.ParseIP(netmask).DefaultMask().Size()
}

// GetLocalBridges finds bridges installed on the local system
// Returns a slice containing bridges names and error/nil
func GetLocalBridges() (installedBridges []string, err error) {
	var netInterfaces []net.Interface
	netInterfaces, err = net.Interfaces()
	if err != nil {
		return
	}
	for _, inet := range netInterfaces {
		if _, err = os.Stat("/sys/class/net/" + inet.Name + "/bridge"); err == nil {
			installedBridges = append(installedBridges, inet.Name)
		}
	}
	if len(installedBridges) == 0 {
		err = errors.New("cannot find any bridge")
	}
	return
}

// ParseInterfacesRH parses network configuration  based on RHEL topology
// and represents the contents as a map
func ParseInterfacesRH() (map[string]map[string]string, error) {
	ifaces, err := net.Interfaces()
	if err != nil {
		return nil, err
	}
	dir := "/etc/sysconfig/network-scripts"
	if _, err := os.Stat(dir); err != nil {
		return nil, err
	}
	pattern := "ifcfg-"
	ignoreFile := "ifcfg-lo"
	var files []string
	err = filepath.Walk(dir, func(ifcfgFile string, info os.FileInfo, err error) error {
		if strings.Contains(ifcfgFile, pattern) && !strings.Contains(ifcfgFile, ignoreFile) {
			files = append(files, ifcfgFile)
		}
		return nil
	})
	ifacesMap := make(map[string]map[string]string)
	for _, iface := range ifaces {
		if !strings.HasPrefix(iface.Name, "veth") || !strings.HasPrefix(iface.Name, "lo") {
			netAddr, err := GetIfaceAddr(iface.Name)
			if err != nil && err != ErrorNoIpv4 {
				return nil, err
			}
			pat := `DEVICE\s*=\s*` + iface.Name
			for _, file := range files {
				if PatternFound(file, pat) {
					var addr string
					switch {
					case netAddr == nil:
						addr = "N/A"
					default:
						addr = netAddr.String()
					}
					label := strings.Split(filepath.Base(file), "-")[1]
					ifacesMap[label] = make(map[string]string)
					ifacesMap[label][iface.Name] = addr
				}
			}
		}
	}
	return ifacesMap, nil
}

// Return the IPv4 address of a network interface
func GetIfaceAddr(name string) (net.Addr, error) {
	iface, err := net.InterfaceByName(name)
	if err != nil {
		return nil, err
	}
	addrs, err := iface.Addrs()
	if err != nil {
		return nil, err
	}
	var addrs4 []net.Addr
	for _, addr := range addrs {
		ip := (addr.(*net.IPNet)).IP
		if ip4 := ip.To4(); len(ip4) == net.IPv4len {
			addrs4 = append(addrs4, addr)
		}
	}
	if len(addrs4) == 0 {
		return nil, ErrorNoIpv4
	}
	return addrs4[0], nil
}
