// Package apperr provides a unified application-error generation interface.
// Errors constructed with apperr can be localized and extended/modified with
// other options. When finalized they are rendered into a RenderedError object,
// which can be converted to native error types for various different
// frameworks like GRPC, Twirp, Plain-HTTP, etc.
//
// Example:
//   err := apperr.Unauthenticated("provided password is invalid",
//       apperr.Localize("INVALID_PASSWORD"))
//
// In a middleware/interceptor:
//
//   var rendered *apperr.RenderedError
//   if a, ok := err.(*apperr.AppError); ok {
//       rendered, _ = apperr.Render(a, apperr.RenderLocalized(adapter, "en-US"))
//   } else {
//       internal := apperr.Internal("internal error")
//       rendered, _ = apperr.Render(internal, apperr.RenderLocalized(adapter, "en-US"))
//   }
//   httpStatus, httpBody, _ := httperr.Convert(rendered)
//
// Example Output:
//   httpStatus = 401 (Unauthorized)
//   httpBody =
//   {
//       "message": "provided password is invalid",
//       "code": "Unauthenticated",
//       "localized": {
//           "userMessage": "The entered password isn't correct. Please try again",
//           "userMessageShort": "Not authenticated",
//           "locale": "en-US"
//       }
//   }
//
// The provided converters are:
//   1) GRPC (go get github.com/harwoeck/apperr/grpcerr)
//      grpcStatus, err := grpcerr.Convert(*RenderedError)
//   2) HTTP (go get github.com/harwoeck/apperr/httperr)
//      httpStatus, httpBody, err := httperr.Convert(*RenderedError)
//   3) Twirp (go get github.com/harwoeck/apperr/twirperr)
//      twirpError := twirperr.Convert(*RenderedError)
package apperr
