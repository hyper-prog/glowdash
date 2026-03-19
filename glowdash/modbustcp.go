/*
	GlowDash - Smart Home Web Dashboard

	(C) 2024-2026 Péter Deák (hyper80@gmail.com)
	License: GPLv2
*/

package main

import (
	"encoding/binary"
	"errors"
	"fmt"
	"net"
	"strconv"
	"time"
)

/* Modbus TCP API Usage
   This code provides a minimal Modbus TCP client.
   The unit ID identifies the target device (slave) and is set once at Dial time.
   Example usage:

	client, err := Dial("192.168.1.100", "502", 1, 5*time.Second)
	if err != nil {  handle error  }
	defer client.Close()

	// Read a single coil at address 0
	coil, err := client.ReadSingleCoil(0)

	// Write coil 5 ON
	err = client.WriteSingleCoil(5, true)

	// Read a single input register at address 200 (read-only, FC 0x04)
	val, err := client.ReadInputRegister(200)

	// Read 4 holding registers starting at address 100
	regs, err := client.ReadHoldingRegisters(100, 4)

	// Write value 1234 to holding register 101
	err = client.WriteSingleRegister(101, 1234)

All functions return errors for protocol violations, connection issues, or Modbus exceptions.
*/

const (
	fcReadCoils            = 0x01
	fcReadDiscreteInputs   = 0x02
	fcReadHoldingRegisters = 0x03
	fcReadInputRegisters   = 0x04
	fcWriteSingleCoil      = 0x05
	fcWriteSingleRegister  = 0x06
)

type Client struct {
	conn   net.Conn
	tid    uint16
	unitID byte
}

// WriteSingleRegister writes a single holding register (16-bit) at the given address.
func (c *Client) WriteSingleRegister(address uint16, value uint16) error {
	pdu := make([]byte, 4)
	binary.BigEndian.PutUint16(pdu[0:2], address)
	binary.BigEndian.PutUint16(pdu[2:4], value)

	data, err := c.sendRequest(fcWriteSingleRegister, pdu)
	if err != nil {
		return err
	}
	if len(data) != 4 {
		return errors.New("Malformed write register response")
	}

	rspAddr := binary.BigEndian.Uint16(data[0:2])
	rspVal := binary.BigEndian.Uint16(data[2:4])
	if rspAddr != address || rspVal != value {
		return fmt.Errorf("Write register verification failed: addr=0x%04X value=0x%04X", rspAddr, rspVal)
	}
	return nil
}

// ReadInputRegister reads a single input register (16-bit, function code 0x04) at the given address.
// Input registers are read-only and typically represent measured values from a device.
func (c *Client) ReadInputRegister(address uint16) (uint16, error) {
	pdu := make([]byte, 4)
	binary.BigEndian.PutUint16(pdu[0:2], address)
	binary.BigEndian.PutUint16(pdu[2:4], 1)

	data, err := c.sendRequest(fcReadInputRegisters, pdu)
	if err != nil {
		return 0, err
	}
	if len(data) < 1 {
		return 0, errors.New("Malformed response: missing byte count")
	}
	byteCount := int(data[0])
	if byteCount != 2 || len(data[1:]) != byteCount {
		return 0, fmt.Errorf("Unexpected byte count in input register response: got %d", byteCount)
	}
	return binary.BigEndian.Uint16(data[1:3]), nil
}

// ReadHoldingRegisters reads one or more holding registers (16-bit) starting at startAddress.
func (c *Client) ReadHoldingRegisters(startAddress uint16, quantity uint16) ([]uint16, error) {
	if quantity == 0 || quantity > 125 {
		return nil, errors.New("Quantity for read holding registers must be 1..125")
	}

	pdu := make([]byte, 4)
	binary.BigEndian.PutUint16(pdu[0:2], startAddress)
	binary.BigEndian.PutUint16(pdu[2:4], quantity)

	data, err := c.sendRequest(fcReadHoldingRegisters, pdu)
	if err != nil {
		return nil, err
	}
	if len(data) < 1 {
		return nil, errors.New("Malformed response: missing byte count")
	}

	byteCount := int(data[0])
	if len(data[1:]) != byteCount {
		return nil, fmt.Errorf("Byte count mismatch: got %d bytes, declared %d", len(data[1:]), byteCount)
	}
	if byteCount != int(quantity)*2 {
		return nil, fmt.Errorf("Expected %d bytes, got %d", quantity*2, byteCount)
	}

	registers := make([]uint16, quantity)
	for i := 0; i < int(quantity); i++ {
		registers[i] = binary.BigEndian.Uint16(data[1+i*2 : 1+i*2+2])
	}
	return registers, nil
}

func Dial(addr string, port string, unitID byte, timeout time.Duration) (*Client, error) {
	conn, err := net.DialTimeout("tcp", net.JoinHostPort(addr, port), timeout)
	if err != nil {
		return nil, err
	}
	return &Client{conn: conn, tid: 1, unitID: unitID}, nil
}

func (c *Client) Close() error {
	if c.conn == nil {
		return nil
	}
	return c.conn.Close()
}

