# Go core EBNF (curated extract)

Source: Go spec go1.26. Curated for Niraluli design — not the full grammar.

## Lexical (selected)

```
letter        = unicode_letter | "_" .
decimal_digit = "0" … "9" .
identifier    = letter { letter | unicode_digit } .
```

### Go keywords (reference only)

```
break        default      func         interface    select
case         defer        go           map          struct
chan         else         goto         package      switch
const        fallthrough  if           range        type
continue     for          import       return       var
```

## Declarations (selected)

```
Declaration  = ConstDecl | TypeDecl | VarDecl .
TopLevelDecl = Declaration | FunctionDecl | MethodDecl .

IdentifierList = identifier { "," identifier } .

VarDecl = "var" ( VarSpec | "(" { VarSpec ";" } ")" ) .
VarSpec = IdentifierList ( Type [ "=" ExpressionList ] | "=" ExpressionList ) .

ShortVarDecl = IdentifierList ":=" ExpressionList .

FunctionDecl = "func" FunctionName [ TypeParameters ] Signature [ FunctionBody ] .
FunctionName = identifier .
FunctionBody = Block .
```

## Blocks and statements (selected)

```
Block         = "{" StatementList "}" .
StatementList = { Statement ";" } .

Statement  = Declaration | LabeledStmt | SimpleStmt |
             GoStmt | ReturnStmt | BreakStmt | ContinueStmt | GotoStmt |
             FallthroughStmt | Block | IfStmt | SwitchStmt | SelectStmt | ForStmt |
             DeferStmt .

SimpleStmt = EmptyStmt | ExpressionStmt | SendStmt | IncDecStmt | Assignment | ShortVarDecl .

ExpressionStmt = Expression .
Assignment     = ExpressionList assign_op ExpressionList .

IfStmt = "if" [ SimpleStmt ";" ] Expression Block [ "else" ( IfStmt | Block ) ] .

ForStmt = "for" [ Condition | ForClause | RangeClause ] Block .

ReturnStmt = "return" [ ExpressionList ] .
```

## Packages (selected)

```
SourceFile    = PackageClause ";" { ImportDecl ";" } { TopLevelDecl ";" } .
PackageClause = "package" PackageName .
PackageName   = identifier .

ImportDecl = "import" ( ImportSpec | "(" { ImportSpec ";" } ")" ) .
```

## Program entry (semantic note)

Go: package name `main` + `func main()` with no params/results.

Niraluli Tamil-0 plans an analogous entry via semantic keywords
(`தொகுப்பு தொடக்கம்` + `செயல்பாடு தொடக்கம்`) — exact EBNF in `grammar/tamil/`.
