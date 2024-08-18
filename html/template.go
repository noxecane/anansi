package html

import (
	"bytes"
	jsonslow "encoding/json"
	"html/template"
	"net/http"

	"github.com/noxecane/anansi"
	"github.com/noxecane/anansi/json"
	"github.com/rs/zerolog"
)

type Template interface {
	Render(*http.Request, http.ResponseWriter, any) error
}

type htmlTemplate struct {
	name string
	tmpl *template.Template
}

func Parse(name string, source ...string) Template {
	var tmpl *template.Template
	if len(source) == 1 {
		tmpl = template.Must(template.New(name).ParseFiles(source...))
	} else if len(source) > 0 {
		tmpl = template.Must(template.ParseFiles(source...))
	} else {
		panic("souce can't be empty")
	}

	return &htmlTemplate{name, tmpl}
}

func (t *htmlTemplate) Render(r *http.Request, w http.ResponseWriter, data any) error {
	log := zerolog.Ctx(r.Context())
	raw, err := json.Marshal(data)

	if data != nil {
		log.UpdateContext(func(ctx zerolog.Context) zerolog.Context {
			buffer := new(bytes.Buffer)
			if err != nil {
				panic(err)
			}

			if err := jsonslow.Compact(buffer, raw); err != nil {
				panic(err)
			}

			return ctx.RawJSON("html_data", buffer.Bytes())
		})
	}

	err = t.tmpl.ExecuteTemplate(w, t.name, data)

	if err == nil {
		log.Info().
			Interface("response_headers", anansi.SimpleMap(w.Header())).
			Msg("")
	} else {
		log.Err(err).
			Interface("response_headers", anansi.SimpleMap(w.Header())).
			Msg("")
	}

	return err
}
