## how to run

    gi run .

## how to trace

    GI_TRACE=1 gi run .

## how to visualize graph

    GI_CALL=call.dot gi run .
    dot -Tsvg -o call.svg call.dot

## how to visualize ast

    GI_AST=ast.txt gi run .