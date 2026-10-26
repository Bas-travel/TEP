package api

import (
	"net/http"
	"strings"

	"github.com/Bas-travel/tep/services/api-go/internal/httpx"
)

func (s *Server) communityRoutes(m *http.ServeMux) {
	m.HandleFunc("GET /v1/community/feed", s.authed(s.communityFeed))
	m.HandleFunc("POST /v1/community/posts", s.authed(s.createPost))
	m.HandleFunc("GET /v1/community/posts/{postId}", s.authed(s.getPost))
	m.HandleFunc("DELETE /v1/community/posts/{postId}", s.authed(s.deletePost))
	m.HandleFunc("POST /v1/community/posts/{postId}/like", s.authed(s.likePost))
	m.HandleFunc("POST /v1/community/posts/{postId}/save", s.authed(s.savePost))
	m.HandleFunc("POST /v1/community/posts/{postId}/hide", s.authed(s.hidePost))
	m.HandleFunc("POST /v1/community/posts/{postId}/report", s.authed(s.reportPost))
	m.HandleFunc("POST /v1/community/posts/{postId}/comments", s.authed(s.createComment))
	m.HandleFunc("POST /v1/community/posts/{postId}/copy-trip", s.authed(s.copyTrip))
	m.HandleFunc("GET /v1/community/groups", s.authed(s.listGroups))
	m.HandleFunc("GET /v1/community/groups/{groupId}", s.authed(s.getGroup))
	m.HandleFunc("POST /v1/community/groups/{groupId}/join", s.authed(s.toggleGroup))
	m.HandleFunc("GET /v1/community/members/{memberId}", s.authed(s.getMember))
	m.HandleFunc("POST /v1/community/members/{memberId}/follow", s.authed(s.toggleFollow))
}

func (s *Server) communityFeed(w http.ResponseWriter, r *http.Request, uid int64) {
	mode, dest := r.URL.Query().Get("mode"), r.URL.Query().Get("destination")
	if mode != "" && mode != "all" && mode != "following" {
		httpx.Invalid(w, r, fe("mode", "must be all or following"))
		return
	}
	var items []Post
	err := s.repo.Tx(r.Context(), func(tx Tx) (e error) { items, e = tx.Feed(uid, mode, dest); return })
	if s.fail(w, r, err) {
		return
	}
	if items == nil {
		items = []Post{}
	}
	httpx.WriteJSON(w, 200, map[string]any{"items": items})
}

func (s *Server) createPost(w http.ResponseWriter, r *http.Request, uid int64) {
	var in struct {
		Type           string `json:"type"`
		Destination    string `json:"destination"`
		GroupID        *int64 `json:"groupId"`
		Text           string `json:"text"`
		ImageObjectKey string `json:"imageObjectKey"`
	}
	if err := httpx.Decode(r, &in); err != nil {
		httpx.Invalid(w, r, fe("body", err.Error()))
		return
	}
	in.Text = strings.TrimSpace(in.Text)
	var errs []httpx.FieldError
	if in.Type != "story" && in.Type != "tip" && in.Type != "question" {
		errs = append(errs, fe("type", "must be story, tip or question"))
	}
	if n := len([]rune(in.Text)); n < 10 || n > 500 {
		errs = append(errs, fe("text", "must be 10..500 characters"))
	}
	if len([]rune(in.Destination)) > 100 {
		errs = append(errs, fe("destination", "max 100 characters"))
	}
	if in.ImageObjectKey != "" && (!strings.HasPrefix(in.ImageObjectKey, "post_image/") || len(in.ImageObjectKey) > 255 || strings.Contains(in.ImageObjectKey, "..")) {
		errs = append(errs, fe("imageObjectKey", "must be a key returned by /media/presign for purpose post_image"))
	}
	if len(errs) > 0 {
		httpx.Invalid(w, r, errs...)
		return
	}
	var out *Post
	err := s.repo.Tx(r.Context(), func(tx Tx) error {
		p, e := tx.AddPost(uid, Post{Type: in.Type, Destination: in.Destination, GroupID: in.GroupID, Text: in.Text}, in.ImageObjectKey, s.now())
		if e == ErrNotFound {
			return invalid(fe("groupId", "unknown group"))
		}
		out = p
		return e
	})
	if s.fail(w, r, err) {
		return
	}
	httpx.WriteJSON(w, 201, out)
}

