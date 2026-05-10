package fixture

var authz = struct {
	Require    func(any)
	RequireAll func(any)
}{}

type Server struct{}

type createWidgetRequest struct {
	Body struct {
		AreaCode string
	}
}

func (s *Server) CreateWidget(req createWidgetRequest) {
	authz.Require(req.Body.AreaCode)
}
