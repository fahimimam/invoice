package api

import (
	"github.com/go-chi/chi/v5"
	"log"
	"net/http"
)

func invoiceRouter(ctrl *InvoiceController) http.Handler {
	h := chi.NewRouter()

	h.Group(func(r chi.Router) {
		// Set up routes
		r.Post("/create", ctrl.CreateInvoiceHandler)
		r.Post("/update", ctrl.UpdateInvoiceHandler)
		r.Post("/delete", ctrl.DeleteInvoiceHandler)
		r.Get("/get/{id}", ctrl.GetInvoiceHandler)
	})

	return h
}

func healthRouter(ctrl *SystemController) http.Handler {
	log.Println("healthRouter")
	h := chi.NewRouter()
	h.Group(func(r chi.Router) {
		// add all system check here, like: api, db connection, ......
		r.Get("/api", ctrl.apiCheck)
	})
	return h
}
