package main

import (
	"flag"
	"html/template"
	"log"
	"net/http"
	"path/filepath"
	"regexp"
)

var tmpl = template.Must(template.New("docsList.html").Parse(`
<!DOCTYPE html>
<html lang="en">
  <head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <meta http-equiv="X-UA-Compatible" content="ie=edge">
    <title>Insomnia API Documents</title>
  </head>
  <body>
    <h1>Insomnia API Documents</h1>
    <ol>
        {{ range .DataFiles }}
        <li><a target="_blank" href="./documenter/index.html#../{{ . }}">{{ . }}</a></li>
        {{ else }}
        <p>No documents for now.</p>
        {{ end }}
        </tbody>
    </ol>
  </body>
</html>
`))

type route struct {
	pattern *regexp.Regexp
	handler http.Handler
}

type RegexpHandler struct {
	routes []*route
}

func (h *RegexpHandler) Handler(pattern *regexp.Regexp, handler http.Handler) {
	h.routes = append(h.routes, &route{pattern, handler})
}

func (h *RegexpHandler) HandleFunc(pattern *regexp.Regexp, handler func(http.ResponseWriter, *http.Request)) {
	h.routes = append(h.routes, &route{pattern, http.HandlerFunc(handler)})
}

func (h *RegexpHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	for _, route := range h.routes {
		if route.pattern.MatchString(r.URL.Path) {
			route.handler.ServeHTTP(w, r)
			return
		}
	}
	// no pattern matched; send 404 response
	http.NotFound(w, r)
}

func handleDocsList(res http.ResponseWriter, req *http.Request) {
	files, err := filepath.Glob("./*.json")
	if err != nil {
		http.Error(res, "Error listing files: "+err.Error(), 500)
		return
	}

	res.Header().Set("Content-Type", "text/html; charset=utf-8")
	err = tmpl.Execute(res, struct {
		DataFiles []string
	}{
		DataFiles: files,
	})
	if err != nil {
		http.Error(res, "Error rendering template: "+err.Error(), 500)
		return
	}
}

func main() {
	port := flag.String("p", "8080", "port to serve on")
	flag.Parse()

	handler := &RegexpHandler{}

	docsListRegex, _ := regexp.Compile("^/$")
	handler.HandleFunc(docsListRegex, handleDocsList)

	restRegex, _ := regexp.Compile("^(/.+$|$)")
	handler.Handler(restRegex, http.FileServer(http.Dir(".")))

	log.Printf("Serving Insomnia API Docs Server on HTTP port: %s\n", *port)
	log.Fatal(http.ListenAndServe(":"+*port, handler))
}
