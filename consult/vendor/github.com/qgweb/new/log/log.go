package main

import (
	"os"
	"github.com/ngaut/log"
	"bufio"
	"io"
	"os/exec"
	"bytes"
	"github.com/qgweb/new/lib/encrypt"
	"fmt"
	"strings"
)

func sscript(buf string) []byte {
	cmd := exec.Command("php", "a.php", encrypt.DefaultBase64.Encode(buf))
	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Start()
	if err != nil {
		log.Info(err)
		return nil
	}
	err = cmd.Wait()
	if err != nil {
		log.Info(err)
		return nil
	}
	return out.Bytes()
}

func main() {
	f, err := os.Open("/home/zb/桌面/20160723_dj.txt")
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	bi := bufio.NewReader(f)
	for {
		line, err := bi.ReadString('\n')
		if err == io.EOF {
			break
		}

		fmt.Println(strings.TrimSpace(line) + "\t" + string(sscript(line)))
	}

}
