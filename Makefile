format:
	yamlfmt -w config.yml
	go fmt

make:
	GOOS=linux GOARCH=amd64 go build -o generator generator.go

deploy: format make
	rsync -avh --delete --exclude index.html --exclude .DS_Store dist/* ipv6-overview.xyz:/home/veloc1ty/html/overview/dist
	rsync -avh config.yml generator index.html.gohtml ipv6-overview.xyz:/home/veloc1ty/html/overview
