package internalhttp

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/ids79/anti-bruteforce/internal/app"
	"github.com/ids79/anti-bruteforce/internal/config"
	"github.com/ids79/anti-bruteforce/internal/logger"
)

type FuncAddList func(context.Context, string) error

type Server struct {
	logg    logger.Logg
	backets app.WorkWithBackets
	iplist  app.WorkWithIPList
	conf    *config.Config
	srv     *http.Server
}

func NewServer(backets app.WorkWithBackets, iplist *app.IPList, logg logger.Logg, config *config.Config) *Server {
	return &Server{
		logg:    logg,
		backets: backets,
		iplist:  iplist,
		conf:    config,
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
		s.addList(w, r, s.iplist.AddWhiteList, "white")
	case "/del-white-list/":
		s.delList(w, r, s.iplist.DelWhiteList, "white")
	case "/add-black-list/":
		s.addList(w, r, s.iplist.AddBlackList, "black")
	case "/del-black-list":
		s.delList(w, r, s.iplist.DelWhiteList, "black")
	case "/auth/":
		s.auth(w, r)
	case "/reset-bucket/":
		s.resetBucket(w, r)
	case "/get-list/":
		s.getList(w, r)
	default:
		s.logg.Info("request not found: ", r.URL.Path)
		http.NotFound(w, r)
	}
}

func (s *Server) addList(w http.ResponseWriter, r *http.Request, f FuncAddList, typeList string) {
	if r.Method == http.MethodGet {
		query := r.URL.Query()
		ipStr := query.Get("ip")
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*2)
		defer cancel()
		err := f(ctx, ipStr)
		switch {
		case err == nil:
			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte("IP was added in the list " + typeList))
		case strings.Contains(err.Error(), app.BadRequestStr):
			s.logg.Error("addList ", typeList, ": ", err)
			http.Error(w, err.Error(), http.StatusBadRequest)
		default:
			s.logg.Error("addList ", typeList, ": ", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	} else {
		http.NotFound(w, r)
	}
}

func (s *Server) delList(w http.ResponseWriter, r *http.Request, f FuncAddList, typeList string) {
	if r.Method == http.MethodGet {
		query := r.URL.Query()
		ipStr := query.Get("ip")
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*2)
		defer cancel()
		err := f(ctx, ipStr)
		switch {
		case err == nil:
			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte("IP was deleted from the list" + typeList))
		case strings.Contains(err.Error(), app.BadRequestStr):
			s.logg.Error("delList ", typeList, err)
			http.Error(w, err.Error(), http.StatusBadRequest)
		default:
			s.logg.Error("delList ", typeList, err)
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
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*2)
	defer cancel()
	w.Header().Set("Content-Type", "application/json")
	if ok, err := s.iplist.IsInBlackList(ctx, ipStr); err != nil {
		s.logg.Error("auth: ", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	} else if ok {
		s.logg.Info(fmt.Sprintf("failed verification for ip: %s - is in blacklist", ipStr))
		w.Write([]byte{0b0})
		return
	}
	if ok, err := s.iplist.IsInWhiteList(ctx, ipStr); err != nil {
		s.logg.Error("auth: ", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	} else if ok {
		s.logg.Info(fmt.Sprintf("success verification for ip: %s - is in whitelist", ipStr))
		w.Write([]byte{0b1})
		return
	}
	ipOK, err := s.backets.AccessVerification(ipStr, app.IP)
	if err != nil {
		s.logg.Error("auth: ", app.IP, ": ", err)
		http.Error(w, "success verification failed", http.StatusBadRequest)
		return
	}
	loginOK, err := s.backets.AccessVerification(login, app.LOGIN)
	if err != nil {
		s.logg.Error("auth: ", app.LOGIN, ": ", err)
		http.Error(w, "success verification failed", http.StatusBadRequest)
		return
	}
	passOK, err := s.backets.AccessVerification(pass, app.PASSWORD)
	if err != nil {
		s.logg.Error("auth: ", app.PASSWORD, ": ", err)
		http.Error(w, "success verification failed", http.StatusBadRequest)
		return
	}
	if ipOK && loginOK && passOK {
		s.logg.Info(fmt.Sprintf("success verification for ip: %s, login: %s, pass: %s", ipStr, login, pass))
		w.Write([]byte{0b1})
	} else {
		s.logg.Info(fmt.Sprintf("failed verification for ip: %s, login: %s, pass: %s", ipStr, login, pass))
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
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*2)
	defer cancel()
	if ipStr == "" {
		s.logg.Info("resetBucket: IP was not specified")
	} else if ip := net.ParseIP(ipStr); ip == nil {
		s.logg.Error("resetBucket: ", ipStr, " - parse IP error")
	}
	if errors.Is(s.backets.ResetBacket(ctx, ipStr, app.IP), app.ErrBacketNotFound) {
		s.logg.Info("resetBucket: ", ipStr, " - ", app.ErrBacketNotFound)
		w.Write([]byte("IP was no found in backets\n"))
	} else {
		s.logg.Info("resetBucket: ", ipStr, " - backet was reset")
		w.Write([]byte("IP backet was reset\n"))
	}
	login := query.Get("login")
	if login == "" {
		s.logg.Info("resetBucket: login was not specified")
	}
	if errors.Is(s.backets.ResetBacket(ctx, login, app.LOGIN), app.ErrBacketNotFound) {
		s.logg.Info("resetBucket: ", login, " - ", app.ErrBacketNotFound)
		w.Write([]byte("login was no found in backets\n"))
	} else {
		s.logg.Info("resetBucket: ", login, " - backet was reset")
		w.Write([]byte("login backet was reset\n"))
	}
	pass := query.Get("pass")
	if pass == "" {
		s.logg.Info("resetBucket: password was not specified")
	}
	if errors.Is(s.backets.ResetBacket(ctx, pass, app.PASSWORD), app.ErrBacketNotFound) {
		s.logg.Info("resetBucket: ", pass, " - ", app.ErrBacketNotFound)
		w.Write([]byte("password was no found in backets"))
	} else {
		s.logg.Info("resetBucket: ", pass, " - backet was reset")
		w.Write([]byte("password backet was reset"))
	}
}

func (s *Server) getList(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		w.Header().Set("Content-Type", "application/json")
		query := r.URL.Query()
		t := query.Get("type")
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*2)
		defer cancel()
		list := s.iplist.GetList(ctx, t)
		body, err := json.Marshal(list)
		if err != nil {
			s.logg.Error("getList, error while marshaling: ", err)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write(body)
	} else {
		http.NotFound(w, r)
	}
}
