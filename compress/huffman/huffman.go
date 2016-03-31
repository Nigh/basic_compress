// huffman.go
package huffman

import (
	"fmt"
	"os"
	"sort"
)

const block_size uint16 = 0x8000
const eob_mark uint16 = 0x1000

type h_node struct {
	left   *h_node
	right  *h_node
	weight uint16
	value  uint16 // 此处应该能表示任意内容
}
type nodelist []*h_node

//为*h_node添加String()方法，便于输出
func (p *h_node) String() string {
	return fmt.Sprintf("[%p]<-this->[%p],\t[%d],\t[%#02X]", p.left, p.right, p.weight, p.value)
}

//为*h_node添加DotString()方法，输出dot格式字符串
func (p *h_node) DotString() string {
	var str string
	if p.left != nil {
		str += fmt.Sprintf("node_%p[label=\"{%d}\"];\n", p, p.weight)
		str += fmt.Sprintf("node_%p->node_%p[headlabel=\"L\"];\n", p, p.left)
		str += fmt.Sprintf("node_%p->node_%p[headlabel=\"R\"];\n", p, p.right)
	} else {
		str += fmt.Sprintf("node_%p[label=\"%d|%#02X\"];\n", p, p.weight, p.value)
	}
	return str
}

// huffman dict struct
type h_dict struct {
	n    uint16 // content
	l    uint8  // length(bit)
	bits uint32
}

// huffman dict struct
type bits struct {
	l    uint8 // length(bit)
	bits uint32
}

/*
每个字典项占2bytes,分为l和r,
如果l!=r,则表示此项非叶子项,
	其左右子枝分别为第l和r项,下标索引即为2*l和2*r
	如果l的索引是它本身,且r==0,则此项为block结束标志
如果l==r,则表示此项为叶子项,
	l=其编码内容
解码时,从根节点开始跟随bit流中的1或0,进行跳转,直到叶子,得到解码内容
*/
//为*h_node添加Decode_dict()方法，输出以此node为root的解码字典
func (p *h_node) Decode_dict(dict *[]uint8) int {
	pos := len(*dict)
	if p.left != nil {
		*dict = append(*dict, 0, 0) // 占位
		left := p.left.Decode_dict(dict)
		right := p.right.Decode_dict(dict)
		(*dict)[pos] = uint8(left / 2)
		(*dict)[pos+1] = uint8(right / 2)
	} else {
		if p.value == eob_mark {
			*dict = append(*dict, uint8(pos/2), 0)
		} else {
			*dict = append(*dict, uint8(p.value), uint8(p.value))
		}
	}
	return pos
}

//为*h_node添加Encode_dict()方法，输出以此node为root的编码字典
// 1 for left, 0 for right
func (p *h_node) Encode_dict(dict *map[uint16]bits, bin bits) {
	if p.left != nil {
		p.left.Encode_dict(dict, bits{bin.l + 1, bin.bits | (0x80000000 >> bin.l)})
		p.right.Encode_dict(dict, bits{bin.l + 1, bin.bits & (^(0x80000000 >> bin.l))})
		return
	}
	(*dict)[p.value] = bits{bin.l, bin.bits}
}

func traverse(p *h_node, hOutput *os.File) {
	if p.left != nil {
		traverse(p.left, hOutput)
	}
	if p.right != nil {
		traverse(p.right, hOutput)
	}
	hOutput.WriteString(p.DotString())
}

// 输出node树的dot格式字符串
func (p *h_node) Dot(index string) {
	hOutput, _ := os.Create("huffman" + index + ".dot")
	hOutput.WriteString("digraph structs {\nnode[shape=record];\n")
	traverse(p, hOutput)
	hOutput.WriteString("}\n")
	hOutput.Close()
}

