# cf-workers-triage
A Go and php based automated cloudflare worker triage system for high volume free cloudflare worker users (more suitable for alist users).

# Program description
## The Go language is responsible for computing cloudflare data
Need to add scheduled tasks to be executed every 2-5 minutes
## The php language is responsible for front-end calculations and 302 jumps to the specified url
### proxy.php
带有特定后缀反向代理直接返回功能的分流程序
Triage program with direct return from reverse proxies with specific suffixes

### go-workers.php
全局走workers的版本
Globally walk the version of workers

### get-links.php
直接获取源link的版本
Get the version of the source link directly

### use-cf-triage.php
使用cloudflare的国家返回头进行分流的版本
A version of triage using cloudflare's country return header

### Rewrite
    location / {
        try_files $uri $uri/ /index.php?$args;
    }

    location = /config.json {
        deny all;
        return 403;
    }
