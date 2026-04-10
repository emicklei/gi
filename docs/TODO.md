## TODO

- About types: https://github.com/golang/example/tree/master/gotypes
- how to handle concurrency. (eval -> native, walk -> simulated?)
- clear with a pointer to a var?
- github.com/fatih/structtag replace with some SDK pkg?
- how to handle makeType of FuncType? and what if FuncType is using local pkg types?
- handle omitzero
- stdtypes is now a two-stage map => make it one big map ??
- generics: https://ehabterra.github.io/ast-extracting-generic-function-signatures
- deprecate varvoy?
- introduce routines (threads) in the VM
- https://www.geeksforgeeks.org/go-language/reflect-makefunc-function-in-golang-with-examples/


## potential blockers
- stepping happens per go-routine; what to do with the others when controlling one of them?
