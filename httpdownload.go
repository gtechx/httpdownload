package main

import (
	"flag"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	. "github.com/gtechx/base/common"
	"golang.org/x/net/proxy"
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

	downloaders := make([]*Downloader, threadnum)
	for i := 0; i < threadnum; i++ {
		if i == threadnum-1 {
			endbyte += remainbytes
		}

		d := NewDownloader(outputdir, rawurl, startbyte, endbyte, i, proxyaddr != "", proxyaddr, proxyauth, bsock5)
		downloaders[i] = d
		go d.Start()

		startbyte += perbytes
		endbyte += perbytes
	}

	for _, d := range downloaders {
		<-d.Done
	}

	for i, d := range downloaders {
		fname := outputdir + fpath + String(d.Index)
		df, err := os.OpenFile(fname, os.O_RDWR, os.ModePerm) //可读写，追加的方式打开（或创建文件）
		if err != nil {
			fmt.Println("open file index ", i, " error:", err.Error())
			break
		}
		f.Seek(int64(d.StartByte), 0)

		buff := make([]byte, 1024)

		for {
			c, err := df.Read(buff)
			if err != nil || c == 0 {
				break
			}

			f.Write(buff[:c])
		}

		df.Close()
		os.Remove(fname)
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

type Downloader struct {
	DownloadPath string
	Url          string
	StartByte    int
	EndByte      int
	Index        int
	UseProxy     bool
	Proxy        string
	ProxyAuth    string
	UseSock5     bool

	Done chan bool
}

func NewDownloader(downloadpath, url string, startbyte, endbyte int, index int, useproxy bool, proxy, proxyauth string, usesock5 bool) *Downloader {
	d := &Downloader{}
	d.DownloadPath = downloadpath
	d.Url = url
	d.StartByte = startbyte
	d.EndByte = endbyte
	d.Index = index
	d.UseProxy = useproxy
	d.Proxy = proxy
	d.ProxyAuth = proxyauth
	d.UseSock5 = usesock5

	d.Done = make(chan bool, 1)

	return d
}

func (d *Downloader) Start() {
	var err error

	purl, _ := url.Parse(d.Url)
	fname := filepath.Base(purl.Path)

	httpTransport := &http.Transport{}
	httpClient := &http.Client{Transport: httpTransport}

	if d.UseProxy {
		// set our socks5 as the dialer
		if d.Proxy != "" {
			var auth *proxy.Auth = nil
			if d.ProxyAuth != "" {
				autharr := strings.Split(d.ProxyAuth, ":")
				if len(autharr) > 1 {
					auth = &proxy.Auth{User: autharr[0], Password: autharr[1]}
				}
			}
			var dialer proxy.Dialer
			if d.UseSock5 {
				dialer, err = proxy.SOCKS5("tcp", d.Proxy, auth, proxy.Direct)
			} else {
				rurl, err := url.Parse(d.Proxy)
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
	}

	req, err := http.NewRequest("GET", d.Url, nil)
	req.Header.Add("Range", `bytes=`+String(d.StartByte)+`-`+String(d.EndByte))
	fmt.Println(d.Index, "Range", `bytes=`+String(d.StartByte)+`-`+String(d.EndByte)+" start")
	resp, err := httpClient.Do(req)

	if err != nil {
		fmt.Println("Range", `bytes=`+String(d.StartByte)+`-`+String(d.EndByte)+" req error ", err.Error())
	} else {
		f, err := os.OpenFile(d.DownloadPath+fname+String(d.Index), os.O_RDWR|os.O_CREATE, os.ModePerm) //可读写，追加的方式打开（或创建文件）
		if err != nil {
			fmt.Println("Range", `bytes=`+String(d.StartByte)+`-`+String(d.EndByte)+" create file error ", err.Error())
			d.Done <- true
			return
		}
		f.Truncate(resp.ContentLength)
		defer f.Close()

		buff := make([]byte, 1024000)

		sum := 0

		for {
			count, err := resp.Body.Read(buff)

			if err != nil {
				fmt.Println("Range", `bytes=`+String(d.StartByte)+`-`+String(d.EndByte)+" read error ", err.Error())
				break
			} else if count == 0 {
				fmt.Println("Range", `bytes=`+String(d.StartByte)+`-`+String(d.EndByte)+" read end ")
			} else {
				sum += count
				//per := fmt.Sprintf("%.2f", float64(sum)/float64(d.EndByte-d.StartByte))
				//fmt.Println("Range index ", d.Index, ` bytes=`+String(d.StartByte)+`-`+String(d.EndByte)+" download percent: ", per, "% "+String(sum)+"/"+String(d.EndByte-d.StartByte), " count ", count)
				f.Write(buff[:count])
			}
		}
	}
	d.Done <- true
}
