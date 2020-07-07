package main

import (
	"bytes"
	"compress/gzip"
	"crypto/tls"
	"fmt"
	cbrotli "github.com/andybalholm/brotli"
	//"github.com/jasonlvhit/gocron"
	"github.com/robfig/cron/v3"
	"gopkg.in/yaml.v2"
	"io"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"strings"
	"sync"
	"time"
)

 var confyaml *conf

type conf struct {
	Listenport string   `yaml:"listenport"` //yaml：yaml格式 enabled：属性的为enabled
	Mode string `yaml:"mode"`
	Heathcheck heathcheckinfo   `yaml:"heathcheck"`
	ProxyUrl    []string `yaml:"baseurl"`
	Verifyuri    uriverify `yaml:"verifyuri"`
	MatchContentType ContentType`yaml:"MatchContentType"`
}
type heathcheckinfo struct {
	Enable bool `yaml:"enable"`
	Interval int `yaml:"interval"`
	CheckUrl string `yaml:"checkurl"`
	Timeout int `yaml:"timeout"`
}
type uriverify struct {
	Enable bool `yaml:"enable"`
	Uri []string `yaml:"uri"`
}
type ContentType struct {
	Enable bool `yaml:"enable"`
	ContentType []string `yaml:"ContentType"`
	replace map[string]string `yaml:"replace"`
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
	var proxybaseurl []string
	if strings.Contains(confyaml.Mode,"dynamic"){
		proxybaseurl = []string{"http://"+req.Host,"https://"+req.Host}
	}else if strings.Contains(confyaml.Mode,"static"){
		proxybaseurl = Random()
	}else {
		proxybaseurl = Random()
	}
	for urlkey,urlvalue := range proxybaseurl {
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
						if (response.StatusCode >= 200 && response.StatusCode <= 300) || urlkey == len(proxybaseurl)-1 {
							for k, v := range response.Header {
								res.Header()[k] = v
							}
							//copy body
							if CheckContentType(response.Header.Get("content-type")) {
								respbody = ToReplaceUrl(respbody,response.Header.Get("Content-Encoding"))
							}
							res.WriteHeader(response.StatusCode)
							res.Write(respbody)
							reqest.Body.Close()
							response.Body.Close()
							break
						}
						reqest.Body.Close()
						response.Body.Close()
						log.Printf("请求失败,url:"+url+",请求方法:"+req.Method+"\n")
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
			}else if urlkey == len(proxybaseurl)-1 {
				res.WriteHeader(502)
				res.Write([]byte("无可用节点\n"))
				log.Printf("请求失败,url:"+url+",请求方法:"+req.Method+"\n")
				log.Printf("当前所有节点请求失败！")
			}
		}
	}
}

//需要检查状态码是否为200的uri
func CheckResponse(requri string)bool  {

	for _,v := range confyaml.Verifyuri.Uri {
     if v == requri{
     	return true
	 }
	}
	return false
}

//洗牌算法
func Random()[]string {
	rwlock.RLock()
	tmparray := healthUrl
	rwlock.RUnlock()
	if len(tmparray) <= 0 {
		log.Printf("无可用节点")
		return []string{"http://1.1.1.1"}
	}
	for i := len(tmparray) - 1; i > 0; i-- {
		num := rand.Intn(i + 1)
		tmparray[i], tmparray[num] = tmparray[num], tmparray[i]
	}

	return tmparray
}
func ToReplaceUrl(before []byte,isgzip string)[]byte  {
	if len(confyaml.MatchContentType.replace) == 0{
		return before
	}
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
		for k,v := range confyaml.MatchContentType.replace {
			undatas = bytes.Replace(undatas,[]byte(k),[]byte(v),-1)
		}
		var buf bytes.Buffer
		g := gzip.NewWriter(&buf)
		g.Write(undatas)
        g.Close()
		return buf.Bytes()
	}else if strings.Contains(isgzip,"br"){
		r := cbrotli.NewReader(reader)
		//defer r.Close()
		if r == nil {
			log.Printf("fail to brotli stream byte")
			return before
		}
		undatas, err := ioutil.ReadAll(r)
		if err != nil {
			log.Printf(err.Error())
			return before
		}
		for k,v := range confyaml.MatchContentType.replace {
			undatas = bytes.Replace(undatas,[]byte(k),[]byte(v),-1)
		}
		var buf bytes.Buffer
		b := cbrotli.NewWriter(&buf)
		b.Write(undatas)
		b.Close()
        return buf.Bytes()
	}else{
		var rpdatas []byte
		for k,v := range confyaml.MatchContentType.replace {
			rpdatas = bytes.Replace(before,[]byte(k),[]byte(v),-1)
		}
		//log.Printf(string(before))
		return rpdatas
	}
}
func CheckContentType(ContentType string)bool{
	if !confyaml.MatchContentType.Enable{
		return false
	}
 for _,v := range confyaml.MatchContentType.ContentType {
	 if strings.Contains(ContentType,v){
		 return true
	 }
 }
    return false
}
var healthUrl []string
var rwlock sync.RWMutex
func CheckHealth(){
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{Transport: tr,Timeout:time.Duration(confyaml.Heathcheck.Timeout)*time.Second}
	var healist []string
	for _,v := range confyaml.ProxyUrl{
        respone,err := client.Head(v+confyaml.Heathcheck.CheckUrl)
        if err != nil{
        	continue
		}else if respone.StatusCode == 200{
			healist = append(healist, v)
		}
	}
	rwlock.Lock()
	healthUrl = healist
	rwlock.Unlock()
}
func main() {
	if confyaml.Heathcheck.Enable {
		CheckHealth()
		nyc, _ := time.LoadLocation("Asia/Shanghai")
		c := cron.New(cron.WithSeconds(), cron.WithLocation(nyc))
		durtime := "@every "+fmt.Sprintf("%d", confyaml.Heathcheck.Interval)+"s"
		idc,_ := c.AddFunc(durtime,CheckHealth)
		//c.Run()
		fmt.Sprintf("%d", idc)
		c.Start()
	}else {
		healthUrl = confyaml.ProxyUrl
	}
	fmt.Printf("开始监听端口："+confyaml.Listenport+"\n")
	http.HandleFunc("/", handleRequestAndRedirect)
	http.ListenAndServe(":"+confyaml.Listenport, nil)
}
