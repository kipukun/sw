package main

import (
	"context"
	"embed"
	"encoding/json"
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"time"
)

type Application int

const (
	Plex Application = iota
	Steam
)

type swr struct {
	Application int
}

type workspaces []struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
	Rect struct {
		X      int `json:"x"`
		Y      int `json:"y"`
		Width  int `json:"width"`
		Height int `json:"height"`
	} `json:"rect"`
	Focus              []int       `json:"focus"`
	Border             string      `json:"border"`
	CurrentBorderWidth int         `json:"current_border_width"`
	Layout             string      `json:"layout"`
	Orientation        string      `json:"orientation"`
	Percent            interface{} `json:"percent"`
	WindowRect         struct {
		X      int `json:"x"`
		Y      int `json:"y"`
		Width  int `json:"width"`
		Height int `json:"height"`
	} `json:"window_rect"`
	DecoRect struct {
		X      int `json:"x"`
		Y      int `json:"y"`
		Width  int `json:"width"`
		Height int `json:"height"`
	} `json:"deco_rect"`
	Geometry struct {
		X      int `json:"x"`
		Y      int `json:"y"`
		Width  int `json:"width"`
		Height int `json:"height"`
	} `json:"geometry"`
	Window         interface{}   `json:"window"`
	Urgent         bool          `json:"urgent"`
	Marks          []interface{} `json:"marks"`
	FullscreenMode int           `json:"fullscreen_mode"`
	Nodes          []interface{} `json:"nodes"`
	FloatingNodes  []interface{} `json:"floating_nodes"`
	Sticky         bool          `json:"sticky"`
	Num            int           `json:"num"`
	Output         string        `json:"output"`
	Type           string        `json:"type"`
	Representation string        `json:"representation"`
	Focused        bool          `json:"focused"`
	Visible        bool          `json:"visible"`
}

//go:embed index.html
var f embed.FS

func main() {

	tmpl := template.Must(template.ParseFS(f, "index.html"))
	s := &http.Server{
		Addr:         ":8080",
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 5 * time.Second,
	}

	http.HandleFunc("/switch", func(rw http.ResponseWriter, r *http.Request) {
		dec := json.NewDecoder(r.Body)
		var sr swr
		err := dec.Decode(&sr)
		if err != nil {
			http.Error(rw, err.Error(), http.StatusInternalServerError)
			return
		}
		cmd := exec.Command("swaymsg", "workspace", fmt.Sprintf("%d", sr.Application))
		out, err := cmd.CombinedOutput()
		if err != nil {
			http.Error(rw, err.Error(), http.StatusInternalServerError)
			return
		}
		if len(out) != 0 {
			http.Error(rw, string(out), http.StatusInternalServerError)
			log.Println("some error from sway:", string(out))
			return
		}
		fmt.Fprint(rw, "success")

	})
	http.HandleFunc("/", func(rw http.ResponseWriter, r *http.Request) {
		bs, err := ioutil.ReadFile("out.json")
		if err != nil {
			http.Error(rw, err.Error(), http.StatusInternalServerError)
			return
		}
		var ws workspaces
		err = json.Unmarshal(bs, &ws)
		if err != nil {
			http.Error(rw, err.Error(), http.StatusInternalServerError)
			return
		}
		d := struct {
			PlexOn, SteamOn bool
		}{
			!ws[0].Focused, ws[0].Focused,
		}
		err = tmpl.Execute(rw, d)
		if err != nil {
			http.Error(rw, err.Error(), http.StatusInternalServerError)
			return
		}
	})
	ctx, cancel := signal.NotifyContext(context.TODO(), os.Interrupt)
	go func() {
		<-ctx.Done()
		s.Close()
	}()
	log.Println("listening on 8080")
	s.ListenAndServe()
	cancel()
}
