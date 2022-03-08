package magic

import (
	"bytes"
	"net/http"
	"strings"
	"text/template"

	"github.com/grafana/agent/pkg/config"

	"gopkg.in/yaml.v2"

	configpage "github.com/grafana/agent/pkg/magic/pages/config"

	"github.com/grafana/agent/pkg/magic/pages/metrics"

	"github.com/grafana/agent/pkg/magic/pages"

	"github.com/gorilla/mux"
)

type Instance struct {
	storage   *Storage
	appender  *Appender
	ninjakeys []ninjaKey
	cfg       *config.Config
}

func NewInstance(cfg *config.Config) *Instance {

	storage := newStorage()
	i := &Instance{
		storage:  storage,
		appender: storage.a,
		cfg:      cfg,
	}
	return i
}

func (i *Instance) ApplyRoutes(r *mux.Router) {
	i.ninjakeys = i.createNinjaArray(r)
	r.HandleFunc("/magic/assets/ninja.js", i.serverJS(ninjajs)).Methods("GET")
	r.HandleFunc("/magic/assets/ninja-action.js", i.serverJS(ninjaaction)).Methods("GET")
	r.HandleFunc("/magic/assets/ninja-header.js", i.serverJS(ninjaheader)).Methods("GET")
	r.HandleFunc("/magic/assets/ninja-footer.js", i.serverJS(ninjafooter)).Methods("GET")
	r.HandleFunc("/magic/assets/base-styles.js", i.serverJS(basestyles)).Methods("GET")

	r.HandleFunc("/magic", i.index).Methods("GET")
}

func (i *Instance) ApplyConfigAndReload() error {
	return nil
}

func (i *Instance) MetricHistory() []Val {
	return i.appender.MetricValues()
}

func (i *Instance) Storage() *Storage {
	return i.storage
}

func (i *Instance) index(writer http.ResponseWriter, request *http.Request) {
	p := pages.Index{TitleText: "Index"}
	output := pages.PageTemplate(&p, i.makeNinja())
	writer.WriteHeader(http.StatusOK)
	writer.Write([]byte(output))
}

func (i *Instance) serverJS(file []byte) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		writer.Header().Set("Content-Type", "application/x-javascript")
		writer.WriteHeader(http.StatusOK)
		_, _ = writer.Write(file)
	}
}

func (i *Instance) makeNinja() string {
	tmp := `
const hotkeys = [
    {{range $val := .}}
                {
                        id: "{{ $val.ID }}",
                        title: "{{ $val.Title }}",
                        keywords: "{{ $val.Keywords }}",
                        handler: () => {                  
                            window.location.href = "{{ $val.URL }}";
                        }
                },
    {{end}}
        ];
`
	t, err := template.New("name").Parse(tmp)
	if err != nil {
		return ""
	}
	bb := bytes.Buffer{}
	err = t.Execute(&bb, i.ninjakeys)
	if err != nil {
		return ""
	}
	return bb.String()
}

func (i *Instance) createNinjaArray(r *mux.Router) []ninjaKey {
	var ninjaKeys = []ninjaKey{
		{
			ID:       "view-metrics",
			Title:    "View Metrics",
			Keywords: "view metrics",
			URL:      "/magic/metrics/view-metrics",
			Handler: func(writer http.ResponseWriter, request *http.Request) {
				mh := i.MetricHistory()
				val := make([]metrics.MetricItem, 0)
				for _, item := range mh {
					j := item
					if j.name == "" {
						break
					}
					val = append(val, &j)
				}
				i.writeResponse(writer, &metrics.View{
					TitleText: "View Metrics",
					Metrics:   val,
				})
			},
		},
		{
			ID:       "view-config",
			Title:    "View Config",
			Keywords: "view config configuration",
			URL:      "/magic/config/view-config",
			Handler: func(writer http.ResponseWriter, request *http.Request) {
				bb, _ := yaml.Marshal(i.cfg)
				v := string(bb)
				v = strings.ReplaceAll(v, "\n", "<br>")
				i.writeResponse(writer, &configpage.Config{
					TitleText: "View Config",
					Config:    v,
				})
			},
		},
		{
			ID:       "windows-exporter",
			Title:    "Enable / Disable Windows Exporter Integration",
			Keywords: "enable disable windows integration",
			URL:      "/magic/integrations/windows-exporter",
			Handler: func(writer http.ResponseWriter, request *http.Request) {

			},
		},
	}

	for _, nk := range ninjaKeys {
		r.HandleFunc(nk.URL, nk.Handler).Methods("GET")
	}
	return ninjaKeys
}

func (i *Instance) writeResponse(writer http.ResponseWriter, p pages.Page) {
	output := pages.PageTemplate(p, i.makeNinja())
	writer.WriteHeader(http.StatusOK)
	writer.Write([]byte(output))
}

type ninjaKey struct {
	ID       string
	Title    string
	Keywords string
	URL      string
	Handler  http.HandlerFunc
}
