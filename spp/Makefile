spp:
	#osx only works with CGO_ENABLED=1
	#CGO_ENABLED=1 go build -ldflags="-s -w" -tags nocgo -o ./server .
	GOOS=linux GOARCH=386 CGO_ENABLED=0 go build -ldflags="-s -w" -tags nocgo -o ./server .