func (c *Client) nextTID() uint16 {
	c.tid++
	if c.tid == 0 {
		c.tid = 1
	}
	return c.tid
}

func (c *Client) sendRequest(function byte, pduData []byte) ([]byte, error) {
	if c.conn == nil {
		return nil, errors.New("Not connected")
	}

	tid := c.nextTID()
	pduLen := 1 + len(pduData) // function + data
	mbapLen := 1 + pduLen      // unit id + pdu

	frame := make([]byte, 7+pduLen)
	binary.BigEndian.PutUint16(frame[0:2], tid)
	binary.BigEndian.PutUint16(frame[2:4], 0) // protocol id = 0 for Modbus
	binary.BigEndian.PutUint16(frame[4:6], uint16(mbapLen))
	frame[6] = c.unitID
	frame[7] = function
	copy(frame[8:], pduData)

	_ = c.conn.SetDeadline(time.Now().Add(5 * time.Second))
	if _, err := c.conn.Write(frame); err != nil {
		return nil, err
	}

	head := make([]byte, 7)
	if _, err := readFull(c.conn, head); err != nil {
		return nil, err
	}

	rspTID := binary.BigEndian.Uint16(head[0:2])
	if rspTID != tid {
		return nil, fmt.Errorf("Transaction ID mismatch: got %d want %d", rspTID, tid)
	}
	if binary.BigEndian.Uint16(head[2:4]) != 0 {
		return nil, errors.New("Invalid protocol ID in response")
	}
	if head[6] != c.unitID {
		return nil, fmt.Errorf("Unit ID mismatch in response: got %d want %d", head[6], c.unitID)
	}

	length := binary.BigEndian.Uint16(head[4:6])
	if length < 2 {
		// length must cover at least unit id (already read) + function code byte
		return nil, errors.New("Invalid length in response")
	}

	payload := make([]byte, int(length)-1) // already consumed unit id
	if _, err := readFull(c.conn, payload); err != nil {
		return nil, err
	}

	if payload[0] == (function | 0x80) {
		if len(payload) < 2 {
			return nil, errors.New("Malformed exception response")
		}
		return nil, fmt.Errorf("Modbus exception code: 0x%02X", payload[1])
	}
	if payload[0] != function {
		return nil, fmt.Errorf("Unexpected function code in response: 0x%02X", payload[0])
	}

	return payload[1:], nil
}

// ReadSingleCoil reads a single coil (1-bit, function code 0x01) at the given address.
func (c *Client) ReadSingleCoil(address uint16) (bool, error) {
	pdu := make([]byte, 4)
	binary.BigEndian.PutUint16(pdu[0:2], address)
	binary.BigEndian.PutUint16(pdu[2:4], 1)

	data, err := c.sendRequest(fcReadCoils, pdu)
	if err != nil {
		return false, err
	}
	if len(data) < 2 {
		return false, errors.New("malformed response: missing byte count or data")
	}
	if data[0] != 1 {
		return false, fmt.Errorf("unexpected byte count in coil response: got %d", data[0])
	}
	return (data[1] & 0x01) == 1, nil
}

func (c *Client) WriteSingleCoil(address uint16, value bool) error {
	pdu := make([]byte, 4)
	binary.BigEndian.PutUint16(pdu[0:2], address)
	if value {
		binary.BigEndian.PutUint16(pdu[2:4], 0xFF00)
	} else {
		binary.BigEndian.PutUint16(pdu[2:4], 0x0000)
	}

	data, err := c.sendRequest(fcWriteSingleCoil, pdu)
	if err != nil {
		return err
	}
	if len(data) != 4 {
		return errors.New("Malformed write response")
	}

	rspAddr := binary.BigEndian.Uint16(data[0:2])
	rspVal := binary.BigEndian.Uint16(data[2:4])
	expectedVal := uint16(0x0000)
	if value {
		expectedVal = 0xFF00
	}
	if rspAddr != address || rspVal != expectedVal {
		return fmt.Errorf("Write verification failed: addr=0x%04X value=0x%04X", rspAddr, rspVal)
	}
	return nil
}

func readFull(conn net.Conn, buf []byte) (int, error) {
	total := 0
	for total < len(buf) {
		n, err := conn.Read(buf[total:])
		if err != nil {
			return total, err
		}
		total += n
	}
	return total, nil
}

func unpackBits(src []byte, count int) []bool {
	out := make([]bool, count)
	for i := 0; i < count; i++ {
		byteIdx := i / 8
		bitIdx := uint(i % 8)
		if byteIdx < len(src) {
			out[i] = ((src[byteIdx] >> bitIdx) & 0x01) == 1
		}
	}
	return out
}

func parseUint16(name string, s string) (uint16, error) {
	v, err := strconv.ParseUint(s, 0, 16)
	if err != nil {
		return 0, fmt.Errorf("Invalid %s: %w", name, err)
	}
	return uint16(v), nil
}

func parseBoolValue(s string) (bool, error) {
	switch s {
	case "1", "true", "on", "ON", "TRUE":
		return true, nil
	case "0", "false", "off", "OFF", "FALSE":
		return false, nil
	default:
		return false, errors.New("Value must be one of: 1,0,true,false,on,off")
	}
}
