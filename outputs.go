package main

import (
	"crypto/tls"
	"fmt"
	"github.com/stianeikeland/go-rpio"
	"github.com/nlopes/slack"
	"log"
	"net"
	"net/mail"
	"net/smtp"
	"os"
	"os/exec"
	"strconv"
	"time"
)

var (
	gpio1 chan (string)
	gpio2 chan (string)
	input chan (bool)
	siren chan (bool)
)

func init() {
	gpio1 = make(chan string)
	gpio2 = make(chan string)
	siren = make(chan bool)
	input = make(chan bool)

	if err := rpio.Open(); err != nil {
		log.Fatal(err)
	}
}

func speak() {
	longv := ""
	sermon := "espeak -g 5 \"" + conf.Speech.Message + ".\n"
	for _, v := range conf.Speech.Read {
		val := strconv.Itoa(arduino_linear_stat[v])
		if v == "H" {
			longv = "humidity"
		}
		if v == "T" {
			longv = "temperature"
		}
		sermon = sermon + longv + " is " + val + ".\n"
	}
	sermon = sermon + "\""
	log.Printf("%s\n", sermon)
	cmd := exec.Command("bash", "-c", sermon)
	err := cmd.Start()
	if err != nil {
		log.Printf("error for speaking")
	}
}

func input_presence() {
	for {
		presence := <-input
		if presence {
			if conf.Verbose {
				log.Printf("Input detected\n")
			}
			// do something useful
		}
	}
}

func test_siren() {
	if conf.Verbose {
		log.Printf("Testing siren...\n")
	}
	siren <- true
}

func siren_mgr() {
	if !conf.Alarms.Siren_enabled {
		return
	}
	if conf.Verbose {
		log.Printf("Siren manager on\n")
	}
	if conf.Outputs["alarm"].PIN == 0 {
		log.Fatal("Alarm configured but GPIO for the relay is not!")
		os.Exit(1)
	}
	var pin rpio.Pin
	// Open and map memory to access gpio, check for errors
	pin = rpio.Pin(conf.Outputs["alarm"].PIN)
	if err := rpio.Open(); err != nil {
		log.Fatal(err)
		os.Exit(1)
	}
	pin.Output()
	pin.High()
	//pin.Low()
	defer rpio.Close()
	for {
		listentome := false
		listentome = <-siren
		if listentome {
			if conf.Verbose {
				log.Printf("Siren ON!!\n")
			}
			pin.Low()
			//pin.High()
			time.Sleep(time.Second * 10)
			if conf.Verbose {
				log.Printf("Siren OFF!!\n")
			}
			pin.High()
			//pin.Low()
			time.Sleep(time.Second * 10)
		}
	}
}

func alarm_mgr() {
	time.Sleep(time.Minute)
	//set a x seconds ticker
	ticker := time.NewTicker(time.Duration(conf.Sensors.Poll_interval) * time.Second)
	// check if presence and U sensor are both setup (it has been just initialized)
	if conf.Alarms.Presence && arduino_linear_stat["U"]==0  {
		log.Fatal("Alarm configured but GPIO for the relay is not!")
		os.Exit(1)
	}

	for _ = range ticker.C {
		lock.Lock()
		actual_temp := arduino_linear_stat["T"]
		//human := arduino_linear_stat["U"]
		lock.Unlock()
		// temperature alarm
		if actual_temp < conf.Alarms.Critical_temp {
			log.Printf("Alarm triggered %d < %d!!\n", actual_temp, conf.Alarms.Critical_temp)
			if conf.Alarms.Email_enabled {
				s := send_email(strconv.Itoa(actual_temp))
				if s {
					log.Println("mail sent")
				}
			}
			if conf.Alarms.Siren_enabled {
				siren <- true
			}
		}
		// presence alarm
		if conf.Alarms.Presence {
			siren <- true
		}
		if conf.Alarms.Slack_token != "none" {
			slack_notify(actual_temp)
		}
	}
}

func command_socket(socket string) (reply string) {
	if socket == "on" {
		gpio1 <- "on"
		reply = "Turning ON"
	} else if socket == "off" {
		gpio1 <- "off"
		reply = "Turning OFF"
	} else {
		reply = "Specify 'on' or 'off'"
	}
	return reply
}

func send_gpio1(gpio1 <-chan string) {
	pin := rpio.Pin(conf.Outputs["socket1"].PIN)
	pin.Output()
	for {
		status := <-gpio1
		log.Printf("Sending %s to GPIO1", status)
		if status == "on" {
			pin.High()
		}
		if status == "off" {
			pin.Low()
		}
	}
}

func send_gpio2(gpio2 <-chan string) {
	pin := rpio.Pin(conf.Outputs["socket2"].PIN)
	pin.Output()
	for {
		status := <-gpio2
		log.Printf("Sending %s to GPIO2", status)
		if status == "on" {
			pin.High()
		}
		if status == "off" {
			pin.Low()
		}
	}
}

func send_email(temp string) (sent bool) {

	from := mail.Address{"", conf.Alarms.Mailbox}
	to := mail.Address{"", conf.Alarms.Mailbox}
	subj := "Greenhouse Temperature Alarm"
	body := "Detected Low temperature (" + temp + "C )\n\n"
	// Setup headers
	headers := make(map[string]string)
	headers["From"] = from.String()
	headers["To"] = to.String()
	headers["Subject"] = subj
	headers["X-Priority"] = "1"
	// Setup message
	message := ""
	for k, v := range headers {
		message += fmt.Sprintf("%s: %s\r\n", k, v)
	}
	message += "\r\n" + body
	// Connect to the SMTP Server
	servername := conf.Alarms.Smtp
	host, _, _ := net.SplitHostPort(servername)
	auth := smtp.PlainAuth("", conf.Alarms.Auth_user, conf.Alarms.Auth_pwd, host)
	// TLS config
	tlsconfig := &tls.Config{
		InsecureSkipVerify: true,
		ServerName:         host,
	}
	c, err := smtp.Dial(servername)
	if err != nil {
		log.Printf("%s", err)
		return false
	}
	c.StartTLS(tlsconfig)
	// Auth
	if err = c.Auth(auth); err != nil {
		log.Printf("%s", err)
		c.Quit()
		return false
	} else {
		// To && From
		if err = c.Mail(from.Address); err != nil {
			log.Printf("%s", err)
		}
		if err = c.Rcpt(to.Address); err != nil {
			log.Printf("%s", err)
		}
		// Data
		w, err := c.Data()
		if err != nil {
			log.Printf("%s", err)
		}
		_, err = w.Write([]byte(message))
		if err != nil {
			log.Printf("%s", err)
		}
		err = w.Close()
		if err != nil {
			log.Printf("%s", err)
		}
		c.Quit()
		return true
	}
}


func slack_notify( temp int) {
        api := slack.New(conf.Alarms.Slack_token)
        attachment := slack.Attachment{
                Pretext: "Alarm:",
                Text:    " Low temperature: " + strconv.Itoa(temp),
                // Uncomment the following part to send a field too
                /*
                        Fields: []slack.AttachmentField{
                                slack.AttachmentField{
                                        Title: "a",
                                        Value: "no",
                                },
                        },
                */
        }

        channelID, timestamp, err := api.PostMessage("greenhouse", slack.MsgOptionText("important", false), slack.MsgOptionAttachments(attachment))
        if err != nil {
                fmt.Printf("%s\n", err)
        } else {
       		fmt.Printf("Message successfully sent to channel %s at %s", channelID, timestamp)
	}
}
