package mockserver

import (
	"context"
	_ "embed"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	jsonpatch "github.com/evanphx/json-patch/v5"
	"github.com/go-openapi/spec"
	"github.com/golang-jwt/jwt/v5"
	"github.com/magodo/azure-rest-api-bridge/log"
	"github.com/magodo/azure-rest-api-bridge/mockserver/swagger"
	"github.com/magodo/azure-rest-api-index/azidx"
)

type Server struct {
	Addr       string
	Port       int
	timeout    time.Duration
	server     *http.Server
	shutdownCh <-chan struct{}

	Idx     azidx.Index
	Specdir string

	// Followings are execution-based
	rnd       swagger.Rnd
	overrides Overrides
	records   []swagger.JSONValue
}

type Overrides []Override

type Override struct {
	PathPattern           regexp.Regexp
	ResponseSelectorMerge string
	ResponseSelectorJSON  string
	ResponseBody          string
	ResponsePatchMerge    string
	ResponsePatchJSON     string
	SynthOption           *swagger.SynthesizerOption
	ExpanderOption        *swagger.ExpanderOption
}

func (ovs Overrides) Match(path string) *Override {
	for _, ov := range ovs {
		ov := ov
		if ov.PathPattern.MatchString(path) {
			return &ov
		}
	}
	return nil
}

type Option struct {
	Addr    string
	Port    int
	Index   string
	SpecDir string
	Timeout time.Duration
}

// New creates a new (uninitialized) mockserver, which can be started, but needs to be initiated in order to work as expected.
func New(opt Option) (*Server, error) {
	b, err := os.ReadFile(opt.Index)
	if err != nil {
		return nil, fmt.Errorf("reading index file %s: %v", opt.Index, err)
	}
	var index azidx.Index
	if err := json.Unmarshal(b, &index); err != nil {
		return nil, fmt.Errorf("unmarshal index file: %v", err)
	}
	return &Server{
		Addr:    opt.Addr,
		Port:    opt.Port,
		Idx:     index,
		Specdir: opt.SpecDir,
		timeout: opt.Timeout,
	}, nil
}

func (srv *Server) writeError(w http.ResponseWriter, err error) {
	log.Error(err.Error())
	w.WriteHeader(http.StatusInternalServerError)
	w.Write([]byte(fmt.Sprintf(`{"error": %q}`, err.Error())))
}

func (srv *Server) Handle(w http.ResponseWriter, r *http.Request) {
	if strings.HasSuffix(r.URL.Path, "/oauth2/v2.0/token") {
		srv.handleToken(w, r)
		return
	}

	ov := srv.overrides.Match(r.URL.Path)

	// Override response body, just return the hardcoded response body
	if ov != nil && ov.ResponseBody != "" {
		log.Debug("override", "type", "body", "url", r.URL.String(), "value", ov.ResponseBody)
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(ov.ResponseBody))
		return
	}

	// Otherwise, we'll synthesize the response based on its swagger definition
	var (
		synthOpt    *swagger.SynthesizerOption
		expanderOpt *swagger.ExpanderOption
	)
	if ov != nil {
		synthOpt = ov.SynthOption
		expanderOpt = ov.ExpanderOption
	}

	resps, expRoot, err := srv.synthResponse(r, synthOpt, expanderOpt)
	if err != nil {
		srv.writeError(w, err)
		return
	}

	b, err := srv.selResponse(resps, ov)
	if err != nil {
		srv.writeError(w, err)
		return
	}

	if ov != nil {
		switch {
		case ov.ResponsePatchMerge != "":
			log.Debug("override", "type", "merge patch", "url", r.URL.String(), "value", ov.ResponsePatchMerge)
			b, err = jsonpatch.MergePatch(b, []byte(ov.ResponsePatchMerge))
			if err != nil {
				srv.writeError(w, err)
				return
			}
		case ov.ResponsePatchJSON != "":
			log.Debug("override", "type", "json patch", "url", r.URL.String(), "value", ov.ResponsePatchJSON)
			patch, err := jsonpatch.DecodePatch([]byte(ov.ResponsePatchJSON))
			if err != nil {
				srv.writeError(w, err)
				return
			}
			b, err = patch.Apply(b)
			if err != nil {
				srv.writeError(w, err)
				return
			}
		}
	}

	log.Debug("server handler", "response", string(b))

	v, err := swagger.UnmarshalJSONToJSONValue(b, expRoot)
	if err != nil {
		srv.writeError(w, fmt.Errorf("unmarshal JSON to JSONValue: %v", err))
		return
	}
	srv.records = append(srv.records, v)
	w.Header().Set("Content-Type", "application/json")
	w.Write(b)
	return
}

func (srv *Server) synthResponse(r *http.Request, synthOpt *swagger.SynthesizerOption, expanderOpt *swagger.ExpanderOption) ([]interface{}, *swagger.Property, error) {
	ref, err := srv.Idx.Lookup(r.Method, *r.URL)
	if err != nil {
		return nil, nil, err
	}
	exp, err := swagger.NewExpanderFromOpRef(spec.MustCreateRef(filepath.Join(srv.Specdir, ref.GetURL().Path)+"#"+ref.GetPointer().String()), expanderOpt)
	if err != nil {
		return nil, nil, err
	}
	if err := exp.Expand(); err != nil {
		return nil, nil, err
	}
	syn := swagger.NewSynthesizer(exp.Root(), &srv.rnd, synthOpt)
	return syn.Synthesize(), exp.Root(), nil
}

