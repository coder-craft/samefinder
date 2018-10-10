package main

import (
	"bytes"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"time"
)

var gp_BackFile_Dir *string
var gp_Filter_Size *int64
var gp_OutFile *string
var gp_Multi_Proc *bool
var g_FilePool = make(map[string][]string)
var g_ErrFile = []string{}

type SameFileData struct {
	FileName string
	AllPath  []string
}
type ReportFile struct {
	SameFile       []SameFileData
	BigFilePath    []string
	TotleFileCount int
	SameFileCount  int
	SameFileSize   int64
	TimeCount      time.Time
}

func main() {
	gp_BackFile_Dir = flag.String("backup-dir", "samefile-560c9cf51", "Move the same file which in current dir to new dir.")
	gp_Filter_Size = flag.Int64("filter-size", 100000000, "Filter the faile size,very big file will make the memory over.")
	gp_OutFile = flag.String("outfile", "samefilefinder.log", "Same file list will be saved.")
	gp_Multi_Proc = flag.Bool("multi", false, "Multithread open may cause the computer no response.")
	flag.Parse()
	rootDir, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		println("Get current path error:", err)
		return
	}
	sameFileSize := int64(0)
	bigPath := []string{}
	startTick := time.Now()
	filepath.Walk(rootDir, func(filePath string, info os.FileInfo, err error) error {
		if err != nil {
			println("File walk error:", err.Error())
			return nil
		}
		if !info.IsDir() && info.Size() < *gp_Filter_Size {
			fileMd5 := MakeMd5File(filePath)
			if len(fileMd5) == 0 {
				g_ErrFile = append(g_ErrFile, filePath)
				println("File:", filePath, " make md5 failed.")
			} else {
				if existPath, ok := g_FilePool[fileMd5]; ok {
					println(filePath, "is repeated.")
					println("Repeated file path:", existPath...)
					sameFileSize += info.Size()
					g_FilePool[fileMd5] = append(g_FilePool[fileMd5], filePath)
				} else {
					g_FilePool[fileMd5] = []string{filePath}
				}
			}
		} else {
			if info.Size() >= *gp_Filter_Size {
				bigPath = append(bigPath, filePath)
			}
		}
		return err
	})
	outFile := &ReportFile{
		BigFilePath: bigPath,
	}
	fileCount := 0
	sameCount := 0
	for _, value := range g_FilePool {
		if len(value) > 1 {
			outFile.SameFile = append(outFile.SameFile, SameFileData{
				FileName: filepath.Base(value[0]),
				AllPath:  value,
			})
			sameCount += 1
		}
		fileCount += len(value)
	}
	outFile.TotleFileCount = fileCount
	outFile.SameFileCount = sameCount
	outFile.SameFileSize = sameFileSize
	outFile.TimeCount = time.Unix(0, time.Now().Sub(startTick).Nanoseconds())
	outFileBuff, err := json.Marshal(outFile)
	if err != nil {
		println("Marshal out file error:", err)
	}
	var formatOut bytes.Buffer
	err = json.Indent(&formatOut, outFileBuff, "", "\t")
	outFileName := fmt.Sprintf("Samefilelist-%v.log", time.Now().Format("2006-01-02-15-04-05"))
	ioutil.WriteFile(outFileName, formatOut.Bytes(), os.ModeType)
}
func MakeMd5File(filename string) string {
	readBuff, readErr := ioutil.ReadFile(filename)
	if readErr != nil {
		println("Read file ", filename, " error:", readErr)
		return ""
	}
	buff := md5.New()
	buff.Write(readBuff)
	return hex.EncodeToString(buff.Sum(nil))
}
func FileReadOnly(m os.FileMode) bool {
	flag := m & os.FileMode(292)
	return uint32(flag) == uint32(292)
}
func FileWriteOnly(m os.FileMode) bool {
	flag := m & os.FileMode(146)
	return uint32(flag) == uint32(146)
}
