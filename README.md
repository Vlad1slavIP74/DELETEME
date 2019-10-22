# Setup

```
go get -u golang.org/x/vgo
./create_database.sh
```

# Build the server

```
vgo build
```

# Run the server

```
./example
```

# Make a request

```
curl -v localhost:8000/people
```
