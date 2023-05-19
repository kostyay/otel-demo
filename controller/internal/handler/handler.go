package handler

import (
	context "context"
	connect_go "github.com/bufbuild/connect-go"
	otelconnect "github.com/bufbuild/connect-opentelemetry-go"
	v1 "github.com/kostyay/otel-demo/controller/api/calculator/v1"
	"github.com/kostyay/otel-demo/controller/api/calculator/v1/calculatorv1connect"
	"github.com/kostyay/otel-demo/controller/internal/domain"
	"google.golang.org/protobuf/types/known/timestamppb"
	"net/http"
)

type Storage interface {
	CreateCalculation(ctx context.Context, owner, expression string) (*domain.Calculation, error)
	GetCalculation(ctx context.Context, id uint) (*domain.Calculation, error)
	GetCalculations(ctx context.Context) ([]*domain.Calculation, error)
	UpdateResult(ctx context.Context, id uint, result float64) error
}

type calculator struct {
	calculatorv1connect.UnimplementedCalculatorServiceHandler
	db Storage
}

func (c *calculator) Calculate(ctx context.Context, req *connect_go.Request[v1.CalculateRequest]) (*connect_go.Response[v1.CalculateResponse], error) {
	res, err := c.db.CreateCalculation(ctx, req.Msg.GetOwner(), req.Msg.GetExpression())
	if err != nil {
		return nil, err
	}

	response := connect_go.NewResponse(&v1.CalculateResponse{
		Id: uint32(res.ID),
	})

	return response, nil
}

func (c *calculator) List(context.Context, *connect_go.Request[v1.ListRequest]) (*connect_go.Response[v1.ListResponse], error) {
	results, err := c.db.GetCalculations(context.Background())
	if err != nil {
		return nil, err
	}

	var calculations []*v1.Calculation
	for _, result := range results {
		calculation := &v1.Calculation{
			Id:         uint32(result.ID),
			Owner:      result.Owner,
			Expression: result.Expression,
			CreatedAt:  timestamppb.New(result.CreatedAt),
			UpdatedAt:  timestamppb.New(result.UpdatedAt),
		}
		if result.Result != nil {
			calculation.Result = *result.Result
		}
		if result.CompletedAt != nil {
			calculation.CompletedAt = timestamppb.New(*result.CompletedAt)
		}

		calculations = append(calculations, calculation)
	}

	response := connect_go.NewResponse(&v1.ListResponse{
		Calculations: calculations,
	})

	return response, nil
}

func New(s Storage) *calculator {
	return &calculator{db: s}
}

func (c *calculator) Register(mux *http.ServeMux) {
	mux.Handle(calculatorv1connect.NewCalculatorServiceHandler(c, connect_go.WithInterceptors(otelconnect.NewInterceptor())))
}
