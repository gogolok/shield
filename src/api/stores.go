package api

import (
	"db"

	"github.com/pborman/uuid"

	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
)

type StoreAPI struct {
	Data *db.ORM
}

func (self StoreAPI) ServeHTTP(w http.ResponseWriter, req *http.Request) {

	switch {
	case match(req, `GET /v1/stores`):
		_a, _b := unusedParam(req)
		_c, _d := pluginParam(req)
		stores, err := self.Data.GetAllAnnotatedStores(_a, _b, _c, _d)
		if err != nil {
			bail(w, err)
			return
		}

		JSON(w, stores)
		return

	case match(req, `POST /v1/stores`):
		if req.Body == nil {
			w.WriteHeader(400)
			return
		}

		var params struct {
			Name     string `json:"name"`
			Summary  string `json:"summary"`
			Plugin   string `json:"plugin"`
			Endpoint string `json:"endpoint"`
		}
		json.NewDecoder(req.Body).Decode(&params)

		if params.Name == "" || params.Plugin == "" || params.Endpoint == "" {
			w.WriteHeader(400)
			return
		}

		id, err := self.Data.CreateStore(params.Plugin, params.Endpoint)
		if err != nil {
			bail(w, err)
			return
		}

		_ = self.Data.AnnotateStore(id, params.Name, params.Summary)
		JSONLiteral(w, fmt.Sprintf(`{"ok":"created","uuid":"%s"}`, id.String()))
		return

	case match(req, `PUT /v1/store/[a-fA-F0-9-]+`):
		if req.Body == nil {
			w.WriteHeader(400)
			return
		}

		var params struct {
			Name     string `json:"name"`
			Summary  string `json:"summary"`
			Plugin   string `json:"plugin"`
			Endpoint string `json:"endpoint"`
		}
		json.NewDecoder(req.Body).Decode(&params)

		if params.Name == "" || params.Summary == "" || params.Plugin == "" || params.Endpoint == "" {
			w.WriteHeader(400)
			return
		}

		re := regexp.MustCompile("^/v1/store/")
		id := uuid.Parse(re.ReplaceAllString(req.URL.Path, ""))
		if err := self.Data.UpdateStore(id, params.Plugin, params.Endpoint); err != nil {
			bail(w, err)
			return
		}
		_ = self.Data.AnnotateStore(id, params.Name, params.Summary)

		JSONLiteral(w, fmt.Sprintf(`{"ok":"updated","uuid":"%s"}`, id.String()))
		return

	case match(req, `DELETE /v1/store/[a-fA-F0-9-]+`):
		re := regexp.MustCompile("^/v1/store/")
		id := uuid.Parse(re.ReplaceAllString(req.URL.Path, ""))
		deleted, err := self.Data.DeleteStore(id)

		if err != nil {
			bail(w, err)
		}
		if !deleted {
			w.WriteHeader(403)
			return
		}

		w.WriteHeader(200)
		return
	}

	w.WriteHeader(415)
	return
}
