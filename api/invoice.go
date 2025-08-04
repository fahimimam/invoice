package api

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/fahimimam/invoice/api/response"
	"github.com/fahimimam/invoice/internal/service"
	"github.com/fahimimam/invoice/logger"
	"github.com/fahimimam/invoice/model"
	"github.com/fahimimam/invoice/utils"
	"github.com/go-chi/chi/v5"
	"net/http"
)

// InvoiceController ...
type InvoiceController struct {
	svc service.InvoiceServiceInterface
	lgr logger.StructLogger
}

// NewInvoiceController ...
func NewInvoiceController(svc service.InvoiceServiceInterface, lgr logger.StructLogger) *InvoiceController {
	return &InvoiceController{
		svc: svc,
		lgr: lgr,
	}
}

// SetLogger ...
func (ic *InvoiceController) SetLogger(lgr logger.StructLogger) {
	ic.lgr = lgr
}

func (ic *InvoiceController) CreateInvoice(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	tid := utils.GetTracingID(ctx)
	ic.lgr.Println("CreateInvoice", tid, "initialize")
	// ##################################################################################################################################################################

	// Extract identity headers
	certPEM := r.Header.Get("X-User-Cert")
	keyPEM := r.Header.Get("X-User-Key")
	mspID := r.Header.Get("X-User-MSPID")

	if certPEM == "" || keyPEM == "" || mspID == "" {
		_ = response.ServeJSON(w, http.StatusBadRequest, fmt.Sprintf("Mandatory headers missing"), nil)
		return
	}

	if err := ic.svc.SetInvoiceIdentityGateway(certPEM, keyPEM, mspID); err != nil {
		_ = response.ServeJSON(w, http.StatusBadRequest, fmt.Sprintf("SetInvoiceIdentityGateway"), map[string]interface{}{
			"status":  http.StatusBadRequest,
			"message": fmt.Sprintf("invalid Identity but required: %v", err),
		})
		return
	}

	defer func() {
		if err := ic.svc.CloseInvoiceIdentityGateway(); err != nil {
			_ = response.ServeJSON(w, http.StatusBadRequest, fmt.Sprintf("CloseInvoiceIdentityGateway"), map[string]interface{}{
				"status":  http.StatusBadRequest,
				"message": fmt.Sprintf("failed to close network gateway connection: %v", err),
			})
		}
	}()

	// Get ID from URL
	id := chi.URLParam(r, "id")
	if utils.IsValueEmpty(id) {
		_ = response.ServeJSON(w, http.StatusBadRequest, fmt.Sprintf("Mandatory url param missing"), map[string]interface{}{
			"status":  http.StatusBadRequest,
			"message": "Identity ID is required",
		})
		return
	}

	// Parse request body
	var invoiceInfo *model.InvoiceInfo
	if err := json.NewDecoder(r.Body).Decode(&invoiceInfo); err != nil {
		_ = response.ServeJSON(w, http.StatusBadRequest, fmt.Sprintf("Invalid request body"), map[string]interface{}{
			"status":  http.StatusBadRequest,
			"message": "Invalid request body: " + err.Error(),
		})
		return
	}

	requiredFields := map[string]interface{}{
		"invoice_date": invoiceInfo.InvoiceDate,
		"user_id":      invoiceInfo.UserId,
		"subtotal":     invoiceInfo.Subtotal,
		//TODO: Add More if needed
	}

	for field, value := range requiredFields {
		if utils.IsValueEmpty(value) {
			_ = response.ServeJSON(w, http.StatusBadRequest, fmt.Sprintf("Required field missing"), map[string]interface{}{
				"status":  http.StatusBadRequest,
				"message": fmt.Sprintf("Field %s is required", field),
			})
			return
		}
	}

	if err := ic.svc.CreateInvoice(ctx, invoiceInfo); err != nil {
		_ = response.ServeJSON(w, http.StatusBadRequest, fmt.Sprintf("failed to create invoice"), map[string]interface{}{
			"status":  http.StatusBadRequest,
			"message": err.Error(),
		})
		return
	}

	err := response.ServeJSON(w, http.StatusOK, "Invoice created successfully", invoiceInfo)
	if err == nil {
		ic.lgr.Println("CreateInvoice", tid, "completed!")
	}
}

