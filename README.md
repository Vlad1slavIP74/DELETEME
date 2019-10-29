# Setup

```
go get -u golang.org/x/vgo
./create_database.sh
```

# Run the server

```
go run ./cmd/server
``` from directory server

# Make a request

```
curl -v localhost:8000/list
```
