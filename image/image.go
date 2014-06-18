package image

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/dorzheh/pipe"
	"github.com/dorzheh/infra/utils"
)

// the structure represents a mapper device
type mapperDevice struct {
	// device name(full path)
	name string
	// mount point the device is mounted
	mountPoint string
	// Linux tree path
	//path string
}

type loopDevice struct {
	// device name(full path)
	name string
	// slice of mapperDevice pointers
	mappers []*mapperDevice
	// counter - amount of mappers the loop device contains
	amountOfMappers uint8
}

// the structure represents loop device manipulation
type image struct {
	// path to the image
	imgpath string
	// path to sirectory intended for mounting the image
	slashpath string
	// platform configuration file (image config)
	conf *Topology
	// do we need to partition and format the image
	needToFormat bool
	// loop device structure
	*loopDevice
}

//// Public methods ////

// New gets a path to configuration directory
// path to temporary directory where the vHDD image supposed to be mounted
// and path to vHDD image.
// Returns a pointer to the structure and error/nil
func New(pathToRawImage, rootfsMp string, imageConfig *Topology) (img *image, err error) {
	needToFormat := false
	if _, err = os.Stat(pathToRawImage); err != nil {
		if err = create(pathToRawImage, imageConfig.HddSizeGb); err != nil {
			return
		}
		needToFormat = true
	}
	img = new(image)
	img.imgpath = pathToRawImage
	img.slashpath = rootfsMp
	img.conf = imageConfig
	img.needToFormat = needToFormat
	img.loopDevice = &loopDevice{}
	img.loopDevice.amountOfMappers = 0
	return
}

// parse processes the RAW image
// Returns error/nil
func (i *image) Parse() error {
	var err error
	if i.loopDevice.name, err = i.bind(); err != nil {
		return err
	}
	if i.needToFormat {
		return i.partTable()
	}
	return i.addMappers()
}

// Customize intended for the target customization
// - pathToPlatformDir - path to directory containing platform configuration XML file
func (i *image) Customize(pathToPlatformDir string) error {
	if i.amountOfMappers == 0 {
		return fmt.Errorf("amount of mappers is 0.Seems you didn't call Parse().")
	}
	return Customize(i.slashpath, pathToPlatformDir)
}

// Release is trying to release the image correctly
// by unmounting the mappers in reverse order
// and the cleans up a temporary stuff
// Returns error or nil
func (i *image) Release() error {
	umount, err := exec.LookPath("umount")
	if err != nil {
		return err
	}
	// Release registered mount points
	var index uint8 = i.loopDevice.amountOfMappers - 1
	for i.loopDevice.amountOfMappers != 0 {
		mounted, err := utils.Mounted(i.loopDevice.mappers[index].name, i.loopDevice.mappers[index].mountPoint)
		if err != nil {
			return err
		}
		if mounted {
			cmd := exec.Command(umount, "-l", i.loopDevice.mappers[index].mountPoint)
			if out, err := cmd.CombinedOutput(); err != nil {
				return fmt.Errorf("%s [%v]", out, err)
			}
			i.loopDevice.amountOfMappers--
		}
	}
	// remove mount point
	if err := utils.RemoveIfExists(true, i.slashpath); err != nil {
		return err
	}
	// "unmap" the image
	if err := exec.Command("kpartx", "-d", i.loopDevice.name).Run(); err != nil {
		return errors.New("kpartx" + " -d " + i.loopDevice.name)
	}
	// unbind the image
	if err := exec.Command("losetup", "-d", i.loopDevice.name).Run(); err != nil {
		return errors.New("losetup -d " + i.loopDevice.name)
	}
	return nil
}

// ReleaseOnInterrupt is trying to release appropriate image
// in case SIGHUP, SIGINT or SIGTERM signal received
func (i *image) ReleaseOnInterrupt() {
	//create a channel for interrupt handler
	interrupt := make(chan os.Signal, 1)
	// create an interrupt handler
	signal.Notify(interrupt, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM)
	// run a seperate goroutine.
	go func() {
		for {
			select {
			case <-interrupt:
				i.Release()
			}
		}
	}()
}

// Exports amount of mappers to the outside of the class
func (i *image) AmountOfMappers() uint8 {
	return i.amountOfMappers
}

/// Private stuff ///

func (i *image) bind() (loopDeviceStr string, err error) {
	loopDeviceBytes, err := exec.Command("losetup", "-f").Output()
	if err != nil {
		return
	}
	loopDeviceStr = strings.TrimSpace(string(loopDeviceBytes))
	if err = exec.Command("losetup", loopDeviceStr, i.imgpath).Run(); err != nil {
		err = fmt.Errorf("cannot bind %s image to %s device", i.imgpath, loopDeviceStr)
		return
	}
	return
}

// partTable creates partition table on the RAW disk
func (i *image) partTable() error {
	fdisk, err := exec.LookPath("fdisk")
	if err != nil {
		return errors.New("fdisk not found")
	}
	if err := utils.CmdPipe("echo", "-e "+i.conf.FdiskCmd, fdisk, i.loopDevice.name); err != nil {
		return err
	}
	return i.makeFs()
}