func (ic *InvoiceController) GetInvoice(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	tid := utils.GetTracingID(ctx)
	ic.lgr.Println("GetInvoice", tid, "initialize")

	// Extract identity headers
	certPEM := r.Header.Get("X-User-Cert")
	keyPEM := r.Header.Get("X-User-Key")
	mspID := r.Header.Get("X-User-MSPID")

	if certPEM == "" || keyPEM == "" || mspID == "" {
		_ = response.ServeJSON(w, http.StatusBadRequest, fmt.Sprintf("Mandatory headers missing"), nil)
		return
	}

	if err := ic.svc.SetInvoiceIdentityGateway(certPEM, keyPEM, mspID); err != nil {
		_ = response.ServeJSON(w, http.StatusBadRequest, fmt.Sprintf("SetInvoiceIdentityGateway"), map[string]interface{}{
			"status":  http.StatusBadRequest,
			"message": fmt.Sprintf("invalid Identity but required: %v", err),
		})
		return
	}

	defer func() {
		if err := ic.svc.CloseInvoiceIdentityGateway(); err != nil {
			_ = response.ServeJSON(w, http.StatusBadRequest, fmt.Sprintf("CloseInvoiceIdentityGateway"), map[string]interface{}{
				"status":  http.StatusBadRequest,
				"message": fmt.Sprintf("failed to close network gateway connection: %v", err),
			})
		}
	}()

	// Parse request
	var request struct {
		ID string `json:"id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		_ = response.ServeJSON(w, http.StatusBadRequest, fmt.Sprintf("Mandatory headers missing"), map[string]interface{}{
			"status":  http.StatusBadRequest,
			"message": "Invalid request body",
		})
		return
	}
	if utils.IsValueEmpty(request.ID) {
		_ = response.ServeJSON(w, http.StatusBadRequest, fmt.Sprintf("Mandatory headers missing"), map[string]interface{}{
			"status":  http.StatusBadRequest,
			"message": "Identity ID is required",
		})
		return
	}

	invoice, err := ic.svc.GetInvoiceById(ctx, request.ID)
	if err != nil {
		_ = response.ServeJSON(w, http.StatusBadRequest, fmt.Sprintf("Invoice not found"), map[string]interface{}{
			"status":  http.StatusBadRequest,
			"message": fmt.Sprintf("Invoice not found %v", err.Error()),
		})
	}

	// Generate PDF using UniDoc
	pdfBytes, err := ic.svc.GeneratePDFInvoice(invoice)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to generate PDF: %v", err), http.StatusInternalServerError)
		return
	}

	// Serve PDF as HTTP Response
	w.Header().Set("Content-Type", "application/pdf")
	w.Header().Set("Content-Disposition", "attachment; filename=invoice_"+request.ID+".pdf")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(pdfBytes)

	ic.lgr.Println("GetInvoice", tid, "completed!")
	return
	//err = response.ServeJSON(w, http.StatusOK, utils.SuccessMessage, invoice)
	//if err == nil {
	//	ic.lgr.Println("GetInvoice", tid, "completed!")
	//}
}

func (ic *InvoiceController) UpdateInvoice(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	tid := utils.GetTracingID(ctx)
	ic.lgr.Println("UpdateInvoice", tid, "initialize")
	// ##################################################################################################################################################################

	// Extract identity headers
	certPEM := r.Header.Get("X-User-Cert")
	keyPEM := r.Header.Get("X-User-Key")
	mspID := r.Header.Get("X-User-MSPID")

	if certPEM == "" || keyPEM == "" || mspID == "" {
		_ = response.ServeJSON(w, http.StatusBadRequest, fmt.Sprintf("Mandatory headers missing"), nil)
		return
	}

	if err := ic.svc.SetInvoiceIdentityGateway(certPEM, keyPEM, mspID); err != nil {
		_ = response.ServeJSON(w, http.StatusBadRequest, fmt.Sprintf("SetInvoiceIdentityGateway"), map[string]interface{}{
			"status":  http.StatusBadRequest,
			"message": fmt.Sprintf("invalid Identity but required: %v", err),
		})
		return
	}

	defer func() {
		if err := ic.svc.CloseInvoiceIdentityGateway(); err != nil {
			_ = response.ServeJSON(w, http.StatusBadRequest, fmt.Sprintf("CloseInvoiceIdentityGateway"), map[string]interface{}{
				"status":  http.StatusBadRequest,
				"message": fmt.Sprintf("failed to close network gateway connection: %v", err),
			})
		}
	}()

	// Parse request body
	var invoiceInfo *model.InvoiceInfo
	if err := json.NewDecoder(r.Body).Decode(&invoiceInfo); err != nil {
		_ = response.ServeJSON(w, http.StatusBadRequest, fmt.Sprintf("Mandatory headers missing"), map[string]interface{}{
			"status":  http.StatusBadRequest,
			"message": "Invalid request body: " + err.Error(),
		})
		return
	}

	requiredFields := map[string]interface{}{
		"invoice_date": invoiceInfo.InvoiceDate,
		"user_id":      invoiceInfo.UserId,
		"subtotal":     invoiceInfo.Subtotal,
	}

	for field, value := range requiredFields {
		if utils.IsValueEmpty(value) {
			_ = response.ServeJSON(w, http.StatusBadRequest, fmt.Sprintf("Mandatory headers missing"), map[string]interface{}{
				"status":  http.StatusBadRequest,
				"message": fmt.Sprintf("Field %s is required", field),
			})
			return
		}
	}

	if err := ic.svc.UpdateInvoice(ctx, invoiceInfo); err != nil {
		_ = response.ServeJSON(w, http.StatusBadRequest, fmt.Sprintf("failed to Update invoice"), map[string]interface{}{
			"status":  http.StatusBadRequest,
			"message": err.Error(),
		})
	}

	err := response.ServeJSON(w, http.StatusOK, "Invoice Updated successfully", invoiceInfo)
	if err == nil {
		ic.lgr.Println("UpdateInvoice", tid, "completed!")
	}

}

func (ic *InvoiceController) DeleteInvoice(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	tid := utils.GetTracingID(ctx)
	ic.lgr.Println("SignUpOrganization", tid, "initialize")

	// Extract identity headers
	certPEM := r.Header.Get("X-User-Cert")
	keyPEM := r.Header.Get("X-User-Key")
	mspID := r.Header.Get("X-User-MSPID")

	if certPEM == "" || keyPEM == "" || mspID == "" {
		_ = response.ServeJSON(w, http.StatusBadRequest, fmt.Sprintf("Mandatory headers missing"), nil)
		return
	}

	if err := ic.svc.SetInvoiceIdentityGateway(certPEM, keyPEM, mspID); err != nil {
		_ = response.ServeJSON(w, http.StatusBadRequest, fmt.Sprintf("SetInvoiceIdentityGateway"), map[string]interface{}{
			"status":  http.StatusBadRequest,
			"message": fmt.Sprintf("invalid Identity but required: %v", err),
		})
		return
	}

	defer func() {
		if err := ic.svc.CloseInvoiceIdentityGateway(); err != nil {
			_ = response.ServeJSON(w, http.StatusBadRequest, fmt.Sprintf("CloseInvoiceIdentityGateway"), map[string]interface{}{
				"status":  http.StatusBadRequest,
				"message": fmt.Sprintf("failed to close network gateway connection: %v", err),
			})
		}
	}()

	// Parse request
	var request struct {
		ID string `json:"id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		_ = response.ServeJSON(w, http.StatusBadRequest, fmt.Sprintf("Mandatory headers missing"), map[string]interface{}{
			"status":  http.StatusBadRequest,
			"message": "Invalid request body",
		})
		return
	}
	if utils.IsValueEmpty(request.ID) {
		_ = response.ServeJSON(w, http.StatusBadRequest, fmt.Sprintf("Mandatory headers missing"), map[string]interface{}{
			"status":  http.StatusBadRequest,
			"message": "Identity ID is required",
		})
		return
	}

	if err := ic.svc.DeleteInvoice(ctx, request.ID); err != nil {
		_ = response.ServeJSON(w, http.StatusBadRequest, fmt.Sprintf("failed to delete invoice"), map[string]interface{}{
			"status":  http.StatusBadRequest,
			"message": err.Error(),
		})
	}

	err := response.ServeJSON(w, http.StatusOK, "Invoice deleted successfully", nil)
	if err == nil {
		ic.lgr.Println("DeleteInvoice", tid, "completed!")
	}
}
