package domain

import (
	"time"

	pb "github.com/kostyay/otel-demo/controller/api/calculator/v1"
	"google.golang.org/protobuf/types/known/timestamppb"
	"gorm.io/gorm"
)

type Calculation struct {
	gorm.Model
	Owner       string
	Expression  string
	Result      *float64
	CompletedAt *time.Time
}

func (c *Calculation) Proto() *pb.Calculation {
	result := &pb.Calculation{
		Id:         uint32(c.ID),
		Owner:      c.Owner,
		Expression: c.Expression,
		UpdatedAt:  timestamppb.New(c.UpdatedAt),
		CreatedAt:  timestamppb.New(c.CreatedAt),
	}

	if c.Result != nil {
		result.Result = *c.Result
	}

	if c.CompletedAt != nil {
		result.CompletedAt = timestamppb.New(*c.CompletedAt)
	}

	return result
}
