

.PHONY: blog dairy

blog:
	go run main.go -type blog

diary:
	go run main.go -type diary
