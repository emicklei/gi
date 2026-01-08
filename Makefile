gen:
	cd cmd/genstdlib && go run .

test: clean examples
	go test -p 16 ./pkg

serial:
	GI_CALL=out.dot GI_AST=test.ast GI_TRACE=1 go test -v -p 1 ./pkg

clean:
	cd pkg/internal/testgraphs && rm -f *.dot *.png *.src *.svg *.ast *.ast.pkg
	cd pkg && rm -f *.dot *.png *.src *.svg *.ast *panic.html

skip:
	cd pkg && go test -v | grep SKIP

bench:
	go test -benchmem -bench=. ./pkg

install:
	cd cmd/gi && go install

# brew install golangci-lint
lint:
	golangci-lint run

# go install golang.org/x/tools/cmd/deadcode@latest
unused:
	cd cmd/gi && deadcode -test .

.PHONY: test clean todo bench install unused examples lint
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
	cd gobyexample/examples && GI_IGNORE_PANIC=1 GI_IGNORE_EXIT=1 treerunner -report output/treerunner-report.json -skip "line-filters" -urlprefix "https://github.com/mmcgrana/gobyexample/tree/master/examples" .
