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

# go install golang.org/x/tools/cmd/deadcode@latest
unused:	
	cd cmd/gi && deadcode .

.PHONY: test clean todo bench install unused examples
examples: install
	cd examples/api_call && go run .
	# cd examples/externalpkg && gi run .
	cd examples/nestedloop && gi run .
	# cd examples/subpkg && GI_TRACE=1 gi run .
	