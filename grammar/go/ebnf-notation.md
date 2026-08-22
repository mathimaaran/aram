# Go EBNF notation (from Go spec go1.26)

The syntax is specified using a variant of Extended Backus-Naur Form (EBNF):

```
Syntax      = { Production } .
Production  = production_name "=" [ Expression ] "." .
Expression  = Term { "|" Term } .
Term        = Factor { Factor } .
Factor      = production_name | token [ "…" token ] | Group | Option | Repetition .
Group       = "(" Expression ")" .
Option      = "[" Expression "]" .
Repetition  = "{" Expression "}" .
```

Operators, increasing precedence:

```
|   alternation
()  grouping
[]  option (0 or 1 times)
{}  repetition (0 to n times)
```

- Lowercase names = lexical (terminal) tokens.
- CamelCase = non-terminals.
- Tokens in `""` or back quotes.
- `a … b` = character range.

Niraluli Tamil EBNF should use this same notation so agents and humans share one style.
