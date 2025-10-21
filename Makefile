test:
	go test -cover ./internal

clean:
	cd internal/testgraphs && rm -f *.dot *.png *.src
	cd internal && rm -f *.dot *.png *.src

todo:
	cd internal && go test -v | grep TODO

bench:
	go test -benchmem -bench=. ./internal

install:
	cd cmd/gi && go install

unused:
	go run honnef.co/go/tools/cmd/staticcheck@latest -checks U1000 ./...