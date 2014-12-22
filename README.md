replacebot
==========

replacebot is an IRC bot written in Go that provides a Vim/Perl/Sed-like regex replacement syntax.

The is, whenever a user enters "s/regex/replacement/flags", their previous message will be replaced using the regex and flags.

Input is of the form "s/regex[/replacement[/flags]]", where replacement, flags, and their associated slashes are optional. The input allows escaped slashes (\/) in the regex or replacement.

Supported flags:
* 'i' (ignore case)
* 'g' (global replace)

The defaults are case sensitive and single replacement.

The user's regex is compiled by Go's [regexp](http://golang.org/pkg/regexp/) package, which supports [re2](https://code.google.com/p/re2/wiki/Syntax) syntax.  See these sites for specifics on the regex format.

Conversations will go something like this:

    <awesomeguy>: 321 Go!
    <lameguy>: I'm winning!
    <awesomeguy>: s/\d+/Go Go
    <replacebot>: <awesomeguy>: Go Go Go!

Building
---------------
replacebot should build like a standard go application, within your `GOPATH`.

    go get github.com/prattmic/replacebot
