package GYJ_0122

import (
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
func (spw *SerialPortWrapper) SendData(rsctl byte, ctl byte, date byte) ([]byte, error) {
	if spw.Port == nil {
		err := spw.OpenPort()
		if err != nil {
			return nil, err
		}
	}
	return SendAndReceiveData(spw.Port, rsctl, ctl, date, spw.Timeout)
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
func SendAndReceiveData(port io.ReadWriteCloser, rsctl byte, ctl byte, date byte, timeout time.Duration) ([]byte, error) {
	sendData := []byte{0x3c, rsctl, ctl, date, 0x3e}
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
