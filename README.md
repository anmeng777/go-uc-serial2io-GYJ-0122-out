目前只有控制输出功能，输入指定私有协议即可进行控制

安装方式：go get github.com/anmeng777/go-uc-serial2io-GYJ-0122-out

引入："github.com/anmeng777/go-uc-serial2io-GYJ-0122-out/GYJ_0122"

调用方式：
  spw := &GYJ_0122.SerialPortWrapper{
			PortName:        "COM1",
			BaudRate:        115200,
			DataBits:        8,
			StopBits:        1,
			MinimumReadSize: 4,
			Timeout:         300 * time.Millisecond,
		}
  getNextFrameNumber := GYJ_0122.FrameNumberGenerator(0x10, 0x19)

  spw.SendData(getNextFrameNumber(), 0x10, 0x01)
