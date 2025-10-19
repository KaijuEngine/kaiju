# How to contribute to Kaiju Engine

## Contributing via Git platform
You are free to contribute to the project via git through bug reports, feature
requests, pull requrests, discussions, video tutorials, etc. Below are some
guides on how to use the platform to contribute in this way.

### **Did you find a bug?**
1. Search issues, both open and closed first
2. If you didn't find the bug, report it via issues with bug tag

### **Pull requeset rules**
1. Ensure you've discussed the issue/addition before starting (issues, discussion, etc.)
2. You must make a pull request to the `staging` branch
3. Title must be short and self explanitory
4. Give a detailed description of the change, why it was made, and what it solves
5. Pull request title should include issue number (eg: #1234)

## Coding guidelines
Every attempt is made to make the code as performant as possible as well as
generate minimal memory garbage. All new code must be thoroughly planned and
designed before being written. This can be via technical design doc, flow
charts, and/or any other type of specification document. Your go code should be
well written prose, fancy code, tricks, and bespoke patterns are fun to code but
not typically welcome.

### Comments and documentation
All public functions, types, and fields must have clean, readable, thorough, and
expressive comments to describe the intent. The comment format should be
formatted the same way as Go's standard library. Comment lines should not
exceed 80 columns in width.

Comments within the code are welcome for when the code is not possible to
express itself in an understandable way. This is typical in tight performance
loops, or code needed to access low-level resources. If your code is otherwise
difficult to understand and needs a comment, consider improving your code first
before writing a comment.

Do not commit TODO, FIXME, or any other sorts of similar comments without first
discussing why it needs to be there and getting approval for it's addition. No
such comment should be committed without an accompanying issue id, regardless
of it's approval. If you create an issue, remove the TODO or FIXME comment and
add in the klib.NotYetImplemented(X) function call, replacing `X` with the id
of the issue.

*Note that the implementation of the `Error() string` error interface public
function does not need to be documented, even though it is a public function.*

### Pointers
Pointers are to be deliberately hand selected and used as sparingly as possible.
Prefer composition of structures with members into a single pointer over
creating multiple pointers that can be passed around. This will require
forethought and thorough design to reduce mistakes. Please review `host.go` for
an example on how this mediator is used to access various systems without
over-use of pointers.

### Errors
Though it's enticing to simply return `fmt.Errorf` or `errors.New`, these are
frowned upon. Having a structure that implements the Error interface is the
preferred method for the Go source code, and so too is it to be the preferred
method within the engine. Typically errors stem from uncontrollable sources,
but make every attempt to resolve the error with a fallback solution as soon as
possible and avoid bubbling up the error if at all possible.

### Interfaces
Interfaces should be used sparingly, only when no other solution is possible.
Typically an interface is to solve an unknown problem that a future developer
may need, or to create a more generic way to interact with a part of a system.
Most of the time, interfaces are not needed. Most interfaces built into the
engine are for generic type constraints, bi-directional communication between
packages, and solutions to larger problems like HTML/CSS parsing.

### 3rd party packages
It is our goal to keep 3rd party packages as minimal as possible. We have an
intent to one day replace all 3rd party packages with our own solutions. Please
do not add any other 3rd party packages into the engine.

### Logging
Use `slog` to write your logs. We've implemented a base logging mechanism
through this interface and may extend it in the future.

### Assembly code
When writing assembly code, ensure that you are correctly locking it to the
target system with the go build flags, as well as providing a fallback method
in Go code. Since you are creating a fallback method in Go anyway, you must
create a benchmark to prove that your assembly implementation is superior to
the go implementation of the code.

### In any other case
Generally, if you would like a guide at how the code should be formatted and
what standards you should hold yourself to, review the existing code in the
repository. When in Rome, do as the Romans do.