func Compress(hFile *os.File, hOutput *os.File) {
	var block_index uint16 = 0
	build_huffman_tree := func(buffer []byte) (leaf nodelist) {
		leaf = make(nodelist, 0, 256)
		// 1.统计
		var table [256]uint16
		for _, v := range buffer {
			table[v]++
		}
		// 2.叶子初始化
		for k, v := range table {
			if v > 0 {
				t_leaf := h_node{nil, nil, v, uint16(k)}
				leaf = append(leaf, &t_leaf)
			}
		}
		// end of block mark
		t_eob := h_node{nil, nil, 1, eob_mark}
		leaf = append(leaf, &t_eob)
		for {
			// 3.排序
			sort.Sort(nodelist(leaf))
			// 4.种树
			if len(leaf) > 1 {
				tleft := leaf[0]
				tright := leaf[1]
				root := h_node{tleft, tright, tleft.weight + tright.weight, 0}
				leaf = leaf[1:]
				leaf[0] = &root
			} else {
				leaf[0].Dot(fmt.Sprintf("%d", block_index))
				block_index += 1
				break
			}
		}
		return leaf
	}
	huffman_compress := func(hOutput *os.File, r_buf []byte, root *nodelist) {
		// byte拼接
		// n 表示最后一个元素占用的bit数
		bits2byte := func(arr []byte, n uint8, bin bits) ([]byte, uint8) {
			for bin.l > 0 {
				if n == 0 {
					arr = append(arr, 0)
				}
				arr[len(arr)-1] |= byte(bin.bits >> uint(n+24))
				temp := 8 - n
				if temp > bin.l {
					n += bin.l
				} else {
					n = 0
				}
				if bin.l > temp {
					bin.l -= temp
				} else {
					bin.l = 0
				}
				bin.bits <<= temp
			}
			ret := arr
			return ret, n
		}
		// 生成编码字典
		dict_e := make(map[uint16]bits)
		(*root)[0].Encode_dict(&dict_e, bits{0, 0})
		// 生成解码字典
		dict_d := make([]uint8, 0, 256)
		(*root)[0].Decode_dict(&dict_d)
		// 写入解码字典
		hOutput.Write([]byte{byte(len(dict_d) / 2)})
		hOutput.Write(dict_d)
		var n uint8 = 0
		var t_bytes []byte // temp byte slice
		for _, v := range r_buf {
			bin := dict_e[uint16(v)]
			t_bytes, n = bits2byte(t_bytes, n, bin)
		}
		t_bytes, n = bits2byte(t_bytes, n, dict_e[eob_mark])
		hOutput.Write(t_bytes)
	}
	hFile.Seek(0, os.SEEK_SET)
	hOutput.Seek(0, os.SEEK_SET)
	buf := make([]byte, block_size)
	for {
		n, err := hFile.Read(buf)
		if n > 0 {
			tree := build_huffman_tree(buf[:n])
			huffman_compress(hOutput, buf[:n], &tree)
		}
		if err != nil {
			if err.Error() == "EOF" {
				hOutput.Close()
				return
			} else {
				panic(err)
			}
		}
	}
}

func Decompress(hFile *os.File, hOutput *os.File) {
	hFile.Seek(0, os.SEEK_SET)
	hOutput.Seek(0, os.SEEK_SET)
	var decode_ptr uint16 = 0 // 指向第n个节点(索引为2*n)
	var debug_cnt uint32 = 0
	huffman_decode := func(bytes byte, dict []byte) bool {
		var i uint8
		for i = 0; i < 8; i++ {
			if bytes&(0x80>>i) > 0 {
				decode_ptr = uint16(dict[decode_ptr*2])
			} else {
				decode_ptr = uint16(dict[decode_ptr*2+1])
			}

			if dict[decode_ptr*2] == dict[decode_ptr*2+1] { // decode
				hOutput.Write([]byte{dict[decode_ptr*2]})
				debug_cnt++
				decode_ptr = 0
			} else if uint16(dict[decode_ptr*2]) == decode_ptr && dict[decode_ptr*2+1] == 0 { // block end
				decode_ptr = 0
				return true
			}
		}
		return false
	}
	for {
		dictLength := make([]byte, 1)
		// 读取字典长度
		_, err := hFile.Read(dictLength)
		if err != nil {
			if err.Error() == "EOF" {
				hOutput.Close()
				return
			} else {
				panic(err)
			}
		}
		// 读取字典
		dict := make([]byte, uint16(dictLength[0])*2)
		_, err = hFile.Read(dict)
		// 解码
		buffer := make([]byte, 0x1)
		for {
			_, err = hFile.Read(buffer)
			if err != nil {
				panic(err)
			}
			if huffman_decode(buffer[0], dict) {
				break
			}
		}
	}
}

func (I nodelist) Len() int {
	return len(I)
}
func (I nodelist) Less(i, j int) bool {
	return I[i].weight < I[j].weight
}
func (I nodelist) Swap(i, j int) {
	I[i], I[j] = I[j], I[i]
}
