package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"time"

	"github.com/lowstz/slackhookgo"
)

type alertMessage struct {
	url     string
	text    string
	color   string
	channel string
}

type connection struct {
	protocol string
	address  string
}

func CheckError(err error) bool {
	if err == nil {
		return false
	}
	log.Printf("error: %s", err)
	return true
}

func (c *connection) Conn() (net.Conn, bool) {
	conn, err := net.DialTimeout(c.protocol, c.address, 3*time.Second)
	errBool := CheckError(err)
	return conn, errBool
}

func (a *alertMessage) SentSlack() bool {
	msg := slackhookgo.NewSlackMessage(
		"Alert",
		a.channel,
	).AddAttachment(
		slackhookgo.MessageAttachment{
			Color: a.color,
			Text:  a.text,
			Title: "<!channel>",
		},
	)
	msg.IconEmoji = ":exclamation:"
	err := slackhookgo.Send(a.url, msg)
	return CheckError(err)
}

func main() {
	var (
		color     string
		status    string
		lastState bool
		protocol  = flag.String("protocol", "tcp", "protocol tcp/udp")
		host      = flag.String("host", "", "destination host")
		port      = flag.Uint("port", 80, "destination port")
		interval  = flag.Uint("interval", 15, "interval check seconds")
		url       = flag.String("url", "", "hook url")
		slack     = flag.Bool("slack", false, "use -slack for send alert in slack")
		channel   = flag.String("channel", "general", "slack channel")
		help      = flag.Bool("help", false, "use -help to see this information")
	)

	flag.StringVar(protocol, "p", "tcp", "protocol tcp/udp")
	flag.StringVar(host, "h", "", "destination host")
	flag.UintVar(interval, "i", 15, "interval check seconds")

	flag.Parse()

	if len(os.Args) == 1 {
		flag.PrintDefaults()
		os.Exit(1)
	} else if *help == true {
		flag.PrintDefaults()
		os.Exit(0)
	} else if *host == "" {
		log.Fatal("param '-host' is empty")
	} else if *slack == true {
		if *url == "" {
			log.Fatal("param '-url' is empty")
		}
	}

	c := connection{
		protocol: *protocol,
		address:  fmt.Sprintf("%s:%v", *host, *port),
	}

	for {
	begin:
		conn, err := c.Conn()
		if err != lastState {
			if err == false { // normal
				conn.Close()
				color = "good"
				status = "reachable"
			} else { // not normal
				time.Sleep(time.Duration(5) * time.Second)
				conn, errr := c.Conn()
				if errr != false {
					color = "danger"
					status = "unreachable"
				} else {
					conn.Close()
					goto begin
				}
			}
			lastState = err // key of success
			if *slack == true {
				am := alertMessage{
					channel: *channel,
					color:   color,
					text:    fmt.Sprintf("Destination host %s:%v %s\n", *host, *port, status),
					url:     *url,
				}
				am.SentSlack()
			}
		}
		time.Sleep(time.Duration(*interval) * time.Second)
	}
}
