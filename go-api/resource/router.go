package resource

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi"
)

// Handler returns RESTful routes for managing Resources
func Handler() http.Handler {
	r := chi.NewRouter()
	r.Get(`/`, indexHandlerFunc)
	r.Get(`/{id}`, getHandlerFunc)
	r.Post(`/`, postHandlerFunc)
	return r
}

// indexHanlderFuc responds with a list of Resource IDs
func indexHandlerFunc(w http.ResponseWriter, r *http.Request) {
	items, err := List(50, 0)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	bytes, err := json.Marshal(items)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Write(bytes)
}

// getHandlerFunc responds with a Resource
func getHandlerFunc(w http.ResponseWriter, r *http.Request) {
	var resource *Resource
	var err error
	if id := chi.URLParam(r, "id"); id != "" {
		resource, err = Get(id)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		if resource.ID == `` {
			http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
			return
		}
	} else {
		http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
		return
	}
	bytes, _ := json.Marshal(*resource)
	w.Write(bytes)
}

// postHandlerFung creates a resource
func postHandlerFunc(w http.ResponseWriter, r *http.Request) {
	resource := Resource{}
	err := json.NewDecoder(r.Body).Decode(&resource)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	err = resource.Create()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	bytes, _ := json.Marshal(map[string]string{
		`id`: resource.ID,
	})
	w.Write(bytes)
}
