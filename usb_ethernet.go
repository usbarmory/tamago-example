// https://github.com/f-secure-foundry/tamago-example
//
// Copyright (c) F-Secure Corporation
// https://foundry.f-secure.com
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package main

import (
	"encoding/binary"
	"strings"

	"github.com/f-secure-foundry/tamago/imx6/usb"

	"gvisor.dev/gvisor/pkg/tcpip"
	"gvisor.dev/gvisor/pkg/tcpip/buffer"
	"gvisor.dev/gvisor/pkg/tcpip/stack"
	"gvisor.dev/gvisor/pkg/tcpip/link/channel"
)

const hostMAC = "1a:55:89:a2:69:42"
const deviceMAC = "1a:55:89:a2:69:41"

// set to true to enable packet sniffing
const sniff = false

// populated by setupStack()
var hostMACBytes []byte
var deviceMACBytes []byte
var link *channel.Endpoint

// Ethernet frame buffers
var rx []byte

// Build a CDC control interface.
func buildControlInterface(device *usb.Device) (iface *usb.InterfaceDescriptor) {
	iface = &usb.InterfaceDescriptor{}
	iface.SetDefaults()

	iface.NumEndpoints = 1
	iface.InterfaceClass = 2
	iface.InterfaceSubClass = 6

	iInterface, _ := device.AddString(`CDC Ethernet Control Model (ECM)`)
	iface.Interface = iInterface

	// Set IAD to be inserted before first interface, to support multiple
	// functions in this same configuration.
	iface.IAD = &usb.InterfaceAssociationDescriptor{}
	iface.IAD.SetDefaults()
	// alternate settings do not count
	iface.IAD.InterfaceCount = 1
	iface.IAD.FunctionClass = iface.InterfaceClass
	iface.IAD.FunctionSubClass = iface.InterfaceSubClass

	iFunction, _ := device.AddString(`CDC`)
	iface.IAD.Function = iFunction

	header := &usb.CDCHeaderDescriptor{}
	header.SetDefaults()

	iface.ClassDescriptors = append(iface.ClassDescriptors, header.Bytes())

	union := &usb.CDCUnionDescriptor{}
	union.SetDefaults()

	iface.ClassDescriptors = append(iface.ClassDescriptors, union.Bytes())

	ethernet := &usb.CDCEthernetDescriptor{}
	ethernet.SetDefaults()

	iMacAddress, _ := device.AddString(strings.ReplaceAll(hostMAC, ":", ""))
	ethernet.MacAddress = iMacAddress

	iface.ClassDescriptors = append(iface.ClassDescriptors, ethernet.Bytes())

	ep2IN := &usb.EndpointDescriptor{}
	ep2IN.SetDefaults()
	ep2IN.EndpointAddress = 0x82
	ep2IN.Attributes = 3
	ep2IN.MaxPacketSize = 16
	ep2IN.Interval = 9
	ep2IN.Function = ECMControl

	iface.Endpoints = append(iface.Endpoints, ep2IN)

	return
}

// Build a CDC data interface.
func buildDataInterface(device *usb.Device) (iface *usb.InterfaceDescriptor) {
	iface = &usb.InterfaceDescriptor{}
	iface.SetDefaults()

	// ECM requires the use of "alternate settings" for its data interface
	iface.AlternateSetting = 1
	iface.NumEndpoints = 2
	iface.InterfaceClass = 10

	iInterface, _ := device.AddString(`CDC Data`)
	iface.Interface = iInterface

	ep1IN := &usb.EndpointDescriptor{}
	ep1IN.SetDefaults()
	ep1IN.EndpointAddress = 0x81
	ep1IN.Attributes = 2
	ep1IN.MaxPacketSize = maxPacketSize
	ep1IN.Function = ECMTx

	iface.Endpoints = append(iface.Endpoints, ep1IN)

	ep1OUT := &usb.EndpointDescriptor{}
	ep1OUT.SetDefaults()
	ep1OUT.EndpointAddress = 0x01
	ep1OUT.Attributes = 2
	ep1OUT.MaxPacketSize = maxPacketSize
	ep1OUT.Function = ECMRx

	iface.Endpoints = append(iface.Endpoints, ep1OUT)

	return
}

func configureECM(device *usb.Device, configurationIndex int) {
	controlInterface := buildControlInterface(device)
	device.Configurations[configurationIndex].AddInterface(controlInterface)

	dataInterface := buildDataInterface(device)
	device.Configurations[configurationIndex].AddInterface(dataInterface)
}

// ECMControl implements the endpoint 2 IN function.
func ECMControl(_ []byte, lastErr error) (in []byte, err error) {
	// ignore for now
	return
}

// ECMTx implements the endpoint 1 IN function, used to transmit Ethernet
// packet from device to host.
func ECMTx(_ []byte, lastErr error) (in []byte, err error) {
	info, valid := link.Read()

	if !valid {
		return
	}

	hdr := info.Pkt.Header.View()
	payload := info.Pkt.Data.ToView()

	proto := make([]byte, 2)
	binary.BigEndian.PutUint16(proto, uint16(info.Proto))

	// Ethernet frame header
	in = append(in, hostMACBytes...)
	in = append(in, deviceMACBytes...)
	in = append(in, proto...)
	// packet header
	in = append(in, hdr...)
	// payload
	in = append(in, payload...)

	return
}

// ECMRx implements the endpoint 1 OUT function, used to receive ethernet
// packet from host to device.
func ECMRx(out []byte, lastErr error) (_ []byte, err error) {
	if len(rx) == 0 && len(out) < 14 {
		return
	}

	rx = append(rx, out...)

	// more data expected or zero length packet
	if len(out) == maxPacketSize {
		return
	}

	hdr := buffer.NewViewFromBytes(rx[0:14])
	proto := tcpip.NetworkProtocolNumber(binary.BigEndian.Uint16(rx[12:14]))
	payload := buffer.NewViewFromBytes(rx[14:])

	pkt := &stack.PacketBuffer{
		LinkHeader: hdr,
		Data:       payload.ToVectorisedView(),
	}

	link.InjectInbound(proto, pkt)

	rx = []byte{}

	return
}
