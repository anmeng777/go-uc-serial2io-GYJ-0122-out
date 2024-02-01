package GYJ_0122

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"github.com/jacobsa/go-serial/serial"
	"io"
	"log"
	"time"
)

// SerialPortWrapper 串口包装器
type SerialPortWrapper struct {
	Port            io.ReadWriteCloser
	PortName        string
	BaudRate        uint
	DataBits        uint
	StopBits        uint
	MinimumReadSize uint
	Timeout         time.Duration
}

type Packet struct {
	Header []byte
	// 数据长度
	Length  uint16
	Command []byte
	Data    []byte
	Tail    []byte
}

type UnpackWrapper struct {
	// 数据包头
	PacketHeader []byte
	// 数据包长度在数据包中的位置，0表示数据包长度依据DataSize指定的字节数
	LengthIndex int
	// 数据包长度的字节数，0表示数据包长度依据DataSize指定的字节数
	LengthSize int
	// 命令码在数据包中的位置
	CommandIndex int
	// 命令码的字节数
	CommandSize int
	// 数据包数据在数据包中的位置
	DataIndex int
	// 数据包数据的字节数, 0表示数据包长度依据LengthSize指定的字节数
	DataSize int
	// 数据包尾
	PacketTail []byte
}

func (uw *UnpackWrapper) Unpack(buf []byte) (int, *Packet, error) {

	if len(buf) < len(uw.PacketHeader)+uw.LengthSize+uw.CommandSize+len(uw.PacketTail) {
		return 0, nil, fmt.Errorf("invalid packet length")
	}

	packet := &Packet{}

	// 包头
	//headerBytes := make([]byte, len(uw.PacketHeader))
	headerBytes, err := safeAccessBytes(buf, 0, len(uw.PacketHeader))
	if err != nil {
		return 0, nil, fmt.Errorf("index out of array1")
	}
	if !bytes.Equal(headerBytes, uw.PacketHeader) {
		//log.Println("Invalid packet header111")
		//buf = buf[1:]
		return -1, nil, fmt.Errorf("invalid packet header")
	}
	packet.Header = headerBytes

	// 包尾
	tailIndex := bytes.Index(buf, uw.PacketTail)
	if tailIndex < 0 {
		return 0, nil, fmt.Errorf("cannot find packet tail")
	}

	// 数据长度
	if uw.LengthIndex > 0 && uw.LengthSize > 0 {
		lengthBytes, err := safeAccessBytes(buf, uw.LengthIndex, uw.LengthSize)
		if err != nil {
			return 0, nil, fmt.Errorf("index out of array2")
		}
		dataLength := binary.LittleEndian.Uint16(lengthBytes)
		if dataLength > 100 {
			log.Println("Invalid packet length")
			//buf = buf[1:]
			return -1, nil, fmt.Errorf("invalid packet length")
		}
		packet.Length = dataLength
		if len(buf) < len(uw.PacketHeader)+uw.LengthSize+uw.CommandSize+int(packet.Length)+len(uw.PacketTail) {
			return 0, nil, fmt.Errorf("invalid packet length")
		}
	}

	// 命令码
	commandBytes, err := safeAccessBytes(buf, uw.CommandIndex, uw.CommandSize)
	if err != nil {
		return 0, nil, fmt.Errorf("index out of array3")
	}
	packet.Command = commandBytes

	if tailIndex < len(uw.PacketHeader)+uw.LengthSize+uw.CommandSize+uw.DataSize+len(uw.PacketTail) {
		return 1, packet, nil
	}

	// 数据
	if packet.Length > 0 {
		dataBytes, err := safeAccessBytes(buf, uw.DataIndex, int(packet.Length))
		if err != nil {
			return 0, nil, fmt.Errorf("index out of array4")
		}
		packet.Data = dataBytes
	} else {
		dataBytes, err := safeAccessBytes(buf, uw.DataIndex, uw.DataSize)
		if err != nil {
			return 0, nil, fmt.Errorf("index out of array5")
		}
		packet.Data = dataBytes
	}

	// 包尾
	if tailIndex != uw.DataIndex+len(packet.Data) {
		return 1, packet, nil
	}
	tailBytes, err := safeAccessBytes(buf, uw.DataIndex+len(packet.Data), len(uw.PacketTail))
	if err != nil {
		return 0, nil, fmt.Errorf("index out of array6")
	}
	if !bytes.Equal(tailBytes, uw.PacketTail) {
		//if packet.Tail[0] != 0x3C || packet.Tail[1] != 0x3E {
		//log.Println("Invalid packet tail")
		//buf = buf[1:]
		return -1, nil, fmt.Errorf("invalid packet tail")
	}
	packet.Tail = tailBytes

	return 1, packet, nil
}

