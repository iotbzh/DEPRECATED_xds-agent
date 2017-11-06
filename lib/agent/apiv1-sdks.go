package agent

// sdksPassthroughInit Declare passthrough routes for sdks
func (s *APIService) sdksPassthroughInit(svr *XdsServer) error {
	svr.PassthroughGet("/sdks")
	svr.PassthroughGet("/sdks/:id")

	return nil
}
