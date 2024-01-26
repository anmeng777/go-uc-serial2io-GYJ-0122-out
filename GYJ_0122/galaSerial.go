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
	Header  []byte
	Length  uint32
	Command byte
	Data    []byte
	Tail    []byte
}

type UnpackWrapper struct {
	PacketHeader []byte
	LengthSize   int
	CommandSize  int
	PacketTail   []byte
}

func (uw *UnpackWrapper) Unpack(buf []byte) (int, *Packet, error) {
	packet := &Packet{}
	headerBytes := make([]byte, len(uw.PacketHeader))
	for i := 0; i < len(uw.PacketHeader); i++ {
		headerBytes[i] = buf[i]
	}
	packet.Header = headerBytes
	if !bytes.Equal(packet.Header, uw.PacketHeader) {
		log.Println("Invalid packet header")
		//buf = buf[1:]
		return -1, nil, fmt.Errorf("invalid packet header")
	}
	lengthBytes := make([]byte, uw.LengthSize)
	for i := 0; i < uw.LengthSize; i++ {
		lengthBytes[i] = buf[len(uw.PacketHeader)+i]
	}
	packet.Length = binary.LittleEndian.Uint32(lengthBytes)
	if packet.Length > 100 {
		log.Println("Invalid packet length")
		//buf = buf[1:]
		return -1, nil, fmt.Errorf("invalid packet length")
	}
	packet.Command = buf[len(uw.PacketHeader)+uw.LengthSize]

	if len(buf) < len(uw.PacketHeader)+uw.LengthSize+int(packet.Length)+len(uw.PacketTail) {
		return 0, nil, fmt.Errorf("invalid packet length")
	}

	packet.Data = make([]byte, packet.Length)
	for i := uint32(0); i < packet.Length; i++ {
		packet.Data[i] = buf[len(uw.PacketHeader)+uw.LengthSize+1+int(i)]
	}

	// Read and check tail
	packet.Tail = make([]byte, len(uw.PacketTail))
	for i := 0; i < len(uw.PacketTail); i++ {
		packet.Tail[i] = buf[len(uw.PacketHeader)+uw.LengthSize+1+int(packet.Length)+i]
	}

	if !bytes.Equal(packet.Tail, uw.PacketTail) {
		//if packet.Tail[0] != 0x3C || packet.Tail[1] != 0x3E {
		log.Println("Invalid packet tail")
		//buf = buf[1:]
		return -1, nil, fmt.Errorf("invalid packet tail")
	}

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
