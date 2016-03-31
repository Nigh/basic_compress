package main

import (
	"flag"
	"fmt"
	"os"

	"./RLE"
	"./checkFile"
	hm "./huffman"
)

func main() {
	var hFile, hOutput *os.File
	var err error
	method := flag.String("m", "rle", "compress method")
	cd := flag.String("d", "comp", "compress or decompress")
	inputFile := flag.String("input", "", "input file path")
	outputFile := flag.String("output", "", "output file path")
	flag.Parse()
	if len(*inputFile) > 0 {
		hFile, err = os.Open(*inputFile)
		isErr(err)
	} else {
		printReadme()
		os.Exit(1)
		return
	}
	if len(*outputFile) == 0 {
		*outputFile = *inputFile + ".output"
	}

	switch *method {
	case "rle":
		fmt.Println("RLE method.\n")
		if *cd == "comp" {
			fmt.Println("Start compress")
			hOutput, err = os.Create(*outputFile)
			isErr(err)
			RLE.Compress(hFile, hOutput)
			isErr(err)
			hOutput.Close()
			fmt.Println("Compress completed!")
		} else if *cd == "dec" {
			fmt.Println("Start decompress")
			hOutput, err = os.Create(*outputFile)
			isErr(err)
			RLE.Decompress(hFile, hOutput)
			hOutput.Close()
			fmt.Println("Decompress completed!")
		} else if *cd == "check" {
			fmt.Println("Start make checkfile")
			checkFile.Create(hFile)
			fmt.Println("Checkfile make completed!")
		} else {
			printReadme()
			os.Exit(1)
			return
		}

	case "lz":
		fmt.Println("LZ is under developing.")
	case "huffman":
		if *cd == "comp" {
			hOutput, err = os.Create(*outputFile)
			isErr(err)
			hm.Compress(hFile, hOutput)
			hOutput.Close()
		} else if *cd == "dec" {
			fmt.Println("Start decompress")
			hOutput, err = os.Create(*outputFile)
			isErr(err)
			hm.Decompress(hFile, hOutput)
			hOutput.Close()
			fmt.Println("Decompress completed!")
		}
	default:
		fmt.Println("invalid method name.")
		os.Exit(1)
		return
	}
	hFile.Close()
	os.Exit(0)
}

func isErr(err error) {
	if err != nil {
		panic(err)
		os.Exit(-2)
	}
}

func printReadme() {
	fmt.Println("Usage:compress [-input <inputFile>] [-output <outputFile>] [-d <comp|dec|check>] [-m <method>]")
	fmt.Println("\n-c: comp for compress, dec for decompress, check for make checkfile")
	fmt.Println("if ignored, compress by default.")
	fmt.Println("\nSupporting method: [rle] [lz]")
	fmt.Println("if no method is assigned, RLE will be used.")
	fmt.Println("\nUsage sample: compress -input qwer.bin -output qwer_comp.bin -d comp -m rle")
}
