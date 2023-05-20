package handler

import (
	context "context"
	connect_go "github.com/bufbuild/connect-go"
	otelconnect "github.com/bufbuild/connect-opentelemetry-go"
	pb "github.com/kostyay/otel-demo/controller/api/calculator/v1"
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

type Math interface {
	Calculate(ctx context.Context, calculation *pb.Calculation) error
}

type calculator struct {
	calculatorv1connect.UnimplementedCalculatorServiceHandler
	db   Storage
	math Math
}

func (c *calculator) Calculate(ctx context.Context, req *connect_go.Request[pb.CalculateRequest]) (*connect_go.Response[pb.CalculateResponse], error) {
	res, err := c.db.CreateCalculation(ctx, req.Msg.GetOwner(), req.Msg.GetExpression())
	if err != nil {
		return nil, err
	}

	err = c.math.Calculate(ctx, &pb.Calculation{
		Id:         uint32(res.ID),
		Owner:      res.Owner,
		Expression: res.Expression,
	})
	if err != nil {
		return nil, err
	}

	response := connect_go.NewResponse(&pb.CalculateResponse{
		Id: uint32(res.ID),
	})

	return response, nil
}

func (c *calculator) List(context.Context, *connect_go.Request[pb.ListRequest]) (*connect_go.Response[pb.ListResponse], error) {
	results, err := c.db.GetCalculations(context.Background())
	if err != nil {
		return nil, err
	}

	var calculations []*pb.Calculation
	for _, result := range results {
		calculation := &pb.Calculation{
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

	response := connect_go.NewResponse(&pb.ListResponse{
		Calculations: calculations,
	})

	return response, nil
}

func New(s Storage, m Math) *calculator {
	return &calculator{db: s, math: m}
}

func (c *calculator) Register(mux *http.ServeMux) {
	mux.Handle(calculatorv1connect.NewCalculatorServiceHandler(c, connect_go.WithInterceptors(otelconnect.NewInterceptor())))
}
