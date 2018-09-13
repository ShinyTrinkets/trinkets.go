package parser

import (
	"encoding/json"
	"regexp"
	"strings"
	"text/template"
)

// ParseBlocks extracts all code blocks from text
// The result will be in the form: {language => content}
func ParseBlocks(body string) map[string]string {
	// TODO: needs extra processing steps,
	// eg: for Javascript, might want to use Babel + Prettify

	langsList := []string{}
	for lng := range CodeBlocks {
		langsList = append(langsList, lng)
	}
	langs := strings.Join(langsList, "|")

	reBlk := regexp.MustCompile("(?sU)```[\t ]?(" + langs + ")[\t ]?[\n\r]+.+[\n\r]```[\n\r]?")
	reLng := regexp.MustCompile("^.+")

	blocks := map[string]string{}

	for _, v := range reBlk.FindAllString(body, -1) {
		s := strings.Trim(v, blankRunes)
		s = strings.Trim(s, "`")
		s = strings.Trim(s, blankRunes)
		lang := reLng.FindString(s)
		if lang == "" {
			continue
		}
		s = strings.Trim(s[len(lang):], blankRunes)
		// The first block of this type
		if blocks[lang] == "" {
			blocks[lang] = s
		} else {
			blocks[lang] += "\n\n" + s
		}
	}

	return blocks
}

// The "this file was generated by Trinkets-Spinal" warning message
func codeGeneratedByMsg(lang string) (str string) {
	cmt := CodeBlocks[lang].Comment
	if lang == "py" {
		str += "#!/usr/bin/env python3\n"
	}
	str += cmt + " THIS FILE IS GENERATED by Trinkets-Spinal\n"
	str += cmt + " Don't edit; Your changes will be overwritten!"
	return
}

// codeLangHeader creates a hash-map called `spinal` = {id, db, log, etc}
// for each language.
func codeLangHeader(front FrontMatter, lang string) (str string) {
	body, _ := json.MarshalIndent(front, "", "  ")
	if lang == "js" {
		str = "const spinal = " + string(body)
	} else if lang == "py" {
		str = "true = True; false = False; null = None\n"
		str += "spinal = " + string(body)
	}
	return
}

// codeLangImports creates the DB and LOG imports for each language.
func codeLangImports(front FrontMatter, lang string) (str string) {
	if lang == "js" {
		str = "let fse = require('fs-extra')\n"
		if front.Db {
			str += ("\n" + dbCode(front, lang) + "\n")
		}
		if front.Log {
			str += ("\n" + logCode(front, lang) + "\n")
		}
		str += "const trigger = require('trinkets/triggers');\n"
	} else if lang == "py" {
		str = "import functools\n"
		str += "print = functools.partial(print, flush=True)\n"
	}
	return
}

// The code for database
func dbCode(front FrontMatter, lang string) (str string) {
	if lang == "js" {
		str = `let FileSync = require('lowdb/adapters/FileSync')
fse.ensureDirSync('dbs/')
const db = require('lowdb')(new FileSync('dbs/{{.Id}}.json'));`
	}
	return renderTemplate(str, front)
}

// The code for logs
func logCode(front FrontMatter, lang string) (str string) {
	if lang == "js" {
		str = `let pinoStream = require('pino-multi-stream')
fse.ensureDirSync('logs/')
const log = pinoStream({
  base: null,
  streams: [{ stream: require('fs').createWriteStream('logs/{{.Id}}.log') }]
});`
	}
	return renderTemplate(str, front)
}

// Helper function to renter a template from a string,
// using the FrontMatter struct.
func renderTemplate(str string, front FrontMatter) string {
	tmpl, err := template.New(front.Id).Parse(str)
	if err != nil {
		panic(err)
	}
	builder := &strings.Builder{}
	err = tmpl.Execute(builder, front)
	if err != nil {
		panic(err)
	}
	return builder.String()
}
