package main

import (
	"bufio"
	"bytes"
	"crypto/tls"
	"fmt"
	"gopkg.in/yaml.v2"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"sync"
	"time"
)

var confyaml conf
var HealthUrl string
var rLock sync.Mutex

type conf struct {
	Listenport string   `yaml:"listenport"` //yaml：yaml格式 enabled：属性的为enabled
	Heathcheck heathcheckinfo   `yaml:"heathcheck"`
	Baseurl    []string `yaml:"baseurl"`
}
type heathcheckinfo struct {
	Timeout int64 `yaml:"timeout"`
	Interval int64 `yaml:"interval"`
}
func init() {
	yamlFile, err := ioutil.ReadFile("conf.yaml")
	//log.Println("yamlFile:", yamlFile)
	if err != nil {
		log.Printf("yamlFile.Get err #%v ", err)
	}
	err = yaml.Unmarshal(yamlFile, &confyaml)
	if err != nil {
		log.Fatalf("Unmarshal: %v", err)
	}
}

func healthcheck() {
	var heaInterval int64
	var heaTimeout int64
	if confyaml.Heathcheck.Interval == 0 {
		heaInterval = 5
	}else {
		heaInterval = confyaml.Heathcheck.Interval
	}
	if confyaml.Heathcheck.Timeout == 0{
		heaTimeout = 2
	}else {
		heaTimeout = confyaml.Heathcheck.Timeout
	}
	for {
		for i := 0; i < len(confyaml.Baseurl); i++ {
			tr := &http.Transport{
				TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
			}
			timeout := time.Duration(heaTimeout) * time.Second
			client := &http.Client{Transport: tr,Timeout:timeout}
			resp, err := client.Get(confyaml.Baseurl[i])
			if err == nil && (resp.StatusCode == 404 || resp.StatusCode == 200) {
				rLock.Lock()
				HealthUrl = confyaml.Baseurl[i]
				log.Printf("当前节点为："+HealthUrl+"\n")
				rLock.Unlock()
				time.Sleep(time.Duration(heaInterval) * time.Second)
			}
		}

	}
}

func handleRequestAndRedirect(res http.ResponseWriter, req *http.Request) {
	// We will get to this...
	var url string
	rLock.Lock()
	//判断url是否可用
	if HealthUrl == ""{
		url = "https://1.1.1.1"+ req.URL.RequestURI()
		log.Printf("当前无可用节点！！\n")
	}else{
		url = HealthUrl + req.URL.RequestURI()
	}
	rLock.Unlock()
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{Transport: tr}
	dd,_ := ioutil.ReadAll(req.Body)
	reqest, err := http.NewRequest(req.Method, url, bytes.NewReader(dd))
	//copy req.header to proxy reqest
	for k, v := range req.Header {
		reqest.Header.Set(k,v[0])
	}
	if err != nil {
		log.Println(err)
		log.Printf("当前请求出错路径："+url+"\n")
		io.WriteString(res, err.Error())
		return
	}
	response, _ := client.Do(reqest)
	defer reqest.Body.Close()
	//copy response.header to res
	for k, v := range response.Header {
		res.Header()[k] = v
	}
	//copy body
	res.WriteHeader(response.StatusCode)
	bufio.NewReader(response.Body).WriteTo(res)
}

func main() {
	go healthcheck()
	fmt.Printf("开始监听端口："+confyaml.Listenport+"\n")
	http.HandleFunc("/", handleRequestAndRedirect)
	http.ListenAndServe(":"+confyaml.Listenport, nil)
}
