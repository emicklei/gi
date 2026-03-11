## TODO

- About types: https://github.com/golang/example/tree/master/gotypes
- how to handle concurrency. (eval -> native, walk -> simulated?)
- fallthrough cannot be used with a type switch.
- clear with a pointer to a var?
- drop StructValues?, use https://stackoverflow.com/questions/57567466/create-a-struct-by-reflection-in-go  this only support Exported fields. Cannot use it.
. fmt.Println for StructValues needs rework
- github.com/fatih/structtag replace with some SDK pkg?
- how to handle makeType of FuncType? and what if FuncType is using local pkg types?
- handle omitzero
- stdtypes is now a two-stage map => make it one big map
- generics: https://ehabterra.github.io/ast-extracting-generic-function-signatures
- should the unnamed results of a function be named?
- put generated code in generated package
- deprecate varvoy?
- introduce routines (threads) in the VM
- https://www.geeksforgeeks.org/go-language/reflect-makefunc-function-in-golang-with-examples/


## potential blockers
- reflect structs can only have exposed fields. for that reason StructValues was created but the SDK is not aware of this. For example, fmt.Println might not work correctly with StructValues.
- stepping happens per go-routine; what to do with the others when controlling one of them?
- should undeclared know the looked-up name and use it later in the flow?  Price is Selector, not Value.

## ideas for Gi Playground
- https://godbolt.org/
- call graphs in tab view
- structexplorer to see all environments and stackframes
