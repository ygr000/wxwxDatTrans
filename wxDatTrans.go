package main

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"sync"
	"time"
)

type Dic struct {
	name       string //名称
	firstIndex uint8  //第一个字节
	lastIndex  uint8  //第二个字节
}
type FileInfo struct {
	filePath   string //文件路径
	fileName   string //文件名称 不包括后缀 .dat
	fileSuffix string // 文件后缀
}

var dicList = []Dic{Dic{".jpg", 0xff, 0xd8},
	Dic{".png", 0x89, 0x50},
	Dic{".gif", 0x47, 0x49},
	Dic{"error", 0x00, 0x00}}

func main() {
	start := time.Now()
	getAllDatFileList(".\\")
	var wg sync.WaitGroup

	// 定义循环值
	var begin = 0
	var ended = 9999
	var num = len(fileInfoArr)

	for i := 0; i <= num/9999; i++ {
		// 判断结束不能大于总值
		if ended > num {
			ended = num
		}

		fmt.Printf(" 文件总数:%d 分%d批次执行 正在执行第%d批次 开始点:%d 结束点:%d  \n", num, num/9999+1, i+1, begin, ended)
		slice := fileInfoArr[begin:ended]
		// 循环赋值
		begin = ended
		ended = ended + 9999

		for _, v := range slice {
			wg.Add(1)
			go changeDat(v, &wg)
		}
		wg.Wait()

	}
	end := time.Now()
	ms := (end.Sub(start).Milliseconds())
	s := (end.Sub(start).Seconds())
	fmt.Printf(" 文件数:%d 总共耗时 %d ms (%f s)   \n", len(fileInfoArr), ms, s)
	fmt.Println("按任意键退出...")
	var input string
	fmt.Scanln(&input)
}

func changeDat(info FileInfo, wg *sync.WaitGroup) {
	defer wg.Done()
	data, err := ioutil.ReadFile(info.filePath + "\\" + info.fileName + info.fileSuffix)
	if err != nil {
		fmt.Println(err)
		return
	}
	addCode, dic, err := getAddCode(data)
	if err != nil {
		fmt.Println(err)
		return
	}
	writeXORAddCodeIntoNewFile(data, addCode, info, dic)
}

/**
 * arr dat文件字节切片
 *  返回 密码addcode 类型dic ,错误
 */
func getAddCode(arr []uint8) (addCode uint8, dic Dic, err error) {
	//遍历 diclist 看dat原本格式
	for _, dic := range dicList {
		addCode = arr[0] ^ dic.firstIndex
		if arr[1]^addCode == dic.lastIndex {
			return addCode, dic, nil
		}
	}
	return 0, dicList[3], errors.New("不是jpg,png,gif")
}

func writeXORAddCodeIntoNewFile(arr []uint8, addCode uint8, info FileInfo, dic Dic) {
	//生成目标路径
	var pos = strings.LastIndex(info.filePath, "\\")
	var willReplace = info.filePath[pos:]
	targetPath := strings.ReplaceAll(info.filePath, willReplace, "\\target"+willReplace+"\\")
	err := os.MkdirAll(targetPath, os.ModePerm)
	//打开文件
	f, err := os.OpenFile(targetPath+info.fileName+dic.name, os.O_RDWR|os.O_CREATE, 0777)
	defer f.Close()
	if err != nil {
		fmt.Println(err)
	}
	//对字节切片每个字节异或
	for i, v := range arr {
		arr[i] = v ^ addCode
	}
	//写入文件
	f.Write(arr)
}

var fileInfoArr = make([]FileInfo, 0)

func getAllDatFileList(parentPath string) {
	parentFileInfo, err := ioutil.ReadDir(parentPath)
	if err != nil {
		fmt.Println(err)
	}
	for _, fi := range parentFileInfo {
		if fi.IsDir() {
			getAllDatFileList(parentPath + "\\" + fi.Name())
		} else {
			if strings.HasSuffix(fi.Name(), ".dat") {
				fileInfoArr = append(fileInfoArr, FileInfo{parentPath, strings.TrimSuffix(fi.Name(), ".dat"), ".dat"})
			}
		}
	}
}