func (srv *Server) selResponse(resps []interface{}, ov *Override) ([]byte, error) {
	if len(resps) == 0 {
		return nil, fmt.Errorf("no responses to select")
	}

	if len(resps) == 1 {
		return json.Marshal(resps[0])
	}

	logDiff := func(resps [][]byte) {
		if len(resps) >= 2 {
			diff, err := jsonpatch.CreateMergePatch(resps[1], resps[0])
			if err == nil {
				log.Warn(fmt.Sprintf("The first two responses have following diff (resp2 -> resp1):\n%s", string(diff)))
			}
		}
	}

	if ov == nil || ov.ResponseSelectorMerge == "" && ov.ResponseSelectorJSON == "" {
		log.Warn(fmt.Sprintf("select the 1st response from %d", len(resps)))
		b1, err := json.Marshal(resps[0])
		if err != nil {
			return nil, err
		}
		if b2, err := json.Marshal(resps[1]); err == nil {
			logDiff([][]byte{b1, b2})
		}
		return b1, nil
	}

	selector := ov.ResponseSelectorMerge
	if selector == "" {
		selector = ov.ResponseSelectorJSON
	}

	var candidates [][]byte
	for _, resp := range resps {
		bOld, err := json.Marshal(resp)
		if err != nil {
			return nil, err
		}

		log.Debug("override", "type", "selector", "sel", selector, "resp", string(bOld))

		// Each selector is a json merge patch, we expect to apply this patch to the response and pick
		// the one that has no difference between the itself and with the patch applied.

		var bNew []byte
		if ov.ResponseSelectorMerge != "" {
			bNew, err = jsonpatch.MergePatch(bOld, []byte(ov.ResponseSelectorMerge))
		} else {
			patch, err := jsonpatch.DecodePatch([]byte(ov.ResponseSelectorJSON))
			if err != nil {
				return nil, fmt.Errorf("decoding response selector json patch: %v", err)
			}
			bNew, err = patch.Apply(bOld)
		}
		if err != nil {
			return nil, fmt.Errorf("applying response selector patch: %v", err)
		}

		if string(bOld) == string(bNew) {
			candidates = append(candidates, bOld)
		}
	}

	if len(candidates) == 0 {
		return nil, fmt.Errorf("no synth response found with the response selector: %s", selector)
	}

	if len(candidates) > 1 {
		log.Warn(fmt.Sprintf("select the 1st response from %d (after selection)", len(candidates)))
		logDiff(candidates)
		return candidates[0], nil
	}

	return candidates[0], nil
}

func (srv *Server) handleToken(w http.ResponseWriter, r *http.Request) {
	now := time.Now()
	exp := now.Add(time.Duration(24) * time.Hour)
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"nbf":   now.Unix(),
		"iat":   now.Unix(),
		"exp":   exp.Unix(),
		"oid":   "00000000-0000-0000-000000000000",
		"appid": "00000000-0000-0000-000000000000",
	})

	tokenString, err := token.SignedString([]byte("secret"))
	if err != nil {
		srv.writeError(w, err)
		return
	}

	type AzureToken struct {
		AccessToken  string `json:"access_token"`
		ExpiresIn    int64  `json:"expires_in"`
		ExtExpiresIn int64  `json:"ext_expires_in"`
		TokenType    string `json:"token_type"`
	}

	tk := AzureToken{
		AccessToken:  tokenString,
		ExpiresIn:    now.Unix(),
		ExtExpiresIn: now.Unix(),
		TokenType:    "Bearer",
	}

	b, err := json.Marshal(tk)
	if err != nil {
		srv.writeError(w, err)
		return
	}
	w.Write(b)
}

func (srv *Server) Start() error {
	mux := http.NewServeMux()
	mux.HandleFunc("/", srv.Handle)
	server := &http.Server{
		Addr:         fmt.Sprintf("%s:%d", srv.Addr, srv.Port),
		ReadTimeout:  srv.timeout,
		WriteTimeout: srv.timeout,
		Handler:      mux,
	}
	shutdownCh := make(chan struct{})
	go func() {
		if err := server.ListenAndServe(); err != http.ErrServerClosed {
			log.Error("HTTP server ListenAndServe", "error", err)
		}
		close(shutdownCh)
	}()

	srv.server = server
	srv.shutdownCh = shutdownCh

	return nil
}

func (srv *Server) Stop(ctx context.Context) error {
	if err := srv.server.Shutdown(ctx); err != nil {
		return err
	}
	<-srv.shutdownCh
	return nil
}

// InitExecution initiates for each execution, for resetting the overrides and the rnd.
func (srv *Server) InitExecution(ov []Override) {
	srv.overrides = ov
	srv.rnd = swagger.NewRnd(nil)
	srv.records = nil
}

func (srv *Server) Records() []swagger.JSONValue {
	return srv.records
}
