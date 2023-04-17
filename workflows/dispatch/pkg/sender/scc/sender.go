package scc

import (
	"context"
	"log"
	"time"

	scc "cloud.google.com/go/securitycenter/apiv1"
	pb "cloud.google.com/go/securitycenter/apiv1/securitycenterpb"
	"github.com/pkg/errors"
	ca "google.golang.org/api/containeranalysis/v1"
	"google.golang.org/protobuf/types/known/structpb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// Sender marshals the occurrence to stdout.
func Sender(ctx context.Context, occ *ca.Occurrence) error {
	if occ == nil {
		return errors.New("occurrence is nil")
	}

	c, err := scc.NewClient(ctx)
	if err != nil {
		return errors.Wrap(err, "failed to create SCC client")
	}
	defer c.Close()

	createTime, err := time.Parse(time.RFC3339, occ.UpdateTime)
	if err != nil {
		return errors.Wrap(err, "failed to parse creation time")
	}

	f, err := c.CreateFinding(ctx, &pb.CreateFindingRequest{
		Parent:    occ.ResourceUri,
		FindingId: occ.Name,
		Finding: &pb.Finding{
			Name:         occ.Name,
			Parent:       occ.ResourceUri,
			ResourceName: occ.Name,
			State:        pb.Finding_ACTIVE,
			Category:     "vulnerability",
			ExternalUri:  occ.ResourceUri,
			EventTime:    timestamppb.New(createTime),

			SourceProperties: map[string]*structpb.Value{
				"severity": {
					Kind: &structpb.Value_StringValue{
						StringValue: occ.Vulnerability.Severity,
					},
				},
			},
		},
	})

	if err != nil {
		return errors.Wrap(err, "failed to create finding")
	}

	log.Printf("Created finding: %s", f.CanonicalName)

	return nil
}
