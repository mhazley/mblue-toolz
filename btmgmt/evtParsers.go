package btmgmt

import (
	"encoding/binary"
	"fmt"
	"github.com/mame82/P4wnP1_aloa/mnetlink"
	"net"
)

/*
// ToDo: Convert these two parsers to interface format

func parseEvtCmdStatus(payload []byte) (cmd CmdCode, status CmdStatus, err error) {
	if len(payload) != 3 {
		err = ErrPayloadFormat
		return
	}
	cmd = CmdCode(binary.LittleEndian.Uint16(payload[0:2]))
	status = CmdStatus(payload[2])
	return
}
*/

/* Parsers */

type ParsePayload interface {
	UpdateFromPayload(pay []byte) (err error)
}

type CommandStatusEvent struct {
	CmdCode CmdCode
	Status  CmdStatus
}

func (cs *CommandStatusEvent) UpdateFromPayload(payload []byte) (err error) {
	if len(payload) != 3 { //exact 3, in contrast to command complete
		return ErrPayloadFormat
	}
	cs.CmdCode = CmdCode(binary.LittleEndian.Uint16(payload[0:2]))
	cs.Status = CmdStatus(payload[2])
	return
}

type CommandCompleteEvent struct {
	CmdCode      CmdCode
	Status       CmdStatus
	ReturnParams []byte
}

func (cc *CommandCompleteEvent) UpdateFromPayload(payload []byte) (err error) {
	if len(payload) < 3 {
		return ErrPayloadFormat
	}
	cc.CmdCode = CmdCode(binary.LittleEndian.Uint16(payload[0:2]))
	cc.Status = CmdStatus(payload[2])
	cc.ReturnParams = payload[3:]
	return
}

type ControllerInformation struct {
	Address           Address
	BluetoothVersion  byte
	Manufacturer      uint16
	SupportedSettings ControllerSettings
	CurrentSettings   ControllerSettings
	ClassOfDevice     DeviceClass // 3, till clear how to parse
	Name              string      //[249]byte, 0x00 terminated
	ShortName         string      //[11]byte, 0x00 terminated

	ServiceNetworkServerGn   bool
	ServiceNetworkServerNap  bool
	ServiceNetworkServerPanu bool
}

func (ci *ControllerInformation) UpdateFromPayload(p []byte) (err error) {
	if len(p) != 280 {
		return ErrPayloadFormat
	}

	ci.Address.UpdateFromPayload(p[0:6])
	ci.BluetoothVersion = p[6]
	ci.Manufacturer = binary.LittleEndian.Uint16(p[7:9])
	ci.SupportedSettings.UpdateFromPayload(p[9:13])
	ci.CurrentSettings.UpdateFromPayload(p[13:17])
	ci.ClassOfDevice.UpdateFromPayload(p[17:20])
	ci.Name = string(zeroTerminateSlice(p[20:269]))
	ci.ShortName = string(zeroTerminateSlice(p[269:]))
	return
}

func (ci ControllerInformation) String() string {
	res := fmt.Sprintf("addr %s version %d manufacturer %d class %s", ci.Address.String(), ci.BluetoothVersion, ci.Manufacturer, ci.ClassOfDevice.String())
	res += fmt.Sprintf("\nSupported settings: %+v", ci.SupportedSettings)
	res += fmt.Sprintf("\nCurrentSettings:    %+v", ci.CurrentSettings)
	res += fmt.Sprintf("\nname %s short name %s", ci.Name, ci.ShortName)
	return res
}

type DeviceClass struct {
	Octets []byte
}

func (c *DeviceClass) String() string {
	return fmt.Sprintf("0x%.2x%.2x%.2x", c.Octets[0], c.Octets[1], c.Octets[2])
}

func (c *DeviceClass) UpdateFromPayload(pay []byte) (err error) {
	if len(pay) != 3 {
		return ErrPayloadFormat
	}
	c.Octets = copyReverse(pay)
	return
}

type Address struct {
	Addr net.HardwareAddr
}

func (a *Address) String() string {
	return a.Addr.String()
}

func (a *Address) UpdateFromPayload(pay []byte) (err error) {
	if len(pay) != 6 {
		return ErrPayloadFormat
	}
	p := copyReverse(pay)
	a.Addr = net.HardwareAddr(p)
	return
}

// ToDo: Stringer interface for ControllerSettings (needs map)
type ControllerSettings struct {
	Powered                 bool
	Connectable             bool
	FastConnectable         bool
	Discoverable            bool
	Bondable                bool
	LinkLevelSecurity       bool
	SecureSimplePairing     bool
	BrEdr                   bool
	HighSpeed               bool
	LowEnergy               bool
	Advertising             bool
	SecureConnections       bool
	DebugKeys               bool
	Privacy                 bool
	ControllerConfiguration bool
	StaticAddress           bool
}

