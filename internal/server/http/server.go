package internalhttp

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/ids79/anti-bruteforcer/internal/app"
	"github.com/ids79/anti-bruteforcer/internal/config"
	"github.com/ids79/anti-bruteforcer/internal/logger"
)

type Server struct {
	logg    logger.Logg
	backets app.WorkWithBackets
	iplist  app.WorkWithIPList
	conf    *config.Config
	srv     *http.Server
}

func NewServer(backets *app.Backets, iplist *app.IPList, logg logger.Logg, config config.Config) *Server {
	return &Server{
		logg:    logg,
		backets: backets,
		iplist:  iplist,
		conf:    &config,
	}
}

func (s *Server) Start() error {
	handler := s.loggingMiddleware()
	server := &http.Server{
		Addr:         s.conf.HTTPServer.Address,
		Handler:      handler,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}
	s.srv = server
	s.logg.Info("starting http server on ", server.Addr)
	server.ListenAndServe()
	if err := server.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
		s.logg.Error("HTTP Start: ", err)
	}
	return nil
}

func (s *Server) Close() {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
	defer cancel()
	s.logg.Info("server http is stopping...")
	if err := s.srv.Shutdown(ctx); err != nil {
		s.logg.Error("failed to stop http server: ", err)
	}
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.URL.Path {
	case "/":
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode("Hello anti-bruteforce!")
	case "/add-white-list/":
		s.addWhite(w, r)
	case "/del-white-list/":
		s.delWhite(w, r)
	case "/add-black-list/":
		s.addBlack(w, r)
	case "/del-black-list":
		s.delBlak(w, r)
	case "/auth/":
		s.auth(w, r)
	case "/reset-bucket/":
		s.resetBucket(w, r)
	case "/get-list/":
		s.getList(w, r)
	default:
		http.NotFound(w, r)
	}
}

func (s *Server) addWhite(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		query := r.URL.Query()
		ipStr := query.Get("ip")
		maskStr := query.Get("mask")
		err := s.iplist.AddWhiteList(ipStr, maskStr)
		switch {
		case err == nil:
			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte("IP was added in the white list"))
		case errors.Is(err, app.ErrBadRequest):
			http.Error(w, err.Error(), http.StatusBadRequest)
		default:
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	} else {
		http.NotFound(w, r)
	}
}

func (s *Server) delWhite(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		query := r.URL.Query()
		ipStr := query.Get("ip")
		err := s.iplist.DelWhiteList(ipStr)
		switch {
		case err == nil:
			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte("IP was deleted from the white list"))
		case errors.Is(err, app.ErrBadRequest):
			http.Error(w, err.Error(), http.StatusBadRequest)
		default:
			http.Error(w, "error when deleting IP", http.StatusInternalServerError)
		}
	} else {
		http.NotFound(w, r)
	}
}

func (s *Server) addBlack(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		query := r.URL.Query()
		ipStr := query.Get("ip")
		maskStr := query.Get("mask")
		err := s.iplist.AddBlackList(ipStr, maskStr)
		switch {
		case err == nil:
			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte("IP was added in the black list"))
		case errors.Is(err, app.ErrBadRequest):
			http.Error(w, err.Error(), http.StatusBadRequest)
		default:
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	} else {
		http.NotFound(w, r)
	}
}

func (s *Server) delBlak(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		query := r.URL.Query()
		ipStr := query.Get("ip")
		err := s.iplist.DelBlackList(ipStr)
		switch {
		case err == nil:
			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte("IP was deleted from the black list"))
		case errors.Is(err, app.ErrBadRequest):
			http.Error(w, err.Error(), http.StatusBadRequest)
		default:
			http.Error(w, "error when deleting IP", http.StatusInternalServerError)
		}
	}
	http.NotFound(w, r)
}

func (s *Server) auth(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.NotFound(w, r)
	}
	query := r.URL.Query()
	ipStr := query.Get("ip")
	login := query.Get("login")
	pass := query.Get("pass")
	w.Header().Set("Content-Type", "application/json")
	if ok, err := s.iplist.IsInBlackList(ipStr); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	} else if ok {
		w.Write([]byte{0b0})
		return
	}
	if ok, err := s.iplist.IsInWhiteList(ipStr); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	} else if ok {
		w.Write([]byte{0b1})
		return
	}
	ipOK, err := s.backets.AccessVerification(ipStr, app.IP)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	loginOK, err := s.backets.AccessVerification(login, app.LOGIN)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	passOK, err := s.backets.AccessVerification(pass, app.PASSWORD)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if ipOK && loginOK && passOK {
		w.Write([]byte{0b1})
	} else {
		w.Write([]byte{0b0})
	}
}

func (s *Server) resetBucket(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.NotFound(w, r)
	}
	w.Header().Set("Content-Type", "application/json")
	query := r.URL.Query()
	ipStr := query.Get("ip")
	if _, err := app.IPtoInt(ipStr); err == nil {
		if errors.Is(s.backets.ResetBacket(ipStr, app.IP), app.ErrBacketNotFound) {
			w.Write([]byte("IP was no found in backets\n"))
		} else {
			w.Write([]byte("IP backet was reset\n"))
		}
	}
	login := query.Get("login")
	if login != "" {
		if errors.Is(s.backets.ResetBacket(login, app.LOGIN), app.ErrBacketNotFound) {
			w.Write([]byte("login was no found in backets\n"))
		} else {
			w.Write([]byte("login backet was reset\n"))
		}
	}
	pass := query.Get("pass")
	if pass != "" {
		if errors.Is(s.backets.ResetBacket(pass, app.PASSWORD), app.ErrBacketNotFound) {
			w.Write([]byte("password was no found in backets"))
		} else {
			w.Write([]byte("password backet was reset"))
		}
	}
}

func (s *Server) getList(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		w.Header().Set("Content-Type", "application/json")
		query := r.URL.Query()
		t := query.Get("type")
		list := s.iplist.GetList(t)
		body, err := json.Marshal(list)
		if err != nil {
			s.logg.Error("error while marshaling: ", err)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write(body)
	} else {
		http.NotFound(w, r)
	}
}
