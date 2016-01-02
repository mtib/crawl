package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"regexp"
	"strings"
)

const (
	raw     = "http[s]{0,1}:\\/\\/([a-zA-Z0-9]{0,}\\.)?%v\\/[a-zA-Z0-9\\/_]{0,}"
	relraw  = " href=[\"']{1}[\\.\\/a-zA-Z0-9_]{1,}[\"']{1}"
	teststr = "http://tibyte.net/index.html https://google.de/tschuee http://google.de/hallo"
)

func getDomain(dom string, rabs *regexp.Regexp, rrel *regexp.Regexp) ([]string, error) {
	var protocol string
	ans, err := http.Get(fmt.Sprintf("http://%v", dom))
	protocol = "http://"
	if err != nil {
		ans, err = http.Get(fmt.Sprintf("https://%v", dom))
		protocol = "https://"
		if err != nil {
			fmt.Printf("Domain [%v] doesn't exist\n", dom)
			//panic(err)
			return nil, nil //TODO: return error
		}
	}
	defer ans.Body.Close()
	text, _ := ioutil.ReadAll(ans.Body)
	longabs := rabs.FindAllString(string(text), -1)
	longrel := rrel.FindAllString(string(text), -1)
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
		res[count] = v
		count++
	}
	for _, v := range longrel {
		res[count] = v
		count++
	}
	return res, nil
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
	for _, l := range links {
		fmt.Println(l)
	}
}
