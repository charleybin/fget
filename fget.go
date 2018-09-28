package main

import (
	"errors"
	"fget/utils"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"
	"time"
)

const bufferSize = 4096
const defaultThreadCount = 1

const CHAN_CODE_SUCC = -1
const CHAN_CODE_ERR = -2

var mainLog = utils.GetLogger("main")

var downloadUrl = flag.String("u", "", "string value for Download URL")
var threadCount = flag.Int("t", defaultThreadCount, "integer value for Thread-Count")
var finalFileName = flag.String("f", "", "string value for output File-Name")
var logFile = flag.Bool("l", false, "boolean value for enable File Log")
var progress = flag.Bool("p", true, "boolean value for show progress")
var debug = flag.Bool("v", false, "boolean value for show debug info")

func printHelp() {
	fmt.Println("-u \"[URL]\" -t [Thread Count] -l [File log] -p [show progress]")
	fmt.Println("File Get Tool. Copy Right reserved by charleybin@outlook.com")
}

type RangePart struct {
	ThreadIndex int
	UrlAddr     string
	Start       int64
	End         int64
	FileName    string
}

type ProgressPart struct {
	Part       RangePart
	CH         (chan int64)
	CurentSize int64
}

func getFileLength(url string) (int64, string) {
	headers := make(map[string]string, 5)
	// headers["Range"] = "bytes=1000000-"

	resp, err := utils.HttpGet(url, headers)
	if nil != err {
		mainLog.Error("request failed", err)
		return -1, ""
	}
	defer resp.Body.Close()

	if *finalFileName != "" {
		return resp.ContentLength, *finalFileName
	}

	fileName := resp.Header.Get("Content-Disposition")

	if fileName == "" {
		lastIndex := strings.LastIndex(url, "/")
		lastEqual := strings.LastIndex(url, "=")
		lastQuestion := strings.LastIndex(url, "?")
		if lastQuestion > 0 && lastQuestion < lastIndex {
			lastIndex = lastQuestion
		}
		if lastEqual > 0 && lastEqual < lastIndex {
			lastIndex = lastEqual
		}
		if lastIndex > 0 {
			fileName = url[lastIndex:]
		} else {
			fileName = fmt.Sprint(time.Now().Unix())
		}
	}

	if strings.LastIndex(fileName, "/") != -1 {
		fileName = fileName[strings.LastIndex(fileName, "/")+1:]
		fileName = strings.TrimSpace(fileName)
	}

	return resp.ContentLength, fileName
}

func getRange(urlAddr string, totalLength int64, split int, fileName string) ([]RangePart, int) {
	splitCount := int64(split)

	if totalLength%splitCount == 0 {
		totalLength = totalLength
	}
	if totalLength < bufferSize {
		split = 1
		splitCount = int64(split)
	}

	averageLen := totalLength / splitCount
	averageTotal := averageLen * int64(split)
	additionSplit := averageTotal < totalLength-1

	if split > 1 && additionSplit {
		split = split + 1
	}
	//	dataList := make([]RangePart, split)
	dataList := []RangePart{}

	for i := 0; i < split; i++ {
		partBegin := averageLen * int64(i)
		partEnd := averageLen*int64(i+1) - 1
		if partEnd > totalLength || i+1 == split {
			partEnd = totalLength - 1
		}

		data := RangePart{
			ThreadIndex: i,
			UrlAddr:     urlAddr,
			Start:       partBegin,
			End:         partEnd,
			FileName:    fmt.Sprintf("%s_%d", fileName, i),
		}
		dataList = append(dataList, data)

		mainLog.Debug("Begin:", data.Start, ",partEnd:", data.End, ",fileName:", data.FileName)
	}

	return dataList, split
}

func CopyBuffer(ch chan<- int64, dst io.Writer, src io.Reader, buf []byte) (written int64, err error) {
	for {
		nr, er := src.Read(buf)
		if nr > 0 {
			nw, ew := dst.Write(buf[0:nr])
			if nw > 0 {
				written += int64(nw)
				break
				//				ch <- int64(nw)
			}
			if ew != nil {
				err = ew
				break
			}
			if nr != nw {
				err = io.ErrShortWrite
				break
			}
		}
		if er != nil {
			if er != io.EOF {
				err = er
			}
			break
		}
	}
	return written, err
}

