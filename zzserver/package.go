package zzserver

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
)

// Packet 封包
func Packet(message []byte) []byte {
	return append(IntToBytes(len(message)), message...)
}

// IntToBytes 整形转换成字节
func IntToBytes(n int) []byte {
	x := int32(n)
	bytesBuffer := bytes.NewBuffer([]byte{})
	binary.Write(bytesBuffer, binary.LittleEndian, x)
	// fmt.Println(bytesBuffer.Bytes())
	return bytesBuffer.Bytes()
}

// Unpack 解包
func Unpack(buffer []byte, readerChannel chan []byte) []byte {
	// fmt.Println("解包:buffer:", buffer)
	length := len(buffer)
	// fmt.Println("解包:length:", length)
	var i int
	for i = 0; i < length; i = i + 1 {
		if length < i+4 {
			break
		}
		messageLength := BytesToInt(buffer[i : i+4])
		if messageLength < 0 || length < i+4+messageLength {
			break
		}
		data := buffer[i+4 : i+4+messageLength]
		readerChannel <- data

		i += 4 + messageLength - 1
	}
	if i == length {
		return make([]byte, 0)
	}
	return buffer[i:]
}

// BytesToInt 字节转换成整形
func BytesToInt(b []byte) int {
	bytesBuffer := bytes.NewBuffer(b)
	var x int32
	_ = binary.Read(bytesBuffer, binary.LittleEndian, &x)
	return int(x)
}

// Packet2 封包2
func Packet2(message []byte) ([]byte, error) {
	var length = int32(len(message))
	var pkg = new(bytes.Buffer)
	//tou
	err := binary.Write(pkg, binary.LittleEndian, length)
	if err != nil {
		return nil, err
	}
	err = binary.Write(pkg, binary.LittleEndian, message)
	if err != nil {
		return nil, err
	}
	return pkg.Bytes(), nil
}

// Unpack2 解包2
func Unpack2(reader *bufio.Reader) ([]byte, error) {
	//buffer返回缓冲区中现有的可读字节数
	bfed := int32(reader.Buffered())
	if bfed <= 4 {
		return nil, errors.New("Buffered()<=4")
	}
	//返回前4个字节
	lengthByte, err := reader.Peek(4)
	if err != nil {
		return nil, err
	}
	//创建buff
	lengthBuff := bytes.NewBuffer(lengthByte)
	var length int32
	//读取内容到 length
	err = binary.Read(lengthBuff, binary.LittleEndian, &length)
	if err != nil {
		return nil, err
	}
	fmt.Println("reader.Buffered() =>", bfed)
	if bfed < length+4 {
		return nil, err
	}
	//读取真正的消息数据
	pack := make([]byte, int(4+length))
	n, err := reader.Read(pack)
	if err != nil {
		return nil, err
	}
	if n != int(4+length) {
		return nil, errors.New("length error")
	}
	return pack[4:], nil
}
