//处理数据-抓取数据-存标签库
package main

import "runtime"

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())
	//	go func() {
	//		ch := make(chan os.Signal)
	//		signal.Notify(ch, syscall.SIGHUP, syscall.SIGTERM)
	//		f, _ := os.Create("cpu_1111")
	//		for {
	//			switch <-ch {
	//			case syscall.SIGHUP:
	//				pprof.StartCPUProfile(f)
	//				fmt.Println("begin")

	//			case syscall.SIGTERM:
	//				fmt.Println("end")
	//				pprof.StopCPUProfile()
	//				f.Close()
	//			}
	//		}
	//	}()
	go Bootstrap("_ck")
	go Bootstrap("_ad")
	select {}

}
