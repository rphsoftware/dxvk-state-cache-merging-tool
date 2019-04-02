package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/base64"
	"encoding/binary"
	"fmt"
	"io"
	"os"
)

func check(e error) {
	if e != nil {
		panic(e)
	}
}

type Entry struct {
	data []byte
}
type Header struct {
	magic []byte
	ver uint32
	len uint32
}
func main() {
	if len(os.Args) < 3 {
		panic("Not enough arguments. Usage: ./merge <list of files separated with space> <target file>")
	}

	numFiles := len(os.Args) - 2 // os.Args[0] is path to this program, the last item is the target file

	files := make([]string, numFiles)

	for i := 0; i < numFiles; i++ {
		files[i] = os.Args[i+1]
	}

	fmt.Println("Reading caches...")
	hdr := Header{}
	entries := make(map[string]Entry)

	first := true

	for i := 0; i < len(files); i++ {
		fmt.Println(files[i], " ")
		fd, err := os.Open(files[i])
		check(err)
		magicNum := make([]byte, 4)
		fd.Read(magicNum)
		sMagicNum := string(magicNum)
		if sMagicNum == "DXVK" {
			fmt.Println("| Magic number OK")
		} else {
			panic("Invalid magic number")
		}
		ver := make([]byte, 4)
		fd.Read(ver)
		nVer := binary.LittleEndian.Uint32(ver)
		fmt.Println("| Version: ", nVer)
		eSize := make([]byte, 4)
		fd.Read(eSize)
		neSize := binary.LittleEndian.Uint32(eSize)
		fmt.Println("| Entry length: ", neSize)

		cnt := 0
		if first == false {
			if hdr.len != neSize {
				panic("Entry length mismatch!")
			}
			if hdr.ver != nVer {
				panic("Version mismatch!")
			}
		}

		hdr.magic = magicNum
		hdr.len = neSize
		hdr.ver = nVer
		first = false
		for {
			entry := make([]byte, neSize)
			_, err := fd.Read(entry)
			if err == io.EOF {
				break
			} else {
				check(err)
				e := Entry{entry}
				h := sha256.Sum256(entry)
				slice := h[:]
				hash := base64.StdEncoding.EncodeToString(slice)
				entries[hash] = e
				cnt++
			}
		}

		fmt.Println("| Loaded ", cnt, " entries")

		fd.Close()
	}

	fmt.Println("Merging...")
	fd, err := os.OpenFile(os.Args[len(os.Args)-1], os.O_CREATE|os.O_RDWR, 0644)
	check(err)

	fd.Write(hdr.magic)
	buf := new(bytes.Buffer)

	binary.Write(buf, binary.LittleEndian, hdr.ver)
	binary.Write(buf, binary.LittleEndian, hdr.len)

	fd.Write(buf.Bytes())

	entriesCnt := 0

	for _,v := range entries {
		fd.Write(v.data)
		entriesCnt++
	}

	fmt.Println("Written ", entriesCnt, " entries")

	defer fd.Close()
}
