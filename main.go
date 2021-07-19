package main

import (
    _ "embed"
	"log"
	"net/http"
	"os/exec"
	"regexp"
	"text/template"
)

const title string = "ConkyWeb"
const port string = ":5500"
const openDefaultBrowser = "xdg-open http://127.0.0.1" + port

//go:embed template.html
var htmlTemplate string

type pairCommand struct {
	designation string
	execution   string
}

var commandsList map[string]string = map[string]string{
	"uptime":   "uptime",
	"user":     "whoami",
	"ips":      "ip a",
	"hostname": "hostname",
	"packages": "if [ -e /usr/bin/dnf ] ; then dnf list --installed | wc -l ; else apt list --installed | wc -l ; fi",
	"kernel":   "uname -a",
	"os":       "lsb_release -a",
	"top":      "ps aux | sort -nk +4 | tail",
	"memory":   "free -h",
	"loads":    "uptime | cut -d' ' -f10-",
	"cpu":      "lscpu | grep 'Model name' | cut -d' ' -f12-",
	"disks":    "df -h",
}

func main() {
	http.HandleFunc("/", serveTemplate)
	log.Print("Le serveur est en ligne, visitez http://127.0.0.1", port)
	exec.Command("bash", "-c", openDefaultBrowser).Start()
	http.ListenAndServe(port, nil)
}

func serveTemplate(response http.ResponseWriter, request *http.Request) {
	log.Println(request.URL, request.UserAgent())

	// lance toutes les commandes dans des go routines
	var returnExecutedCommand = make(chan pairCommand)
	for designation, execution := range commandsList {
		command := pairCommand{designation, execution}
		go runCommand(command, returnExecutedCommand)
	}

	// recupération des résultats
	var commandsProcessed = make(map[string]string) // obliger de faire un make sinon le map reste nil
	for i := 0; i < len(commandsList); i++ {
		getExecutedCommand := <-returnExecutedCommand
		commandsProcessed[getExecutedCommand.designation] = getExecutedCommand.execution
	}

	indexTemplate, err := template.New("").Parse(htmlTemplate)
	if err != nil {
		panic(err)
	}

	data := struct {
		Title    string
		Commands map[string]string
	}{
		Title:    title,
		Commands: commandsProcessed,
	}
	indexTemplate.Execute(response, data)
}

func runCommand(commands pairCommand, returnExecutedCommand chan pairCommand) {
	output, errcmd := exec.Command("bash", "-c", commands.execution).CombinedOutput()
	if errcmd != nil {
		emptyResult := pairCommand{commands.designation, "La commande " + commands.execution + " n'existe pas"}
		log.Print(emptyResult)
		returnExecutedCommand <- emptyResult
	} else {
		regexp, _ := regexp.Compile(`\n`)
		if commands.designation == "user" {
			htmlFormated := regexp.ReplaceAllString(string(output), "")
			commands.execution = htmlFormated
		} else {
			htmlFormated := regexp.ReplaceAllString(string(output), "<br>")
			commands.execution = htmlFormated
		}
		returnExecutedCommand <- commands
	}
}
