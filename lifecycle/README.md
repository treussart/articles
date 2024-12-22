# github.com/treussart/articles/lifecycle

## Run go program

````bash
go run -C lifecycle main.go task.go handlers.go recover.go
curl http://localhost:9000/task
ctrl-c
````

Output :

```
{"level":"info","service":"lifecycle","duration":0.000399375,"time":"2024-12-22T18:16:55+01:00","message":"Service started successfully"}
{"level":"info","service":"lifecycle","time":"2024-12-22T18:16:55+01:00","message":"Performing task started..."}
{"level":"info","service":"lifecycle","time":"2024-12-22T18:16:59+01:00","message":"Performing task handler started..."}
^C{"level":"info","service":"lifecycle","time":"2024-12-22T18:17:05+01:00","message":"HTTP server cancelled"}
{"level":"info","service":"lifecycle","duration":9.955820834,"time":"2024-12-22T18:17:05+01:00","message":"Gracefully shutting down service..."}
{"level":"info","service":"lifecycle","time":"2024-12-22T18:17:05+01:00","message":"Performing task ended..."}
{"level":"info","service":"lifecycle","time":"2024-12-22T18:17:05+01:00","message":"Task cancelled"}
{"level":"info","service":"lifecycle","time":"2024-12-22T18:17:09+01:00","message":"Performing task handler ended..."}
{"level":"info","service":"lifecycle","duration":4.224514959,"time":"2024-12-22T18:17:09+01:00","message":"Shutdown service complete"}
```