// makeFs intended for creating file system
func (i *image) makeFs() error {
	mappers, err := getMappers(i.loopDevice.name)
	if err != nil {
		return err
	}
	if err := i.validatePconf(len(mappers)); err != nil {
		return err
	}
	for index, part := range i.conf.Partitions {
		mapper := mappers[index]
		// create SWAP and do not add to the mappers slice
		if strings.ToLower(part.FileSystem) == "swap" {
			mkswap, err := exec.LookPath("mkswap")
			if err != nil {
				return errors.New("mkswap not found")
			}
			if out, err := exec.Command(mkswap, "-L", part.Label, mapper).CombinedOutput(); err != nil {
				return fmt.Errorf("%s [%v]", out, err)
			}
		} else {
			// treat a "none-swap" mapper
			mkfs, err := exec.LookPath("mkfs")
			if err != nil {
				return errors.New("mkfs not found")
			}
			cmd := fmt.Sprintf("-t %v -L %v %v %v", part.FileSystem,
				part.Label, part.FileSystemArgs, mapper)
			if out, err := exec.Command(mkfs, strings.Fields(cmd)...).CombinedOutput(); err != nil {
				return fmt.Errorf("%s [%v]", out, err)
			}
			if err := i.addMapper(mapper, part.MountPoint); err != nil {
				return err
			}
		}
	}
	return nil
}

// addMappers registers appropriate mappers
func (i *image) addMappers() error {
	mappers, err := getMappers(i.loopDevice.name)
	if err != nil {
		return err
	}
	if err := i.validatePconf(len(mappers)); err != nil {
		return err
	}
	for index, part := range i.conf.Partitions {
		mapper := mappers[index]
		// create SWAP and do not add to the mappers slice
		if strings.ToLower(part.FileSystem) != "swap" {
			if err := i.addMapper(mapper, part.MountPoint); err != nil {
				return err
			}
		}
	}
	return nil
}

// addMapper registers appropriate mapper and it's mount point
func (i *image) addMapper(mapperDeviceName, path string) error {
	mountPoint := filepath.Join(i.slashpath, path)
	if err := utils.CreateDirRecursively(mountPoint, 0755, 0, 0, false); err != nil {
		return err
	}
	mounted, err := utils.Mounted(mapperDeviceName, mountPoint)
	if err != nil {
		return err
	}
	if !mounted {
		cmd := exec.Command("mount", mapperDeviceName, mountPoint)
		if out, err := cmd.CombinedOutput(); err != nil {
			return fmt.Errorf("%s [%v]", out, err)
		}
	}
	// add mapper
	i.mappers = append(i.mappers,
		&mapperDevice{
			name:       mapperDeviceName,
			mountPoint: mountPoint,
		},
	)
	i.loopDevice.amountOfMappers++
	return nil
}

// validatePconf is responsible for platform configuration file validation
func (i *image) validatePconf(amountOfMappers int) error {
	if amountOfMappers != len(i.conf.Partitions) {
		return fmt.Errorf("amount of partitions defined = %v, actual amount is %v",
			len(i.conf.Partitions), amountOfMappers)
	}
	return nil
}

// MakeBootable is responsible for making RAW disk bootable
func MakeBootable(pathToGrubBin, pathToImage string) error {
	p := pipe.Line(
		pipe.Printf("device (hd0) %s\nroot (hd0,0)\nsetup (hd0)\n", pathToImage),
		pipe.Exec(pathToGrubBin),
	)
	output, err := pipe.CombinedOutput(p)
	//fmt.Printf("%s", output)
	if err != nil {
		return fmt.Errorf("%s [%v]", output, err)
	}
	return nil
}

// create is intended for creating RAW image
func create(pathToRawImage string, hddSize int) error {
	finfo, err := os.Stat(pathToRawImage)
	switch {
	case err != nil && !strings.Contains(err.Error(), "no such file or directory"):
		return err

	case err == nil && finfo.IsDir():
		return errors.New(pathToRawImage + " is a directory")
	}
	ddCmd, err := exec.LookPath("dd")
	if err != nil {
		return errors.New("dd not found")
	}
	cmdStr := fmt.Sprintf("if=/dev/zero of=%s count=1 bs=1 seek=%vG",
		pathToRawImage, hddSize)
	out, err := exec.Command(ddCmd, strings.Fields(cmdStr)...).CombinedOutput()
	if err != nil {
		return fmt.Errorf("%s [%v]", out, err)
	}
	return nil
}

// getMappers is responsible for finding mappers bound to appropriate loop device
// and providing the stuff as a slice
func getMappers(loopDeviceName string) (mappers []string, err error) {
	kpartx, err := exec.LookPath("kpartx")
	if err != nil {
		return
	}
	err = exec.Command(kpartx, "-av", loopDeviceName).Run()
	if err != nil {
		return
	}
	// somehow on RHEL based systems refresh might take some time therefore
	// no mappers are available until then
	duration, err := time.ParseDuration("1s")
	if err != nil {
		return
	}
	time.Sleep(duration)
	filepath.Walk("/dev", func(mapperDeviceName string, info os.FileInfo, err error) error {
		if strings.Contains(mapperDeviceName, "/"+filepath.Base(loopDeviceName)+"p") {
			// skip extended partition
			if !strings.HasSuffix(mapperDeviceName, "4") {
				mappers = append(mappers, mapperDeviceName)
			}
		}
		return nil
	})
	if len(mappers) == 0 {
		err = errors.New("mappers not found")
	}
	return
}