func (s *Server) getPost(w http.ResponseWriter, r *http.Request, uid int64) {
	id, ok := pathID(r, "postId")
	if !ok {
		httpx.NotFound(w, r)
		return
	}
	var out PostDetail
	err := s.repo.Tx(r.Context(), func(tx Tx) error {
		p, e := tx.PostByID(uid, id)
		if e != nil {
			return e
		}
		cs, e := tx.Comments(id)
		out = PostDetail{Post: *p, Comments: cs}
		return e
	})
	if s.fail(w, r, err) {
		return
	}
	httpx.WriteJSON(w, 200, out)
}

func (s *Server) deletePost(w http.ResponseWriter, r *http.Request, uid int64) {
	id, ok := pathID(r, "postId")
	if !ok {
		httpx.NotFound(w, r)
		return
	}
	if s.fail(w, r, s.repo.Tx(r.Context(), func(tx Tx) error { return tx.DeletePost(uid, id) })) {
		return
	}
	w.WriteHeader(204)
}

func (s *Server) likePost(w http.ResponseWriter, r *http.Request, uid int64) {
	id, ok := pathID(r, "postId")
	if !ok {
		httpx.NotFound(w, r)
		return
	}
	var liked bool
	var n int
	err := s.repo.Tx(r.Context(), func(tx Tx) (e error) { liked, n, e = tx.TogglePostLike(uid, id); return })
	if s.fail(w, r, err) {
		return
	}
	httpx.WriteJSON(w, 200, map[string]any{"liked": liked, "likeCount": n})
}

func (s *Server) savePost(w http.ResponseWriter, r *http.Request, uid int64) {
	id, ok := pathID(r, "postId")
	if !ok {
		httpx.NotFound(w, r)
		return
	}
	var saved bool
	err := s.repo.Tx(r.Context(), func(tx Tx) (e error) { saved, e = tx.TogglePostSave(uid, id); return })
	if s.fail(w, r, err) {
		return
	}
	httpx.WriteJSON(w, 200, map[string]any{"saved": saved})
}

func (s *Server) hidePost(w http.ResponseWriter, r *http.Request, uid int64) {
	id, ok := pathID(r, "postId")
	if !ok {
		httpx.NotFound(w, r)
		return
	}
	if s.fail(w, r, s.repo.Tx(r.Context(), func(tx Tx) error { return tx.HidePost(uid, id) })) {
		return
	}
	w.WriteHeader(204)
}

func (s *Server) reportPost(w http.ResponseWriter, r *http.Request, uid int64) {
	id, ok := pathID(r, "postId")
	if !ok {
		httpx.NotFound(w, r)
		return
	}
	var in struct {
		Reason string `json:"reason"`
	}
	if err := httpx.Decode(r, &in); err != nil {
		httpx.Invalid(w, r, fe("body", err.Error()))
		return
	}
	switch in.Reason {
	case "spam", "misinformation", "harassment", "privacy", "other":
	default:
		httpx.Invalid(w, r, fe("reason", "must be spam, misinformation, harassment, privacy or other"))
		return
	}
	if s.fail(w, r, s.repo.Tx(r.Context(), func(tx Tx) error { return tx.ReportPost(uid, id, in.Reason, s.now()) })) {
		return
	}
	w.WriteHeader(204)
}

