package main

import (
	"bytes"
	"github.com/anmeng777/go-uc-serial2io-GYJ-0122-out/GYJ_0122"
	"log"
	"time"
)

const (
	HelpSignal = 0x30
	TalkStart  = 0x0E
)

func main1() {
	//b1 := []byte("hello world")
	//b2 := []byte("l2o")
	//
	//// 使用 bytes.Index() 函数
	//index := bytes.Index(b1, b2)
	//fmt.Println(index) // 2

	buf := []byte{0x3c, 0x00, 0x00, 0x00, 0x3e, 0x3c, 0x00, 0x00, 0x00, 0x3e}
	index := bytes.Index(buf, []byte{0x3e})
	log.Printf("index: %v", buf[index+1:])
}

func main() {
	spw := &GYJ_0122.SerialPortWrapper{
		PortName:        "COM3",
		BaudRate:        9600,
		DataBits:        8,
		StopBits:        1,
		MinimumReadSize: 4,
		Timeout:         300 * time.Millisecond,
	}
	uw := &GYJ_0122.UnpackWrapper{
		PacketHeader: []byte{0x3c},
		LengthIndex:  0,
		LengthSize:   0,
		CommandIndex: 2,
		CommandSize:  1,
		DataIndex:    3,
		DataSize:     4,
		PacketTail:   []byte{0x3e},
	}
	MinimumPacket := len(uw.PacketHeader) + uw.LengthSize + uw.CommandSize + len(uw.PacketTail)
	var getNextFrameNumber = GYJ_0122.FrameNumberGenerator(0x30, 0x39)

	var buf []byte

	for {
		tempBuf, err := spw.ReceiveData(1024)
		if err != nil {
			continue
		}
		if len(tempBuf) == 0 {
			continue
		}
		buf = append(buf, tempBuf...)

		log.Printf("接收到: %v", tempBuf)
		log.Printf("全部未处理: %v", buf)

		for {
			if len(buf) < MinimumPacket {
				break
			}

			status, packet, err := uw.Unpack(buf)
			if status == 0 {
				break
			}
			if err != nil {
				buf = buf[1:]
				continue
			}

			switch packet.Command[0] {
			case HelpSignal:
				//time.Sleep(200 * time.Millisecond)
				log.Printf("Received help signal")
				if packet.Data[0] == 0x00 || packet.Data[1] == 0x00 {
					spw.SendData([]byte{0x3c, getNextFrameNumber(), 0x31, 0x00, 0x3e})
				} else {
					spw.SendData([]byte{0x3c, getNextFrameNumber(), 0x31, 0x01, 0x3e})
				}

			case 0x31:
				log.Printf("Received 0x31")
			default:
				log.Printf("Unknown command: %x", packet.Command)
			}
			if uw.PacketTail != nil {
				log.Printf("解包: %v", buf[:bytes.Index(buf, uw.PacketTail)+1])
				buf = buf[bytes.Index(buf, uw.PacketTail)+1:]
			} else {
				log.Printf("解包: %v", buf[:len(uw.PacketHeader)+uw.LengthSize+1+int(packet.Length)+len(uw.PacketTail)+1])
				buf = buf[len(uw.PacketHeader)+uw.LengthSize+1+int(packet.Length)+len(uw.PacketTail):]
			}
		}

		// 增加延时
		time.Sleep(100 * time.Millisecond)
	}
}
