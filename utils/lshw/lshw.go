package lshw

//
// A simple wrapper for lshw
//

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"
	"sync"
)

type Class string

const (
	// used to refer to the whole machine (laptop, server, desktop computer)
	Systems Class = "system"
	// internal bus converter (PCI-to-PCI brige, AGP bridge, PCMCIA controler, host bridge)
	Bridge Class = "bridge"
	// memory bank that can contain data, executable code, etc.
	// RAM, BIOS, firmware, extension ROM
	Memory Class = "memory"
	// execution processor	 (CPUs, RAID controller on a SCSI bus)
	Processor Class = "processor"
	// memory address range extension ROM, video memory
	Address Class = "address"
	// storage controller	(SCSI controller, IDE controller)
	Storage Class = "storage"
	// random-access storage device discs, optical storage (CD-ROM, DVD±RW...)
	Disk Class = "disk"
	// sequential-access storage device (DAT, DDS)
	Tape Class = "tape"
	// device-connecting bus (USB, SCSI, Firewire)
	Bus Class = "bus"
	// network interface (Ethernet, FDDI, WiFi, Bluetooth)
	Network Class = "network"
	// display adapter (EGA/VGA, UGA...)
	Display Class = "display"
	// user input device (keyboards, mice, joysticks...)
	Input Class = "input"
	// printing device (printer, all-in-one)
	Printer Class = "printer"
	// audio/video device (sound card, TV-output card, video acquisition card)
	Multimedia Class = "multimedia"
	// line communication device (serial ports, modem)
	Communication Class = "communication"
	// energy source (power supply, internal battery)
	Power Class = "power"
	// disk volume	(filesystem, swap, etc.)
	Volume Class = "volume"
	// generic device (used when no pre-defined class is suitable)
	Generic Class = "generic"
	// Print everything
	All Class = "all"
)

type Format string

const (
	FormatXML     Format = "-xml"     // output hardware tree as XML
	FormatJSON    Format = "-json"    // output hardware tree as JSON
	FormatHTML    Format = "-html"    // output hardware tree as HTML
	FormatShort   Format = "-short"   // output hardware paths
	FormatBusinfo Format = "-businfo" // output bus information
	FormatEmpty   Format = ""
)

type Config struct {
	Class  []Class
	Format Format
}

type lshw struct {
	cmd    *exec.Cmd
	config *Config
	lock   sync.Mutex
}

func New(path string, config *Config) (l *lshw, err error) {
	l = new(lshw)
	if path == "" {
		path, err = exec.LookPath("lshw")
		if err != nil {
			return
		}
	}
	l.cmd = exec.Command(path)
	l.config = new(Config)
	l.SetConfig(config)
	return
}

func (l *lshw) cmdreset() {
	l.lock.Lock()
	l.cmd.Args = l.cmd.Args[:1]
	l.lock.Unlock()
}

func (l *lshw) SetClass(class Class) {
	l.lock.Lock()
	if class == All {
		l.config.Class = []Class{}
	}
	l.config.Class = append(l.config.Class, class)
	l.lock.Unlock()
}

func (l *lshw) SetFormat(format Format) {
	l.lock.Lock()
	l.config.Format = format
	l.lock.Unlock()
}

func (l *lshw) SetConfig(config *Config) {
	l.lock.Lock()
	defer l.lock.Unlock()
	l.config.Format = config.Format
	if len(config.Class) == 0 {
		l.config.Class[0] = All
		return
	}
	for _, el := range config.Class {
		if el == All {
			l.config.Class = []Class{}
			l.config.Class = append(l.config.Class, el)
			break
		}
		l.config.Class = append(l.config.Class, el)
	}
}

func (l *lshw) Cmd() string {
	l.makeCmd()
	return strings.Join(l.cmd.Args, " ")
}

func (l *lshw) Execute() (out []byte, err error) {
	l.cmdreset()

	l.lock.Lock()
	defer l.lock.Unlock()

	l.makeCmd()
	out, err = l.cmd.CombinedOutput()
	if err != nil {
		err = fmt.Errorf("%s [%s]", out, err)
	}
	return
}

func (l *lshw) WriteToFile(file string) error {
	out, err := l.Execute()
	if err != nil {
		return err
	}
	l.lock.Lock()
	defer l.lock.Unlock()

	if err := ioutil.WriteFile(file, out, 0); err != nil {
		return err
	}
	return nil
}

func (l *lshw) WriteToStdout() error {
	out, err := l.Execute()
	if err != nil {
		return err
	}
	l.lock.Lock()
	defer l.lock.Unlock()

	_, err = os.Stdout.Write(out)
	return err
}

func (l *lshw) Version() (string, error) {
	l.cmdreset()
	l.cmd.Args = append(l.cmd.Args, "-version")
	ver, err := l.cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("%s [%s]", ver, err)
	}
	return string(ver), nil
}

func (l *lshw) makeCmd() {
	if l.config.Class[0] != All {
		for _, el := range l.config.Class {
			l.cmd.Args = append(l.cmd.Args, "-C", string(el))
		}
	}
	if l.config.Format != FormatEmpty {
		l.cmd.Args = append(l.cmd.Args, string(l.config.Format))
	}
}
