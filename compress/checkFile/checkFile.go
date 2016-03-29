package checkFile

import (
	"os"
)

func Create(hFile *os.File) {
	var checkSum uint16 = 0
	hFile.Seek(0, os.SEEK_SET)
	buf := make([]byte, 0x100)
	for {
		n, err := hFile.Read(buf)
		if n > 1 {
			for i := 0; i < n>>1; i++ {
				checkSum += uint16(buf[i*2]) + (uint16(buf[i*2+1]) << 8)
			}
		}
		if err != nil {
			if err.Error() == "EOF" {
				break
			} else {
				panic(err)
			}
		}
	}
	hCheck, _ := os.Create("checkFile.chk")
	temp := make([]byte, 2)
	temp[0] = byte(checkSum & 0xFF)
	temp[1] = byte((checkSum >> 8) & 0xFF)
	hCheck.Write(temp)
	hCheck.Close()
}
