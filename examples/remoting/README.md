# remoting

Remote execution of Go source on the server upon handling a HTTP request.

## start server

    go run .

## send Go source to execute

    curl -v --data-binary "@function.script" "http://localhost:8080/gi?func=doit"