func (cd *ControllerSettings) UpdateFromPayload(pay []byte) (err error) {
	if len(pay) != 4 {
		return ErrPayloadFormat
	}
	//b := (pay)[0]
	b := mnetlink.Hbo().Uint32(pay[0:4])

	cd.Powered = testBit(b, 0)
	cd.Connectable = testBit(b, 1)
	cd.FastConnectable = testBit(b, 2)
	cd.Discoverable = testBit(b, 3)
	cd.Bondable = testBit(b, 4)
	cd.LinkLevelSecurity = testBit(b, 5)
	cd.SecureSimplePairing = testBit(b, 6)
	cd.BrEdr = testBit(b, 7)
	cd.HighSpeed = testBit(b, 8)
	cd.LowEnergy = testBit(b, 9)
	cd.Advertising = testBit(b, 10)
	cd.SecureConnections = testBit(b, 11)
	cd.DebugKeys = testBit(b, 12)
	cd.Privacy = testBit(b, 13)
	cd.ControllerConfiguration = testBit(b, 14)
	cd.StaticAddress = testBit(b, 15)
	return nil
}

type ControllerIndexList struct {
	Indices []uint16
}

func (cil *ControllerIndexList) String() string {
	res := "Controller Index List: "
	for _, ctrlIdx := range cil.Indices {
		res += fmt.Sprintf("%d ", ctrlIdx)
	}
	return res
}

func (cil *ControllerIndexList) UpdateFromPayload(p []byte) (err error) {
	if len(p) < 2 {
		return ErrPayloadFormat
	}
	numIndices := binary.LittleEndian.Uint16(p[0:2])
	cil.Indices = make([]uint16, numIndices)
	off := 2
	for i, _ := range cil.Indices {
		cil.Indices[i] = binary.LittleEndian.Uint16(p[off : off+2])
		off += 2
	}
	return
}

type SupportedCommands struct {
	Commands []CmdCode
	Events   []EvtCode
}

func (sc *SupportedCommands) String() string {
	res := "Supported commands: "
	for _, cmd := range sc.Commands {
		res += fmt.Sprintf("%d ", cmd)
	}
	res += "Supported events: "
	for _, evt := range sc.Events {
		res += fmt.Sprintf("%d ", evt)
	}
	return res
}

func (sc *SupportedCommands) UpdateFromPayload(p []byte) (err error) {
	if len(p) < 4 {
		return ErrPayloadFormat
	}
	numCommands := binary.LittleEndian.Uint16(p[0:2])
	numEvents := binary.LittleEndian.Uint16(p[2:4])
	sc.Commands = make([]CmdCode, numCommands)
	sc.Events = make([]EvtCode, numEvents)
	off := 4
	for i, _ := range sc.Commands {
		uiCmd := binary.LittleEndian.Uint16(p[off : off+2])
		sc.Commands[i] = CmdCode(uiCmd)
		off += 2
	}
	for i, _ := range sc.Events {
		uiEvt := binary.LittleEndian.Uint16(p[off : off+2])
		sc.Events[i] = EvtCode(uiEvt)
		off += 2
	}
	return nil
}

type VersionInformation struct {
	Version  uint8
	Revision uint16
}

func (v *VersionInformation) UpdateFromPayload(pay []byte) (err error) {
	if len(pay) != 3 {
		return ErrPayloadFormat
	}
	v.Version = pay[0]
	v.Revision = binary.LittleEndian.Uint16(pay[1:3])
	return
}

func (v VersionInformation) String() string {
	return fmt.Sprintf("Version %d.%d", v.Version, v.Revision)
}

type AddressType uint8

const (
	BR_EDR AddressType = iota
	LE_PUBLIC
	LE_RANDOM
)

func (a AddressType) String() string {
	switch a {
	case BR_EDR:
		return "BR/EDR"
	case LE_PUBLIC:
		return "LE PUBLIC"
	case LE_RANDOM:
		return "LE RANDOM"
	}
	return "unknown"
}

type ConnectionAddress struct {
	Addr     net.HardwareAddr
	AddrType AddressType
}

func (c *ConnectionAddress) UpdateFromPayload(pay []byte) (err error) {
	if len(pay) != 7 {
		return ErrPayloadFormat
	}
	p := copyReverse(pay[0:6])
	c.Addr = net.HardwareAddr(p)
	c.AddrType = AddressType(pay[6])
	return
}

type ConnectionInfoList struct {
	ConnectionCount uint16
	ConnectionList  []ConnectionAddress
}

func (c *ConnectionInfoList) UpdateFromPayload(pay []byte) (err error) {
	if len(pay) < 2 {
		return ErrPayloadFormat
	}
	c.ConnectionCount = binary.LittleEndian.Uint16(pay[0:2])
	c.ConnectionList = make([]ConnectionAddress, c.ConnectionCount)

	for i, _ := range c.ConnectionList {
		startIdx := 2 + (i * 8)
		endIdx := startIdx + 7
		c.ConnectionList[i].UpdateFromPayload(pay[startIdx:endIdx])
	}
	return
}

func (c ConnectionInfoList) String() string {
	res := "Connection List: "
	for _, conn := range c.ConnectionList {
		res += fmt.Sprintf("%v [%v]  ", conn.Addr.String(), conn.AddrType.String())
	}
	return res
}
