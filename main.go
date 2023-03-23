package main

//go:generate ./convert.sh

import (
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"os"
	"strings"

	_ "embed"

	"github.com/TV4/nids"
)

//go:embed data/companies.json
var dataCompanies []byte

//go:embed data/populations.json
var dataPopulations []byte

const defaultPort = "8000"

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = defaultPort
	}

	app, err := NewApp(dataCompanies, dataPopulations)
	if err != nil {
		panic(err)
	}

	http.HandleFunc("/", app.Handler)

	fmt.Println("Listening on port", port)
	http.ListenAndServe(":"+port, nil)
}

func NewApp(dataCompanies, dataPopulations []byte) (*App, error) {
	var v []Company

	if err := json.Unmarshal(dataCompanies, &v); err != nil {
		return nil, err
	}

	c := Companies{}

	for _, company := range v {
		name, _, _ := strings.Cut(company.Name, " (")

		c[nids.Case(name)] = company
	}

	var p Populations

	if err := json.Unmarshal(dataPopulations, &p); err != nil {
		return nil, err
	}

	t, err := template.New("index").Parse(html)
	if err != nil {
		return nil, err
	}

	return &App{companies: c, populations: p, template: t}, nil
}

type App struct {
	companies   map[string]Company
	populations Populations
	template    *template.Template
}

func (app *App) Handler(w http.ResponseWriter, r *http.Request) {
	pn, cn, found := strings.Cut(r.URL.Path[1:], "/")
	if !found {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	c, ok := app.companies[strings.ToLower(cn)]
	if !ok {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	p, ok := app.populations[strings.ToLower(pn)]
	if !ok {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	f := float64(c.Employees) / float64(p)

	// fmt.Fprintf(w, format, c.Name, pn, f)

	app.template.Execute(w, Value{
		Company:    c,
		Population: p,
		Location:   pn,
		Times:      f,
	})
}

type Value struct {
	Company    Company
	Population int
	Location   string
	Times      float64
}

type Company struct {
	Rank      int
	Name      string
	Symbol    string
	Employees int
	Price     float64
	Country   string
}

type Companies map[string]Company
type Populations map[string]int

var html = `<!doctype html>
<html lang="en" data-theme="dark">
  <head>
    <meta charset="utf-8">
    <meta name="viewport" content="width=device-width, initial-scale=1">
		<link rel="stylesheet" href="https://unpkg.com/@picocss/pico@1.*/css/pico.classless.min.css">
		<title>How many?</title>
  </head>
  <body>
    <main>
			<article>
				<h1>
					The number of employees at 
					<mark>{{.Company.Name}}</mark> 
					equals the population of 
					<mark>{{.Location}}</mark>… 
					<u>{{printf "%.2f" .Times}} times</u>
					{{if gt .Times 1.0}}🤯{{else}}😐{{end}}
				</h1>
			</article>
		</main>
  </body>
</html>
`
