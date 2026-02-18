// Package protoerr builds googleapis rpc errdetails proto messages from
// a resolved apperr error. It is used as a shared building block by the
// connecterr and grpcerr packages so that the detail-mapping logic is
// defined in one place.
package protoerr

import (
	"time"

	apperrdetails "github.com/harwoeck/apperr/errdetails"
	protoerrdetails "google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/durationpb"
)

// Details builds a slice of proto.Message from the resolved error's detail
// fields. The returned messages are the standard googleapis rpc errdetails
// types (LocalizedMessage, ErrorInfo, RequestInfo, ResourceInfo,
// BadRequest, PreconditionFailure, QuotaFailure, Help, RetryInfo,
// DebugInfo).
//
// The caller is responsible for attaching them to the transport-specific
// error type (e.g. *connect.Error or *status.Status).
func Details(resolved *apperrdetails.ResolvedError) []proto.Message {
	var msgs []proto.Message

	if resolved.Localized != nil {
		msgs = append(msgs, &protoerrdetails.LocalizedMessage{
			Locale:  resolved.Localized.Locale.String(),
			Message: resolved.Localized.Text,
		})
	}

	if resolved.ErrorInfo != nil {
		msgs = append(msgs, &protoerrdetails.ErrorInfo{
			Reason:   resolved.ErrorInfo.Reason,
			Domain:   resolved.ErrorInfo.Domain,
			Metadata: resolved.ErrorInfo.Metadata,
		})
	}

	if resolved.RequestInfo != nil {
		msgs = append(msgs, &protoerrdetails.RequestInfo{
			RequestId:   resolved.RequestInfo.RequestID,
			ServingData: resolved.RequestInfo.ServingData,
		})
	}

	if resolved.ResourceInfo != nil {
		msgs = append(msgs, &protoerrdetails.ResourceInfo{
			ResourceType: resolved.ResourceInfo.Type,
			ResourceName: resolved.ResourceInfo.Name,
			Owner:        resolved.ResourceInfo.Owner,
			Description:  resolved.ResourceInfo.Description,
		})
	}

	if len(resolved.FieldViolations) > 0 {
		br := &protoerrdetails.BadRequest{}
		for _, fv := range resolved.FieldViolations {
			br.FieldViolations = append(br.FieldViolations, &protoerrdetails.BadRequest_FieldViolation{
				Field:       fv.Field,
				Description: fv.Description,
			})
		}
		msgs = append(msgs, br)
	}

	if len(resolved.PreconditionViolations) > 0 {
		pf := &protoerrdetails.PreconditionFailure{}
		for _, pv := range resolved.PreconditionViolations {
			pf.Violations = append(pf.Violations, &protoerrdetails.PreconditionFailure_Violation{
				Type:        pv.Type,
				Subject:     pv.Subject,
				Description: pv.Description,
			})
		}
		msgs = append(msgs, pf)
	}

	if len(resolved.QuotaViolations) > 0 {
		qf := &protoerrdetails.QuotaFailure{}
		for _, qv := range resolved.QuotaViolations {
			qf.Violations = append(qf.Violations, &protoerrdetails.QuotaFailure_Violation{
				Subject:     qv.Subject,
				Description: qv.Description,
			})
		}
		msgs = append(msgs, qf)
	}

	if len(resolved.HelpLinks) > 0 {
		h := &protoerrdetails.Help{}
		for _, link := range resolved.HelpLinks {
			h.Links = append(h.Links, &protoerrdetails.Help_Link{
				Url:         link.URL,
				Description: link.Description,
			})
		}
		msgs = append(msgs, h)
	}

	if resolved.RetryInfo != nil {
		msgs = append(msgs, &protoerrdetails.RetryInfo{
			RetryDelay: durationpb.New(time.Duration(resolved.RetryInfo.Delay) * time.Millisecond),
		})
	}

	if resolved.DebugInfo != nil {
		msgs = append(msgs, &protoerrdetails.DebugInfo{
			StackEntries: resolved.DebugInfo.StackEntries,
			Detail:       resolved.DebugInfo.Detail,
		})
	}

	return msgs
}
