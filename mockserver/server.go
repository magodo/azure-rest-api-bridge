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

	"github.com/go-openapi/spec"
	"github.com/golang-jwt/jwt/v5"
	"github.com/magodo/azure-rest-api-bridge/log"
	"github.com/magodo/azure-rest-api-bridge/mockserver/swagger"
	"github.com/magodo/azure-rest-api-index/azidx"
)

type Server struct {
	addr       string
	port       int
	server     *http.Server
	shutdownCh <-chan struct{}

	idx     azidx.Index
	specdir string

	overrides []Override
}

type Override struct {
	PathPattern regexp.Regexp
	Response    []byte
}

type Option struct {
	Addr    string
	Port    int
	Index   string
	SpecDir string
}

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
		addr:    opt.Addr,
		port:    opt.Port,
		idx:     index,
		specdir: opt.SpecDir,
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

	ref, err := srv.idx.Lookup(r.Method, *r.URL)
	if err != nil {
		srv.writeError(w, err)
		return
	}
	exp, err := swagger.NewExpanderFromGet(spec.MustCreateRef(filepath.Join(srv.specdir, ref.GetURL().Path) + "#" + ref.GetPointer().String()))
	if err != nil {
		srv.writeError(w, err)
		return
	}
	if err := exp.Expand(); err != nil {
		srv.writeError(w, err)
		return
	}
	syn := swagger.NewSynthesizer(exp.Root(), swagger.NewRnd(nil))
	resps := syn.Synthesize()
	resp := resps[0]
	b, err := json.Marshal(resp)
	if err != nil {
		srv.writeError(w, err)
		return
	}
	w.Write(b)
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
		Addr:         fmt.Sprintf("%s:%d", srv.addr, srv.port),
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
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

func (srv *Server) UpdateOverrides(ov []Override) {
	srv.overrides = ov
}
