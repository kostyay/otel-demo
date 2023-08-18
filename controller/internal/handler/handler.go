package handler

import (
	context "context"
	"fmt"
	"github.com/kostyay/otel-demo/common/log"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
	"net/http"

	connect_go "github.com/bufbuild/connect-go"
	otelconnect "github.com/bufbuild/connect-opentelemetry-go"
	pb "github.com/kostyay/otel-demo/controller/api/calculator/v1"
	"github.com/kostyay/otel-demo/controller/api/calculator/v1/calculatorv1connect"
	"github.com/kostyay/otel-demo/controller/internal/domain"
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
	// use existing span attached to the context by the middleware
	span := trace.SpanFromContext(ctx)
	// add some attributes to the span
	span.SetAttributes(attribute.String("owner", req.Msg.GetOwner()))

	// log with traceid
	log.WithContext(ctx).Info("Got calculation request!")
	if req.Msg.GetOwner() == "error" {
		span.RecordError(fmt.Errorf("owner is invalid"))
		span.SetStatus(codes.Error, "owner is invalid")
		return nil, connect_go.NewError(connect_go.CodeInvalidArgument, fmt.Errorf("owner is invalid"))
	}

	// some span events
	span.AddEvent("Creating calculation in database")
	res, err := c.db.CreateCalculation(ctx, req.Msg.GetOwner(), req.Msg.GetExpression())
	if err != nil {
		return nil, err
	}

	span.SetAttributes(attribute.Int("id", int(res.ID)))

	span.AddEvent(fmt.Sprintf("Dispatching calculation %d reqeust to math service", res.ID))

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

func (c *calculator) List(ctx context.Context, req *connect_go.Request[pb.ListRequest]) (*connect_go.Response[pb.ListResponse], error) {
	results, err := c.db.GetCalculations(ctx)
	if err != nil {
		return nil, err
	}

	var calculations []*pb.Calculation
	for _, result := range results {
		calculations = append(calculations, result.Proto())
	}

	response := connect_go.NewResponse(&pb.ListResponse{
		Calculations: calculations,
	})

	return response, nil
}

func (c *calculator) Get(ctx context.Context, req *connect_go.Request[pb.GetRequest]) (*connect_go.Response[pb.GetResponse], error) {
	res, err := c.db.GetCalculation(ctx, uint(req.Msg.GetId()))
	if err != nil {
		return nil, err
	}

	response := connect_go.NewResponse(&pb.GetResponse{
		Calculation: res.Proto(),
	})

	return response, nil
}

func (c *calculator) Cleanup(ctx context.Context, req *connect_go.Request[pb.CleanupRequest]) (*connect_go.Response[pb.CleanupResponse], error) {
	return connect_go.NewResponse(&pb.CleanupResponse{}), nil
}
func New(s Storage, m Math) *calculator {
	return &calculator{db: s, math: m}
}

func (c *calculator) Register(mux *http.ServeMux) {
	mux.Handle(calculatorv1connect.NewCalculatorServiceHandler(c, connect_go.WithInterceptors(otelconnect.NewInterceptor())))
}
