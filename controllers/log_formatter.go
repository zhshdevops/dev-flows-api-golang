package controllers

import "strings"

func FormatLog(data string, isHtml bool) (buildLogs string) {
	var log []string
	if data != "" {

		dataLine := strings.Split(data, "\n")
		for _, d := range dataLine {
			log = strings.Split(d, " ")
			buildLogs = `<font color="#ffc20e">['` + log[0] + `']</font> ` + log[1] + `\n`
		}

	}

	return
}

func formatDate() {

}
