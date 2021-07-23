package main

import (
	"bufio"
	"crypto/md5"
	"encoding/base64"
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"strings"
	"time"

	"github.com/fatih/color"
)

func login(conn net.Conn, botBinaryPath string) {
	user, err := user.Current()
	if err != nil {
		log.Fatalln(err.Error())
	}

	// Calculate md5 hash of the binary
	file, err := os.Open(botBinaryPath)
	if err != nil {
		fmt.Println(color.RedString("Failed to read your bot"))
		os.Exit(3)
	}
	fileReader := bufio.NewReader(file)

	hash := md5.New()
	if _, err := io.Copy(hash, fileReader); err != nil {
		log.Fatal(err)
	}
	sum := hash.Sum(nil)
	base64hash := base64.StdEncoding.EncodeToString(sum)

	botName := filepath.Base(botBinaryPath)

	username := user.Username
	login_packet := username + "#" + botName + "#" + base64hash[:5]

	log.Println("Logged in with:", color.YellowString(login_packet))
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

	serverReader := bufio.NewReader(conn)
	for {
		dataFromServer, err := serverReader.ReadString('\n')

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
			log.Println(color.RedString("Failed to read from bot. Have you flushed with 'cout << endl'?"))
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

		play(conn, botBinary)

		if !loop {
			break
		}

		log.Println("Starting a new game in 5 seconds")
		time.Sleep(time.Second * 5)
	}

}
