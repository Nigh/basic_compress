// huffman.go
package huffman

import (
	"fmt"
	"os"
	"sort"
)

const block_size uint16 = 0x8000

type h_node struct {
	left   *h_node
	right  *h_node
	weight uint16
	value  byte
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

//为*h_node添加Dict()方法，输出以此node为root的解码字典
func (p *h_node) Dict(dict *[]uint8) {
	if p.left != nil {
		*dict = append(*dict, len(dict)+2, len(dict)+4)
		p.left.Dict(dict)
		p.right.Dict(dict)
	} else {
		*dict = append(*dict, p.value, p.value)
	}
}

// huffman dict struct
type h_dict struct {
	n    byte  // content
	l    uint8 // length(bit)
	bits uint32
}

// 1 for left, 0 for right
func get_huffman_dict(p *h_node, dict *[]h_dict, length uint8, code uint32) {
	if p.left != nil {
		get_huffman_dict(p.left, dict, length+1, code|(0x80000000>>length))
		get_huffman_dict(p.right, dict, length+1, code&(^(0x80000000 >> length)))
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
	//fmt.Print(p.String())
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
		//fmt.Println(table)
		// 2.叶子初始化
		for k, v := range table {
			if v > 0 {
				t_leaf := h_node{nil, nil, v, byte(k)}
				leaf = append(leaf, &t_leaf)
			}
		}
		//for _, v := range leaf {
		//	fmt.Println(v.String())
		//}
		for {
			// 3.排序
			sort.Sort(nodelist(leaf))
			//fmt.Println("-----------After Sort-----------")
			//for _, v := range leaf {
			//	fmt.Println(v.String())
			//}
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
				/*
					dict := make([]h_dict, 0, 256)
					get_huffman_dict(leaf[0], &dict, 0, 0)
					fmt.Println("")
					for _, v := range dict {
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
