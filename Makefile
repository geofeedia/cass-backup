all: deps build
	    
deps:
	go get -u github.com/rjeczalik/notify
	go get -u github.com/aws/aws-sdk-go/aws
	go get -u github.com/aws/aws-sdk-go/aws/session
	go get -u github.com/aws/aws-sdk-go/service/s3/s3manager
	go get -u github.com/aws/aws-sdk-go/aws/ec2metadata
	go get -u google.golang.org/cloud/compute/metadata
	go get -u golang.org/x/net/context
	go get -u golang.org/x/oauth2/google
	go get -u google.golang.org/api/storage/v1

build:
	export CGO_ENABLED=0
	export GOOS=linux
	go build -ldflags "-s" -a -installsuffix cgo -o ./cass-backup