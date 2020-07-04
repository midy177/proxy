##proxy是什么?
后端服务器代理应用，同时支持设置需要校验uri请求状态码是否符合需求，当不符合时尝试请求所有后端
服务，直到代理请求返回的状态码为200时将代理请求结果转发给请求，或者遍历后端服务器组无理想状态
时将最后请求的后端服务器返回信息返回给用户
者

##配置文件，yaml格式, conf.yaml

```
listenport: 8080   #监听端口
heathcheck:
  timeout: 1    #代理请求超时时间
verifyuri:     #需要校验请求状态码直到为200的url，是一个数组
  - /v1/chain/get_info
  - /v1/chain/get_table_rows
  - /v1/chain/push_transaction
baseurl:      #后端服务器组列表，是一个数组
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
MatchContentType: #需要替换文本内容的文件类型
  - text/html
  - application/x-javascript
  - text/plain
  - text/javascript
  - application/javascript
  - application/x-javascript
  - application/xhtml+xml
urlreplace: #替换的文本 替换前: 替换后
  morecoin.zendesk.com: morecoin.zendesk.more.top
  support.morecoin.com: morecoin.zendesk.more.top
  static.zdassets.com: morecoin.zendesk.more.top
  p19.zdassets.com: morecoin.zendesk.more.top
```