func downloadFileByRange(segSize chan<- int64, partInfo RangePart) error {
	rangeValue := fmt.Sprintf("bytes=%d-%d", partInfo.Start, partInfo.End)
	headers := make(map[string]string, 5)
	headers["Range"] = rangeValue

	mainLog.Debug("downloadFileByRange:", rangeValue)

	resp, err := utils.HttpGet(partInfo.UrlAddr, headers)
	if nil != err {
		mainLog.Error("request failed", err)
		return err
	}
	defer resp.Body.Close()

	cleanExistFile(partInfo.FileName)

	file, err := os.Create(partInfo.FileName)
	if err != nil {
		return err
	}
	//	defer file.Close()

	byteData := make([]byte, bufferSize)

	readedLength := int64(0)
	totalLength := partInfo.End - partInfo.Start
	for readedLength < totalLength {
		n, err := CopyBuffer(segSize, file, resp.Body, byteData)
		if err != nil {
			file.Close()
			segSize <- CHAN_CODE_ERR
			return err
		}
		readedLength = readedLength + n
		segSize <- n
		mainLog.Debug("readedLength:", readedLength)
		//		if *progress {
		//			pct := int64(readedLength * 100 / totalLength)
		//			fmt.Fprintf(os.Stdout, "%s,%3d", partInfo.FileName, pct)
		//		}
	}

	file.Close()
	segSize <- CHAN_CODE_SUCC

	return nil
}

func cleanExistFile(fn string) {
	tfp, terr := os.Open(fn)
	if terr == nil {
		tfp.Close()
		os.Remove(fn)
	}
}

func combineFiles(finalName string, rangeList []RangePart, listSize int) error {
	if listSize == 1 {
		return os.Rename(rangeList[0].FileName, finalName)
	}
	cleanExistFile(finalName)

	ffp, ferr := os.Create(finalName)
	if ferr != nil {
		return ferr
	}
	defer ffp.Close()
	for _, rangeSegement := range rangeList {
		fp, err := os.Open(rangeSegement.FileName)
		if err != nil {
			return err
		}
		_, cpErr := io.Copy(ffp, fp)
		fp.Close()
		if cpErr != nil {
			return cpErr
		}
	}

	for _, rangeSegement := range rangeList {
		err := os.Remove(rangeSegement.FileName)
		if err != nil {
			fmt.Println("delete file error:", rangeSegement.FileName, err)
		}
	}
	return nil
}

func getFile(urlAddr string, split int) (error, string) {
	if split < 1 || split > 100 {
		return errors.New("exceed limit 100"), ""
	}

	length, fileName := getFileLength(urlAddr)

	if length <= 0 {
		mainLog.Error("download failed")

		return errors.New("download failed"), ""
	}
	mainLog.Debug("Content-Length:", length)

	rangeList, listSize := getRange(urlAddr, length, split, fileName)

	mainLog.Debug("fileName:", fileName, ",length:", length, ",listSize:", listSize, "rangeList:", rangeList)

	chList := []*ProgressPart{}

	for _, rangeSegement := range rangeList {
		ch := make(chan int64)
		go downloadFileByRange(ch, rangeSegement)
		prog := ProgressPart{CH: ch, Part: rangeSegement}
		chList = append(chList, &prog)
	}

	downloadSize := int64(0)
	for downloadSize < length && len(chList) > 0 {
		for n, prog := range chList {
			loadedSize := <-prog.CH
			if loadedSize < 0 {
				chList = append(chList[:n], chList[n+1:]...)
				break
			} else {
				downloadSize = downloadSize + loadedSize
				prog.CurentSize = prog.CurentSize + loadedSize
				totalLength := prog.Part.End - prog.Part.Start

				if *progress {
					pct := int64(prog.CurentSize * 100 / totalLength)
					fmt.Fprintf(os.Stdout, "[%s:%3d%s]\t", prog.Part.FileName, pct, "%")
				}
			}
		}
		fmt.Fprintf(os.Stdout, "\r")
	}

	if downloadSize >= length {
		cerr := combineFiles(fileName, rangeList, listSize)
		if cerr != nil {
			mainLog.Error("combineFiles error:", cerr)
			return cerr, ""
		}
	}

	succ := fmt.Sprintf("download finish. total length:%d", downloadSize)
	mainLog.Debug(succ)
	fmt.Println("\n", succ)

	return nil, ""
}

func checkDownloadUrl() {
	if *downloadUrl == "" {
		//		mainLog.Error("please specify Download URL through -u parameter")
		//		printHelp()
		//		os.Exit(1)
		*downloadUrl = "http://106.12.24.114/files/nginx.tar.gz"
	}
}

func main() {

	flag.Parse()

	utils.SetDebug(!*logFile)
	utils.SetLogDebug(*debug)

	checkDownloadUrl()

	// getFile("http://down1.tupwk.com.cn/qhwkdownpage/978-7-302-43674-4.zip")
	getFile(*downloadUrl, *threadCount)
	//	getRange("http://", 1022881, 1, "testfile")
}
