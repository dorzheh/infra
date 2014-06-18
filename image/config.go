// Parses platform configuration file (XML)

// Configuration example:
//
//<?xml version="1.0" encoding="UTF-8"?>
//<Platforms>
//  <Platform>
//    <Name>Test</Name>
//    <Type>1</Type>
//	<Cpus>2</Cpus>
//	<RamSizeGb>3</RamSizeGb>
//	<HddSizeGb>5</HddSizeGb>
//	<FdiskCmd>n\np\n1\n\n+%vM\nn\np\n2\n\n\nt\n2\n82\na\n1\nw\n</FdiskCmd>
//    <Description>Topology for release xxxx</Description>
//    <Partition>
//	  <Sequence>1</Sequence>
//	  <SizeMb>3045</SizeMb>
//      <Label>SLASH</Label>
//      <MountPoint>/</MountPoint>
//      <FileSystem>ext4</FileSystem>
//	  <FileSystemArgs></FileSystemArgs>
//	</Partition>
//	<Partition>
//	  <Sequence>2</Sequence>
//	  <SizeMb>400</SizeMb>
//      <Label>SWAP</Label>
//      <MountPoint>SWAP</MountPoint>
//      <FileSystem>swap</FileSystem>
//	  <FileSystemArgs></FileSystemArgs>
//	</Partition>
//  </Platform>
//</Platforms>`

package image

import (
	"bytes"
	"encoding/xml"
	"io/ioutil"
)

type TopologyType uint8

type Platforms struct {
	Platforms []Topology `xml:"Topology"`
}

type Topology struct {
	Name        string      `xml:"Name"`
	Type        string      `xml:"Type"`
	Cpus        int         `xml:"Cpus"`
	RamSizeMb   int         `xml:"RamSizeMb"`
	HddSizeGb   int         `xml:"HddSizeGb"`
	Description string      `xml:"Description"`
	FdiskCmd    string      `xml: "FdiskCmd"`
	Partitions  []Partition `xml:"Partition"`
}

type Partition struct {
	Sequence       string `xml:"Sequence"`
	SizeMb         int    `xml:"SizeMb"`
	Label          string `xml:"Label"`
	MountPoint     string `xml:"MountPoint"`
	FileSystem     string `xml:"FileSystem"`
	FileSystemArgs string `xml:"FileSystemArgs"`
	Description    string `xml:"description"`
}

// Parse is responsible for parsing appropriate XML file
func ParseConfigFile(xmlpath string) (*Platforms, error) {
	fb, err := ioutil.ReadFile(xmlpath)
	if err != nil {
		return nil, err
	}
	return ParseConfig(fb)
}

func ParseConfig(fb []byte) (*Platforms, error) {
	buf := bytes.NewBuffer(fb)
	p := new(Platforms)
	decoded := xml.NewDecoder(buf)
	if err := decoded.Decode(p); err != nil {
		return nil, err
	}
	return p, nil
}

// TypToTopology returns a topology configuration related to a type
func (p *Platforms) TypeToTopology(topotype TopologyType) *Topology {
	return &p.Platforms[topotype]
}
