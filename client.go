package main

import (
	"bufio"
	"io"
	"log"
	"net"
	"os"
	"os/exec"
	"os/user"
	"strings"
	"time"
)

func login(conn net.Conn, botBinary string) {
	user, err := user.Current()
	if err != nil {
		log.Fatalln(err.Error())
	}

	username := user.Username
	token := "asd"
	login_packet := token + " " + username + "#" + botBinary
	log.Println("Logged in with:", login_packet)
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
	log.Println("Bot started")

	log.Println("Waiting for game to start...")

	for {
		dataFromServer, err := bufio.NewReader(conn).ReadString('\n')

		if strings.HasPrefix(dataFromServer, "CTRL") {
			// This is a special packet from server, not meant for the bot
			log.Println(dataFromServer)
			break
		}

		if err != nil {
			log.Println("Failed to read from server", err)
			break
		}

		log.Println("Got this from server:", dataFromServer)

		_, err = io.WriteString(stdin, dataFromServer)
		if err != nil {
			log.Println("Failed to send to bot")
			break
		}

		log.Println("Sent to bot")

		reader := bufio.NewReader(stdout)
		slurp, err := reader.ReadString('\n')
		if err != nil {
			log.Println("Failed to read from bot")
			break
		}

		log.Println("Got this from bot:", slurp)

		conn.Write([]byte(slurp))

		log.Println("Sent to server. Waiting for other bot to play...")
	}

	bot.Process.Kill()

	// log.Println("Bot finished:", bot.ProcessState.ExitCode())
}

func main() {
	server := "localhost:8081"
	botBinary := "./bot"

	for {
		conn, err := net.Dial("tcp", server)

		if err != nil {
			log.Fatalln("Error connecting:", err.Error())
		}

		log.Println("Connected to game server at", server)

		login(conn, botBinary)

		log.Println("-----------------------")
		play(conn, botBinary)
		log.Println("Game endeded. Starting a new one in 5 seconds")
		log.Println("-----------------------")

		time.Sleep(time.Second * 5)
	}

}
