# Pearl

A scripting language that takes the good parts of Perl (text processing, regex, pragmatism) and leaves behind the bad (line noise, TMTOWTDI, cryptic sigils).

Based on the interpreter from Thorsten Ball's "Writing an Interpreter in Go", but extended significantly.

## Building

```bash
go build -o pearl .
```

## Usage

```bash
# start the REPL
./pearl

# run a file
./pearl examples/hello.pearl
./pearl -f examples/text.pearl

# evaluate an expression
./pearl -e 'print("hello")'

# check syntax without running
./pearl -check -f myfile.pearl
```

## Quick Tour

### Variables (no sigils!)

```pearl
let name = "alice"
let age = 30
let pi = 3.14159
let active = true
let nothing = null
```

### String Interpolation

```pearl
let x = 10
print("x is {x}")
print("x squared is {x * x}")
```

### Arrays and Maps

```pearl
let fruits = ["apple", "banana", "cherry"]
let person = {"name": "bob", "age": 25}

print(fruits[0])        # apple
print(person["name"])   # bob
```

### Control Flow

```pearl
if x > 10 {
    print("big")
} else {
    print("small")
}

for item in items {
    print(item)
}

for i in 0..10 {
    print(i)
}

while x > 0 {
    x = x - 1
}
```

### Functions

```pearl
fn greet(name, loud = false) {
    if loud {
        print("HELLO {upper(name)}!")
    } else {
        print("hello {name}")
    }
}

greet("world")
greet("world", loud = true)
```

### Regex

```pearl
let text = "email me at bob@test.com"

# match check
if text ~ /\w+@\w+\.\w+/ {
    print("found email!")
}

# extract groups
let m = match(text, /(\w+)@(\w+)\.(\w+)/)
print(m[1])  # bob

# replace
let clean = replace(text, /\w+@\w+\.\w+/, "[REDACTED]")
```

### Pipelines

```pearl
let result = "  HELLO WORLD  "
    |> trim()
    |> lower()
    |> split(" ")
    |> join("-")
# result is "hello-world"
```

### Functional Style

```pearl
let nums = [1, 2, 3, 4, 5]

let doubled = map(nums, fn(x) { x * 2 })
let evens = filter(nums, fn(x) { x % 2 == 0 })
let sum = reduce(nums, fn(a, b) { a + b }, 0)
```

## Built-in Functions

### String Functions
- `len(s)` - length
- `upper(s)`, `lower(s)` - case conversion
- `trim(s)`, `ltrim(s)`, `rtrim(s)` - whitespace removal
- `split(s, delim)` - split into array
- `join(arr, delim)` - join array into string
- `substr(s, start, len)` - substring
- `contains(s, needle)` - check if contains
- `starts_with(s, prefix)`, `ends_with(s, suffix)`
- `replace(s, old, new)`, `replace_all(s, old, new)`
- `repeat(s, n)` - repeat n times
- `reverse(s)` - reverse string
- `lines(s)` - split by newlines
- `chars(s)` - split into characters
- `find(s, needle)` - find index

### Regex Functions
- `match(s, regex)` - returns array of matches or null
- `match_all(s, regex)` - returns all matches
- `regex(pattern)` - compile a regex from string

### Array Functions
- `len(arr)` - length
- `push(arr, item)`, `pop(arr)` - end operations
- `shift(arr)`, `unshift(arr, item)` - start operations
- `slice(arr, start, end)` - sub-array
- `sort(arr)` - sort (returns new array)
- `reverse(arr)` - reverse
- `unique(arr)` - remove duplicates
- `flatten(arr)` - flatten nested arrays
- `contains(arr, item)` - check membership
- `find(arr, item)` - find index

### Functional
- `map(arr, fn)` - transform each element
- `filter(arr, fn)` - keep matching elements
- `reduce(arr, fn, init)` - reduce to single value

### Map Functions
- `keys(map)` - get all keys
- `values(map)` - get all values

### Type Conversion
- `int(x)`, `float(x)`, `str(x)`
- `type(x)` - get type as string

### Other
- `print(...)` - output
- `range(n)` or `range(start, end)` - create range

## Why "Pearl"?

It's like Perl, but:
- No `$@%` sigils
- No `$_` magic variable
- No TMTOWTDI (there's more than one way to do it)
- Readable by default
- Good error messages
- One obvious way to do things

The gem, not the mess.
