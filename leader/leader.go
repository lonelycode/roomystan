package leader

import (
	"github.com/upsight/dinghy"
	"log"
	"net/http"
)

type Leader struct {
	OnLeader   func() error
	OnFollower func() error
}

func (l *Leader) Start(nodes []string, listen string) {
	nodes = append(nodes, listen)

	din, err := dinghy.New(
		listen,
		nodes,
		l.OnLeader,
		l.OnFollower,
		&dinghy.DiscardLogger{},
		dinghy.DefaultElectionTickRange,
		dinghy.DefaultHeartbeatTickRange,
	)

	if err != nil {
		log.Fatal(err)
	}

	for _, route := range din.Routes() {
		http.HandleFunc(route.Path, route.Handler)
	}

	go func() {
		if err := din.Start(); err != nil {
			log.Fatal(err)
		}
	}()

	log.Fatal(http.ListenAndServe(listen, nil))
}
