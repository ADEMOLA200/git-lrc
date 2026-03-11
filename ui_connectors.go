package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/knadh/koanf/parsers/toml"
	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/v2"
	"github.com/urfave/cli/v2"
)

const defaultUIPort = 8090

const (
	aiConnectorsSectionBegin = "# BEGIN lrc managed ai_connectors"
	aiConnectorsSectionEnd   = "# END lrc managed ai_connectors"
)

type uiRuntimeConfig struct {
	APIURL        string
	JWT           string
	RefreshJWT    string
	OrgID         string
	UserEmail     string
	UserID        string
	FirstName     string
	LastName      string
	AvatarURL     string
	OrgName       string
	ConfigPath    string
	ConfigErr     string
	ConfigMissing bool
}

type aiConnectorRemote struct {
	ID            int64  `json:"id"`
	ProviderName  string `json:"provider_name"`
	ConnectorName string `json:"connector_name"`
	APIKey        string `json:"api_key"`
	BaseURL       string `json:"base_url"`
	SelectedModel string `json:"selected_model"`
	DisplayOrder  int    `json:"display_order"`
	CreatedAt     string `json:"created_at"`
	UpdatedAt     string `json:"updated_at"`
}

type connectorManagerServer struct {
	cfg    *uiRuntimeConfig
	client *http.Client
	mu     sync.Mutex
}

type authRefreshRequest struct {
	RefreshToken string `json:"refresh_token"`
}

type authRefreshResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

func runUI(c *cli.Context) error {
	cfg, err := loadUIRuntimeConfig()
	if err != nil {
		return err
	}

	ln, port, err := pickServePort(defaultUIPort, 20)
	if err != nil {
		return fmt.Errorf("failed to reserve UI port: %w", err)
	}

	srv := &connectorManagerServer{
		cfg:    cfg,
		client: &http.Client{Timeout: 30 * time.Second},
	}

	mux := http.NewServeMux()
	mux.Handle("/static/", http.StripPrefix("/static/", getStaticHandler()))
	mux.HandleFunc("/", srv.handleIndex)
	mux.HandleFunc("/api/ui/session-status", srv.handleSessionStatus)
	mux.HandleFunc("/api/ui/auth/reauth", srv.handleReauthenticate)
	mux.HandleFunc("/api/ui/connectors/reorder", srv.handleReorder)
	mux.HandleFunc("/api/ui/connectors/validate-key", srv.handleValidateKey)
	mux.HandleFunc("/api/ui/connectors/ollama/models", srv.handleOllamaModels)
	mux.HandleFunc("/api/ui/connectors/", srv.handleConnectorByID)
	mux.HandleFunc("/api/ui/connectors", srv.handleConnectors)

	httpServer := &http.Server{Handler: mux}
	go func() {
		if err := httpServer.Serve(ln); err != nil && err != http.ErrServerClosed {
			log.Printf("ui server error: %v", err)
		}
	}()

	url := fmt.Sprintf("http://localhost:%d", port)
	fmt.Printf("\n🌐 git-lrc Manager UI available at: %s\n\n", highlightURL(url))
	go func() {
		time.Sleep(300 * time.Millisecond)
		_ = openURL(url)
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)
	<-quit

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	return httpServer.Shutdown(ctx)
}

func loadUIRuntimeConfig() (*uiRuntimeConfig, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to determine home directory: %w", err)
	}

	configPath := filepath.Join(homeDir, ".lrc.toml")
	cfg := &uiRuntimeConfig{
		APIURL:        defaultAPIURL,
		ConfigPath:    configPath,
		ConfigMissing: false,
	}

	if _, err := os.Stat(configPath); err != nil {
		if os.IsNotExist(err) {
			cfg.ConfigErr = fmt.Sprintf("config file not found at %s", configPath)
			cfg.ConfigMissing = true
			return cfg, nil
		}
		cfg.ConfigErr = fmt.Sprintf("failed to read config file %s: %v", configPath, err)
		return cfg, nil
	}

	k := koanf.New(".")
	if err := k.Load(file.Provider(configPath), toml.Parser()); err != nil {
		cfg.ConfigErr = fmt.Sprintf("failed to load config file %s: %v", configPath, err)
		return cfg, nil
	}

	apiURL := strings.TrimSpace(k.String("api_url"))
	if apiURL == "" {
		apiURL = defaultAPIURL
	}

	cfg.APIURL = apiURL
	cfg.JWT = strings.TrimSpace(k.String("jwt"))
	cfg.RefreshJWT = strings.TrimSpace(k.String("refresh_token"))
	cfg.OrgID = strings.TrimSpace(k.String("org_id"))
	cfg.UserEmail = strings.TrimSpace(k.String("user_email"))
	cfg.UserID = strings.TrimSpace(k.String("user_id"))
	cfg.FirstName = strings.TrimSpace(k.String("user_first_name"))
	cfg.LastName = strings.TrimSpace(k.String("user_last_name"))
	cfg.AvatarURL = strings.TrimSpace(k.String("avatar_url"))
	cfg.OrgName = strings.TrimSpace(k.String("org_name"))

	return cfg, nil
}
