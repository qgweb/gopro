package compress

import (
	"archive/tar"
	"archive/zip"
	"bytes"
	"compress/gzip"
	"errors"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

type Archive struct {
	bts []byte
	dst string
	src string
}

const ERR_DST_FILE_NOT_EXIST = "目标文件不存在"
const ERR_SRC_FILE_NOT_EXIST = "源文件不存在"
const ERR_DATA_EMPTY = "压缩数据为空"

func isExist(src string) bool {
	_, err := os.Stat(src)
	return err == nil || os.IsExist(err)
}

func NewArchive(src string, dst string) *Archive {
	return &Archive{bts: make([]byte, 0), dst: dst, src: src}
}

// tar 压缩
func (this *Archive) Tar() error {
	//判断是否存在
	if !isExist(this.src) {
		return errors.New(ERR_SRC_FILE_NOT_EXIST)
	}

	bs := bytes.NewBuffer(nil)
	tw := tar.NewWriter(bs)
	fistr, _ := filepath.Abs(this.src)
	PATH, _ := filepath.Split(fistr)

	filepath.Walk(fistr, func(path string, info os.FileInfo, err error) error {
		var new_path = ""
		if info.IsDir() {
			new_path = strings.TrimRight(strings.Replace(path, PATH, "", -1), string(os.PathSeparator))
		} else {
			new_path = strings.Replace(path, PATH, "", -1)
		}

		h, err := tar.FileInfoHeader(info, "")
		h.Name = filepath.Clean(new_path)

		if err := tw.WriteHeader(h); err != nil {
			return err
		}
		if !info.IsDir() {
			ff, _ := os.Open(path)
			defer ff.Close()
			io.Copy(tw, ff)
		}
		return nil
	})

	tw.Close()

	this.bts = bs.Bytes()
	return nil
}

// gz 压缩
func (this *Archive) GZ() error {
	//判断是否有数据
	if len(this.bts) == 0 {
		return errors.New(ERR_DATA_EMPTY)
	}

	bs := bytes.NewBuffer(nil)
	gz := gzip.NewWriter(bs)
	_, err := gz.Write(this.bts)
	if err != nil {
		return err
	}
	gz.Flush()
	gz.Close()

	this.bts = bs.Bytes()
	return nil
}

func (this *Archive) Zip() error {
	//判断是否存在
	if !isExist(this.src) {
		return errors.New(ERR_SRC_FILE_NOT_EXIST)
	}

	buf := bytes.NewBuffer(make([]byte, 0, 10*1024*1024)) // 创建一个读写缓冲
	myzip := zip.NewWriter(buf)                           // 用压缩器包装该缓冲
	// 用Walk方法来将所有目录下的文件写入zip
	err := filepath.Walk(this.src, func(path string, info os.FileInfo, err error) error {
		var file []byte
		if err != nil {
			return filepath.SkipDir
		}
		header, err := zip.FileInfoHeader(info) // 转换为zip格式的文件信息
		if err != nil {
			return filepath.SkipDir
		}
		header.Name, _ = filepath.Rel(filepath.Dir(this.src), path)

		if !info.IsDir() {
			// 确定采用的压缩算法（这个是内建注册的deflate）
			header.Method = 8
			file, err = ioutil.ReadFile(path) // 获取文件内容
			if err != nil {
				return filepath.SkipDir
			}
		} else {
			file = nil
		}
		// 上面的部分如果出错都返回filepath.SkipDir
		// 下面的部分如果出错都直接返回该错误
		// 目的是尽可能的压缩目录下的文件，同时保证zip文件格式正确
		w, err := myzip.CreateHeader(header) // 创建一条记录并写入文件信息
		if err != nil {
			return err
		}
		_, err = w.Write(file) // 非目录文件会写入数据，目录不会写入数据
		if err != nil {        // 因为目录的内容可能会修改
			return err // 最关键的是我不知道咋获得目录文件的内容
		}
		return nil
	})
	if err != nil {
		return err
	}
	myzip.Close() // 关闭压缩器，让压缩器缓冲中的数据写入buf

	this.bts = buf.Bytes()

	return nil
}

// 保存到目标文件中
func (this *Archive) Save() error {
	f, err := os.Create(this.dst)
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = f.Write(this.bts)
	if err != nil {
		return err
	}
	return nil
}
