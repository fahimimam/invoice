package api

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"github.com/triapex/auth/api/response"
	"github.com/triapex/auth/config"
	"github.com/triapex/auth/internal/gateways"
	"github.com/triapex/auth/internal/infra"
	"github.com/triapex/auth/internal/service"
	"github.com/triapex/auth/logger"
	"github.com/triapex/auth/model"
	"github.com/triapex/auth/utils"
	"google.golang.org/grpc"
	"net/http"
	"strings"
)

// OrgController ...
type InvoiceController struct {
	svc            service.InvoiceService
	grpc           *grpc.ClientConn
	lgr            logger.StructLogger
	identityConfig *config.Invoice
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

	// decode request body to OrganizationInfo
	body := &invoiceCreatePld{}
	if err := json.NewDecoder(r.Body).Decode(body); err != nil {
		_ = response.ServeJSON(w, http.StatusBadRequest, err.Error(), nil)
		return
	}

	if body.OrganizationName == "" {
		_ = response.ServeJSON(w, http.StatusBadRequest, "invoice name is required", nil)
		return
	}

	body.OrganizationName = strings.ToLower(strings.TrimSpace(body.OrganizationName))
	createResponse, err := uc.svc.CreateOrganization(ctx, &model.invoice{
		UniqueOrganizationIdentifier: fmt.Sprintf("org_%v", uuid.New()),
		OrganizationName:             body.OrganizationName,
	})
	if err != nil {
		if errors.Is(err, infra.ErrDuplicateKey) {
			_ = response.ServeJSON(w, http.StatusBadRequest, "invoice name already exists", nil)
			return
		}
		_ = response.ServeJSON(w, http.StatusBadRequest, err.Error(), nil)
		return
	}

	err = response.ServeJSON(w, http.StatusOK, utils.SuccessMessage, map[string]interface{}{
		"organization_id":                createResponse.ID,
		"organization_name":              createResponse.OrganizationName,
		"unique_organization_identifier": createResponse.UniqueOrganizationIdentifier,
	})
	if err == nil {
		uc.lgr.Println("Create invoice", tid, "complete!")
	}

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
	identityGateway, contract, err := gateways.NewGatewayForIdentity(uc.grpc, certPEM, keyPEM, mspID, uc.IdentityConfig)
	if err != nil {
		_ = response.ServeJSON(w, http.StatusBadRequest, "", map[string]interface{}{
			"status":  http.StatusUnauthorized,
			"message": "Invalid identity credentials: " + err.Error(),
		})
		return
	}
	defer identityGateway.Close()

	// Parse request body
	var idnty model.Identity
	if err := json.NewDecoder(r.Body).Decode(&idnty); err != nil {
		_ = response.ServeJSON(w, http.StatusBadRequest, fmt.Sprintf("Mandatory headers missing"), nil)
		return
	}
	// Validate required fields
	if utils.IsValueEmpty(idnty.Id) {
		_ = response.ServeJSON(w, http.StatusBadRequest, fmt.Sprintf("Mandatory headers missing"), map[string]interface{}{
			"status":  http.StatusBadRequest,
			"message": "Invoice ID is required",
		})
		return
	}
	requiredFields := map[string]string{
		"firstName":  idnty.FirstName,
		"phone":      idnty.Phone,
		"nationalID": idnty.NationalID,
	}
	for field, value := range requiredFields {
		if utils.IsValueEmpty(value) {
			_ = response.ServeJSON(w, http.StatusBadRequest, fmt.Sprintf("Mandatory field missing"), map[string]interface{}{
				"status":  http.StatusBadRequest,
				"message": fmt.Sprintf("Field %s is required", field),
			})
			return
		}
	}

	// Submit transaction
	assetJSON, marshalErr := json.Marshal(idnty)
	if marshalErr != nil {
		_ = response.ServeJSON(w, http.StatusBadRequest, fmt.Sprintf("Mandatory field missing"), map[string]interface{}{
			"status":  http.StatusInternalServerError,
			"message": "Error marshaling identity: " + err,
		})
		return
	}

	if _, err := contract.SubmitTransaction("CreateIdentity", string(assetJSON)); err != nil {
		respondJSON(w, http.StatusInternalServerError, map[string]interface{}{
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

	// decode request body to OrganizationInfo
	body := &invoiceCreatePld{}
	if err := json.NewDecoder(r.Body).Decode(body); err != nil {
		_ = response.ServeJSON(w, http.StatusBadRequest, err.Error(), nil)
		return
	}
	createResponse, err := uc.svc.UpdateOrganization(ctx, &model.invoice{
		OrganizationName: body.OrganizationName,
	})
	if err != nil {
		_ = response.ServeJSON(w, http.StatusBadRequest, err.Error(), nil)
		return
	}

	err = response.ServeJSON(w, http.StatusOK, utils.SuccessMessage, map[string]interface{}{
		"organization_id":                createResponse.ID,
		"organization_name":              createResponse.OrganizationName,
		"unique_organization_identifier": createResponse.UniqueOrganizationIdentifier,
	})
	if err == nil {
		uc.lgr.Println("Update invoice", tid, "complete!")
	}
}

func (uc *InvoiceController) DeleteInvoiceHandler(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	tid := utils.GetTracingID(ctx)
	uc.lgr.Println("SignUpOrganization", tid, "initialize")

	// decode request body to OrganizationInfo
	body := &invoiceCreatePld{}
	if err := json.NewDecoder(r.Body).Decode(body); err != nil {
		_ = response.ServeJSON(w, http.StatusBadRequest, err.Error(), nil)
		return
	}
	createResponse, err := uc.svc.UpdateOrganization(ctx, &model.invoice{
		OrganizationName: body.OrganizationName,
	})
	if err != nil {
		_ = response.ServeJSON(w, http.StatusBadRequest, err.Error(), nil)
		return
	}

	err = response.ServeJSON(w, http.StatusOK, utils.SuccessMessage, map[string]interface{}{
		"organization_id":                createResponse.ID,
		"organization_name":              createResponse.OrganizationName,
		"unique_organization_identifier": createResponse.UniqueOrganizationIdentifier,
	})
	if err == nil {
		uc.lgr.Println("Update invoice", tid, "complete!")
	}
}

func (uc *InvoiceController) GetInvoiceHandler(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	tid := utils.GetTracingID(ctx)
	uc.lgr.Println("SignUpOrganization", tid, "initialize")

	// decode request body to OrganizationInfo
	body := &invoiceCreatePld{}
	if err := json.NewDecoder(r.Body).Decode(body); err != nil {
		_ = response.ServeJSON(w, http.StatusBadRequest, err.Error(), nil)
		return
	}
	createResponse, err := uc.svc.UpdateOrganization(ctx, &model.invoice{
		OrganizationName: body.OrganizationName,
	})
	if err != nil {
		_ = response.ServeJSON(w, http.StatusBadRequest, err.Error(), nil)
		return
	}

	err = response.ServeJSON(w, http.StatusOK, utils.SuccessMessage, map[string]interface{}{
		"organization_id":                createResponse.ID,
		"organization_name":              createResponse.OrganizationName,
		"unique_organization_identifier": createResponse.UniqueOrganizationIdentifier,
	})
	if err == nil {
		uc.lgr.Println("Update invoice", tid, "complete!")
	}
}
