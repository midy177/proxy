package main

import (
	"bytes"
	"compress/gzip"
	"crypto/tls"
	"fmt"
	"gopkg.in/yaml.v2"
	"io"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"strings"
	"time"
)

 var confyaml *conf

type conf struct {
	Listenport string   `yaml:"listenport"` //yaml：yaml格式 enabled：属性的为enabled
	Heathcheck heathcheckinfo   `yaml:"heathcheck"`
	Baseurl    []string `yaml:"baseurl"`
	Verifyuri    []string `yaml:"verifyuri"`
	MatchContentType []string `yaml:"MatchContentType"`
	UrlReplace     map[string]string `yaml:"urlreplace"`
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
    //healthcheck
	if requrl == "/healthcheck"{
		res.WriteHeader(200)
		res.Write([]byte("I'm health!"))
		return
	}
    reqmethod := req.Method
    var dd []byte
    if req.Body != nil{
    	defer req.Body.Close()
		dd,_ = ioutil.ReadAll(req.Body)
	}else {dd = nil}
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
				respbody,err := ioutil.ReadAll(response.Body)
				if err == nil {
					if needcheck {
						if (response.StatusCode >= 200 && response.StatusCode <= 300) || urlkey == len(confyaml.Baseurl)-1 {
							for k, v := range response.Header {
								res.Header()[k] = v
							}
							//copy body
							if CheckContentType(response.Header.Get("content-type")) {
								respbody = ToReplaceUrl(respbody,response.Header.Get("Content-Encoding"))
							}
							res.WriteHeader(response.StatusCode)
							res.Write(respbody)
							//bufio.NewReader(response.Body).WriteTo(res)
							reqest.Body.Close()
							response.Body.Close()
							break
						}
						reqest.Body.Close()
						response.Body.Close()
						log.Printf("请求失败,url:"+urlvalue+",请求方法:"+req.Method+"\n")
					}else {
						for k, v := range response.Header {
							res.Header()[k] = v
						}
						//copy body
						if CheckContentType(response.Header.Get("Content-Type")) {
							respbody = ToReplaceUrl(respbody,response.Header.Get("Content-Encoding"))
						}
						res.WriteHeader(response.StatusCode)
						res.Write(respbody)
						//bufio.NewReader(response.Body).WriteTo(res)
						reqest.Body.Close()
						response.Body.Close()
						break
					}
				}else {
					log.Printf(err.Error())
				}
			}else if urlkey == len(confyaml.Baseurl)-1 {
				res.WriteHeader(502)
				res.Write([]byte("无可用节点\n"))
				log.Printf("请求失败,url:"+urlvalue+",请求方法:"+req.Method+"\n")
				log.Printf("当前所有节点请求失败！")
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
func ToReplaceUrl(before []byte,isgzip string)[]byte  {
	//log.Printf(string(before))
	//var reader bytes.Reader
	reader := bytes.NewReader(before)
	if strings.Contains(isgzip,"gzip") {
		r,err := gzip.NewReader(reader)
		defer r.Close()
		if err != nil {
			log.Printf(err.Error())
			return before
		}
		undatas, err := ioutil.ReadAll(r)
		if err != nil {
			log.Printf(err.Error())
			return before
		}
		for k,v := range confyaml.UrlReplace {
			undatas = bytes.Replace(undatas,[]byte(k),[]byte(v),-1)
		}
		var buf bytes.Buffer
		g := gzip.NewWriter(&buf)
		g.Write(undatas)
        g.Close()
		if err != nil {
			log.Printf(err.Error())
			return before
		}
		return buf.Bytes()
	}else{
		var rpdatas []byte
		for k,v := range confyaml.UrlReplace {
			rpdatas = bytes.Replace(before,[]byte(k),[]byte(v),-1)
		}
		//log.Printf(string(before))
		return rpdatas
	}
}
func CheckContentType(ContentType string)bool{
 for _,v := range confyaml.MatchContentType {
	 if strings.Contains(ContentType,v){
		 return true
	 }
 }
    return false
}

func main() {
	fmt.Printf("开始监听端口："+confyaml.Listenport+"\n")
	http.HandleFunc("/", handleRequestAndRedirect)
	http.ListenAndServe(":"+confyaml.Listenport, nil)
}
