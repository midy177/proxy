##proxy是什么?
后端服务器代理应用，同时支持设置需要校验uri请求状态码是否符合需求，当不符合时尝试请求所有后端
服务，直到代理请求返回的状态码为200时将代理请求结果转发给请求，或者遍历后端服务器组无理想状态
时将最后请求的后端服务器返回信息返回给用户
者

##配置文件，yaml格式, conf.yaml

```
listenport: 80
mode: static   #static 静态配置，指定后端host, dynamic  动态host，动态请求需要访问的URL default: static
heathcheck:   #健康检查配置
  enable: true #健康检查开关
  interval: 60  #健康检查间隔
  checkurl: /healthcheck  #健康检查路径
  timeout: 10 #http请求超时时间
verifyuri: #校验代理请求对应,状态码是否为200，单不是200时遍历所有健康节点
  enable: true  #开关
  uri:
    - /v1/chain/get_info
    - /v1/chain/get_table_rows
    - /v1/chain/push_transaction
baseurl:  #代理url
  #- https://morecoin.zendesk.more.top
  - https://eosnode.pizzadex.io
  - https://nodes.get-scatter.com
  - http://eospush.tokenpocket.pro
  - http://openapi.eos.ren
  - https://mainnet.eosio.sg
  - https://api.eossweden.org
  - https://eos.greymass.com
  - https://mainnet.meet.one
  - https://eos.newdex.one
  - https://api.eos.education
  - https://api.eosauthority.com
  - https://api.eoscleaner.com
  - https://api.zbeos.com
  - https://api.eos.wiki
  - https://api.blockpool.com
  - https://api.bitmars.one
  - https://api.eosrio.io
  - https://api.eosargentina.io
  - https://api.eostitan.com
  - https://api.cypherglass.com
  - https://api.acroeos.one
  - https://mainnet.genereos.io
MatchContentType: #需要替换的ContentType类型
  enable: false   #开关
  ContentType:
    - text/html
    - application/x-javascript
    - text/plain
    - text/javascript
    - application/javascript
    - application/x-javascript
    - application/xhtml+xml
  replace: #替换字符 格式说明  old: new
    morecoin.zendesk.com: morecoin.zendesk.more.top
```
###dokcer 运行
```
docker run -it -d 1228022817/proxy:latest
```
###docker-compose
```
version: '3'
services:
  proxy:
    image: 1228022817/proxy:latest
    container_name: proxy
    ports:
      - "8080:80"
    restart: always
    volumes:
      - /home/localhost/conf.yaml:/conf.yaml
```
