# CommonWeb
通过上下行分流来绕过某些防火墙

## 原理
客户端主动建立两个连接，一个只上传数据，另一个只下载数据，服务端再把两个连接合成一个。  
底层传输协议使用http，下载使用GET请求，服务器返回分段数据（`Transfer-Encoding: chunked`）；上传使用POST请求，也是分段发送。

![image](https://raw.githubusercontent.com/sduoduo233/commonweb/master/image.jpg)

## 如何使用
服务端启动参数
```
-server -local 127.0.0.1:8080 -remote 127.0.0.1:8081
```
客户端启动参数
```
-local 127.0.0.1:8081 -upUrl http://your_domain.com/up -downUrl http://your_domain.com/down
```
Nginx
```nginx
server {
  ...

  # 可以把xxxxxx换成随机字符串防止主动探测
  location /up_xxxxxx {
    # 因为使用到 chunked 编码，所以必须加这几行
    proxy_http_version 1.1;
    proxy_buffering off; 
    proxy_request_buffering off;
    proxy_pass http://127.0.0.1:8080/up;
  }

  location /down_xxxxxx {
    proxy_http_version 1.1;
    proxy_buffering off;
    proxy_pass http://127.0.0.1:8080/down;
  }
}
```
v2ray server
```json
{
    "inbounds": [
        {
            "listen": "127.0.0.1",
            "port": 8081,
            "protocol": "vmess",
            "settings": {
                "clients": [
                    {
                        "id": "xxxxxx"
                    }
                ]
            }
        }
    ],
    "outbounds": [
        {
            "protocol": "freedom"
        }
    ]
}
```
v2ray client
```json
{
  "inbounds": [
    {
      "port": 8085,
      "protocol": "socks"
    }
  ],
  "outbounds": [
    {
      "protocol": "vmess",
      "settings": {
        "vnext": [
          {
            "address": "127.0.0.1",
            "port": 8081,
            "users": [
              {
                "id": "xxxxxx"
              }
            ]
          }
        ]
      }
    }
  ]
}
```