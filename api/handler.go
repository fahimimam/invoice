package api

import (
	"github.com/go-chi/chi/v5"
	"net/http"
)

func invoiceRouter(ctrl *InvoiceController) http.Handler {
	h := chi.NewRouter()

	h.Group(func(r chi.Router) {
		// Set up routes
		r.Post("/create", ctrl.CreateInvoice)
		r.Post("/update", ctrl.UpdateInvoice)
		r.Post("/delete", ctrl.DeleteInvoice)
		r.Get("/get/{id}", ctrl.GetInvoice)
	})

	return h
}
