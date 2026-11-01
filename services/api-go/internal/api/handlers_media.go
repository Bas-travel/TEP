package api

import (
	"net/http"
	"strconv"
	"time"

	"github.com/Bas-travel/tep/services/api-go/internal/httpx"
	"github.com/Bas-travel/tep/services/api-go/internal/objstore"
)

var uploadTypes = map[string]string{"image/jpeg": ".jpg", "image/png": ".png", "image/webp": ".webp"}

// presignUpload returns a short-lived PUT URL for direct-to-storage uploads. Keys are namespaced by purpose and user,
// with an unguessable name, so a client can never overwrite another user's object.
func (s *Server) presignUpload(w http.ResponseWriter, r *http.Request, uid int64) {
	var in struct {
		ContentType string `json:"contentType"`
		Purpose     string `json:"purpose"`
	}
	if err := httpx.Decode(r, &in); err != nil {
		httpx.Invalid(w, r, fe("body", err.Error()))
		return
	}
	var errs []httpx.FieldError
	ext, okType := uploadTypes[in.ContentType]
	if !okType {
		errs = append(errs, fe("contentType", "must be image/jpeg, image/png or image/webp"))
	}
	switch in.Purpose {
	case "post_image", "avatar":
	case "listing_image":
		// Listing media is managed by partners/staff (portals are post-MVP), never by travelers.
		httpx.WriteProblem(w, r, 403, httpx.CodeForbidden, "Forbidden", "listing images are managed by partners")
		return
	default:
		errs = append(errs, fe("purpose", "must be listing_image, post_image or avatar"))
	}
	if len(errs) > 0 {
		httpx.Invalid(w, r, errs...)
		return
	}
	key := in.Purpose + "/" + itoa64(uid) + "/" + randID("") + ext
	const ttl = 10 * time.Minute
	var url string
	if s.cfg.S3AccessKey == "" || s.cfg.S3Endpoint == "" {
		if s.cfg.APIMode != "mock" {
			httpx.WriteProblem(w, r, 503, httpx.CodeInternal, "Service Unavailable", "object storage is not configured")
			return
		}
		url = "https://mock.tep.local/upload/" + key // DEV mock mode: nothing is stored
	} else {
		sg := objstore.Signer{Endpoint: s.cfg.S3Endpoint, Bucket: s.cfg.S3Bucket, Region: s.cfg.S3Region, AccessKey: s.cfg.S3AccessKey, SecretKey: s.cfg.S3SecretKey}
		var err error
		if url, err = sg.Presign("PUT", key, ttl, s.now()); err != nil {
			s.fail(w, r, err)
			return
		}
	}
	httpx.WriteJSON(w, 200, map[string]any{"uploadUrl": url, "objectKey": key, "expiresIn": int(ttl.Seconds())})
}

func itoa64(n int64) string { return strconv.FormatInt(n, 10) }
