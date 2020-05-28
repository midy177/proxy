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
	"math/rand"
	"net/http"
	"time"
)

 var confyaml *conf

type conf struct {
	Listenport string   `yaml:"listenport"` //yaml：yaml格式 enabled：属性的为enabled
	Heathcheck heathcheckinfo   `yaml:"heathcheck"`
	Baseurl    []string `yaml:"baseurl"`
	Verifyuri    []string `yaml:"verifyuri"`
}
type heathcheckinfo struct {
	Timeout int `yaml:"timeout"`
}
func init() {
	yamlFile, err := ioutil.ReadFile("conf.yaml")
	//log.Println("yamlFile:", yamlFile)
	if err != nil {
		log.Printf("yamlFile.Get err #%v ", err)
	}
	err = yaml.Unmarshal(yamlFile,&confyaml)
	if err != nil {
		log.Fatalf("Unmarshal: %v", err)
	}
}

func handleRequestAndRedirect(res http.ResponseWriter, req *http.Request) {
	// We will get to this...
    requrl := req.URL.RequestURI()
	if requrl == "/"{
		res.WriteHeader(200)
		res.Write([]byte("I'm health!"))
		return
	}
    reqmethod := req.Method
	dd,_ := ioutil.ReadAll(req.Body)
	needcheck := CheckResponse(requrl)
	for urlkey,urlvalue := range Random(){
		var url string
		url = urlvalue + requrl
		tr := &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		}
		client := &http.Client{Transport: tr,Timeout:time.Duration(confyaml.Heathcheck.Timeout)*time.Second}
		reqest, err := http.NewRequest(reqmethod, url, bytes.NewReader(dd))
		//copy req.header to proxy reqest
		for k, v := range req.Header {
			reqest.Header.Set(k,v[0])
		}
		if err != nil {
			log.Println(err)
			log.Printf("请求出错："+url+"\n")
			io.WriteString(res, err.Error())
		}else{
			response, err := client.Do(reqest)
			//copy response.header to res
			if err == nil {
				if needcheck {
					if response.StatusCode == 200 || urlkey == len(confyaml.Baseurl)-1 {
						for k, v := range response.Header {
							res.Header()[k] = v
						}
						//copy body
						res.WriteHeader(response.StatusCode)
						bufio.NewReader(response.Body).WriteTo(res)
						reqest.Body.Close()
						response.Body.Close()
						break
					}
					reqest.Body.Close()
					response.Body.Close()
					log.Printf("请求失败,url:"+urlvalue+",请求方法:"+req.Method+"\n")
				}else{
					for k, v := range response.Header {
						res.Header()[k] = v
					}
					//copy body
					res.WriteHeader(response.StatusCode)
					bufio.NewReader(response.Body).WriteTo(res)
					reqest.Body.Close()
					response.Body.Close()
					break
				}
			}else if urlkey == len(confyaml.Baseurl)-1 {
				res.WriteHeader(502)
				res.Write([]byte("无可用节点\n"))
				log.Printf("请求失败,url:"+urlvalue+",请求方法:"+req.Method+"\n")
			}
		}
	}
}

//需要检查状态码是否为200的uri
func CheckResponse(requri string)bool  {
	for _,v := range confyaml.Verifyuri {
     if v == requri{
     	return true
	 }
	}
	return false
}

//洗牌算法
func Random()[]string {
	tmparray := confyaml.Baseurl
	if len(confyaml.Baseurl) <= 0 {
		return []string{"http://1.1.1.1"}
	}
	for i := len(tmparray) - 1; i > 0; i-- {
		num := rand.Intn(i + 1)
		tmparray[i], tmparray[num] = tmparray[num], tmparray[i]
	}

	return tmparray
}

func main() {
	fmt.Printf("开始监听端口："+confyaml.Listenport+"\n")
	http.HandleFunc("/", handleRequestAndRedirect)
	http.ListenAndServe(":"+confyaml.Listenport, nil)
}
