package RLE

import (
	"os"
)

const cp_length uint8 = 4

func Compress(hFile *os.File, hOutput *os.File) {
	hFile.Seek(0, os.SEEK_SET)
	hOutput.Seek(0, os.SEEK_SET)
	buf := make([]byte, 0x1000)
	compressor := rle_compressor(hOutput)
	for {
		n, err := hFile.Read(buf)
		if n > 0 {
			for i := 0; i < n; i++ {
				compressor(buf[i], 0)
			}
		}
		if err != nil {
			if err.Error() == "EOF" {
				compressor(0, 1)
				return
			} else {
				panic(err)
			}
		}
	}
}

type tRLE_block struct {
	head struct {
		arr    [3]byte
		length byte
	}
	length uint16
	buffer [0x10000]byte
}

func rle_compressor(hOutput *os.File) func(input byte, cmd byte) {
	// 重数标记：最高位表示是否为重数
	// 低7位表示长度，当低7位全为1时，表示为扩展重数标记

	var repeat_count uint16 = 0
	var cblock tRLE_block
	cblock.head.length = 0
	cblock.length = 0
	initBlock := func(block *tRLE_block) {
		block.head.length = 0
		block.length = 0
	}
	blockLengthUpdate := func(block *tRLE_block) {
		if block.head.length > 0 { // length计算
			if block.length >= 0x7F {
				block.head.arr[0] = (block.head.arr[0] & 0x80) | 0x7F
				block.head.arr[1] = byte(block.length & 0xFF)
				block.head.arr[2] = byte((block.length >> 8) & 0xFF)
				block.head.length = 3
			} else {
				block.head.arr[0] = (block.head.arr[0] & 0x80) | byte(block.length&0x7F)
				block.head.length = 1
			}
		}
	}
	writeBlock := func(block *tRLE_block) {
		blockLengthUpdate(block)
		hOutput.Write(cblock.head.arr[0:cblock.head.length])
		if cblock.head.arr[0]&0x80 > 0 { // 重数块
			hOutput.Write(cblock.buffer[0:1])
		} else {
			hOutput.Write(cblock.buffer[0:cblock.length])
		}
		//fmt.Println("block:")
		//fmt.Print(cblock.head.length, cblock.head.arr, "\n")
		//fmt.Print(cblock.length, cblock.buffer[0:cblock.length], "\n")
	}
	// cmd: 0 for push 1 for pop
	return func(input byte, cmd byte) {
		if cmd == 1 {
			if cblock.head.length == 0 {
				cblock.head.length = 1
				cblock.head.arr[0] &= 0x7F
			}
			//_slice := make([]byte, cblock.length)
			writeBlock(&cblock)

		} else if cmd == 0 {
			cblock.buffer[cblock.length] = input
			cblock.length++
			if cblock.length < 1 {
				return
			}

			if cblock.length > 0xFFF0 {
				writeBlock(&cblock)
				initBlock(&cblock)
				repeat_count = 0
			}

			if cblock.buffer[cblock.length-1] == cblock.buffer[cblock.length-2] {
				repeat_count++
			} else {
				repeat_count = 0
			}

			if cblock.length == uint16(cp_length) { // 确定数据块属性
				cblock.head.length = 1
				if repeat_count >= uint16(cp_length-1) {
					cblock.head.arr[0] = 0x80 // 重数块
				} else {
					cblock.head.arr[0] = 0x00 // 非重数块
				}
			}

			if cblock.head.length > 0 { // 数据块终止
				if cblock.head.arr[0] >= 0x80 && repeat_count == 0 {
					cblock.length -= 1
					writeBlock(&cblock)
					initBlock(&cblock)
					cblock.buffer[0] = input
					cblock.length++
				} else if cblock.head.arr[0] <= 0x7F && repeat_count == uint16(cp_length-1) {
					cblock.length -= uint16(cp_length - 1)

					cblock.length -= 1
					writeBlock(&cblock)
					initBlock(&cblock)

					cblock.head.arr[0] = 0x80
					cblock.head.length = 1
					cblock.buffer[0] = input
					cblock.length = uint16(cp_length)
					cblock.buffer[cblock.length-1] = input
				}
			}

		}
	}
}

// 解压缩
func Decompress(hInput *os.File, hOutput *os.File) {
	var mark [3]byte
	var length, i uint16
	var err error = nil
	var index uint32 = 0
	hInput.Seek(0, os.SEEK_SET)
	hOutput.Seek(0, os.SEEK_SET)
	for err == nil {

		temp := make([]byte, 1)
		_, err := hInput.Read(temp)
		mark[0] = temp[0]
		if err != nil || mark[0] == 0 {
			break
		}
		if (mark[0] & 0x7F) == 0x7F {
			hInput.Read(temp)
			mark[1] = temp[0]
			hInput.Read(temp)
			mark[2] = temp[0]
			length = uint16(mark[1])
			length += uint16(mark[2]) << 8
		} else {
			length = uint16(mark[0] & 0x7F)
		}
		index += uint32(length)
		if mark[0]&0x80 > 0 { // 重数块
			hInput.Read(temp)
			for i = 0; i < length; i++ {
				hOutput.Write(temp)
			}
		} else { // 非重数块
			for i = 0; i < length; i++ {
				hInput.Read(temp)
				hOutput.Write(temp)
			}
		}
	}
}
