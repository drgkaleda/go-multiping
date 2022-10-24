APPNAME=mping

all:
	CGO_ENABLED=0 go build -o $(APPNAME) cmd/main.go

clean:
	go clean
	rm -f $(APPNAME)
