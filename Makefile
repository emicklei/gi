gen:
	cd cmd/genstdlib && go run .

test:
	GI_TRACE=1 go test -cover ./internal

serial:
	GI_TRACE=1 go test -v -p 1 ./internal

clean:
	cd internal/testgraphs && rm -f *.dot *.png *.src *.svg *.ast
	cd internal && rm -f *.dot *.png *.src *.svg *.ast

todo:
	cd internal && go test -v | grep SKIP

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
	cd examples/subpkg && GI_TRACE=1 gi run .
	
# runs all programs provided by GoByExample and reports failures
.PHONY: gobyexample
gobyexample:
	cd cmd/treerunner && go install
	if [ ! -d "gobyexample" ]; then git clone https://github.com/mmcgrana/gobyexample; fi	
	mkdir -p gobyexample/examples/output
	cd gobyexample/examples && GI_IGNORE_PANIC=1 GI_IGNORE_EXIT=1 treerunner -report output/treerunner-report.json .
