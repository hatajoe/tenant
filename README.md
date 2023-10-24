# tenant

WIP

tenant is an intercepter proxy for local web application development.

# how it works

Assuming a server running in the cloud, start server 1. (PORT 8081)
```
% go run cmd/example/server1/main.go
```

Assuming a server running in the local, start server 2. (PORT 8082)
```
% go run cmd/example/server2/main.go
```

Assuming a tenant running in the same VPC as server1. (PORT 8080)
```
% go run cmd/tenant/main.go
```

tenant pass through all HTTP requests to server1.
```
% curl localhost:8080/
server1
```

if you want to intercept all requests to server1 and forward them to server2, run tenantctl

```
% go run cmd/tenantctl/main.go
```

no header:
```
% curl localhost:8080/
server1
```

add header:
```
% curl -H "USER: hatajoe" localhost:8080/
server2
```
