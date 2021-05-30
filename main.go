package main

import (
	"html/template"
	"log"
	"net/http"
	"os/exec"
	"regexp"
)

const title string = "ConkyWeb"

var commands map[string]string = map[string]string{
	"uptime":   "uptime",
	"user":     "whoami",
	"ips":      "ip a",
	"hostname": "hostname",
	"packages": "dnf list --installed | wc -l", // TODO prévoir pour debian distro
	"kernel":   "uname -a",
	// "os":       "lsb_release -a",
	// TODO ignorer les commandes inconnues - mais mettre une alerte
	"top":    "ps aux | sort -nk +4 | tail",
	"memory": "free -h",
	"loads":  "uptime | cut -d' ' -f10-",
	"cpu":    "lscpu | grep 'Model name' | cut -d' ' -f12-",
	"disks":  "df -h",
}

func main() {
	http.HandleFunc("/", serveTemplate)
	log.Println("Le serveur est en ligne, visitez http://127.0.0.1:5500")
	// TODO ouverture auto du navigateur par défaut ? `xdg-open http://127.0.0.1:5500`
	http.ListenAndServe(":5500", nil)
}

func serveTemplate(res http.ResponseWriter, req *http.Request) {
	log.Println(req.URL, req.UserAgent())

	var commandsExec = make(map[string]string) // obliger de faire un make sinon le map reste nil
	for title, command := range commands {
		commandsExec[title] = runCommand(command)
	}

	indextemplate, err := template.ParseFiles("indextemplate.html")
	// TODO include html file to only have one file
	if err != nil {
		panic(err)
	}

	data := struct {
		Title    string
		Commands map[string]string
	}{
		Title:    title,
		Commands: commandsExec,
	}
	indextemplate.Execute(res, data)
}

func runCommand(command string) string {
	output, errcmd := exec.Command("bash", "-c", command).CombinedOutput()
	if errcmd != nil {
		log.Fatal("La commande ", command, " n'existe pas")
	}
	regexp, _ := regexp.Compile(`\n`)
	formatoutput := regexp.ReplaceAllString(string(output), "<br>")

	result := string(formatoutput)
	return result
}
