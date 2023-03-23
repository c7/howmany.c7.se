package main

//go:generate ./convert.sh

import (
	"encoding/json"
	"fmt"
	"html/template"
	"math/rand"
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

	fmt.Printf("Listening on http://0.0.0.0:%s\n", port)
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

func parseRequestPath(path string) (string, string, bool) {
	return strings.Cut(path[1:], "/")
}

func (app *App) Handler(w http.ResponseWriter, r *http.Request) {
	switch r.URL.Path {
	case "/":
		app.index(w, r)
	default:
		app.show(w, r)
	}
}

func (app *App) randomLocations() []string {
	var locations []string

	for l := range app.populations {
		locations = append(locations, l)
	}

	rand.Shuffle(len(locations), func(i, j int) {
		locations[i], locations[j] = locations[j], locations[i]
	})

	return locations
}

func (app *App) randomCompaniesWithMoreEmployeesThan(n int) []string {
	var companies []string

	for name, c := range app.companies {
		if c.Employees > n {
			name = strings.Title(name)

			if len(name) == len(c.Name) {
				name = c.Name
			}

			companies = append(companies, name)
		}
	}

	rand.Shuffle(len(companies), func(i, j int) {
		companies[i], companies[j] = companies[j], companies[i]
	})

	return companies
}

func (app *App) randomPath() string {
	location := app.randomLocations()[0]
	population := app.populations[location]

	companies := app.randomCompaniesWithMoreEmployeesThan(population)

	if len(companies) == 0 {
		return "/Gotland/Accenture"
	}

	return fmt.Sprintf("/%s/%s", strings.Title(location), strings.ReplaceAll(companies[0], " ", "-"))
}

func (app *App) index(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, app.randomPath(), http.StatusFound)
}

func (app *App) show(w http.ResponseWriter, r *http.Request) {
	pn, cn, found := parseRequestPath(r.URL.Path)
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

	v := Value{
		Company:    c,
		Population: p,
		Location:   strings.ReplaceAll(pn, "-", " "),
		Times:      float64(c.Employees) / float64(p),
	}

	app.template.Execute(w, v)
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
  	<style>
			:root {
  			--primary: #546e7a;
  			--primary-hover: #607d8b;
  			--primary-focus: rgba(84, 110, 122, 0.25);
  			--primary-inverse: #FFF;
				--typography-spacing-vertical: 0;
			}
		</style>
	</head>
  <body>
    <main>
			<article>
				<header>
					<form action="/" method="GET">
						<input type="submit" value="Randomize 🔀">
					</form>
				</header>
				<h1>
					The number of employees at 
					<mark>{{.Company.Name}}</mark> 
					equals the population of 
					<mark>{{.Location}}</mark>… 
					<u>{{printf "%.2f" .Times}} times</u>
					{{if gt .Times 1.0}}🤯{{else}}😐{{end}}
				</h1>
				<footer>
					<table role="grid">
						<tbody>
							<tr>
								<th scope="row" width="1">#️</th>
								<td>{{.Company.Rank}}</td>
							</tr>
							<tr>
								<th scope="row">📈</th>
								<td>{{.Company.Symbol}}</td>
							</tr>
							<tr>
								<th scope="row">👥</th>
								<td>{{.Company.Employees}}</td>
							</tr>
							<tr>
								<th scope="row">🌎</th>
								<td>{{.Company.Country}}</td>
							</tr>
						</tbody>
					</table>
				</footer>
			</article>
		</main>
  </body>
</html>
`
