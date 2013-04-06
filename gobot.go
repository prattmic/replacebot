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

    // s/([^/]+)/([^/]+)(?:/([ig])*)*
    reg, err := regexp.Compile("s/(.+)/(.*)(?:/([ig])*)*")
    if err != nil {
        logging.Warn("Regexp compile error: %s", err)
    }

    match := reg.FindStringSubmatch(message)
    if err != nil {
        logging.Warn("Regexp error: %s", err)
    }

    logging.Info("%v", match)

    if match != nil {
        replacement_regex := match[1]
        repl := match[2]
        flags := match[3]
        global := false

        // Replacement must be given
        if replacement_regex == "" {
            return
        } 

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
