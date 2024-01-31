package main

import (
	"github.com/anmeng777/go-uc-serial2io-GYJ-0122-out/GYJ_0122"
	"log"
	"time"
)

const (
	HelpSignal = 0x30
	TalkStart  = 0x0E
)

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
				log.Printf("Received help signal %v", packet.Data)
			default:
				log.Printf("Unknown command: %x", packet.Command)
			}
			buf = buf[len(uw.PacketHeader)+uw.LengthSize+1+int(packet.Length)+len(uw.PacketTail):]
		}

		// 增加延时
		time.Sleep(100 * time.Millisecond)
	}
}