// SerialReceiver 定义一个接收串口消息的接口
type SerialReceiver interface {
	ReceiveData() ([]byte, error)
}

// OpenPort 打开串口
func (spw *SerialPortWrapper) OpenPort() error {
	port, err := OpenSerialPort(spw.PortName, spw.BaudRate, spw.DataBits, spw.StopBits, spw.MinimumReadSize)
	if err != nil {
		return err
	}
	spw.Port = port
	return nil
}

// SendData 发送数据
func (spw *SerialPortWrapper) SendData(sendData []byte) error {
	if spw.Port == nil {
		err := spw.OpenPort()
		if err != nil {
			return err
		}
	}

	_, err := spw.Port.Write(sendData)
	if err != nil {
		return err
	}
	return nil
	//return SendAndReceiveData(spw.Port, sendData, spw.Timeout)
}

// ReceiveData 从串口接收数据
func (spw *SerialPortWrapper) ReceiveData(length uint) ([]byte, error) {
	if spw.Port == nil {
		err := spw.OpenPort()
		if err != nil {
			return nil, err
		}
	}
	buf := make([]byte, length)
	n, err := spw.Port.Read(buf)
	if err != nil {
		return nil, err
	}
	return buf[:n], nil
}

// OpenSerialPort 打开串口
func OpenSerialPort(portName string, baudRate uint, dataBits uint, stopBits uint, minimumReadSize uint) (io.ReadWriteCloser, error) {
	// 配置串口参数
	options := serial.OpenOptions{
		PortName:        portName,
		BaudRate:        baudRate,
		DataBits:        dataBits,
		StopBits:        stopBits,
		MinimumReadSize: minimumReadSize,
	}

	// 打开串口
	port, err := serial.Open(options)
	if err != nil {
		return nil, err
	}

	return port, nil
}

// SendAndReceiveData 发送和接收数据
func SendAndReceiveData(port io.ReadWriteCloser, sendData []byte, timeout time.Duration) ([]byte, error) {
	//sendData := []byte{0x3c, rsctl, ctl, date, 0x3e}
	// 接收数据
	buf := make([]byte, 128)

	for i := 0; i < 3; i++ {
		n, err := port.Write(sendData)
		if err != nil {
			return nil, err
		}
		log.Printf("Sent %d bytes: %v", n, sendData)

		// 等待300毫秒
		time.Sleep(timeout)

		n, err = port.Read(buf)
		if err != nil {
			return nil, err
		}

		if n > 0 {
			fmt.Printf("Received %d bytes: %v", n, buf[:n])
			return buf[:n], nil
		} else if i == 2 {
			return nil, fmt.Errorf("no response after 3 attempts")
		}
	}

	return nil, nil
}

// FrameNumberGenerator 帧号生成器
func FrameNumberGenerator(startIndex byte, endIndex byte) func() byte {
	current := startIndex
	return func() byte {
		if current > endIndex {
			current = startIndex
		}
		result := current
		current++
		return result
	}
}

func safeAccess(slice []byte, index int) (byte, error) {
	if index < 0 || index >= len(slice) {
		return 0, fmt.Errorf("index out of bounds")
	}
	return slice[index], nil
}

func safeAccessBytes(slice []byte, index int, length int) ([]byte, error) {
	if index < 0 || index >= len(slice) {
		return nil, fmt.Errorf("index out of bounds")
	}
	if index+length > len(slice) {
		return nil, fmt.Errorf("index out of bounds")
	}
	return slice[index : index+length], nil
}
