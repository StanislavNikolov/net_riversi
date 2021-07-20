package main

import (
	"bufio"
	"flag"
	"io"
	"log"
	"net"
	"os"
	"os/exec"
	"os/user"
	"strings"
	"time"

	"github.com/fatih/color"
)

func login(conn net.Conn, botBinary string) {
	user, err := user.Current()
	if err != nil {
		log.Fatalln(err.Error())
	}

	username := user.Username
	token := "asd"
	login_packet := token + " " + username + "#" + botBinary
	log.Println("Logged in with:", username)
	conn.Write([]byte(login_packet + "\n"))
}

func play(conn net.Conn, botBinary string) {
	bot := exec.Command(botBinary)

	stdin, err := bot.StdinPipe()
	if err != nil {
		log.Fatal(err)
	}

	stdout, err := bot.StdoutPipe()
	if err != nil {
		log.Fatal(err)
	}

	bot.Stderr = os.Stderr // let the bot print to our stderr

	err = bot.Start()
	if err != nil {
		log.Fatal(err)
	}

	log.Println("Waiting for game to start...")

	for {
		dataFromServer, err := bufio.NewReader(conn).ReadString('\n')

		if err != nil {
			log.Println(color.RedString("Failed to talk with server. Maybe you lost connection?"), err)
			break
		}

		// The following are special packets not meant for the bot
		if strings.HasPrefix(dataFromServer, "CTRL") {
			log.Println(color.BlueString(dataFromServer))
			continue
		}

		if strings.HasPrefix(dataFromServer, "EXIT") {
			break
		}

		log.Printf("server -> bot:%s", color.YellowString(dataFromServer))

		_, err = io.WriteString(stdin, dataFromServer)
		if err != nil {
			log.Println(color.RedString("Failed to talk with your bot. Maybe it crashed?"))
			break
		}

		reader := bufio.NewReader(stdout)
		slurp, err := reader.ReadString('\n')
		if err != nil {
			log.Println(color.RedString("Failed to read from bot"))
			break
		}

		log.Printf("server <- bot:%s", color.YellowString(slurp))

		conn.Write([]byte(slurp))

		log.Println("Sent to server. Waiting for other bot to play...")
	}

	bot.Process.Kill()

	// log.Println("Bot finished:", bot.ProcessState.ExitCode())
}

func main() {
	var server string
	var botBinary string
	var loop bool

	flag.StringVar(&server, "server", "localhost:8081", "Address of the server to connect to. For example: 192.168.0.123:8081")
	flag.StringVar(&botBinary, "bot", "./bot", "Location of the executable binary to play with")
	flag.BoolVar(&loop, "loop", false, "Automatically start a new game after the last one ended")
	flag.Parse()

	for {
		conn, err := net.Dial("tcp", server)

		if err != nil {
			log.Fatalln("Error connecting:", err.Error())
		}

		login(conn, botBinary)

		log.Println("-----------------------")
		play(conn, botBinary)

		if !loop {
			break
		}

		log.Println("Starting a new game in 5 seconds")
		time.Sleep(time.Second * 5)
	}

}
