package agent

// sdksPassthroughInit Declare passthrough routes for sdks
func (s *APIService) sdksPassthroughInit(svr *XdsServer) error {
	svr.PassthroughGet("/sdks")
	svr.PassthroughGet("/sdk/:id")

	return nil
}
