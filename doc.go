// Package apperr provides a unified application-error generation interface.
// Errors constructed with apperr can be localized and extended/modified with
// other options. When finalized they are rendered into a RenderedError object,
// which can be converted to native error types for various different
// frameworks like GRPC, Twirp, Plain-HTTP, etc.
//
// Example:
//
//	err := apperr.Unauthenticated("provided password is invalid",
//	    apperr.Localize("INVALID_PASSWORD"))
//
// Setup:
//
//	// configure i18n adapter
//	i18nAdapter := NewI18nAdapter(config)
//
//	// configure language matcher with available languages for i18n
//	matcher := language.NewMatcher([]language.Tag{language.English, language.German})
//
// In a middleware/interceptor:
//
//		// call request handler and get error back
//		err := handler(r, w)
//
//		// check if request failed with apperr, or for unknown reasons (then
//		// default to Internal
//		var ae *apperr.AppError
//		if x, ok := err.(*apperr.AppError); ok {
//		   ae = x
//		} else {
//		    ae = apperr.Internal("internal error")
//		}
//
//	 // get best match for user language
//		t, q, err := language.ParseAcceptLanguage(r.Header.Get("Accept-Language"))
//		userLang, _, _ := matcher.Match(t...)
//
//	 // finalize apperror to something we can return to users
//		rendered := apperrutils.Render(ae, apperr.RenderLocalized(i18nAdapter, userLang))
//
//	 // convert rendered error to the output format of our protocol
//		httpStatus, httpBody, _ := httperr.Convert(rendered)
//
// Example Output:
//
//	httpStatus = 401 (Unauthorized)
//	httpBody =
//	{
//	    "message": "provided password is invalid",
//	    "code": "Unauthenticated",
//	    "localized": {
//	        "userMessage": "The entered password isn't correct. Please try again",
//	        "userMessageShort": "Not authenticated",
//	        "locale": "en-US"
//	    }
//	}
//
// The provided converters are:
//  1. GRPC (go get github.com/harwoeck/apperr/grpcerr)
//     grpcStatus, err := grpcerr.Convert(*RenderedError)
//  2. HTTP (go get github.com/harwoeck/apperr/httperr)
//     httpStatus, httpBody, err := httperr.Convert(*RenderedError)
//  3. Twirp (go get github.com/harwoeck/apperr/twirperr)
//     twirpError := twirperr.Convert(*RenderedError)
//  4. Terminal (go get github.com/harwoeck/apperr/terminalerr)
//     fmt.Println(terminalerr.Convert(*RenderedError))
package apperr
