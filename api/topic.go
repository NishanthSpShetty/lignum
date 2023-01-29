package api

import (
	"encoding/json"
	"net/http"
)

type Topic struct {
	Name         string `json:"name"`
	MessageCount uint64 `json:"message_count"`
}

func (s *Server) handleTopicGet(w http.ResponseWriter, req *http.Request) {
	topics := make([]Topic, 0)
	topix := s.message.GetTopics()

	for _, topic := range topix {
		topics = append(topics, Topic{Name: topic.GetName(), MessageCount: topic.GetCurrentOffset()})
	}
	json.NewEncoder(w).Encode(topics)
}

func (s *Server) TopicHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch req.Method {
		//	case "POST":
		//		s.handleTopicPost(w, req)
		case "GET":
			s.handleTopicGet(w, req)
		default:
			http.Error(w, "request method must be one of [ GET, POST ].", http.StatusMethodNotAllowed)
		}
	}
}