func (s *Server) createComment(w http.ResponseWriter, r *http.Request, uid int64) {
	id, ok := pathID(r, "postId")
	if !ok {
		httpx.NotFound(w, r)
		return
	}
	var in struct {
		Text string `json:"text"`
	}
	if err := httpx.Decode(r, &in); err != nil {
		httpx.Invalid(w, r, fe("body", err.Error()))
		return
	}
	in.Text = strings.TrimSpace(in.Text)
	if n := len([]rune(in.Text)); n < 1 || n > 1000 {
		httpx.Invalid(w, r, fe("text", "must be 1..1000 characters"))
		return
	}
	var c Comment
	err := s.repo.Tx(r.Context(), func(tx Tx) (e error) { c, e = tx.AddComment(uid, id, in.Text, s.now()); return })
	if s.fail(w, r, err) {
		return
	}
	httpx.WriteJSON(w, 201, c)
}

// copyTrip saves the itinerary shared in a "trip" post to the caller's saved trips.
func (s *Server) copyTrip(w http.ResponseWriter, r *http.Request, uid int64) {
	id, ok := pathID(r, "postId")
	if !ok {
		httpx.NotFound(w, r)
		return
	}
	var t SavedTrip
	err := s.repo.Tx(r.Context(), func(tx Tx) error {
		p, e := tx.PostByID(uid, id)
		if e != nil {
			return e
		}
		if p.SharedTrip == nil {
			return invalid(fe("postId", "this post does not share a trip"))
		}
		t, e = tx.AddTrip(uid, SavedTripInput{Destination: p.Destination, Days: p.SharedTrip.Days, Travelers: 1, Budget: p.SharedTrip.Budget, Source: "community"}, s.now())
		return e
	})
	if s.fail(w, r, err) {
		return
	}
	httpx.WriteJSON(w, 201, t)
}

func (s *Server) listGroups(w http.ResponseWriter, r *http.Request, uid int64) {
	var items []Group
	err := s.repo.Tx(r.Context(), func(tx Tx) (e error) { items, e = tx.Groups(uid); return })
	if s.fail(w, r, err) {
		return
	}
	httpx.WriteJSON(w, 200, map[string]any{"items": items})
}

func (s *Server) getGroup(w http.ResponseWriter, r *http.Request, uid int64) {
	id, ok := pathID(r, "groupId")
	if !ok {
		httpx.NotFound(w, r)
		return
	}
	var g *GroupDetail
	err := s.repo.Tx(r.Context(), func(tx Tx) (e error) { g, e = tx.GroupByID(uid, id); return })
	if s.fail(w, r, err) {
		return
	}
	httpx.WriteJSON(w, 200, g)
}

func (s *Server) toggleGroup(w http.ResponseWriter, r *http.Request, uid int64) {
	id, ok := pathID(r, "groupId")
	if !ok {
		httpx.NotFound(w, r)
		return
	}
	var joined bool
	err := s.repo.Tx(r.Context(), func(tx Tx) (e error) { joined, e = tx.ToggleGroup(uid, id); return })
	if s.fail(w, r, err) {
		return
	}
	httpx.WriteJSON(w, 200, map[string]any{"joined": joined})
}

func (s *Server) getMember(w http.ResponseWriter, r *http.Request, uid int64) {
	var m *MemberProfile
	err := s.repo.Tx(r.Context(), func(tx Tx) (e error) { m, e = tx.Member(uid, r.PathValue("memberId")); return })
	if s.fail(w, r, err) {
		return
	}
	httpx.WriteJSON(w, 200, m)
}

func (s *Server) toggleFollow(w http.ResponseWriter, r *http.Request, uid int64) {
	handle := r.PathValue("memberId")
	var following bool
	err := s.repo.Tx(r.Context(), func(tx Tx) error {
		me, e := tx.UserByID(uid, false)
		if e != nil {
			return e
		}
		if me.Handle == handle {
			return invalid(fe("memberId", "you cannot follow yourself"))
		}
		following, e = tx.ToggleFollow(uid, handle)
		return e
	})
	if s.fail(w, r, err) {
		return
	}
	httpx.WriteJSON(w, 200, map[string]any{"following": following})
}
