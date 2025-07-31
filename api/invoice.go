package api

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/go-chi/chi/v5"
	"github.com/triapex/auth/api/response"
	"github.com/triapex/auth/config"
	"github.com/triapex/auth/internal/gateways"
	"github.com/triapex/auth/internal/service"
	"github.com/triapex/auth/logger"
	"github.com/triapex/auth/model"
	"github.com/triapex/auth/utils"
	"google.golang.org/grpc"
	"net/http"
)

// OrgController ...
type InvoiceController struct {
	svc            service.InvoiceService
	grpc           *grpc.ClientConn
	lgr            logger.StructLogger
	invoiceConfig  *config.Invoice
	identityConfig *config.Identity
}

// NewInvoiceController ...
func NewInvoiceController(svc service.InvoiceService, grpc *grpc.ClientConn, lgr logger.StructLogger) *InvoiceController {
	return &InvoiceController{
		svc:  svc,
		grpc: grpc,
		lgr:  lgr,
	}
}

// SetLogger ...
func (uc *InvoiceController) SetLogger(lgr logger.StructLogger) {
	uc.lgr = lgr
}

type invoiceCreatePld struct {
	OrganizationName string `json:"organization_name"`
}

// CreateInvoiceHandler ...
func (uc *InvoiceController) CreateInvoiceHandler(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	tid := utils.GetTracingID(ctx)
	uc.lgr.Println("SignUpOrganization", tid, "initialize")
	// ##################################################################################################################################################################

	// Extract identity headers
	certPEM := r.Header.Get("X-User-Cert")
	keyPEM := r.Header.Get("X-User-Key")
	mspID := r.Header.Get("X-User-MSPID")

	if certPEM == "" || keyPEM == "" || mspID == "" {
		_ = response.ServeJSON(w, http.StatusBadRequest, fmt.Sprintf("Mandatory headers missing"), nil)
		return
	}

	// Create gateway connection for this identity
	identityGateway, identityContract, err := gateways.NewGatewayForIdentity(uc.grpc, certPEM, keyPEM, mspID, uc.identityConfig)
	if err != nil {
		_ = response.ServeJSON(w, http.StatusBadRequest, "", map[string]interface{}{
			"status":  http.StatusUnauthorized,
			"message": "Invalid identity credentials: " + err.Error(),
		})
		return
	}
	defer identityGateway.Close()

	// Parse request body
	// Get ID from URL
	id := chi.URLParam(r, "id")
	if utils.IsValueEmpty(id) {
		_ = response.ServeJSON(w, http.StatusBadRequest, fmt.Sprintf("Mandatory headers missing"), map[string]interface{}{
			"status":  http.StatusBadRequest,
			"message": "Identity ID is required",
		})
		return
	}

	// Evaluate transaction
	result, err := identityContract.EvaluateTransaction("ReadIdentity", id)
	if err != nil {
		_ = response.ServeJSON(w, http.StatusBadRequest, fmt.Sprintf("Mandatory headers missing"), map[string]interface{}{
			"status":  http.StatusNotFound,
			"message": "Identity not found: " + err.Error(),
		})
		return
	}

	var identity *model.Identity
	if err = json.Unmarshal(result, &identity); err != nil {
		_ = response.ServeJSON(w, http.StatusBadRequest, fmt.Sprintf("Mandatory headers missing"), map[string]interface{}{
			"status":  http.StatusInternalServerError,
			"message": "Error parsing identity data:  " + err.Error(),
		})
		return
	}

	// Check if identity is verified or not
	if identity == nil || identity.Id != id {
		_ = response.ServeJSON(w, http.StatusBadRequest, fmt.Sprintf("Mandatory headers missing"), map[string]interface{}{
			"status":  http.StatusNotFound,
			"message": "Identity not found",
		})
	}

	// Create gateway connection for this identity
	invoiceGateway, invoiceContract, err := gateways.NewGatewayForInvoice(uc.grpc, certPEM, keyPEM, mspID, uc.invoiceConfig)
	if err != nil {
		_ = response.ServeJSON(w, http.StatusBadRequest, fmt.Sprintf("Mandatory headers missing"), map[string]interface{}{
			"status":  http.StatusUnauthorized,
			"message": "Invalid identity credentials: " + err.Error(),
		})
		return
	}
	defer invoiceGateway.Close()

	// Parse request body
	var invoiceDetails *model.InvoiceInfo
	if err := json.NewDecoder(r.Body).Decode(&invoiceDetails); err != nil {
		_ = response.ServeJSON(w, http.StatusBadRequest, fmt.Sprintf("Mandatory headers missing"), map[string]interface{}{
			"status":  http.StatusBadRequest,
			"message": "Invalid request body: " + err.Error(),
		})
		return
	}

	requiredFields := map[string]string{
		"invoice_date": invoiceDetails.InvoiceDate,
		"user_id":      invoiceDetails.UserId,
		"subtotal":     invoiceDetails.Subtotal,
		//TODO: Add More if needed
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

	// Submit transaction
	assetJSON, err := json.Marshal(invoiceDetails)
	if err != nil {
		_ = response.ServeJSON(w, http.StatusBadRequest, fmt.Sprintf("Mandatory headers missing"), map[string]interface{}{
			"status":  http.StatusInternalServerError,
			"message": "Error marshaling identity: " + err.Error(),
		})
		return
	}

	// TODO: Move NAME to constants
	if _, err := invoiceContract.SubmitTransaction("CreateInvoice", string(assetJSON)); err != nil {
		_ = response.ServeJSON(w, http.StatusBadRequest, fmt.Sprintf("Mandatory headers missing"), map[string]interface{}{
			"status":  http.StatusInternalServerError,
			"message": "Chaincode error: " + err.Error(),
		})
		return
	}

}

// UpdateInvoiceHandler ...
func (uc *InvoiceController) UpdateInvoiceHandler(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	tid := utils.GetTracingID(ctx)
	uc.lgr.Println("SignUpOrganization", tid, "initialize")
	// ##################################################################################################################################################################

	// Extract identity headers
	certPEM := r.Header.Get("X-User-Cert")
	keyPEM := r.Header.Get("X-User-Key")
	mspID := r.Header.Get("X-User-MSPID")

	if certPEM == "" || keyPEM == "" || mspID == "" {
		_ = response.ServeJSON(w, http.StatusBadRequest, fmt.Sprintf("Mandatory headers missing"), nil)
		return
	}

	// Create gateway connection for this identity
	identityGateway, identityContract, err := gateways.NewGatewayForIdentity(uc.grpc, certPEM, keyPEM, mspID, uc.identityConfig)
	if err != nil {
		_ = response.ServeJSON(w, http.StatusBadRequest, "", map[string]interface{}{
			"status":  http.StatusUnauthorized,
			"message": "Invalid identity credentials: " + err.Error(),
		})
		return
	}
	defer identityGateway.Close()

	// Parse request body
	// Get ID from URL
	id := chi.URLParam(r, "id")
	if utils.IsValueEmpty(id) {
		_ = response.ServeJSON(w, http.StatusBadRequest, fmt.Sprintf("Mandatory headers missing"), map[string]interface{}{
			"status":  http.StatusBadRequest,
			"message": "Identity ID is required",
		})
		return
	}

	// Evaluate transaction
	result, err := identityContract.EvaluateTransaction("ReadIdentity", id)
	if err != nil {
		_ = response.ServeJSON(w, http.StatusBadRequest, fmt.Sprintf("Mandatory headers missing"), map[string]interface{}{
			"status":  http.StatusNotFound,
			"message": "Identity not found: " + err.Error(),
		})
		return
	}

	var identity *model.Identity
	if err = json.Unmarshal(result, &identity); err != nil {
		_ = response.ServeJSON(w, http.StatusBadRequest, fmt.Sprintf("Mandatory headers missing"), map[string]interface{}{
			"status":  http.StatusInternalServerError,
			"message": "Error parsing identity data:  " + err.Error(),
		})
		return
	}

	// Check if identity is verified or not
	if identity == nil || identity.Id != id {
		_ = response.ServeJSON(w, http.StatusBadRequest, fmt.Sprintf("Mandatory headers missing"), map[string]interface{}{
			"status":  http.StatusNotFound,
			"message": "Identity not found",
		})
	}

	// Create gateway connection for this identity
	invoiceGateway, invoiceContract, err := gateways.NewGatewayForInvoice(uc.grpc, certPEM, keyPEM, mspID, uc.invoiceConfig)
	if err != nil {
		_ = response.ServeJSON(w, http.StatusBadRequest, fmt.Sprintf("Mandatory headers missing"), map[string]interface{}{
			"status":  http.StatusUnauthorized,
			"message": "Invalid identity credentials: " + err.Error(),
		})
		return
	}
	defer invoiceGateway.Close()

	// Parse request body
	var invoiceDetails *model.InvoiceInfo
	if err := json.NewDecoder(r.Body).Decode(&invoiceDetails); err != nil {
		_ = response.ServeJSON(w, http.StatusBadRequest, fmt.Sprintf("Mandatory headers missing"), map[string]interface{}{
			"status":  http.StatusBadRequest,
			"message": "Invalid request body: " + err.Error(),
		})
		return
	}

	requiredFields := map[string]string{
		"invoice_date": invoiceDetails.InvoiceDate,
		"user_id":      invoiceDetails.UserId,
		"subtotal":     invoiceDetails.Subtotal,
		//TODO: Add More if needed
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

	// Submit transaction
	assetJSON, err := json.Marshal(invoiceDetails)
	if err != nil {
		_ = response.ServeJSON(w, http.StatusBadRequest, fmt.Sprintf("Mandatory headers missing"), map[string]interface{}{
			"status":  http.StatusInternalServerError,
			"message": "Error marshaling identity: " + err.Error(),
		})
		return
	}

	// TODO: Move NAME to constants
	if _, err := invoiceContract.SubmitTransaction("UpdateInvoice", string(assetJSON)); err != nil {
		_ = response.ServeJSON(w, http.StatusBadRequest, fmt.Sprintf("Mandatory headers missing"), map[string]interface{}{
			"status":  http.StatusInternalServerError,
			"message": "Chaincode error: " + err.Error(),
		})
		return
	}
}

func (uc *InvoiceController) DeleteInvoiceHandler(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	tid := utils.GetTracingID(ctx)
	uc.lgr.Println("SignUpOrganization", tid, "initialize")

	// Extract identity headers
	certPEM := r.Header.Get("X-User-Cert")
	keyPEM := r.Header.Get("X-User-Key")
	mspID := r.Header.Get("X-User-MSPID")

	if certPEM == "" || keyPEM == "" || mspID == "" {
		_ = response.ServeJSON(w, http.StatusBadRequest, fmt.Sprintf("Mandatory headers missing"), nil)
		return
	}

	// Create gateway connection for this identity
	invoiceGateway, invoiceContract, err := gateways.NewGatewayForInvoice(uc.grpc, certPEM, keyPEM, mspID, uc.invoiceConfig)
	if err != nil {
		_ = response.ServeJSON(w, http.StatusBadRequest, fmt.Sprintf("Mandatory headers missing"), map[string]interface{}{
			"status":  http.StatusUnauthorized,
			"message": "Invalid identity credentials: " + err.Error(),
		})
		return
	}
	defer invoiceGateway.Close()
	// decode request body to OrganizationInfo
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

	// Submit transaction
	if _, err = invoiceContract.SubmitTransaction("DeleteInvoice", request.ID); err != nil {
		_ = response.ServeJSON(w, http.StatusBadRequest, fmt.Sprintf("Mandatory headers missing"), map[string]interface{}{
			"status":  http.StatusInternalServerError,
			"message": "Chaincode error: " + err.Error(),
		})
		return
	}

	_ = response.ServeJSON(w, http.StatusBadRequest, fmt.Sprintf("Mandatory headers missing"), map[string]interface{}{
		"invoice_id": request.ID,
	})
}

func (uc *InvoiceController) GetInvoiceHandler(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	tid := utils.GetTracingID(ctx)
	uc.lgr.Println("SignUpOrganization", tid, "initialize")

	// Extract identity headers
	certPEM := r.Header.Get("X-User-Cert")
	keyPEM := r.Header.Get("X-User-Key")
	mspID := r.Header.Get("X-User-MSPID")

	if certPEM == "" || keyPEM == "" || mspID == "" {
		_ = response.ServeJSON(w, http.StatusBadRequest, fmt.Sprintf("Mandatory headers missing"), nil)
		return
	}

	// Create gateway connection for this identity
	invoiceGateway, invoiceContract, err := gateways.NewGatewayForInvoice(uc.grpc, certPEM, keyPEM, mspID, uc.invoiceConfig)
	if err != nil {
		_ = response.ServeJSON(w, http.StatusBadRequest, fmt.Sprintf("Mandatory headers missing"), map[string]interface{}{
			"status":  http.StatusUnauthorized,
			"message": "Invalid identity credentials: " + err.Error(),
		})
		return
	}
	defer invoiceGateway.Close()
	// decode request body to OrganizationInfo
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
	// Evaluate transaction
	result, err := invoiceContract.EvaluateTransaction("ReadInvoice", request.ID)
	if err != nil {
		_ = response.ServeJSON(w, http.StatusBadRequest, fmt.Sprintf("Mandatory headers missing"), map[string]interface{}{
			"status":  http.StatusNotFound,
			"message": "Identity not found: " + err.Error(),
		})
		return
	}

	var invoice model.InvoiceInfo
	if err = json.Unmarshal(result, &invoice); err != nil {
		_ = response.ServeJSON(w, http.StatusBadRequest, fmt.Sprintf("Mandatory headers missing"), map[string]interface{}{
			"status":  http.StatusInternalServerError,
			"message": "Error parsing identity data:  " + err.Error(),
		})
		return
	}

	err = response.ServeJSON(w, http.StatusOK, utils.SuccessMessage, map[string]interface{}{
		"invoice_id": invoice.InvoiceId,
	})
	if err == nil {
		uc.lgr.Println("Update invoice", tid, "complete!")
	}
}
