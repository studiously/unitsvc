package unitsvc

import (
	"context"
	"net/url"
	"strings"

	"github.com/go-kit/kit/endpoint"
	httptransport "github.com/go-kit/kit/transport/http"
	"github.com/google/uuid"
	"github.com/studiously/introspector"
	"github.com/studiously/unitsvc/models"
)

type Endpoints struct {
	ListUnitsEndpoint  endpoint.Endpoint
	GetUnitEndpoint    endpoint.Endpoint
	CreateUnitEndpoint endpoint.Endpoint
	RenameUnitEndpoint endpoint.Endpoint
	DeleteUnitEndpoint endpoint.Endpoint
}

func MakeServerEndpoints(s Service) Endpoints {
	return Endpoints{
		ListUnitsEndpoint:  MakeListUnitsEndpoint(s),
		GetUnitEndpoint:    MakeGetUnitEndpoint(s),
		CreateUnitEndpoint: MakeCreateUnitEndpoint(s),
		RenameUnitEndpoint: MakeRenameUnitEndpoint(s),
		DeleteUnitEndpoint: MakeDeleteUnitEndpoint(s),
	}
}

func MakeClientEndpoints(instance string) (Endpoints, error) {
	if !strings.HasPrefix(instance, "http") {
		instance = "http://" + instance
	}
	tgt, err := url.Parse(instance)
	if err != nil {
		return Endpoints{}, err
	}
	tgt.Path = ""

	options := []httptransport.ClientOption{
		httptransport.ClientBefore(introspector.FromHTTPContext()),
	}

	return Endpoints{
		ListUnitsEndpoint:  httptransport.NewClient("GET", tgt, EncodeListUnitsRequest, DecodeListUnitsResponse, options...).Endpoint(),
		GetUnitEndpoint:    httptransport.NewClient("GET", tgt, EncodeGetUnitRequest, DecodeGetUnitResponse, options...).Endpoint(),
		CreateUnitEndpoint: httptransport.NewClient("POST", tgt, EncodeCreateUnitRequest, DecodeCreateUnitResponse, options...).Endpoint(),
		RenameUnitEndpoint: httptransport.NewClient("PATCH", tgt, EncodeRenameUnitRequest, DecodeRenameUnitResponse, options...).Endpoint(),
		DeleteUnitEndpoint: httptransport.NewClient("DELETE", tgt, EncodeDeleteUnitRequest, DecodeDeleteUnitResponse, options...).Endpoint(),
	}, nil
}

func (e Endpoints) ListUnits(ctx context.Context, classID uuid.UUID) ([]uuid.UUID, error) {
	request := listUnitsRequest{ClassID: classID}
	response, err := e.ListUnitsEndpoint(ctx, request)
	if err != nil {
		return nil, err
	}
	resp := response.(listUnitsResponse)
	return resp.Units, resp.Error
}

func (e Endpoints) GetUnit(ctx context.Context, unitID uuid.UUID) (*models.Unit, error) {
	request := getUnitRequest{UnitID: unitID}
	response, err := e.GetUnitEndpoint(ctx, request)
	if err != nil {
		return nil, err
	}
	resp := response.(getUnitResponse)
	return resp.Unit, resp.Error
}

func (e Endpoints) CreateUnit(ctx context.Context, classID uuid.UUID, title string) error {
	request := createUnitRequest{ClassID: classID, Title: title}
	response, err := e.CreateUnitEndpoint(ctx, request)
	if err != nil {
		return err
	}
	resp := response.(createUnitResponse)
	return resp.Error
}

func (e Endpoints) RenameUnit(ctx context.Context, unitID uuid.UUID, title string) error {
	request := renameUnitRequest{UnitID: unitID, Title: title}
	response, err := e.RenameUnitEndpoint(ctx, request)
	if err != nil {
		return err
	}
	resp := response.(renameUnitResponse)
	return resp.Error
}

func (e Endpoints) DeleteUnit(ctx context.Context, unitID uuid.UUID) error {
	request := deleteUnitRequest{UnitID: unitID}
	response, err := e.DeleteUnitEndpoint(ctx, request)
	if err != nil {
		return err
	}
	resp := response.(deleteUnitResponse)
	return resp.Error
}

func MakeListUnitsEndpoint(s Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		req := request.(listUnitsRequest)
		units, e := s.ListUnits(ctx, req.ClassID)
		return listUnitsResponse{units, e}, nil
	}
}

type listUnitsRequest struct {
	ClassID uuid.UUID `json:"class_id"`
}

type listUnitsResponse struct {
	Units []uuid.UUID `json:"units,omitempty"`
	Error error `json:"error,omitempty"`
}

func (r listUnitsResponse) error() error {
	return r.Error
}

func MakeGetUnitEndpoint(s Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		req := request.(getUnitRequest)
		unit, e := s.GetUnit(ctx, req.UnitID)
		return getUnitResponse{unit, e}, nil
	}
}

type getUnitRequest struct {
	UnitID uuid.UUID `json:"unit_id"`
}

type getUnitResponse struct {
	Unit  *models.Unit `json:"unit,omitempty"`
	Error error `json:"error,omitempty"`
}

func (r getUnitResponse) error() error {
	return r.Error
}

func MakeCreateUnitEndpoint(s Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		req := request.(createUnitRequest)
		e := s.CreateUnit(ctx, req.ClassID, req.Title)
		return createUnitResponse{e}, nil
	}
}

type createUnitRequest struct {
	ClassID uuid.UUID `json:"class_id"`
	Title   string `json:"title"`
}

type createUnitResponse struct {
	Error error `json:"error,omitempty"`
}

func (r createUnitResponse) error() error {
	return r.Error
}

func MakeRenameUnitEndpoint(s Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		req := request.(renameUnitRequest)
		e := s.RenameUnit(ctx, req.UnitID, req.Title)
		return renameUnitResponse{e}, nil
	}
}

type renameUnitRequest struct {
	UnitID uuid.UUID `json:"unit_id"`
	Title  string `json:"title"`
}

type renameUnitResponse struct {
	Error error `json:"error,omitempty"`
}

func (r renameUnitResponse) error() error {
	return r.Error
}

func MakeDeleteUnitEndpoint(s Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		req := request.(deleteUnitRequest)
		e := s.DeleteUnit(ctx, req.UnitID)
		return deleteUnitResponse{e}, nil
	}
}

type deleteUnitRequest struct {
	UnitID uuid.UUID `json:"unit_id"`
}

type deleteUnitResponse struct {
	Error error `json:"error,omitempty"`
}

func (r deleteUnitResponse) error() error {
	return r.Error
}
