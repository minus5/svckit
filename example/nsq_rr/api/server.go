package api

import "fmt"

type Service interface {
	Add(int, int) (int, error)
}

type Server struct {
	svc Service
}

func NewServer(svc Service) *Server {
	return &Server{svc: svc}
}

func (s *Server) Serve(i interface{}) (interface{}, error) {
	switch req := i.(type) {
	case *addReq:
		z, err := s.svc.Add(req.X, req.Y)
		if err != nil {
			return nil, err
		}
		return &addRsp{Z: z}, nil
	default:
		return nil, fmt.Errorf("unknown type %T", req)
	}
}
