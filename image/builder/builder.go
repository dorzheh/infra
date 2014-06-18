package builder

import (
	"github.com/dorzheh/infra/image"
)

// organizing the stuff

// BuildImage encapsulates entire logic related to a
// virtual appliance creation and population
// filler - image.Rootfs interface implementation
// config - platform configuration
// pathToRawImage - path to the virtual appliance HDD image
// pathToRootfsMp - path to "/" mount point (a mapper containing / content)
// pathToPlatformDir -path to directory containing appropriate configuration (XML file)
// and a stuff intended for the image/platform customization
func BuildImage(filler image.Rootfs, config *image.Topology, pathToRawImage, pathToRootfsMp, pathToPlatformDir string) error {
	img, err := image.New(config, pathToRawImage, pathToRootfsMp)
	if err != nil {
		return err
	}
	// parse the vHDD image
	if err := img.Parse(); err != nil {
		return err
	}
	img.ReleaseOnInterrupt()
	// create rootfs
	if err := filler.MakeRootfs(pathToRootfsMp); err != nil {
		return err
	}
	if pathToPlatformDir != "" {
		if err := img.Customize(pathToPlatformDir); err != nil {
			return err
		}
	}
	defer func() {
		if err := img.Release(); err != nil {
			panic(err)
		}

	}()
	// install application.
	if err := filler.InstallApp(pathToRootfsMp); err != nil {
		return err
	}
	return nil
}
