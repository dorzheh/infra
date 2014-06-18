// System related functions

package utils

import (
	"syscall"
)

// SysinfoRam returns total amount of installed RAM in Megabytes
func SysinfoRam() (int, error) {
	sysInfoBufPtr := new(syscall.Sysinfo_t)
	if err := syscall.Sysinfo(sysInfoBufPtr); err != nil {
		return 0, err
	}
	return int(sysInfoBufPtr.Totalram / 1024 / 1024), nil
}

type Utsname syscall.Utsname

func uname() (*syscall.Utsname, error) {
	uts := &syscall.Utsname{}
	if err := syscall.Uname(uts); err != nil {
		return nil, err
	}
	return uts, nil
}
