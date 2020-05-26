package main

import (
	"crypto/tls"
	"fmt"
	"gopkg.in/yaml.v2"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"time"
)

var confyaml conf
var baseurl string
type conf struct {
	Listenport string `yaml:"listenport"` //yaml：yaml格式 enabled：属性的为enabled
	Baseurl    []string `yaml:"baseurl"`
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

func healthcheck()  {
	for{
		for i := 0 ;i <= len(confyaml.Baseurl);i++ {
			tr:=&http.Transport{
				TLSClientConfig:&tls.Config{InsecureSkipVerify:true},
			}
			client:=&http.Client{Transport:tr}
			resp,err := client.Get(confyaml.Baseurl[i])
			if err == nil &&(resp.StatusCode == 404 || resp.StatusCode == 200){
				baseurl = confyaml.Baseurl[i]
				break
			}
		}
		time.Sleep(time.Duration(5)*time.Second)
	}
}

func handleRequestAndRedirect(res http.ResponseWriter, req *http.Request) {
	// We will get to this...
	//fmt.Printf(req.Method)
	url := baseurl+req.URL.RequestURI()
		tr:=&http.Transport{
			TLSClientConfig:&tls.Config{InsecureSkipVerify:true},
		}
		client:=&http.Client{Transport:tr}
	reqest, err := http.NewRequest(req.Method, url, req.Body)
	reqest.Header = req.Header
		if err!=nil{
			fmt.Println(err)
			io.WriteString(res,err.Error())
			return
		}
	response, _ := client.Do(reqest)
		defer reqest.Body.Close()
	body, err := ioutil.ReadAll(response.Body)
	if err!=nil{
		fmt.Println(err)
		io.WriteString(res,err.Error())
		return
	}
	//copy header
	for k, v := range response.Header {
		res.Header()[k] = v
	}
	//copy body
	io.WriteString(res,string(body))

}

func main() {
	go healthcheck()
	http.HandleFunc("/", handleRequestAndRedirect)
	http.ListenAndServe(":"+confyaml.Listenport, nil)
}
