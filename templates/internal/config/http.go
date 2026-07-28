// [[ when (modeIs "http") ]]
package config

import (
	"fmt"
	"strings"
	"time"
)

// HTTPSection maps YAML/ENV/flag block `http`.
//
// The timeout and size limits bound how long a single connection may occupy a
// goroutine and how much a client may send, so a slow or hostile peer cannot
// exhaust server resources.
type HTTPSection struct {
	ListenAddr        string        `yaml:"listen_addr,omitempty"`
	ReadTimeout       time.Duration `yaml:"read_timeout,omitempty"`
	ReadHeaderTimeout time.Duration `yaml:"read_header_timeout,omitempty"`
	WriteTimeout      time.Duration `yaml:"write_timeout,omitempty"`
	IdleTimeout       time.Duration `yaml:"idle_timeout,omitempty"`
	ShutdownTimeout   time.Duration `yaml:"shutdown_timeout,omitempty"`
	MaxHeaderBytes    int           `yaml:"max_header_bytes,omitempty"`
	MaxBodyBytes      int64         `yaml:"max_body_bytes,omitempty"`
}

func (h HTTPSection) WithDefaults() HTTPSection {
	if h.ListenAddr == "" {
		h.ListenAddr = DefaultHTTPListenAddr
	}
	if h.ReadTimeout == 0 {
		h.ReadTimeout = DefaultHTTPReadTimeout
	}
	if h.ReadHeaderTimeout == 0 {
		h.ReadHeaderTimeout = DefaultHTTPReadHeaderTimeout
	}
	if h.WriteTimeout == 0 {
		h.WriteTimeout = DefaultHTTPWriteTimeout
	}
	if h.IdleTimeout == 0 {
		h.IdleTimeout = DefaultHTTPIdleTimeout
	}
	if h.ShutdownTimeout == 0 {
		h.ShutdownTimeout = DefaultHTTPShutdownTimeout
	}
	if h.MaxHeaderBytes == 0 {
		h.MaxHeaderBytes = DefaultHTTPMaxHeaderBytes
	}
	if h.MaxBodyBytes == 0 {
		h.MaxBodyBytes = DefaultHTTPMaxBodyBytes
	}
	return h
}

func (h HTTPSection) Validate() error {
	if strings.TrimSpace(h.ListenAddr) == "" {
		return fmt.Errorf("http.listen_addr must be non-empty")
	}
	for _, f := range []struct {
		name string
		val  time.Duration
	}{
		{"http.read_timeout", h.ReadTimeout},
		{"http.read_header_timeout", h.ReadHeaderTimeout},
		{"http.write_timeout", h.WriteTimeout},
		{"http.idle_timeout", h.IdleTimeout},
		{"http.shutdown_timeout", h.ShutdownTimeout},
	} {
		if f.val <= 0 {
			return fmt.Errorf("%s must be positive, got %s", f.name, f.val)
		}
	}
	if h.MaxHeaderBytes <= 0 {
		return fmt.Errorf("http.max_header_bytes must be positive, got %d", h.MaxHeaderBytes)
	}
	if h.MaxBodyBytes <= 0 {
		return fmt.Errorf("http.max_body_bytes must be positive, got %d", h.MaxBodyBytes)
	}
	return nil
}
