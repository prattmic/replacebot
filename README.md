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
replacebot uses the [goirc](https://github.com/fluffle/goirc) IRC framework, however it currently tracks the master branch, which is incomptible with the go1 branch that `go get` will fetch.

I recommend creating a src/github.com/fluffle/ folder in the replacebot directory, and cloning goirc there.  Then you can set your GOPATH to include the current directory, so the package will be found when you build.

    export GOPATH=$(pwd):$GOPATH

replacebot is also dependent upon [golog](https://github.com/fluffle/goirc), which is in turn dependent on [goevent](https://github.com/fluffle/goevent), however, the go1 branch of these works fine, so you should be able to `go get` them no problem.

The bot nick, server, and channel are specified as constants at the top of replacebot.go.  Modify them to your liking.  Joining multiple channels is not currently supported, but should be trivial to add.  Simply add multiple Join calls in the "connected" function handler.
