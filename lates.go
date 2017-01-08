package main

import (
	"html/template"
	"log"
	"net/http"
	"sync"
	"time"
)

var (
	indexTemp *template.Template
	store     LatesStore
)

func init() {
	indexTemp = template.Must(template.ParseFiles("index.html"))
	store = LatesStore{cache: make(map[string]Late), lastReq: time.Now()}
}

func main() {
	http.HandleFunc("/", renderLates)
	http.HandleFunc("/submit", handleSubmit)

	log.Println("listening on port 8080...")
	http.ListenAndServe(":8080", nil)
}

func renderLates(w http.ResponseWriter, r *http.Request) {
	payload := struct {
		Lates []Late
		Now   time.Time
	}{
		Lates: store.List(),
		Now:   time.Now(),
	}
	indexTemp.Execute(w, payload)
}

func handleSubmit(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		err := r.ParseForm()
		if err != nil {
			log.Printf("handleSubmit: got err when trying to parse form, err: %s", err)
		} else {
			name := r.PostFormValue("name")
			veg := r.PostFormValue("vegetarian") == "on"
			fridge := r.PostFormValue("refrigerated") == "on"
			store.Add(Late{Name: name, Vegetarian: veg, Refrigerated: fridge})
		}
	}
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

// Late represents an individual late request.
type Late struct {
	// The requester's name.
	Name string

	// Vegeterian represents whether or not the requester wants to have only veggie options.
	Vegetarian bool

	// Refrigerated represents whether or not the requester wants to have their late refridgerated.
	Refrigerated bool
}

// LatesStore is an in-memory map that keeps track of all the lates.
type LatesStore struct {
	// Muxtex controls concurrent access to the cache and last request time.
	sync.Mutex

	// Cache is the underlying map from requester names to their Late request.
	cache map[string]Late

	// lastReq is the last time any member of LatesStore was accessed.
	lastReq time.Time
}

func (ls *LatesStore) list() []Late {
	defer ls.checkTime()
	var out []Late
	for _, late := range ls.cache {
		out = append(out, late)
	}
	return out
}

// List returns all of the Lates currently inside of the LatesStore.
func (ls *LatesStore) List() []Late {
	ls.Lock()
	defer ls.Unlock()
	return ls.list()
}

func (ls *LatesStore) add(l Late) {
	defer ls.checkTime()
	ls.cache[l.Name] = l
}

// Add adds a new late to the LatesStore.
func (ls *LatesStore) Add(l Late) {
	ls.Lock()
	defer ls.Unlock()
	ls.add(l)
}

// checkTime expires old lates the date has changed in between
// now and the last time LatesStore was accessed.
func (ls *LatesStore) checkTime() {
	oldYear, oldMonth, oldDay := ls.lastReq.Date()
	newYear, newMonth, newDay := time.Now().Date()

	// Time usually moves forward, so this should trigger when a new day has started.
	if oldYear != newYear || oldMonth != newMonth || oldDay != newDay {
		log.Printf("Dumping logs from date: %s \n", ls.lastReq.Format("01.02.06"))
		for name, late := range ls.cache {
			log.Printf("%s: %+v\n", name, late)
		}
		ls.cache = make(map[string]Late)
	}
	ls.lastReq = time.Now()
}
