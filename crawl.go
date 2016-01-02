package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"regexp"
	"strings"
	"time"
)

const (
	raw     = "http[s]{0,1}:\\/\\/([a-zA-Z0-9]{0,}\\.)?%v\\/[a-zA-Z0-9\\/_]{0,}"
	relraw  = " href=[\"']{1}[\\.\\/a-zA-Z0-9_]{1,}[\"']{1}"
	teststr = "http://tibyte.net/index.html https://google.de/tschuee http://google.de/hallo"
)

type crawlerError string
type webspace map[string]int
type addc chan string

func (w webspace) String() string {
	ans := ""
	for k, v := range w {
		ans = fmt.Sprintf("%v%v", ans, fmt.Sprintf("(%vx) %s\n", v, k))
	}
	return ans
}

func (e crawlerError) Error() string {
	return fmt.Sprintf("Crawl error: %s", string(e))
}

func cutdomain(s string) string {
	for strings.HasSuffix(s, "/") {
		s = s[:len(s)-1]
	}
	return s
}

func getDomain(dom string, rabs *regexp.Regexp, rrel *regexp.Regexp) ([]string, error) {
	if strings.HasPrefix(dom, "https://") {
		dom = dom[8:]
	}
	if strings.HasPrefix(dom, "http://") {
		dom = dom[7:]
	}
	for strings.HasSuffix(dom, "/") {
		dom = dom[:len(dom)-1]
	}
	var protocol string
	ans, err := http.Get(fmt.Sprintf("http://%v", dom))
	protocol = "http://"
	if err != nil {
		ans, err = http.Get(fmt.Sprintf("https://%v", dom))
		protocol = "https://"
		if err != nil {
			fmt.Printf("Domain [%v] doesn't exist\n", dom)
			fmt.Println(err)
			return nil, crawlerError("http and https not supported")
		}
	}
	defer ans.Body.Close()
	text, _ := ioutil.ReadAll(ans.Body)
	longabs := rabs.FindAllString(string(text), -1)
	longrel := rrel.FindAllString(string(text), -1)
	if i := strings.Index(dom, "/"); i != -1 {
		dom = dom[:i]
	}
	for k, v := range longrel {
		longrel[k] = v[7 : len(v)-1]
		if longrel[k][0] == '/' {
			longrel[k] = fmt.Sprintf("%v%v%v", protocol, dom, longrel[k])
		} else {
			longrel[k] = fmt.Sprintf("%v%v/%v", protocol, dom, longrel[k])
		}
	}
	// TODO: make this a map!
	res := make([]string, len(longabs)+len(longrel))
	count := 0
	for _, v := range longabs {
		res[count] = cutdomain(v)
		count++
	}
	for _, v := range longrel {
		res[count] = cutdomain(v)
		count++
	}
	return res, nil
}

func fill(addr string, c *addc, r1, r2 *regexp.Regexp) {
	links, err := getDomain(addr, r1, r2)
	if err != nil {
		fmt.Printf("Could not read %v\n", addr)
		*c <- "none"
		return
	}
	for _, l := range links {
		*c <- l
	}
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: crawl <domain>\ne.g: crawl example.org")
		return
	}
	url := os.Args[1]
	savurl := strings.Replace(strings.Replace(url, ".", "\\.", -1), "/", "\\/", -1)
	reg := fmt.Sprintf(raw, savurl)
	r, err := regexp.Compile(reg)
	if err != nil {
		fmt.Println("Could not compile regexp")
		//panic(err)
		return
	}
	rrel, _ := regexp.Compile(relraw)
	//fmt.Printf("Crawling %v with:\n%v\n%v\n", url, reg, relraw)
	links, _ := getDomain(url, r, rrel)
	maxsize := 100
	collection := make(webspace, maxsize)
	for _, v := range links {
		n, ok := collection[v]
		if ok {
			collection[v] = n + 1
		} else {
			collection[v] = 1
		}
	}
	lres := make(addc, 500)
	start := 0
	end := len(collection)
	count := 1
	for {
		fmt.Printf("%v. iteration (%v links)\n", count+1, end)
		count++
		new := end - start
		if new == 0 {
			break
		}
		for k := range collection {
			if start < end {
				fmt.Println(start, end)
				go fill(k, &lres, r, rrel)
				start++
			} else {
				continue
			}
		}
		for index := 0; index < new; index++ {
			ns := <-lres
			fmt.Println(ns)
			if ns == "none" {
				continue
			}
			n, ok := collection[ns]
			if ok {
				collection[ns] = n + 1
			} else {
				collection[ns] = 1
				fmt.Println(ns)
			}
		}
		end = len(collection)
	}
	time.Sleep(1 * time.Second)
	//close(lres)
	fmt.Println(collection)
}
