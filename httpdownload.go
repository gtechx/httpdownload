package main

import (
	"flag"
	"fmt"
	. "github.com/gtechx/base/common"
	"golang.org/x/net/proxy"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"
)

var rawurl string = ""
var outputdir string = "download/"
var proxyaddr string = "127.0.0.1:1080"
var proxyauth string = "user:password"
var threadnum int = 8
var bsock5 bool = true

func main() {
	//http://cfhcable.dl.sourceforge.net/project/boost/boost-binaries/1.64.0/boost_1_64_0-msvc-14.0-64.exe
	purl := flag.String("url", "", "-url=")
	pdir := flag.String("outdir", "./download/", "-outdir=")
	pproxy := flag.String("proxy", "", "-proxy=")
	pthreadnum := flag.Int("threadnum", 8, "-threadnum=")
	pproxyauth := flag.String("proxyauth", "", "-proxyauth=")
	pbsock5 := flag.Bool("sock5", true, "-socke5=")

	flag.Parse()

	rawurl = *purl //"http://discuzt.cr180.com/"
	outputdir = *pdir
	proxyaddr = *pproxy
	threadnum = *pthreadnum
	proxyauth = *pproxyauth
	bsock5 = *pbsock5

	if rawurl == "" {
		fmt.Println("-usage:")
		fmt.Println("-url=url")
		fmt.Println("-outdir=[dir] default is download/")
		fmt.Println("-proxy=[proxy]")
		fmt.Println("-threadnum=[num] default is 8")
		fmt.Println("-proxyauth=[user:password]")
		fmt.Println("-sock5=[true/false] default is true")
		return
	}

	testhttp()
}

func testhttp() {
	// setup a http client
	httpTransport := &http.Transport{}
	httpClient := &http.Client{Transport: httpTransport}
	var err error
	// set our socks5 as the dialer
	if proxyaddr != "" {
		var auth *proxy.Auth = nil
		if proxyauth != "" {
			autharr := strings.Split(proxyauth, ":")
			if len(autharr) > 1 {
				auth = &proxy.Auth{User: autharr[0], Password: autharr[1]}
			}
		}
		var dialer proxy.Dialer
		if bsock5 {
			dialer, err = proxy.SOCKS5("tcp", proxyaddr, auth, proxy.Direct)
		} else {
			rurl, err := url.Parse(proxyaddr)
			if err == nil {
				dialer, err = proxy.FromURL(rurl, proxy.Direct)
			}
		}
		if err != nil {
			fmt.Fprintln(os.Stderr, "can't connect to the proxy:", err)
			//os.Exit(1)
		} else {
			httpTransport.Dial = dialer.Dial
		}
	}

	//resp, err := httpClient.Get(rawurl)

	// client := &http.Client{
	// 	CheckRedirect: redirectPolicyFunc,
	// }

	//resp, err := httpClient.Get("http://example.com")
	// ...
	//downloadurl := "https://github.com/tam7t/xmpp/archive/master.zip" // "http://cfhcable.dl.sourceforge.net/project/boost/boost-binaries/1.64.0/boost_1_64_0-msvc-14.0-64.exe" //"https://superb-dca2.dl.sourceforge.net/project/tdm-gcc/TDM-GCC%20Installer/tdm64-gcc-5.1.0-2.exe"
	req, err := http.NewRequest("HEAD", rawurl, nil)
	// ...
	//req.Header.Add("Range", `bytes=0-499,601-999`)
	resp, err := httpClient.Do(req)
	fmt.Println(resp)
	fmt.Println(resp.Header)
	fmt.Println(resp.Status)
	fmt.Println("contentlength:", resp.ContentLength)
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	if resp.ContentLength <= 0 {
		fmt.Println("content length is not right:", resp.ContentLength)
		return
	}

	//threadnum := 8
	perbytes := int(resp.ContentLength) / threadnum
	remainbytes := int(resp.ContentLength) % threadnum

	startbyte := 0
	endbyte := perbytes - 1
	var endchan chan int
	endchan = make(chan int, threadnum)
	count := threadnum
	starttime := time.Now().Unix()

	purl, _ := url.Parse(rawurl)
	fpath := filepath.Base(purl.Path)                                 //"./master.zip"
	f, err1 := os.OpenFile(fpath, os.O_RDWR|os.O_CREATE, os.ModePerm) //可读写，追加的方式打开（或创建文件）
	if err1 != nil {
		panic(err1)
		return
	}
	f.Truncate(resp.ContentLength)
	defer f.Close()

	for i := 0; i < threadnum; i++ {
		if i == threadnum-1 {
			endbyte += remainbytes
		}
		go func(startbyte, endbyte, index int) {
			lstart := startbyte
			lend := endbyte

			// setup a http client
			httpTransport := &http.Transport{}
			httpClient := &http.Client{Transport: httpTransport}
			// set our socks5 as the dialer
			if proxyaddr != "" {
				var auth *proxy.Auth = nil
				if proxyauth != "" {
					autharr := strings.Split(proxyauth, ":")
					if len(autharr) > 1 {
						auth = &proxy.Auth{User: autharr[0], Password: autharr[1]}
					}
				}
				var dialer proxy.Dialer
				if bsock5 {
					dialer, err = proxy.SOCKS5("tcp", proxyaddr, auth, proxy.Direct)
				} else {
					rurl, err := url.Parse(proxyaddr)
					if err == nil {
						dialer, err = proxy.FromURL(rurl, proxy.Direct)
					}
				}
				if err != nil {
					fmt.Fprintln(os.Stderr, "can't connect to the proxy:", err)
					//os.Exit(1)
				} else {
					httpTransport.Dial = dialer.Dial
				}
			}

			req, err := http.NewRequest("GET", rawurl, nil)
			req.Header.Add("Range", `bytes=`+String(lstart)+`-`+String(lend))
			fmt.Println(index, "Range", `bytes=`+String(lstart)+`-`+String(lend)+" start")
			resp, err := httpClient.Do(req)
			if err != nil {
				fmt.Println("Range", `bytes=`+String(lstart)+`-`+String(lend)+" req error ", err.Error())
			} else {
				data, err := ioutil.ReadAll(resp.Body)
				if err != nil {
					fmt.Println("Range", `bytes=`+String(lstart)+`-`+String(lend)+" read error ", err.Error())
				} else {
					//fmt.Println("Range", `bytes=`+String(lstart)+`-`+String(lend)+" end:contentlength ", resp.ContentLength, " status:", resp.Status)
					f.Seek(int64(startbyte), 0)
					f.Write(data)
				}
				endchan <- 1
			}
		}(startbyte, endbyte, i)
		startbyte += perbytes
		endbyte += perbytes
	}

	for _ = range endchan {
		count -= 1
		if count == 0 {
			break
		}
	}
	usetime := time.Now().Unix() - starttime
	speed := float64(resp.ContentLength) / float64(usetime) / 1024.0

	fmt.Printf("download end, use %d ms, speed:%.2f KB/s", usetime, speed)

	// if speed < 1.0 {
	// 	fmt.Printf("download end, use %d ms, speed:%.2f KB/s", usetime, speed)
	// 	fmt.Println("download end, use", usetime, "seconds, speed:", String(Int64))+"KB/s")
	// }else{
	// 	fmt.Printf("download end, use %d ms, speed:%.2f KB/s", usetime, speed)
	// }
	// fmt.Println("download end, use", usetime, "seconds, speed:", String(Int64))+"KB/s")
	// a := 0
	// fmt.Scanln(&a)
}
