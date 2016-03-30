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
func (p *h_node) Encode_dict(dict *[]h_dict, length uint8, code uint32) {
	if p.left != nil {
		p.left.Encode_dict(dict, length+1, code|(0x80000000>>length))
		p.right.Encode_dict(dict, length+1, code&(^(0x80000000 >> length)))
		return
	}
	*dict = append(*dict, h_dict{p.value, length, code})
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

type huffman_block struct {
	blockLength uint16 // 块长度
	dictLength  uint16 // 字典长度
	// 数据长度=blockLength-dictLength
	buffer [block_size]byte
}

func Compress(hFile *os.File, hOutput *os.File) {
	//var cblock huffman_block
	var block_index uint16 = 0
	build_huffman_tree := func(buffer []byte) {
		leaf := make(nodelist, 0, 256)
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

				dict_d := make([]uint8, 0, 256)
				leaf[0].Decode_dict(&dict_d)
				dict_e := make([]h_dict, 0, 256)
				leaf[0].Encode_dict(&dict_e, 0, 0)
				/*
					for k, v := range dict_d {
						if k&0x1 > 0 {
							fmt.Println(v)
						} else {
							fmt.Print(k/2, ":", v, ",")
						}
					}
					for _, v := range dict_e {
						fmt.Printf("[%#02X,%d,%s]\n", v.n, v.l, string([]rune(fmt.Sprintf("%032b", v.bits))[:v.l]))
					}
				*/
				break
			}
		}
	}
	hFile.Seek(0, os.SEEK_SET)
	hOutput.Seek(0, os.SEEK_SET)
	buf := make([]byte, block_size)
	for {
		n, err := hFile.Read(buf)
		if n > 0 {
			build_huffman_tree(buf[:n])
		}
		if err != nil {
			if err.Error() == "EOF" {
				return
			} else {
				panic(err)
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
