package main

import "fmt"
import "regexp"
import "strings"
import irc "github.com/fluffle/goirc/client"
import "log"

const (
	BOT_NICK = "replacebot"
	//BOT_SERVER = "art1.mae.ncsu.edu:6667"
	//BOT_CHANNEL = "#arc"
	BOT_SERVER  = "chat.freenode.net:6667"
	BOT_CHANNEL = "#ncsulug"
	BOT_SOURCE  = "https://github.com/prattmic/replacebot"
)

var last_message map[string]string

// Yay! Regex!
// Regex for matching {vim,perl,sed}-style replacement lines.
// Input is of the form "user: s/regex[/replacement[/flags]]"
// Where user, replacement, flags, and their assosciated slashes are optional
// The input allows escaped slashes (\/) in the regex or replacement
// When used with FindSubmatch, match[1] = user, match[2] = regex, match[3] = replacement, match[4] = flags
// As a single line: "^(?:(.+?)[:, ]\\s*)?s/((?:\\\\/|[^/])+)(?:/((?:\\\\/|[^/])*)(?:/((?:[ig])*))?)?" */
var search *regexp.Regexp = regexp.MustCompile(
	// Start at the beginning of the line
	`^` +
		// Match arbitraru (optional) username in group (1), with a delimiting character and some whitespace
		`(?:(.+?)[:, ]\s*)?` +
		// Beginning of search command
		`s/` +
		// Match search regexp in group (2)
		// One or more (you have to provide something to search for!) non-slashes (/), unless slash is escaped (\/)
		`((?:\\/|[^/])+)` +
		// Everything after the search regexp is optional (`s/blah` is valid and replaces with nothing)
		`(?:` +
		// Match / separating search and replacement
		`/` +
		// Match replacement string in group (3)
		`((?:\\/|[^/])*)` +
		// Everything after replacement is optional (/flags) (`s/blah/bleh` is valid)
		`(?:` +
		// Match / separating replacement and flags
		`/` +
		// Zero or more i (ignore case) or g (global replace) flags in group (4)
		`([ig]*)` +
		// End optional group for (/flags)
		`)?` +
		// End optional group for (/replacement/flags)
		`)?`)

func main() {
	last_message = make(map[string]string)

	c := irc.SimpleClient(BOT_NICK)

	// Add handlers to do things here!
	// e.g. join a channel on connect.
	c.HandleFunc("connected",
		func(conn *irc.Conn, line *irc.Line) {
			log.Printf("Connected to %s as %s", c.Config().Server, c.Config().Me.Nick)
			conn.Join(BOT_CHANNEL)
		})

	// And a signal on disconnect
	quit := make(chan bool)
	c.HandleFunc("disconnected",
		func(conn *irc.Conn, line *irc.Line) {
			log.Print("Disconnected")
			quit <- true
		})

	// Watch for messages
	c.HandleFunc("PRIVMSG", privmsg)

	// Tell client to connect.
	if err := c.ConnectTo(BOT_SERVER); err != nil {
		fmt.Printf("Connection error: %s\n", err)
		return
	}

	// Wait for disconnect
	<-quit
}

/* Watch for messages.
 * Keep track of last thing every user said, and if they perform a replacement,
 * do the replacement for them and respond */
func privmsg(conn *irc.Conn, line *irc.Line) {
	channel := line.Args[0]
	message := strings.Join(line.Args[1:], "")

	log.Printf("%s to %s: %s", line.Nick, channel, message)

	match := search.FindStringSubmatch(message)

	if match != nil {
		log.Printf("Complete match: \"%s\"", match[0])
		log.Printf("Matched user: \"%s\"", match[1])
		log.Printf("Matched regex: \"%s\"", match[2])
		log.Printf("Matched replacement: \"%s\"", match[3])
		log.Printf("Matched flags: \"%s\"", match[4])
	}

	if match != nil {
		user := match[1]
		replacement_regex := match[2]
		repl := match[3]
		flags := match[4]
		global := false

		// Replacement must be given
		if replacement_regex == "" {
			return
		}

		// Default to current user
		if user == "" {
			user = line.Nick
		}

		// Replace escaped slashes with real ones
		replacement_regex = strings.Replace(replacement_regex, "\\/", "/", -1)
		repl = strings.Replace(repl, "\\/", "/", -1)

		// Ignore case
		if strings.Contains(flags, "i") {
			replacement_regex = fmt.Sprintf("(?i)%s", replacement_regex)
		}

		// Global replacement
		if strings.Contains(flags, "g") {
			global = true
		}

		// User's last message
		m, ok := last_message[user]
		if !ok {
			return
		}

		// Apply their regex
		regex, err := regexp.Compile(replacement_regex)
		if err != nil {
			conn.Privmsg(channel, fmt.Sprintf("%s: error compiling regex: %s", line.Nick, err))
			return
		}

		i := 0
		fixed := regex.ReplaceAllStringFunc(m, func(s string) string {
			if global {
				return repl
			} else if i < 1 {
				// non-global replacement only done once
				i = i + 1
				return repl
			}

			return s
		})

		/* Limit result to one message
		 * My limited testing indicates that 438 (?)
		 * characters is the maximum */
		if len(fixed) > 438 {
			b := []byte(fixed)
			fixed = string(b[0:438])
		}

		// Send out replaced message
		conn.Privmsg(channel, fmt.Sprintf("<%s>: %s", user, fixed))

		// Consider this corrected message the user's last message
		last_message[user] = fixed
	} else {
		// Not a replacement, just remember this message
		last_message[line.Nick] = message

		// Send source link
		if strings.EqualFold(message, BOT_NICK+": source") {
			conn.Privmsg(channel, fmt.Sprintf("%s: "+BOT_SOURCE, line.Nick))
		} else if strings.EqualFold(message, BOT_NICK+": help") {
			conn.Privmsg(channel, fmt.Sprintf("%s: I will search and replace your last message when you "+
				"use the format s/regex/replacement/flags (don't direct at me). "+
				"Flags are i (ignore case) and g (global replacement).  "+
				"Replacements directed at a user will replace their last message.  "+
				"See source for more details.", line.Nick))
		}
	}
}
