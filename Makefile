.PHONY: all blog dairy

all:
	go run main.go -type all

blog:
	go run main.go -type blog

diary:
	go run main.go -type diary
