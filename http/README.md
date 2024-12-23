# github.com/treussart/articles/http

## Run go program

Term 1:
````bash
go run -C http/stdlib main.go handlers.go recover.go
````

Term 2:
````bash
curl --max-time 3 http://localhost:9000/taskcancel
````

Output :

```
{"level":"info","service":"http","time":"2024-12-23T06:31:38+01:00","message":"Performing task handler started..."}
{"level":"info","service":"http","time":"2024-12-23T06:31:46+01:00","message":"Performing task handler ended..."}
```


Term 1:
````bash
go run -C http/stdlib main.go handlers.go recover.go
````

Output :

```
{"level":"info","service":"lifecycle","time":"2024-12-23T00:22:38+01:00","message":"Performing task handler started..."}
{"level":"info","service":"lifecycle","time":"2024-12-23T00:22:48+01:00","message":"Performing task handler ended..."}
```

Term 2:
````bash
curl  http://localhost:9000/taskcancel
````

Output :

```
server timed out, request exceeded 8s
```


```bash
go run -C http/gin main.go handlers.go recover.go logger.go
```

Avec Gin (handler http.HandlerFunc) :

```bash
curl --max-time 1 http://localhost:9000/taskcancel
```

Output : 

```
{"level":"info","service":"http/gin","time":"2024-12-23T07:03:47+01:00","message":"Performing task handler started..."}
{"level":"warn","service":"http/gin","net.ip.source":"::1","http.path":"/task","http.method":"GET","http.proto":"HTTP/1.1","http.duration_seconds":1.001911083,"http.status_code":408,"time":"2024-12-23T07:03:48+01:00"}
{"level":"info","service":"http/gin","time":"2024-12-23T07:03:55+01:00","message":"Performing task handler ended..."}
```

Avec Gin (handler gin.HandlerFunc) :

```bash
curl --max-time 1 http://localhost:9000/taskcancelgin
```

Output : 

```
{"level":"info","service":"http/gin","time":"2024-12-23T07:04:58+01:00","message":"Performing task handler started..."}
{"level":"warn","service":"http/gin","net.ip.source":"::1","http.path":"/taskgin","http.method":"GET","http.proto":"HTTP/1.1","error":"context canceled","http.duration_seconds":1.0075145,"http.status_code":408,"time":"2024-12-23T07:04:59+01:00"}
{"level":"info","service":"http/gin","time":"2024-12-23T07:05:06+01:00","message":"Performing task handler ended..."}
```
