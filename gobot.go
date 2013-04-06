package main

import "fmt"
import "regexp"
import "strings"
import irc "github.com/fluffle/goirc/client"
import "github.com/fluffle/golog/logging"

var bot_nick string = "gobot"
var bot_server string = "art1.mae.ncsu.edu:6667"
var bot_channel string = "#arc"
//var bot_nick string = "prattmic_gobot"
//var bot_server string = "rajaniemi.freenode.net:6667"
//var bot_channel string = "#ncsulug"

var last_message map[string]string

func main() {
    last_message = make(map[string]string)

    // Create a config and fiddle with it first:
    cfg := irc.NewConfig(bot_nick)

    //cfg.SSL = true
    cfg.Server = bot_server

    cfg.NewNick = func(n string) string { return n + "^" }
    c := irc.Client(cfg)

    // Add handlers to do things here!
    // e.g. join a channel on connect.
    c.HandleFunc("connected",
        func(conn *irc.Conn, line *irc.Line) {
            logging.Info("Connected to %s as %s", cfg.Server, cfg.Me.Nick);
            conn.Join(bot_channel)
        })

    // And a signal on disconnect
    quit := make(chan bool)
    c.HandleFunc("disconnected",
        func(conn *irc.Conn, line *irc.Line) { 
            logging.Info("Disconnected");
            quit <- true
        })

    c.HandleFunc("PRIVMSG", privmsg)

    // Tell client to connect.
    if err := c.Connect(); err != nil {
        fmt.Printf("Connection error: %s\n", err)
        return
    }

    // Wait for disconnect
    <-quit
}

func privmsg(conn *irc.Conn, line *irc.Line) {
    channel := line.Args[0]
    message := strings.Join(line.Args[1:], "")

    logging.Info("%s to %s: %s", line.Nick, channel, message);

    /* Yay! Regex!
     * Regex for matching vim-style replacement lines.
     * Input is of the form "s/regex[/replacement[/flags]]"
     * Where replacement, flags, and their assosciated slashes are optional
     * The input allows escaped slashes (\/) in the regex or replacement
     * When used with FindSubmatch, match[1] = regex, match[2] = replacement, match[3] = flags
     * As a single line: "s/((?:\\\\/|[^/])+)(?:/((?:\\\\/|[^/])*)(?:/((?:[ig])*))?)?" */
    reg, err := regexp.Compile("s/" +   // Beginning of search command
                                "(" +   // Start search regex group (1)
                                    "(?:\\\\/|[^/])+" + // One or more (you have to provide something to search for!) non-slashes (/), unless slash is escaped (\/)
                                                        // 4 backslashes are required to match one backslash, as it is an escape character for Go and for the regexp engine
                                                        // The engine will only match a literal \ if you escape it, thus \\. but to create a literal \ in Go, you also must
                                                        // escape that
                                ")" +   // End search regex group (1)
                                "(?:" + // Start non-capturing group for the remainder of the string
                                        // This group is optional (note the ? at the very end of the expression), meaning "s/blah" is valid, replacing blah with nothing
                                    "/" +   // Match / separating search and replacement
                                    "(" +   // Start replacement group (2)
                                        "(?:\\\\/|[^/])*" + // Zero or more (can replace with nothing) non-slashes (/), unless slash is escaped (\/)
                                    ")" +   // End replacement group (2)
                                    "(?:" + // Start non-capturing group for the remainder of the string (/flags)
                                            // This group is optional (note the ? at the end of this section, meaning "s/blah/bleh" is valid, replacing blah with bleh
                                        "/" +   // Match / separating replacement and flags
                                        "(" +   // Start flags group (3)
                                            "[ig]*" +   // Zero or more i (ignore case) or g (global replace) flags
                                        ")" +   // End flags group (3)
                                    ")?" +  // End optional group for (/flags)
                                ")?") // End optional group for (/replacement/flags)

    if err != nil {
        logging.Warn("Regexp compile error: %s", err)
    }

    match := reg.FindStringSubmatch(message)
    if err != nil {
        logging.Warn("Regexp error: %s", err)
    }

    if match != nil {
        logging.Info("Complete match: \"%s\"", match[0])
        logging.Info("Matched regex: \"%s\"", match[1])
        logging.Info("Matched replacement: \"%s\"", match[2])
        logging.Info("Matched flags: \"%s\"", match[3])
    }

    if match != nil {
        replacement_regex := match[1]
        repl := match[2]
        flags := match[3]
        global := false

        // Replacement must be given
        if replacement_regex == "" {
            return
        } 

        // Replace escaped slashes with real ones
        replacement_regex = strings.Replace(replacement_regex, "\\/", "/", -1)
        repl = strings.Replace(repl, "\\/", "/", -1)

        // Ignore case
        if strings.Contains(flags, "i") {
            replacement_regex = fmt.Sprintf("(?i)%s", replacement_regex)
        }

        // Global
        if strings.Contains(flags, "g") {
            global = true
        }

        m, ok := last_message[line.Nick]
        if !ok {
            return
        }

        regex, err := regexp.Compile(replacement_regex)
        if err != nil {
            conn.Privmsg(channel, fmt.Sprintf("%s: error compiling regex: %s", line.Nick, err))
            return
        }

        i := 0
        fixed := regex.ReplaceAllStringFunc(m, func (s string) string {
            if global {
                return repl
            } else if i < 1 {
                i = i + 1
                return repl
            }

            return s
        })

        conn.Privmsg(channel, fmt.Sprintf("<%s>: %s", line.Nick, fixed))

        last_message[line.Nick] = fixed
    } else {
        last_message[line.Nick] = message
    }
}
