# github.com/treussart/articles/http/server

[Link to published article](https://medium.com/@matthieu.treussart/golang-http-request-management-on-the-server-side-c7f83fba2d6a)

## Stdlib

### Simulate client-side cancelled requests

Term 1:
````bash
go run -C http/stdlib main.go handlers.go recover.go
````

Term 2:
````bash
curl --max-time 1 http://localhost:9000/taskcancel
````

Output :

```
{"level":"info","service":"http/server","time":"2024-12-23T08:33:13+01:00","message":"Performing task handler started..."}
{"level":"warn","service":"http/server","error":"context canceled","time":"2024-12-23T08:33:14+01:00","message":"handleRequestCtx error"}
```

### Simulate server-side cancelled requests

Term 1:
````bash
go run -C http/stdlib main.go handlers.go recover.go
````

Term 2:
```bash
curl  http://localhost:9000/taskcancel
```

Output :

```
{"level":"info","service":"http/server","time":"2024-12-23T08:32:50+01:00","message":"Performing task handler started..."}
{"level":"warn","service":"http/server","error":"context deadline exceeded","time":"2024-12-23T08:32:55+01:00","message":"handleRequestCtx error"}
```


## Gin

### Simulate client-side cancelled requests

Term 1:
````bash
go run -C http/gin main.go handlers.go recover.go timeout.go logger.go
````

Term 2:
````bash
curl --max-time 1 http://localhost:9000/taskcancel
````

Output :

```
{"level":"info","service":"http/gin","time":"2024-12-23T21:28:00+01:00","message":"Performing task handler started..."}
{"level":"warn","service":"http/gin","net.ip.source":"::1","http.path":"/taskcancel","http.method":"GET","http.proto":"HTTP/1.1","error":"context canceled","http.duration_seconds":1.006406,"http.status_code":408,"time":"2024-12-23T21:28:01+01:00"}
```

### Simulate server-side cancelled requests

Term 1:
````bash
go run -C http/gin main.go handlers.go recover.go timeout.go logger.go
````

Term 2:
```bash
curl  http://localhost:9000/taskcancel
```

Output :

```
{"level":"info","service":"http/gin","time":"2024-12-23T21:28:51+01:00","message":"Performing task handler started..."}
{"level":"warn","service":"http/gin","net.ip.source":"::1","http.path":"/taskcancel","http.method":"GET","http.proto":"HTTP/1.1","error":"context deadline exceeded","http.duration_seconds":5.00140075,"http.status_code":408,"time":"2024-12-23T21:28:56+01:00"}
```
