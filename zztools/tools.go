package zztools

import (
	"crypto/md5"
	"crypto/rand"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"fmt"
	"github.com/buger/jsonparser"
	"io"
	"os"
	"strconv"
	"sync/atomic"
	"time"
)

var objectIdCounter uint32 = 0

var machineId = readMachineId()

type ObjectId string

func readMachineId() []byte {
	var sum [3]byte
	id := sum[:]
	hostname, err1 := os.Hostname()
	if err1 != nil {
		_, err2 := io.ReadFull(rand.Reader, id)
		if err2 != nil {
			panic(fmt.Errorf("cannot get hostname: %v; %v", err1, err2))
		}
		return id
	}
	hw := md5.New()
	hw.Write([]byte(hostname))
	copy(id, hw.Sum(nil))
	return id
}

// GUID returns a new unique ObjectId.
// 4byte 时间，
// 3byte 机器ID
// 2byte pid
// 3byte 自增ID
func GetGUID() ObjectId {
	var b [12]byte
	// Timestamp, 4 bytes, big endian
	binary.BigEndian.PutUint32(b[:], uint32(time.Now().Unix()))
	// Machine, first 3 bytes of md5(hostname)
	b[4] = machineId[0]
	b[5] = machineId[1]
	b[6] = machineId[2]
	// Pid, 2 bytes, specs don't specify endianness, but we use big endian.
	pid := os.Getpid()
	b[7] = byte(pid >> 8)
	b[8] = byte(pid)
	// Increment, 3 bytes, big endian
	i := atomic.AddUint32(&objectIdCounter, 1)
	b[9] = byte(i >> 16)
	b[10] = byte(i >> 8)
	b[11] = byte(i)
	return ObjectId(b[:])
}

// Hex returns a hex representation of the ObjectId.
// 返回16进制对应的字符串
func (id ObjectId) Hex() string {
	return hex.EncodeToString([]byte(id))
}
func JsonUnsafeGetInt(json []byte, key string) (v int) {
	str, err := jsonparser.GetUnsafeString(json, key)
	v, err = strconv.Atoi(str)
	if err != nil {
		return 0
	}
	return v
}

func StringArrayToIntArray(s []string) (arr []int, err error) {
	if len(s) == 0 {
		return nil, errors.New("string  length == 0")
	}
	arr = make([]int, len(s))
	for pos, v := range s {
		arr[pos], err = strconv.Atoi(v)
		if err != nil {
			return nil, err
		}
	}
	return
}

//定时器
func SetTimer(dura time.Duration, proc func()) {
	go func(d time.Duration, f func()) {
		ticker := time.NewTicker(d)
		for {
			select {
			case <-ticker.C:
				go f()
			}
		}
	}(dura, proc)
}

func GetSecond(stime time.Time) int {
	d := time.Now().Sub(stime).Seconds()
	if d > 1 {
		return int(d)
	} else if d < 1 && d >= 0.5 {
		return 1
	} else {
		return 0
	}
}
