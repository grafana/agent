package actorsystem

import (
	"github.com/AsynkronIT/protoactor-go/actor"
	"github.com/gorilla/mux"
	"github.com/grafana/agent/pkg/agentflow/config"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"log"
	"net/http"
)

// StartActorSystem creates the actor framework, instantiates the orchestrator and sends all the init and start messages
// to the nodes. It runs until stopped.
func StartActorSystem(filePath string) {
	as := actor.NewActorSystem()
	root := actor.NewRootContext(as, nil)
	cfgStr, err := ioutil.ReadFile(filePath)
	if err != nil {
		panic(err)
	}
	cfg := &config.Config{}
	err = yaml.Unmarshal(cfgStr, cfg)
	if err != nil {
		panic(err)
	}
	orch := NewOrchestrator(*cfg)
	err = orch.StartActorSystem(as, root)
	if err != nil {
		panic(err)
	}
	router := mux.NewRouter()
	router.HandleFunc("/mermaid", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(orch.GenerateMermaid()))
	})

	router.HandleFunc("/nodes", func(w http.ResponseWriter, r *http.Request) {
		bb, _ := yaml.Marshal(orch.NodeList())
		w.Write(bb)
	})

	router.HandleFunc("/nodes/{name}", func(writer http.ResponseWriter, request *http.Request) {
		vars := mux.Vars(request)
		key := vars["name"]
		writer.Write(orch.GetNodeStatus(key))

	})
	log.Fatal(http.ListenAndServe(":12345", router))
}
