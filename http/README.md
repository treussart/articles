# github.com/treussart/articles/http

## Run go program

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
